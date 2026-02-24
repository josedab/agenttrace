.PHONY: all build test lint dev dev-api dev-web clean help format docs-build docs-serve migrate-pg-up migrate-pg-down migrate-ch-up migrate-ch-down setup health-check doctor

# Default target
all: lint test build

## help: Show this help message
help:
	@echo "AgentTrace Monorepo Commands:"
	@echo ""
	@echo "  make setup        - One-command dev environment setup"
	@echo "  make dev          - Start dev environment (API + web, Ctrl+C stops both)"
	@echo "  make dev-api      - Start only the API server (with databases)"
	@echo "  make dev-web      - Start only the web frontend"
	@echo "  make build        - Build all components"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run all linters"
	@echo "  make format       - Format all code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start databases via Docker Compose"
	@echo "  make docker-down  - Stop databases"
	@echo "  make health-check - Verify API and web are running"
	@echo "  make doctor       - Check all prerequisites are installed"
	@echo "  make docs-build   - Build documentation site"
	@echo "  make docs-serve   - Serve documentation locally"
	@echo ""
	@echo "Database migrations:"
	@echo "  make migrate-pg-up   - Run PostgreSQL migrations"
	@echo "  make migrate-pg-down - Roll back one PostgreSQL migration"
	@echo "  make migrate-ch-up   - Run ClickHouse migrations"
	@echo "  make migrate-ch-down - Roll back one ClickHouse migration"
	@echo ""
	@echo "Component targets:"
	@echo "  make test-api     - Run Go backend tests"
	@echo "  make test-web     - Run frontend tests"
	@echo "  make test-sdk-py  - Run Python SDK tests"
	@echo "  make test-sdk-ts  - Run TypeScript SDK tests"
	@echo "  make test-sdk-go  - Run Go SDK tests"

# ============================================
# Development
# ============================================

## dev: Start development environment
dev: docker-up
	@echo "Starting API server..."
	@cd api && go run cmd/server/main.go &
	@API_PID=$$!; \
	trap "kill $$API_PID 2>/dev/null; exit" INT TERM; \
	echo "API server PID: $$API_PID"; \
	echo "Starting web dev server..."; \
	cd web && npm run dev; \
	kill $$API_PID 2>/dev/null

## dev-api: Start only the API server
dev-api: docker-up
	@echo "Starting API server..."
	cd api && go run cmd/server/main.go

## dev-web: Start only the web frontend
dev-web:
	@echo "Starting web dev server..."
	cd web && npm run dev

## docker-up: Start development databases
docker-up:
	docker compose -f deploy/docker-compose.dev.yml up -d

## docker-down: Stop development databases
docker-down:
	docker compose -f deploy/docker-compose.dev.yml down

## setup: One-command development environment setup
setup: docker-up
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@echo "Installing web dependencies..."
	cd web && npm install
	@echo "Downloading Go modules..."
	cd api && go mod download
	@echo "Running PostgreSQL migrations..."
	$(MAKE) migrate-pg-up
	@echo "Running ClickHouse migrations..."
	$(MAKE) migrate-ch-up
	@echo ""
	@echo "Setup complete! Run 'make dev' to start the development environment."

# ============================================
# Database Migrations
# ============================================

## migrate-pg-up: Run PostgreSQL migrations
migrate-pg-up:
	cd api && $(MAKE) migrate-pg-up

## migrate-pg-down: Roll back one PostgreSQL migration
migrate-pg-down:
	cd api && $(MAKE) migrate-pg-down

## migrate-ch-up: Run ClickHouse migrations
migrate-ch-up:
	cd api && $(MAKE) migrate-ch-up

## migrate-ch-down: Roll back one ClickHouse migration
migrate-ch-down:
	cd api && $(MAKE) migrate-ch-down

# ============================================
# Health Check
# ============================================

## health-check: Verify API and web are running
health-check:
	@echo "Checking API (http://localhost:8080/health)..."
	@for i in 1 2 3 4 5; do \
		if curl -sf http://localhost:8080/health > /dev/null 2>&1; then \
			echo "  ✓ API is healthy"; \
			break; \
		fi; \
		if [ $$i -eq 5 ]; then \
			echo "  ✗ API is not responding"; \
		else \
			sleep 2; \
		fi; \
	done
	@echo "Checking Web (http://localhost:3000)..."
	@for i in 1 2 3 4 5; do \
		if curl -sf http://localhost:3000 > /dev/null 2>&1; then \
			echo "  ✓ Web is healthy"; \
			break; \
		fi; \
		if [ $$i -eq 5 ]; then \
			echo "  ✗ Web is not responding"; \
		else \
			sleep 2; \
		fi; \
	done

