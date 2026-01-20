# ============================================================================
# Multi-Stage Dockerfile for Devtron Documentation RAG Server
# Optimized for minimal image size and fast builds
# ============================================================================

# Stage 1: Builder - Install dependencies
FROM python:3.12-slim AS builder

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
        libgomp1 \
        && \
    apt-get clean && \
    rm -rf \
        /var/lib/apt/lists/* \
        /tmp/* \
        /var/tmp/* \
        /usr/share/doc/* \
        /usr/share/man/* \
        /usr/share/locale/*

# Copy Python packages from builder
COPY --from=builder /root/.local /root/.local

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
    TRANSFORMERS_CACHE=/tmp/transformers \
    HF_HOME=/tmp/huggingface \
    TORCH_HOME=/tmp/torch \
    PIP_NO_CACHE_DIR=1

# Expose port
EXPOSE 8000

# Run the application
CMD ["python", "api.py"]

