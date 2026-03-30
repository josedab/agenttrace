---
sidebar_position: 7
title: Deployment Issues
---

# Deployment Issues

## ClickHouse Connection Failed

**Symptom**: API logs show ClickHouse connection errors

**Solutions**:

1. **Check ClickHouse health**:
   ```bash
   curl http://localhost:8123/ping
   # Should return "Ok."
   ```

2. **Verify credentials**:
   ```bash
   # Check environment variables
   echo $CLICKHOUSE_USER
   echo $CLICKHOUSE_PASSWORD
   ```

3. **Check network connectivity**:
   ```bash
   docker compose exec api ping clickhouse
   ```

## ClickHouse Authentication Errors

**Symptom**: `Authentication failed` or `Code: 516. DB::Exception: default: Authentication failed`

**Solutions**:

1. **Verify user and password match between API and ClickHouse**:
   ```bash
   # Check what the API is configured to use
   grep CLICKHOUSE deploy/.env

   # Test connection directly
   curl "http://localhost:8123/?user=agenttrace&password=agenttrace" \
     --data "SELECT 1"
   # Should return: 1
   ```

2. **Check if the user exists in ClickHouse**:
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "SHOW USERS"
   ```

3. **Recreate the user if needed** (development only):
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "CREATE USER IF NOT EXISTS agenttrace IDENTIFIED BY 'agenttrace'"
   docker compose exec clickhouse clickhouse-client \
     --query "GRANT ALL ON agenttrace.* TO agenttrace"
   ```

4. **Ensure `CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT` is enabled** in docker-compose:
   ```yaml
   clickhouse:
     environment:
       CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT: 1  # Required for SQL-based user management
   ```

## ClickHouse Memory Issues

**Symptom**: `Memory limit (total) exceeded` or ClickHouse OOM kills

**Solutions**:

1. **Check current memory usage**:
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "SELECT formatReadableSize(memory_usage) FROM system.processes"
   ```

2. **Increase memory limits in docker-compose**:
   ```yaml
   clickhouse:
     deploy:
       resources:
         limits:
           memory: 4G   # Increase from default
     environment:
       # Limit per-query memory (prevents single query from using all RAM)
       CLICKHOUSE_MAX_MEMORY_USAGE: "2000000000"  # 2GB per query
   ```

3. **Optimize large queries**:
   ```bash
   # Check which queries are consuming the most memory
   docker compose exec clickhouse clickhouse-client \
     --query "SELECT query, memory_usage FROM system.processes ORDER BY memory_usage DESC LIMIT 5"
   ```

4. **For development environments with limited RAM** (e.g., < 4GB):
   ```bash
   # Add to ClickHouse config or environment
   CLICKHOUSE_MAX_SERVER_MEMORY_USAGE_RATIO: "0.5"  # Use at most 50% of available RAM
   ```

## ClickHouse Database Not Found

**Symptom**: `Code: 81. DB::Exception: Database agenttrace doesn't exist`

**Solutions**:

1. **Create the database**:
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "CREATE DATABASE IF NOT EXISTS agenttrace"
   ```

2. **Run migrations**:
   ```bash
   make migrate-ch-up
   ```

3. **Verify the database exists**:
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "SHOW DATABASES"
   ```

## ClickHouse Slow Queries

**Symptom**: Trace queries are slow, dashboard takes a long time to load

**Solutions**:

1. **Check the query log for slow queries**:
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "SELECT query, query_duration_ms FROM system.query_log
              WHERE query_duration_ms > 1000
              ORDER BY query_duration_ms DESC LIMIT 10"
   ```

2. **Verify table engines and partitioning** are correct (migrations should set this up):
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "SELECT name, engine, partition_key FROM system.tables
              WHERE database = 'agenttrace'"
   ```

3. **Force optimize tables** if merge operations are lagging:
   ```bash
   docker compose exec clickhouse clickhouse-client \
     --query "OPTIMIZE TABLE agenttrace.traces FINAL"
   ```

## PostgreSQL Connection Issues

**Symptom**: "connection refused" to PostgreSQL

**Solutions**:

1. **Wait for PostgreSQL to be ready**:
   ```bash
   # PostgreSQL takes a few seconds to start
   docker compose logs postgres
   # Look for "database system is ready to accept connections"
   ```

2. **Check credentials match**:
   ```yaml
   # In docker-compose.yml
   postgres:
     environment:
       POSTGRES_USER: agenttrace
       POSTGRES_PASSWORD: agenttrace  # Must match API config
   ```

## Migrations Not Applied

**Symptom**: API errors about missing tables

**Solution**:
```bash
# Run migrations manually
cd api
make migrate-pg-up
make migrate-ch-up
```
