# Devtron Documentation API - Sample Requests & Responses

This document provides sample API requests and responses for the Devtron Documentation RAG Server.

## ⚠️ Important for Athena-BE Integration

**If you're integrating with Athena-BE (or any service with LLM capabilities):**
- ✅ **Always use `use_llm=false`** to avoid double token consumption
- ✅ See [MCP_INTEGRATION_GUIDE.md](./MCP_INTEGRATION_GUIDE.md) for integration details
- ✅ See [ARCHITECTURE_DECISION.md](./ARCHITECTURE_DECISION.md) for cost/performance analysis

## Base URL
```
http://localhost:8000
```

## API Endpoints

### 1. Health Check

**Endpoint:** `GET /health`

**Description:** Check the health status of the API and database connection.

#### Request
```bash
curl -X GET http://localhost:8000/health
```

#### Response (200 OK)
```json
{
  "status": "healthy",
  "database": "connected",
  "docs_indexed": true
}
```

#### Response when not indexed (200 OK)
```json
{
  "status": "healthy",
  "database": "connected",
  "docs_indexed": false
}
```

---

### 2. Re-index Documentation

**Endpoint:** `POST /reindex`

**Description:** Sync and re-index documentation from GitHub repository.

#### Request - Incremental Update
```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{
    "force": false
  }'
```

#### Response (200 OK)
```json
{
  "status": "success",
  "message": "Incremental update completed",
  "documents_processed": 23,
  "changed_files": 5
}
```

#### Request - Force Full Re-index
```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{
    "force": true
  }'
```

#### Response (200 OK)
```json
{
  "status": "success",
  "message": "Full re-index completed",
  "documents_processed": 156,
  "changed_files": 12
}
```

#### Response - No Changes Detected
```json
{
  "status": "success",
  "message": "No changes detected, index is up to date",
  "documents_processed": 0,
  "changed_files": 0
}
```

---

### 3. Search Documentation

**Endpoint:** `POST /search`

**Description:** Perform semantic search over Devtron documentation. Returns relevant documentation chunks based on vector similarity.

**Recommended:** Use `use_llm=false` for MCP tool integration with Athena-BE to avoid double token consumption.

#### Request - Basic Search (Recommended for Athena-BE)
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application using Devtron?",
    "max_results": 5,
    "use_llm": false
  }'
```

#### Response (200 OK)

```json
{
  "query": "How do I deploy an application using Devtron?",
  "results": [
    {
      "title": "Deploying Applications",
      "source": "docs/user-guide/deploying-application/README.md",
      "header": "Deploy Application",
      "content": "To deploy an application in Devtron:\n\n1. Navigate to Applications\n2. Click 'Create New'\n3. Select your Git repository\n4. Configure build settings\n5. Set deployment configuration\n6. Click 'Deploy'\n\nDevtron will automatically build and deploy your application to the configured Kubernetes cluster.",
      "score": 0.89
    },
    {
      "title": "Application Deployment Guide",
      "source": "docs/user-guide/creating-application/workflow/README.md",
      "header": "Workflow Configuration",
      "content": "Workflows in Devtron define how your application is built and deployed. A typical workflow includes:\n\n- CI Pipeline: Builds your Docker image\n- CD Pipeline: Deploys to Kubernetes\n- Pre/Post deployment hooks\n\nYou can configure multiple environments and promotion strategies.",
      "score": 0.85
    },
    {
      "title": "Quick Start Guide",
      "source": "docs/getting-started/README.md",
      "header": "Getting Started",
      "content": "Devtron is a Kubernetes-native application delivery platform. To get started:\n\n1. Install Devtron on your cluster\n2. Connect your Git repositories\n3. Create your first application\n4. Configure CI/CD pipelines\n5. Deploy to your environments",
      "score": 0.82
    },
    {
      "title": "CI/CD Pipeline Setup",
      "source": "docs/user-guide/creating-application/workflow/ci-pipeline.md",
      "header": "CI Pipeline Configuration",
      "content": "The CI pipeline builds your application from source code. Configure:\n\n- Source code repository\n- Build context and Dockerfile\n- Pre-build and post-build scripts\n- Docker registry for image storage\n\nDevtron supports multiple build strategies including Docker, Buildpacks, and custom scripts.",
      "score": 0.78
    },
    {
      "title": "Environment Configuration",
      "source": "docs/user-guide/global-configurations/cluster-and-environments.md",
      "header": "Managing Environments",
      "content": "Environments in Devtron represent deployment targets (dev, staging, production). Each environment is associated with a Kubernetes namespace and cluster. You can configure environment-specific values and secrets.",
      "score": 0.75
    }
  ],
  "llm_response": null,
  "total_results": 5
}
```

**Note:** `llm_response` is `null` when `use_llm=false`. Process these results in Athena-BE with your LLM to generate enhanced responses.

---

#### Request - RBAC Configuration Search

```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to configure RBAC in Devtron?",
    "max_results": 3,
    "use_llm": false
  }'
