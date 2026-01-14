#!/bin/bash
# Setup script for Devtron Documentation MCP Server

set -e

echo "🚀 Setting up Devtron Documentation MCP Server..."

# Check Python version
echo "📋 Checking Python version..."
python_version=$(python3 --version 2>&1 | awk '{print $2}')
required_version="3.9"

if [ "$(printf '%s\n' "$required_version" "$python_version" | sort -V | head -n1)" != "$required_version" ]; then
    echo "❌ Python 3.9+ required. Found: $python_version"
    exit 1
fi
echo "✅ Python version: $python_version"

# Create virtual environment
echo "📦 Creating virtual environment..."
if [ ! -d "venv" ]; then
    python3 -m venv venv
    echo "✅ Virtual environment created"
else
    echo "✅ Virtual environment already exists"
fi

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Upgrade pip
echo "⬆️  Upgrading pip..."
pip install --upgrade pip

# Install dependencies
echo "📥 Installing dependencies..."
pip install -r requirements.txt

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env file with your AWS credentials"
else
    echo "✅ .env file already exists"
fi

# Check AWS credentials
echo "🔐 Checking AWS credentials..."
if [ -z "$AWS_ACCESS_KEY_ID" ] && [ -z "$AWS_PROFILE" ]; then
    echo "⚠️  AWS credentials not found in environment"
    echo "   Please configure AWS credentials using one of these methods:"
    echo "   1. Edit .env file with AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY"
    echo "   2. Run 'aws configure' to set up AWS CLI profile"
    echo "   3. Set AWS_PROFILE environment variable"
else
    echo "✅ AWS credentials configured"
fi

# Create directories
echo "📁 Creating directories..."
mkdir -p devtron-docs
echo "✅ Directories created"

# Check PostgreSQL
echo ""
echo "🗄️  Checking PostgreSQL..."
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL client found"
    echo ""
    echo "To set up the database, run:"
    echo "  ./setup_database.sh"
else
    echo "⚠️  PostgreSQL client not found"
    echo ""
    echo "Please install PostgreSQL or use Docker:"
    echo "  Docker: docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres ankane/pgvector:latest"
    echo "  Or use: docker-compose up -d postgres"
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Configure AWS credentials (if not done already)"
echo "2. Set up PostgreSQL database: ./setup_database.sh"
echo "3. Enable AWS Bedrock Titan Embeddings in AWS Console"
echo "4. Run the server: python server.py"
echo ""
echo "For more information, see README.md"

