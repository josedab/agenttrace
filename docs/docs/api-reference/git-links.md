---
sidebar_position: 10
title: Git Links API
description: Endpoints for linking git commits and repositories to traces for code-trace correlation.
---

# Git Links API

Git links associate traces with git commits, branches, and repositories, enabling correlation between agent actions and code changes.

## Create Git Link

```
POST /api/public/git-links
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `traceId` | string | Yes | Associated trace ID |
| `commitSha` | string | Yes | Full git commit SHA |
| `repository` | string | Yes | Repository identifier (e.g., `owner/repo`) |
| `branch` | string | No | Branch name |
| `commitMessage` | string | No | Commit message |
| `authorName` | string | No | Commit author |
| `metadata` | object | No | Custom metadata |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/git-links" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc123",
    "commitSha": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
    "repository": "acme/my-app",
    "branch": "feature/auto-refactor",
    "commitMessage": "refactor: apply agent suggestions",
    "authorName": "AI Agent"
  }'
```

### Response

```json
{
  "id": "gl-abc123",
  "traceId": "trace-abc123",
  "commitSha": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
  "repository": "acme/my-app",
  "branch": "feature/auto-refactor",
  "commitMessage": "refactor: apply agent suggestions",
  "authorName": "AI Agent",
  "metadata": {},
  "createdAt": "2024-01-15T10:30:00Z"
}
```

## List Git Links

```
GET /api/public/git-links
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |
| `traceId` | string | – | Filter by trace |
| `repository` | string | – | Filter by repository |
| `commitSha` | string | – | Filter by commit SHA |
| `branch` | string | – | Filter by branch |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/git-links?repository=acme/my-app&branch=main" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "data": [
    {
      "id": "gl-abc123",
      "traceId": "trace-abc123",
      "commitSha": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
      "repository": "acme/my-app",
      "branch": "feature/auto-refactor",
      "commitMessage": "refactor: apply agent suggestions",
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": { "totalCount": 8, "hasMore": false, "nextCursor": null }
}
```

## Delete Git Link

```
DELETE /api/public/git-links/:gitLinkId
```

### Response

```json
{ "success": true }
```
