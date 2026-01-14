# Devtron Central API - Project Overview

## 🎯 What is Central API?

**Devtron Central API** is a Go-based REST API service that serves as a centralized hub for Devtron-related metadata, release information, and auxiliary services. It acts as a backend service that provides essential data to Devtron installations and related tools.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Central API Server                    │
│                      (Port 8080)                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Release    │  │   Module     │  │   Currency   │ │
│  │   Notes      │  │   Metadata   │  │   Exchange   │ │
│  │   Service    │  │   Service    │  │   Service    │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐                    │
│  │   CI/CD      │  │   Webhook    │                    │
│  │   Metadata   │  │   Handler    │                    │
│  │   Service    │  │              │                    │
│  └──────────────┘  └──────────────┘                    │
│                                                          │
└─────────────────────────────────────────────────────────┘
           │                    │                │
           ▼                    ▼                ▼
    ┌──────────┐        ┌──────────┐    ┌──────────┐
    │  GitHub  │        │   Blob   │    │ External │
    │   API    │        │ Storage  │    │   APIs   │
    └──────────┘        └──────────┘    └──────────┘
```

## 📦 Core Services & Use Cases

### 1. **Release Notes Service**

**Purpose**: Manage and serve Devtron release information

**Use Cases**:
- ✅ Fetch latest Devtron releases from GitHub
- ✅ Display release notes in Devtron dashboard
- ✅ Check for updates and new versions
- ✅ Show prerequisite information for upgrades
- ✅ Webhook integration for automatic updates

**API Endpoints**:
- `GET /release/notes` - Get all releases with pagination
- `POST /release/webhook` - GitHub webhook for release events

**How it works**:
1. Fetches releases from GitHub API
2. Caches in memory for performance
3. Stores latest tag in blob storage (S3/Azure/GCP)
4. Auto-updates on GitHub webhook events
5. Serves with pagination support

### 2. **Module Management Service**

**Purpose**: Provide information about Devtron modules and integrations

**Use Cases**:
- ✅ List available Devtron modules (CI/CD, Security, Cost, etc.)
- ✅ Get module metadata and versions
- ✅ Check module compatibility
- ✅ Integration marketplace information

**API Endpoints**:
- `GET /modules` - Get all modules (v1)
- `GET /v2/modules` - Get all modules (v2 with enhanced metadata)
- `GET /module?name={name}` - Get specific module by name

**Module Examples**:
- CI/CD Module
- Security Scanning Module
- Cost Optimization Module
- GitOps Module
- Monitoring Integration

### 3. **CI/CD Build Metadata Service**

**Purpose**: Serve metadata for CI/CD build configurations

**Use Cases**:
- ✅ Provide Dockerfile templates for different languages
- ✅ Buildpack metadata for auto-detection
- ✅ Language-specific build configurations
- ✅ Container image build optimization

**API Endpoints**:
- `GET /dockerfileTemplate` - Get Dockerfile template metadata
- `GET /buildpackMetadata` - Get buildpack metadata

**Supported Languages/Frameworks**:
- Node.js
- Python
- Java
- Go
- PHP
- Ruby
- .NET
- And more...

### 4. **Currency Exchange Service**

**Purpose**: Provide real-time currency exchange rates

**Use Cases**:
- ✅ Cost calculations in different currencies
- ✅ Multi-currency billing support
- ✅ Cloud cost conversions
- ✅ Financial reporting

**API Endpoints**:
- `GET /currency/rates?base={currency}` - Get exchange rates

**Features**:
- Real-time rates from external APIs
- Caching for performance
- Multiple base currency support
- Automatic rate updates

### 5. **Webhook Handler**

**Purpose**: Process GitHub webhooks for automated updates

**Use Cases**:
- ✅ Auto-update release notes on new GitHub releases
- ✅ Trigger cache invalidation
- ✅ Notify connected systems
- ✅ Secure webhook validation

**Security**:
- HMAC signature verification
- Secret-based authentication
- Request validation

## 🔧 Technical Stack

### **Backend**:
- **Language**: Go 1.19+
- **Framework**: Gorilla Mux (HTTP router)
- **DI**: Google Wire (dependency injection)
- **Logging**: Uber Zap

### **Storage**:
- **Blob Storage**: AWS S3 / Azure Blob / GCP Storage
- **Cache**: In-memory (map-based)

### **External Integrations**:
- **GitHub API**: Release data
- **Currency APIs**: Exchange rates
- **Cloud Storage**: Blob persistence

### **Build & Deploy**:
- **Build**: Make + Wire
- **Container**: Docker (Alpine-based)
- **Port**: 8080

## 📊 Data Flow Examples

### Example 1: Getting Latest Release

```
User/Dashboard
    │
    ├─> GET /release/notes
    │
    ▼
