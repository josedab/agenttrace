.PHONY: all build test lint dev dev-api dev-web clean help format docs-build docs-serve migrate-pg-up migrate-pg-down migrate-ch-up migrate-ch-down

# Default target
all: lint test build

## help: Show this help message
help:
	@echo "AgentTrace Monorepo Commands:"
	@echo ""
	@echo "  make dev          - Start dev environment (databases + API + web)"
	@echo "  make dev-api      - Start only the API server (with databases)"
	@echo "  make dev-web      - Start only the web frontend"
	@echo "  make build        - Build all components"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run all linters"
	@echo "  make format       - Format all code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start databases via Docker Compose"
	@echo "  make docker-down  - Stop databases"
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
	@echo "Starting web dev server..."
	@cd web && npm run dev

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
test: test-api test-sdk-go test-sdk-ts test-sdk-py

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
