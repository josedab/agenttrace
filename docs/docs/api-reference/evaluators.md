---
sidebar_position: 8
title: Evaluators API
description: Endpoints for configuring and managing automated evaluators.
---

# Evaluators API

Evaluators are configurable scoring functions that automatically assess trace and observation quality.

## Create Evaluator

```
POST /api/public/evaluators
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Evaluator name |
| `description` | string | No | Evaluator description |
| `type` | string | Yes | `LLM_JUDGE`, `RULE_BASED`, or `CUSTOM` |
| `config` | object | Yes | Type-specific configuration |
| `scoreName` | string | Yes | Name of the score produced |
| `dataType` | string | No | Output data type: `NUMERIC`, `BOOLEAN`, `CATEGORICAL` |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/evaluators" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hallucination-check",
    "type": "LLM_JUDGE",
    "scoreName": "hallucination",
    "dataType": "BOOLEAN",
    "config": {
      "model": "claude-3-sonnet",
      "prompt": "Does the output contain hallucinated facts? Respond true or false.",
      "inputMapping": {
        "input": "{{trace.input}}",
        "output": "{{trace.output}}"
      }
    }
  }'
```

### Response

```json
{
  "id": "eval-abc123",
  "name": "hallucination-check",
  "type": "LLM_JUDGE",
  "scoreName": "hallucination",
  "dataType": "BOOLEAN",
  "config": { "model": "claude-3-sonnet", "prompt": "..." },
  "enabled": true,
  "createdAt": "2024-01-15T10:00:00Z"
}
```

## List Evaluators

```
GET /api/public/evaluators
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–100) |
| `cursor` | string | – | Pagination cursor |
| `type` | string | – | Filter by evaluator type |
| `enabled` | boolean | – | Filter by enabled status |

### Response

```json
{
  "data": [
    {
      "id": "eval-abc123",
      "name": "hallucination-check",
      "type": "LLM_JUDGE",
      "scoreName": "hallucination",
      "dataType": "BOOLEAN",
      "enabled": true,
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": { "totalCount": 5, "hasMore": false, "nextCursor": null }
}
```

## Update Evaluator

```
PATCH /api/public/evaluators/:evaluatorId
```

### Request Body

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | boolean | Enable or disable the evaluator |
| `config` | object | Updated configuration |
| `description` | string | Updated description |

### Example

```bash
curl -X PATCH "https://api.agenttrace.io/api/public/evaluators/eval-abc123" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{ "enabled": false }'
```

## Delete Evaluator

```
DELETE /api/public/evaluators/:evaluatorId
```

### Response

```json
{ "success": true }
```
