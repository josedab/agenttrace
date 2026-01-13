# AgentTrace Production Deployment Guide

This guide covers deploying AgentTrace in a production environment. AgentTrace consists of several components:

- **API Server** - HTTP/GraphQL API for trace ingestion and querying
- **Worker** - Background job processor for cost calculation, evaluations, and data retention
- **PostgreSQL** - Metadata storage (users, organizations, projects, prompts, datasets)
- **ClickHouse** - High-performance trace and observation storage
- **Redis** - Job queue, rate limiting, and caching
- **MinIO** (optional) - Object storage for exports

## Prerequisites

- Docker and Docker Compose (v2.0+)
- 4GB RAM minimum (8GB+ recommended for production)
- 20GB disk space minimum (SSD recommended for ClickHouse)
- Domain name with SSL certificate (for production)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/deploy
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your production values
```

3. Start all services:
```bash
docker compose up -d
```

4. Run database migrations:
```bash
docker compose exec api /app/server migrate up
```

5. Verify deployment:
```bash
curl http://localhost:8080/health
```

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_PASSWORD` | PostgreSQL password | `strong-password-here` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | `strong-password-here` |
| `REDIS_PASSWORD` | Redis password | `strong-password-here` |
| `JWT_SECRET` | JWT signing secret (32+ chars) | `your-256-bit-secret` |
| `ENCRYPTION_KEY` | Encryption key for API keys (32 chars) | `32-char-encryption-key` |
| `NEXTAUTH_URL` | Web app public URL | `https://app.example.com` |
| `NEXTAUTH_SECRET` | NextAuth.js secret | `your-nextauth-secret` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_PORT` | API server port | `8080` |
| `WEB_PORT` | Web frontend port | `3000` |
| `POSTGRES_USER` | PostgreSQL user | `agenttrace` |
| `POSTGRES_DB` | PostgreSQL database | `agenttrace` |
| `CLICKHOUSE_USER` | ClickHouse user | `default` |
| `CLICKHOUSE_DB` | ClickHouse database | `agenttrace` |
| `MINIO_ROOT_USER` | MinIO root user | `agenttrace` |
| `MINIO_ROOT_PASSWORD` | MinIO root password | Required if using MinIO |
| `OPENAI_API_KEY` | OpenAI API key for LLM evaluations | Empty |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | Empty |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | Empty |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | Empty |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | Empty |

## Production Configuration

### 1. Generate Secure Secrets

```bash
# Generate JWT secret
openssl rand -base64 32

# Generate encryption key (exactly 32 characters)
openssl rand -hex 16

# Generate NextAuth secret
openssl rand -base64 32
```

### 2. Configure TLS/SSL

For production, configure the nginx reverse proxy with SSL certificates:

```bash
# Create certs directory
mkdir -p certs

# Option 1: Use Let's Encrypt (certbot)
certbot certonly --standalone -d api.example.com -d app.example.com

