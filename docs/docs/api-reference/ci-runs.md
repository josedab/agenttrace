---
sidebar_position: 11
title: CI Runs API
description: Endpoints for tracking CI/CD pipeline runs and linking them to traces.
---

# CI Runs API

CI runs track continuous integration pipeline executions and link them to AgentTrace traces for end-to-end observability.

## Create CI Run

```
POST /api/public/ci-runs
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | CI run name or pipeline identifier |
| `traceId` | string | No | Associated trace ID |
| `provider` | string | Yes | CI provider (`github_actions`, `gitlab_ci`, `jenkins`, `circleci`) |
| `runId` | string | Yes | External run identifier from the CI provider |
| `repository` | string | Yes | Repository identifier (e.g., `owner/repo`) |
| `branch` | string | No | Branch name |
| `commitSha` | string | No | Commit SHA that triggered the run |
| `status` | string | Yes | `pending`, `running`, `success`, `failure`, `cancelled` |
| `url` | string | No | URL to the CI run |
| `metadata` | object | No | Custom metadata |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/ci-runs" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "eval-pipeline",
    "traceId": "trace-abc123",
    "provider": "github_actions",
    "runId": "7890123456",
    "repository": "acme/my-app",
    "branch": "main",
    "commitSha": "a1b2c3d4e5f6",
    "status": "running",
    "url": "https://github.com/acme/my-app/actions/runs/7890123456"
  }'
```

### Response

```json
{
  "id": "ci-abc123",
  "name": "eval-pipeline",
  "traceId": "trace-abc123",
  "provider": "github_actions",
  "runId": "7890123456",
  "repository": "acme/my-app",
  "branch": "main",
  "commitSha": "a1b2c3d4e5f6",
  "status": "running",
  "url": "https://github.com/acme/my-app/actions/runs/7890123456",
  "metadata": {},
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

## Update CI Run

```
PATCH /api/public/ci-runs/:ciRunId
```

### Request Body

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Updated status |
| `metadata` | object | Merged with existing metadata |

### Example

```bash
curl -X PATCH "https://api.agenttrace.io/api/public/ci-runs/ci-abc123" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "status": "success",
    "metadata": {"duration_ms": 45200, "tests_passed": 142}
  }'
```

## List CI Runs

```
GET /api/public/ci-runs
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |
| `traceId` | string | – | Filter by trace |
| `repository` | string | – | Filter by repository |
| `status` | string | – | Filter by status |
| `provider` | string | – | Filter by CI provider |

### Response

```json
{
  "data": [
    {
      "id": "ci-abc123",
      "name": "eval-pipeline",
      "provider": "github_actions",
      "repository": "acme/my-app",
      "status": "success",
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": { "totalCount": 15, "hasMore": false, "nextCursor": null }
}
```
