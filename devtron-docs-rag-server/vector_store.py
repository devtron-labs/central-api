"""
Vector Store using PostgreSQL pgvector and local embeddings
"""

import logging
import json
import math
import os
import asyncio
from typing import List, Dict, Any, Callable, Optional
from pathlib import Path
import hashlib

import psycopg2
from psycopg2.extras import execute_values
from psycopg2.pool import SimpleConnectionPool
from sentence_transformers import SentenceTransformer

from settings import EMBEDDING_MODEL

logger = logging.getLogger(__name__)

# Chunks are about 1000 characters; longer token runs are cut instead of slowing every batch.
MAX_TOKENS = int(os.getenv("EMBEDDING_MAX_TOKENS", "256"))


def cpu_limit() -> int:
    """CPUs this container may use: its cgroup quota when one is set, else the host count."""
    for quota_file, period_file in (("/sys/fs/cgroup/cpu.max", None),
                                    ("/sys/fs/cgroup/cpu/cpu.cfs_quota_us", "/sys/fs/cgroup/cpu/cpu.cfs_period_us")):
        try:
            if period_file is None:
                quota, period = Path(quota_file).read_text().split()
            else:
                quota, period = Path(quota_file).read_text().strip(), Path(period_file).read_text().strip()
            if quota not in ("max", "-1"):
                return max(1, math.ceil(int(quota) / int(period)))
        except (OSError, ValueError):
            continue
    return os.cpu_count() or 1


class LocalEmbeddings:
    """Local embeddings using a sentence-transformers model."""

    def __init__(self, model_name: str = EMBEDDING_MODEL):
        """
        Initialize local embedding model.

        Args:
            model_name: HuggingFace model name
        """
        logger.info(f"Loading embedding model: {model_name}")

        # Verify cache directory exists
        cache_dir = os.getenv('SENTENCE_TRANSFORMERS_HOME')
        if cache_dir and os.path.exists(cache_dir):
            logger.info(f"Using cached model from: {cache_dir}")
        else:
            logger.warning(f"Cache directory not found: {cache_dir}")

        try:
            # Load model with optimizations for CPU inference
            import torch

            # Disable gradient computation (we're only doing inference)
            torch.set_grad_enabled(False)

            # Load model - it will use SENTENCE_TRANSFORMERS_HOME env var automatically
            self.model = SentenceTransformer(model_name)

            # Set model to evaluation mode for faster inference
            self.model.eval()

            # One thread per CPU the container is allowed, so a bigger limit is actually used.
            threads = int(os.getenv("EMBEDDING_THREADS", cpu_limit()))
            torch.set_num_threads(threads)
            self.model.max_seq_length = min(MAX_TOKENS, self.model.max_seq_length)
            logger.info(f"PyTorch threads: {threads}, max tokens per chunk: {self.model.max_seq_length}")

            self.dimension = self.model.get_sentence_embedding_dimension()
            logger.info(f"Embedding model loaded (dimension: {self.dimension})")
        except Exception as e:
            logger.error(f"Failed to load embedding model: {str(e)}")
            logger.error(f"Cache directory: {cache_dir}")
            logger.error(f"Cache exists: {os.path.exists(cache_dir) if cache_dir else 'N/A'}")
            raise

    def embed_documents(self, texts: List[str]) -> List[List[float]]:
        """
        Embed multiple documents.

        Args:
            texts: List of text strings to embed

        Returns:
            List of embedding vectors
        """
        # Add instruction prefix for better retrieval (recommended by BGE)
        texts_with_prefix = [f"passage: {text}" for text in texts]

        # Optimized settings for CPU inference
        # batch_size=16 is optimal for CPU (balances speed vs memory)
        # convert_to_tensor=False avoids unnecessary tensor conversions
        embeddings = self.model.encode(
            texts_with_prefix,
            show_progress_bar=False,
            batch_size=32,
            convert_to_numpy=True,
            normalize_embeddings=False,
            device='cpu'  # Explicitly use CPU
        )
        return embeddings.tolist()

    def embed_query(self, text: str) -> List[float]:
        """
        Embed a single query.

        Args:
            text: Text to embed

        Returns:
            Embedding vector
        """
        # Add instruction prefix for queries (recommended by BGE)
        text_with_prefix = f"query: {text}"
        embedding = self.model.encode(text_with_prefix, show_progress_bar=False)
        return embedding.tolist()


