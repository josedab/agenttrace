# AgentTrace Kubernetes Deployment Guide

This guide covers deploying AgentTrace on Kubernetes. It includes raw manifests for flexibility and can be adapted for Helm, Kustomize, or other deployment tools.

## Architecture Overview

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                    Kubernetes Cluster                    │
                    │                                                          │
  ┌─────────┐       │   ┌──────────┐     ┌───────────────────────────────┐    │
  │ Ingress │───────┼──▶│ Service  │────▶│  API Deployment (3 replicas) │    │
  │ (HTTPS) │       │   └──────────┘     └───────────────────────────────┘    │
  └─────────┘       │                                   │                      │
                    │                                   ▼                      │
                    │                    ┌──────────────────────────────┐      │
                    │                    │  Worker Deployment (2 reps) │      │
                    │                    └──────────────────────────────┘      │
                    │                                   │                      │
                    │         ┌─────────────────────────┼───────────────┐      │
                    │         ▼                         ▼               ▼      │
                    │   ┌──────────┐           ┌──────────────┐  ┌──────────┐  │
                    │   │PostgreSQL│           │  ClickHouse  │  │  Redis   │  │
                    │   │(StatefulSet)         │(StatefulSet) │  │(StatefulSet) │
                    │   └──────────┘           └──────────────┘  └──────────┘  │
                    │         │                         │               │      │
                    │         ▼                         ▼               ▼      │
                    │   ┌──────────┐           ┌──────────────┐  ┌──────────┐  │
                    │   │  PVC     │           │    PVC       │  │  PVC     │  │
                    │   └──────────┘           └──────────────┘  └──────────┘  │
                    └─────────────────────────────────────────────────────────┘
```

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured with cluster access
- 8GB+ RAM available across nodes
- Storage class that supports ReadWriteOnce PVCs
- Ingress controller (nginx-ingress recommended)
- cert-manager (optional, for automatic TLS)

## Quick Start

### 1. Create Namespace

```bash
kubectl create namespace agenttrace
kubectl config set-context --current --namespace=agenttrace
```

### 2. Create Secrets

```bash
# Generate secure passwords
POSTGRES_PASSWORD=$(openssl rand -base64 32)
CLICKHOUSE_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -hex 16)

# Create secret
kubectl create secret generic agenttrace-secrets \
  --from-literal=postgres-password=$POSTGRES_PASSWORD \
  --from-literal=clickhouse-password=$CLICKHOUSE_PASSWORD \
  --from-literal=redis-password=$REDIS_PASSWORD \
  --from-literal=jwt-secret=$JWT_SECRET \
  --from-literal=encryption-key=$ENCRYPTION_KEY
```

### 3. Apply Manifests

```bash
# Apply in order (dependencies first)
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f postgres.yaml
kubectl apply -f clickhouse.yaml
kubectl apply -f redis.yaml

# Wait for databases to be ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
kubectl wait --for=condition=ready pod -l app=clickhouse --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis --timeout=60s

# Apply application
kubectl apply -f api.yaml
kubectl apply -f worker.yaml
kubectl apply -f ingress.yaml
```

### 4. Run Migrations

```bash
kubectl exec -it deployment/agenttrace-api -- /app/server migrate up
```

### 5. Verify Deployment

```bash
kubectl get pods
kubectl logs deployment/agenttrace-api
curl -k https://api.your-domain.com/health
```

## Configuration

### Environment Variables

All configuration is managed through ConfigMaps and Secrets. See `configmap.yaml` for the full list.

Key configuration sections:

| Section | Description |
|---------|-------------|
| Server | API host, port, environment |
| PostgreSQL | Connection settings, pool sizes |
| ClickHouse | Analytics database connection |
| Redis | Cache and queue connection |
| JWT | Authentication token settings |
| Worker | Background job configuration |
| Rate Limiting | Request throttling |

### Resource Recommendations

| Component | CPU Request | CPU Limit | Memory Request | Memory Limit |
|-----------|-------------|-----------|----------------|--------------|
| API | 200m | 1000m | 256Mi | 1Gi |
| Worker | 100m | 500m | 256Mi | 512Mi |
| PostgreSQL | 250m | 1000m | 512Mi | 2Gi |
| ClickHouse | 500m | 2000m | 1Gi | 4Gi |
| Redis | 100m | 500m | 128Mi | 512Mi |

### Storage Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| PostgreSQL | 10Gi | 50Gi | Metadata, users, projects |
| ClickHouse | 50Gi | 200Gi+ | Traces, observations |
| Redis | 1Gi | 5Gi | Ephemeral, can use emptyDir |

## Production Hardening

### 1. Enable Pod Security

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: agenttrace
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### 2. Network Policies

Apply network policies to restrict traffic:

```bash
kubectl apply -f network-policies.yaml
```

### 3. Pod Disruption Budgets

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: agenttrace-api-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: agenttrace-api
```

