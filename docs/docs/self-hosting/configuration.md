---
sidebar_position: 4
title: "Configuration Reference"
description: "All AgentTrace environment variables with descriptions and default values."
---

# Configuration Reference

This page documents all environment variables used to configure AgentTrace. These are defined in `deploy/.env.example` and can be set in your `.env` file (Docker Compose) or ConfigMap/Secrets (Kubernetes).

## Database Credentials (Required)

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_USER` | PostgreSQL username | `agenttrace` |
| `POSTGRES_PASSWORD` | PostgreSQL password | — (required) |
| `POSTGRES_DB` | PostgreSQL database name | `agenttrace` |
| `CLICKHOUSE_USER` | ClickHouse username | `default` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | — (required) |
| `CLICKHOUSE_DB` | ClickHouse database name | `agenttrace` |
| `REDIS_PASSWORD` | Redis password | — (required) |
| `MINIO_ROOT_USER` | MinIO root username | `agenttrace` |
| `MINIO_ROOT_PASSWORD` | MinIO root password | — (required) |

## Security Keys (Required)

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | Secret key for signing JWT tokens. Generate with `openssl rand -base64 32`. | — (required) |
| `ENCRYPTION_KEY` | Key for encrypting sensitive data at rest (API keys, credentials). Generate with `openssl rand -base64 32`. | — (required) |

## NextAuth Configuration (Required)

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXTAUTH_URL` | Public-facing URL of your AgentTrace instance (e.g., `https://agenttrace.your-company.com`). | — (required) |
| `NEXTAUTH_SECRET` | Secret for NextAuth session encryption. Generate with `openssl rand -base64 32`. | — (required) |

## OAuth Providers (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_CLIENT_ID` | Google OAuth client ID for Google Sign-In. | — |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret. | — |
| `GITHUB_CLIENT_ID` | GitHub OAuth app client ID for GitHub Sign-In. | — |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth app client secret. | — |

## API Configuration (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `API_HOST` | Host address the API server binds to. | `0.0.0.0` |
| `API_PORT` | Port the API server listens on. | `8080` |
| `NEXT_PUBLIC_API_URL` | Internal URL for the web frontend to reach the API. In Docker Compose, use the service name (e.g., `http://api:8080`). | `http://api:8080` |
| `LOG_LEVEL` | Log verbosity. Options: `debug`, `info`, `warn`, `error`. | `info` |

## External Services (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key. Required for LLM-as-Judge evaluators and the prompt playground. | — |

## Deployment (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `VERSION` | Docker image tag to use. | `latest` |
| `WEB_PORT` | Port to expose the web dashboard on the host. | `3000` |

## Connection URLs

These are typically constructed from the individual credential variables above and configured in the Docker Compose file. Override them if using external managed services.

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | Full PostgreSQL connection string. | `postgres://agenttrace:<password>@postgres:5432/agenttrace` |
| `CLICKHOUSE_URL` | Full ClickHouse connection string. | `clickhouse://default:<password>@clickhouse:9000/agenttrace` |
| `REDIS_URL` | Full Redis connection string. | `redis://:<password>@redis:6379` |
| `MINIO_ENDPOINT` | MinIO or S3-compatible endpoint. | `minio:9000` |

## Using Managed Services

You can replace any infrastructure component with a managed service by overriding the connection URL:

```bash
# Use AWS RDS for PostgreSQL
DATABASE_URL=postgres://user:pass@your-rds-instance.amazonaws.com:5432/agenttrace

# Use ClickHouse Cloud
CLICKHOUSE_URL=clickhouse://default:pass@your-instance.clickhouse.cloud:9440/agenttrace?secure=true

# Use AWS ElastiCache for Redis
REDIS_URL=redis://your-elasticache.amazonaws.com:6379

# Use AWS S3 instead of MinIO
MINIO_ENDPOINT=s3.amazonaws.com
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
S3_BUCKET=your-agenttrace-bucket
S3_REGION=us-east-1
```

## SSO Configuration (Enterprise)

| Variable | Description | Default |
|----------|-------------|---------|
| `SSO_ENABLED` | Enable SSO authentication. | `false` |
| `AGENTTRACE_SSO_DEBUG` | Enable debug logging for SSO. | `false` |

See [Single Sign-On](../enterprise/sso.md) for provider-specific configuration.

## Generating Secrets

Generate all required secrets at once:

```bash
echo "JWT_SECRET=$(openssl rand -base64 32)"
echo "ENCRYPTION_KEY=$(openssl rand -base64 32)"
echo "NEXTAUTH_SECRET=$(openssl rand -base64 32)"
echo "POSTGRES_PASSWORD=$(openssl rand -base64 24)"
echo "CLICKHOUSE_PASSWORD=$(openssl rand -base64 24)"
echo "REDIS_PASSWORD=$(openssl rand -base64 24)"
echo "MINIO_ROOT_PASSWORD=$(openssl rand -base64 24)"
```

## Related

- [Docker Compose Deployment](./docker-compose.md) — using these variables with Docker Compose
- [Kubernetes Deployment](./kubernetes.md) — using these variables in Kubernetes
- [Self-Hosting Overview](./overview.md) — architecture and requirements
