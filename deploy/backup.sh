#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="${1:-$SCRIPT_DIR/backups/$timestamp}"
mkdir -p "$backup_dir"

postgres_user="${POSTGRES_USER:-agenttrace}"
postgres_db="${POSTGRES_DB:-agenttrace}"
clickhouse_user="${CLICKHOUSE_USER:-agenttrace}"
clickhouse_password="${CLICKHOUSE_PASSWORD:?CLICKHOUSE_PASSWORD required}"
clickhouse_db="${CLICKHOUSE_DB:-agenttrace}"

[[ "$clickhouse_db" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || {
  echo "invalid ClickHouse database name: $clickhouse_db" >&2
  exit 1
}

writers_stopped=true
restart_writers() {
  if [[ "$writers_stopped" == "true" ]]; then
    docker compose up -d api worker >/dev/null
  fi
}
trap restart_writers EXIT

echo "Pausing application writers..."
docker compose stop api worker

echo "Backing up PostgreSQL..."
docker compose exec -T postgres \
  pg_dump -U "$postgres_user" -Fc "$postgres_db" > "$backup_dir/postgres.dump"

clickhouse_file="clickhouse-$timestamp.zip"
echo "Backing up ClickHouse..."
docker compose exec -T clickhouse clickhouse-client \
  --user "$clickhouse_user" \
  --password "$clickhouse_password" \
  --database "$clickhouse_db" \
  --query "BACKUP DATABASE $clickhouse_db TO Disk('backups', '$clickhouse_file')"
docker compose cp "clickhouse:/backups/$clickhouse_file" "$backup_dir/$clickhouse_file"

cat > "$backup_dir/manifest.txt" <<EOF
created_at=$timestamp
postgres_database=$postgres_db
clickhouse_database=$clickhouse_db
clickhouse_file=$clickhouse_file
EOF

if [[ "${MINIO_ENABLED:-false}" == "true" ]]; then
  echo "External object storage is enabled; back it up using the provider's replication or snapshot tooling."
fi

restart_writers
writers_stopped=false
trap - EXIT

echo "Backup complete: $backup_dir"
