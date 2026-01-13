# AgentTrace Disaster Recovery Runbook

This runbook covers procedures for backup, recovery, and disaster scenarios for AgentTrace deployments.

## Table of Contents

1. [Backup Procedures](#backup-procedures)
2. [Recovery Procedures](#recovery-procedures)
3. [Disaster Scenarios](#disaster-scenarios)
4. [Testing Procedures](#testing-procedures)
5. [Contacts and Escalation](#contacts-and-escalation)

---

## Backup Procedures

### PostgreSQL Backup

PostgreSQL stores critical metadata: users, organizations, projects, API keys, prompts, datasets.

#### Manual Backup

```bash
# Docker Compose
docker compose exec postgres pg_dump -U agenttrace -Fc agenttrace > backup_$(date +%Y%m%d_%H%M%S).dump

# Kubernetes
kubectl exec -it postgres-0 -n agenttrace -- pg_dump -U agenttrace -Fc agenttrace > backup_$(date +%Y%m%d_%H%M%S).dump
```

#### Automated Daily Backup Script

```bash
#!/bin/bash
# /opt/agenttrace/scripts/backup-postgres.sh

set -euo pipefail

BACKUP_DIR="/backups/postgres"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/agenttrace_${DATE}.dump"

# Create backup directory
mkdir -p "${BACKUP_DIR}"

# Create backup
PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump \
  -h "${POSTGRES_HOST:-localhost}" \
  -U "${POSTGRES_USER:-agenttrace}" \
  -Fc \
  agenttrace > "${BACKUP_FILE}"

# Compress backup
gzip "${BACKUP_FILE}"

# Upload to S3 (optional)
if [ -n "${S3_BUCKET:-}" ]; then
  aws s3 cp "${BACKUP_FILE}.gz" "s3://${S3_BUCKET}/postgres/"
fi

# Clean old backups
find "${BACKUP_DIR}" -name "*.dump.gz" -mtime +${RETENTION_DAYS} -delete

echo "PostgreSQL backup completed: ${BACKUP_FILE}.gz"
```

#### Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: agenttrace
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16-alpine
            command: ["/bin/sh", "-c"]
            args:
            - |
              pg_dump -h postgres -U agenttrace -Fc agenttrace | \
              gzip > /backup/postgres_$(date +%Y%m%d).dump.gz
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
              claimName: backup-pvc
          restartPolicy: OnFailure
```

### ClickHouse Backup

ClickHouse stores trace and observation data. Due to volume, consider incremental or partition-based backups.

#### Using clickhouse-backup

```bash
# Install clickhouse-backup in ClickHouse container
# Configuration in /etc/clickhouse-backup/config.yml

# Create backup
clickhouse-backup create daily_$(date +%Y%m%d)

# List backups
clickhouse-backup list

# Upload to remote storage
clickhouse-backup upload daily_$(date +%Y%m%d)
```

#### clickhouse-backup Configuration

```yaml
# /etc/clickhouse-backup/config.yml
general:
  remote_storage: s3
  disable_progress_bar: true

clickhouse:
  host: localhost
  port: 9000
  username: agenttrace
  password: ${CLICKHOUSE_PASSWORD}
  timeout: 5m
  freeze_by_part: false

s3:
  access_key: ${AWS_ACCESS_KEY_ID}
  secret_key: ${AWS_SECRET_ACCESS_KEY}
  bucket: agenttrace-backups
  region: us-east-1
  path: clickhouse/
  compression_format: gzip
```

#### Manual Table Backup

```sql
-- Freeze table partitions (creates hard links)
ALTER TABLE traces FREEZE PARTITION tuple();
ALTER TABLE observations FREEZE PARTITION tuple();
ALTER TABLE scores FREEZE PARTITION tuple();

-- Backup files are in /var/lib/clickhouse/shadow/
```

#### Partition-Based Backup (Recommended for Large Datasets)

```bash
#!/bin/bash
# Backup only recent partitions (last 7 days)

BACKUP_DATE=$(date -d "7 days ago" +%Y%m%d)

clickhouse-client --query "
  ALTER TABLE traces FREEZE PARTITION '${BACKUP_DATE}'
"

# Copy frozen data to backup location
rsync -av /var/lib/clickhouse/shadow/ /backups/clickhouse/
```

### Redis Backup

Redis stores job queues and cache. For most deployments, Redis data is ephemeral.

```bash
# Trigger RDB snapshot
docker compose exec redis redis-cli -a ${REDIS_PASSWORD} BGSAVE

# Copy RDB file
docker compose cp redis:/data/dump.rdb ./redis_backup_$(date +%Y%m%d).rdb
```

### Full System Backup Checklist

| Component | Frequency | Retention | Priority |
|-----------|-----------|-----------|----------|
| PostgreSQL | Daily | 30 days | Critical |
| ClickHouse | Daily | 14 days | High |
| Redis RDB | Weekly | 7 days | Low |
| Secrets/Config | On change | Indefinite | Critical |
| TLS Certificates | On renewal | Previous version | Critical |

---

## Recovery Procedures

### PostgreSQL Recovery

#### Full Restore

```bash
# Stop API and worker to prevent writes
docker compose stop api worker

# Restore from backup
docker compose exec -T postgres pg_restore \
  -U agenttrace \
  -d agenttrace \
  --clean \
  --if-exists \
  < backup.dump

# Or from compressed backup
gunzip -c backup.dump.gz | docker compose exec -T postgres pg_restore \
  -U agenttrace \
  -d agenttrace \
  --clean \
  --if-exists

# Restart services
docker compose start api worker
```

#### Point-in-Time Recovery (if WAL archiving enabled)

```bash
# 1. Stop PostgreSQL
docker compose stop postgres

# 2. Remove current data
rm -rf /var/lib/postgresql/data/*

# 3. Restore base backup
pg_basebackup -D /var/lib/postgresql/data -Fp -Xs -P

# 4. Configure recovery
cat > /var/lib/postgresql/data/recovery.conf << EOF
restore_command = 'cp /backups/wal/%f %p'
recovery_target_time = '2024-01-15 10:30:00'
EOF

# 5. Start PostgreSQL (will replay WAL to target time)
docker compose start postgres
```

### ClickHouse Recovery

#### Full Restore

```bash
# 1. Stop services
docker compose stop api worker

# 2. Download backup
clickhouse-backup download daily_20240115

# 3. Restore backup
clickhouse-backup restore daily_20240115

# 4. Start services
docker compose start api worker
```

#### Restore Specific Tables

```bash
# Restore only traces table
clickhouse-backup restore daily_20240115 --table=traces

# Restore with different name (for comparison)
clickhouse-backup restore daily_20240115 --table=traces --restore-table=traces_restored
```

#### Manual Partition Restore

```sql
-- 1. Detach current partition
ALTER TABLE traces DETACH PARTITION '20240115';

-- 2. Copy backup files to detached directory
-- /var/lib/clickhouse/data/agenttrace/traces/detached/

-- 3. Attach restored partition
ALTER TABLE traces ATTACH PARTITION '20240115';
```

### Redis Recovery

```bash
# 1. Stop Redis
docker compose stop redis

# 2. Replace RDB file
docker compose cp redis_backup.rdb redis:/data/dump.rdb

# 3. Start Redis
docker compose start redis
```

---

## Disaster Scenarios

### Scenario 1: Complete Data Center Failure

**Impact:** All services unavailable, potential data loss

**Recovery Steps:**

1. **Assess damage** - Determine which components are affected
2. **Provision new infrastructure** - In alternate region/provider
3. **Restore secrets** - From secure backup (Vault, 1Password, etc.)
4. **Deploy infrastructure** - Using IaC (Terraform/Pulumi)
5. **Restore PostgreSQL** - Critical path, restore first
6. **Restore ClickHouse** - May take longer due to volume
7. **Deploy application** - API, Worker, Web
8. **Verify functionality** - Run health checks
9. **Update DNS** - Point to new infrastructure
10. **Monitor closely** - Watch for issues post-recovery

**RTO:** 4-8 hours (depending on data volume)
**RPO:** Last backup (typically < 24 hours)

### Scenario 2: PostgreSQL Corruption

**Symptoms:** API errors, authentication failures, missing projects

**Recovery Steps:**

1. **Stop writes** - Scale API/Worker to 0
   ```bash
   kubectl scale deployment agenttrace-api --replicas=0
   kubectl scale deployment agenttrace-worker --replicas=0
   ```

2. **Assess corruption**
   ```sql
   -- Check for corruption
   SELECT * FROM pg_catalog.pg_stat_user_tables;

   -- Run consistency check
   SELECT pg_catalog.pg_database_size('agenttrace');
   ```

3. **Attempt repair**
   ```bash
   # Reindex if index corruption
   docker compose exec postgres psql -U agenttrace -c "REINDEX DATABASE agenttrace;"
   ```

4. **If repair fails, restore from backup**
   ```bash
   pg_restore -U agenttrace -d agenttrace --clean < latest_backup.dump
   ```

5. **Restart services**
   ```bash
   kubectl scale deployment agenttrace-api --replicas=3
   kubectl scale deployment agenttrace-worker --replicas=2
   ```

### Scenario 3: ClickHouse Out of Disk Space

**Symptoms:** Ingestion failures, query timeouts

**Recovery Steps:**

1. **Check disk usage**
   ```sql
   SELECT
     database, table,
     formatReadableSize(sum(bytes)) as size
   FROM system.parts
   GROUP BY database, table
   ORDER BY sum(bytes) DESC;
   ```

2. **Drop old partitions** (if retention allows)
   ```sql
   -- Drop data older than 90 days
   ALTER TABLE traces DROP PARTITION WHERE toDate(timestamp) < today() - 90;
   ALTER TABLE observations DROP PARTITION WHERE toDate(timestamp) < today() - 90;
   ```

3. **Optimize tables** (merge small parts)
   ```sql
   OPTIMIZE TABLE traces FINAL;
   OPTIMIZE TABLE observations FINAL;
   ```

4. **Expand storage** (if needed)
   ```bash
   # Kubernetes - resize PVC
   kubectl patch pvc clickhouse-data -p '{"spec":{"resources":{"requests":{"storage":"500Gi"}}}}'
   ```

### Scenario 4: Redis Queue Backup

**Symptoms:** Jobs not processing, API slowdown

**Recovery Steps:**

1. **Check queue depth**
   ```bash
   redis-cli -a $REDIS_PASSWORD LLEN asynq:default:pending
   ```

2. **Scale workers**
   ```bash
   kubectl scale deployment agenttrace-worker --replicas=5
   ```

3. **If queue is corrupted, clear and requeue**
   ```bash
   # WARNING: This will lose pending jobs
   redis-cli -a $REDIS_PASSWORD FLUSHDB

   # Restart workers
   kubectl rollout restart deployment agenttrace-worker
   ```

### Scenario 5: API Key Compromise

**Symptoms:** Unexpected API activity, unauthorized access

**Recovery Steps:**

1. **Identify compromised key**
   ```sql
   SELECT id, name, last_used_at, created_at
   FROM api_keys
   WHERE project_id = 'affected_project_id';
   ```

2. **Revoke compromised key**
   ```sql
   UPDATE api_keys SET revoked_at = NOW() WHERE id = 'compromised_key_id';
   ```

3. **Generate new key** (via UI or API)

4. **Audit usage**
   ```sql
   SELECT * FROM audit_logs
   WHERE api_key_id = 'compromised_key_id'
   ORDER BY created_at DESC;
   ```

5. **Notify affected users**

---

## Testing Procedures

### Monthly DR Test Checklist

- [ ] Verify backup jobs are running
- [ ] Download and verify PostgreSQL backup integrity
- [ ] Test PostgreSQL restore to staging environment
- [ ] Verify ClickHouse backup completeness
- [ ] Test ClickHouse restore for one table
- [ ] Document any issues and update runbook

### Quarterly Full DR Test

1. **Provision isolated test environment**
2. **Restore all databases from backup**
3. **Deploy application**
4. **Run integration tests**
5. **Measure actual RTO**
6. **Document gaps and improve**

### Backup Verification Script

```bash
#!/bin/bash
# /opt/agenttrace/scripts/verify-backups.sh

set -euo pipefail

echo "=== Backup Verification Report ==="
echo "Date: $(date)"
echo

# Check PostgreSQL backup
LATEST_PG=$(ls -t /backups/postgres/*.dump.gz 2>/dev/null | head -1)
if [ -n "$LATEST_PG" ]; then
  PG_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_PG")) / 3600 ))
  PG_SIZE=$(du -h "$LATEST_PG" | cut -f1)
  echo "PostgreSQL: $LATEST_PG ($PG_SIZE, ${PG_AGE}h old)"

  # Verify backup integrity
  if gunzip -t "$LATEST_PG" 2>/dev/null; then
    echo "  Status: OK"
  else
    echo "  Status: CORRUPT"
    exit 1
  fi
else
  echo "PostgreSQL: NO BACKUP FOUND"
  exit 1
fi

# Check ClickHouse backup
if command -v clickhouse-backup &>/dev/null; then
  LATEST_CH=$(clickhouse-backup list 2>/dev/null | tail -1)
  echo "ClickHouse: $LATEST_CH"
else
  echo "ClickHouse: clickhouse-backup not installed"
fi

echo
echo "=== Verification Complete ==="
```

---

## Contacts and Escalation

### Escalation Matrix

| Severity | Response Time | Escalation Path |
|----------|---------------|-----------------|
| SEV1 (Outage) | 15 min | On-call -> Engineering Lead -> CTO |
| SEV2 (Degraded) | 1 hour | On-call -> Engineering Lead |
| SEV3 (Minor) | 4 hours | On-call |

### Communication Templates

#### Incident Start

```
[INCIDENT] AgentTrace - {severity} - {description}

Impact: {user impact description}
Started: {timestamp}
Status: Investigating

Current actions:
- {action 1}
- {action 2}

Next update in 30 minutes.
```

#### Incident Resolved

```
[RESOLVED] AgentTrace - {severity} - {description}

Impact: {user impact description}
Duration: {start} - {end} ({duration})
Root cause: {brief description}

Full post-mortem to follow within 48 hours.
```

### External Dependencies

| Service | Support Contact | SLA |
|---------|-----------------|-----|
| AWS | aws.amazon.com/support | Per contract |
| ClickHouse Cloud | support@clickhouse.com | Per contract |
| Cloudflare | cloudflare.com/support | Per plan |

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2024-01-15 | 1.0 | Initial runbook |
