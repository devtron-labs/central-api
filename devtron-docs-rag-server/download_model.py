#!/usr/bin/env python3
"""
Pre-download embedding model to cache it in Docker image.
This prevents the model from being downloaded on every container startup.
"""

import logging
import os
import sys
from sentence_transformers import SentenceTransformer

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

def download_model(model_name: str = "BAAI/bge-large-en-v1.5"):
    """Download and cache the embedding model."""

    # Verify cache directories are set
    cache_dir = os.getenv('SENTENCE_TRANSFORMERS_HOME')
    logger.info(f"Cache directory: {cache_dir}")
    logger.info(f"Downloading embedding model: {model_name}")
    logger.info("This will download ~1.34GB and may take several minutes...")

    try:
        # Download model - it will use SENTENCE_TRANSFORMERS_HOME env var automatically
        model = SentenceTransformer(model_name)
        dimension = model.get_sentence_embedding_dimension()

        logger.info(f"✓ Model downloaded successfully!")
        logger.info(f"  Model: {model_name}")
        logger.info(f"  Embedding dimension: {dimension}")
        logger.info(f"  Cache location: {cache_dir}")

        # Verify the cache exists
        if cache_dir and os.path.exists(cache_dir):
            logger.info(f"  Cache verified at: {cache_dir}")
            # List contents
            for root, dirs, files in os.walk(cache_dir):
                logger.info(f"    {root}: {len(files)} files")

        return True
    except Exception as e:
        logger.error(f"✗ Failed to download model: {str(e)}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    model_name = sys.argv[1] if len(sys.argv) > 1 else "BAAI/bge-large-en-v1.5"
    success = download_model(model_name)
    sys.exit(0 if success else 1)

