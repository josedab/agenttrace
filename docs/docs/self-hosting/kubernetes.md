---
sidebar_position: 3
title: "Kubernetes Deployment"
description: "Deploy AgentTrace on Kubernetes using kustomize with the manifests in deploy/kubernetes/."
---

# Kubernetes Deployment

This guide covers deploying AgentTrace on Kubernetes using the kustomize manifests provided in `deploy/kubernetes/`. This is the recommended approach for production environments requiring high availability and scalability.

## Prerequisites

- Kubernetes cluster 1.26+
- `kubectl` configured with cluster access
- `kustomize` v5+ (or `kubectl` with built-in kustomize)
- Persistent volume provisioner (for database storage)
- Ingress controller (nginx-ingress or traefik recommended)

## Repository Structure

The Kubernetes manifests are located at `deploy/kubernetes/`:

```
deploy/kubernetes/
├── kustomization.yaml       # Kustomize configuration
├── namespace.yaml           # agenttrace namespace
├── configmap.yaml           # Application configuration
├── secrets.yaml.example     # Secret template (copy and fill in)
├── api.yaml                 # API server deployment + service
├── worker.yaml              # Background worker deployment
├── postgres.yaml            # PostgreSQL StatefulSet
├── clickhouse.yaml          # ClickHouse StatefulSet
├── redis.yaml               # Redis deployment
├── ingress.yaml             # Ingress resource
├── network-policies.yaml    # Network segmentation
└── hpa.yaml                 # Horizontal Pod Autoscaler
```

## Deployment Steps

### 1. Create Secrets

Copy the secrets template and fill in your values:

```bash
cd deploy/kubernetes
cp secrets.yaml.example secrets.yaml
```

Edit `secrets.yaml` with base64-encoded values:

```bash
# Generate base64 values
echo -n 'your-secure-password' | base64

# Generate a random secret
openssl rand -base64 32 | tr -d '\n' | base64
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: agenttrace-secrets
  namespace: agenttrace
type: Opaque
data:
  postgres-password: <base64-encoded>
  clickhouse-password: <base64-encoded>
  redis-password: <base64-encoded>
  jwt-secret: <base64-encoded>
  encryption-key: <base64-encoded>
  nextauth-secret: <base64-encoded>
```

### 2. Configure the Application

Edit `configmap.yaml` with your domain and settings:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agenttrace-config
  namespace: agenttrace
data:
  NEXTAUTH_URL: "https://agenttrace.your-company.com"
  API_HOST: "0.0.0.0"
  API_PORT: "8080"
  LOG_LEVEL: "info"
```

### 3. Configure Ingress

Edit `ingress.yaml` with your domain and TLS settings:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agenttrace-ingress
  namespace: agenttrace
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - agenttrace.your-company.com
      secretName: agenttrace-tls
  rules:
    - host: agenttrace.your-company.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: agenttrace-api
                port:
                  number: 8080
          - path: /
            pathType: Prefix
            backend:
              service:
                name: agenttrace-web
                port:
                  number: 3000
```

### 4. Apply Manifests

```bash
# Preview what will be applied
kubectl kustomize deploy/kubernetes/

# Apply all resources
kubectl apply -k deploy/kubernetes/

# Verify pods are running
kubectl -n agenttrace get pods
```

### 5. Run Migrations

```bash
kubectl -n agenttrace exec deploy/agenttrace-api -- /app/server migrate up
```

### 6. Verify Deployment

```bash
# Check pod status
kubectl -n agenttrace get pods

# Check service endpoints
kubectl -n agenttrace get svc

# Test API health
kubectl -n agenttrace port-forward svc/agenttrace-api 8080:8080
curl http://localhost:8080/health
```

## Storage Configuration

### Persistent Volumes

The StatefulSets for PostgreSQL and ClickHouse require persistent volumes. Adjust storage class and size in the manifests:

```yaml
volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      storageClassName: gp3    # Adjust for your cloud provider
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi       # Adjust based on expected volume
```

## Autoscaling

The `hpa.yaml` manifest configures horizontal pod autoscaling for the API server:

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

## Upgrading

```bash
# Update image tags in kustomization.yaml, then:
kubectl apply -k deploy/kubernetes/

# Run any new migrations
kubectl -n agenttrace exec deploy/agenttrace-api -- /app/server migrate up

# Monitor rollout
kubectl -n agenttrace rollout status deploy/agenttrace-api
```

## Related

- [Configuration Reference](./configuration.md) — all environment variables
- [Scaling Guide](./scaling.md) — scaling strategies for each component
- [Backup & Restore](./backup.md) — backup procedures for Kubernetes
