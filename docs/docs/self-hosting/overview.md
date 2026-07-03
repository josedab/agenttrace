---
sidebar_position: 1
---

# Self-Hosting Overview

AgentTrace can run entirely in your infrastructure. The supported repository deployment paths are Docker Compose and Kubernetes/Kustomize.

## Architecture

```text
Browser
  |
Load balancer / ingress
  |------------------|
Web (Next.js)     API (Go/Fiber) ---- Worker (Asynq)
                       |       |        |
                  PostgreSQL ClickHouse Redis
                                      |
                              External S3 storage
```

| Component | Purpose |
|-----------|---------|
| Web | Dashboard and browser UI |
| API | Authentication, REST/GraphQL APIs, ingestion, health, and metrics |
| Worker | Queued evaluations, exports, notifications, and cleanup |
| PostgreSQL | Relational metadata |
| ClickHouse | Traces, observations, scores, and analytics |
| Redis | Queue and distributed rate limiting |
| S3-compatible storage (optional) | Export object storage |

## Docker Compose

Use Compose for a single-node installation or small team:

```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/deploy
cp .env.example .env
# Replace all change-me values and configure public URLs.
docker compose up -d --build
./smoke-test.sh
```

Migrations are an automatic startup gate.

[Docker Compose guide](/self-hosting/docker-compose)

## Kubernetes

Use the bundled Kustomize base for orchestrated deployments:

```bash
cp deploy/kubernetes/secrets.yaml.example deploy/kubernetes/secrets.yaml
# Replace secret placeholders and domain/image values.
kubectl apply -f deploy/kubernetes/secrets.yaml
kubectl apply -k deploy/kubernetes/
```

API and worker init containers run serialized migrations before application startup.

[Kubernetes guide](/self-hosting/kubernetes)

## Capacity Baseline

| Environment | CPU | Memory | Storage |
|-------------|-----|--------|---------|
| Evaluation/small team | 4 cores | 8 GB | 50 GB SSD |
| Initial production baseline | 8 cores | 32 GB | 500 GB SSD |

Actual ClickHouse sizing depends on trace volume, payload size, and retention. Load-test with representative data before setting production limits.

## Deployment Requirements

- Use immutable API and web image tags.
- Build the web image with the public `NEXT_PUBLIC_API_URL`.
- Store secrets in a secret manager, not source control.
- Terminate TLS before exposing the application.
- Back up PostgreSQL, ClickHouse, and external object storage independently.
- Monitor `/health`, `/readyz`, `/metrics`, worker queues, disk, and backup freshness.
- Test a clean installation and an upgrade against a copy of production data.

## Configuration

See the [Configuration Reference](/self-hosting/configuration) for supported component variables and DSN aliases.
