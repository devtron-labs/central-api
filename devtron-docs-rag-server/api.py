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
import boto3
from botocore.config import Config

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
bedrock_runtime = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global doc_processor, vector_store, bedrock_runtime

    logger.info("Initializing Devtron Documentation API Server...")

    # Configuration from environment
    docs_repo_url = os.getenv("DOCS_REPO_URL", "https://github.com/devtron-labs/devtron")
    docs_path = os.getenv("DOCS_PATH", "./devtron-docs")
    aws_region = os.getenv("AWS_REGION", "us-east-1")

    # Embedding model configuration
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
    logger.info("✓ Documentation processor initialized")

    logger.info("Initializing vector store with database connection...")
    try:
        vector_store = VectorStore(
            db_host=db_host,
            db_port=db_port,
            db_name=db_name,
            db_user=db_user,
            db_password=db_password,
            embedding_model=embedding_model
        )
        logger.info("✓ Vector store initialized successfully")
    except Exception as e:
        logger.error("✗ FATAL: Failed to initialize vector store")
        logger.error(f"Error: {str(e)}")
        logger.error(f"Database: {db_user}@{db_host}:{db_port}/{db_name}")
        logger.error("")
        logger.error("Troubleshooting steps:")
        logger.error("1. Check if PostgreSQL container is running:")
        logger.error("   docker-compose ps postgres-pgvector")
        logger.error("")
        logger.error("2. Check PostgreSQL logs:")
        logger.error("   docker-compose logs postgres-pgvector")
        logger.error("")
        logger.error("3. Verify connection details in docker-compose.yml")
        logger.error("")
        logger.error("4. Ensure you're using a pgvector-enabled PostgreSQL image:")
        logger.error("   pgvector/pgvector:pg14 or ankane/pgvector:v0.5.1")
        raise

    # Initialize Bedrock runtime for LLM (optional - only for enhanced responses)
    try:
        bedrock_runtime = boto3.client(
            service_name='bedrock-runtime',
            region_name=aws_region,
            config=Config(read_timeout=300)
        )
        logger.info("AWS Bedrock initialized for LLM responses")
    except Exception as e:
        logger.warning(f"AWS Bedrock not available: {e}. LLM responses will be disabled.")
        bedrock_runtime = None

    # Check if database needs indexing
    if vector_store.needs_indexing():
        logger.warning("⚠️  Database is empty - no documents indexed")
        logger.warning("   Call POST /docs/index to index documentation")
    else:
        logger.info("✓ Database already has indexed documents")

    logger.info("Server initialization complete")

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
    use_llm: bool = Field(
        False,
        description="Whether to use LLM for enhanced response. "
                    "Recommended: false for MCP tools (let caller handle LLM to avoid double token usage)"
    )
    llm_model: str = Field(
        "anthropic.claude-3-haiku-20240307-v1:0",
        description="Bedrock model ID (only used if use_llm=true)"
    )


class SearchResult(BaseModel):
    title: str
    source: str
    header: str
    content: str
    score: float


class SearchResponse(BaseModel):
    query: str
    results: List[SearchResult]
    llm_response: Optional[str] = None
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

    Optionally uses LLM to generate an enhanced response based on search results.
    """
    try:
        logger.info(f"Searching for: {request.query}")

        # Check if index exists
        if vector_store.needs_indexing():
            raise HTTPException(
                status_code=400,
                detail="Documentation not indexed. Please call /reindex first."
            )

        # Perform vector search
        results = await vector_store.search(request.query, max_results=request.max_results)

        llm_response = None
        if request.use_llm and results:
            if bedrock_runtime is None:
                logger.warning("LLM requested but AWS Bedrock not available")
                llm_response = "LLM responses are not available. AWS Bedrock is not configured."
            else:
                # Generate LLM response using search results as context
                llm_response = await generate_llm_response(
                    query=request.query,
                    search_results=results,
                    model_id=request.llm_model
                )

        return SearchResponse(
            query=request.query,
            results=[SearchResult(**r) for r in results],
            llm_response=llm_response,
            total_results=len(results)
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Search failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")


async def generate_llm_response(query: str, search_results: List[dict], model_id: str) -> str:
    """
    Generate LLM response using search results as context.

    Args:
        query: User's search query
        search_results: List of search results from vector store
        model_id: Bedrock model ID to use

    Returns:
        LLM-generated response
    """
    try:
        # Build context from search results
        context_parts = []
        for i, result in enumerate(search_results, 1):
            context_parts.append(
                f"[Document {i}]\n"
                f"Title: {result['title']}\n"
                f"Source: {result['source']}\n"
                f"Content:\n{result['content']}\n"
            )

        context = "\n---\n".join(context_parts)

        # Build prompt
        prompt = f"""You are a helpful assistant for Devtron documentation. Answer the user's question based on the provided documentation context.

Documentation Context:
{context}

User Question: {query}

Instructions:
- Answer based ONLY on the provided documentation context
- Be concise and accurate
- If the context doesn't contain enough information, say so
- Include relevant code examples or commands if present in the context
- Format your response in markdown

Answer:"""

        # Call Bedrock
        if "claude" in model_id.lower():
            # Claude models
            body = {
                "anthropic_version": "bedrock-2023-05-31",
                "max_tokens": 2000,
                "messages": [
                    {
                        "role": "user",
                        "content": prompt
                    }
                ],
                "temperature": 0.7
            }

            response = bedrock_runtime.invoke_model(
                modelId=model_id,
                body=str.encode(str(body))
            )

            import json
            response_body = json.loads(response['body'].read())
            return response_body['content'][0]['text']

        else:
            # Other models (Titan, etc.)
            body = {
                "inputText": prompt,
                "textGenerationConfig": {
                    "maxTokenCount": 2000,
                    "temperature": 0.7,
                    "topP": 0.9
                }
            }

            response = bedrock_runtime.invoke_model(
                modelId=model_id,
                body=str.encode(str(body))
            )

            import json
            response_body = json.loads(response['body'].read())
            return response_body['results'][0]['outputText']

    except Exception as e:
        logger.error(f"LLM generation failed: {e}", exc_info=True)
        return f"Error generating LLM response: {str(e)}"


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