# ============================================
# Build
# ============================================

## build: Build all components
build: build-api build-web build-sdk-ts

build-api:
	@echo "Building API..."
	cd api && go build -o bin/server ./cmd/server
	cd api && go build -o bin/worker ./cmd/worker

build-web:
	@echo "Building web..."
	cd web && npm run build

build-sdk-ts:
	@echo "Building TypeScript SDK..."
	cd sdk/typescript && npm run build

# ============================================
# Test
# ============================================

## test: Run all tests
test: test-api test-web test-sdk-go test-sdk-ts test-sdk-py

test-api:
	@echo "Testing Go backend..."
	cd api && go test -race ./...

test-web:
	@echo "Testing web frontend..."
	cd web && npm test

test-sdk-py:
	@echo "Testing Python SDK..."
	cd sdk/python && pytest

test-sdk-ts:
	@echo "Testing TypeScript SDK..."
	cd sdk/typescript && npm test

test-sdk-go:
	@echo "Testing Go SDK..."
	cd sdk/go && go test -race ./...

test-cli:
	@echo "Testing CLI..."
	cd sdk/cli && go test -race ./...

# ============================================
# Lint
# ============================================

## lint: Run all linters
lint: lint-api lint-web lint-sdk-py lint-sdk-ts

lint-api:
	@echo "Linting Go backend..."
	cd api && golangci-lint run ./...

lint-web:
	@echo "Linting web frontend..."
	cd web && npm run lint

lint-sdk-py:
	@echo "Linting Python SDK..."
	cd sdk/python && ruff check .

lint-sdk-ts:
	@echo "Linting TypeScript SDK..."
	cd sdk/typescript && npm run lint

# ============================================
# Clean
# ============================================

## clean: Clean all build artifacts
clean:
	rm -rf api/bin/
	rm -rf web/.next/
	rm -rf sdk/typescript/dist/
	rm -f api/coverage.out

# ============================================
# Format
# ============================================

## format: Format all code
format:
	@echo "Formatting Go backend..."
	cd api && gofmt -w .
	@echo "Formatting web frontend..."
	cd web && npx prettier --write .
	@echo "Formatting Python SDK..."
	cd sdk/python && ruff format .

# ============================================
# Documentation
# ============================================

## docs-build: Build the documentation site
docs-build:
	cd docs && npm run build

## docs-serve: Serve the documentation site locally
docs-serve:
	cd docs && npm start

# ============================================
# Doctor
# ============================================

## doctor: Check that all prerequisites are installed
doctor:
	@echo "Checking prerequisites..."
	@echo ""
	@if command -v go > /dev/null 2>&1; then \
		echo "  ✓ Go: $$(go version | awk '{print $$3}')"; \
	else \
		echo "  ✗ Go: not found (install from https://go.dev/dl/)"; \
	fi
	@if command -v node > /dev/null 2>&1; then \
		echo "  ✓ Node.js: $$(node --version)"; \
	else \
		echo "  ✗ Node.js: not found (install from https://nodejs.org/)"; \
	fi
	@if command -v docker > /dev/null 2>&1; then \
		echo "  ✓ Docker: $$(docker --version | awk '{print $$3}' | tr -d ',')"; \
	else \
		echo "  ✗ Docker: not found (install from https://docs.docker.com/get-docker/)"; \
	fi
	@if command -v golangci-lint > /dev/null 2>&1; then \
		echo "  ✓ golangci-lint: $$(golangci-lint --version 2>&1 | awk '{print $$4}')"; \
	else \
		echo "  ✗ golangci-lint: not found (install from https://golangci-lint.run/welcome/install/)"; \
	fi
	@if command -v migrate > /dev/null 2>&1; then \
		echo "  ✓ migrate: installed"; \
	else \
		echo "  ✗ migrate: not found (brew install golang-migrate or go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest)"; \
	fi
	@if command -v pre-commit > /dev/null 2>&1; then \
		echo "  ✓ pre-commit: $$(pre-commit --version)"; \
	else \
		echo "  ✗ pre-commit: not found (pip install pre-commit)"; \
	fi
	@echo ""
