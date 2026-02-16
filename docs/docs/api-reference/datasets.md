---
sidebar_position: 7
title: Datasets API
description: Endpoints for managing datasets and dataset items for evaluation and testing.
---

# Datasets API

Datasets store collections of input/expected-output pairs used for evaluation runs and regression testing.

## Create Dataset

```
POST /api/public/datasets
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Dataset name (unique per project) |
| `description` | string | No | Dataset description |
| `metadata` | object | No | Custom metadata |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/datasets" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "code-review-golden",
    "description": "Golden test set for code review evaluations"
  }'
```

### Response

```json
{
  "id": "dataset-abc123",
  "name": "code-review-golden",
  "description": "Golden test set for code review evaluations",
  "metadata": {},
  "itemCount": 0,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-15T10:00:00Z"
}
```

## List Datasets

```
GET /api/public/datasets
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Max results (1–1000) |
| `cursor` | string | – | Pagination cursor |

### Response

```json
{
  "data": [
    {
      "id": "dataset-abc123",
      "name": "code-review-golden",
      "description": "Golden test set for code review evaluations",
      "itemCount": 24,
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": { "totalCount": 3, "hasMore": false, "nextCursor": null }
}
```

## Get Dataset

```
GET /api/public/datasets/:datasetId
```

### Response

```json
{
  "id": "dataset-abc123",
  "name": "code-review-golden",
  "description": "Golden test set for code review evaluations",
  "metadata": {},
  "itemCount": 24,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-20T09:00:00Z"
}
```

## Create Dataset Item

```
POST /api/public/datasets/:datasetId/items
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `input` | any | Yes | Input data for the test case |
| `expectedOutput` | any | No | Expected output |
| `metadata` | object | No | Custom metadata |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/datasets/dataset-abc123/items" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "input": {"code": "def foo(): pass", "language": "python"},
    "expectedOutput": {"issues": ["Empty function body"]},
    "metadata": {"difficulty": "easy"}
  }'
```

### Response

```json
{
  "id": "item-001",
  "datasetId": "dataset-abc123",
  "input": {"code": "def foo(): pass", "language": "python"},
  "expectedOutput": {"issues": ["Empty function body"]},
  "metadata": {"difficulty": "easy"},
  "createdAt": "2024-01-15T11:00:00Z"
}
```

## Delete Dataset

```
DELETE /api/public/datasets/:datasetId
```

### Response

```json
{ "success": true }
```
