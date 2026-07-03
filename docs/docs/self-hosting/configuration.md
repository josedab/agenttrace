---
sidebar_position: 4
title: "Configuration Reference"
description: "Production environment variables for AgentTrace."
---

# Configuration Reference

The canonical copy-ready configuration is [`deploy/.env.example`](https://github.com/agenttrace/agenttrace/blob/main/deploy/.env.example).

## Required Deployment Values

| Variable | Purpose |
|----------|---------|
| `VERSION` | Immutable AgentTrace API and web image tag |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `CLICKHOUSE_PASSWORD` | ClickHouse password |
| `REDIS_PASSWORD` | Redis password |
| `JWT_SECRET` | API JWT signing secret |
| `NEXTAUTH_SECRET` | Web session secret |
| `NEXTAUTH_URL` | Public web URL |
| `NEXT_PUBLIC_API_URL` | Public API URL embedded into the web image |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins allowed by the API |

Placeholder values beginning with `change-me` are rejected in production.

The web container also uses `API_INTERNAL_URL` for server-side authentication calls and `AUTH_TRUST_HOST=true` when it runs behind the configured proxy/ingress.

## Server and Browser Security

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` / `API_HOST` | `0.0.0.0` | API bind address |
| `SERVER_PORT` / `API_PORT` | `8080` | API port |
| `SERVER_ENV` / `ENVIRONMENT` | `development` | Set to `production` in deployments |
| `SERVER_CSRF_ENABLED` | `true` | Enable CSRF middleware |
| `SERVER_SECURE_COOKIES` | `false` | Require secure cookies; production Compose sets `true` |
| `CORS_ALLOWED_ORIGINS` | `*` | Explicit origins are required in production |
| `CORS_ALLOW_CREDENTIALS` | `false` | Allow browser credentials for approved origins |
| `PUBLIC_URL` | — | Public web origin used when returning share links and dashboard links |

## Privacy and No-Egress

| Variable | Default | Description |
|----------|---------|-------------|
| `PRIVACY_NO_EGRESS` | `false` | Block outbound integrations at runtime |
| `PRIVACY_REDACTION_ENABLED` | `true` | Enable deterministic redaction for public/shareable data |

See [Local and Private Mode](/self-hosting/privacy-mode) for conflict validation and effective capabilities.

## Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | — | JWT signing key |
| `JWT_ACCESS_EXPIRY_MINUTES` | `1440` | Access-token lifetime |
| `JWT_REFRESH_EXPIRY_DAYS` | `7` | Refresh-token lifetime |
| `JWT_ISSUER` | `agenttrace` | Expected token issuer |

## Database Configuration

Component variables and full connection URLs are both supported. Full URLs override component values.

| Service | Component variables | Full URL |
|---------|---------------------|----------|
| PostgreSQL | `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_SSL_MODE` | `DATABASE_URL` |
| ClickHouse | `CLICKHOUSE_HOST`, `CLICKHOUSE_PORT`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DB` | `CLICKHOUSE_DSN` |
| Redis | `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` | `REDIS_URL` |

`CLICKHOUSE_DB` must remain `agenttrace`; the packaged ClickHouse migrations target that canonical database.

For an internal database without TLS, set both:

```bash
POSTGRES_SSL_MODE=disable
POSTGRES_ALLOW_INSECURE=true
```

Managed production databases should use `require`, `verify-ca`, or `verify-full`.

## Object Storage

| Variable | Default |
|----------|---------|
| `MINIO_ENABLED` | `true` |
| `MINIO_ENDPOINT` | `localhost:9002` |
| `MINIO_ACCESS_KEY` | `agenttrace` |
| `MINIO_SECRET_KEY` | — |
| `MINIO_USE_SSL` | `false` |
| `MINIO_BUCKET` | `agenttrace-exports` |

## Optional Integrations

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` / `EVAL_API_KEY` | LLM evaluations |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | Google OAuth |
| `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` | GitHub OAuth |
| `GITHUB_REPORTING_ENABLED` | Enable optional outcome report delivery |
| `GITHUB_API_URL` | GitHub REST API base URL |
| `GITHUB_REPORT_TOKEN` | Token used only for optional report delivery |
| `OAUTH_CALLBACK_SECRET` | Shared secret authenticating Auth.js callbacks to the API; required with OAuth |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook verification |
| `SENTRY_ENABLED`, `SENTRY_DSN` | Error reporting |
| `SENTRY_ENVIRONMENT`, `SENTRY_RELEASE` | Sentry release context |
| `SENTRY_SAMPLE_RATE`, `SENTRY_TRACES_SAMPLE_RATE` | Sentry sampling |

## Observability

| Variable | Default |
|----------|---------|
| `LOG_LEVEL` | `info` |
| `LOG_FORMAT` | `json` |
| `OTEL_RECEIVER_ENABLED` | `false` |
| `OTEL_RECEIVER_GRPC_PORT` | `4317` |

Prometheus metrics are available from the API at `/metrics`.
