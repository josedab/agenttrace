#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "--yes" || -z "${2:-}" ]]; then
  echo "usage: $0 --yes <backup-directory>" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
backup_dir="$(cd "$2" && pwd)"
cd "$SCRIPT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

[[ -f "$backup_dir/postgres.dump" ]] || { echo "missing postgres.dump" >&2; exit 1; }
[[ -f "$backup_dir/manifest.txt" ]] || { echo "missing manifest.txt" >&2; exit 1; }

manifest_value() {
  local key="$1"
  awk -F= -v key="$key" '$1 == key { sub(/^[^=]*=/, ""); print; exit }' \
    "$backup_dir/manifest.txt"
}

clickhouse_file="$(manifest_value clickhouse_file)"
clickhouse_database="$(manifest_value clickhouse_database)"

[[ -n "$clickhouse_file" ]] || { echo "manifest missing clickhouse_file" >&2; exit 1; }
[[ -n "$clickhouse_database" ]] || { echo "manifest missing clickhouse_database" >&2; exit 1; }

clickhouse_backup="$backup_dir/$clickhouse_file"
[[ -f "$clickhouse_backup" ]] || { echo "missing $clickhouse_file" >&2; exit 1; }
[[ "$clickhouse_database" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || {
  echo "invalid ClickHouse database name: $clickhouse_database" >&2
  exit 1
}
[[ "$clickhouse_file" == "$(basename "$clickhouse_file")" ]] || {
  echo "invalid ClickHouse backup filename" >&2
  exit 1
}

postgres_user="${POSTGRES_USER:-agenttrace}"
postgres_db="${POSTGRES_DB:-agenttrace}"
clickhouse_user="${CLICKHOUSE_USER:-agenttrace}"
clickhouse_password="${CLICKHOUSE_PASSWORD:?CLICKHOUSE_PASSWORD required}"
redis_password="${REDIS_PASSWORD:?REDIS_PASSWORD required}"

[[ "$postgres_user" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || {
  echo "invalid PostgreSQL user name: $postgres_user" >&2
  exit 1
}
[[ "$postgres_db" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || {
  echo "invalid PostgreSQL database name: $postgres_db" >&2
  exit 1
}

echo "Stopping application writers..."
docker compose stop api worker web

echo "Restoring PostgreSQL..."
docker compose exec -T postgres psql -U "$postgres_user" -d postgres \
  -v ON_ERROR_STOP=1 \
  -c "DROP DATABASE IF EXISTS \"$postgres_db\" WITH (FORCE)" \
  -c "CREATE DATABASE \"$postgres_db\" OWNER \"$postgres_user\""
docker compose exec -T postgres \
  pg_restore -U "$postgres_user" -d "$postgres_db" --exit-on-error --no-owner \
  < "$backup_dir/postgres.dump"

echo "Restoring ClickHouse..."
docker compose cp "$clickhouse_backup" "clickhouse:/backups/$clickhouse_file"
docker compose exec -T -u 0 clickhouse chown 101:101 "/backups/$clickhouse_file"
docker compose exec -T clickhouse clickhouse-client \
  --user "$clickhouse_user" \
  --password "$clickhouse_password" \
  --multiquery \
  --query "DROP DATABASE IF EXISTS $clickhouse_database; RESTORE DATABASE $clickhouse_database FROM Disk('backups', '$clickhouse_file')"

echo "Clearing newer Redis queue and rate-limit state..."
docker compose exec -T -e REDISCLI_AUTH="$redis_password" redis redis-cli FLUSHDB >/dev/null

if [[ "${MINIO_ENABLED:-false}" == "true" ]]; then
  echo "External object storage is enabled; restore it with the provider's tooling before restarting writers."
fi

echo "Restarting application..."
docker compose up -d api worker web
echo "Restore complete."
