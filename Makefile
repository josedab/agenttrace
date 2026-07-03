.PHONY: all build test lint dev dev-full dev-hot dev-hot-full dev-api dev-api-full dev-worker dev-web clean help format format-check docs-build docs-serve migrate-pg-up migrate-pg-down migrate-ch-up migrate-ch-down setup setup-full _setup health-check doctor test-e2e-api test-e2e-web generate test-coverage test-coverage-api test-coverage-web docker-up docker-up-full docker-down

# Default target
all: lint test build

## help: Show this help message
help:
	@echo "AgentTrace Monorepo Commands:"
	@echo ""
	@echo "  make setup        - Set up minimal development (PostgreSQL + ClickHouse)"
	@echo "  make setup-full   - Set up all services, including Redis and MinIO"
	@echo "  make dev          - Start minimal API + web development"
	@echo "  make dev-full     - Start API + web with Redis and MinIO"
	@echo "  make dev-hot      - Start minimal development with Go hot-reload"
	@echo "  make dev-hot-full - Start full development with Go hot-reload"
	@echo "  make dev-api      - Start only the minimal API stack"
	@echo "  make dev-api-full - Start the API with all services"
	@echo "  make dev-worker   - Start the background worker and all services"
	@echo "  make dev-web      - Start only the web frontend"
	@echo "  make build        - Build all components"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run all linters"
	@echo "  make format       - Format all code"
	@echo "  make format-check - Check formatting without modifying files"
	@echo "  make generate     - Run GraphQL code generation"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start PostgreSQL and ClickHouse"
	@echo "  make docker-up-full - Start PostgreSQL, ClickHouse, Redis, and MinIO"
	@echo "  make docker-down  - Stop local services"
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
	@echo "  make test-e2e-api - Run API end-to-end tests"
	@echo "  make test-e2e-web - Run web end-to-end tests"
	@echo "  make test-coverage     - Generate HTML coverage reports"
	@echo "  make test-coverage-open - Generate and open coverage reports in browser"

# ============================================
# Development
# ============================================

dev: DEV_REDIS=false
dev: DEV_MINIO=false
dev: DEV_RATE_LIMIT=false
dev: docker-up

dev-full: DEV_REDIS=true
dev-full: DEV_MINIO=true
dev-full: DEV_RATE_LIMIT=true
dev-full: docker-up-full

## dev: Start API and web in minimal or full mode
dev dev-full:
	@set -e; \
	echo "Starting API server..."; \
	cd api; \
	set -a; if [ -f .env ] && [ "$${AGENTTRACE_DEVCONTAINER:-false}" != "true" ]; then . ./.env; fi; set +a; \
	export REDIS_ENABLED=$(DEV_REDIS) MINIO_ENABLED=$(DEV_MINIO) RATE_LIMIT_ENABLED=$(DEV_RATE_LIMIT); \
	go run ./cmd/server & \
	API_PID=$$!; \
	WORKER_PID=""; \
	if [ "$(DEV_REDIS)" = "true" ]; then \
		echo "Starting background worker..."; \
		go run ./cmd/worker & \
		WORKER_PID=$$!; \
	fi; \
	cd ../web; \
	cleanup() { \
		kill $$API_PID 2>/dev/null || true; \
		if [ -n "$$WORKER_PID" ]; then kill $$WORKER_PID 2>/dev/null || true; fi; \
	}; \
	trap cleanup EXIT INT TERM; \
	sleep 1; \
	if ! kill -0 $$API_PID 2>/dev/null; then \
		echo "ERROR: API server failed to start. Check the output above."; \
		wait $$API_PID; \
		exit 1; \
	fi; \
	if [ -n "$$WORKER_PID" ] && ! kill -0 $$WORKER_PID 2>/dev/null; then \
		echo "ERROR: Background worker failed to start. Check the output above."; \
		wait $$WORKER_PID; \
		exit 1; \
	fi; \
	echo "Starting web dev server..."; \
	npm run dev

dev-hot: DEV_REDIS=false
dev-hot: DEV_MINIO=false
dev-hot: DEV_RATE_LIMIT=false
dev-hot: docker-up

dev-hot-full: DEV_REDIS=true
dev-hot-full: DEV_MINIO=true
dev-hot-full: DEV_RATE_LIMIT=true
dev-hot-full: docker-up-full

