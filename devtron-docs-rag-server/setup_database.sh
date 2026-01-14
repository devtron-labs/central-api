#!/bin/bash
# Database setup script for Devtron MCP Documentation Server

set -e

echo "🗄️  Setting up PostgreSQL database for Devtron MCP Server..."

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-devtron_docs}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}

echo "📋 Configuration:"
echo "   Host: $POSTGRES_HOST"
echo "   Port: $POSTGRES_PORT"
echo "   Database: $POSTGRES_DB"
echo "   User: $POSTGRES_USER"

# Check if PostgreSQL is running
echo ""
echo "🔍 Checking PostgreSQL connection..."
if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -lqt &>/dev/null; then
    echo "❌ Cannot connect to PostgreSQL at $POSTGRES_HOST:$POSTGRES_PORT"
    echo ""
    echo "Please ensure PostgreSQL is running. You can:"
    echo "  1. Install PostgreSQL locally: https://www.postgresql.org/download/"
    echo "  2. Use Docker: docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres ankane/pgvector:latest"
    echo "  3. Use docker-compose: docker-compose up -d postgres"
    exit 1
fi

echo "✅ PostgreSQL is running"

# Create database if it doesn't exist
echo ""
echo "📦 Creating database '$POSTGRES_DB' if it doesn't exist..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$POSTGRES_DB'" | grep -q 1 || \
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -c "CREATE DATABASE $POSTGRES_DB"

echo "✅ Database '$POSTGRES_DB' is ready"

# Enable pgvector extension
echo ""
echo "🔧 Enabling pgvector extension..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "CREATE EXTENSION IF NOT EXISTS vector;"

echo "✅ pgvector extension enabled"

# Create tables (will be created by the application, but we can verify)
echo ""
echo "📊 Database setup complete!"
echo ""
echo "You can now run the MCP server with:"
echo "  python server.py"
echo ""
echo "Or run tests with:"
echo "  python test_server.py"

