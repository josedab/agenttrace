---
sidebar_position: 3
title: "Kubernetes Deployment"
description: "Deploy AgentTrace with the bundled Kustomize resources."
---

# Kubernetes Deployment

The Kustomize base deploys the API, worker, web dashboard, PostgreSQL, ClickHouse, Redis, ingress, services, and persistent volumes.

## Prerequisites

- Kubernetes 1.27+
- `kubectl` with Kustomize support
- A default or explicitly configured storage class
- An ingress controller
- API and web images published with the same immutable release tag

The published web image uses same-origin API requests by default, which matches the combined ingress manifest. Only provide a build-time URL when the browser must reach the API on a different origin:

```bash
docker build \
  --build-arg NEXT_PUBLIC_API_URL=https://agenttrace.example.com \
  -t agenttrace/web:0.1.0 \
  web

docker build \
  --build-arg VERSION=0.1.0 \
  -t agenttrace/api:0.1.0 \
  api
```

## Configure

```bash
cd deploy/kubernetes
cp secrets.yaml.example secrets.yaml
```

Replace every secret placeholder, including:

- PostgreSQL, ClickHouse, and Redis passwords
- JWT and NextAuth secrets

Update these placeholders in the manifests:

- `agenttrace.example.com` in `configmap.yaml`, `web.yaml`, and `ingress.yaml`
- image tags in `kustomization.yaml`
- storage sizes/classes in StatefulSets

Preview the rendered resources:

```bash
kubectl kustomize deploy/kubernetes/
```

## Deploy

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/secrets.yaml
kubectl apply -k deploy/kubernetes/
```

API and worker pods run `/app/migrate -path /app/migrations up` in init containers. Migrations are idempotent and serialized with a PostgreSQL advisory lock, so concurrent pod starts are safe.

## Verify

```bash
kubectl -n agenttrace get pods
kubectl -n agenttrace get svc
kubectl -n agenttrace rollout status deploy/agenttrace-api
kubectl -n agenttrace rollout status deploy/agenttrace-worker
kubectl -n agenttrace rollout status deploy/agenttrace-web
```

Check migration init-container logs:

```bash
kubectl -n agenttrace logs deploy/agenttrace-api -c migrate
```

Port-forward for a direct smoke test:

```bash
kubectl -n agenttrace port-forward svc/agenttrace-api 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/readyz
curl http://localhost:8080/metrics
```

## Upgrading

1. Back up PostgreSQL, ClickHouse, and any external object storage.
2. Publish API and web images with a new immutable tag.
3. Update `kustomization.yaml`.
4. Review `kubectl diff -k deploy/kubernetes/`.
5. Apply and monitor rollouts.

```bash
kubectl diff -k deploy/kubernetes/
kubectl apply -k deploy/kubernetes/
kubectl -n agenttrace rollout status deploy/agenttrace-api
kubectl -n agenttrace rollout status deploy/agenttrace-worker
kubectl -n agenttrace rollout status deploy/agenttrace-web
```

The init migration gate runs before each new API and worker pod becomes ready.

## Network Policies and Autoscaling

`network-policies.yaml` and `hpa.yaml` are provided as opt-in production overlays. Review namespace labels, allowed outbound destinations, and metrics-server availability before enabling them.

## Related

- [Configuration Reference](./configuration.md)
- [Docker Compose Deployment](./docker-compose.md)
- [Backup & Restore](./backup.md)