```

#### Response (200 OK)

```json
{
  "query": "How to configure RBAC in Devtron?",
  "results": [
    {
      "title": "User Access Management",
      "source": "docs/user-guide/global-configurations/authorization/user-access.md",
      "header": "RBAC Configuration",
      "content": "Devtron provides Role-Based Access Control (RBAC) to manage user permissions. You can:\n\n- Create custom roles with specific permissions\n- Assign roles to users or groups\n- Control access at application, environment, or cluster level\n- Integrate with SSO providers (OAuth, LDAP, SAML)\n\nRBAC policies are enforced at both API and UI levels.",
      "score": 0.92
    },
    {
      "title": "Permission Groups",
      "source": "docs/user-guide/global-configurations/authorization/permission-groups.md",
      "header": "Creating Permission Groups",
      "content": "Permission groups allow you to bundle permissions and assign them to multiple users. To create a permission group:\n\n1. Go to Global Configurations → Authorization\n2. Click 'Add Group'\n3. Define permissions (View, Create, Edit, Delete)\n4. Assign to applications/environments\n5. Add users to the group",
      "score": 0.88
    },
    {
      "title": "SSO Integration",
      "source": "docs/user-guide/global-configurations/authorization/sso/README.md",
      "header": "Single Sign-On Setup",
      "content": "Devtron supports SSO integration for enterprise authentication. Supported providers:\n\n- Google OAuth\n- GitHub OAuth\n- GitLab OAuth\n- LDAP/Active Directory\n- SAML 2.0\n\nConfigure SSO in Global Configurations → Authorization → SSO Login Services.",
      "score": 0.81
    }
  ],
  "llm_response": null,
  "total_results": 3
}
```

---

#### Request - Helm Chart Deployment

```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "helm chart deployment",
    "max_results": 2,
    "use_llm": false
  }'
```

#### Response (200 OK)

```json
{
  "query": "helm chart deployment",
  "results": [
    {
      "title": "Helm Chart Deployment",
      "source": "docs/user-guide/deploying-application/deploying-helm-charts.md",
      "header": "Deploy Helm Charts",
      "content": "Devtron supports deploying Helm charts from various sources:\n\n- Public Helm repositories (Bitnami, Stable, etc.)\n- Private Helm repositories\n- Git repositories containing Helm charts\n- OCI registries\n\nTo deploy a Helm chart:\n1. Go to Chart Store\n2. Search for your chart\n3. Click 'Deploy'\n4. Configure values\n5. Select environment and deploy",
      "score": 0.94
    },
    {
      "title": "Chart Store",
      "source": "docs/user-guide/deploying-application/chart-store.md",
      "header": "Using Chart Store",
      "content": "The Chart Store provides a curated collection of Helm charts. You can:\n\n- Browse available charts\n- View chart details and versions\n- Deploy charts with custom values\n- Manage deployed chart instances\n\nCharts can be deployed to multiple environments with different configurations.",
      "score": 0.87
    }
  ],
  "llm_response": null,
  "total_results": 2
}
```

---

### Error Responses

#### 400 Bad Request - Documentation Not Indexed
```json
{
  "detail": "Documentation not indexed. Please call /reindex first."
}
```

#### 500 Internal Server Error - Search Failed
```json
{
  "detail": "Search failed: Connection to database lost"
}
```

#### 503 Service Unavailable - Health Check Failed
```json
{
  "detail": "Service unhealthy: Unable to connect to PostgreSQL database"
}
```

---

## Python Client Example (Recommended for Athena-BE)

```python
import requests
import json

BASE_URL = "http://localhost:8000"

class DevtronDocsClient:
    """Client for Devtron Documentation RAG API."""

    def __init__(self, base_url: str = "http://localhost:8000"):
        self.base_url = base_url

    def health_check(self):
        """Check API health status."""
        response = requests.get(f"{self.base_url}/health")
        return response.json()

    def reindex(self, force: bool = False):
        """Re-index documentation from GitHub."""
        response = requests.post(
            f"{self.base_url}/reindex",
            json={"force": force}
        )
        return response.json()

    def search(self, query: str, max_results: int = 5):
        """
        Search documentation (without LLM).
        Returns raw results for processing in Athena-BE.
        """
        response = requests.post(
            f"{self.base_url}/search",
            json={
                "query": query,
                "max_results": max_results,
                "use_llm": False  # Let Athena-BE handle LLM
            }
        )
        return response.json()


# Usage Example
client = DevtronDocsClient()

# 1. Health check
health = client.health_check()
print("Health:", health)

# 2. Re-index (if needed)
if not health.get("docs_indexed"):
    print("Indexing documentation...")
    reindex_result = client.reindex(force=True)
    print("Reindex:", reindex_result)

# 3. Search documentation
query = "How do I set up CI/CD pipeline?"
result = client.search(query, max_results=5)

print(f"\nQuery: {result['query']}")
print(f"Total Results: {result['total_results']}\n")

# Display results
for i, doc in enumerate(result['results'], 1):
    print(f"{i}. {doc['title']} (Score: {doc['score']:.2f})")
    print(f"   Source: {doc['source']}")
    print(f"   Header: {doc.get('header', 'N/A')}")
    print(f"   Content: {doc['content'][:150]}...\n")

