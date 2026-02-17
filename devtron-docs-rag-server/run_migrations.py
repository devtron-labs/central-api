#!/usr/bin/env python3
"""
Database Migration Runner
Runs SQL migrations in order to set up the database schema.
Supports up/down migrations from scripts/sql/ directory.
"""

import os
import sys
import logging
from pathlib import Path
import psycopg2
from psycopg2.extensions import ISOLATION_LEVEL_AUTOCOMMIT

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def get_db_connection():
    """Create database connection."""
    db_host = os.getenv("POSTGRES_HOST", "localhost")
    db_port = int(os.getenv("POSTGRES_PORT", "5432"))
    db_name = os.getenv("POSTGRES_DB", "devtron_docs")
    db_user = os.getenv("POSTGRES_USER", "postgres")
    db_password = os.getenv("POSTGRES_PASSWORD", "postgres")

    try:
        conn = psycopg2.connect(
            host=db_host,
            port=db_port,
            database=db_name,
            user=db_user,
            password=db_password
        )
        conn.set_isolation_level(ISOLATION_LEVEL_AUTOCOMMIT)
        return conn
    except psycopg2.OperationalError as e:
        logger.error(f"Failed to connect to database: {e}")
        logger.info("Attempting to create database...")
        
        # Try to connect to default 'postgres' database to create our database
        try:
            conn = psycopg2.connect(
                host=db_host,
                port=db_port,
                database="postgres",
                user=db_user,
                password=db_password
            )
            conn.set_isolation_level(ISOLATION_LEVEL_AUTOCOMMIT)
            
            with conn.cursor() as cur:
                cur.execute(f"CREATE DATABASE {db_name};")
                logger.info(f"Database '{db_name}' created successfully")
            
            conn.close()
            
            # Now connect to the newly created database
            return psycopg2.connect(
                host=db_host,
                port=db_port,
                database=db_name,
                user=db_user,
                password=db_password
            )
        except Exception as create_error:
            logger.error(f"Failed to create database: {create_error}")
            raise


def get_applied_migrations(conn):
    """Get list of already applied migrations."""
    try:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT version FROM schema_migrations ORDER BY version;
            """)
            return {row[0] for row in cur.fetchall()}
    except psycopg2.Error:
        # Table doesn't exist yet, no migrations applied
        return set()


def run_migration(conn, migration_file: Path):
    """Run a single migration file."""
    logger.info(f"Running migration: {migration_file.name}")

    try:
        with open(migration_file, 'r') as f:
            sql = f.read()

        with conn.cursor() as cur:
            cur.execute(sql)

        logger.info(f"✓ Migration {migration_file.name} completed successfully")
        return True
    except Exception as e:
        logger.error(f"✗ Migration {migration_file.name} failed: {e}")
        logger.error(f"   Error details: {str(e)}")
        return False


def run_migrations():
    """Run all pending migrations from scripts/sql/ directory."""
    logger.info("Starting database migrations...")

    # Get migrations directory - use root scripts/sql/ directory
    # Path: devtron-docs-rag-server/run_migrations.py -> ../scripts/sql/
    migrations_dir = Path(__file__).parent.parent / "scripts" / "sql"

    if not migrations_dir.exists():
        logger.error(f"Migrations directory not found: {migrations_dir}")
        return False

    # Get all UP migration files (e.g., 1_release_notes.up.sql, 2_pgvector_docs.up.sql)
    migration_files = sorted(migrations_dir.glob("*.up.sql"))

    if not migration_files:
        logger.warning("No migration files found")
        return True

    logger.info(f"Found {len(migration_files)} migration file(s)")

    # Connect to database
    try:
        conn = get_db_connection()
        logger.info("Database connection established")
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        return False

    try:
        # Get already applied migrations
        applied = get_applied_migrations(conn)
        logger.info(f"Already applied migrations: {len(applied)}")

        # Run pending migrations
        pending_count = 0
        for migration_file in migration_files:
            # Extract version from filename (e.g., "2_pgvector_docs.up.sql" -> "2")
            version = migration_file.stem.split('_')[0]

            if version in applied:
                logger.info(f"⊘ Skipping already applied migration: {migration_file.name}")
                continue

            pending_count += 1
            if not run_migration(conn, migration_file):
                logger.error("Migration failed, stopping")
                return False

        if pending_count == 0:
            logger.info("✓ All migrations are up to date")
        else:
            logger.info(f"✓ Successfully applied {pending_count} migration(s)")

        return True

    finally:
        conn.close()
        logger.info("Database connection closed")


if __name__ == "__main__":
    success = run_migrations()
    sys.exit(0 if success else 1)