## dev-hot: Start development with Go hot-reload
dev-hot dev-hot-full:
	@if ! command -v air > /dev/null 2>&1; then \
		echo "Error: 'air' is not installed. Install with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi
	@set -e; \
	echo "Starting API server with hot-reload..."; \
	cd api; \
	set -a; if [ -f .env ] && [ "$${AGENTTRACE_DEVCONTAINER:-false}" != "true" ]; then . ./.env; fi; set +a; \
	export REDIS_ENABLED=$(DEV_REDIS) MINIO_ENABLED=$(DEV_MINIO) RATE_LIMIT_ENABLED=$(DEV_RATE_LIMIT); \
	air & \
	AIR_PID=$$!; \
	WORKER_PID=""; \
	if [ "$(DEV_REDIS)" = "true" ]; then \
		echo "Starting background worker..."; \
		go run ./cmd/worker & \
		WORKER_PID=$$!; \
	fi; \
	cd ../web; \
	cleanup() { \
		kill $$AIR_PID 2>/dev/null || true; \
		if [ -n "$$WORKER_PID" ]; then kill $$WORKER_PID 2>/dev/null || true; fi; \
	}; \
	trap cleanup EXIT INT TERM; \
	echo "Starting web dev server..."; \
	npm run dev

## dev-api: Start only the API server with core services
dev-api: docker-up
	@echo "Starting API server..."
	cd api && $(MAKE) run-core

## dev-api-full: Start only the API server with all services
dev-api-full: docker-up-full
	@echo "Starting API server with all services..."
	cd api && $(MAKE) run-full

## dev-worker: Start the background worker with all services
dev-worker: docker-up-full
	@echo "Starting background worker..."
	cd api && $(MAKE) run-worker

## dev-web: Start only the web frontend
dev-web:
	@echo "Starting web dev server..."
	cd web && npm run dev

## docker-up: Start core development databases
docker-up:
	docker compose -f deploy/docker-compose.dev.yml up -d --wait postgres clickhouse

## docker-up-full: Start all development services
docker-up-full:
	docker compose -f deploy/docker-compose.dev.yml --profile full up -d --wait postgres clickhouse redis minio
	docker compose -f deploy/docker-compose.dev.yml --profile full run --rm createbuckets

## docker-down: Stop development databases
docker-down:
	docker compose -f deploy/docker-compose.dev.yml --profile full down

# Prerequisite check (used by setup)
_check-prerequisites:
	@echo "Checking prerequisites..."
	@MISSING=""; \
	if ! command -v go > /dev/null 2>&1; then \
		MISSING="$$MISSING\n  ✗ Go not found (install from https://go.dev/dl/)"; \
	fi; \
	if ! command -v node > /dev/null 2>&1; then \
		MISSING="$$MISSING\n  ✗ Node.js not found (install from https://nodejs.org/)"; \
	fi; \
	if ! command -v docker > /dev/null 2>&1; then \
		MISSING="$$MISSING\n  ✗ Docker not found (install from https://docs.docker.com/get-docker/)"; \
	fi; \
	if ! docker compose version > /dev/null 2>&1; then \
		MISSING="$$MISSING\n  ✗ Docker Compose v2 not found (install the Docker Compose plugin)"; \
	fi; \
	if [ -n "$$MISSING" ]; then \
		echo "Missing required tools:"; \
		printf "$$MISSING\n"; \
		echo ""; \
		echo "Install the missing prerequisites and try again."; \
		echo "Run 'make doctor' for a full prerequisites check."; \
		exit 1; \
	fi; \
	echo "  ✓ All prerequisites found"

setup: _check-prerequisites
	$(MAKE) docker-up
	$(MAKE) _setup

setup-full: _check-prerequisites
	$(MAKE) docker-up-full
	$(MAKE) _setup

## _setup: Install dependencies and migrate core databases
_setup:
	@echo "Copying environment files (if missing)..."
	@if [ ! -f api/.env ]; then cp api/.env.example api/.env && echo "  Created api/.env"; else echo "  api/.env already exists, skipping"; fi
	@if [ ! -f web/.env.local ]; then cp web/.env.example web/.env.local && echo "  Created web/.env.local"; else echo "  web/.env.local already exists, skipping"; fi
	@echo "Installing web dependencies..."
	cd web && npm install
	@echo "Downloading Go modules..."
	cd api && go mod download
	@echo "Running PostgreSQL migrations..."
	$(MAKE) migrate-pg-up
	@echo "Running ClickHouse migrations..."
	$(MAKE) migrate-ch-up
	@echo ""
	@echo "Setup complete. Run 'make dev' for the minimal stack or 'make dev-full' for all services."

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
	cd api && go build -o bin/migrate ./cmd/migrate

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

## test-coverage: Generate and open HTML coverage reports
test-coverage: test-coverage-api test-coverage-web

test-coverage-api:
	@echo "Generating Go backend coverage report..."
	cd api && go test -coverprofile=coverage.out ./...
	cd api && go tool cover -html=coverage.out -o coverage.html
	@echo "  API coverage report: api/coverage.html"