# 4. Now process with Athena-BE's LLM
# Format context for LLM
context = "\n\n---\n\n".join([
    f"[Document {i+1}]\n"
    f"Title: {doc['title']}\n"
    f"Source: {doc['source']}\n"
    f"Content:\n{doc['content']}"
    for i, doc in enumerate(result['results'])
])

print("Context prepared for Athena-BE LLM:")
print(f"Total context length: {len(context)} characters")

# Send to Athena-BE's LLM (pseudo-code)
# athena_llm_response = athena_llm.generate(
#     prompt=f"Question: {query}\n\nContext:\n{context}\n\nAnswer:"
# )
```

---

## JavaScript/Node.js Client Example (Recommended for Athena-BE)

```javascript
const axios = require('axios');

class DevtronDocsClient {
  constructor(baseURL = 'http://localhost:8000') {
    this.client = axios.create({ baseURL });
  }

  async healthCheck() {
    const { data } = await this.client.get('/health');
    return data;
  }

  async reindex(force = false) {
    const { data } = await this.client.post('/reindex', { force });
    return data;
  }

  async search(query, maxResults = 5) {
    /**
     * Search documentation without LLM.
     * Returns raw results for processing in Athena-BE.
     */
    const { data } = await this.client.post('/search', {
      query,
      max_results: maxResults,
      use_llm: false  // Let Athena-BE handle LLM
    });
    return data;
  }

  formatContextForLLM(results) {
    /**
     * Format search results into context for LLM.
     */
    return results.map((doc, index) =>
      `[Document ${index + 1}]\n` +
      `Title: ${doc.title}\n` +
      `Source: ${doc.source}\n` +
      `Content:\n${doc.content}`
    ).join('\n\n---\n\n');
  }
}

// Usage Example
async function main() {
  try {
    const client = new DevtronDocsClient();

    // 1. Health check
    const health = await client.healthCheck();
    console.log('Health:', health);

    // 2. Re-index if needed
    if (!health.docs_indexed) {
      console.log('Indexing documentation...');
      const reindexResult = await client.reindex(true);
      console.log('Reindex:', reindexResult);
    }

    // 3. Search documentation
    const query = 'How to configure environment variables?';
    const result = await client.search(query, 5);

    console.log(`\nQuery: ${result.query}`);
    console.log(`Total Results: ${result.total_results}\n`);

    // Display results
    result.results.forEach((doc, index) => {
      console.log(`${index + 1}. ${doc.title} (Score: ${doc.score.toFixed(2)})`);
      console.log(`   Source: ${doc.source}`);
      console.log(`   Header: ${doc.header || 'N/A'}`);
      console.log(`   Content: ${doc.content.substring(0, 150)}...\n`);
    });

    // 4. Format context for Athena-BE's LLM
    const context = client.formatContextForLLM(result.results);
    console.log('Context prepared for Athena-BE LLM:');
    console.log(`Total context length: ${context.length} characters`);

    // Send to Athena-BE's LLM (pseudo-code)
    // const athenaResponse = await athenaLLM.generate({
    //   prompt: `Question: ${query}\n\nContext:\n${context}\n\nAnswer:`
    // });

  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
  }
}

main();
```

---

## cURL Examples Collection

### Complete Workflow (Recommended for Athena-BE)

```bash
# 1. Check health
curl -X GET http://localhost:8000/health

# 2. Initial indexing (one-time)
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'

# 3. Search for deployment docs (no LLM)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "kubernetes deployment",
    "max_results": 5,
    "use_llm": false
  }'

# 4. Search for troubleshooting docs (no LLM)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to troubleshoot failed deployments?",
    "max_results": 5,
    "use_llm": false
  }'

# 5. Search for CI/CD pipeline docs (no LLM)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "CI/CD pipeline configuration",
    "max_results": 3,
    "use_llm": false
  }'

# 6. Incremental update (daily/hourly sync)
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": false}'
```

---

## Notes

1. **Recommended for Athena-BE**: Always use `use_llm=false` to avoid double token consumption. Process results in Athena-BE with your LLM.

2. **Search Scores**: Scores range from 0.0 to 1.0, with higher scores indicating better semantic similarity. Filter results with score < 0.7 if needed.

3. **Max Results**: Limited to 20 results per request to ensure performance. Recommended: 3-5 results for optimal LLM context.

4. **Re-indexing**:
   - Initial: `force: true` (5-10 minutes for ~150 docs)
   - Incremental: `force: false` (30-60 seconds, only changed files)
   - Schedule incremental updates hourly or daily

5. **Performance**:
   - Search (no LLM): <500ms
   - Network transfer: ~50ms
   - Total for Athena-BE: ~550ms + your LLM processing time

6. **Context Preparation**: Take the `results` array and format it for your LLM. See Python/JavaScript examples above.

7. **No AWS Credentials Needed**: When using `use_llm=false`, you don't need to configure AWS Bedrock credentials in this API.
