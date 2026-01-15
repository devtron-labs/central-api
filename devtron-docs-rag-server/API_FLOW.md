# API Flow & Architecture

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Application                       │
│                    (Web App / CLI / Chatbot)                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/REST
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FastAPI Server (Port 8000)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   /health    │  │   /reindex   │  │      /search         │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└────────┬────────────────────┬────────────────────┬──────────────┘
         │                    │                    │
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌──────────────────────┐
│   PostgreSQL    │  │  GitHub Repo    │  │   AWS Bedrock        │
│   + pgvector    │  │  (Devtron Docs) │  │   (Claude LLM)       │
│                 │  │                 │  │   [Optional]         │
│  Vector Store   │  │  Markdown Files │  │                      │
└─────────────────┘  └─────────────────┘  └──────────────────────┘
```

## Request Flow Diagrams

### 1. Health Check Flow

```
Client                    API Server              PostgreSQL
  │                          │                        │
  │──── GET /health ────────▶│                        │
  │                          │                        │
  │                          │──── Check Connection ─▶│
  │                          │                        │
  │                          │◀──── Status OK ────────│
  │                          │                        │
  │◀─── 200 OK ──────────────│                        │
  │  {                       │                        │
  │    "status": "healthy",  │                        │
  │    "database": "connected"                        │
  │  }                       │                        │
```

### 2. Re-index Flow

```
Client              API Server           GitHub              PostgreSQL
  │                    │                    │                     │
  │─ POST /reindex ───▶│                    │                     │
  │  {"force": true}   │                    │                     │
  │                    │                    │                     │
  │                    │─── git pull ──────▶│                     │
  │                    │                    │                     │
  │                    │◀── docs files ─────│                     │
  │                    │                    │                     │
  │                    │─── Process Markdown Files ───            │
  │                    │    (Split into chunks)                   │
  │                    │                                          │
  │                    │─── Generate Embeddings ──                │
  │                    │    (BAAI/bge-large-en-v1.5)              │
  │                    │                                          │
  │                    │─── Store Vectors ───────────────────────▶│
  │                    │                                          │
  │                    │◀─── Confirm Stored ──────────────────────│
  │                    │                                          │
  │◀─── 200 OK ────────│                                          │
  │  {                                                            │
  │    "status": "success",                                       │
  │    "documents_processed": 156                                 │
  │  }                                                            │
```

### 3. Search Flow (Without LLM)

```
Client              API Server              PostgreSQL
  │                    │                        │
  │─ POST /search ────▶│                        │
  │  {                 │                        │
  │    "query": "...", │                        │
  │    "use_llm": false│                        │
  │  }                 │                        │
  │                    │                        │
  │                    │─── Generate Query ─────│
  │                    │    Embedding           │
  │                    │                        │
  │                    │─── Vector Search ─────▶│
  │                    │    (Cosine Similarity) │
  │                    │                        │
  │                    │◀─── Top K Results ─────│
  │                    │                        │
  │◀─── 200 OK ────────│                        │
  │  {                                          │
  │    "results": [...],                        │
  │    "llm_response": null                     │
  │  }                                          │
```

### 4. Search Flow (With LLM)

```
Client         API Server        PostgreSQL      AWS Bedrock
  │               │                  │                │
  │─ POST ───────▶│                  │                │
  │  /search      │                  │                │
  │  {            │                  │                │
  │   "use_llm":  │                  │                │
  │    true       │                  │                │
  │  }            │                  │                │
  │               │                  │                │
  │               │─── Vector ──────▶│                │
  │               │    Search        │                │
  │               │                  │                │
  │               │◀─── Results ─────│                │
  │               │                  │                │
  │               │─── Build Context ─                │
  │               │    from Results                   │
  │               │                                   │
  │               │─── Invoke LLM ───────────────────▶│
  │               │    (Claude)                       │
  │               │                                   │
  │               │◀─── Generated Response ───────────│
  │               │                                   │
  │◀─── 200 OK ───│                                   │
  │  {                                                │
  │    "results": [...],                              │
  │    "llm_response": "..."                          │
  │  }                                                │
