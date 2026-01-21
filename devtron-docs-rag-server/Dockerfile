# ============================================================================
# Multi-Stage Dockerfile for Devtron Documentation RAG Server
# Optimized for minimal image size and fast builds
# ============================================================================

# Stage 1: Builder - Install dependencies
FROM python:3.12-slim AS builder

# Set cache directories BEFORE downloading anything
ENV TRANSFORMERS_CACHE=/root/.cache/huggingface \
    HF_HOME=/root/.cache/huggingface \
    TORCH_HOME=/root/.cache/torch \
    SENTENCE_TRANSFORMERS_HOME=/root/.cache/torch/sentence_transformers

# Install build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        gcc \
        g++ \
        git \
        && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy and install Python dependencies
COPY requirements.txt .

# Install to user site-packages for easy copying
RUN pip install --user --no-cache-dir -r requirements.txt && \
    # Remove unnecessary files from installed packages
    find /root/.local -type d -name "tests" -exec rm -rf {} + 2>/dev/null || true && \
    find /root/.local -type d -name "test" -exec rm -rf {} + 2>/dev/null || true && \
    find /root/.local -type d -name "docs" -exec rm -rf {} + 2>/dev/null || true && \
    find /root/.local -type f -name "*.pyc" -delete && \
    find /root/.local -type f -name "*.pyo" -delete && \
    find /root/.local -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true && \
    find /root/.local -type f -name "*.c" -delete && \
    find /root/.local -type f -name "*.pyx" -delete && \
    find /root/.local -type f -name "*.md" -delete 2>/dev/null || true && \
    find /root/.local -name "*.dist-info" -type d -exec sh -c 'rm -rf {}/RECORD {}/INSTALLER {}/direct_url.json' \; 2>/dev/null || true

# Pre-download embedding model to cache it in the image
# This prevents downloading ~1.34GB on every container startup
COPY download_model.py .
RUN python download_model.py BAAI/bge-large-en-v1.5 && \
    rm download_model.py && \
    echo "Verifying model cache..." && \
    ls -lah /root/.cache/torch/sentence_transformers/ && \
    echo "Model cache verified!"

# ============================================================================
# Stage 2: Runtime - Minimal production image
FROM python:3.12-slim

LABEL maintainer="Devtron Labs"
LABEL description="Devtron Documentation RAG Server - Optimized for embeddings"

# Install only runtime dependencies (no build tools)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        curl \
        libpq-dev \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Copy Python packages from builder
COPY --from=builder /root/.local /root/.local

# Copy pre-downloaded model cache from builder
COPY --from=builder /root/.cache /root/.cache

# Set PATH to include user site-packages
ENV PATH=/root/.local/bin:$PATH

WORKDIR /app

# Copy application code (only necessary files)
COPY api.py doc_processor.py vector_store.py ./

# Create necessary directories
RUN mkdir -p /data/devtron-docs

# Environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    TRANSFORMERS_CACHE=/root/.cache/huggingface \
    HF_HOME=/root/.cache/huggingface \
    TORCH_HOME=/root/.cache/torch \
    SENTENCE_TRANSFORMERS_HOME=/root/.cache/torch/sentence_transformers \
    HF_HUB_OFFLINE=1 \
    TRANSFORMERS_OFFLINE=1 \
    PIP_NO_CACHE_DIR=1

# Expose port
EXPOSE 8000

# Run the application
CMD ["python", "api.py"]

