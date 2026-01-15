# Complete API Guide - Sample Requests & Responses

## 🚀 Quick Reference

**Base URL**: `http://localhost:8000`

**Available Endpoints**:
- `GET /health` - Health check
- `POST /reindex` - Re-index documentation
- `POST /search` - Semantic search with optional LLM

---

## 📋 Complete Examples

### Example 1: Health Check

**Request:**
```bash
curl -X GET http://localhost:8000/health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "database": "connected",
  "docs_indexed": true
}
```

**When to use**: Check if service is running and database is connected

---

### Example 2: Initial Documentation Indexing

**Request:**
```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{
    "force": true
  }'
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Full re-index completed",
  "documents_processed": 156,
  "changed_files": 12
}
```

**Time**: 5-10 minutes for initial indexing

**When to use**: First time setup or when you want to rebuild the entire index

---

### Example 3: Incremental Update

**Request:**
```bash
curl -X POST http://localhost:8000/reindex \
  -H "Content-Type: application/json" \
  -d '{
    "force": false
  }'
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Incremental update completed",
  "documents_processed": 5,
  "changed_files": 2
}
```

**Time**: 30-60 seconds

**When to use**: Daily/hourly sync to get latest documentation changes

---

### Example 4: Basic Search (No LLM)

**Request:**
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I deploy an application using Devtron?",
    "max_results": 3,
    "use_llm": false
  }'
```

**Response (200 OK):**
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
    }
  ],
  "llm_response": null,
  "total_results": 3
}
```

**Time**: ~200-500ms

**When to use**: Fast lookups, autocomplete, when you just need relevant docs

---

### Example 5: Enhanced Search (With LLM)

**Request:**
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

**Response (200 OK):**
```json
{
  "query": "How do I deploy an application using Devtron?",
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
    },
    {
      "title": "Quick Start Guide",
      "source": "docs/getting-started/README.md",
      "header": "Getting Started",
      "content": "Devtron is a Kubernetes-native application delivery platform...",
      "score": 0.82
    },
    {
      "title": "CI/CD Pipeline Setup",
      "source": "docs/user-guide/creating-application/workflow/ci-pipeline.md",
      "header": "CI Pipeline Configuration",
      "content": "The CI pipeline builds your application from source code...",
      "score": 0.78
    },
    {
      "title": "Environment Configuration",
      "source": "docs/user-guide/global-configurations/cluster-and-environments.md",
      "header": "Managing Environments",
      "content": "Environments in Devtron represent deployment targets...",
      "score": 0.75
    }
  ],
  "llm_response": "# Deploying an Application with Devtron\n\nBased on the Devtron documentation, here's a comprehensive guide on how to deploy an application:\n\n## Prerequisites\n\nBefore deploying an application, ensure you have:\n- Devtron installed on your Kubernetes cluster\n- Access to the Devtron dashboard\n- A Git repository with your application code\n- Docker registry configured\n\n## Step-by-Step Deployment Process\n\n### 1. Navigate to Applications\n- Open the Devtron dashboard\n- Go to the **Applications** section from the main menu\n\n### 2. Create New Application\n- Click on **'Create New'** button\n- Provide a unique application name\n- Select or create a project for organization\n\n### 3. Connect Git Repository\n- Select your Git repository from the connected sources\n- Choose the branch you want to deploy from\n- Specify the build context (usually root directory)\n\n### 4. Configure Build Settings (CI Pipeline)\n\nThe CI pipeline builds your Docker image:\n\n- **Source Code**: Already selected from step 3\n- **Dockerfile**: Specify the path to your Dockerfile\n- **Build Context**: Set the directory for the build\n- **Pre-build Scripts**: Add any scripts to run before building (optional)\n- **Post-build Scripts**: Add scripts to run after building (optional)\n- **Docker Registry**: Select where to push the built image\n\nDevtron supports multiple build strategies:\n- Docker builds\n- Cloud Native Buildpacks\n- Custom build scripts\n\n### 5. Set Deployment Configuration (CD Pipeline)\n\nConfigure how your application will be deployed:\n\n- **Environment**: Select target environment (dev, staging, production)\n- **Namespace**: Choose or create a Kubernetes namespace\n- **Deployment Strategy**: \n  - Rolling update (default)\n  - Blue-green deployment\n  - Canary deployment\n- **Environment Variables**: Set application-specific variables\n- **Secrets**: Configure sensitive data securely\n- **Pre-deployment Hooks**: Scripts to run before deployment (optional)\n- **Post-deployment Hooks**: Scripts to run after deployment (optional)\n\n### 6. Deploy\n\n- Review all configurations\n- Click **'Deploy'** to trigger the deployment pipeline\n- Monitor the deployment progress in real-time\n\n## What Happens During Deployment\n\nWhen you click Deploy, Devtron automatically:\n\n1. **Builds** your Docker image using the CI pipeline\n2. **Pushes** the image to your configured Docker registry\n3. **Deploys** the application to your Kubernetes cluster\n4. **Monitors** the deployment status and health\n5. **Provides** real-time logs and metrics\n\n## Workflow Components\n\nA typical Devtron workflow includes:\n\n- **CI Pipeline**: Builds your Docker image from source code\n- **CD Pipeline**: Deploys the image to Kubernetes environments\n- **Pre/Post Deployment Hooks**: Custom scripts for deployment lifecycle management\n\n## Multi-Environment Deployment\n\nYou can configure multiple environments and set up promotion strategies:\n\n- **Development**: For testing new features\n- **Staging**: For pre-production validation\n- **Production**: For live applications\n\nDevtron allows you to promote applications through these stages with approval workflows.\n\n## Best Practices\n\n1. Start with a development environment first\n2. Test thoroughly before promoting to production\n3. Use environment-specific configurations\n4. Implement proper secret management\n5. Set up monitoring and alerts\n6. Use deployment strategies like canary for production\n\n## Troubleshooting\n\nIf deployment fails:\n- Check the build logs in the CI pipeline\n- Verify deployment logs in the CD pipeline\n- Ensure Kubernetes cluster has sufficient resources\n- Validate environment variables and secrets\n- Check network connectivity and registry access\n\nDevtron provides comprehensive logging and monitoring to help identify and resolve issues quickly.",
  "total_results": 5
}
```

