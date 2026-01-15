# Devtron Documentation API - Sample Requests & Responses

This document provides sample API requests and responses for the Devtron Documentation RAG Server.

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

**Description:** Perform semantic search over Devtron documentation with optional LLM-enhanced responses.

#### Request - Basic Search (with LLM)
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application using Devtron?",
    "max_results": 5,
    "use_llm": true,
    "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
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
  "llm_response": "# Deploying an Application with Devtron\n\nBased on the documentation, here's how to deploy an application using Devtron:\n\n## Step-by-Step Process\n\n1. **Navigate to Applications**: Go to the Applications section in the Devtron dashboard\n\n2. **Create New Application**: Click on 'Create New' to start the application creation process\n\n3. **Connect Git Repository**: Select and connect your Git repository containing the application source code\n\n4. **Configure Build Settings**: Set up your CI pipeline by configuring:\n   - Source code repository details\n   - Build context and Dockerfile location\n   - Pre-build and post-build scripts (if needed)\n   - Docker registry for storing built images\n\n5. **Set Deployment Configuration**: Configure your CD pipeline:\n   - Select target environment (dev, staging, production)\n   - Configure environment-specific values and secrets\n   - Set up pre/post deployment hooks if required\n\n6. **Deploy**: Click 'Deploy' to trigger the deployment\n\n## What Happens Next\n\nDevtron will automatically:\n- Build your Docker image using the CI pipeline\n- Push the image to your configured registry\n- Deploy the application to your Kubernetes cluster\n- Monitor the deployment status\n\n## Workflow Components\n\nA typical Devtron workflow includes:\n- **CI Pipeline**: Builds your Docker image from source\n- **CD Pipeline**: Deploys to Kubernetes environments\n- **Pre/Post Deployment Hooks**: Custom scripts for deployment lifecycle\n\nYou can configure multiple environments and set up promotion strategies to move applications through different stages (dev → staging → production).\n\nDevtron supports multiple build strategies including Docker, Cloud Native Buildpacks, and custom build scripts.",
  "total_results": 5
}
```

#### Request - Search Without LLM
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

#### Request - Limited Results
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "helm chart deployment",
    "max_results": 2,
    "use_llm": true
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
  "llm_response": "# Deploying Helm Charts in Devtron\n\nDevtron provides comprehensive support for Helm chart deployments. Here's what you need to know:\n\n## Supported Helm Chart Sources\n\nDevtron can deploy Helm charts from:\n- **Public Helm Repositories**: Bitnami, Stable charts, and other public repos\n- **Private Helm Repositories**: Your organization's private chart repositories\n- **Git Repositories**: Charts stored in Git\n- **OCI Registries**: Charts stored in OCI-compliant registries\n\n## Deployment Process\n\n1. **Access Chart Store**: Navigate to the Chart Store in Devtron\n2. **Search for Chart**: Find the Helm chart you want to deploy\n3. **Click Deploy**: Initiate the deployment process\n4. **Configure Values**: Customize the chart values for your use case\n5. **Select Environment**: Choose the target environment\n6. **Deploy**: Execute the deployment\n\n## Chart Store Features\n\nThe Chart Store provides:\n- Curated collection of popular Helm charts\n- Chart details and version information\n- Custom values configuration\n- Multi-environment deployment support\n- Management of deployed chart instances\n\nYou can deploy the same chart to multiple environments with different configurations, making it easy to maintain consistency across dev, staging, and production.",
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

## Python Client Example

```python
import requests
import json

BASE_URL = "http://localhost:8000"

# Health check
response = requests.get(f"{BASE_URL}/health")
print("Health:", response.json())

# Re-index documentation
reindex_data = {"force": False}
response = requests.post(f"{BASE_URL}/reindex", json=reindex_data)
print("Reindex:", response.json())

# Search with LLM
search_data = {
    "query": "How do I set up CI/CD pipeline?",
    "max_results": 5,
    "use_llm": True,
    "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
}
response = requests.post(f"{BASE_URL}/search", json=search_data)
result = response.json()

print(f"\nQuery: {result['query']}")
print(f"Total Results: {result['total_results']}\n")

for i, doc in enumerate(result['results'], 1):
    print(f"{i}. {doc['title']} (Score: {doc['score']})")
    print(f"   Source: {doc['source']}")
    print(f"   {doc['content'][:100]}...\n")

if result['llm_response']:
    print("LLM Response:")
    print(result['llm_response'])
```

---

## JavaScript/Node.js Client Example

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:8000';

async function searchDocs() {
  try {
    // Health check
    const health = await axios.get(`${BASE_URL}/health`);
    console.log('Health:', health.data);

    // Search documentation
    const searchResponse = await axios.post(`${BASE_URL}/search`, {
      query: 'How to configure environment variables?',
      max_results: 5,
      use_llm: true,
      llm_model: 'anthropic.claude-3-haiku-20240307-v1:0'
    });

    const { query, results, llm_response, total_results } = searchResponse.data;

    console.log(`\nQuery: ${query}`);
    console.log(`Total Results: ${total_results}\n`);

    results.forEach((doc, index) => {
      console.log(`${index + 1}. ${doc.title} (Score: ${doc.score})`);
      console.log(`   Source: ${doc.source}`);
      console.log(`   ${doc.content.substring(0, 100)}...\n`);
    });

    if (llm_response) {
      console.log('LLM Response:');
      console.log(llm_response);
    }
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
  }
}

searchDocs();
```

---

## cURL Examples Collection

### Complete Workflow
```bash
# 1. Check health
curl -X GET http://localhost:8000/health

# 2. Initial indexing
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": true}'

# 3. Search without LLM (faster)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "kubernetes deployment",
    "max_results": 3,
    "use_llm": false
  }'

# 4. Search with LLM (comprehensive answer)
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to troubleshoot failed deployments?",
    "max_results": 5,
    "use_llm": true,
    "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
  }'

# 5. Incremental update (daily sync)
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{"force": false}'
```

---

## Notes

1. **LLM Availability**: LLM responses require AWS Bedrock configuration. If not available, `llm_response` will contain an error message.

2. **Search Scores**: Scores range from 0.0 to 1.0, with higher scores indicating better semantic similarity.

3. **Max Results**: Limited to 20 results per request to ensure performance.

4. **Re-indexing**: Incremental updates are faster and recommended for regular syncs. Use `force: true` only when needed.

5. **Performance**: Search typically completes in <500ms. LLM responses add 2-5 seconds depending on the model.


