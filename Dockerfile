# ============================================================================
# OPTIMIZED MULTI-STAGE DOCKERFILE
# Reduces image size from 1GB+ to ~600-700MB
# PyTorch supports both CPU and GPU automatically
# ============================================================================

# Stage 1: Build Go application
FROM golang:1.19.9-alpine3.18 AS go-builder

RUN apk add --no-cache git gcc musl-dev make && \
    go install github.com/google/wire/cmd/wire@latest

WORKDIR /go/src/github.com/devtron-labs/central-api

# Cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build Go binary (static, stripped)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux make && \
    strip --strip-all central-api || true

# ============================================================================
# Stage 2: Build Python dependencies
FROM python:3.11-slim AS python-builder

# Install minimal build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        gcc \
        g++ \
        git \
        && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY devtron-docs-rag-server/requirements.txt .

# Install Python packages (PyTorch supports both CPU and GPU)
RUN pip install --no-cache-dir --user -r requirements.txt && \
    # Remove test files and documentation
    find /root/.local -type d -name "tests" -exec rm -rf {} + 2>/dev/null || true && \
    find /root/.local -type d -name "test" -exec rm -rf {} + 2>/dev/null || true && \
    find /root/.local -type d -name "docs" -exec rm -rf {} + 2>/dev/null || true && \
    # Remove bytecode
    find /root/.local -type f -name "*.pyc" -delete && \
    find /root/.local -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true

# ============================================================================
# Stage 3: Minimal runtime image
FROM python:3.11-slim

LABEL maintainer="Devtron Labs"
LABEL description="Central API with RAG Documentation Server - Optimized"

# Install only essential runtime dependencies + curl for debugging
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        supervisor \
        libgomp1 \
        curl \
        && \
    apt-get clean && \
    rm -rf \
        /var/lib/apt/lists/* \
        /tmp/* \
        /var/tmp/* \
        /usr/share/doc/* \
        /usr/share/man/* \
        /usr/share/locale/* \
        /var/cache/apt/*

# Copy Go binary (already stripped)
COPY --from=go-builder /go/src/github.com/devtron-labs/central-api/central-api /app/central-api

# Copy minimal config files
COPY ./DockerfileTemplateData.json /DockerfileTemplateData.json
COPY ./BuildpackMetadata.json /BuildpackMetadata.json

# Copy Python dependencies (already cleaned)
COPY --from=python-builder /root/.local /root/.local
ENV PATH=/root/.local/bin:$PATH

# Copy Python application (only necessary files)
WORKDIR /app/rag-server
COPY devtron-docs-rag-server/api.py \
     devtron-docs-rag-server/doc_processor.py \
     devtron-docs-rag-server/vector_store.py \
     ./

# Setup directories
RUN mkdir -p /data/devtron-docs /var/log/supervisor /etc/supervisor/conf.d

# Copy supervisor config
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Environment variables
ENV DOCS_PATH=/data/devtron-docs \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    DOCS_RAG_SERVER_URL=http://localhost:8000 \
    PIP_NO_CACHE_DIR=1 \
    TRANSFORMERS_CACHE=/tmp/transformers \
    HF_HOME=/tmp/huggingface \
    TORCH_HOME=/tmp/torch

WORKDIR /app

EXPOSE 8080 8000

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health')" || exit 1

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]