**Time**: ~2-5 seconds (includes LLM processing)

**When to use**: Chatbots, user support, when you need a comprehensive answer

**Note**: Requires AWS Bedrock configuration. If not available, `llm_response` will contain an error message.

---

### Example 6: Search for Specific Topic (RBAC)

**Request:**
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to configure RBAC and user permissions?",
    "max_results": 3,
    "use_llm": true,
    "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
  }'
```

**Response (200 OK):**
```json
{
  "query": "How to configure RBAC and user permissions?",
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
  "llm_response": "# Configuring RBAC and User Permissions in Devtron\n\nDevtron provides comprehensive Role-Based Access Control (RBAC) to manage user permissions effectively. Here's how to configure it:\n\n## Understanding Devtron RBAC\n\nDevtron's RBAC system allows you to:\n- Create custom roles with specific permissions\n- Assign roles to individual users or groups\n- Control access at multiple levels (application, environment, cluster)\n- Integrate with enterprise SSO providers\n- Enforce policies at both API and UI levels\n\n## Setting Up RBAC\n\n### 1. Access Authorization Settings\n\n- Navigate to **Global Configurations** in the Devtron dashboard\n- Click on **Authorization**\n- You'll see options for User Access, Permission Groups, and SSO\n\n### 2. Create Permission Groups\n\nPermission groups allow you to bundle permissions and assign them to multiple users:\n\n**Steps:**\n1. Go to **Global Configurations → Authorization**\n2. Click **'Add Group'**\n3. Define permissions:\n   - **View**: Read-only access\n   - **Create**: Ability to create new resources\n   - **Edit**: Modify existing resources\n   - **Delete**: Remove resources\n4. Assign permissions to specific:\n   - Applications\n   - Environments\n   - Clusters\n5. Add users to the group\n\n### 3. Assign Roles to Users\n\nYou can assign roles at different levels:\n\n**Application Level:**\n- Control who can view, edit, or deploy specific applications\n- Set different permissions for different apps\n\n**Environment Level:**\n- Restrict access to production environments\n- Allow broader access to development/staging\n\n**Cluster Level:**\n- Manage who can access entire Kubernetes clusters\n- Useful for multi-cluster setups\n\n### 4. Configure SSO Integration (Optional)\n\nFor enterprise authentication, Devtron supports multiple SSO providers:\n\n**Supported Providers:**\n- Google OAuth\n- GitHub OAuth\n- GitLab OAuth\n- LDAP/Active Directory\n- SAML 2.0\n\n**Configuration:**\n1. Go to **Global Configurations → Authorization → SSO Login Services**\n2. Select your SSO provider\n3. Enter provider-specific credentials and settings\n4. Test the integration\n5. Enable for your organization\n\n## Permission Types\n\n### Standard Permissions\n\n- **View**: Read-only access to resources\n- **Create**: Create new applications, pipelines, etc.\n- **Edit**: Modify existing configurations\n- **Delete**: Remove resources\n- **Admin**: Full control over resources\n\n### Special Permissions\n\n- **Trigger**: Execute CI/CD pipelines\n- **Approve**: Approve deployment requests\n- **Manage**: Configure settings and integrations\n\n## Best Practices\n\n1. **Principle of Least Privilege**: Grant only necessary permissions\n2. **Use Groups**: Manage permissions through groups rather than individual users\n3. **Separate Environments**: Restrict production access to authorized personnel\n4. **Regular Audits**: Review and update permissions periodically\n5. **SSO Integration**: Use SSO for centralized authentication\n6. **Document Roles**: Maintain documentation of role definitions\n\n## Example RBAC Setup\n\n### Developer Role\n- **Applications**: View, Create, Edit (dev/staging only)\n- **Environments**: Deploy to dev/staging\n- **Clusters**: View only\n\n### DevOps Engineer Role\n- **Applications**: Full access\n- **Environments**: Deploy to all environments\n- **Clusters**: Manage cluster configurations\n\n### Manager Role\n- **Applications**: View all\n- **Environments**: Approve production deployments\n- **Clusters**: View metrics and logs\n\n## Troubleshooting\n\n**Issue**: User can't access application\n- Check if user is assigned to correct permission group\n- Verify permissions are set for the specific application\n- Ensure SSO integration is working (if enabled)\n\n**Issue**: SSO login fails\n- Verify SSO provider credentials\n- Check network connectivity\n- Review SSO provider logs\n\nDevtron's RBAC system provides fine-grained control over user access, ensuring security while maintaining flexibility for your team's workflow.",
  "total_results": 3
}
```

---

### Example 7: Helm Chart Deployment Query

**Request:**
```bash
curl -X POST http://localhost:8000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "helm chart deployment",
    "max_results": 2,
    "use_llm": false
  }'
