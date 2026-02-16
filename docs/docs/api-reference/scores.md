---
sidebar_position: 5
title: Scores API
description: Endpoints for creating and retrieving evaluation scores on traces and observations.
---

# Scores API

Scores attach evaluation results to traces or observations. Supported types: **numeric** (0–1 float), **boolean** (true/false), and **categorical** (string label).

## Create Score

```
POST /api/public/scores
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `traceId` | string | Yes | Target trace ID |
| `observationId` | string | No | Target observation ID |
| `name` | string | Yes | Score name |
| `value` | number/boolean/string | Yes | Score value |
| `dataType` | string | No | `NUMERIC`, `BOOLEAN`, or `CATEGORICAL` (auto-detected) |
| `comment` | string | No | Human-readable comment |
| `source` | string | No | Score source: `API`, `EVAL`, `ANNOTATION` |

### Example – Numeric Score

```bash
curl -X POST "https://api.agenttrace.io/api/public/scores" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc123",
    "name": "relevance",
    "value": 0.92,
    "dataType": "NUMERIC",
    "comment": "Highly relevant response"
  }'
```

### Example – Boolean Score

```bash
curl -X POST "https://api.agenttrace.io/api/public/scores" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc123",
    "observationId": "obs-456",
    "name": "hallucination",
    "value": false,
    "dataType": "BOOLEAN"
  }'
```

### Example – Categorical Score

```bash
curl -X POST "https://api.agenttrace.io/api/public/scores" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc123",
    "name": "sentiment",
    "value": "positive",
    "dataType": "CATEGORICAL"
  }'
```

### Response

```json
{
  "id": "score-789",
  "traceId": "trace-abc123",
  "observationId": null,
  "name": "relevance",
  "value": 0.92,
  "dataType": "NUMERIC",
  "comment": "Highly relevant response",
  "source": "API",
  "createdAt": "2024-01-15T10:35:00Z"
}
```

## List Scores

```
GET /api/public/scores
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |
| `traceId` | string | – | Filter by trace |
| `observationId` | string | – | Filter by observation |
| `name` | string | – | Filter by score name |
| `source` | string | – | Filter by source |
| `dataType` | string | – | Filter by data type |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/scores?traceId=trace-abc123&name=relevance" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "data": [
    {
      "id": "score-789",
      "traceId": "trace-abc123",
      "observationId": null,
      "name": "relevance",
      "value": 0.92,
      "dataType": "NUMERIC",
      "source": "API",
      "comment": "Highly relevant response",
      "createdAt": "2024-01-15T10:35:00Z"
    }
  ],
  "meta": {
    "totalCount": 12,
    "hasMore": false,
    "nextCursor": null
  }
}
```
