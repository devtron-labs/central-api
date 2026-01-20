"""
Vector Store using PostgreSQL pgvector and Local Embeddings (BAAI/bge-large-en-v1.5)
"""

import logging
import json
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
        self.model = SentenceTransformer(model_name)
        self.dimension = self.model.get_sentence_embedding_dimension()
        logger.info(f"Model loaded. Embedding dimension: {self.dimension}")

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
        embeddings = self.model.encode(texts_with_prefix, show_progress_bar=False)
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
        logger.info("Initializing Vector Store with PostgreSQL pgvector")
        logger.info(f"Database Configuration:")
        logger.info(f"  Host: {db_host}")
        logger.info(f"  Port: {db_port}")
        logger.info(f"  Database: {db_name}")
        logger.info(f"  User: {db_user}")
        logger.info(f"  Embedding Model: {embedding_model}")

        # Initialize connection pool
        try:
            logger.info("Creating database connection pool...")
            self.pool = SimpleConnectionPool(
                minconn=1,
                maxconn=10,
                host=db_host,
                port=db_port,
                database=db_name,
                user=db_user,
                password=db_password
            )
            logger.info("✓ Database connection pool created successfully")

            # Test connection
            logger.info("Testing database connection...")
            conn = self.pool.getconn()
            try:
                with conn.cursor() as cur:
                    cur.execute("SELECT version();")
                    version = cur.fetchone()[0]
                    logger.info(f"✓ Database connection successful!")
                    logger.info(f"  PostgreSQL version: {version}")
            finally:
                self.pool.putconn(conn)

        except psycopg2.OperationalError as e:
            logger.error("✗ Failed to connect to PostgreSQL database")
            logger.error(f"  Error: {str(e)}")
            logger.error(f"  Connection details: {db_user}@{db_host}:{db_port}/{db_name}")
            logger.error("  Possible issues:")
            logger.error("    - PostgreSQL server is not running")
            logger.error("    - Incorrect host or port")
            logger.error("    - Database does not exist")
            logger.error("    - Invalid credentials")
            logger.error("    - Network/firewall issues")
            raise
        except Exception as e:
            logger.error(f"✗ Unexpected error during database connection: {str(e)}")
            logger.error(f"  Error type: {type(e).__name__}")
            raise

        # Initialize local embeddings
        logger.info("Loading embedding model...")
        self.embeddings = LocalEmbeddings(model_name=embedding_model)
        self.embedding_dimension = self.embeddings.dimension
        logger.info(f"✓ Embedding model loaded (dimension: {self.embedding_dimension})")

        # Initialize database schema
        logger.info("Initializing database schema...")
        self._init_database()

        logger.info("✓ Vector store initialization complete!")

    def _init_database(self):
        """Initialize database schema with pgvector extension."""
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                # Enable pgvector extension
                try:
                    logger.info("Checking pgvector extension...")
                    cur.execute("CREATE EXTENSION IF NOT EXISTS vector;")
                    logger.info("✓ pgvector extension is available")
                except psycopg2.Error as e:
                    logger.error("✗ Failed to enable pgvector extension")
                    logger.error(f"  Error: {str(e)}")
                    logger.error("  Make sure you're using a PostgreSQL image with pgvector support")
                    logger.error("  Recommended: pgvector/pgvector:pg14 or ankane/pgvector:v0.5.1")
                    raise

                # Create documents table with dynamic embedding dimension
                try:
                    logger.info(f"Creating documents table (embedding dimension: {self.embedding_dimension})...")
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
                    logger.info("✓ Documents table ready")
                except psycopg2.Error as e:
                    logger.error("✗ Failed to create documents table")
                    logger.error(f"  Error: {str(e)}")
                    raise

                # Create index for vector similarity search
                try:
                    logger.info("Creating vector similarity index (IVFFlat)...")
                    cur.execute("""
                        CREATE INDEX IF NOT EXISTS documents_embedding_idx
                        ON documents USING ivfflat (embedding vector_cosine_ops)
                        WITH (lists = 100);
                    """)
                    logger.info("✓ Vector similarity index ready")
                except psycopg2.Error as e:
                    logger.error("✗ Failed to create vector index")
                    logger.error(f"  Error: {str(e)}")
                    raise

                # Create index for source lookups
                try:
                    logger.info("Creating source index...")
                    cur.execute("""
                        CREATE INDEX IF NOT EXISTS documents_source_idx
                        ON documents(source);
                    """)
                    logger.info("✓ Source index ready")
                except psycopg2.Error as e:
                    logger.error("✗ Failed to create source index")
                    logger.error(f"  Error: {str(e)}")
                    raise

                conn.commit()
                logger.info("✓ Database schema initialization complete")

                # Log table statistics
                cur.execute("SELECT COUNT(*) FROM documents;")
                doc_count = cur.fetchone()[0]
                logger.info(f"  Current document count: {doc_count}")

        except Exception as e:
            logger.error(f"✗ Database initialization failed: {str(e)}")
            raise
        finally:
            self.pool.putconn(conn)
    
    def needs_indexing(self) -> bool:
        """Check if the database needs initial indexing."""
        logger.info("Checking if database needs indexing...")
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT COUNT(*) FROM documents;")
                count = cur.fetchone()[0]

                if count == 0:
                    logger.info("✓ Database is empty - indexing needed")
                else:
                    logger.info(f"✓ Database already has {count} documents - indexing not needed")

                return count == 0
        except Exception as e:
            logger.error(f"✗ Failed to check document count: {str(e)}")
            raise
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
        
        logger.info(f"Indexing {len(documents)} documents...")
        
        # Process documents in batches
        batch_size = 10
        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            await self._index_batch(batch)
        
        logger.info("Indexing complete")

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

        # Generate embeddings
        logger.info(f"Generating embeddings for {len(rows)} chunks...")
        texts = [row['content'] for row in rows]
        embeddings = self.embeddings.embed_documents(texts)

        # Insert into database
        conn = self.pool.getconn()
        try:
            with conn.cursor() as cur:
                # Prepare data for batch insert
                values = [
                    (
                        row['id'],
                        row['title'],
                        row['source'],
                        row['header'],
                        row['content'],
                        row['chunk_index'],
                        embeddings[i]
                    )
                    for i, row in enumerate(rows)
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
                logger.info(f"Indexed batch of {len(rows)} chunks")
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
        logger.info(f"Searching for: '{query}' (max_results: {max_results})")

        try:
            # Generate query embedding
            logger.info("Generating query embedding...")
            query_embedding = self.embeddings.embed_query(query)
            logger.info(f"✓ Query embedding generated (dimension: {len(query_embedding)})")

            # Search in PostgreSQL using cosine similarity
            logger.info("Executing vector similarity search...")
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

                    logger.info(f"✓ Found {len(formatted_results)} results")
                    if formatted_results:
                        logger.info(f"  Top result: '{formatted_results[0]['title']}' (score: {formatted_results[0]['score']:.4f})")

                    return formatted_results
            finally:
                self.pool.putconn(conn)

        except Exception as e:
            logger.error(f"✗ Search failed: {str(e)}")
            logger.error(f"  Error type: {type(e).__name__}")
            raise

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