### 4. Horizontal Pod Autoscaling

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
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 5. External Secrets (Recommended)

For production, use External Secrets Operator with your secrets manager:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: agenttrace-secrets
spec:
  secretStoreRef:
    name: vault-backend
    kind: ClusterSecretStore
  target:
    name: agenttrace-secrets
  data:
  - secretKey: postgres-password
    remoteRef:
      key: agenttrace/postgres
      property: password
```

## Monitoring

### Prometheus ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: agenttrace-api
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app: agenttrace-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Grafana Dashboard

Import the AgentTrace dashboard from `monitoring/grafana-dashboard.json` or create alerts for:

- API error rate > 1%
- Request latency p99 > 500ms
- Worker queue depth > 1000
- Database connection pool exhaustion
- ClickHouse query duration > 10s

## Backup and Disaster Recovery

### PostgreSQL Backup with CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres -U agenttrace agenttrace | \
              gzip > /backup/agenttrace-$(date +%Y%m%d).sql.gz
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: agenttrace-secrets
                  key: postgres-password
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: postgres-backup-pvc
          restartPolicy: OnFailure
```

### ClickHouse Backup

Use clickhouse-backup for ClickHouse data:

```bash
kubectl exec -it clickhouse-0 -- clickhouse-backup create daily-backup
kubectl exec -it clickhouse-0 -- clickhouse-backup upload daily-backup
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl describe pod <pod-name>

# Check events
kubectl get events --sort-by='.lastTimestamp'

# Check resource constraints
kubectl top nodes
kubectl top pods
```

### Database Connection Issues

```bash
# Test PostgreSQL connectivity
kubectl run -it --rm pg-client --image=postgres:16-alpine -- \
  psql "postgresql://agenttrace:$POSTGRES_PASSWORD@postgres:5432/agenttrace"

# Test ClickHouse connectivity
kubectl run -it --rm ch-client --image=clickhouse/clickhouse-client -- \
  clickhouse-client --host clickhouse --user agenttrace --password $CLICKHOUSE_PASSWORD
```

### API Health Check Failing

```bash
# Check API logs
kubectl logs deployment/agenttrace-api --tail=100

# Check readiness probe
kubectl describe pod -l app=agenttrace-api | grep -A5 Readiness

# Test internal health endpoint
kubectl exec -it deployment/agenttrace-api -- wget -qO- http://localhost:8080/health
```

### Worker Not Processing Jobs

```bash
# Check worker logs
kubectl logs deployment/agenttrace-worker --tail=100

# Check Redis connection
kubectl exec -it deployment/agenttrace-worker -- redis-cli -h redis -a $REDIS_PASSWORD PING

# Check queue depth
kubectl exec -it redis-0 -- redis-cli -a $REDIS_PASSWORD LLEN asynq:default:pending
```

## Upgrades

### Rolling Update

```bash
# Update image
kubectl set image deployment/agenttrace-api api=agenttrace/api:v1.2.0
kubectl set image deployment/agenttrace-worker worker=agenttrace/api:v1.2.0

# Run migrations
kubectl exec -it deployment/agenttrace-api -- /app/server migrate up

# Verify
kubectl rollout status deployment/agenttrace-api
```

### Rollback

```bash
kubectl rollout undo deployment/agenttrace-api
kubectl rollout undo deployment/agenttrace-worker
```

## Uninstall

```bash
kubectl delete namespace agenttrace
```

This will remove all resources including PVCs. For data preservation, backup databases first.
