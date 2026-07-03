---
sidebar_position: 1
title: Authentication
description: API key and JWT authentication methods for AgentTrace API access.
---

# Authentication

AgentTrace supports **API keys** for programmatic access and **JWT access tokens** for dashboard sessions. Credentials are accepted only in headers, never in query strings.

## API keys

The secret credential starts with `sk-at-` and is returned only when a key is created. The corresponding `pk-at-` value is a public identifier and cannot authenticate by itself.

### Header authentication

```bash
curl "https://api.agenttrace.io/api/public/traces" \
  -H "X-API-Key: sk-at-key-id.secret"
```

Bearer authentication is also supported:

```bash
curl "https://api.agenttrace.io/api/public/traces" \
  -H "Authorization: Bearer sk-at-key-id.secret"
```

### Basic authentication

Use the public identifier as the username and secret credential as the password:

```bash
curl "https://api.agenttrace.io/api/public/traces" \
  -u "pk-at-key-id:sk-at-key-id.secret"
```

### Creating API keys

```text
POST /api/v1/projects/:projectId/api-keys
```

This management endpoint requires a JWT access token and access to the project.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Descriptive name for the key |
| `scopes` | string[] | No | Permission scopes; defaults to the standard SDK scopes |
| `expiresAt` | string | No | ISO 8601 expiration; defaults to one year |

```bash
curl -X POST "https://api.agenttrace.io/api/v1/projects/$PROJECT_ID/api-keys" \
  -H "Authorization: Bearer $JWT_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "python-sdk-prod",
    "scopes": ["traces:read", "traces:write", "scores:write"],
    "expiresAt": "2027-12-31T23:59:59Z"
  }'
```

```json
{
  "id": "key-abc123",
  "name": "python-sdk-prod",
  "publicKey": "pk-at-0123456789abcdef0123456789abcdef",
  "secretKey": "sk-at-0123456789abcdef0123456789abcdef.0123456789abcdef",
  "secretKeyPreview": "cdef",
  "scopes": ["traces:read", "traces:write", "scores:write"],
  "expiresAt": "2027-12-31T23:59:59Z",
  "createdAt": "2026-01-15T10:00:00Z"
}
```

Store `secretKey` securely; it is not returned again.

### Key ownership and rotation

Every API key records the user that created it (`createdBy`). AgentTrace uses this
owner as the acting user for endpoints that are constrained by a user foreign key
— Eval Hub publish/fork/run, replay sessions and annotations, and prompt/migration
imports. When a request authenticates with an **owned** key, it is attributed to
that owner exactly as if the owner had made the request with a JWT.

Keys created before ownership tracking existed have no `createdBy` (**legacy
keys**). Such keys can still call ingestion and other project-scoped, non-user
endpoints, but calls to user-attributed endpoints are rejected with an explicit
`401 Unauthorized` ("Actor ID not found") instead of failing later with an
opaque error. Share links are exempt because they are not user-foreign-key
constrained; a legacy key is attributed by its key identity there.

To rotate: create a new key (which captures the current creator as owner),
migrate clients to it, then revoke the old key. Rotating a legacy key by
recreating it is the supported path to enable user attribution for machine
callers.

## JWT access tokens

Credentials and configured OAuth providers use the same backend session flow. Password login is available at:

```text
POST /api/auth/login
```

```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

The response contains `accessToken`, `refreshToken`, `expiresAt`, and `user`. Send the access token as a bearer token:

```bash
curl "https://api.agenttrace.io/api/v1/projects" \
  -H "Authorization: Bearer $JWT_ACCESS_TOKEN"
```

Access-token lifetime, refresh-token lifetime, and issuer are configured with `JWT_ACCESS_EXPIRY_MINUTES`, `JWT_REFRESH_EXPIRY_DAYS`, and `JWT_ISSUER`.

## Core scopes

| Scope | Description |
|---|---|
| `traces:read` | Read traces, sessions, checkpoints, and related trace data |
| `traces:write` | Ingest and update traces and related trace data |
| `traces:delete` | Delete traces |
| `observations:read` | Read observations |
| `observations:write` | Create and update observations |
| `scores:read` | Read scores |
| `scores:write` | Create and update scores |
| `scores:delete` | Delete scores |
| `prompts:read` | Read prompts |
| `prompts:write` | Create and update prompts |
| `prompts:delete` | Delete prompts |
| `datasets:read` | Read datasets |
| `datasets:write` | Manage datasets |
| `datasets:delete` | Delete datasets |
| `evaluators:read` | Read evaluators |
| `evaluators:write` | Manage evaluators |
| `evaluators:delete` | Delete evaluators |
| `admin:write` | Permit all scoped API-key operations |
