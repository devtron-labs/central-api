# MCP Tool Example

This document shows how to create MCP tools in a separate repository that call the Devtron Documentation API.

## Architecture

```
┌─────────────────────────┐
│  Your MCP Server Repo   │
│  (Separate Repository)  │
│                         │
│  ┌──────────────────┐   │
│  │   MCP Tools      │   │      HTTP Requests
│  │   - search_docs  │───┼──────────────────┐
│  │   - reindex_docs │   │                  │
│  └──────────────────┘   │                  ▼
└─────────────────────────┘         ┌────────────────────┐
                                    │  Central API       │
                                    │  (This Repo)       │
                                    │                    │
                                    │  /search           │
                                    │  /reindex          │
                                    └────────────────────┘
```

## Example MCP Server Implementation

Create a new repository with the following structure:

```
my-mcp-server/
├── server.py
├── requirements.txt
└── .env
```

### `requirements.txt`

```
mcp>=1.0.0
requests>=2.31.0
python-dotenv>=1.0.0
```

### `.env`

```bash
# Devtron Documentation API URL
DOCS_API_URL=http://localhost:8000

# Optional: API Key if you add authentication
# DOCS_API_KEY=your-api-key-here
```

### `server.py`

```python
#!/usr/bin/env python3
"""
MCP Server that provides Devtron documentation tools
by calling the central Devtron Documentation API.
"""

import os
import requests
from typing import Any
from dotenv import load_dotenv

from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

# Load environment variables
load_dotenv()

# Configuration
DOCS_API_URL = os.getenv("DOCS_API_URL", "http://localhost:8000")
API_KEY = os.getenv("DOCS_API_KEY")  # Optional

# Initialize MCP server
app = Server("devtron-docs-mcp")


def call_api(endpoint: str, method: str = "GET", data: dict = None) -> dict:
    """
    Call the Devtron Documentation API.
    
    Args:
        endpoint: API endpoint (e.g., "/search")
        method: HTTP method (GET or POST)
        data: Request body for POST requests
        
    Returns:
        API response as dictionary
    """
    url = f"{DOCS_API_URL}{endpoint}"
    headers = {"Content-Type": "application/json"}
    
    # Add API key if configured
    if API_KEY:
        headers["X-API-Key"] = API_KEY
    
    if method == "GET":
        response = requests.get(url, headers=headers)
    else:
        response = requests.post(url, json=data, headers=headers)
    
    response.raise_for_status()
    return response.json()


@app.list_tools()
async def list_tools() -> list[Tool]:
    """List available MCP tools."""
    return [
        Tool(
            name="search_devtron_docs",
            description="Search Devtron documentation using semantic search with LLM-enhanced responses",
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "The search query"
                    },
                    "max_results": {
                        "type": "integer",
                        "description": "Maximum number of results (1-20)",
                        "default": 5
                    },
                    "use_llm": {
                        "type": "boolean",
                        "description": "Whether to use LLM for enhanced response",
                        "default": True
                    }
                },
                "required": ["query"]
            }
        ),
        Tool(
            name="reindex_devtron_docs",
            description="Re-index Devtron documentation from GitHub",
            inputSchema={
                "type": "object",
                "properties": {
                    "force": {
                        "type": "boolean",
                        "description": "Force full re-index",
                        "default": False
                    }
                }
            }
        )
    ]


@app.call_tool()
async def call_tool(name: str, arguments: Any) -> list[TextContent]:
    """Handle tool calls."""
    
    if name == "search_devtron_docs":
        # Call the search API
        response = call_api(
            "/search",
            method="POST",
            data={
                "query": arguments["query"],
                "max_results": arguments.get("max_results", 5),
                "use_llm": arguments.get("use_llm", True)
            }
        )
        
        # Format response
        if response.get("llm_response"):
            # Return LLM response if available
            result = response["llm_response"]
            
            # Optionally add sources
            if response.get("results"):
                result += "\n\n**Sources:**\n"
                for i, r in enumerate(response["results"][:3], 1):
                    result += f"{i}. {r['title']} - {r['source']}\n"
        else:
            # Return search results
            result = f"Found {response['total_results']} results:\n\n"
            for i, r in enumerate(response["results"], 1):
                result += f"{i}. **{r['title']}**\n"
                result += f"   Source: {r['source']}\n"
                result += f"   Score: {r['score']:.2f}\n"
                result += f"   {r['content'][:200]}...\n\n"
        
        return [TextContent(type="text", text=result)]
    
    elif name == "reindex_devtron_docs":
        # Call the reindex API
        response = call_api(
            "/reindex",
            method="POST",
            data={"force": arguments.get("force", False)}
        )
        
        result = f"✅ {response['message']}\n"
        result += f"Documents processed: {response['documents_processed']}\n"
        result += f"Changed files: {response['changed_files']}"
        
        return [TextContent(type="text", text=result)]
    
    else:
        raise ValueError(f"Unknown tool: {name}")


async def main():
    """Run the MCP server."""
    async with stdio_server() as (read_stream, write_stream):
        await app.run(read_stream, write_stream, app.create_initialization_options())


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
```

