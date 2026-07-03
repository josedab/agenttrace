---
sidebar_position: 6
title: "Backup & Restore"
description: "Back up and restore PostgreSQL, ClickHouse, and external object storage."
---

# Backup & Restore

AgentTrace data is split across three persistent systems:

| Component | Data |
|-----------|------|
| PostgreSQL | Users, organizations, projects, prompts, datasets, and configuration |
| ClickHouse | Traces, observations, scores, and analytics |
| External object storage | Exported objects, when enabled |

Redis queue state is ephemeral and is not included in backups.

## Docker Compose

The production Compose file configures a ClickHouse `backups` disk and includes tested scripts.

### Create a Backup

```bash
cd deploy
./backup.sh
```

Backups are written to `deploy/backups/<UTC timestamp>/` by default and contain:

- `postgres.dump`
- a ClickHouse backup archive
- `manifest.txt`

The API and worker are paused while PostgreSQL and ClickHouse are captured so both backups share one write boundary. They restart automatically even if the backup fails.

Choose another destination:

```bash
./backup.sh /secure/off-host/path/agenttrace-backup
```

Copy backup output off the Docker host. A backup stored beside the source volumes is not disaster recovery.

### Restore

Restore only into a disposable or intentionally replaced environment:

```bash
./restore.sh --yes /secure/off-host/path/agenttrace-backup
```

The script stops API, worker, and web; restores PostgreSQL and ClickHouse; clears newer Redis queue and rate-limit state; then restarts the application.

Run the deployment smoke test afterward:

```bash
./smoke-test.sh
```

## Kubernetes

Use provider-managed snapshots or scheduled backup jobs:

- PostgreSQL: `pg_dump -Fc` to durable object storage.
- ClickHouse: configure an S3 backup disk or use `clickhouse-backup`.
- External object storage: enable provider replication/versioning and test object restore separately.

Example PostgreSQL dump:

```bash
kubectl -n agenttrace exec sts/postgres -- \
  pg_dump -U agenttrace -Fc agenttrace > postgres.dump
```

The bundled ClickHouse configuration enables a local `backups` disk on its persistent volume. For production disaster recovery, replace or supplement it with off-cluster object storage.

## Schedule and Verification

Recommended minimum:

| Environment | Schedule | Restore test |
|-------------|----------|--------------|
| Staging | Daily | Monthly |
| Production | Every 6 hours for PostgreSQL; daily for ClickHouse/object storage | Monthly |

Monitor backup age, size, and exit status. A backup is not considered valid until it has been restored and the API smoke test passes.
