.PHONY: all build run run-core run-full run-worker test lint clean migrate-up migrate-down generate docker-up docker-up-full docker-down dev dev-full dev-setup dev-setup-full

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=agenttrace-server
WORKER_BINARY=agenttrace-worker
MIGRATE_BINARY=agenttrace-migrate
MIGRATE_VERSION?=v4.18.3
MIGRATE_CLI=$(GOCMD) run -tags 'postgres clickhouse' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
DEV_COMPOSE=docker compose -f ../deploy/docker-compose.dev.yml

# Directories
CMD_SERVER=./cmd/server
CMD_WORKER=./cmd/worker
MIGRATIONS_PG=./migrations/postgres
MIGRATIONS_CH=./migrations/clickhouse
LOCAL_CLICKHOUSE_MIGRATE_URL=clickhouse://localhost:9000?username=agenttrace&password=agenttrace&database=agenttrace&x-multi-statement=true
CLICKHOUSE_MIGRATE_URL?=$(LOCAL_CLICKHOUSE_MIGRATE_URL)

# Database URLs (defaults for local dev)
POSTGRES_URL?=postgres://agenttrace:agenttrace@localhost:5432/agenttrace?sslmode=disable
CLICKHOUSE_URL?=clickhouse://agenttrace:agenttrace@localhost:9000/agenttrace

all: lint test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(CMD_SERVER)
	$(GOBUILD) -o bin/$(WORKER_BINARY) $(CMD_WORKER)
	$(GOBUILD) -o bin/$(MIGRATE_BINARY) ./cmd/migrate

run:
	@set -a; if [ -f .env ] && [ "$${AGENTTRACE_DEVCONTAINER:-false}" != "true" ]; then . ./.env; fi; set +a; exec $(GOCMD) run $(CMD_SERVER)

run-core:
	@set -a; if [ -f .env ] && [ "$${AGENTTRACE_DEVCONTAINER:-false}" != "true" ]; then . ./.env; fi; set +a; \
		export REDIS_ENABLED=false MINIO_ENABLED=false RATE_LIMIT_ENABLED=false; \
		exec $(GOCMD) run $(CMD_SERVER)

run-full:
	@set -a; if [ -f .env ] && [ "$${AGENTTRACE_DEVCONTAINER:-false}" != "true" ]; then . ./.env; fi; set +a; \
		export REDIS_ENABLED=true MINIO_ENABLED=true RATE_LIMIT_ENABLED=true; \
		exec $(GOCMD) run $(CMD_SERVER)

run-worker:
	@set -a; if [ -f .env ] && [ "$${AGENTTRACE_DEVCONTAINER:-false}" != "true" ]; then . ./.env; fi; set +a; \
		export REDIS_ENABLED=true MINIO_ENABLED=true; \
		exec $(GOCMD) run $(CMD_WORKER)

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-integration:
	$(GOTEST) -v -race -tags=integration ./...

test-e2e:
	@echo "Running E2E tests..."
	@echo "Ensure AGENTTRACE_API_KEY is set and API is running"
	$(GOTEST) -v -tags=e2e ./tests/e2e/...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
	rm -f coverage.out

# Dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Database migrations
migrate-pg-up:
	DATABASE_URL="$(POSTGRES_URL)" $(GOCMD) run ./cmd/migrate -database postgres -path ./migrations up

migrate-pg-down:
	DATABASE_URL="$(POSTGRES_URL)" $(GOCMD) run ./cmd/migrate -database postgres -path ./migrations down

migrate-pg-create:
	@read -p "Enter migration name: " name; \
	$(MIGRATE_CLI) create -ext sql -dir $(MIGRATIONS_PG) -seq $$name

migrate-ch-up:
	CLICKHOUSE_DSN="$(CLICKHOUSE_MIGRATE_URL)" $(GOCMD) run ./cmd/migrate -database clickhouse -path ./migrations up

migrate-ch-down:
	CLICKHOUSE_DSN="$(CLICKHOUSE_MIGRATE_URL)" $(GOCMD) run ./cmd/migrate -database clickhouse -path ./migrations down

# GraphQL generation
generate:
	go run github.com/99designs/gqlgen generate

# Docker
docker-up:
	$(DEV_COMPOSE) up -d --wait postgres clickhouse

docker-up-full:
	$(DEV_COMPOSE) --profile full up -d --wait postgres clickhouse redis minio
	$(DEV_COMPOSE) --profile full run --rm createbuckets

docker-down:
	$(DEV_COMPOSE) --profile full down

docker-logs:
	$(DEV_COMPOSE) logs -f

# Development helpers
dev: docker-up migrate-pg-up migrate-ch-up run-core

dev-setup: deps docker-up
	$(MAKE) migrate-pg-up
	$(MAKE) migrate-ch-up
	@echo "Development environment ready!"

dev-full: docker-up-full migrate-pg-up migrate-ch-up run-full

dev-setup-full: deps docker-up-full
	$(MAKE) migrate-pg-up
	$(MAKE) migrate-ch-up
	@echo "Full development environment ready!"

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-amd64 $(CMD_SERVER)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-arm64 $(CMD_SERVER)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-amd64 $(CMD_SERVER)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-arm64 $(CMD_SERVER)
