#!/usr/bin/env python3
"""
Devtron Documentation API Server
REST API for documentation search and re-indexing using PostgreSQL pgvector and local embeddings.
"""

import asyncio
import logging
import os
from datetime import datetime, timezone
from enum import Enum
from typing import Any, Dict, List, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from doc_processor import DocumentationProcessor
from settings import EMBEDDING_MODEL
from vector_store import VectorStore

# Configure logging
logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global instances
doc_processor: Optional[DocumentationProcessor] = None
vector_store: Optional[VectorStore] = None



class IndexState(str, Enum):
    IDLE = "idle"
    RUNNING = "running"
    DONE = "done"
    FAILED = "failed"


class IndexMode(str, Enum):
    REBUILD = "rebuild"
    UPDATE = "update"


class ServiceStatus(str, Enum):
    HEALTHY = "healthy"


class DatabaseStatus(str, Enum):
    CONNECTED = "connected"


# The one indexing job this server runs at a time, and how far it has got.
index_job: Dict[str, Any] = {"state": IndexState.IDLE, "mode": None, "started_at": None, "finished_at": None,
                             "documents_total": 0, "documents_done": 0, "error": None}
index_task: Optional[asyncio.Task] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global doc_processor, vector_store
    logger.info("Initializing Devtron Documentation API Server...")
    # Configuration from environment
    docs_repo_url = os.getenv("DOCS_REPO_URL", "https://github.com/devtron-labs/devtron-documentation")
    docs_path = os.getenv("DOCS_PATH", "./docs")
    embedding_model = EMBEDDING_MODEL
    chunk_size = int(os.getenv("CHUNK_SIZE", "1000"))
    chunk_overlap = int(os.getenv("CHUNK_OVERLAP", "0"))

    # PostgreSQL configuration
    db_host = os.getenv("POSTGRES_HOST", "localhost")
    db_port = int(os.getenv("POSTGRES_PORT", "5432"))
    db_name = os.getenv("POSTGRES_DB", "devtron_docs")
    db_user = os.getenv("POSTGRES_USER", "postgres")
    db_password = os.getenv("POSTGRES_PASSWORD", "postgres")

    logger.info("Starting Devtron Documentation RAG Server")

    # Initialize components
    logger.info("Initializing documentation processor...")
    doc_processor = DocumentationProcessor(
        docs_repo_url,
        docs_path,
        chunk_size=chunk_size,
        chunk_overlap=chunk_overlap
    )
    logger.info("Documentation processor initialized")
    logger.info("Initializing vector store with database connection...")
    vector_store = VectorStore(
        db_host=db_host,
        db_port=db_port,
        db_name=db_name,
        db_user=db_user,
        db_password=db_password,
        embedding_model=embedding_model
    )
    logger.info("Vector store initialized successfully")

    # An index that is missing or built for another model is rebuilt in the background.
    problem = vector_store.schema_problem()
    if problem:
        logger.warning(f"Index does not fit embedding model {embedding_model} ({problem}); starting a rebuild")
        _start_index(IndexMode.REBUILD)
    elif vector_store.needs_indexing():
        logger.info(f"Index is empty; starting a rebuild with embedding model {embedding_model}")
        _start_index(IndexMode.REBUILD)
    else:
        conn = vector_store.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT COUNT(*) FROM documents;")
                doc_count = cur.fetchone()[0]
                logger.info(f"Ready to serve queries ({doc_count} chunks indexed)")
        finally:
            vector_store.pool.putconn(conn)

    logger.info("Server startup complete")

    yield

    # Cleanup
    if vector_store:
        vector_store.close()
    logger.info("Server shutdown complete")


