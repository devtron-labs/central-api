"""
Documentation Processor
Handles cloning, syncing, and processing of Devtron documentation from GitHub.
"""

import logging
import os
import re
from pathlib import Path
from typing import List, Dict, Optional
import hashlib

import git
from git import Repo
from langchain_text_splitters import MarkdownTextSplitter

logger = logging.getLogger(__name__)


class DocumentationProcessor:
    """Processes Devtron documentation from GitHub repository."""
    
    def __init__(self, repo_url: str, local_path: str, chunk_size: int = 1000, chunk_overlap: int = 0):
        """
        Initialize the documentation processor.

        Args:
            repo_url: GitHub repository URL
            local_path: Local path to clone/store the repository
            chunk_size: Size of text chunks for splitting
            chunk_overlap: Overlap between chunks
        """
        self.repo_url = repo_url
        self.local_path = Path(local_path)
        self.repo: Optional[Repo] = None
        self.docs_dir = self.local_path / "docs"

        # Initialize markdown splitter
        self.md_splitter = MarkdownTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap
        )
        logger.info(f"Initialized MarkdownTextSplitter with chunk_size={chunk_size}, chunk_overlap={chunk_overlap}")
        
    async def sync_docs(self) -> List[str]:
        """
        Sync documentation from GitHub.
        
        Returns:
            List of changed file paths
        """
        changed_files = []
        
        try:
            if not self.local_path.exists():
                logger.info(f"Cloning repository from {self.repo_url}...")
                self.repo = Repo.clone_from(self.repo_url, self.local_path)
                logger.info("Repository cloned successfully")
                # All files are new
                changed_files = self._get_all_markdown_files()
            else:
                logger.info("Pulling latest changes...")
                self.repo = Repo(self.local_path)
                
                # Get current commit
                old_commit = self.repo.head.commit
                
                # Pull changes
                origin = self.repo.remotes.origin
                origin.pull()
                
                # Get new commit
                new_commit = self.repo.head.commit
                
                # Find changed files
                if old_commit != new_commit:
                    diff = old_commit.diff(new_commit)
                    for item in diff:
                        if item.a_path.endswith('.md') and item.a_path.startswith('docs/'):
                            changed_files.append(item.a_path)
                    logger.info(f"Found {len(changed_files)} changed documentation files")
                else:
                    logger.info("No changes detected")
        
        except Exception as e:
            logger.error(f"Error syncing documentation: {e}", exc_info=True)
            raise
        
        return changed_files
    
    def _get_all_markdown_files(self) -> List[str]:
        """Get all markdown files in the docs directory."""
        markdown_files = []
        
        if self.docs_dir.exists():
            for md_file in self.docs_dir.rglob("*.md"):
                rel_path = md_file.relative_to(self.local_path)
                markdown_files.append(str(rel_path))
        
        return markdown_files
    
    async def get_all_documents(self) -> List[Dict[str, str]]:
        """
        Get all documentation files as processed documents.
        
        Returns:
            List of document dictionaries with metadata
        """
        documents = []
        markdown_files = self._get_all_markdown_files()
        
        for file_path in markdown_files:
            doc = await self._process_markdown_file(file_path)
            if doc:
                documents.append(doc)
        
        logger.info(f"Processed {len(documents)} documents")
        return documents
    
    async def get_documents_by_paths(self, paths: List[str]) -> List[Dict[str, str]]:
        """
        Get specific documents by their paths.
        
        Args:
            paths: List of file paths
            
        Returns:
            List of processed documents
        """
        documents = []
        
        for path in paths:
            doc = await self._process_markdown_file(path)
            if doc:
                documents.append(doc)
        
        return documents
    
    async def get_document_by_path(self, path: str) -> Optional[str]:
        """
        Get a specific document by path.
        
        Args:
            path: Relative path to the document
            
        Returns:
            Document content or None
        """
        file_path = self.local_path / path
        
        if file_path.exists() and file_path.suffix == '.md':
            try:
                return file_path.read_text(encoding='utf-8')
            except Exception as e:
                logger.error(f"Error reading file {path}: {e}")
                return None
        
        return None
    
    async def list_sections(self, filter_term: str = "") -> List[Dict[str, str]]:
        """
        List all documentation sections.
        
        Args:
            filter_term: Optional filter string
            
        Returns:
            List of section metadata
        """
        sections = []
        markdown_files = self._get_all_markdown_files()
        
        for file_path in markdown_files:
            if filter_term and filter_term.lower() not in file_path.lower():
                continue
            
            title = self._extract_title_from_path(file_path)
            sections.append({
                "title": title,
                "path": file_path
            })
        
        return sections

    async def _process_markdown_file(self, file_path: str) -> Optional[Dict[str, str]]:
        """
        Process a markdown file into a document.

        Args:
            file_path: Relative path to the markdown file

        Returns:
            Document dictionary or None
        """
        full_path = self.local_path / file_path

        if not full_path.exists():
            logger.warning(f"File not found: {file_path}")
            return None

        try:
            content = full_path.read_text(encoding='utf-8')

            # Extract title from first heading or filename
            title = self._extract_title(content, file_path)

            # Chunk the content for better retrieval
            chunks = self._chunk_markdown(content, file_path)

            # Create document ID
            doc_id = hashlib.md5(file_path.encode()).hexdigest()

            # Return the main document (we'll handle chunking in vector store)
            return {
                "id": doc_id,
                "title": title,
                "content": content,
                "source": file_path,
                "chunks": chunks
            }

        except Exception as e:
            logger.error(f"Error processing file {file_path}: {e}")
            return None

    def _extract_title(self, content: str, file_path: str) -> str:
        """Extract title from markdown content or filename."""
        # Try to find first H1 heading
        match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
        if match:
            return match.group(1).strip()

        # Fallback to filename
        return self._extract_title_from_path(file_path)

    def _extract_title_from_path(self, file_path: str) -> str:
        """Extract a readable title from file path."""
        path = Path(file_path)
        # Remove .md extension and convert dashes/underscores to spaces
        title = path.stem.replace('-', ' ').replace('_', ' ')
        # Capitalize words
        return title.title()

    def _chunk_markdown(self, content: str, source: str, chunk_size: int = 1000) -> List[Dict[str, str]]:
        """
        Chunk markdown content using MarkdownTextSplitter.

        Args:
            content: Markdown content
            source: Source file path
            chunk_size: Target size for chunks (in characters) - not used, kept for compatibility

        Returns:
            List of chunks with metadata
        """
        chunks = []

        # Use MarkdownTextSplitter to split content
        text_chunks = self.md_splitter.split_text(content)

        for i, chunk_text in enumerate(text_chunks):
            # Extract header from chunk if present
            header_match = re.search(r'^(#{1,6}\s+.+)$', chunk_text, re.MULTILINE)
            header = header_match.group(1) if header_match else ""

            chunks.append({
                "content": chunk_text.strip(),
                "header": header,
                "source": source
            })

        logger.debug(f"Split {source} into {len(chunks)} chunks")
        return chunks

