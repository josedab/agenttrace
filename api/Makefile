.PHONY: all build run test lint clean migrate-up migrate-down generate docker-up docker-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=agenttrace-server
WORKER_BINARY=agenttrace-worker

# Directories
CMD_SERVER=./cmd/server
CMD_WORKER=./cmd/worker
MIGRATIONS_PG=./migrations/postgres
MIGRATIONS_CH=./migrations/clickhouse

# Database URLs (defaults for local dev)
POSTGRES_URL?=postgres://agenttrace:agenttrace@localhost:5432/agenttrace?sslmode=disable
CLICKHOUSE_URL?=clickhouse://agenttrace:agenttrace@localhost:9000/agenttrace

all: lint test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(CMD_SERVER)
	$(GOBUILD) -o bin/$(WORKER_BINARY) $(CMD_WORKER)

run:
	$(GOCMD) run $(CMD_SERVER)

run-worker:
	$(GOCMD) run $(CMD_WORKER)

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
	migrate -path $(MIGRATIONS_PG) -database "$(POSTGRES_URL)" up

migrate-pg-down:
	migrate -path $(MIGRATIONS_PG) -database "$(POSTGRES_URL)" down 1

migrate-pg-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PG) -seq $$name

migrate-ch-up:
	@for f in $(MIGRATIONS_CH)/*.up.sql; do \
		echo "Running $$f..."; \
		clickhouse-client --host localhost --user agenttrace --password agenttrace --database agenttrace --multiquery < $$f; \
	done

migrate-ch-down:
	@for f in $$(ls -r $(MIGRATIONS_CH)/*.down.sql | head -1); do \
		echo "Rolling back $$f..."; \
		clickhouse-client --host localhost --user agenttrace --password agenttrace --database agenttrace --multiquery < $$f; \
	done

# GraphQL generation
generate:
	go run github.com/99designs/gqlgen generate

# Docker
docker-up:
	docker compose -f ../deploy/docker-compose.dev.yml up -d

docker-down:
	docker compose -f ../deploy/docker-compose.dev.yml down

docker-logs:
	docker compose -f ../deploy/docker-compose.dev.yml logs -f

# Development helpers
dev: docker-up migrate-pg-up migrate-ch-up run

dev-setup: deps docker-up
	@echo "Waiting for databases to be ready..."
	@sleep 5
	$(MAKE) migrate-pg-up
	$(MAKE) migrate-ch-up
	@echo "Development environment ready!"

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-amd64 $(CMD_SERVER)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-arm64 $(CMD_SERVER)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-amd64 $(CMD_SERVER)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-arm64 $(CMD_SERVER)
