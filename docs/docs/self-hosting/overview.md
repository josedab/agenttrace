---
sidebar_position: 1
---

# Self-Hosting Overview

AgentTrace can be self-hosted in your own infrastructure for complete data control, compliance requirements, or air-gapped environments.

## Architecture

```
                                   ┌─────────────────┐
                                   │   Load Balancer │
                                   │   (nginx/traefik)│
                                   └────────┬────────┘
                                            │
                        ┌───────────────────┼───────────────────┐
                        │                   │                   │
                        ▼                   ▼                   ▼
               ┌────────────────┐  ┌────────────────┐  ┌────────────────┐
               │  Web Frontend  │  │   Go Backend   │  │    Workers     │
               │  (Next.js)     │  │   (Fiber)      │  │   (Asynq)      │
               │   Port 3000    │  │   Port 8080    │  │                │
               └────────────────┘  └────────────────┘  └────────────────┘
                        │                   │                   │
                        │                   │                   │
        ┌───────────────┼───────────────────┼───────────────────┤
        │               │                   │                   │
        ▼               ▼                   ▼                   ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  PostgreSQL  │ │  ClickHouse  │ │    Redis     │ │    MinIO     │
│  (Metadata)  │ │  (Traces)    │ │   (Queue)    │ │  (Storage)   │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

## Components

| Component | Purpose | Default Port |
|-----------|---------|--------------|
| Web Frontend | Next.js dashboard | 3000 |
| API Server | Go Fiber backend | 8080 |
| Workers | Background job processing | - |
| PostgreSQL | Metadata, users, prompts | 5432 |
| ClickHouse | Traces, observations, metrics | 8123/9000 |
| Redis | Job queue, caching | 6379 |
| MinIO | Object storage (optional) | 9000 |

## Deployment Options

### Docker Compose (Development/Small Teams)

Best for: Development, small teams (< 50 users)

```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/deploy
cp .env.example .env
# Edit .env with your settings
docker compose up -d
```

[Docker Compose Guide →](/self-hosting/docker-compose)

### Kubernetes (Production)

Best for: Production, large teams, high availability

```bash
helm repo add agenttrace https://charts.agenttrace.io
helm install agenttrace agenttrace/agenttrace \
  --namespace agenttrace \
  --create-namespace \
  -f values.yaml
```

[Kubernetes Guide →](/self-hosting/kubernetes)

## System Requirements

### Minimum (Development)

| Resource | Requirement |
|----------|-------------|
| CPU | 4 cores |
| Memory | 8 GB |
| Storage | 50 GB SSD |
| OS | Linux, macOS, or Windows with WSL2 |

### Recommended (Production)

| Resource | Requirement |
|----------|-------------|
| CPU | 8+ cores |
| Memory | 32+ GB |
| Storage | 500+ GB SSD |
| Network | 1 Gbps |

### ClickHouse Sizing

ClickHouse storage depends on trace volume:

| Daily Traces | Storage/Month | Memory |
|--------------|---------------|--------|
| 10,000 | ~5 GB | 8 GB |
| 100,000 | ~50 GB | 16 GB |
| 1,000,000 | ~500 GB | 32 GB |

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace
```

### 2. Configure Environment

```bash
cd deploy
cp .env.example .env
```

Edit `.env` with your settings:

```bash
# Required
AGENTTRACE_SECRET_KEY=your-secret-key-min-32-chars
POSTGRES_PASSWORD=your-postgres-password
CLICKHOUSE_PASSWORD=your-clickhouse-password

# Optional
AGENTTRACE_BASE_URL=https://agenttrace.your-company.com
SMTP_HOST=smtp.your-company.com
SMTP_FROM=agenttrace@your-company.com
```

### 3. Start Services

```bash
docker compose up -d
```

### 4. Run Migrations

```bash
docker compose exec api /app/server migrate up
```

### 5. Access AgentTrace

