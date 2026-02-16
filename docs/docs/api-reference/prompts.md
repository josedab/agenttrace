---
sidebar_position: 6
title: Prompts API
description: Endpoints for managing versioned prompts and compiling them with variables.
---

# Prompts API

Prompts are versioned templates managed in AgentTrace. Retrieve them by name and version, and compile them with variables at runtime.

## Get Prompt

Fetch a prompt by name. Optionally filter by version or label.

```
GET /api/public/prompts
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `name` | string | – | **Required.** Prompt name |
| `version` | integer | – | Specific version number |
| `label` | string | – | Label filter (e.g., `production`, `staging`) |

### Example Request

```bash
curl "https://api.agenttrace.io/api/public/prompts?name=code-review&label=production" \
  -H "X-API-Key: sk-at-..."
```

### Response

```json
{
  "id": "prompt-abc123",
  "name": "code-review",
  "version": 3,
  "type": "text",
  "prompt": "Review the following {{language}} code for bugs and improvements:\n\n```\n{{code}}\n```",
  "config": {
    "model": "claude-3-sonnet",
    "temperature": 0.2,
    "maxTokens": 2048
  },
  "labels": ["production"],
  "variables": ["language", "code"],
  "createdAt": "2024-01-10T08:00:00Z",
  "updatedAt": "2024-01-12T14:30:00Z"
}
```

## Compile Prompt

Compile a prompt by substituting variables with provided values.

```
POST /api/public/prompts/:promptName/compile
```

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `promptName` | string | The prompt name |

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `variables` | object | Yes | Key-value pairs for template variables |
| `version` | integer | No | Specific version to compile |
| `label` | string | No | Label filter |

### Example Request

```bash
curl -X POST "https://api.agenttrace.io/api/public/prompts/code-review/compile" \
  -H "X-API-Key: sk-at-..." \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "language": "Python",
      "code": "def add(a, b): return a + b"
    },
    "label": "production"
  }'
```

### Response

```json
{
  "compiledPrompt": "Review the following Python code for bugs and improvements:\n\n```\ndef add(a, b): return a + b\n```",
  "version": 3,
  "config": {
    "model": "claude-3-sonnet",
    "temperature": 0.2,
    "maxTokens": 2048
  }
}
```

## Prompt Types

| Type | Description | Template Format |
|------|-------------|-----------------|
| `text` | Plain text prompt | `{{variable}}` placeholders |
| `chat` | Chat message array | `{{variable}}` in message content |

### Chat Prompt Example

```json
{
  "type": "chat",
  "prompt": [
    {"role": "system", "content": "You are a {{role}} assistant."},
    {"role": "user", "content": "{{user_input}}"}
  ]
}
```
