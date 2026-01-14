# PostgreSQL pgvector Setup Guide

This guide explains how to set up and use PostgreSQL with pgvector extension for the Devtron MCP Documentation Server.

## 🎯 Why pgvector?

**Advantages over ChromaDB:**
- ✅ **Production-ready**: Battle-tested PostgreSQL database
- ✅ **ACID compliance**: Full transactional support
- ✅ **Scalability**: Handle millions of vectors efficiently
- ✅ **Familiar tooling**: Standard SQL, backup/restore, monitoring
- ✅ **Multi-user**: Concurrent access with proper locking
- ✅ **Cloud-native**: Easy deployment on AWS RDS, Google Cloud SQL, Azure
- ✅ **Advanced indexing**: IVFFlat and HNSW indexes for fast search
- ✅ **Integration**: Works with existing PostgreSQL infrastructure

## 📋 Prerequisites

- Python 3.9+
- PostgreSQL 12+ with pgvector extension
- AWS credentials (for Bedrock Titan embeddings)

## 🚀 Quick Start

### Option 1: Docker (Recommended for Development)

The easiest way to get started is using Docker:

```bash
# Start PostgreSQL with pgvector
docker-compose up -d postgres

# Verify it's running
docker-compose ps
```

This will start PostgreSQL on port 5432 with:
- Database: `devtron_docs`
- User: `postgres`
- Password: `postgres`

### Option 2: Local PostgreSQL Installation

#### macOS (Homebrew)
```bash
# Install PostgreSQL
brew install postgresql@15

# Start PostgreSQL
brew services start postgresql@15

# Install pgvector
brew install pgvector

# Or build from source
cd /tmp
git clone --branch v0.5.1 https://github.com/pgvector/pgvector.git
cd pgvector
make
make install
```

#### Ubuntu/Debian
```bash
# Install PostgreSQL
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib

# Install build dependencies
sudo apt-get install -y postgresql-server-dev-15 build-essential

# Install pgvector
cd /tmp
git clone --branch v0.5.1 https://github.com/pgvector/pgvector.git
cd pgvector
make
sudo make install

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### Windows
```powershell
# Install PostgreSQL from https://www.postgresql.org/download/windows/

# Install pgvector (requires Visual Studio Build Tools)
# Download from: https://github.com/pgvector/pgvector/releases
# Follow installation instructions in the release notes
```

### Option 3: Cloud Providers

#### AWS RDS
1. Create PostgreSQL 15+ instance
2. Enable pgvector extension:
   ```sql
   CREATE EXTENSION vector;
   ```

#### Google Cloud SQL
1. Create PostgreSQL 15+ instance
2. Enable pgvector extension via Cloud SQL flags

#### Azure Database for PostgreSQL
1. Create Flexible Server with PostgreSQL 15+
2. Enable pgvector extension

## ⚙️ Configuration

### 1. Environment Variables

Edit `.env` file:

```bash
# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=devtron_docs
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

# AWS Bedrock Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
```

### 2. Database Setup

Run the setup script:

```bash
./setup_database.sh
```

This will:
- Check PostgreSQL connection
- Create database if it doesn't exist
- Enable pgvector extension
- Verify setup

## 🏗️ Database Schema

The MCP server automatically creates this schema:

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Documents table
CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    source TEXT NOT NULL,
    header TEXT,
    content TEXT NOT NULL,
    chunk_index INTEGER,
    embedding vector(1536),  -- Titan embeddings are 1536-dimensional
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vector similarity search index (IVFFlat)
CREATE INDEX documents_embedding_idx 
ON documents USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Source lookup index
CREATE INDEX documents_source_idx ON documents(source);
```

## 🔍 Vector Search

pgvector supports multiple distance metrics:

- **Cosine distance** (default): `<=>` operator
- **L2 distance**: `<->` operator  
- **Inner product**: `<#>` operator

Example search query:
```sql
SELECT 
    title,
    content,
    1 - (embedding <=> '[0.1, 0.2, ...]'::vector) as similarity
FROM documents
ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector
LIMIT 5;
```

## 📊 Performance Tuning

### Index Types

**IVFFlat** (default):
- Good for most use cases
- Faster build time
- Moderate search speed

```sql
CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
```

**HNSW** (for large datasets):
- Better search performance
- Slower build time
- More memory usage

```sql
CREATE INDEX ON documents USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

### Connection Pooling

The MCP server uses connection pooling (1-10 connections) for optimal performance.

Adjust in `vector_store.py`:
```python
self.pool = SimpleConnectionPool(
    minconn=1,
    maxconn=10,  # Adjust based on load
    ...
)
```

### PostgreSQL Configuration

For better performance, tune these settings in `postgresql.conf`:

```ini
# Memory
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB

# Connections
max_connections = 100

# Maintenance
maintenance_work_mem = 128MB
```

## 🔐 Security

### Production Recommendations

1. **Use strong passwords**:
   ```bash
   POSTGRES_PASSWORD=$(openssl rand -base64 32)
   ```

2. **Restrict network access**:
   ```ini
   # postgresql.conf
   listen_addresses = 'localhost'
   ```

3. **Use SSL connections**:
   ```python
   conn = psycopg2.connect(
       ...,
       sslmode='require'
   )
   ```

4. **Create dedicated user**:
   ```sql
   CREATE USER devtron_mcp WITH PASSWORD 'secure_password';
   GRANT ALL PRIVILEGES ON DATABASE devtron_docs TO devtron_mcp;
   ```

## 🧪 Testing

Run the test suite:

```bash
# Activate virtual environment
source venv/bin/activate

# Run tests
python test_server.py
```

## 🐳 Docker Deployment

### Development
```bash
docker-compose up -d
```

### Production
```bash
# Build and run
docker-compose -f docker-compose.yml up -d

# View logs
docker-compose logs -f mcp-docs-server

# Stop
docker-compose down
```

## 📈 Monitoring

### Check database size
```sql
SELECT pg_size_pretty(pg_database_size('devtron_docs'));
```

### Check table size
```sql
SELECT pg_size_pretty(pg_total_relation_size('documents'));
```

### Check index usage
```sql
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE tablename = 'documents';
```

### Active connections
```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'devtron_docs';
```

## 🔄 Backup & Restore

### Backup
```bash
pg_dump -h localhost -U postgres devtron_docs > backup.sql
```

### Restore
```bash
psql -h localhost -U postgres devtron_docs < backup.sql
```

## 🆘 Troubleshooting

### Connection refused
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Start PostgreSQL (macOS)
brew services start postgresql@15

# Start PostgreSQL (Linux)
sudo systemctl start postgresql
```

### Extension not found
```sql
-- Check available extensions
SELECT * FROM pg_available_extensions WHERE name = 'vector';

-- If not available, reinstall pgvector
```

### Slow queries
```sql
-- Analyze query performance
EXPLAIN ANALYZE
SELECT * FROM documents
ORDER BY embedding <=> '[...]'::vector
LIMIT 5;

-- Rebuild index if needed
REINDEX INDEX documents_embedding_idx;
```

## 📚 Additional Resources

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [AWS RDS PostgreSQL](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html)

---

**Next Steps**: After setup, run `python server.py` to start the MCP server!

