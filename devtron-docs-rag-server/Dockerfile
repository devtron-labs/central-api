# ============================================================================
# Dockerfile for Devtron Documentation RAG Server
# Optimized Python-only image for embedding-based document search
# ============================================================================

FROM python:3.11-slim

LABEL maintainer="Devtron Labs"
LABEL description="Devtron Documentation RAG Server - Optimized for embeddings"

# Install system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        curl \
        libgomp1 \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt && \
    # Clean up pip cache and unnecessary files
    rm -rf /root/.cache/pip && \
    find /usr/local/lib/python3.11 -type d -name "tests" -exec rm -rf {} + 2>/dev/null || true && \
    find /usr/local/lib/python3.11 -type d -name "test" -exec rm -rf {} + 2>/dev/null || true && \
    find /usr/local/lib/python3.11 -type f -name "*.pyc" -delete && \
    find /usr/local/lib/python3.11 -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true

# Copy application code
COPY api.py doc_processor.py vector_store.py ./

# Create necessary directories
RUN mkdir -p /data/devtron-docs

# Environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    DOCS_PATH=/data/devtron-docs \
    DOCS_REPO_URL=https://github.com/devtron-labs/devtron \
    POSTGRES_HOST=localhost \
    POSTGRES_PORT=5432 \
    POSTGRES_DB=devtron_docs \
    POSTGRES_USER=postgres \
    POSTGRES_PASSWORD=postgres \
    TRANSFORMERS_CACHE=/tmp/transformers \
    HF_HOME=/tmp/huggingface \
    TORCH_HOME=/tmp/torch

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Run the application
CMD ["python", "api.py"]

