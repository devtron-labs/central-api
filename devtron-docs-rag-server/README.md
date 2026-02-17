# Devtron Documentation API

A REST API service that provides semantic search over Devtron documentation using local embeddings (BAAI/bge-large-en-v1.5) and PostgreSQL pgvector.

## Features

- 🔍 **Semantic Search**: Find relevant documentation using natural language queries
- 🤖 **Local Embeddings**: Uses BAAI/bge-large-en-v1.5 model (no AWS dependency for embeddings)
- 📝 **Smart Chunking**: MarkdownTextSplitter for optimal document chunking
- 🔄 **Auto-Sync**: Automatically syncs with GitHub documentation repository
- 🗄️ **PostgreSQL + pgvector**: Production-ready vector database
- 💡 **Optional LLM**: AWS Bedrock Claude for enhanced responses (optional)
- 🔄 **Incremental Updates**: Only re-indexes changed files on sync
- 🐳 **Docker Support**: Easy deployment with Docker Compose

## 🎯 For Athena-BE / MCP Tool Integration

**Important:** If you're integrating this with Athena-BE (which already has LLM capabilities):

- ✅ **Use `use_llm=false`** in all search requests
- ✅ **Let Athena-BE handle LLM processing** to avoid double token consumption
- ✅ **No AWS credentials needed** in this API
- ✅ **See [MCP_INTEGRATION_GUIDE.md](./MCP_INTEGRATION_GUIDE.md)** for detailed integration guide

**Why?** Using `use_llm=true` would cause LLM to be called twice (once here, once in Athena-BE), doubling your token costs and latency!

## Architecture

```
┌─────────────────┐
│  GitHub Docs    │
│  Repository     │
└────────┬────────┘
         │ git pull
         ▼
┌─────────────────────────┐
│  Doc Processor          │
│  - Clone/Sync           │
│  - MarkdownTextSplitter │
│  - Chunk (1000 chars)   │
└────────┬────────────────┘
         │
         ▼
┌──────────────────────────┐      ┌──────────────────┐
│ Local Embeddings         │◄─────┤  Vector Store    │
│ BAAI/bge-large-en-v1.5   │      │  (PostgreSQL +   │
│ (1024 dimensions)        │      │   pgvector)      │
└──────────────────────────┘      └────────┬─────────┘
                                           │
                                           ▼
                                  ┌────────────────────┐
                                  │   FastAPI Server   │
                                  │   - /search        │
                                  │   - /reindex       │
                                  │   - /health        │
                                  └────────┬───────────┘
                                           │
                                           ▼
                                  ┌────────────────────┐
                                  │   MCP Tools        │
                                  │   (Separate Repo)  │
                                  │   - Call APIs      │
                                  └────────────────────┘

Optional (for LLM responses):
┌──────────────────┐
│ AWS Bedrock      │
│ Claude Models    │
└──────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Python 3.9+
- PostgreSQL 12+ with pgvector extension
- Docker (optional, recommended)
- AWS Account with Bedrock access (optional - only for LLM enhanced responses)

### Option 1: Docker (Recommended)

```bash
cd mcp-docs-server

# Copy and configure environment
cp .env.example .env
# Edit .env (AWS credentials optional - only needed for LLM responses)

# Start all services (PostgreSQL + API)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f docs-api
```

The API will be available at `http://localhost:8000`

### Option 2: Local Setup

1. **Install PostgreSQL with pgvector**:
   See [PGVECTOR_SETUP.md](PGVECTOR_SETUP.md) for detailed instructions.

2. **Install Python dependencies**:
```bash
cd mcp-docs-server
pip install -r requirements.txt
```

3. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Setup database**:
```bash
./setup_database.sh
```

5. **Configure AWS credentials** (choose one method):

   **Option A: Environment variables**
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

   **Option B: AWS CLI profile**
   ```bash
   aws configure
   # Or use existing profile
   export AWS_PROFILE=your_profile
   ```

6. **Enable AWS Bedrock** (if not already enabled):
   - Go to AWS Console → Bedrock → Model access
   - Request access to:
     - "Titan Embeddings G1 - Text" (for embeddings)
     - "Claude 3 Haiku" (for LLM responses)
   - Wait for approval (usually instant)

