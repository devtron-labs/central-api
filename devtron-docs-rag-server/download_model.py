#!/usr/bin/env python3
"""
Pre-download embedding model to cache it in Docker image.
This prevents the model from being downloaded on every container startup.
"""

import logging
import sys
from sentence_transformers import SentenceTransformer

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

def download_model(model_name: str = "BAAI/bge-large-en-v1.5"):
    """Download and cache the embedding model."""
    logger.info(f"Downloading embedding model: {model_name}")
    logger.info("This will download ~1.34GB and may take several minutes...")
    
    try:
        model = SentenceTransformer(model_name)
        dimension = model.get_sentence_embedding_dimension()
        
        logger.info(f"✓ Model downloaded successfully!")
        logger.info(f"  Model: {model_name}")
        logger.info(f"  Embedding dimension: {dimension}")
        logger.info(f"  Model is now cached and ready to use")
        
        return True
    except Exception as e:
        logger.error(f"✗ Failed to download model: {str(e)}")
        return False

if __name__ == "__main__":
    model_name = sys.argv[1] if len(sys.argv) > 1 else "BAAI/bge-large-en-v1.5"
    success = download_model(model_name)
    sys.exit(0 if success else 1)

