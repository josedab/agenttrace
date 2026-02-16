---
sidebar_position: 9
title: Checkpoints API
description: Endpoints for creating and managing code state checkpoints linked to traces.
---

# Checkpoints API

Checkpoints capture snapshots of code state at specific points during agent execution, enabling replay and debugging.

## Create Checkpoint

```
POST /api/public/checkpoints
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `traceId` | string | Yes | Associated trace ID |
| `observationId` | string | No | Associated observation ID |
| `name` | string | Yes | Checkpoint name |
| `description` | string | No | What this checkpoint captures |
| `snapshot` | object | Yes | Code state snapshot |
| `metadata` | object | No | Custom metadata |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/checkpoints" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc123",
    "name": "pre-refactor",
    "description": "State before applying refactoring suggestions",
    "snapshot": {
      "files": {
        "src/main.py": "def hello(): print(\"world\")",
        "src/utils.py": "def add(a, b): return a + b"
      },
      "commitSha": "a1b2c3d"
    },
    "metadata": {"step": 3}
  }'
```

### Response

```json
{
  "id": "cp-abc123",
  "traceId": "trace-abc123",
  "observationId": null,
  "name": "pre-refactor",
  "description": "State before applying refactoring suggestions",
  "snapshot": {
    "files": { "src/main.py": "...", "src/utils.py": "..." },
    "commitSha": "a1b2c3d"
  },
  "metadata": {"step": 3},
  "createdAt": "2024-01-15T10:30:00Z"
}
```

## List Checkpoints

```
GET /api/public/checkpoints
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |
| `traceId` | string | – | Filter by trace |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/checkpoints?traceId=trace-abc123" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "data": [
    {
      "id": "cp-abc123",
      "traceId": "trace-abc123",
      "name": "pre-refactor",
      "description": "State before applying refactoring suggestions",
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": { "totalCount": 2, "hasMore": false, "nextCursor": null }
}
```

## Get Checkpoint

```
GET /api/public/checkpoints/:checkpointId
```

Returns the full checkpoint including the snapshot payload.

## Delete Checkpoint

```
DELETE /api/public/checkpoints/:checkpointId
```

### Response

```json
{ "success": true }
```