## 📡 API Usage

### Start the API Server

```bash
# Using Docker
docker-compose up -d

# Or locally
python api.py
```

The API will be available at `http://localhost:8000`

### Interactive Documentation

Visit these URLs in your browser:
- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc

### API Endpoints

#### 1. Health Check
```bash
curl http://localhost:8000/health
```

#### 2. Re-index Documentation
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

#### 3. Search Documentation
```bash
# Search with LLM response
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "max_results": 5,
    "use_llm": true
  }'

# Search without LLM (faster)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "max_results": 10,
    "use_llm": false
  }'
```

### Testing the API

Run the test suite:
```bash
python test_api.py
```

For detailed API documentation, see [API_DOCUMENTATION.md](API_DOCUMENTATION.md)

#### 1. `search_docs`
Search documentation using semantic search.

**Parameters**:
- `query` (string, required): Search query
- `max_results` (integer, optional): Maximum results to return (default: 5)

**Example**:
```json
{
  "query": "How do I deploy an application?",
  "max_results": 3
}
```

#### 2. `get_doc_by_path`
Retrieve a specific documentation file by path.

**Parameters**:
- `path` (string, required): Relative path to the documentation file

**Example**:
```json
{
  "path": "docs/user-guide/deploying-application.md"
}
```

#### 3. `sync_docs`
Manually trigger documentation synchronization from GitHub.

**Parameters**: None

**Example**:
```json
{}
```

#### 4. `list_doc_sections`
List all available documentation sections.

**Parameters**:
- `filter` (string, optional): Filter sections by keyword

**Example**:
```json
{
  "filter": "user-guide"
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DOCS_REPO_URL` | GitHub repository URL | `https://github.com/devtron-labs/devtron` |
| `DOCS_PATH` | Local path for cloned docs | `./devtron-docs` |
| `CHROMA_DB_PATH` | ChromaDB persistence path | `./chroma_db` |
| `AWS_REGION` | AWS region for Bedrock | `us-east-1` |
| `AWS_ACCESS_KEY_ID` | AWS access key | - |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | - |
| `LOG_LEVEL` | Logging level | `INFO` |

## How It Works

### 1. Documentation Sync
- Clones the Devtron docs repository from GitHub
- On subsequent runs, pulls latest changes
- Detects modified files using git diff

### 2. Document Processing
- Parses markdown files
- Extracts titles and metadata
- Chunks content by headers (H2, H3) for better retrieval
- Maintains source references

### 3. Vectorization
- **When**: On first run and when files change
- **Where**: Stored in local ChromaDB (persisted to disk)
- **How**: AWS Bedrock Titan generates embeddings
- **Cost**: Free tier covers ~1M tokens/month

### 4. Search
- Converts query to embedding using Bedrock Titan
- Performs similarity search in ChromaDB
- Returns top-k most relevant chunks with metadata

## Integration with Chatbot

To integrate with your Python chatbot:

```python
from mcp import ClientSession
from mcp.client.stdio import stdio_client

# Connect to MCP server
async with stdio_client("python", ["server.py"]) as (read, write):
    async with ClientSession(read, write) as session:
        # Initialize
        await session.initialize()
        
        # Search docs
        result = await session.call_tool(
            "search_docs",
            {"query": "How to configure CI/CD pipeline?", "max_results": 3}
        )
        
        # Use result in your chatbot context
        context = result[0].text
```

## Troubleshooting

### AWS Bedrock Access Denied
- Ensure you've requested access to Titan Embeddings in AWS Console
- Check your AWS credentials are correct
- Verify your region supports Bedrock (us-east-1, us-west-2, etc.)

### ChromaDB Errors
- Delete `./chroma_db` directory and restart to rebuild index
- Check disk space for vector storage

### Git Sync Issues
- Ensure you have internet connectivity
- Check GitHub repository URL is correct
- For private repos, configure git credentials

## Performance

- **Initial indexing**: ~2-5 minutes for full Devtron docs
- **Search latency**: <500ms per query
- **Update sync**: Only re-indexes changed files (~10-30 seconds)
- **Storage**: ~50-100MB for ChromaDB vectors

## License

Apache License 2.0 - Same as Devtron project

