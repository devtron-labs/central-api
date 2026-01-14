#!/bin/bash
# Quick start script for Devtron Documentation API

set -e

echo "🚀 Starting Devtron Documentation API..."
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo ""
    echo "⚠️  IMPORTANT: Please edit .env file with your AWS credentials!"
    echo ""
    echo "Required configuration:"
    echo "  - AWS_ACCESS_KEY_ID"
    echo "  - AWS_SECRET_ACCESS_KEY"
    echo "  - AWS_REGION"
    echo ""
    read -p "Press Enter after you've configured .env, or Ctrl+C to exit..."
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Check if Docker is available
if command -v docker &> /dev/null && command -v docker-compose &> /dev/null; then
    echo "🐳 Docker detected. Starting with Docker Compose..."
    echo ""
    
    # Start services
    docker-compose up -d
    
    echo ""
    echo "✅ Services started!"
    echo ""
    echo "📊 Service Status:"
    docker-compose ps
    
    echo ""
    echo "⏳ Waiting for services to be ready..."
    sleep 5
    
    # Check health
    echo ""
    echo "🔍 Checking API health..."
    if curl -s http://localhost:8000/health > /dev/null 2>&1; then
        echo "✅ API is healthy!"
    else
        echo "⚠️  API not responding yet. Check logs with: docker-compose logs -f docs-api"
    fi
    
    echo ""
    echo "📚 Next steps:"
    echo "  1. Index documentation: curl -X POST http://localhost:8000/reindex -H 'Content-Type: application/json' -d '{\"force\": true}'"
    echo "  2. Test search: python test_api.py"
    echo "  3. View API docs: http://localhost:8000/docs"
    echo "  4. View logs: docker-compose logs -f docs-api"
    echo ""
    
else
    echo "🐍 Docker not found. Starting locally..."
    echo ""
    
    # Check if virtual environment exists
    if [ ! -d "venv" ]; then
        echo "📦 Creating virtual environment..."
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    echo "🔧 Activating virtual environment..."
    source venv/bin/activate
    
    # Install dependencies
    echo "📥 Installing dependencies..."
    pip install -q --upgrade pip
    pip install -q -r requirements.txt
    
    # Check PostgreSQL
    echo ""
    echo "🗄️  Checking PostgreSQL..."
    POSTGRES_HOST=${POSTGRES_HOST:-localhost}
    POSTGRES_PORT=${POSTGRES_PORT:-5432}
    POSTGRES_USER=${POSTGRES_USER:-postgres}
    POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
    
    if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -lqt &>/dev/null; then
        echo "❌ PostgreSQL not running!"
        echo ""
        echo "Please start PostgreSQL:"
        echo "  Option 1: docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres ankane/pgvector:latest"
        echo "  Option 2: brew services start postgresql@15"
        echo "  Option 3: sudo systemctl start postgresql"
        echo ""
        exit 1
    fi
    
    echo "✅ PostgreSQL is running"
    
    # Setup database
    echo ""
    echo "🔧 Setting up database..."
    ./setup_database.sh
    
    # Start API server
    echo ""
    echo "🚀 Starting API server..."
    echo ""
    python api.py &
    API_PID=$!
    
    # Wait for API to start
    echo "⏳ Waiting for API to start..."
    sleep 3
    
    # Check health
    if curl -s http://localhost:8000/health > /dev/null 2>&1; then
        echo "✅ API is running!"
        echo ""
        echo "📚 Next steps:"
        echo "  1. Index documentation: curl -X POST http://localhost:8000/reindex -H 'Content-Type: application/json' -d '{\"force\": true}'"
        echo "  2. Test search: python test_api.py"
        echo "  3. View API docs: http://localhost:8000/docs"
        echo ""
        echo "To stop the server: kill $API_PID"
        echo ""
        
        # Keep script running
        wait $API_PID
    else
        echo "❌ API failed to start. Check the logs above."
        kill $API_PID 2>/dev/null || true
        exit 1
    fi
fi

