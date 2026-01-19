#!/bin/bash

# Start script for integrated Central API + RAG Server

set -e

echo "🚀 Starting Central API with integrated RAG Server..."
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose not found. Please install docker-compose."
    exit 1
fi

# Build and start services
echo "📦 Building Docker images..."
docker-compose build

echo ""
echo "🏃 Starting services..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check health
echo ""
echo "🏥 Checking service health..."

# Check Go server
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Central API (Go) is healthy"
else
    echo "❌ Central API (Go) is not responding"
fi

# Check Python RAG server (via proxy)
if curl -s http://localhost:8080/docs/health > /dev/null; then
    echo "✅ RAG Server (Python) is healthy"
else
    echo "❌ RAG Server (Python) is not responding"
fi

echo ""
echo "📊 Service Status:"
docker-compose ps

echo ""
echo "📝 Logs:"
echo "  - View all logs:        docker-compose logs -f"
echo "  - View Go logs:         docker-compose exec central-api tail -f /var/log/supervisor/central-api.out.log"
echo "  - View Python logs:     docker-compose exec central-api tail -f /var/log/supervisor/rag-server.out.log"
echo "  - View supervisor logs: docker-compose exec central-api tail -f /var/log/supervisor/supervisord.log"

echo ""
echo "🧪 Test Commands:"
echo "  # Health check"
echo "  curl http://localhost:8080/health"
echo ""
echo "  # RAG server health (via proxy)"
echo "  curl http://localhost:8080/docs/health"
echo ""
echo "  # Index documentation"
echo "  curl -X POST http://localhost:8080/docs/reindex -H 'Content-Type: application/json' -d '{\"force\": true}'"
echo ""
echo "  # Search documentation"
echo "  curl -X POST http://localhost:8080/docs/search -H 'Content-Type: application/json' -d '{\"query\": \"deployment\", \"max_results\": 3, \"use_llm\": false}'"

echo ""
echo "🎉 Services are running!"
echo "   Central API: http://localhost:8080"
echo "   RAG Endpoints: http://localhost:8080/docs/*"
echo ""
echo "To stop: docker-compose down"

