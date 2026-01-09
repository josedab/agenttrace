#!/bin/bash
set -e

echo "=== AgentTrace Dev Container Setup ==="

# Install Go tools
echo "Installing Go tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/99designs/gqlgen@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Setup API
echo "Setting up API..."
cd /workspace/api
go mod download

# Setup Web
echo "Setting up Web frontend..."
cd /workspace/web
npm install

# Setup Python SDK
echo "Setting up Python SDK..."
cd /workspace/sdk/python
python -m pip install --upgrade pip
pip install -e ".[dev]"

# Setup TypeScript SDK
echo "Setting up TypeScript SDK..."
cd /workspace/sdk/typescript
npm install

# Setup Go SDK
echo "Setting up Go SDK..."
cd /workspace/sdk/go
go mod download

# Setup CLI
echo "Setting up CLI..."
cd /workspace/sdk/cli
go mod download

# Create MinIO bucket
echo "Creating MinIO bucket..."
mc alias set local http://minio:9000 minioadmin minioadmin || true
mc mb local/agenttrace --ignore-existing || true

echo "=== Setup Complete ==="
echo ""
echo "To get started:"
echo "  1. Run database migrations:  cd /workspace/api && make migrate-pg-up && make migrate-ch-up"
echo "  2. Start the API server:     cd /workspace/api && go run cmd/server/main.go"
echo "  3. Start the web frontend:   cd /workspace/web && npm run dev"
echo ""
echo "Access the application at http://localhost:3000"
