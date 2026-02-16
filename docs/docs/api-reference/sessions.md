---
sidebar_position: 4
title: Sessions API
description: Endpoints for managing sessions that group related traces together.
---

# Sessions API

Sessions group related traces into a logical unit, such as a user conversation or a multi-step agent workflow.

## List Sessions

```
GET /api/public/sessions
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |
| `fromTimestamp` | string | – | Start time (ISO 8601) |
| `toTimestamp` | string | – | End time (ISO 8601) |
| `orderBy` | string | createdAt | Sort field |
| `order` | string | desc | Sort direction (`asc`/`desc`) |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/sessions?limit=20&order=desc" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "data": [
    {
      "id": "session-abc123",
      "projectId": "project-xyz",
      "traceCount": 5,
      "totalCost": 0.0145,
      "totalLatency": 12500,
      "firstTraceAt": "2024-01-15T10:00:00Z",
      "lastTraceAt": "2024-01-15T10:15:00Z",
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "totalCount": 87,
    "hasMore": true,
    "nextCursor": "eyJpZCI6InNlc3Npb24tYWJjMTIyIn0="
  }
}
```

## Get Session

```
GET /api/public/sessions/:sessionId
```

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `sessionId` | string | The session ID |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/sessions/session-abc123" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "id": "session-abc123",
  "projectId": "project-xyz",
  "traceCount": 5,
  "totalCost": 0.0145,
  "totalLatency": 12500,
  "firstTraceAt": "2024-01-15T10:00:00Z",
  "lastTraceAt": "2024-01-15T10:15:00Z",
  "createdAt": "2024-01-15T10:00:00Z",
  "traces": [
    {
      "id": "trace-001",
      "name": "user-query",
      "latency": 2500,
      "totalCost": 0.0029,
      "createdAt": "2024-01-15T10:00:00Z"
    },
    {
      "id": "trace-002",
      "name": "tool-execution",
      "latency": 4000,
      "totalCost": 0.0048,
      "createdAt": "2024-01-15T10:05:00Z"
    }
  ]
}
```

## Session Grouping

Sessions are created implicitly when a trace includes a `sessionId`. All traces sharing the same `sessionId` are grouped together.

```bash
# Both traces join the same session
curl -X POST "https://api.agenttrace.io/api/public/ingestion" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "batch": [
      {
        "type": "trace-create",
        "id": "evt-1",
        "timestamp": "2024-01-15T10:00:00Z",
        "body": {
          "id": "trace-001",
          "name": "step-1",
          "sessionId": "session-abc123"
        }
      },
      {
        "type": "trace-create",
        "id": "evt-2",
        "timestamp": "2024-01-15T10:05:00Z",
        "body": {
          "id": "trace-002",
          "name": "step-2",
          "sessionId": "session-abc123"
        }
      }
    ]
  }'
```
