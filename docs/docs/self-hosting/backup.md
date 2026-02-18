---
sidebar_position: 6
title: "Backup & Restore"
description: "Backup and restore procedures for PostgreSQL, ClickHouse, and MinIO in AgentTrace."
---

# Backup & Restore

This guide covers backup and restore procedures for all AgentTrace data stores. Regular backups are critical for disaster recovery and should be part of your production runbook.

## What to Back Up

| Component | Data | Priority |
|-----------|------|----------|
| PostgreSQL | Users, projects, prompts, API keys, configuration | **Critical** |
| ClickHouse | Traces, observations, scores, metrics | **Critical** |
| MinIO | Uploaded files, attachments | Important |
| Redis | Job queue state | Low (ephemeral) |

## PostgreSQL Backup

### Manual Backup

```bash
# Docker Compose
docker compose exec postgres pg_dump -U agenttrace -Fc agenttrace > pg_backup_$(date +%F).dump

# Kubernetes
kubectl -n agenttrace exec sts/postgres -- \
  pg_dump -U agenttrace -Fc agenttrace > pg_backup_$(date +%F).dump
```

The `-Fc` flag creates a custom-format dump that supports parallel restore and selective table restore.

### Automated Daily Backup

Add a backup service to your `docker-compose.override.yml`:

```yaml
services:
  pg-backup:
    image: prodrigestivill/postgres-backup-local:16
    environment:
      POSTGRES_HOST: postgres
      POSTGRES_DB: agenttrace
      POSTGRES_USER: agenttrace
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      SCHEDULE: "@daily"
      BACKUP_KEEP_DAYS: 30
      BACKUP_KEEP_WEEKS: 8
      BACKUP_KEEP_MONTHS: 6
    volumes:
      - pg_backups:/backups
```

### Restore PostgreSQL

```bash
# Docker Compose
docker compose exec -T postgres pg_restore -U agenttrace -d agenttrace --clean < pg_backup_2024-01-15.dump

# Kubernetes
kubectl -n agenttrace exec -i sts/postgres -- \
  pg_restore -U agenttrace -d agenttrace --clean < pg_backup_2024-01-15.dump
```

:::caution
The `--clean` flag drops existing objects before restoring. Ensure you have a current backup before restoring.
:::

## ClickHouse Backup

### Manual Backup

```bash
# Docker Compose — backup specific tables
docker compose exec clickhouse clickhouse-client \
  --query "BACKUP TABLE agenttrace.traces, agenttrace.observations, agenttrace.scores \
           TO Disk('backups', 'ch_backup_$(date +%F).zip')"

# Kubernetes
kubectl -n agenttrace exec sts/clickhouse -- clickhouse-client \
  --query "BACKUP TABLE agenttrace.traces, agenttrace.observations, agenttrace.scores \
           TO Disk('backups', 'ch_backup_$(date +%F).zip')"
```

### Backup to S3

For off-site backups, configure a backup disk pointing to S3:

```xml
<!-- clickhouse/config.d/backup_disk.xml -->
<storage_configuration>
  <disks>
    <s3_backups>
      <type>s3</type>
      <endpoint>https://s3.amazonaws.com/your-backup-bucket/clickhouse/</endpoint>
      <access_key_id>YOUR_KEY</access_key_id>
      <secret_access_key>YOUR_SECRET</secret_access_key>
    </s3_backups>
  </disks>
  <backups>
    <allowed_disk>s3_backups</allowed_disk>
  </backups>
</storage_configuration>
```

Then back up to S3:

```sql
BACKUP TABLE agenttrace.traces, agenttrace.observations
TO Disk('s3_backups', 'ch_backup_2024-01-15.zip');
```

### Restore ClickHouse

```bash
docker compose exec clickhouse clickhouse-client \
  --query "RESTORE TABLE agenttrace.traces, agenttrace.observations, agenttrace.scores \
           FROM Disk('backups', 'ch_backup_2024-01-15.zip')"
```

## MinIO Backup

### Using mc (MinIO Client)

```bash
# Install mc
docker run --rm -it --entrypoint sh minio/mc

# Configure alias
mc alias set agenttrace http://localhost:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD}

# Mirror to local directory
mc mirror agenttrace/agenttrace-bucket ./minio_backup_$(date +%F)/

# Mirror to another S3 bucket
mc mirror agenttrace/agenttrace-bucket s3/your-backup-bucket/minio/
```

### Restore MinIO

```bash
mc mirror ./minio_backup_2024-01-15/ agenttrace/agenttrace-bucket
```

## Full Backup Script

Create a comprehensive backup script:

```bash
#!/bin/bash
# scripts/backup.sh
set -euo pipefail

BACKUP_DIR="./backups/$(date +%F_%H%M)"
mkdir -p "$BACKUP_DIR"

echo "==> Backing up PostgreSQL..."
docker compose exec -T postgres pg_dump -U agenttrace -Fc agenttrace \
  > "$BACKUP_DIR/postgres.dump"

echo "==> Backing up ClickHouse..."
docker compose exec clickhouse clickhouse-client \
  --query "BACKUP TABLE agenttrace.traces, agenttrace.observations, agenttrace.scores \
           TO Disk('backups', 'backup_$(date +%F_%H%M).zip')"

echo "==> Backing up MinIO..."
docker run --rm --network=agenttrace_default \
  -v "$BACKUP_DIR:/backup" minio/mc \
  mirror agenttrace/agenttrace-bucket /backup/minio/

echo "==> Backup complete: $BACKUP_DIR"
```

## Backup Schedule Recommendations

| Environment | PostgreSQL | ClickHouse | MinIO |
|-------------|-----------|-----------|-------|
| Development | Weekly | Weekly | Manual |
| Staging | Daily | Daily | Weekly |
| Production | Every 6 hours | Daily | Daily |

## Retention Policy

| Backup Type | Retention |
|-------------|-----------|
| Daily | 30 days |
| Weekly | 12 weeks |
| Monthly | 12 months |

## Verification

Regularly test your backups by restoring to a staging environment:

```bash
# Spin up a test environment
docker compose -f docker-compose.test.yml up -d

# Restore and verify
pg_restore -U agenttrace -d agenttrace < backup/postgres.dump
curl http://localhost:8080/health
```

## Related

- [Self-Hosting Overview](./overview.md) — architecture overview
- [Docker Compose Deployment](./docker-compose.md) — Docker Compose setup
- [Configuration Reference](./configuration.md) — environment variables
