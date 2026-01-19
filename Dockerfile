# Stage 1: Build Go application
FROM golang:1.19.9-alpine3.18 AS build-env
RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
RUN go install github.com/google/wire/cmd/wire@latest
WORKDIR /go/src/github.com/devtron-labs/central-api
ADD . /go/src/github.com/devtron-labs/central-api
RUN GOOS=linux make

# Stage 2: Final image with both Go and Python
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    git \
    supervisor \
    && rm -rf /var/lib/apt/lists/*

# Copy Go binary
COPY --from=build-env /go/src/github.com/devtron-labs/central-api/central-api /app/central-api
COPY ./DockerfileTemplateData.json /DockerfileTemplateData.json
COPY ./BuildpackMetadata.json /BuildpackMetadata.json

# Copy Python RAG server
WORKDIR /app/rag-server
COPY devtron-docs-rag-server/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY devtron-docs-rag-server/api.py .
COPY devtron-docs-rag-server/doc_processor.py .
COPY devtron-docs-rag-server/vector_store.py .
COPY devtron-docs-rag-server/run_migrations.py .

# Copy migration scripts from root
COPY scripts /app/scripts

# Create directories for data persistence
RUN mkdir -p /data/devtron-docs

# Set environment variables
ENV DOCS_PATH=/data/devtron-docs
ENV PYTHONUNBUFFERED=1
ENV DOCS_RAG_SERVER_URL=http://localhost:8000

# Copy supervisor configuration
RUN mkdir -p /var/log/supervisor /etc/supervisor/conf.d
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

WORKDIR /app

# Expose ports
EXPOSE 8080 8000

# Start both services using supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]