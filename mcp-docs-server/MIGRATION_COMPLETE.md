# ✅ Migration Complete: Local Embeddings

## Summary

The Devtron Documentation API has been successfully migrated from AWS Bedrock Titan embeddings to **local embeddings** using BAAI/bge-large-en-v1.5.

## What Changed

### 🎯 Key Changes

1. **Embeddings**: AWS Bedrock Titan → BAAI/bge-large-en-v1.5 (local)
2. **Chunking**: Custom header-based → MarkdownTextSplitter
3. **AWS Dependency**: Required → Optional (only for LLM responses)
4. **Vector Dimension**: 1536 → 1024

### ✅ Benefits

- ✅ **No AWS dependency** for core search functionality
- ✅ **No costs** for embeddings
- ✅ **Faster** - no network latency
- ✅ **Works offline** after initial model download
- ✅ **Better chunking** with MarkdownTextSplitter
- ✅ **Configurable** chunk size and overlap

## Files Modified

### Core Application
1. **`vector_store.py`**
   - Replaced `BedrockEmbeddings` with `LocalEmbeddings`
   - Uses `SentenceTransformer` for embeddings
   - Dynamic embedding dimension based on model

2. **`doc_processor.py`**
   - Added `MarkdownTextSplitter` for chunking
   - Configurable chunk size and overlap
   - Better markdown structure preservation

3. **`api.py`**
   - Added embedding model configuration
   - AWS Bedrock now optional
   - Graceful degradation when AWS not available

### Configuration
4. **`requirements.txt`**
   - Added: `sentence-transformers`, `torch`, `langchain`, `langchain-text-splitters`
   - AWS dependencies now optional

5. **`.env.example`**
   - Added: `EMBEDDING_MODEL`, `CHUNK_SIZE`, `CHUNK_OVERLAP`
   - AWS credentials now commented (optional)

### Documentation
6. **`README.md`** - Updated architecture and features
7. **`CHANGES.md`** - Detailed migration guide
8. **`MIGRATION_COMPLETE.md`** - This file

## Quick Start (New Installation)

```bash
cd mcp-docs-server

# Copy environment file
cp .env.example .env

# Start with Docker
docker-compose up -d

# Or install locally
pip install -r requirements.txt
python api.py &

# Index documentation
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'

# Test search
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{"query": "How to deploy?", "use_llm": false}'
```

## Migration (Existing Installation)

```bash
# 1. Update dependencies
pip install -r requirements.txt

# 2. Update .env file
cat >> .env << EOF
EMBEDDING_MODEL=BAAI/bge-large-en-v1.5
CHUNK_SIZE=1000
CHUNK_OVERLAP=0
EOF

# 3. Drop old table (dimension changed)
psql -h localhost -U postgres -d devtron_docs -c "DROP TABLE IF EXISTS documents;"

# 4. Restart API
docker-compose restart docs-api
# Or: python api.py &

# 5. Re-index
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

## Configuration

### Embedding Model

Default: `BAAI/bge-large-en-v1.5` (1024 dimensions)

Alternatives:
```bash
# Smaller, faster (384 dimensions)
EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2

# Medium (768 dimensions)
EMBEDDING_MODEL=sentence-transformers/all-mpnet-base-v2
```

### Chunking

```bash
# Default
CHUNK_SIZE=1000
CHUNK_OVERLAP=0

# More granular
CHUNK_SIZE=500
CHUNK_OVERLAP=50

# More context
CHUNK_SIZE=2000
CHUNK_OVERLAP=200
```

### AWS Bedrock (Optional)

Only needed for LLM-enhanced responses:

```bash
# Optional - comment out if not needed
# AWS_REGION=us-east-1
# AWS_ACCESS_KEY_ID=your_key
# AWS_SECRET_ACCESS_KEY=your_secret
```

## Testing

```bash
# Run test suite
python test_api.py

# Manual test - search without LLM
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "max_results": 5,
    "use_llm": false
  }'

# Manual test - search with LLM (requires AWS)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "max_results": 5,
    "use_llm": true
  }'
```

## Architecture

```
GitHub Docs → Doc Processor (MarkdownTextSplitter)
                    ↓
         Local Embeddings (BAAI/bge-large-en-v1.5)
                    ↓
         PostgreSQL + pgvector (1024-dim vectors)
                    ↓
              FastAPI Server
                    ↓
         /search, /reindex, /health
                    ↓
              MCP Tools (separate repo)

Optional: AWS Bedrock Claude (for LLM responses)
```

## Performance

### First Run
- Model download: ~1.3GB (one-time)
- Initial indexing: 2-5 minutes

### Subsequent Runs
- Embedding generation: ~50-100ms per chunk (local)
- Search: 100-300ms
- With LLM: 1-3 seconds (if AWS configured)

## Troubleshooting

### Model Download Fails
```bash
# Pre-download manually
python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('BAAI/bge-large-en-v1.5')"
```

### Dimension Mismatch Error
```bash
# Re-create table with new dimension
psql -h localhost -U postgres -d devtron_docs -c "DROP TABLE IF EXISTS documents;"
# Restart API and re-index
```

### Out of Memory
```bash
# Use smaller model
EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2
```

## Next Steps

1. ✅ Test the API with local embeddings
2. ✅ Re-index your documentation
3. ✅ Update your MCP tools (no changes needed - API is compatible)
4. ✅ (Optional) Configure AWS for LLM responses
5. ✅ Deploy to production

## Documentation

- **`GETTING_STARTED.md`** - Quick start guide
- **`CHANGES.md`** - Detailed migration guide
- **`API_DOCUMENTATION.md`** - API reference
- **`README.md`** - Main documentation
- **`MCP_TOOL_EXAMPLE.md`** - MCP integration

---

**Status**: ✅ **MIGRATION COMPLETE**

The API now uses local embeddings and works without AWS credentials for core search functionality!

