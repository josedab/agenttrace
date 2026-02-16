---
sidebar_position: 1
title: Authentication
description: API key and JWT authentication methods for AgentTrace API access.
---

# Authentication

AgentTrace supports two authentication methods: **API Keys** for programmatic access and **JWT tokens** for dashboard sessions.

## API Keys

API keys are used by SDKs and direct API calls. Pass the key via the `X-API-Key` header or as Basic Auth credentials.

### Header Authentication

```bash
curl -X GET "https://api.agenttrace.io/api/public/traces" \
  -H "X-API-Key: sk-at-your-api-key"
```

### Basic Authentication

Use your public key as username and secret key as password:

```bash
curl -X GET "https://api.agenttrace.io/api/public/traces" \
  -u "pk-at-public-key:sk-at-secret-key"
```

### Creating API Keys

```
POST /api/public/api-keys
```

#### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Descriptive name for the key |
| `scopes` | string[] | No | Permission scopes (default: all) |
| `expiresAt` | string | No | Expiration date (ISO 8601) |

#### Example

```bash
curl -X POST "https://api.agenttrace.io/api/public/api-keys" \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "python-sdk-prod",
    "scopes": ["traces:read", "traces:write", "scores:write"],
    "expiresAt": "2025-12-31T23:59:59Z"
  }'
```

#### Response

```json
{
  "id": "key-abc123",
  "name": "python-sdk-prod",
  "publicKey": "pk-at-xxxxxxxxxxxx",
  "secretKey": "sk-at-xxxxxxxxxxxx",
  "scopes": ["traces:read", "traces:write", "scores:write"],
  "expiresAt": "2025-12-31T23:59:59Z",
  "createdAt": "2024-01-15T10:00:00Z"
}
```

> **Note:** The `secretKey` is only returned once at creation time. Store it securely.

## JWT Tokens

JWT tokens are used by the AgentTrace dashboard and web UI. Obtain a token via the login endpoint.

```
POST /api/public/auth/login
```

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

#### Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 86400,
  "user": {
    "id": "user-123",
    "email": "user@example.com",
    "role": "admin"
  }
}
```

### Using the JWT Token

Pass the token in the `Authorization` header with the `Bearer` scheme:

```bash
curl -X GET "https://api.agenttrace.io/api/public/projects" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

## Scopes Reference

| Scope | Description |
|-------|-------------|
| `traces:read` | Read traces and observations |
| `traces:write` | Create and update traces |
| `scores:read` | Read scores |
| `scores:write` | Create scores |
| `prompts:read` | Read prompts |
| `prompts:write` | Create and update prompts |
| `datasets:read` | Read datasets |
| `datasets:write` | Manage datasets |
