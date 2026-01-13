# AgentTrace Monitoring Guide

This guide covers setting up monitoring and observability for AgentTrace deployments.

## Overview

AgentTrace exposes metrics, structured logs, and health endpoints for comprehensive monitoring. The recommended stack is:

- **Prometheus** - Metrics collection
- **Grafana** - Visualization and alerting
- **Loki** - Log aggregation (optional)
- **AlertManager** - Alert routing

## Health Endpoints

| Endpoint | Description | Use Case |
|----------|-------------|----------|
| `/health` | Full health check with dependencies | Human debugging |
| `/livez` | Liveness probe | Kubernetes liveness |
| `/readyz` | Readiness probe | Kubernetes readiness, load balancers |
| `/version` | Version and uptime | Debugging, dashboards |

### Health Response Format

```json
{
  "status": "healthy",
  "checks": {
    "postgres": {"status": "healthy", "latency_ms": 2},
    "clickhouse": {"status": "healthy", "latency_ms": 5},
    "redis": {"status": "healthy", "latency_ms": 1}
  },
  "version": "1.0.0",
  "uptime": "2h15m30s"
}
```

## Prometheus Metrics

The API exposes Prometheus metrics at `/metrics` with the following key metrics:

### HTTP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `http_request_duration_seconds` | Histogram | Request latency distribution |
| `http_request_size_bytes` | Histogram | Request body sizes |
| `http_response_size_bytes` | Histogram | Response body sizes |

### Business Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `traces_ingested_total` | Counter | Total traces ingested |
| `observations_ingested_total` | Counter | Total observations ingested |
| `scores_created_total` | Counter | Total scores created |
| `api_keys_created_total` | Counter | API keys generated |

### Worker Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `worker_jobs_processed_total` | Counter | Jobs processed by type and status |
| `worker_job_duration_seconds` | Histogram | Job processing time |
| `worker_queue_depth` | Gauge | Current queue depth by queue |
| `worker_active_jobs` | Gauge | Currently processing jobs |

### Database Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `db_connections_active` | Gauge | Active database connections |
| `db_connections_idle` | Gauge | Idle database connections |
| `db_query_duration_seconds` | Histogram | Query execution time |
| `clickhouse_insert_rows_total` | Counter | Rows inserted to ClickHouse |

## Prometheus Configuration

### Kubernetes ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: agenttrace-api
  namespace: agenttrace
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app: agenttrace-api
  namespaceSelector:
    matchNames:
    - agenttrace
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
```

### Docker Compose / Static Config

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'agenttrace-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: /metrics
    scheme: http

  - job_name: 'agenttrace-worker'
    static_configs:
      - targets: ['worker:8081']
    metrics_path: /metrics
    scheme: http

  - job_name: 'clickhouse'
    static_configs:
      - targets: ['clickhouse:9363']
    metrics_path: /metrics

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

## Grafana Dashboards

### Recommended Dashboards

1. **AgentTrace Overview** - High-level metrics
2. **API Performance** - Latency, throughput, errors
3. **Worker Performance** - Queue depths, job processing
4. **Database Health** - Connection pools, query performance
5. **ClickHouse Metrics** - Inserts, merges, query times

### Sample Dashboard JSON

```json
{
  "title": "AgentTrace Overview",
  "panels": [
    {
      "title": "Request Rate",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(http_requests_total{job=\"agenttrace-api\"}[5m])",
          "legendFormat": "{{method}} {{path}}"
        }
      ]
    },
    {
      "title": "Error Rate",
      "type": "singlestat",
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{job=\"agenttrace-api\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{job=\"agenttrace-api\"}[5m])) * 100"
        }
      ]
    },
    {
      "title": "P99 Latency",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"agenttrace-api\"}[5m]))",
          "legendFormat": "p99"
        }
      ]
    },
    {
      "title": "Traces Ingested",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(traces_ingested_total[5m])",
          "legendFormat": "traces/sec"
        }
      ]
    },
    {
      "title": "Worker Queue Depth",
      "type": "graph",
      "targets": [
        {
          "expr": "worker_queue_depth",
          "legendFormat": "{{queue}}"
        }
      ]
    }
  ]
}
```

## Alerting Rules

### Critical Alerts

```yaml
groups:
- name: agenttrace-critical
  rules:
  - alert: AgentTraceDown
    expr: up{job="agenttrace-api"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "AgentTrace API is down"
      description: "AgentTrace API has been down for more than 1 minute"

  - alert: HighErrorRate
    expr: |
      sum(rate(http_requests_total{job="agenttrace-api",status=~"5.."}[5m]))
      / sum(rate(http_requests_total{job="agenttrace-api"}[5m])) > 0.05
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      description: "Error rate is above 5% for the last 5 minutes"

  - alert: DatabaseConnectionExhausted
    expr: db_connections_active / db_connections_max > 0.9
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Database connections nearly exhausted"
      description: "More than 90% of database connections are in use"
```

### Warning Alerts

```yaml
groups:
- name: agenttrace-warning
  rules:
  - alert: HighLatency
    expr: |
      histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="agenttrace-api"}[5m])) > 1
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "High API latency"
      description: "P99 latency is above 1 second"

  - alert: WorkerQueueBacklog
    expr: worker_queue_depth > 10000
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "Worker queue backlog"
      description: "Worker queue depth is above 10000 for 15 minutes"

  - alert: ClickHouseSlowQueries
    expr: |
      histogram_quantile(0.99, rate(clickhouse_query_duration_seconds_bucket[5m])) > 10
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "Slow ClickHouse queries"
      description: "P99 ClickHouse query time is above 10 seconds"

  - alert: LowDiskSpace
    expr: |
      (node_filesystem_avail_bytes{mountpoint="/var/lib/clickhouse"}
      / node_filesystem_size_bytes{mountpoint="/var/lib/clickhouse"}) < 0.15
    for: 30m
    labels:
      severity: warning
    annotations:
      summary: "Low disk space on ClickHouse volume"
      description: "Less than 15% disk space remaining"