```

**Response (200 OK):**
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

## 🔧 Integration Examples

### Python Client

```python
import requests

class DevtronDocsClient:
    def __init__(self, base_url="http://localhost:8000"):
        self.base_url = base_url

    def health_check(self):
        response = requests.get(f"{self.base_url}/health")
        return response.json()

    def reindex(self, force=False):
        response = requests.post(
            f"{self.base_url}/reindex",
            json={"force": force}
        )
        return response.json()

    def search(self, query, max_results=5, use_llm=True):
        response = requests.post(
            f"{self.base_url}/search",
            json={
                "query": query,
                "max_results": max_results,
                "use_llm": use_llm,
                "llm_model": "anthropic.claude-3-haiku-20240307-v1:0"
            }
        )
        return response.json()

# Usage
client = DevtronDocsClient()

# Check health
print(client.health_check())

# Search
result = client.search("How to deploy applications?")
print(f"Found {result['total_results']} results")
if result['llm_response']:
    print(result['llm_response'])
```

### JavaScript/Node.js Client

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

  async search(query, maxResults = 5, useLLM = true) {
    const { data } = await this.client.post('/search', {
      query,
      max_results: maxResults,
      use_llm: useLLM,
      llm_model: 'anthropic.claude-3-haiku-20240307-v1:0'
    });
    return data;
  }
}

// Usage
const client = new DevtronDocsClient();

(async () => {
  // Check health
  const health = await client.healthCheck();
  console.log('Health:', health);

  // Search
  const result = await client.search('How to deploy applications?');
  console.log(`Found ${result.total_results} results`);
  if (result.llm_response) {
    console.log(result.llm_response);
  }
})();
```

---

## 📊 Response Time Comparison

| Search Type | Avg Time | Use Case |
|-------------|----------|----------|
| No LLM | 200-500ms | Fast lookups, autocomplete |
| With LLM (Haiku) | 2-3s | Chatbots, detailed answers |
| With LLM (Sonnet) | 4-6s | Complex queries, analysis |

---

## ⚠️ Error Responses

### 400 - Documentation Not Indexed
```json
{
  "detail": "Documentation not indexed. Please call /reindex first."
}
```

**Solution**: Run `/reindex` endpoint first

### 500 - Search Failed
```json
{
  "detail": "Search failed: Connection to database lost"
}
```

**Solution**: Check database connectivity

### 503 - Service Unhealthy
```json
{
  "detail": "Service unhealthy: Unable to connect to PostgreSQL database"
}
```

**Solution**: Verify PostgreSQL is running

---

## 📚 Additional Resources

- **Quick Start**: See `QUICK_START.md`
- **API Flow Diagrams**: See `API_FLOW.md`
- **Detailed Examples**: See `API_EXAMPLES.md`
- **Main Documentation**: See `README.md`

---

## ✅ Testing Checklist

- [ ] Health check returns `"status": "healthy"`
- [ ] Re-index completes successfully
- [ ] Search without LLM returns results
- [ ] Search with LLM returns enhanced response
- [ ] Incremental update works
- [ ] Error handling works correctly

---

**Last Updated**: 2026-01-15