# Initialize FastAPI app
app = FastAPI(
    title="Devtron Documentation API",
    description="REST API for semantic search over Devtron documentation",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Request/Response Models
class SearchRequest(BaseModel):
    query: str = Field(..., description="Search query", min_length=1)
    max_results: int = Field(5, description="Maximum number of results", ge=1, le=20)


class SearchResult(BaseModel):
    title: str
    source: str
    header: str
    content: str
    score: float


class SearchResponse(BaseModel):
    query: str
    results: List[SearchResult]
    total_results: int


class IndexStatus(BaseModel):
    state: IndexState
    mode: Optional[IndexMode]
    started_at: Optional[str]
    finished_at: Optional[str]
    documents_total: int
    documents_done: int
    error: Optional[str]


class HealthResponse(BaseModel):
    status: ServiceStatus
    database: DatabaseStatus
    docs_indexed: bool
    indexing: IndexState


# API Endpoints
@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    try:
        needs_indexing = vector_store.needs_indexing()
        return HealthResponse(
            status=ServiceStatus.HEALTHY,
            database=DatabaseStatus.CONNECTED,
            docs_indexed=not needs_indexing,
            indexing=index_job["state"]
        )
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(status_code=503, detail=f"Service unhealthy: {str(e)}")


def _now() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds")


async def _run_index(mode: IndexMode) -> None:
    """Sync the docs and index them: everything for a rebuild, only what changed for an update."""
    try:
        await doc_processor.sync_docs()
        documents = await doc_processor.get_all_documents()
        if mode is IndexMode.REBUILD:
            vector_store.rebuild_schema()
            pending = documents
        else:
            indexed = vector_store.indexed_hashes()
            on_disk = {doc["source"] for doc in documents}
            vector_store.delete_sources([source for source in indexed if source not in on_disk])
            pending = [doc for doc in documents if indexed.get(doc["source"]) != doc["content_hash"]]

        index_job["documents_total"] = len(pending)
        logger.info(f"Index {mode.value}: {len(pending)} of {len(documents)} documents to embed")
        await vector_store.index_documents(
            pending, on_progress=lambda done: index_job.update(documents_done=done))
        index_job.update(state=IndexState.DONE, finished_at=_now())
        logger.info(f"Index {mode.value} complete: {len(pending)} documents embedded")
    except Exception as e:
        logger.error(f"Index {mode.value} failed: {e}", exc_info=True)
        index_job.update(state=IndexState.FAILED, finished_at=_now(), error=str(e))


def _start_index(mode: IndexMode) -> IndexStatus:
    """Start an indexing job in the background, refusing a second one."""
    global index_task
    if index_job["state"] is IndexState.RUNNING:
        raise HTTPException(status_code=409, detail=f"An index {index_job['mode'].value} is already running.")
    if mode is IndexMode.UPDATE:
        problem = vector_store.schema_problem()
        if problem:
            raise HTTPException(status_code=409,
                                detail=f"Index does not fit the embedding model ({problem}). Call /index/rebuild.")
    index_job.update(state=IndexState.RUNNING, mode=mode, started_at=_now(), finished_at=None,
                     documents_total=0, documents_done=0, error=None)
    index_task = asyncio.create_task(_run_index(mode))
    logger.info(f"Index {mode.value} started in the background")
    return IndexStatus(**index_job)


@app.post("/index/rebuild", response_model=IndexStatus, status_code=202)
async def rebuild_index():
    """Delete every indexed chunk and index all documentation again, in the background."""
    return _start_index(IndexMode.REBUILD)


@app.post("/index/update", response_model=IndexStatus, status_code=202)
async def update_index():
    """Index only new and changed documents and drop removed ones, in the background."""
    return _start_index(IndexMode.UPDATE)


@app.get("/index/status", response_model=IndexStatus)
async def index_status():
    """Progress of the current or last indexing job."""
    return IndexStatus(**index_job)


@app.post("/search", response_model=SearchResponse)
async def search_documentation(request: SearchRequest):
    """
    Search documentation using semantic search.

    Returns relevant documentation chunks based on vector similarity.
    """
    try:
        logger.info(f"Searching for: {request.query}")

        # Check if index exists
        if vector_store.needs_indexing():
            raise HTTPException(
                status_code=400,
                detail="Documentation not indexed. Please call /index/rebuild first."
            )

        # Perform vector search
        results = await vector_store.search(request.query, max_results=request.max_results)

        return SearchResponse(
            query=request.query,
            results=[SearchResult(**r) for r in results],
            total_results=len(results)
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Search failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")

    uvicorn.run(
        "api:app",
        host=host,
        port=port,
        reload=os.getenv("ENV", "production") == "development"
    )