Open `http://localhost:3000` in your browser.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AGENTTRACE_SECRET_KEY` | Encryption key (32+ chars) | Required |
| `AGENTTRACE_BASE_URL` | Public URL | `http://localhost:3000` |
| `DATABASE_URL` | PostgreSQL connection | `postgres://...` |
| `CLICKHOUSE_URL` | ClickHouse connection | `clickhouse://...` |
| `REDIS_URL` | Redis connection | `redis://localhost:6379` |
| `MINIO_ENDPOINT` | MinIO/S3 endpoint | `localhost:9000` |
| `SMTP_HOST` | SMTP server for emails | - |
| `LOG_LEVEL` | Log verbosity | `info` |

### TLS/SSL

For production, configure TLS:

```yaml
# docker-compose.override.yml
services:
  traefik:
    command:
      - "--certificatesresolvers.letsencrypt.acme.email=admin@your-company.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
```

### Authentication

Configure authentication providers:

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret

# GitHub OAuth
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret

# SAML/SSO (Enterprise)
SSO_ENABLED=true
```

## Upgrading

### Check Current Version

```bash
docker compose exec api /app/server version
```

### Upgrade Steps

```bash
# 1. Backup
./scripts/backup.sh

# 2. Pull new images
docker compose pull

# 3. Stop services
docker compose down

# 4. Run migrations
docker compose run --rm api /app/server migrate up

# 5. Start services
docker compose up -d

# 6. Verify
docker compose logs -f api
```

## Backup & Restore

### Automated Backups

Enable automated backups:

```yaml
services:
  backup:
    image: agenttrace/backup:latest
    environment:
      BACKUP_SCHEDULE: "0 2 * * *"  # 2 AM daily
      BACKUP_RETENTION_DAYS: 30
      S3_BUCKET: your-backup-bucket
```

### Manual Backup

```bash
# Backup PostgreSQL
docker compose exec postgres pg_dump -U agenttrace > backup.sql

# Backup ClickHouse
docker compose exec clickhouse clickhouse-client \
  --query "BACKUP TABLE traces, observations TO Disk('backups', 'backup.zip')"
```

### Restore

```bash
# Restore PostgreSQL
cat backup.sql | docker compose exec -T postgres psql -U agenttrace

# Restore ClickHouse
docker compose exec clickhouse clickhouse-client \
  --query "RESTORE TABLE traces, observations FROM Disk('backups', 'backup.zip')"
```

## Monitoring

### Health Checks

```bash
# API health
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "version": "1.0.0",
  "postgres": "connected",
  "clickhouse": "connected",
  "redis": "connected"
}
```

### Prometheus Metrics

Metrics are exposed at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

Key metrics:
- `agenttrace_traces_total` - Total traces ingested
- `agenttrace_observations_total` - Total observations
- `agenttrace_api_requests_total` - API request count
- `agenttrace_api_latency_seconds` - API latency histogram

### Grafana Dashboard

Import the AgentTrace dashboard:

1. Go to Grafana > Dashboards > Import
2. Enter dashboard ID: `12345`
3. Select your Prometheus data source

## Troubleshooting

### Common Issues

#### Container won't start

```bash
# Check logs
docker compose logs api

# Check resource usage
docker stats
```

#### Database connection failed

```bash
# Test PostgreSQL
docker compose exec postgres pg_isready

# Test ClickHouse
docker compose exec clickhouse clickhouse-client --query "SELECT 1"
```

#### High memory usage

ClickHouse can be memory-intensive. Adjust settings:

```xml
<!-- clickhouse/config.d/memory.xml -->
<max_memory_usage>8000000000</max_memory_usage>
<max_memory_usage_for_user>4000000000</max_memory_usage_for_user>
```

### Getting Help

- [GitHub Issues](https://github.com/agenttrace/agenttrace/issues)
- [Discord Community](https://discord.gg/agenttrace)
- [Enterprise Support](mailto:enterprise@agenttrace.io)
