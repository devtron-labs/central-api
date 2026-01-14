# Devtron Documentation API

REST API for semantic search over Devtron documentation with LLM-enhanced responses.

## 🚀 Quick Start

### Start the API Server

```bash
# Using Docker (Recommended)
docker-compose up -d

# Or locally
python api.py
```

The API will be available at `http://localhost:8000`

### Interactive API Documentation

Once running, visit:
- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc

## 📡 API Endpoints

### 1. Health Check

Check if the API is running and database is connected.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "database": "connected",
  "docs_indexed": true
}
```

**Example**:
```bash
curl http://localhost:8000/health
```

---

### 2. Re-index Documentation

Sync and re-index documentation from GitHub.

**Endpoint**: `POST /reindex`

**Request Body**:
```json
{
  "force": false
}
```

**Parameters**:
- `force` (boolean, optional): Force full re-index even if no changes detected. Default: `false`

**Response**:
```json
{
  "status": "success",
  "message": "Incremental update completed",
  "documents_processed": 15,
  "changed_files": 3
}
```

**Example**:
```bash
# Incremental update (only changed files)
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": false}'

# Force full re-index
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

**Use Cases**:
- Call this endpoint periodically (e.g., daily) to keep docs up-to-date
- Call with `force: true` after major documentation changes
- Call on first deployment to initialize the index

---

### 3. Search Documentation

Search documentation using semantic search with optional LLM-enhanced response.

**Endpoint**: `POST /search`

**Request Body**:
```json
{
  "query": "How do I deploy an application?",
  "max_results": 5,
  "use_llm": true,
  "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
}
```

**Parameters**:
- `query` (string, required): Search query
- `max_results` (integer, optional): Maximum number of results (1-20). Default: `5`
- `use_llm` (boolean, optional): Whether to generate LLM response. Default: `true`
- `llm_model` (string, optional): Bedrock model ID. Default: `"anthropic.claude-3-haiku-20240307-v1:0"`

**Available Models**:
- `anthropic.claude-3-haiku-20240307-v1:0` (Fast, cost-effective)
- `anthropic.claude-3-sonnet-20240229-v1:0` (Balanced)
- `anthropic.claude-3-opus-20240229-v1:0` (Most capable)
- `amazon.titan-text-express-v1` (AWS Titan)

**Response**:
```json
{
  "query": "How do I deploy an application?",
  "results": [
    {
      "title": "Deploying Applications",
      "source": "docs/user-guide/deploying-application/README.md",
      "header": "Quick Start",
      "content": "To deploy an application in Devtron...",
      "score": 0.89
    }
  ],
  "llm_response": "To deploy an application in Devtron, follow these steps:\n\n1. **Create Application**...",
  "total_results": 5
}
```

**Example**:
```bash
# Search with LLM response
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "max_results": 5,
    "use_llm": true
  }'

# Search without LLM (just vector search)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "max_results": 10,
    "use_llm": false
  }'
```

**Response Fields**:
- `query`: The original search query
- `results`: Array of search results from vector database
  - `title`: Document title
  - `source`: File path in repository
  - `header`: Section header (if applicable)
  - `content`: Relevant content chunk
  - `score`: Similarity score (0-1, higher is better)
- `llm_response`: LLM-generated answer based on search results (if `use_llm: true`)
- `total_results`: Number of results returned

---

## 🔧 Integration Examples

### Python

```python
import requests

API_URL = "http://localhost:8000"

# Search documentation
response = requests.post(
    f"{API_URL}/search",
    json={
        "query": "How to configure CI/CD pipeline?",
        "max_results": 5,
        "use_llm": True
    }
)

data = response.json()
print(f"LLM Response: {data['llm_response']}")
print(f"\nFound {data['total_results']} results:")
for result in data['results']:
    print(f"- {result['title']} (score: {result['score']:.2f})")
```

### JavaScript/Node.js

```javascript
const API_URL = "http://localhost:8000";

async function searchDocs(query) {
  const response = await fetch(`${API_URL}/search`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query: query,
      max_results: 5,
      use_llm: true
    })
  });
  
  const data = await response.json();
  console.log('LLM Response:', data.llm_response);
  console.log('Results:', data.results);
}

searchDocs("How to configure CI/CD pipeline?");
```

### cURL

```bash
# Search
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{"query": "How to configure CI/CD pipeline?", "use_llm": true}'

# Re-index
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": false}'
```

---

## 🔐 Authentication (Optional)

For production deployment, you should add authentication. Here's how to add API key authentication:

### Add to `.env`:
```bash
API_KEY=your-secret-api-key-here
```

### Modify `api.py`:
```python
from fastapi import Header, HTTPException

async def verify_api_key(x_api_key: str = Header(...)):
    if x_api_key != os.getenv("API_KEY"):
        raise HTTPException(status_code=401, detail="Invalid API key")
    return x_api_key

# Add to endpoints
@app.post("/search", dependencies=[Depends(verify_api_key)])
async def search_documentation(request: SearchRequest):
    ...
```

### Usage with API key:
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{"query": "How to deploy?"}'
```

---

## 📊 Response Format Design

The API returns responses in a structured format optimized for different use cases:

### For Chatbots/LLM Integration
Use `use_llm: true` to get a ready-to-use response:
```json
{
  "llm_response": "Formatted markdown response ready to display"
}
```

### For Custom UI/Search
Use `use_llm: false` to get raw search results:
```json
{
  "results": [
    {
      "title": "...",
      "content": "...",
      "score": 0.89
    }
  ]
}
```

### For Hybrid Approach
Use `use_llm: true` to get both:
- `llm_response`: For direct display
- `results`: For showing sources/references

---

## 🚀 Deployment

### Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f docs-api

# Stop services
docker-compose down
```

### Kubernetes

See `k8s/` directory for Kubernetes manifests (to be created).

### Cloud Deployment

The API can be deployed to:
- AWS ECS/Fargate
- Google Cloud Run
- Azure Container Instances
- Any platform supporting Docker containers

---

## 📈 Performance

- **Search latency**: ~100-300ms (vector search only)
- **LLM latency**: ~1-3s (with Claude Haiku)
- **Throughput**: ~100 requests/second (with proper scaling)
- **Database**: Supports millions of document chunks

---

## 🐛 Troubleshooting

### Documentation not indexed
```bash
# Check health
curl http://localhost:8000/health

# If docs_indexed: false, run reindex
curl -X POST http://localhost:8000/reindex -H "Content-Type: application/json" -d '{"force": true}'
```

### Slow responses
- Reduce `max_results` parameter
- Use faster LLM model (Claude Haiku)
- Set `use_llm: false` for faster responses

### Database connection errors
```bash
# Check PostgreSQL is running
docker-compose ps

# Restart services
docker-compose restart
```

---

## 📚 Next Steps

1. **Deploy the API** to your infrastructure
2. **Create MCP tools** in your separate repo that call these APIs
3. **Set up periodic re-indexing** (cron job or scheduled task)
4. **Add monitoring** and logging
5. **Configure authentication** for production use

---

For more details, see:
- [PGVECTOR_SETUP.md](PGVECTOR_SETUP.md) - Database setup
- [README.md](README.md) - General information

