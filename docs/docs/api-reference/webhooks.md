---
sidebar_position: 12
title: Webhooks API
description: Endpoints for configuring webhooks and subscribing to AgentTrace events.
---

# Webhooks API

Webhooks deliver real-time HTTP callbacks when events occur in AgentTrace, such as trace completion or score creation.

## Create Webhook

```
POST /api/public/webhooks
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | Yes | HTTPS endpoint to receive events |
| `events` | string[] | Yes | Event types to subscribe to |
| `secret` | string | No | Shared secret for HMAC signature verification |
| `enabled` | boolean | No | Whether the webhook is active (default: `true`) |
| `description` | string | No | Description |

### Event Types

| Event | Description |
|-------|-------------|
| `trace.created` | A new trace was created |
| `trace.completed` | A trace finished execution |
| `score.created` | A score was attached to a trace |
| `evaluation.completed` | An evaluator finished running |
| `ci-run.completed` | A CI run status changed to terminal |
| `dataset.updated` | A dataset or its items were modified |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/webhooks" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/hooks/agenttrace",
    "events": ["trace.completed", "score.created"],
    "secret": "whsec_your_secret_key",
    "description": "Notify on trace completion and scoring"
  }'
```

### Response

```json
{
  "id": "wh-abc123",
  "url": "https://example.com/hooks/agenttrace",
  "events": ["trace.completed", "score.created"],
  "enabled": true,
  "description": "Notify on trace completion and scoring",
  "createdAt": "2024-01-15T10:00:00Z"
}
```

## List Webhooks

```
GET /api/public/webhooks
```

### Response

```json
{
  "data": [
    {
      "id": "wh-abc123",
      "url": "https://example.com/hooks/agenttrace",
      "events": ["trace.completed", "score.created"],
      "enabled": true,
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": { "totalCount": 2, "hasMore": false, "nextCursor": null }
}
```

## Update Webhook

```
PATCH /api/public/webhooks/:webhookId
```

### Request Body

| Field | Type | Description |
|-------|------|-------------|
| `url` | string | Updated URL |
| `events` | string[] | Updated event subscriptions |
| `enabled` | boolean | Enable or disable |

## Delete Webhook

```
DELETE /api/public/webhooks/:webhookId
```

### Response

```json
{ "success": true }
```

## Webhook Payload Format

All webhook deliveries use the following envelope:

```json
{
  "id": "evt-789",
  "type": "trace.completed",
  "timestamp": "2024-01-15T10:35:00Z",
  "data": {
    "traceId": "trace-abc123",
    "name": "code-review",
    "latency": 5000,
    "totalCost": 0.0023
  }
}
```

### Signature Verification

Requests include an `X-AgentTrace-Signature` header containing an HMAC-SHA256 digest of the payload body using your webhook secret:

```
X-AgentTrace-Signature: sha256=abc123def456...
```
