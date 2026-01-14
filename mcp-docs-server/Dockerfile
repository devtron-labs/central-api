FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY api.py .
COPY doc_processor.py .
COPY vector_store.py .
COPY .env.example .

# Create directories for data persistence
RUN mkdir -p /data/devtron-docs

# Set environment variables
ENV DOCS_PATH=/data/devtron-docs
ENV PYTHONUNBUFFERED=1
ENV HOST=0.0.0.0
ENV PORT=8000

# Expose API port
EXPOSE 8000

# Run the API server
CMD ["python", "api.py"]

