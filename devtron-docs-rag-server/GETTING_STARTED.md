# Getting Started with Devtron Documentation API

This guide will help you get the Devtron Documentation API up and running in 5 minutes.

## 🎯 What You're Building

A REST API that provides:
- **Semantic search** over Devtron documentation
- **LLM-enhanced responses** using AWS Bedrock
- **Auto-sync** from GitHub
- **Production-ready** PostgreSQL storage

## 📋 Prerequisites

Before you start, make sure you have:

- [ ] **Docker & Docker Compose** (recommended) OR Python 3.9+
- [ ] **AWS Account** with Bedrock access
- [ ] **AWS Credentials** (Access Key ID & Secret Access Key)

## 🚀 Quick Start (5 Minutes)

### Step 1: Clone and Navigate

```bash
cd mcp-docs-server
```

### Step 2: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add your AWS credentials
nano .env  # or use your favorite editor
```

**Required configuration in `.env`:**
```bash
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here
AWS_REGION=us-east-1
```

### Step 3: Enable AWS Bedrock Models

1. Go to [AWS Console → Bedrock → Model Access](https://console.aws.amazon.com/bedrock/home#/modelaccess)
2. Click "Manage model access"
3. Enable these models:
   - ✅ **Titan Embeddings G1 - Text** (for embeddings)
   - ✅ **Claude 3 Haiku** (for LLM responses)
4. Click "Save changes"
5. Wait for approval (usually instant)

### Step 4: Start the API

```bash
# One command to start everything!
./start.sh
```

This will:
- Start PostgreSQL with pgvector
- Start the API server
- Set up the database
- Show you the status

### Step 5: Index Documentation

```bash
# Index the documentation (takes 2-5 minutes)
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

### Step 6: Test It!

```bash
# Run the test suite
python test_api.py
```

Or try a manual search:

```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application?",
    "use_llm": true
  }'
```

## 🎉 Success!

Your API is now running at `http://localhost:8000`

### What's Next?

1. **View Interactive Docs**: http://localhost:8000/docs
2. **Read API Documentation**: [API_DOCUMENTATION.md](API_DOCUMENTATION.md)
3. **Create MCP Tools**: [MCP_TOOL_EXAMPLE.md](MCP_TOOL_EXAMPLE.md)

## 📡 Using the API

### Search Documentation

```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to configure CI/CD pipeline?",
    "max_results": 5,
    "use_llm": true
  }'
```

**Response:**
```json
{
  "query": "How to configure CI/CD pipeline?",
  "results": [...],
  "llm_response": "To configure a CI/CD pipeline in Devtron...",
  "total_results": 5
}
```

### Re-index Documentation

```bash
# Incremental update (only changed files)
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": false}'

# Full re-index
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

### Health Check

```bash
curl http://localhost:8000/health
```

## 🔧 Common Tasks

### View Logs

```bash
# Docker
docker-compose logs -f docs-api

# Local
# Logs are printed to console
```

### Stop the API

```bash
# Docker
docker-compose down

# Local
# Press Ctrl+C or kill the process
```

### Restart the API

```bash
# Docker
docker-compose restart docs-api

# Local
./start.sh
```

### Update Documentation

```bash
# Sync latest docs from GitHub
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": false}'
```

## 🐛 Troubleshooting

### "Cannot connect to PostgreSQL"

**Docker:**
```bash
docker-compose up -d postgres
docker-compose ps  # Check if postgres is running
```

**Local:**
```bash
# Install PostgreSQL with pgvector
# See PGVECTOR_SETUP.md for detailed instructions
```

### "AWS credentials not found"

Make sure `.env` file has:
```bash
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
AWS_REGION=us-east-1
```

### "Documentation not indexed"

Run the reindex command:
```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

### "Slow responses"

- Use `"use_llm": false` for faster responses
- Reduce `max_results` parameter
- Check AWS Bedrock quotas

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [README.md](README.md) | Main documentation |
| [API_DOCUMENTATION.md](API_DOCUMENTATION.md) | Complete API reference |
| [PGVECTOR_SETUP.md](PGVECTOR_SETUP.md) | PostgreSQL setup guide |
| [MCP_TOOL_EXAMPLE.md](MCP_TOOL_EXAMPLE.md) | MCP integration example |
| [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) | Technical details |

## 🎯 Next Steps

### For MCP Integration

1. Create a new repository for your MCP server
2. Follow the example in [MCP_TOOL_EXAMPLE.md](MCP_TOOL_EXAMPLE.md)
3. Create MCP tools that call this API
4. Use in Claude Desktop or other MCP clients

### For Production Deployment

1. Deploy PostgreSQL to managed service (AWS RDS, etc.)
2. Deploy API to container platform (ECS, Cloud Run, etc.)
3. Add API key authentication
4. Set up HTTPS with domain name
5. Configure periodic re-indexing (cron job)

### For Development

1. Explore the API at http://localhost:8000/docs
2. Modify `api.py` to add custom endpoints
3. Customize LLM prompts in `generate_llm_response()`
4. Add caching, rate limiting, etc.

## 💡 Tips

- **Periodic Updates**: Set up a cron job to call `/reindex` daily
- **Faster Responses**: Use `use_llm: false` for quick searches
- **Better Answers**: Use Claude Sonnet instead of Haiku for complex queries
- **Cost Optimization**: Bedrock Titan embeddings are free tier eligible
- **Monitoring**: Add logging and metrics for production use

## 🆘 Need Help?

- Check the [API Documentation](API_DOCUMENTATION.md)
- Review [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)
- See [PGVECTOR_SETUP.md](PGVECTOR_SETUP.md) for database issues

---

**Ready to integrate?** See [MCP_TOOL_EXAMPLE.md](MCP_TOOL_EXAMPLE.md) for creating MCP tools that call this API!