```

## Sample Response Comparison

### Basic Search Response (No LLM)

**Request:**
```json
{
  "query": "deploy application",
  "max_results": 2,
  "use_llm": false
}
```

**Response Time:** ~200ms

**Response:**
```json
{
  "query": "deploy application",
  "results": [
    {
      "title": "Deploying Applications",
      "source": "docs/user-guide/deploying-application/README.md",
      "header": "Deploy Application",
      "content": "To deploy an application in Devtron: 1. Navigate to Applications...",
      "score": 0.89
    },
    {
      "title": "Application Deployment Guide",
      "source": "docs/user-guide/creating-application/workflow/README.md",
      "header": "Workflow Configuration",
      "content": "Workflows in Devtron define how your application is built...",
      "score": 0.85
    }
  ],
  "llm_response": null,
  "total_results": 2
}
```

**Use Case:** Fast lookups, autocomplete, quick reference

---

### Enhanced Search Response (With LLM)

**Request:**
```json
{
  "query": "deploy application",
  "max_results": 5,
  "use_llm": true,
  "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
}
```

**Response Time:** ~3000ms (3 seconds)

**Response:**
```json
{
  "query": "deploy application",
  "results": [
    {
      "title": "Deploying Applications",
      "source": "docs/user-guide/deploying-application/README.md",
      "header": "Deploy Application",
      "content": "To deploy an application in Devtron: 1. Navigate to Applications...",
      "score": 0.89
    }
    // ... 4 more results
  ],
  "llm_response": "# How to Deploy an Application in Devtron\n\nBased on the documentation, here's a comprehensive guide:\n\n## Prerequisites\n- Devtron installed on your Kubernetes cluster\n- Git repository with your application code\n- Docker registry configured\n\n## Step-by-Step Process\n\n1. **Navigate to Applications**\n   - Open Devtron dashboard\n   - Go to Applications section\n\n2. **Create New Application**\n   - Click 'Create New'\n   - Provide application name and project\n\n3. **Configure Git Repository**\n   - Connect your Git repository\n   - Select branch and build context\n\n4. **Set Up CI Pipeline**\n   - Configure Dockerfile or buildpack\n   - Add pre/post build scripts if needed\n   - Select Docker registry\n\n5. **Configure CD Pipeline**\n   - Choose target environment\n   - Set deployment strategy (rolling, blue-green, canary)\n   - Configure environment variables and secrets\n\n6. **Deploy**\n   - Click 'Deploy' to trigger the pipeline\n   - Monitor deployment progress\n\nDevtron will automatically build your Docker image and deploy it to Kubernetes.",
  "total_results": 5
}
```

**Use Case:** Chatbots, detailed answers, user support, documentation assistance

## Performance Metrics

| Operation | Avg Time | Notes |
|-----------|----------|-------|
| Health Check | <50ms | Simple DB ping |
| Search (No LLM) | 200-500ms | Vector similarity search |
| Search (With LLM) | 2-5s | Includes LLM inference |
| Re-index (Incremental) | 30-60s | Only changed files |
| Re-index (Full) | 5-10min | All documentation |

## Error Handling Flow

```
Client                    API Server
  │                          │
  │─── POST /search ────────▶│
  │                          │
  │                          │─── Check if indexed
  │                          │
  │                          │    ❌ Not indexed
  │                          │
  │◀─── 400 Bad Request ─────│
  │  {                       │
  │    "detail": "Documentation not indexed"
  │  }                       │
  │                          │
  │─── POST /reindex ───────▶│
  │                          │
  │◀─── 200 OK ──────────────│
  │                          │
  │─── POST /search ────────▶│
  │                          │
  │◀─── 200 OK ──────────────│
  │  { "results": [...] }    │
```

## Integration Patterns

### Pattern 1: Direct API Calls
```
User → Your App → Devtron Docs API → Response
```
Best for: Custom applications, internal tools

### Pattern 2: Cached Responses
```
User → Your App → Cache → Devtron Docs API
                    ↓
                Response
```
Best for: High-traffic applications, repeated queries

### Pattern 3: Async Processing
```
User → Queue → Background Worker → Devtron Docs API
  ↓                                        ↓
Immediate                              Store Result
Response                                    ↓
                                    Notify User
```
Best for: Batch processing, scheduled updates

## Security Considerations

1. **API Authentication**: Add API key validation in production
2. **Rate Limiting**: Implement rate limits per client
3. **Input Validation**: Already handled by Pydantic models
4. **CORS**: Configure allowed origins in production
5. **AWS Credentials**: Use IAM roles instead of access keys
6. **Database**: Use strong passwords, enable SSL

## Scaling Recommendations

- **Horizontal Scaling**: Run multiple API instances behind load balancer
- **Database**: Use PostgreSQL read replicas for search queries
- **Caching**: Add Redis for frequently accessed results
- **CDN**: Cache static responses at edge locations

