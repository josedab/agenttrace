---
sidebar_position: 3
title: Local Development
description: Run AgentTrace locally with no deployed services for tests or only the services required by your feature.
---

# Local Development

AgentTrace supports three local workflows so contributors do not need to run the full production architecture for every change.

## Prerequisites

- Go 1.25.12+
- Node.js 20+
- Docker with Docker Compose v2
- Make

Python is only required for Python SDK work.

## Workflow Matrix

| Workflow | Deployed services | Use it for |
|----------|-------------------|------------|
| Tests only | None | Backend, frontend, and SDK unit or component tests |
| Core development | PostgreSQL, ClickHouse | API, dashboard, tracing, prompts, datasets, and most product work |
| Full development | PostgreSQL, ClickHouse, Redis, MinIO | Background workers, async exports, distributed rate limiting, and object storage |

PostgreSQL stores relational metadata and ClickHouse stores traces and analytics, so both are required for a functional API. Redis and MinIO are optional for normal local application work.

## Tests Without Services

Default tests do not require deployed databases, caches, queues, or object storage:

```bash
make test
```

Run a smaller suite while iterating:

```bash
make test-api
make test-web
make test-sdk-py
make test-sdk-ts
make test-sdk-go
```

Database integration and end-to-end tests are opt-in. Repository integration tests skip unless their documented test host variables are set.

## Minimal Application Stack

Set up dependencies, start PostgreSQL and ClickHouse, and run migrations:

```bash
make doctor
make setup
```

Start the API and dashboard:

```bash
make dev
```

The API runs with Redis, distributed rate limiting, MinIO, and queue-backed exports disabled. Export endpoints return `503 Service Unavailable` with instructions to use the full stack rather than failing at startup.

Useful component commands:

```bash
make dev-api
make dev-web
make dev-hot
```

## Full Application Stack

Use the full profile when working on workers or export infrastructure:

```bash
make setup-full
make dev-full
```

Additional commands:

```bash
make dev-api-full
make dev-hot-full
make dev-worker
```

The full profile starts Redis and MinIO in addition to PostgreSQL and ClickHouse.

## Service Management

```bash
# Core services only
make docker-up

# Core plus Redis and MinIO
make docker-up-full

# Stop either stack
make docker-down
```

Docker Compose waits for service health checks before returning.

## Database Migrations

No global migration CLI or host-installed PostgreSQL or ClickHouse client is required. The Make targets run a pinned `golang-migrate` version through Go:

```bash
make migrate-pg-up
make migrate-ch-up
```

Rollback targets are also available:

```bash
make migrate-pg-down
make migrate-ch-down
```

## Dev Container

Opening the repository in the devcontainer starts only PostgreSQL and ClickHouse by default. To add the optional services, run this from the host repository:

```bash
docker compose --project-name agenttrace-devcontainer \
  -f .devcontainer/docker-compose.yml \
  --profile full up -d redis minio
```

Set `REDIS_ENABLED=true`, `MINIO_ENABLED=true`, and `RATE_LIMIT_ENABLED=true` when running the API with those services.
