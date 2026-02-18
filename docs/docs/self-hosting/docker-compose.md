---
sidebar_position: 2
title: "Docker Compose Deployment"
description: "Production Docker Compose deployment guide with proper secrets, TLS, and backup configuration."
---

# Docker Compose Deployment

This guide walks through deploying AgentTrace with Docker Compose, suitable for development environments and small-to-medium teams (up to ~50 users).

## Prerequisites

- Docker Engine 24+ and Docker Compose v2
- Minimum 4 CPU cores and 8 GB RAM
- 50 GB available disk space
- A domain name (for TLS in production)

## Quick Start

```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/deploy
cp .env.example .env
# Edit .env with your configuration
docker compose up -d
```

## Configuration

### Environment File

Copy and edit the environment file at `deploy/.env.example`:

```bash
cp .env.example .env
```

At minimum, set these required values:

```bash
# Database credentials
POSTGRES_PASSWORD=your-secure-postgres-password
CLICKHOUSE_PASSWORD=your-secure-clickhouse-password
REDIS_PASSWORD=your-secure-redis-password
MINIO_ROOT_PASSWORD=your-secure-minio-password

# Security keys — generate with: openssl rand -base64 32
JWT_SECRET=your-jwt-secret-key
ENCRYPTION_KEY=your-encryption-key

# NextAuth
NEXTAUTH_URL=https://your-domain.com
NEXTAUTH_SECRET=your-nextauth-secret
```

See [Configuration Reference](./configuration.md) for all available environment variables.

### Generating Secrets

Never use default or weak secrets in production. Generate strong values:

```bash
# Generate a random 32-byte secret
openssl rand -base64 32

# Generate all secrets at once
for key in JWT_SECRET ENCRYPTION_KEY NEXTAUTH_SECRET; do
  echo "${key}=$(openssl rand -base64 32)"
done
```

## Production Setup

### TLS with Traefik

The default `docker-compose.yml` includes Traefik for reverse proxying. For TLS, create a `docker-compose.override.yml`:

```yaml
services:
  traefik:
    command:
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@your-domain.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    volumes:
      - letsencrypt:/letsencrypt

volumes:
  letsencrypt:
```

### Data Persistence

Ensure volumes are configured for data persistence:

```yaml
volumes:
  postgres_data:
    driver: local
  clickhouse_data:
    driver: local
  redis_data:
    driver: local
  minio_data:
    driver: local
```

### Resource Limits

Set resource limits to prevent any single service from consuming all resources:

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
  clickhouse:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

## Running Migrations

After the first start, run database migrations:

```bash
docker compose exec api /app/server migrate up
```

## Verifying the Deployment

```bash
# Check all services are running
docker compose ps

# Verify API health
curl http://localhost:8080/health

# Check logs for errors
docker compose logs --tail=50 api
docker compose logs --tail=50 web
```

The dashboard is available at `http://localhost:3000` (or your configured `NEXTAUTH_URL`).

## Backup

Set up automated daily backups:

```bash
# Manual PostgreSQL backup
docker compose exec postgres pg_dump -U agenttrace > backup-$(date +%F).sql

# Manual ClickHouse backup
docker compose exec clickhouse clickhouse-client \
  --query "BACKUP TABLE traces, observations TO Disk('backups', 'backup-$(date +%F).zip')"
```

See [Backup & Restore](./backup.md) for comprehensive backup procedures.

## Upgrading

```bash
# 1. Back up your data
./scripts/backup.sh

# 2. Pull latest images
docker compose pull

# 3. Stop and restart with new images
docker compose down
docker compose run --rm api /app/server migrate up
docker compose up -d

# 4. Verify
docker compose logs -f api
```

## Troubleshooting

### Services fail to start

```bash
docker compose logs <service-name>
docker stats  # Check resource usage
```

### Database connection errors

```bash
docker compose exec postgres pg_isready
docker compose exec clickhouse clickhouse-client --query "SELECT 1"
docker compose exec redis redis-cli ping
```

### Port conflicts

If ports 3000, 8080, or database ports are in use, override them in `.env`:

```bash
WEB_PORT=3001
API_PORT=8081
```

## Related

- [Configuration Reference](./configuration.md) — all environment variables
- [Kubernetes Deployment](./kubernetes.md) — for larger-scale deployments
- [Scaling Guide](./scaling.md) — horizontal scaling strategies
- [Backup & Restore](./backup.md) — backup procedures
