#!/bin/bash
# Startup script for RAG server
# Runs migrations and starts the API server

set -e

echo "========================================="
echo "Devtron Documentation RAG Server Startup"
echo "========================================="
echo ""

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if python3 -c "
import psycopg2
import os
try:
    conn = psycopg2.connect(
        host=os.getenv('POSTGRES_HOST', 'localhost'),
        port=int(os.getenv('POSTGRES_PORT', '5432')),
        database='postgres',
        user=os.getenv('POSTGRES_USER', 'postgres'),
        password=os.getenv('POSTGRES_PASSWORD', 'postgres')
    )
    conn.close()
    exit(0)
except:
    exit(1)
" 2>/dev/null; then
        echo "✓ PostgreSQL is ready"
        break
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "  Attempt $RETRY_COUNT/$MAX_RETRIES - PostgreSQL not ready yet, waiting..."
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "✗ PostgreSQL is not available after $MAX_RETRIES attempts"
    exit 1
fi

echo ""

# Enable pgvector extension
echo "🔧 Enabling pgvector extension..."
python3 -c "
import psycopg2
import os
import sys

try:
    conn = psycopg2.connect(
        host=os.getenv('POSTGRES_HOST', 'localhost'),
        port=int(os.getenv('POSTGRES_PORT', '5432')),
        database=os.getenv('POSTGRES_DB', 'devtron_docs'),
        user=os.getenv('POSTGRES_USER', 'postgres'),
        password=os.getenv('POSTGRES_PASSWORD', 'postgres')
    )
    conn.autocommit = True

    with conn.cursor() as cur:
        cur.execute('CREATE EXTENSION IF NOT EXISTS vector;')
        print('✓ pgvector extension enabled')

    conn.close()
    sys.exit(0)
except Exception as e:
    print(f'✗ Failed to enable pgvector extension: {e}')
    print('  Make sure you are using a PostgreSQL image with pgvector support')
    print('  Recommended: ankane/pgvector:v0.5.1 or pgvector/pgvector:pg16')
    sys.exit(1)
"

if [ $? -ne 0 ]; then
    echo "✗ pgvector extension setup failed"
    exit 1
fi

echo ""

# Run database migrations
echo "📦 Running database migrations..."
python3 run_migrations.py

if [ $? -ne 0 ]; then
    echo "✗ Database migrations failed"
    exit 1
fi

echo "✓ Database migrations completed"
echo ""

# Start the API server
echo "🚀 Starting API server..."
echo "   Host: ${HOST:-0.0.0.0}"
echo "   Port: ${PORT:-8000}"
echo "   Auto-index: ${AUTO_INDEX_ON_STARTUP:-true}"
echo ""

exec python3 api.py

