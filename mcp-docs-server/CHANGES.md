# Changes: Local Embeddings Migration

## Summary

The Devtron Documentation API has been updated to use **local embeddings** instead of AWS Bedrock Titan for text embeddings. This removes the AWS dependency for the core search functionality.

## What Changed

### ✅ New Features

1. **Local Embeddings Model**: BAAI/bge-large-en-v1.5
   - No AWS dependency for embeddings
   - Runs locally on your machine
   - 1024-dimensional vectors
   - Better performance for retrieval tasks

2. **MarkdownTextSplitter**: Intelligent document chunking
   - Uses LangChain's MarkdownTextSplitter
   - Configurable chunk size (default: 1000 characters)
   - Configurable chunk overlap (default: 0)
   - Better preservation of markdown structure

3. **Optional AWS Bedrock**: Now only needed for LLM responses
   - Search works without AWS credentials
   - LLM-enhanced responses require AWS Bedrock (optional)
   - Graceful degradation if AWS not configured

### 🔧 Technical Changes

#### 1. Dependencies (`requirements.txt`)
**Added:**
- `sentence-transformers>=2.2.2` - For local embeddings
- `torch>=2.0.0` - Required by sentence-transformers
- `langchain>=0.1.0` - For text splitting
- `langchain-text-splitters>=0.0.1` - MarkdownTextSplitter

**Changed:**
- AWS Bedrock (boto3) is now optional

#### 2. Vector Store (`vector_store.py`)
**Changed:**
- `BedrockEmbeddings` → `LocalEmbeddings`
- Uses `SentenceTransformer` instead of AWS Bedrock API
- Embedding dimension: 1536 (Titan) → 1024 (BGE)
- Added instruction prefixes for better retrieval:
  - Documents: `"passage: {text}"`
  - Queries: `"query: {text}"`

#### 3. Document Processor (`doc_processor.py`)
**Changed:**
- Custom header-based chunking → `MarkdownTextSplitter`
- Configurable chunk size and overlap
- Better handling of markdown structure

#### 4. API Server (`api.py`)
**Changed:**
- AWS region parameter removed from VectorStore initialization
- Added embedding model configuration
- Added chunk size/overlap configuration
- AWS Bedrock initialization is now optional
- Graceful error handling when AWS not available

#### 5. Configuration (`.env.example`)
**Added:**
```bash
EMBEDDING_MODEL=BAAI/bge-large-en-v1.5
CHUNK_SIZE=1000
CHUNK_OVERLAP=0
```

**Changed:**
- AWS credentials are now commented out (optional)

## Migration Guide

### For New Installations

No changes needed! Just follow the updated `GETTING_STARTED.md`.

### For Existing Installations

#### Step 1: Update Dependencies

```bash
cd mcp-docs-server
pip install -r requirements.txt
```

This will install:
- sentence-transformers
- torch
- langchain
- langchain-text-splitters

**Note**: First run will download the BAAI/bge-large-en-v1.5 model (~1.3GB)

#### Step 2: Update Environment Variables

Edit your `.env` file:

```bash
# Add these new variables
EMBEDDING_MODEL=BAAI/bge-large-en-v1.5
CHUNK_SIZE=1000
CHUNK_OVERLAP=0

# AWS credentials are now optional (only for LLM responses)
# You can comment them out if you don't need LLM responses
# AWS_ACCESS_KEY_ID=...
# AWS_SECRET_ACCESS_KEY=...
```

#### Step 3: Re-index Documentation

**Important**: The embedding dimension changed from 1536 to 1024, so you need to re-index:

```bash
# Drop the old table (this will delete existing embeddings)
psql -h localhost -U postgres -d devtron_docs -c "DROP TABLE IF EXISTS documents;"

# Restart the API (it will recreate the table with new dimension)
docker-compose restart docs-api

# Or if running locally:
python api.py &

# Re-index all documentation
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

#### Step 4: Test

```bash
# Test search
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to deploy an application?",
    "use_llm": false
  }'
```

## Benefits

### 1. No AWS Dependency for Core Functionality
- ✅ Search works without AWS credentials
- ✅ No AWS costs for embeddings
- ✅ No API rate limits
- ✅ Works offline (after model download)

### 2. Better Performance
- ✅ BAAI/bge-large-en-v1.5 is optimized for retrieval
- ✅ Faster embedding generation (local GPU if available)
- ✅ No network latency

### 3. Better Chunking
- ✅ MarkdownTextSplitter preserves structure
- ✅ Configurable chunk size
- ✅ Better context preservation

### 4. Cost Savings
- ✅ No AWS Bedrock embedding costs
- ✅ AWS only needed for optional LLM responses

## Comparison

| Feature | Before (AWS Bedrock Titan) | After (Local BGE) |
|---------|---------------------------|-------------------|
| **Embedding Model** | amazon.titan-embed-text-v1 | BAAI/bge-large-en-v1.5 |
| **Dimensions** | 1536 | 1024 |
| **AWS Required** | Yes | No (optional for LLM) |
| **Cost** | Free tier, then $0.0001/1K tokens | Free |
| **Speed** | Network latency | Local (faster) |
| **Offline** | No | Yes (after download) |
| **Chunking** | Custom header-based | MarkdownTextSplitter |
| **Chunk Size** | Fixed ~1000 chars | Configurable |

## Troubleshooting

### Model Download Issues

**Problem**: Model download fails or is slow

**Solution**:
```bash
# Pre-download the model
python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('BAAI/bge-large-en-v1.5')"
```

### Memory Issues

**Problem**: Out of memory when loading model

**Solution**:
- Ensure at least 4GB RAM available
- Close other applications
- Use a smaller model: `EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2`

### Dimension Mismatch Error

**Problem**: `ERROR: dimension mismatch`

**Solution**: You need to re-index (see Step 3 above)

## Configuration Options

### Using a Different Embedding Model

You can use any SentenceTransformer model:

```bash
# Smaller, faster (384 dimensions)
EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2

# Larger, more accurate (768 dimensions)
EMBEDDING_MODEL=sentence-transformers/all-mpnet-base-v2

# Default (1024 dimensions)
EMBEDDING_MODEL=BAAI/bge-large-en-v1.5
```

**Note**: Changing the model requires re-indexing.

### Adjusting Chunk Size

```bash
# Smaller chunks (more granular search)
CHUNK_SIZE=500
CHUNK_OVERLAP=50

# Larger chunks (more context)
CHUNK_SIZE=2000
CHUNK_OVERLAP=200
```

## Next Steps

1. ✅ Update dependencies
2. ✅ Update environment variables
3. ✅ Re-index documentation
4. ✅ Test search functionality
5. ✅ (Optional) Configure AWS for LLM responses

For questions or issues, see the updated documentation:
- `GETTING_STARTED.md` - Quick start guide
- `API_DOCUMENTATION.md` - API reference
- `README.md` - Main documentation

