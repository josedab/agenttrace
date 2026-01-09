---
sidebar_position: 2
---

# Traces API

Traces represent complete units of work in AgentTrace. Each trace can contain multiple observations (spans, generations, events).

## Create Trace

Create a new trace.

```
POST /v1/traces
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Name of the trace |
| `input` | any | No | Input data |
| `output` | any | No | Output data |
| `metadata` | object | No | Custom metadata |
| `tags` | string[] | No | Tags for filtering |
| `sessionId` | string | No | Session to group this trace with |
| `userId` | string | No | User who initiated the trace |
| `version` | string | No | Version identifier |
| `level` | string | No | Log level (DEBUG, INFO, WARNING, ERROR) |
| `public` | boolean | No | Whether trace is publicly accessible |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/v1/traces" \
  -H "Authorization: Bearer sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "code-review",
    "input": {
      "file": "main.py",
      "content": "def hello(): print(\"world\")"
    },
    "metadata": {
      "repository": "my-app",
      "branch": "main"
    },
    "tags": ["production", "code-review"],
    "sessionId": "session-123"
  }'
```

### Response

```json
{
  "id": "trace-abc123",
  "name": "code-review",
  "projectId": "project-xyz",
  "input": {
    "file": "main.py",
    "content": "def hello(): print(\"world\")"
  },
  "output": null,
  "metadata": {
    "repository": "my-app",
    "branch": "main"
  },
  "tags": ["production", "code-review"],
  "sessionId": "session-123",
  "userId": null,
  "version": null,
  "level": "DEFAULT",
  "public": false,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

## Get Trace

Retrieve a trace by ID.

```
GET /v1/traces/:traceId
```

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `traceId` | string | The trace ID |

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `includeObservations` | boolean | false | Include all observations |
| `includeScores` | boolean | false | Include all scores |

### Example Request

```bash
curl "https://api.agenttrace.io/v1/traces/trace-abc123?includeObservations=true&includeScores=true" \
  -H "Authorization: Bearer sk-at-..."
```

### Response

```json
{
  "id": "trace-abc123",
  "name": "code-review",
  "projectId": "project-xyz",
  "input": {
    "file": "main.py",
    "content": "def hello(): print(\"world\")"
  },
  "output": {
    "review": "Code looks good!",
    "suggestions": []
  },
  "metadata": {
    "repository": "my-app"
  },
  "tags": ["production"],
  "sessionId": "session-123",
  "level": "DEFAULT",
  "latency": 1234,
  "totalCost": 0.0023,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:31:00Z",
  "observations": [
    {
      "id": "obs-123",
      "traceId": "trace-abc123",
      "type": "GENERATION",
      "name": "analyze-code",
      "model": "claude-3-sonnet",
      "input": [...],
      "output": "...",
      "startTime": "2024-01-15T10:30:01Z",
      "endTime": "2024-01-15T10:30:05Z",
      "latency": 4000,
      "usage": {
        "inputTokens": 150,
        "outputTokens": 200
      },
      "cost": {
        "input": 0.0015,
        "output": 0.0006,
        "total": 0.0021
      }
    }
  ],
  "scores": [
    {
      "id": "score-456",
      "name": "quality",
      "value": 0.95,
      "comment": "Excellent review",
      "source": "API",
      "createdAt": "2024-01-15T10:32:00Z"
    }
  ]
}
```

## List Traces

List traces with filtering and pagination.

```
GET /v1/traces
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1-1000) |
| `cursor` | string | - | Pagination cursor |
| `name` | string | - | Filter by name (exact match) |
| `sessionId` | string | - | Filter by session |
| `userId` | string | - | Filter by user |
| `tags` | string | - | Filter by tag (comma-separated) |
| `level` | string | - | Filter by level |
| `fromTimestamp` | string | - | Start time (ISO 8601) |
| `toTimestamp` | string | - | End time (ISO 8601) |
| `orderBy` | string | createdAt | Field to order by |
| `order` | string | desc | Order direction (asc/desc) |

### Example Request

```bash
curl "https://api.agenttrace.io/v1/traces?limit=20&tags=production&fromTimestamp=2024-01-01T00:00:00Z" \
  -H "Authorization: Bearer sk-at-..."
```