test-coverage-web:
	@echo "Generating web frontend coverage report..."
	cd web && npx vitest run --coverage
	@echo "  Web coverage report: web/coverage/"

## test-coverage-open: Open coverage reports in browser
test-coverage-open: test-coverage
	@echo "Opening coverage reports..."
	@if [ -f api/coverage.html ]; then open api/coverage.html 2>/dev/null || xdg-open api/coverage.html 2>/dev/null || echo "Open api/coverage.html manually"; fi
	@if [ -d web/coverage ]; then open web/coverage/index.html 2>/dev/null || xdg-open web/coverage/index.html 2>/dev/null || echo "Open web/coverage/index.html manually"; fi

## test-e2e-api: Run API end-to-end tests
test-e2e-api:
	@echo "Running API E2E tests..."
	cd api && $(MAKE) test-e2e

## test-e2e-web: Run web end-to-end tests
test-e2e-web:
	@echo "Running web E2E tests..."
	cd web && npx playwright test

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
	@echo "Formatting Go SDK..."
	cd sdk/go && gofmt -w .
	@echo "Formatting CLI..."
	cd sdk/cli && gofmt -w .

## format-check: Verify formatting without modifying files
format-check:
	@echo "Checking Go backend formatting..."
	@UNFORMATTED=$$(cd api && gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "  ✗ Unformatted Go files in api/:"; echo "$$UNFORMATTED" | sed 's/^/    /'; \
		FORMAT_FAIL=1; \
	else \
		echo "  ✓ api/ Go formatting OK"; \
	fi; \
	echo "Checking web frontend formatting..."; \
	if ! (cd web && npx prettier --check . 2>/dev/null); then \
		echo "  ✗ Prettier check failed in web/"; \
		FORMAT_FAIL=1; \
	else \
		echo "  ✓ web/ formatting OK"; \
	fi; \
	echo "Checking Python SDK formatting..."; \
	if ! (cd sdk/python && ruff format --check . 2>/dev/null); then \
		echo "  ✗ Ruff format check failed in sdk/python/"; \
		FORMAT_FAIL=1; \
	else \
		echo "  ✓ sdk/python/ formatting OK"; \
	fi; \
	UNFORMATTED_SDK=$$(cd sdk/go && gofmt -l .); \
	if [ -n "$$UNFORMATTED_SDK" ]; then \
		echo "  ✗ Unformatted Go files in sdk/go/:"; echo "$$UNFORMATTED_SDK" | sed 's/^/    /'; \
		FORMAT_FAIL=1; \
	else \
		echo "  ✓ sdk/go/ Go formatting OK"; \
	fi; \
	UNFORMATTED_CLI=$$(cd sdk/cli && gofmt -l .); \
	if [ -n "$$UNFORMATTED_CLI" ]; then \
		echo "  ✗ Unformatted Go files in sdk/cli/:"; echo "$$UNFORMATTED_CLI" | sed 's/^/    /'; \
		FORMAT_FAIL=1; \
	else \
		echo "  ✓ sdk/cli/ Go formatting OK"; \
	fi; \
	if [ -n "$$FORMAT_FAIL" ]; then \
		echo ""; \
		echo "Formatting issues found. Run 'make format' to fix."; \
		exit 1; \
	fi; \
	echo ""; \
	echo "All formatting checks passed."

# ============================================
# Code Generation
# ============================================

## generate: Run GraphQL code generation
generate:
	@echo "Running GraphQL code generation..."
	cd api && go run github.com/99designs/gqlgen generate

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
	@if docker compose version > /dev/null 2>&1; then \
		echo "  ✓ Docker Compose: $$(docker compose version --short)"; \
	else \
		echo "  ✗ Docker Compose: not found (install the Docker Compose v2 plugin)"; \
	fi
	@if command -v golangci-lint > /dev/null 2>&1; then \
		echo "  ✓ golangci-lint: $$(golangci-lint --version 2>&1 | awk '{print $$4}')"; \
	else \
		echo "  ✗ golangci-lint: not found (install from https://golangci-lint.run/welcome/install/)"; \
	fi
	@if command -v pre-commit > /dev/null 2>&1; then \
		echo "  ✓ pre-commit: $$(pre-commit --version)"; \
	else \
		echo "  ✗ pre-commit: not found (pip install pre-commit)"; \
	fi
	@if command -v air > /dev/null 2>&1; then \
		echo "  ✓ air: installed (hot-reload)"; \
	else \
		echo "  ✗ air: not found (go install github.com/air-verse/air@latest) [optional, for hot-reload]"; \
	fi
	@echo ""
