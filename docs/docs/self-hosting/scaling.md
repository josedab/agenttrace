---
sidebar_position: 5
title: "Scaling Guide"
description: "Horizontal scaling for API, ClickHouse replication, PostgreSQL read replicas, and Redis Sentinel."
---

# Scaling Guide

This guide covers scaling strategies for each AgentTrace component to handle increased load. AgentTrace is designed for horizontal scaling at every layer.

## Architecture Overview

```
                    ┌──────────────────┐
                    │  Load Balancer   │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │  API #1  │  │  API #2  │  │  API #3  │
        └──────────┘  └──────────┘  └──────────┘
              │              │              │
    ┌─────────┴──────────────┴──────────────┴─────────┐
    │                                                   │
    ▼                    ▼                ▼              ▼
┌──────────┐    ┌──────────────┐   ┌──────────┐  ┌──────────┐
│PostgreSQL│    │  ClickHouse  │   │  Redis   │  │  MinIO   │
│ Primary  │    │   Cluster    │   │ Sentinel │  │ Cluster  │
│ + Replica│    │              │   │          │  │          │
└──────────┘    └──────────────┘   └──────────┘  └──────────┘
```

## API Server Scaling

The API server is stateless and scales horizontally by adding more replicas.

### Docker Compose

```bash
docker compose up -d --scale api=3
```

Add a load balancer (nginx or traefik) in front of the API instances.

### Kubernetes

Use the Horizontal Pod Autoscaler defined in `deploy/kubernetes/hpa.yaml`:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: agenttrace-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: agenttrace-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### Worker Scaling

Scale background workers independently for heavy evaluation or ingestion workloads:

```bash
# Docker Compose
docker compose up -d --scale worker=3

# Kubernetes
kubectl -n agenttrace scale deploy/agenttrace-worker --replicas=3
```

## ClickHouse Scaling

ClickHouse stores all trace and observation data. It is typically the first component to need scaling.

### Vertical Scaling

For moderate workloads, increase ClickHouse resources:

| Daily Traces | Recommended Memory | Recommended CPU |
|--------------|-------------------|-----------------|
| < 100K | 8 GB | 2 cores |
| 100K–500K | 16 GB | 4 cores |
| 500K–2M | 32 GB | 8 cores |
| > 2M | 64 GB+ | 16+ cores |

### Replication

For high availability, configure ClickHouse with replication using ClickHouse Keeper:

```xml
<!-- clickhouse/config.d/cluster.xml -->
<remote_servers>
  <agenttrace_cluster>
    <shard>
      <replica>
        <host>clickhouse-1</host>
        <port>9000</port>
      </replica>
      <replica>
        <host>clickhouse-2</host>
        <port>9000</port>
      </replica>
    </shard>
  </agenttrace_cluster>
</remote_servers>
```

### Sharding

For very large deployments, shard data across multiple ClickHouse nodes:

```xml
<remote_servers>
  <agenttrace_cluster>
    <shard>
      <replica><host>ch-shard1-replica1</host></replica>
      <replica><host>ch-shard1-replica2</host></replica>
    </shard>
    <shard>
      <replica><host>ch-shard2-replica1</host></replica>
      <replica><host>ch-shard2-replica2</host></replica>
    </shard>
  </agenttrace_cluster>
</remote_servers>
```

## PostgreSQL Scaling

PostgreSQL stores metadata, users, prompts, and project configuration.

### Read Replicas

Offload read-heavy queries (dashboard, prompt fetching) to read replicas:

```bash
# Primary
DATABASE_URL=postgres://agenttrace:pass@pg-primary:5432/agenttrace

# Read replica (for dashboard queries)
DATABASE_READ_URL=postgres://agenttrace:pass@pg-replica:5432/agenttrace
```

### Connection Pooling

Use PgBouncer for connection pooling in high-concurrency environments:

```yaml
services:
  pgbouncer:
    image: edoburu/pgbouncer:latest
    environment:
      DATABASE_URL: postgres://agenttrace:pass@postgres:5432/agenttrace
      POOL_MODE: transaction
      MAX_CLIENT_CONN: 1000
      DEFAULT_POOL_SIZE: 50
```

## Redis Scaling

Redis is used for job queues (Asynq) and caching.

### Redis Sentinel

For high availability, deploy Redis Sentinel:

```yaml
services:
  redis-master:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}

  redis-replica:
    image: redis:7-alpine
    command: redis-server --replicaof redis-master 6379 --requirepass ${REDIS_PASSWORD}

  redis-sentinel:
    image: redis:7-alpine
    command: redis-sentinel /etc/sentinel.conf
```

Update the connection string:

```bash
REDIS_URL=redis+sentinel://sentinel1:26379,sentinel2:26379,sentinel3:26379/mymaster
```

## MinIO / Object Storage Scaling

For large-scale file storage, switch to a managed S3-compatible service:

```bash
MINIO_ENDPOINT=s3.amazonaws.com
S3_BUCKET=agenttrace-storage
S3_REGION=us-east-1
```

## Monitoring Scale

Monitor these metrics to know when to scale:

| Metric | Scale Signal | Action |
|--------|-------------|--------|
| API CPU > 70% | API overloaded | Add API replicas |
| API response time > 500ms | API overloaded | Add API replicas |
| ClickHouse query time > 2s | Query bottleneck | Add memory or replicas |
| PostgreSQL connections > 80% | Connection saturation | Add PgBouncer or replicas |
| Redis memory > 80% | Cache pressure | Increase memory or add replicas |
| Worker queue depth growing | Processing backlog | Add worker replicas |

## Related

- [Self-Hosting Overview](./overview.md) — architecture and system requirements
- [Kubernetes Deployment](./kubernetes.md) — deploying on Kubernetes
- [Configuration Reference](./configuration.md) — all environment variables