```

## Structured Logging

AgentTrace outputs JSON-formatted logs for easy parsing. Key fields:

| Field | Description |
|-------|-------------|
| `level` | Log level (debug, info, warn, error) |
| `ts` | Timestamp (ISO 8601) |
| `msg` | Log message |
| `caller` | Source file and line |
| `request_id` | Request correlation ID |
| `user_id` | Authenticated user ID |
| `project_id` | Project context |
| `trace_id` | OpenTelemetry trace ID |
| `duration_ms` | Operation duration |

### Sample Log Entry

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "msg": "trace ingested",
  "caller": "handler/otel.go:125",
  "request_id": "abc123",
  "project_id": "proj_xyz",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "observations_count": 15,
  "duration_ms": 45
}
```

### Loki Configuration

```yaml
# loki-config.yaml
server:
  http_listen_port: 3100

ingester:
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1

schema_config:
  configs:
    - from: 2024-01-01
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    cache_location: /loki/cache
    shared_store: filesystem
  filesystem:
    directory: /loki/chunks
```

### Promtail Configuration

```yaml
# promtail-config.yaml
server:
  http_listen_port: 9080

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: agenttrace
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'
      - source_labels: ['__meta_docker_container_label_com_docker_compose_service']
        target_label: 'service'
    pipeline_stages:
      - json:
          expressions:
            level: level
            msg: msg
            request_id: request_id
      - labels:
          level:
```

## Kubernetes Monitoring Stack

### Install Prometheus Operator

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin
```

### Install Loki Stack

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set promtail.enabled=true \
  --set grafana.enabled=false  # Use existing Grafana
```

### Add AgentTrace Datasource to Grafana

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitoring
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus-server:80
      isDefault: true
    - name: Loki
      type: loki
      url: http://loki:3100
```

## Performance Baselines

### Recommended SLOs

| Metric | Target | Critical |
|--------|--------|----------|
| API Availability | 99.9% | 99.5% |
| P50 Latency | < 50ms | < 200ms |
| P99 Latency | < 500ms | < 2s |
| Error Rate | < 0.1% | < 1% |
| Trace Ingestion | > 1000/s | > 100/s |

### Capacity Planning

Monitor these metrics for scaling decisions:

- **CPU Usage** > 70% sustained: Add API replicas
- **Memory Usage** > 80%: Increase limits or add replicas
- **Worker Queue Depth** > 5000: Add worker replicas
- **ClickHouse Disk** > 80%: Expand storage or enable retention
- **PostgreSQL Connections** > 80%: Increase pool size

## Troubleshooting

### High Latency

1. Check ClickHouse query times:
   ```promql
   histogram_quantile(0.99, rate(clickhouse_query_duration_seconds_bucket[5m]))
   ```

2. Check database connection pool:
   ```promql
   db_connections_active / db_connections_max
   ```

3. Review slow query logs in ClickHouse:
   ```sql
   SELECT * FROM system.query_log
   WHERE query_duration_ms > 1000
   ORDER BY event_time DESC
   LIMIT 20
   ```

### High Error Rate

1. Check error breakdown:
   ```promql
   sum by (path, status) (rate(http_requests_total{status=~"5.."}[5m]))
   ```

2. Review error logs:
   ```logql
   {job="agenttrace"} |= "error" | json
   ```

3. Check database connectivity:
   ```bash
   kubectl exec -it deployment/agenttrace-api -- wget -qO- http://localhost:8080/health
   ```

### Worker Backlog

1. Check queue depth by priority:
   ```promql
   worker_queue_depth by (queue)
   ```

2. Check job processing rate:
   ```promql
   rate(worker_jobs_processed_total[5m])
   ```

3. Look for failed jobs:
   ```logql
   {job="agenttrace-worker"} |= "failed" | json
   ```
