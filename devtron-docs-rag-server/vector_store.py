"""
Vector Store using PostgreSQL pgvector and Local Embeddings (BAAI/bge-large-en-v1.5)
"""

import logging
import json
import os
import asyncio
from typing import List, Dict, Any, Optional
from pathlib import Path
import hashlib

import psycopg2
from psycopg2.extras import execute_values
from psycopg2.pool import SimpleConnectionPool
from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)


class LocalEmbeddings:
    """Local embeddings using BAAI/bge-large-en-v1.5 model."""

    def __init__(self, model_name: str = "BAAI/bge-large-en-v1.5"):
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
            # Load model - it will use SENTENCE_TRANSFORMERS_HOME env var automatically
            self.model = SentenceTransformer(model_name)
            self.dimension = self.model.get_sentence_embedding_dimension()
            logger.info(f"✓ Embedding model loaded (dimension: {self.dimension})")
        except Exception as e:
            logger.error(f"✗ Failed to load embedding model: {str(e)}")
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

        # Use very small batch size for CPU to minimize blocking time
        # batch_size=2 processes 2 texts at a time, reducing memory and blocking
        embeddings = self.model.encode(
            texts_with_prefix,
            show_progress_bar=False,
            batch_size=2,
            convert_to_numpy=True,
            normalize_embeddings=False
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
        embedding_model: str = "BAAI/bge-large-en-v1.5"
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
                    logger.info(f"✓ Database connected successfully")
            finally:
                self.pool.putconn(conn)

        except psycopg2.OperationalError as e:
            logger.error(f"✗ Database connection failed: {str(e)}")
            logger.error(f"Connection: {db_user}@{db_host}:{db_port}/{db_name}")
            raise
        except Exception as e:
            logger.error(f"✗ Unexpected error: {str(e)}")
            raise

        # Initialize local embeddings
        logger.info("Loading embedding model...")
        self.embeddings = LocalEmbeddings(model_name=embedding_model)
        self.embedding_dimension = self.embeddings.dimension

        # Initialize database schema
        logger.info("Initializing database schema...")
        self._init_database()
        logger.info("✓ Vector store ready")

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
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                    );
                """)

                # Create index for vector similarity search
                cur.execute("""
                    CREATE INDEX IF NOT EXISTS documents_embedding_idx
                    ON documents USING ivfflat (embedding vector_cosine_ops)
                    WITH (lists = 100);
                """)

                # Create index for source lookups
                cur.execute("""
                    CREATE INDEX IF NOT EXISTS documents_source_idx
                    ON documents(source);
                """)

                conn.commit()

                # Log table statistics
                cur.execute("SELECT COUNT(*) FROM documents;")
                doc_count = cur.fetchone()[0]
                logger.info(f"✓ Schema initialized ({doc_count} documents indexed)")

        except psycopg2.Error as e:
            logger.error(f"✗ Database schema initialization failed: {str(e)}")
            raise
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
    
    async def index_documents(self, documents: List[Dict[str, Any]]) -> None:
        """
        Index documents into the vector store.

        Args:
            documents: List of document dictionaries
        """
        if not documents:
            logger.warning("No documents to index")
            return

        logger.info(f"Starting indexing: {len(documents)} documents")

        # Process documents one at a time to minimize memory and allow health checks
        batch_size = 1
        total_batches = len(documents)

        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            batch_num = i + 1
            logger.info(f"Processing document {batch_num}/{total_batches}: {batch[0].get('title', 'Unknown')}")
            await self._index_batch(batch)

            # Yield control to event loop to allow health checks to respond
            await asyncio.sleep(0.1)

        logger.info(f"✓ Indexing complete: {len(documents)} documents")

    async def _index_batch(self, documents: List[Dict[str, Any]]) -> None:
        """Index a batch of documents."""
        rows = []

        for doc in documents:
            # Index each chunk separately for better retrieval
            chunks = doc.get('chunks', [])

            if not chunks:
                # If no chunks, index the whole document
                chunks = [{
                    'content': doc['content'],
                    'header': doc['title'],
                    'source': doc['source']
                }]

            for idx, chunk in enumerate(chunks):
                chunk_id = f"{doc['id']}_chunk_{idx}"
                rows.append({
                    'id': chunk_id,
                    'title': doc['title'],
                    'source': doc['source'],
                    'header': chunk.get('header', ''),
                    'content': chunk['content'],
                    'chunk_index': idx
                })

        logger.info(f"Processing {len(rows)} chunks from {len(documents)} document(s)")

        # Process chunks in very small sub-batches to avoid blocking health checks
        # Reduced to 5 chunks at a time (~10-15 seconds per sub-batch)
        chunk_batch_size = 5
        total_chunks = len(rows)

        conn = self.pool.getconn()
        try:
            for chunk_start in range(0, total_chunks, chunk_batch_size):
                chunk_end = min(chunk_start + chunk_batch_size, total_chunks)
                chunk_batch = rows[chunk_start:chunk_end]

                # Generate embeddings for this sub-batch
                logger.info(f"  Embedding chunks {chunk_start+1}-{chunk_end}/{total_chunks}...")
                texts = [row['content'] for row in chunk_batch]

                # Run embedding in thread pool to avoid blocking event loop
                loop = asyncio.get_event_loop()
                embeddings = await loop.run_in_executor(
                    None,
                    self.embeddings.embed_documents,
                    texts
                )

                # Insert into database
                with conn.cursor() as cur:
                    # Prepare data for batch insert
                    values = [
                        (
                            chunk_batch[i]['id'],
                            chunk_batch[i]['title'],
                            chunk_batch[i]['source'],
                            chunk_batch[i]['header'],
                            chunk_batch[i]['content'],
                            chunk_batch[i]['chunk_index'],
                            embeddings[i]
                        )
                        for i in range(len(chunk_batch))
                    ]

                    # Batch insert
                    execute_values(
                        cur,
                        """
                        INSERT INTO documents
                        (id, title, source, header, content, chunk_index, embedding)
                        VALUES %s
                        ON CONFLICT (id) DO UPDATE SET
                            title = EXCLUDED.title,
                            source = EXCLUDED.source,
                            header = EXCLUDED.header,
                            content = EXCLUDED.content,
                            chunk_index = EXCLUDED.chunk_index,
                            embedding = EXCLUDED.embedding,
                            updated_at = CURRENT_TIMESTAMP
                        """,
                        values
                    )
                    conn.commit()
                    logger.info(f"  ✓ Stored {len(chunk_batch)} chunks")

                # Yield control to event loop to allow health checks
                await asyncio.sleep(0.1)

            logger.info(f"✓ Document complete: {total_chunks} chunks indexed")
        finally:
            self.pool.putconn(conn)

    async def update_documents(self, documents: List[Dict[str, Any]]) -> None:
        """
        Update specific documents in the vector store.

        Args:
            documents: List of document dictionaries to update
        """
        if not documents:
            return

        logger.info(f"Updating {len(documents)} documents...")

        # Delete old versions
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                for doc in documents:
                    cur.execute(
                        "DELETE FROM documents WHERE source = %s",
                        (doc['source'],)
                    )
                conn.commit()
        finally:
            self.pool.putconn(conn)

        # Re-index the documents
        await self.index_documents(documents)

        logger.info("Update complete")

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

    def reset(self) -> None:
        """Reset the vector store (delete all data)."""
        logger.warning("Resetting vector store...")
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("TRUNCATE TABLE documents;")
                conn.commit()
                logger.info("Vector store reset complete")
        finally:
            self.pool.putconn(conn)

    def close(self) -> None:
        """Close all database connections."""
        if self.pool:
            self.pool.closeall()
            logger.info("Database connections closed")

