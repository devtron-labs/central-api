# MCP Tool Integration Guide for Athena-BE

## 🎯 Recommended Architecture

### Why NOT to use `use_llm=true` in the RAG API

When integrating with Athena-BE (which already has LLM capabilities), you should **NOT** use the RAG API's built-in LLM feature. Here's why:

#### ❌ Problem with Double LLM Processing

```
User Query
    ↓
Athena-BE
    ↓
RAG API (use_llm=true) ← LLM Call #1 (costs tokens)
    ↓
Returns enhanced response
    ↓
Athena-BE processes further ← LLM Call #2 (costs MORE tokens)
    ↓
Final response to user

Result: DOUBLE TOKEN CONSUMPTION! 💸💸
```

#### ✅ Recommended Approach

```
User Query
    ↓
Athena-BE
    ↓
RAG API (use_llm=false) ← Just vector search (fast, no LLM cost)
    ↓
Returns raw search results
    ↓
Athena-BE formats context + calls LLM ← LLM Call (single token usage)
    ↓
Final response to user

Result: SINGLE TOKEN CONSUMPTION! ✅
```

---

## 🔐 AWS Credentials Configuration

The RAG API uses AWS Bedrock for LLM (when `use_llm=true`). Authentication is handled via:

### Option 1: Environment Variables (Recommended for Docker)
```bash
# In .env file or docker-compose.yml
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here
```

### Option 2: AWS Profile (Recommended for Local Development)
```bash
# In .env file
AWS_REGION=us-east-1
AWS_PROFILE=default  # Uses ~/.aws/credentials
```

### Option 3: IAM Role (Recommended for Production)
When running on AWS (ECS, EKS, EC2), attach an IAM role with Bedrock permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:aws:bedrock:*::foundation-model/anthropic.claude-*"
      ]
    }
  ]
}
```

**Note:** For Athena-BE integration, you likely **don't need** to configure AWS credentials in the RAG API since you'll use `use_llm=false`.

---

## 🛠️ MCP Tool Implementation

### Recommended MCP Tool Structure

```python
# In Athena-BE MCP tool

import requests
from typing import List, Dict

class DevtronDocsTool:
    """MCP Tool for searching Devtron documentation."""
    
    def __init__(self, rag_api_url: str = "http://localhost:8000"):
        self.rag_api_url = rag_api_url
    
    def search_docs(self, query: str, max_results: int = 5) -> List[Dict]:
        """
        Search Devtron documentation using vector similarity.
        
        Args:
            query: User's search query
            max_results: Maximum number of results to return
            
        Returns:
            List of relevant documentation chunks with metadata
        """
        response = requests.post(
            f"{self.rag_api_url}/search",
            json={
                "query": query,
                "max_results": max_results,
                "use_llm": False  # ← Important: Let Athena-BE handle LLM
            }
        )
        
        if response.status_code != 200:
            raise Exception(f"Search failed: {response.text}")
        
        data = response.json()
        return data["results"]
    
    def format_context_for_llm(self, search_results: List[Dict]) -> str:
        """
        Format search results into context for LLM.
        
        Args:
            search_results: Results from search_docs()
            
        Returns:
            Formatted context string for LLM prompt
        """
        if not search_results:
            return "No relevant documentation found."
        
        context_parts = []
        for i, result in enumerate(search_results, 1):
            context_parts.append(
                f"[Document {i}]\n"
                f"Title: {result['title']}\n"
                f"Source: {result['source']}\n"
                f"Section: {result.get('header', 'N/A')}\n"
                f"Relevance Score: {result['score']:.2f}\n"
                f"Content:\n{result['content']}\n"
            )
        
        return "\n---\n".join(context_parts)
    
    def answer_question(self, query: str, llm_client) -> str:
        """
        Answer user question using RAG + LLM.
        
        Args:
            query: User's question
            llm_client: Athena-BE's LLM client
            
        Returns:
            LLM-generated answer based on documentation
        """
        # Step 1: Get relevant docs from RAG API
        search_results = self.search_docs(query, max_results=5)
        
        if not search_results:
            return "I couldn't find relevant documentation for your question."
        
        # Step 2: Format context
        context = self.format_context_for_llm(search_results)
        
        # Step 3: Create prompt for LLM
        prompt = f"""You are a helpful assistant for Devtron, a Kubernetes application delivery platform.

User Question: {query}

Relevant Documentation:
{context}

Instructions:
- Answer the user's question based ONLY on the provided documentation
- Be specific and include step-by-step instructions when applicable
- If the documentation doesn't contain enough information, say so
- Format your response in markdown
- Include relevant examples or commands if present in the documentation

Answer:"""
        
        # Step 4: Call Athena-BE's LLM (single token usage)
        response = llm_client.generate(prompt)
        
        return response