## Usage

### 1. Start the Central API

In the `central-api` repository:

```bash
cd mcp-docs-server
docker-compose up -d
```

### 2. Start Your MCP Server

In your separate MCP repository:

```bash
# Install dependencies
pip install -r requirements.txt

# Configure API URL
echo "DOCS_API_URL=http://localhost:8000" > .env

# Run the MCP server
python server.py
```

### 3. Use in Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "devtron-docs": {
      "command": "python",
      "args": ["/path/to/your/mcp-server/server.py"]
    }
  }
}
```

### 4. Test the Tools

In Claude Desktop, you can now use:

```
Search Devtron documentation for "How to deploy an application"
```

Claude will call your MCP tool, which will call the central API, and return the response.

## Benefits of This Architecture

1. **Separation of Concerns**: 
   - Central API handles documentation indexing and search
   - MCP tools handle user interaction

2. **Reusability**: 
   - Multiple MCP servers can use the same central API
   - API can be called from web apps, CLI tools, etc.

3. **Scalability**: 
   - Central API can be deployed once and shared
   - Easy to add caching, rate limiting, etc.

4. **Maintainability**: 
   - Update documentation logic in one place
   - MCP tools remain simple and focused

5. **Flexibility**:
   - Can add authentication to the API
   - Can deploy API separately from MCP tools
   - Can use different LLM models per MCP server

## Advanced: Adding Authentication

If you add API key authentication to the central API:

### In Central API (`api.py`):

```python
from fastapi import Header, HTTPException, Depends

async def verify_api_key(x_api_key: str = Header(...)):
    expected_key = os.getenv("API_KEY")
    if not expected_key or x_api_key != expected_key:
        raise HTTPException(status_code=401, detail="Invalid API key")
    return x_api_key

@app.post("/search", dependencies=[Depends(verify_api_key)])
async def search_documentation(request: SearchRequest):
    ...
```

### In MCP Server (`.env`):

```bash
DOCS_API_URL=http://localhost:8000
DOCS_API_KEY=your-secret-api-key
```

The MCP server code already handles this with the `API_KEY` environment variable.

## Deployment

### Central API
- Deploy to AWS ECS, Cloud Run, or any container platform
- Use managed PostgreSQL (RDS, Cloud SQL, etc.)
- Set up HTTPS with a domain name

### MCP Server
- Keep it local (runs on user's machine)
- Or deploy to a server if needed
- Configure `DOCS_API_URL` to point to deployed API

## Next Steps

1. Create your MCP server repository
2. Copy the example code above
3. Customize the tools as needed
4. Add more tools (e.g., `get_doc_by_path`, `list_topics`, etc.)
5. Deploy the central API to production
6. Share the API URL with your team

---

For more information:
- [API Documentation](API_DOCUMENTATION.md)
- [MCP Protocol](https://modelcontextprotocol.io/)