class VectorStore:
    """Vector store for documentation using PostgreSQL with pgvector."""

    def __init__(
        self,
        db_host: str = "localhost",
        db_port: int = 5432,
        db_name: str = "devtron_docs",
        db_user: str = "postgres",
        db_password: str = "postgres",
        embedding_model: str = EMBEDDING_MODEL
    ):
        """
        Initialize vector store.

        Args:
            db_host: PostgreSQL host
            db_port: PostgreSQL port
            db_name: Database name
            db_user: Database user
            db_password: Database password
            embedding_model: HuggingFace model name for embeddings
        """
        # Initialize connection pool
        try:
            logger.info(f"Connecting to database: {db_host}:{db_port}/{db_name}")
            self.pool = SimpleConnectionPool(
                minconn=1,
                maxconn=10,
                host=db_host,
                port=db_port,
                database=db_name,
                user=db_user,
                password=db_password
            )

            # Test connection
            conn = self.pool.getconn()
            try:
                with conn.cursor() as cur:
                    cur.execute("SELECT version();")
                    version = cur.fetchone()[0]
                    logger.info(f"Database connected successfully")

                    # Log connection details for debugging
                    cur.execute("SELECT current_database(), current_schema();")
                    db, schema = cur.fetchone()
                    logger.info(f"Connected to database: {db}, schema: {schema}")
            finally:
                self.pool.putconn(conn)

        except psycopg2.OperationalError as e:
            logger.error(f"Database connection failed: {str(e)}")
            logger.error(f"Connection: {db_user}@{db_host}:{db_port}/{db_name}")
            raise
        except Exception as e:
            logger.error(f"Unexpected error: {str(e)}")
            raise

        # Initialize local embeddings
        logger.info("Loading embedding model...")
        self.embeddings = LocalEmbeddings(model_name=embedding_model)
        self.embedding_dimension = self.embeddings.dimension

        # Initialize database schema
        logger.info("Initializing database schema...")
        self._init_database()
        logger.info("Vector store ready")

    def _init_database(self):
        """Initialize database schema with pgvector extension."""
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                # Enable pgvector extension
                cur.execute("CREATE EXTENSION IF NOT EXISTS vector;")

                # Create documents table with dynamic embedding dimension
                cur.execute(f"""
                    CREATE TABLE IF NOT EXISTS documents (
                        id TEXT PRIMARY KEY,
                        title TEXT NOT NULL,
                        source TEXT NOT NULL,
                        header TEXT,
                        content TEXT NOT NULL,
                        chunk_index INTEGER,
                        embedding vector({self.embedding_dimension}),
                        content_hash TEXT,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                    );
                """)

                # No vector index: a few thousand rows are scanned exactly in milliseconds,
                # and an ivfflat index with one probe would skip most of them.

                # Create index for source lookups
                cur.execute("""
                    CREATE INDEX IF NOT EXISTS documents_source_idx
                    ON documents(source);
                """)

                conn.commit()

                # Log table statistics
                cur.execute("SELECT COUNT(*) FROM documents;")
                doc_count = cur.fetchone()[0]
                logger.info(f"Schema initialized ({doc_count} documents indexed)")

        except psycopg2.Error as e:
            logger.error(f"Database schema initialization failed: {str(e)}")
            raise
        finally:
            self.pool.putconn(conn)
    
    def schema_problem(self) -> Optional[str]:
        """Why the table cannot hold this model's vectors, or None when it can."""
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT attname, atttypmod FROM pg_attribute
                    WHERE attrelid = 'documents'::regclass AND attname IN ('embedding', 'content_hash')
                      AND NOT attisdropped;
                """)
                columns = dict(cur.fetchall())
        finally:
            self.pool.putconn(conn)
        if columns.get("embedding") != self.embedding_dimension:
            return (f"table holds {columns.get('embedding')}-dimension vectors, "
                    f"the model produces {self.embedding_dimension}")
        if "content_hash" not in columns:
            return "table has no content_hash column"
        return None

    def rebuild_schema(self) -> None:
        """Empty the table and fit it to the current model."""
        logger.warning("Clearing the documents table for a full rebuild...")
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("TRUNCATE TABLE documents;")
                cur.execute("DROP INDEX IF EXISTS documents_embedding_idx;")
                cur.execute(f"ALTER TABLE documents ALTER COLUMN embedding TYPE vector({self.embedding_dimension});")
                cur.execute("ALTER TABLE documents ADD COLUMN IF NOT EXISTS content_hash TEXT;")
            conn.commit()
        finally:
            self.pool.putconn(conn)

    def indexed_hashes(self) -> Dict[str, Optional[str]]:
        """The content hash stored for each indexed source file."""
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT DISTINCT source, content_hash FROM documents;")
                return dict(cur.fetchall())
        finally:
            self.pool.putconn(conn)

    def delete_sources(self, sources: List[str]) -> None:
        """Remove every chunk of the given source files."""
        if not sources:
            return
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("DELETE FROM documents WHERE source = ANY(%s);", (sources,))
            conn.commit()
        finally:
            self.pool.putconn(conn)

    def needs_indexing(self) -> bool:
        """Check if the database needs initial indexing."""
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT COUNT(*) FROM documents;")
                count = cur.fetchone()[0]
                return count == 0
        finally:
            self.pool.putconn(conn)
    
    async def index_documents(self, documents: List[Dict[str, Any]],
                              on_progress: Optional[Callable[[int], None]] = None) -> None:
        """
        Replace the given documents in the vector store.

        Args:
            documents: List of document dictionaries
            on_progress: Called with the number of documents stored so far
        """
        if not documents:
            logger.info("No documents to index")
            return

        logger.info(f"Starting indexing: {len(documents)} documents")
        batch_size = 8
        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            await self._index_batch(batch)
            if on_progress:
                on_progress(i + len(batch))
            # Yield control to event loop to allow health checks and searches to respond
            await asyncio.sleep(0)

        logger.info(f"Indexing complete: {len(documents)} documents")

    async def _index_batch(self, documents: List[Dict[str, Any]]) -> None:
        """Embed one batch of documents and store it in a single transaction."""
        rows = []
        for doc in documents:
            chunks = doc.get('chunks') or [{'content': doc['content'], 'header': doc['title']}]
            for idx, chunk in enumerate(chunks):
                rows.append((f"{doc['id']}_chunk_{idx}", doc['title'], doc['source'],
                             chunk.get('header', ''), chunk['content'], idx, doc.get('content_hash')))

        # Run embedding in a thread so the event loop keeps serving requests
        loop = asyncio.get_event_loop()
        embeddings = await loop.run_in_executor(
            None, self.embeddings.embed_documents, [row[4] for row in rows])

        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                # Old chunks go in the same transaction, so a document is never half replaced
                cur.execute("DELETE FROM documents WHERE source = ANY(%s);",
                            ([doc['source'] for doc in documents],))
                execute_values(
                    cur,
                    """
                    INSERT INTO documents
                    (id, title, source, header, content, chunk_index, content_hash, embedding)
                    VALUES %s
                    """,
                    [row + (embedding,) for row, embedding in zip(rows, embeddings)]
                )
            conn.commit()
            logger.info(f"  Stored {len(rows)} chunks from {len(documents)} document(s)")
        except Exception as e:
            conn.rollback()
            logger.error(f"  Failed to store chunks: {str(e)}", exc_info=True)
            raise
        finally:
            self.pool.putconn(conn)

    async def search(self, query: str, max_results: int = 5) -> List[Dict[str, Any]]:
        """
        Search for relevant documents using vector similarity.

        Args:
            query: Search query
            max_results: Maximum number of results to return

        Returns:
            List of search results with metadata
        """
        # Generate query embedding
        query_embedding = self.embeddings.embed_query(query)

        # Search in PostgreSQL using cosine similarity
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT
                        id,
                        title,
                        source,
                        header,
                        content,
                        1 - (embedding <=> %s::vector) as similarity
                    FROM documents
                    ORDER BY embedding <=> %s::vector
                    LIMIT %s
                    """,
                    (query_embedding, query_embedding, max_results)
                )

                results = cur.fetchall()

                # Format results
                formatted_results = []
                for row in results:
                    formatted_results.append({
                        'id': row[0],
                        'title': row[1],
                        'source': row[2],
                        'header': row[3] or '',
                        'content': row[4],
                        'score': float(row[5])
                    })

                logger.info(f"Search: '{query}' -> {len(formatted_results)} results")
                return formatted_results
        finally:
            self.pool.putconn(conn)

    def close(self) -> None:
        """Close all database connections."""
        if self.pool:
            self.pool.closeall()
            logger.info("Database connections closed")