Central API
    │
    ├─> Check in-memory cache
    │   └─> If cached: return immediately
    │
    ├─> Check blob storage for latest tag
    │   └─> If same as cache: return cache
    │
    ├─> Fetch from GitHub API
    │   └─> Parse release data
    │   └─> Extract prerequisites
    │
    ├─> Update cache
    ├─> Update blob storage
    │
    └─> Return releases to user
```

### Example 2: GitHub Webhook Flow

```
GitHub Release Event
    │
    ├─> POST /release/webhook
    │   └─> Validate HMAC signature
    │
    ▼
Central API
    │
    ├─> Parse webhook payload
    ├─> Fetch new release from GitHub
    ├─> Update in-memory cache
    ├─> Update blob storage
    │
    └─> Return success
```

## 🚀 Deployment

### **Environment Variables**:
```bash
# Blob Storage (AWS S3 example)
BLOB_STORAGE_PROVIDER=S3
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
AWS_DEFAULT_REGION=us-east-1
AWS_S3_BUCKET_NAME=devtron-central-api

# GitHub
GITHUB_TOKEN=xxx  # For API rate limits

# Webhook
WEBHOOK_SECRET=xxx  # For signature validation
```

### **Running Locally**:
```bash
# Build
make build

# Run
./central-api
```

### **Docker**:
```bash
# Build image
docker build -t central-api:latest .

# Run container
docker run -p 8080:8080 \
  -e BLOB_STORAGE_PROVIDER=S3 \
  -e AWS_ACCESS_KEY_ID=xxx \
  central-api:latest
```

## 📁 Project Structure

```
central-api/
├── api/                    # HTTP handlers and routing
│   ├── RestHandler.go     # Main REST handlers
│   ├── Router.go          # Route definitions
│   ├── currency/          # Currency service handlers
│   └── handler/           # Common handler utilities
├── pkg/                   # Business logic services
│   ├── ReleaseNoteService.go
│   ├── CiBuildMetadataService.go
│   ├── WebhookSecretValidator.go
│   └── currency/          # Currency service logic
├── client/                # External API clients
│   ├── GitHubClient.go
│   ├── ModuleConfig.go
│   └── BlobConfig.go
├── common/                # Shared models and types
│   ├── bean.go
│   ├── BuildpackMetadata.go
│   └── DockerfileTemplateMetadata.go
├── mcp-docs-server/       # MCP server for documentation
│   ├── server.py
│   ├── doc_processor.py
│   ├── vector_store.py
│   └── README.md
├── App.go                 # Application entry point
├── Wire.go                # Dependency injection config
├── main.go                # Main function
└── Dockerfile             # Container definition
```

## 🔌 API Reference

### Health Check
```bash
GET /health
Response: {"code": 200, "result": "OK"}
```

### Release Notes
```bash
GET /release/notes?offset=0&size=10
Response: {
  "code": 200,
  "result": [
    {
      "tagName": "v0.7.0",
      "releaseName": "Devtron v0.7.0",
      "body": "Release notes...",
      "createdAt": "2024-01-01T00:00:00Z",
      "prerequisite": true,
      "prerequisiteMessage": "Upgrade instructions..."
    }
  ]
}
```

### Modules
```bash
GET /modules
Response: {
  "code": 200,
  "result": [
    {"id": 1, "name": "cicd"},
    {"id": 2, "name": "security"}
  ]
}
```

### Currency Rates
```bash
GET /currency/rates?base=USD
Response: {
  "code": 200,
  "result": {
    "base": "USD",
    "rates": {
      "EUR": 0.85,
      "GBP": 0.73,
      "INR": 83.12
    }
  }
}
```

## 🎯 Who Uses This?

1. **Devtron Dashboard**: Displays release notes and updates
2. **Devtron CLI**: Checks for new versions
3. **Devtron Installations**: Fetches module metadata
4. **CI/CD Pipelines**: Gets build templates
5. **Cost Management**: Currency conversions
6. **Integration Tools**: Module discovery

## 🔐 Security

- ✅ CORS enabled for cross-origin requests
- ✅ Webhook signature validation
- ✅ Secure blob storage access
- ✅ No sensitive data in responses
- ✅ Rate limiting (via GitHub token)

## 📈 Performance

- **In-memory caching**: Fast response times
- **Blob storage**: Reduces GitHub API calls
- **Lazy loading**: Only fetch when needed
- **Retry logic**: Resilient to transient failures

## 🆕 Recent Addition: MCP Documentation Server

A new **Model Context Protocol (MCP) server** has been added to provide semantic search over Devtron documentation:

- **Location**: `mcp-docs-server/`
- **Purpose**: Enable chatbots to access Devtron docs
- **Technology**: Python, ChromaDB, AWS Bedrock Titan
- **Features**: Semantic search, auto-sync, incremental updates

See `mcp-docs-server/README.md` for details.

## 📝 License

Apache License 2.0 - Copyright (c) 2024 Devtron Inc.

---

**Maintained by**: Devtron Labs  
**Repository**: https://github.com/devtron-labs/central-api