# Option 2: Use your own certificates
cp /path/to/fullchain.pem certs/
cp /path/to/privkey.pem certs/
```

Create `nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        server api:8080;
    }

    upstream web {
        server web:3000;
    }

    # Redirect HTTP to HTTPS
    server {
        listen 80;
        server_name api.example.com app.example.com;
        return 301 https://$host$request_uri;
    }

    # API server
    server {
        listen 443 ssl http2;
        server_name api.example.com;

        ssl_certificate /etc/nginx/certs/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;

        location / {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # SSE support
            proxy_buffering off;
            proxy_cache off;
            proxy_read_timeout 86400s;
        }
    }

    # Web frontend
    server {
        listen 443 ssl http2;
        server_name app.example.com;

        ssl_certificate /etc/nginx/certs/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;

        location / {
            proxy_pass http://web;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

Start with the nginx proxy profile:

```bash
docker compose --profile with-proxy up -d
```

### 3. Database Configuration

#### PostgreSQL Tuning

For production workloads, create a custom PostgreSQL config:

```sql
-- postgresql.conf optimizations
shared_buffers = 256MB
effective_cache_size = 768MB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 4MB
min_wal_size = 1GB
max_wal_size = 4GB
max_worker_processes = 4
max_parallel_workers_per_gather = 2
max_parallel_workers = 4
max_parallel_maintenance_workers = 2
```

#### ClickHouse Tuning

For high-volume trace ingestion:

```xml
<!-- config.d/performance.xml -->
<clickhouse>
    <max_concurrent_queries>100</max_concurrent_queries>
    <max_memory_usage>4000000000</max_memory_usage>
    <max_memory_usage_for_all_queries>6000000000</max_memory_usage_for_all_queries>

    <merge_tree>
        <max_suspicious_broken_parts>5</max_suspicious_broken_parts>
        <parts_to_delay_insert>150</parts_to_delay_insert>
        <parts_to_throw_insert>300</parts_to_throw_insert>
        <max_avg_part_size_for_too_many_parts>1073741824</max_avg_part_size_for_too_many_parts>
    </merge_tree>
</clickhouse>
```

### 4. Scaling

#### Horizontal Scaling

For high-traffic deployments:

```yaml
# docker-compose.override.yml
services:
  api:
    deploy:
      replicas: 3

  worker:
    deploy:
      replicas: 2
```

With a load balancer, update nginx upstream:

```nginx
upstream api {
    least_conn;
    server api-1:8080;
    server api-2:8080;
    server api-3:8080;
}
```

#### Redis Cluster (Optional)

For high availability, use Redis Sentinel or Redis Cluster:

```yaml
services:
  redis-master:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}

  redis-replica:
    image: redis:7-alpine
    command: redis-server --appendonly yes --replicaof redis-master 6379 --masterauth ${REDIS_PASSWORD} --requirepass ${REDIS_PASSWORD}
    depends_on:
      - redis-master
```

## Database Migrations

### Running Migrations

```bash
# PostgreSQL migrations
docker compose exec api /app/server migrate up

# ClickHouse migrations (if separate)
docker compose exec api /app/server migrate-clickhouse up
```

### Rollback

```bash
docker compose exec api /app/server migrate down 1
```

## Monitoring

### Health Endpoints

| Endpoint | Description |
|----------|-------------|
| `/health` | Full health check (all dependencies) |
| `/livez` | Kubernetes liveness probe |
| `/readyz` | Kubernetes readiness probe |
| `/version` | Version and uptime info |

### Prometheus Metrics

The API exposes Prometheus metrics at `/metrics` (when enabled):

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'agenttrace-api'
    static_configs:
      - targets: ['api:8080']
```

Key metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency histogram
- `traces_ingested_total` - Total traces ingested
- `observations_ingested_total` - Total observations ingested
- `worker_jobs_processed_total` - Background jobs processed

### Logging

Logs are in JSON format for easy parsing:

```bash
# View API logs
docker compose logs -f api

# View worker logs
docker compose logs -f worker
```

Configure log aggregation with Loki, Elasticsearch, or your preferred solution.

## Backup and Recovery

### PostgreSQL Backup

```bash
# Create backup
docker compose exec postgres pg_dump -U agenttrace agenttrace > backup.sql

# Restore backup
docker compose exec -T postgres psql -U agenttrace agenttrace < backup.sql
```

### ClickHouse Backup

```bash
# Using clickhouse-backup tool
docker compose exec clickhouse clickhouse-backup create backup_name
docker compose exec clickhouse clickhouse-backup upload backup_name
```

### Automated Backups

Create a cron job for regular backups:

```bash
# /etc/cron.d/agenttrace-backup
0 2 * * * root /opt/agenttrace/scripts/backup.sh >> /var/log/agenttrace-backup.log 2>&1
```

## Security Checklist

- [ ] Use strong, unique passwords for all services
- [ ] Enable TLS/SSL for all external traffic
- [ ] Configure firewall rules (only expose ports 80/443)
- [ ] Set up rate limiting (enabled by default)
- [ ] Rotate JWT secrets periodically
- [ ] Enable PostgreSQL SSL (`sslmode=require`)
- [ ] Configure ClickHouse user permissions
- [ ] Set up audit logging
- [ ] Regular security updates

## Troubleshooting

### Common Issues

**API server won't start:**
```bash
# Check logs
docker compose logs api

# Verify database connectivity
docker compose exec api /app/server healthcheck
```

**ClickHouse connection issues:**
```bash
# Verify ClickHouse is running
docker compose exec clickhouse clickhouse-client --query "SELECT 1"

# Check ClickHouse logs
docker compose logs clickhouse
```

**High memory usage:**
```bash
# Check container stats
docker stats

# Adjust ClickHouse memory limits in config
```

**Slow trace queries:**
```bash
# Check ClickHouse query log
docker compose exec clickhouse clickhouse-client --query "SELECT * FROM system.query_log ORDER BY event_time DESC LIMIT 10"
```

### Getting Help

- GitHub Issues: https://github.com/agenttrace/agenttrace/issues
- Documentation: https://docs.agenttrace.io
- Community Discord: https://discord.gg/agenttrace

## Upgrading

1. Pull latest images:
```bash
docker compose pull
```

2. Review changelog for breaking changes

3. Run migrations:
```bash
docker compose exec api /app/server migrate up
```

4. Restart services:
```bash
docker compose up -d
```

5. Verify health:
```bash
curl http://localhost:8080/health
```

## Kubernetes Deployment

For Kubernetes deployments, see the manifests in `/deploy/kubernetes/`.

### Quick Start

```bash
# Create namespace and secrets
kubectl create namespace agenttrace
kubectl create secret generic agenttrace-secrets \
  --from-literal=postgres-password=$(openssl rand -base64 32) \
  --from-literal=clickhouse-password=$(openssl rand -base64 32) \
  --from-literal=redis-password=$(openssl rand -base64 32) \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=encryption-key=$(openssl rand -hex 16) \
  -n agenttrace

# Apply with Kustomize
kubectl apply -k deploy/kubernetes/

# Or apply manifests individually
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/configmap.yaml
kubectl apply -f deploy/kubernetes/postgres.yaml
kubectl apply -f deploy/kubernetes/clickhouse.yaml
kubectl apply -f deploy/kubernetes/redis.yaml
kubectl apply -f deploy/kubernetes/api.yaml
kubectl apply -f deploy/kubernetes/worker.yaml
kubectl apply -f deploy/kubernetes/ingress.yaml

# Run migrations
kubectl exec -it deployment/agenttrace-api -n agenttrace -- /app/server migrate up
```

For production hardening, also apply:
- `network-policies.yaml` - Network segmentation
- `hpa.yaml` - Horizontal Pod Autoscaling

See `/deploy/kubernetes/README.md` for the full Kubernetes deployment guide.
