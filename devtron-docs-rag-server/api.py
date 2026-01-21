#!/usr/bin/env python3
"""
Devtron Documentation API Server
REST API for documentation search and re-indexing using PostgreSQL pgvector and local embeddings.
"""

import asyncio
import logging
import os
from typing import List, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from doc_processor import DocumentationProcessor
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


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global doc_processor, vector_store
    logger.info("Initializing Devtron Documentation API Server...")
    # Configuration from environment
    docs_repo_url = os.getenv("DOCS_REPO_URL", "https://github.com/devtron-labs/devtron")
    docs_path = os.getenv("DOCS_PATH", "./devtron-docs")
    embedding_model = os.getenv("EMBEDDING_MODEL", "BAAI/bge-large-en-v1.5")
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

    # Check if database needs indexing
    if vector_store.needs_indexing():
        logger.info("⚠️  Database is empty - call POST /docs/index to index documentation")
    else:
        conn = vector_store.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT COUNT(*) FROM documents;")
                doc_count = cur.fetchone()[0]
                logger.info(f"✓ Ready to serve queries ({doc_count} chunks indexed)")
        finally:
            vector_store.pool.putconn(conn)

    logger.info("✓ Server startup complete")

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


class IndexRequest(BaseModel):
    force: bool = Field(False, description="Force full re-index even if documents already exist")


class IndexResponse(BaseModel):
    status: str
    message: str
    documents_indexed: int
    total_chunks: int


class HealthResponse(BaseModel):
    status: str
    database: str
    docs_indexed: bool


# API Endpoints
@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    try:
        needs_indexing = vector_store.needs_indexing()
        return HealthResponse(
            status="healthy",
            database="connected",
            docs_indexed=not needs_indexing
        )
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(status_code=503, detail=f"Service unhealthy: {str(e)}")


@app.post("/index", response_model=IndexResponse)
async def index_documentation(request: IndexRequest):
    """
    Index documentation from GitHub into the vector database.

    This endpoint:
    1. Syncs the latest documentation from GitHub
    2. Processes all markdown files
    3. Generates embeddings
    4. Stores vectors in PostgreSQL with pgvector

    If documents already exist and force=false, it will skip indexing.
    If force=true, it will clear existing data and re-index everything.
    """
    try:
        # Check if already indexed
        if not request.force and not vector_store.needs_indexing():
            logger.info("Documentation already indexed. Use force=true to re-index.")
            # Get current count
            conn = vector_store.pool.getconn()
            try:
                with conn.cursor() as cur:
                    cur.execute("SELECT COUNT(*) FROM documents;")
                    doc_count = cur.fetchone()[0]
                    cur.execute("SELECT COUNT(DISTINCT source) FROM documents;")
                    source_count = cur.fetchone()[0]
            finally:
                vector_store.pool.putconn(conn)

            return IndexResponse(
                status="skipped",
                message=f"Documentation already indexed ({source_count} documents, {doc_count} chunks). Use force=true to re-index.",
                documents_indexed=source_count,
                total_chunks=doc_count
            )

        # If force=true, reset the database
        if request.force and not vector_store.needs_indexing():
            logger.info("Force re-index requested. Clearing existing data...")
            vector_store.reset()
            logger.info("✓ Existing data cleared")

        logger.info("Starting documentation indexing...")

        # Sync docs from GitHub
        logger.info("Syncing documentation from GitHub...")
        changed_files = await doc_processor.sync_docs()
        logger.info(f"✓ Synced documentation: {len(changed_files)} files")

        # Get all documents
        logger.info("Processing documentation files...")
        documents = await doc_processor.get_all_documents()
        logger.info(f"✓ Found {len(documents)} documents to process")

        if not documents:
            logger.warning("No documents found to index")
            return IndexResponse(
                status="error",
                message="No documents found in repository",
                documents_indexed=0,
                total_chunks=0
            )

        # Index documents (this will chunk them and create embeddings)
        logger.info("Generating embeddings and indexing into database...")
        await vector_store.index_documents(documents)

        # Get final counts
        conn = vector_store.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT COUNT(*) FROM documents;")
                total_chunks = cur.fetchone()[0]
        finally:
            vector_store.pool.putconn(conn)

        logger.info(f"✓ Indexing complete: {len(documents)} documents, {total_chunks} chunks")

        return IndexResponse(
            status="success",
            message=f"Successfully indexed {len(documents)} documents into {total_chunks} chunks",
            documents_indexed=len(documents),
            total_chunks=total_chunks
        )

    except Exception as e:
        logger.error(f"Indexing failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Indexing failed: {str(e)}")


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
                detail="Documentation not indexed. Please call /index first."
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
