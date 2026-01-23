# Quick Start Guide - Devtron Documentation RAG Server

## Prerequisites

- Docker and Docker Compose installed
- AWS credentials (for LLM features - optional)
- 4GB RAM minimum
- 10GB disk space

## Setup & Run

### 1. Clone and Configure

```bash
cd devtron-docs-rag-server
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` file:

```bash
# Required
POSTGRES_DB=devtron_docs
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password

# Optional - for LLM features
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key

# Optional - customize
EMBEDDING_MODEL=BAAI/bge-large-en-v1.5
CHUNK_SIZE=1000
LOG_LEVEL=INFO
```

### 3. Start Services

```bash
docker-compose up -d
```

Check logs:
```bash
docker-compose logs -f docs-api
```

### 4. Verify Health

```bash
curl http://localhost:8000/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "docs_indexed": false
}
```

### 5. Index Documentation

```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

This will:
- Clone Devtron documentation from GitHub
- Process markdown files
- Generate embeddings
- Store in PostgreSQL with pgvector

Expected response:
```json
{
  "status": "success",
  "message": "Full re-index completed",
  "documents_processed": 156,
  "changed_files": 12
}
```

⏱️ **Time**: Initial indexing takes 5-10 minutes depending on your hardware.

### 6. Search Documentation

**Simple search (no LLM):**
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to deploy applications?",
    "max_results": 3,
    "use_llm": false
  }'
```

**Enhanced search (with LLM):**
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to deploy applications?",
    "max_results": 5,
    "use_llm": true,
    "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
  }'
```

## Common Use Cases

### Daily Documentation Sync

Set up a cron job for incremental updates:

```bash
# Add to crontab (runs daily at 2 AM)
0 2 * * * curl -X POST http://localhost:8000/reindex -H "Content-Type: application/json" -d '{"force": false}'
```

### Integration with Chatbot

```python
import requests

def ask_devtron_docs(question: str) -> str:
    response = requests.post(
        "http://localhost:8000/search",
        json={
            "query": question,
            "max_results": 5,
            "use_llm": True
        }
    )
    data = response.json()
    return data.get("llm_response", "No answer found")

# Usage
answer = ask_devtron_docs("How do I configure RBAC?")
print(answer)
```

### Slack Bot Integration

```python
from slack_bolt import App
import requests

app = App(token="xoxb-your-token")

@app.message("!docs")
def handle_docs_query(message, say):
    query = message['text'].replace('!docs', '').strip()
    
    response = requests.post(
        "http://localhost:8000/search",
        json={"query": query, "max_results": 3, "use_llm": True}
    )
    
    result = response.json()
    say(result.get("llm_response", "No results found"))

app.start(port=3000)
```

## Troubleshooting

### Issue: "Documentation not indexed"
**Solution:** Run the reindex endpoint first:
```bash
curl -X POST http://localhost:8000/reindex -H "Content-Type: application/json" -d '{"force": true}'
```

### Issue: Database connection failed
**Solution:** Check PostgreSQL is running:
```bash
docker-compose ps
docker-compose logs postgres
```

### Issue: LLM responses not working
**Solution:** 
1. Check AWS credentials are set in `.env`
2. Verify AWS Bedrock access in your region
3. Search without LLM: `"use_llm": false`

### Issue: Slow search performance
**Solution:**
- Reduce `max_results` (default: 5)
- Disable LLM for faster responses
- Check database indexes are created

## Performance Tips

1. **Use incremental updates**: Set `"force": false` for daily syncs
2. **Limit results**: Use `max_results: 3-5` for best performance
3. **Cache responses**: Implement caching layer for common queries
4. **Disable LLM**: Use `"use_llm": false` when speed is critical

## Monitoring

View logs:
```bash
docker-compose logs -f docs-api
```

Check resource usage:
```bash
docker stats
```

## Stopping Services

```bash
docker-compose down
```

Keep data:
```bash
docker-compose down
```

Remove all data:
```bash
docker-compose down -v
```

## Next Steps

- See [API_EXAMPLES.md](./API_EXAMPLES.md) for detailed API documentation
- See [README.md](./README.md) for architecture details
- Configure production settings in `.env`
- Set up monitoring and alerting
- Implement rate limiting for production use