# Usage in Athena-BE
tool = DevtronDocsTool(rag_api_url="http://docs-rag-api:8000")

# When user asks a question
user_query = "How do I deploy an application in Devtron?"
answer = tool.answer_question(user_query, athena_llm_client)
print(answer)
```

---

## 📊 Performance & Cost Comparison

### Scenario: User asks "How to deploy applications?"

#### ❌ Using `use_llm=true` (Not Recommended)

| Step | Service | Action | Tokens | Cost | Time |
|------|---------|--------|--------|------|------|
| 1 | RAG API | Vector search | 0 | $0 | 200ms |
| 2 | RAG API | LLM call #1 | ~2000 | $0.005 | 2s |
| 3 | Athena-BE | LLM call #2 | ~3000 | $0.0075 | 3s |
| **Total** | | | **5000** | **$0.0125** | **5.2s** |

#### ✅ Using `use_llm=false` (Recommended)

| Step | Service | Action | Tokens | Cost | Time |
|------|---------|--------|--------|------|------|
| 1 | RAG API | Vector search | 0 | $0 | 200ms |
| 2 | Athena-BE | LLM call | ~3000 | $0.0075 | 3s |
| **Total** | | | **3000** | **$0.0075** | **3.2s** |

**Savings:** 40% tokens, 40% cost, 38% faster! 🎉

---

## 🚀 Quick Start for Athena-BE Integration

### 1. Start the RAG API
```bash
cd devtron-docs-rag-server
docker-compose up -d
```

### 2. Index Documentation (One-time)
```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

### 3. Test Search (No LLM)
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to deploy applications?",
    "max_results": 5,
    "use_llm": false
  }'
```

### 4. Integrate in Athena-BE
Use the `DevtronDocsTool` class from above, or create your own MCP tool wrapper.

---

## 🔧 Configuration for Athena-BE

### Environment Variables

```bash
# In Athena-BE .env or config
DEVTRON_DOCS_RAG_API_URL=http://docs-rag-api:8000
DEVTRON_DOCS_MAX_RESULTS=5
DEVTRON_DOCS_MIN_SCORE=0.7  # Filter results below this score
```

### Docker Compose Integration

```yaml
# In Athena-BE docker-compose.yml
services:
  athena-be:
    # ... your existing config
    environment:
      - DEVTRON_DOCS_RAG_API_URL=http://docs-rag-api:8000
    depends_on:
      - docs-rag-api
  
  docs-rag-api:
    image: devtron-docs-rag-server:latest
    ports:
      - "8000:8000"
    environment:
      - POSTGRES_HOST=postgres
      - POSTGRES_DB=devtron_docs
      # No AWS credentials needed if use_llm=false
```

---

## 📝 Example API Responses

### Search Response (use_llm=false)

```json
{
  "query": "How to deploy applications?",
  "results": [
    {
      "title": "Deploying Applications",
      "source": "docs/user-guide/deploying-application/README.md",
      "header": "Deploy Application",
      "content": "To deploy an application in Devtron:\n\n1. Navigate to Applications\n2. Click 'Create New'\n3. Select your Git repository...",
      "score": 0.89
    },
    {
      "title": "Application Deployment Guide",
      "source": "docs/user-guide/creating-application/workflow/README.md",
      "header": "Workflow Configuration",
      "content": "Workflows in Devtron define how your application is built and deployed...",
      "score": 0.85
    }
  ],
  "llm_response": null,
  "total_results": 2
}
```

**What Athena-BE should do:**
1. Extract `results` array
2. Format into context for your LLM
3. Call your LLM with the context
4. Return enhanced response to user

---

## ⚠️ Important Notes

1. **Always use `use_llm=false`** when calling from Athena-BE
2. **No AWS credentials needed** in RAG API if you're not using its LLM
3. **Filter by score** - Results with score < 0.7 may not be relevant
4. **Combine with other sources** - You can merge docs with other context in Athena-BE
5. **Cache results** - Consider caching frequent queries to reduce latency

---

## 🎯 Summary

**For Athena-BE MCP Tool:**
- ✅ Use `use_llm=false` in all requests
- ✅ Let Athena-BE handle LLM processing
- ✅ No AWS credentials needed in RAG API
- ✅ Saves tokens, cost, and latency
- ✅ More flexible for combining multiple sources

**The RAG API's LLM feature (`use_llm=true`) is useful for:**
- Standalone applications without LLM capabilities
- Direct API consumers (CLI tools, simple bots)
- Testing/debugging the search quality

---

**Last Updated:** 2026-01-15

