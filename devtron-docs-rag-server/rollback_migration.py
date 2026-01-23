#!/usr/bin/env python3
"""
Database Migration Rollback Script
Rolls back the last applied migration using the corresponding .down.sql file.
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

    conn = psycopg2.connect(
        host=db_host,
        port=db_port,
        database=db_name,
        user=db_user,
        password=db_password
    )
    conn.set_isolation_level(ISOLATION_LEVEL_AUTOCOMMIT)
    return conn


def get_last_migration(conn):
    """Get the last applied migration."""
    try:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT version, description, applied_at 
                FROM schema_migrations 
                ORDER BY version DESC 
                LIMIT 1;
            """)
            result = cur.fetchone()
            if result:
                return {
                    'version': result[0],
                    'description': result[1],
                    'applied_at': result[2]
                }
            return None
    except psycopg2.Error as e:
        logger.error(f"Failed to get last migration: {e}")
        return None


def rollback_migration(version: str):
    """Rollback a specific migration version."""
    logger.info(f"Starting rollback of migration version {version}...")
    
    # Get migrations directory
    migrations_dir = Path(__file__).parent.parent / "scripts" / "sql"
    
    if not migrations_dir.exists():
        logger.error(f"Migrations directory not found: {migrations_dir}")
        return False
    
    # Find the down migration file
    down_file = migrations_dir / f"{version}_*.down.sql"
    down_files = list(migrations_dir.glob(f"{version}_*.down.sql"))
    
    if not down_files:
        logger.error(f"Down migration file not found for version {version}")
        return False
    
    down_file = down_files[0]
    logger.info(f"Found down migration: {down_file.name}")
    
    # Connect to database
    try:
        conn = get_db_connection()
        logger.info("Database connection established")
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        return False
    
    try:
        # Read and execute down migration
        with open(down_file, 'r') as f:
            sql = f.read()
        
        logger.info(f"Executing rollback: {down_file.name}")
        with conn.cursor() as cur:
            cur.execute(sql)
        
        # Remove migration record
        with conn.cursor() as cur:
            cur.execute(
                "DELETE FROM schema_migrations WHERE version = %s",
                (version,)
            )
        
        logger.info(f"✓ Migration {version} rolled back successfully")
        return True
        
    except Exception as e:
        logger.error(f"✗ Rollback failed: {e}")
        logger.error(f"   Error details: {str(e)}")
        return False
    finally:
        conn.close()
        logger.info("Database connection closed")


def main():
    """Main rollback function."""
    logger.info("Database Migration Rollback Tool")
    logger.info("=" * 50)
    
    # Connect to database
    try:
        conn = get_db_connection()
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        return False
    
    # Get last migration
    last_migration = get_last_migration(conn)
    conn.close()
    
    if not last_migration:
        logger.warning("No migrations to rollback")
        return True
    
    # Show migration info
    logger.info(f"Last applied migration:")
    logger.info(f"  Version: {last_migration['version']}")
    logger.info(f"  Description: {last_migration['description']}")
    logger.info(f"  Applied at: {last_migration['applied_at']}")
    logger.info("")
    
    # Confirm rollback
    if len(sys.argv) > 1 and sys.argv[1] == '--yes':
        confirm = 'yes'
    else:
        confirm = input("Do you want to rollback this migration? (yes/no): ").lower()
    
    if confirm != 'yes':
        logger.info("Rollback cancelled")
        return True
    
    # Perform rollback
    return rollback_migration(last_migration['version'])


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)

