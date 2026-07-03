---
sidebar_position: 2
title: "Docker Compose Deployment"
description: "Deploy the complete AgentTrace stack with migrations, health gates, and pinned dependencies."
---

# Docker Compose Deployment

The production Compose stack runs PostgreSQL, ClickHouse, Redis, migrations, the API, worker, and web dashboard. API and worker startup is blocked until database migrations complete successfully.

## Prerequisites

- Docker Engine 24+
- Docker Compose 2.20+
- 4 CPU cores, 8 GB RAM, and 50 GB disk minimum
- A public application URL and TLS termination for production

## Configure

```bash
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/deploy
cp .env.example .env
```

Replace every `change-me` value. In particular:

```bash
VERSION=0.1.0
POSTGRES_PASSWORD="$(openssl rand -base64 24)"
CLICKHOUSE_PASSWORD="$(openssl rand -base64 24)"
REDIS_PASSWORD="$(openssl rand -base64 24)"
JWT_SECRET="$(openssl rand -base64 32)"
NEXTAUTH_SECRET="$(openssl rand -base64 32)"

NEXTAUTH_URL=https://agenttrace.example.com
NEXT_PUBLIC_API_URL=https://agenttrace.example.com
CORS_ALLOWED_ORIGINS=https://agenttrace.example.com
```

`NEXT_PUBLIC_API_URL` is embedded into the web image at build time. Rebuild the web image whenever it changes.

The bundled PostgreSQL container does not terminate TLS, so the example uses:

```bash
POSTGRES_SSL_MODE=disable
POSTGRES_ALLOW_INSECURE=true
```

Use `POSTGRES_SSL_MODE=require` and remove the insecure override when connecting to a TLS-enabled managed database.

## Deploy

```bash
docker compose up -d --build
```

The `migrate` service is a one-shot startup gate. No separate migration command is required.

To add the bundled Nginx reverse proxy:

```bash
docker compose --profile with-proxy up -d --build
```

Terminate TLS at your load balancer, ingress, or a hardened Nginx override.

## Verify

```bash
docker compose ps
docker compose logs migrate
./smoke-test.sh
```

The migration and bucket services must exit with code `0`; all long-running services should be healthy or running.

## Manual Migration Check

Migrations are idempotent and serialized with a PostgreSQL advisory lock:

```bash
docker compose run --rm migrate
```

## Upgrade

1. Back up PostgreSQL, ClickHouse, and any external object storage.
2. Set `VERSION` to the immutable application release tag.
3. Rebuild or pull the new images.
4. Start the stack; migrations run before API and worker rollout.
5. Run the smoke test.

```bash
docker compose pull
docker compose up -d --build
./smoke-test.sh
```

Do not delete database volumes during an upgrade.

If upgrading from a pre-release build that authenticated with a bare `pk-` public identifier, rotate those keys after upgrade. Current releases authenticate only with the one-time `sk-at-...` secret (or an explicit public/secret Basic Auth pair).

## Backup

- PostgreSQL: schedule `pg_dump -Fc` and test `pg_restore`.
- ClickHouse: configure a backup disk or use `clickhouse-backup`; snapshotting a live data directory alone is not a complete backup strategy.
- External object storage: use the provider's versioning, replication, and backup tooling.

See [Backup & Restore](./backup.md) for platform-specific procedures.

## Troubleshooting

```bash
docker compose logs migrate
docker compose logs api worker
docker compose exec postgres pg_isready -U agenttrace -d agenttrace
docker compose exec clickhouse clickhouse-client \
  --user agenttrace --password "$CLICKHOUSE_PASSWORD" --query "SELECT 1"
docker compose exec redis redis-cli -a "$REDIS_PASSWORD" ping
```

## Related

- [Configuration Reference](./configuration.md)
- [Kubernetes Deployment](./kubernetes.md)
- [Backup & Restore](./backup.md)
