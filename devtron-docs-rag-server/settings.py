"""Settings shared by the server and the image build."""

import os

# The one place the embedding model is named; the EMBEDDING_MODEL env var overrides it.
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL") or "BAAI/bge-small-en-v1.5"
