---
sidebar_position: 3
title: Observations API
description: Endpoints for managing observations (spans, generations, events) within traces.
---

# Observations API

Observations represent individual units of work within a trace. There are three types: **SPAN** (generic timed operation), **GENERATION** (LLM call), and **EVENT** (point-in-time marker).

## List Observations

```
GET /api/public/observations
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |
| `traceId` | string | – | Filter by trace |
| `type` | string | – | Filter by type: `SPAN`, `GENERATION`, `EVENT` |
| `name` | string | – | Filter by name |
| `fromTimestamp` | string | – | Start time (ISO 8601) |
| `toTimestamp` | string | – | End time (ISO 8601) |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/observations?traceId=trace-abc123&type=GENERATION&limit=10" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "data": [
    {
      "id": "obs-456",
      "traceId": "trace-abc123",
      "type": "GENERATION",
      "name": "llm-call",
      "parentObservationId": "obs-100",
      "model": "claude-3-sonnet",
      "input": [{"role": "user", "content": "Summarize this code"}],
      "output": "This function handles...",
      "startTime": "2024-01-15T10:30:01Z",
      "endTime": "2024-01-15T10:30:03Z",
      "usage": {"inputTokens": 120, "outputTokens": 85},
      "cost": {"input": 0.0012, "output": 0.0004, "total": 0.0016},
      "level": "DEFAULT",
      "metadata": {}
    }
  ],
  "meta": {
    "totalCount": 42,
    "hasMore": true,
    "nextCursor": "eyJpZCI6Im9icy00NTUifQ=="
  }
}
```

## Get Observation

```
GET /api/public/observations/:observationId
```

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `observationId` | string | The observation ID |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/observations/obs-456" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "id": "obs-456",
  "traceId": "trace-abc123",
  "type": "GENERATION",
  "name": "llm-call",
  "parentObservationId": "obs-100",
  "model": "claude-3-sonnet",
  "input": [{"role": "user", "content": "Summarize this code"}],
  "output": "This function handles authentication...",
  "startTime": "2024-01-15T10:30:01Z",
  "endTime": "2024-01-15T10:30:03Z",
  "completionStartTime": "2024-01-15T10:30:02Z",
  "usage": {"inputTokens": 120, "outputTokens": 85},
  "cost": {"input": 0.0012, "output": 0.0004, "total": 0.0016},
  "level": "DEFAULT",
  "metadata": {"provider": "anthropic"},
  "version": "1.0.0"
}
```

## Observation Schema

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `traceId` | string | Parent trace ID |
| `type` | string | `SPAN`, `GENERATION`, or `EVENT` |
| `name` | string | Observation name |
| `parentObservationId` | string | Parent observation for nesting |
| `model` | string | Model name (GENERATION only) |
| `input` | any | Input data |
| `output` | any | Output data |
| `startTime` | string | Start timestamp (ISO 8601) |
| `endTime` | string | End timestamp (ISO 8601) |
| `usage` | object | Token usage (`inputTokens`, `outputTokens`, `totalTokens`) |
| `cost` | object | Cost breakdown (`input`, `output`, `total`) |
| `level` | string | `DEBUG`, `DEFAULT`, `WARNING`, `ERROR` |
| `metadata` | object | Custom metadata |