### Response

```json
{
  "data": [
    {
      "id": "trace-abc123",
      "name": "code-review",
      "sessionId": "session-123",
      "input": {...},
      "output": {...},
      "latency": 1234,
      "totalCost": 0.0023,
      "level": "DEFAULT",
      "tags": ["production"],
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "totalCount": 156,
    "hasMore": true,
    "nextCursor": "eyJpZCI6InRyYWNlLWFiYzEyMiJ9"
  }
}
```

## Update Trace

Update an existing trace.

```
PATCH /v1/traces/:traceId
```

### Request Body

| Field | Type | Description |
|-------|------|-------------|
| `output` | any | Output data |
| `metadata` | object | Merged with existing metadata |
| `tags` | string[] | Replaces existing tags |
| `level` | string | Log level |
| `public` | boolean | Public visibility |

### Example Request

```bash
curl -X PATCH "https://api.agenttrace.io/v1/traces/trace-abc123" \
  -H "Authorization: Bearer sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "output": {
      "review": "Code looks good!",
      "suggestions": []
    },
    "level": "INFO"
  }'
```

## Delete Trace

Delete a trace and all its observations.

```
DELETE /v1/traces/:traceId
```

### Example Request

```bash
curl -X DELETE "https://api.agenttrace.io/v1/traces/trace-abc123" \
  -H "Authorization: Bearer sk-at-..."
```

### Response

```json
{
  "success": true
}
```

## Batch Ingestion

Ingest multiple traces and observations in a single request. This is the recommended approach for high-volume scenarios.

```
POST /api/public/ingestion
```

### Request Body

| Field | Type | Description |
|-------|------|-------------|
| `batch` | array | Array of events to ingest |

Each event in the batch has:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Event type: `trace-create`, `observation-create`, `score-create` |
| `id` | string | Yes | Idempotency key |
| `timestamp` | string | Yes | Event timestamp (ISO 8601) |
| `body` | object | Yes | Event data |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/ingestion" \
  -H "Authorization: Bearer sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "batch": [
      {
        "type": "trace-create",
        "id": "evt-1",
        "timestamp": "2024-01-15T10:30:00Z",
        "body": {
          "id": "trace-123",
          "name": "my-task",
          "input": {"query": "hello"}
        }
      },
      {
        "type": "observation-create",
        "id": "evt-2",
        "timestamp": "2024-01-15T10:30:01Z",
        "body": {
          "id": "obs-456",
          "traceId": "trace-123",
          "type": "GENERATION",
          "name": "llm-call",
          "model": "gpt-4",
          "input": [{"role": "user", "content": "hello"}],
          "output": "Hi there!",
          "startTime": "2024-01-15T10:30:01Z",
          "endTime": "2024-01-15T10:30:02Z",
          "usage": {
            "inputTokens": 5,
            "outputTokens": 3
          }
        }
      },
      {
        "type": "score-create",
        "id": "evt-3",
        "timestamp": "2024-01-15T10:30:05Z",
        "body": {
          "traceId": "trace-123",
          "name": "quality",
          "value": 0.95
        }
      }
    ]
  }'
```

### Response

```json
{
  "successes": [
    {"id": "evt-1", "status": 201},
    {"id": "evt-2", "status": 201},
    {"id": "evt-3", "status": 201}
  ],
  "errors": []
}
```

## Get Trace Tree

Get the complete observation tree for a trace.

```
GET /v1/traces/:traceId/tree
```

### Response

```json
{
  "trace": {
    "id": "trace-abc123",
    "name": "code-review",
    "latency": 5000
  },
  "observations": [
    {
      "id": "obs-1",
      "parentId": null,
      "type": "SPAN",
      "name": "analyze",
      "latency": 5000,
      "children": [
        {
          "id": "obs-2",
          "parentId": "obs-1",
          "type": "GENERATION",
          "name": "llm-call",
          "latency": 3000,
          "children": []
        },
        {
          "id": "obs-3",
          "parentId": "obs-1",
          "type": "SPAN",
          "name": "post-process",
          "latency": 500,
          "children": []
        }
      ]
    }
  ]
}
```

## Webhooks

Configure webhooks to receive trace events. See [Webhooks](/api-reference/webhooks) for details.
