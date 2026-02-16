---
sidebar_position: 2
title: Installation
description: How to install and deploy AgentTrace using Docker Compose or Kubernetes.
---

# Installation

This guide walks you through deploying AgentTrace on your own infrastructure. Docker Compose is the fastest way to get started; Kubernetes is recommended for production.

## Prerequisites

| Requirement | Minimum Version |
|-------------|-----------------|
| [Docker](https://docs.docker.com/get-docker/) | 24.0+ |
| [Docker Compose](https://docs.docker.com/compose/install/) | 2.20+ |
| Git | 2.30+ |

**Hardware** (minimum for development):

- 4 CPU cores
- 8 GB RAM
- 50 GB SSD

For production sizing, see the [Self-Hosting Overview](/self-hosting/overview).

## Docker Compose (Recommended)

Docker Compose deploys all components — API server, web dashboard, PostgreSQL, ClickHouse, Redis, and MinIO — in a single command.

### 1. Clone the Repository

```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/deploy
```

### 2. Configure Environment Variables

```bash
cp .env.example .env
```

Open `.env` and set the **required** values:

```bash
# Database passwords — change these!
POSTGRES_PASSWORD=change-me-to-a-secure-password
CLICKHOUSE_PASSWORD=change-me-to-a-secure-password
REDIS_PASSWORD=change-me-to-a-secure-password
MINIO_ROOT_PASSWORD=change-me-to-a-secure-password

# Security keys — generate with: openssl rand -base64 32
JWT_SECRET=<generated-secret>
ENCRYPTION_KEY=<generated-secret>

# NextAuth
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=<generated-secret>
```

See [`deploy/.env.example`](https://github.com/agenttrace/agenttrace/blob/main/deploy/.env.example) for the full list of available variables, including optional OAuth and SMTP settings.

### 3. Start AgentTrace

```bash
docker compose up -d
```

This pulls the images and starts all services. On first run it may take a few minutes.

### 4. Run Database Migrations

```bash
docker compose exec api /app/server migrate up
```

### 5. Verify the Deployment

```bash
# Check that all containers are healthy
docker compose ps

# Test the API health endpoint
curl http://localhost:8080/health
```

You should see a response like:

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "postgres": "connected",
  "clickhouse": "connected",
  "redis": "connected"
}
```

### 6. Open the Dashboard

Navigate to [http://localhost:3000](http://localhost:3000) in your browser to access the AgentTrace dashboard. Create your first account to get started.

### Stopping and Restarting

```bash
# Stop all services (data is preserved in Docker volumes)
docker compose down

# Restart
docker compose up -d
```

### Upgrading

```bash
# Pull the latest images
docker compose pull

# Restart with the new version
docker compose down && docker compose up -d

# Run any new migrations
docker compose exec api /app/server migrate up
```

## Kubernetes

For production deployments with high availability, use the Kubernetes manifests in `deploy/kubernetes/`.

### Prerequisites

| Requirement | Minimum Version |
|-------------|-----------------|
| Kubernetes cluster | 1.27+ |
| kubectl | Matching cluster version |
| Helm (optional) | 3.12+ |

### Using Kustomize

```bash
# Review and edit the secrets template
cd deploy/kubernetes
cp secrets.yaml.example secrets.yaml
# Edit secrets.yaml with your base64-encoded credentials

# Apply all resources
kubectl apply -k .
```

This creates a dedicated `agenttrace` namespace and deploys all components including the API server, workers, PostgreSQL, ClickHouse, Redis, ingress, network policies, and horizontal pod autoscalers.

### Using Helm

```bash
helm repo add agenttrace https://charts.agenttrace.io
helm install agenttrace agenttrace/agenttrace \
  --namespace agenttrace \
  --create-namespace \
  -f values.yaml
```

### Verify the Cluster

```bash
kubectl get pods -n agenttrace
kubectl logs -n agenttrace deployment/api --tail=50
```

## What's Next?

Now that AgentTrace is running:

- [Quickstart](/getting-started/quickstart) — Get an API key and instrument your first agent
- [Create Your First Trace](/getting-started/first-trace) — End-to-end tutorial with Python, TypeScript, and Go
- [Core Concepts](/getting-started/concepts) — Understand traces, observations, and sessions
