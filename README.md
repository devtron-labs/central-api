# Devtron Central API

A centralized REST API service for Devtron metadata, release information, and auxiliary services.

## 📚 Table of Contents

- [Overview](#overview)
- [Services](#services)
- [MCP Documentation Server](#mcp-documentation-server)
- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Documentation](#documentation)

## 🎯 Overview

**Devtron Central API** is a Go-based REST API that provides:
- 📦 Release notes and version information
- 🔧 Module metadata and configurations
- 🏗️ CI/CD build templates and metadata
- 💱 Currency exchange rates
- 🔔 GitHub webhook handling

**Port**: 8080
**Language**: Go 1.19+
**Framework**: Gorilla Mux

For detailed information, see [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)

## 🚀 Services

### 1. Release Notes Service
Manages Devtron releases from GitHub with caching and blob storage.

**Endpoints**:
- `GET /release/notes` - Get releases with pagination
- `POST /release/webhook` - GitHub webhook handler

### 2. Module Management
Provides Devtron module information and metadata.

**Endpoints**:
- `GET /modules` - List all modules
- `GET /v2/modules` - Enhanced module list
- `GET /module?name={name}` - Get module by name

### 3. CI/CD Metadata
Serves build templates and buildpack information.

**Endpoints**:
- `GET /dockerfileTemplate` - Dockerfile templates
- `GET /buildpackMetadata` - Buildpack metadata

### 4. Currency Exchange
Real-time currency conversion rates.

**Endpoints**:
- `GET /currency/rates?base={currency}` - Exchange rates

### 5. Health Check
Service health monitoring.

**Endpoints**:
- `GET /health` - Health status

## 🤖 MCP Documentation Server

**NEW**: A Model Context Protocol (MCP) server for semantic search over Devtron documentation.

### Features
- 🔍 Semantic search using AWS Bedrock Titan embeddings
- 📦 ChromaDB vector storage
- 🔄 Auto-sync with GitHub documentation
- 💰 Free tier (AWS Bedrock)
- ⚡ Fast (<500ms search)

### Quick Start

```bash
cd mcp-docs-server
./setup.sh
python server.py
```

### Documentation
- [Quick Start Guide](mcp-docs-server/QUICKSTART.md) - 5-minute setup
- [Integration Guide](mcp-docs-server/INTEGRATION_GUIDE.md) - Chatbot integration
- [Solution Summary](mcp-docs-server/SOLUTION_SUMMARY.md) - Architecture details
- [Full README](mcp-docs-server/README.md) - Complete documentation

## 🏃 Quick Start

### Central API (Go)

```bash
# Build
make build

# Run
./central-api
```

### With Docker

```bash
docker build -t central-api:latest .
docker run -p 8080:8080 central-api:latest
```

## 📡 API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Get Releases
```bash
curl http://localhost:8080/release/notes?offset=0&size=10
```

### Get Modules
```bash
curl http://localhost:8080/modules
```

### Get Currency Rates
```bash
curl http://localhost:8080/currency/rates?base=USD
```

For complete API documentation, see [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)

## 📖 Documentation

### Central API
- [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md) - Complete project overview
- [spec/api.yaml](spec/api.yaml) - OpenAPI specification

### MCP Documentation Server
- [QUICKSTART.md](mcp-docs-server/QUICKSTART.md) - 5-minute setup
- [README.md](mcp-docs-server/README.md) - User guide
- [INTEGRATION_GUIDE.md](mcp-docs-server/INTEGRATION_GUIDE.md) - Integration instructions
- [SOLUTION_SUMMARY.md](mcp-docs-server/SOLUTION_SUMMARY.md) - Architecture
- [ALTERNATIVES_COMPARISON.md](mcp-docs-server/ALTERNATIVES_COMPARISON.md) - Solution comparison
- [FILES_OVERVIEW.md](mcp-docs-server/FILES_OVERVIEW.md) - File reference

### Implementation
- [IMPLEMENTATION_COMPLETE.md](IMPLEMENTATION_COMPLETE.md) - Implementation summary

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Central API (Go)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │ Release  │  │ Modules  │  │ Currency │             │
│  │  Notes   │  │ Metadata │  │ Exchange │             │
│  └──────────┘  └──────────┘  └──────────┘             │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│           MCP Documentation Server (Python)              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │  GitHub  │  │ ChromaDB │  │ Bedrock  │             │
│  │   Sync   │  │  Vector  │  │  Titan   │             │
│  └──────────┘  └──────────┘  └──────────┘             │
└─────────────────────────────────────────────────────────┘
```

## 🛠️ Development

### Prerequisites
- Go 1.19+
- Make
- Wire (for dependency injection)

### Build
```bash
make build
```

### Run Tests
```bash
go test ./...
```

### Generate Wire
```bash
make wire
```

## 🐳 Docker

### Build Image
```bash
docker build -t central-api:latest .
```

### Run Container
```bash
docker run -p 8080:8080 \
  -e BLOB_STORAGE_PROVIDER=S3 \
  -e AWS_ACCESS_KEY_ID=xxx \
  central-api:latest
```

## 📝 License

Apache License 2.0 - Copyright (c) 2024 Devtron Inc.

## 🤝 Contributing

Contributions are welcome! Please read the contributing guidelines before submitting PRs.

## 📞 Support

- Documentation: See files listed above
- Issues: GitHub Issues
- Website: https://devtron.ai

---

**Maintained by**: Devtron Labs
**Repository**: https://github.com/devtron-labs/central-api