---
sidebar_position: 1
---

# API Reference

AgentTrace provides a comprehensive REST API for all platform functionality. The API is designed to be compatible with Langfuse, making migration seamless.

## Base URL

**Cloud**: `https://api.agenttrace.io`

**Self-hosted**: Your configured API URL (e.g., `https://agenttrace.your-company.com/api`)

## Authentication

All API requests require authentication using an API key. Include your key in the `Authorization` header:

```bash
curl -X GET "https://api.agenttrace.io/v1/traces" \
  -H "Authorization: Bearer sk-at-your-api-key"
```

See [Authentication](/api-reference/authentication) for more details.

## Content Type

All requests and responses use JSON:

```
Content-Type: application/json
```

## Rate Limiting

| Tier | Requests/minute | Burst |
|------|-----------------|-------|
| Free | 60 | 10 |
| Pro | 600 | 100 |
| Enterprise | 6000 | 1000 |

Rate limit headers are included in all responses:

```
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 599
X-RateLimit-Reset: 1704067200
```

## Pagination

List endpoints support cursor-based pagination:

```bash
GET /v1/traces?limit=50&cursor=eyJpZCI6IjEyMzQifQ==
```

Response includes pagination info:

```json
{
  "data": [...],
  "meta": {
    "totalCount": 1234,
    "hasMore": true,
    "nextCursor": "eyJpZCI6IjU2Nzgifq=="
  }
}
```

## Error Handling

Errors return appropriate HTTP status codes with detailed messages:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "The 'traceId' field is required",
    "details": {
      "field": "traceId"
    }
  }
}
```

### Common Error Codes

| Status | Code | Description |
|--------|------|-------------|
| 400 | `invalid_request` | Malformed request body or parameters |
| 401 | `unauthorized` | Missing or invalid API key |
| 403 | `forbidden` | API key lacks required permissions |
| 404 | `not_found` | Resource doesn't exist |
| 409 | `conflict` | Resource already exists |
| 429 | `rate_limited` | Too many requests |
| 500 | `internal_error` | Server error |

## API Endpoints

### Core

| Endpoint | Description |
|----------|-------------|
| [POST /v1/traces](/api-reference/traces#create-trace) | Create a new trace |
| [GET /v1/traces](/api-reference/traces#list-traces) | List traces |
| [GET /v1/traces/:id](/api-reference/traces#get-trace) | Get a trace |
| [POST /v1/observations](/api-reference/observations#create-observation) | Create an observation |
| [GET /v1/sessions](/api-reference/sessions) | List sessions |
| [POST /v1/scores](/api-reference/scores#create-score) | Create a score |

### Prompts

| Endpoint | Description |
|----------|-------------|
| [POST /v1/prompts](/api-reference/prompts#create-prompt) | Create a prompt |
| [GET /v1/prompts](/api-reference/prompts#list-prompts) | List prompts |
| [GET /v1/prompts/:name](/api-reference/prompts#get-prompt) | Get prompt by name |
| [POST /v1/prompts/:name/compile](/api-reference/prompts#compile-prompt) | Compile prompt with variables |

### Datasets

| Endpoint | Description |
|----------|-------------|
| [POST /v1/datasets](/api-reference/datasets#create-dataset) | Create a dataset |
| [GET /v1/datasets](/api-reference/datasets#list-datasets) | List datasets |
| [POST /v1/datasets/:id/items](/api-reference/datasets#add-item) | Add item to dataset |
| [POST /v1/datasets/:id/runs](/api-reference/datasets#create-run) | Create experiment run |

### Evaluators

| Endpoint | Description |
|----------|-------------|
| [POST /v1/evaluators](/api-reference/evaluators#create-evaluator) | Create an evaluator |
| [GET /v1/evaluators](/api-reference/evaluators#list-evaluators) | List evaluators |
| [POST /v1/evaluators/:id/execute](/api-reference/evaluators#execute) | Execute evaluator |

### Agent Features

| Endpoint | Description |
|----------|-------------|
| [POST /v1/checkpoints](/api-reference/checkpoints#create-checkpoint) | Create checkpoint |
| [GET /v1/checkpoints](/api-reference/checkpoints#list-checkpoints) | List checkpoints |
| [POST /v1/git-links](/api-reference/git-links#create-git-link) | Link commit to trace |
| [POST /v1/ci-runs](/api-reference/ci-runs#create-ci-run) | Create CI run |

### Ingestion (Langfuse Compatible)

| Endpoint | Description |
|----------|-------------|
| [POST /api/public/ingestion](/api-reference/traces#ingestion) | Batch ingestion endpoint |

## OpenAPI Specification

Download the full OpenAPI specification:

- [openapi.yaml](/openapi.yaml) - OpenAPI 3.0 spec
- [openapi.json](/openapi.json) - JSON format

## SDKs

Official SDKs handle authentication, retries, and batching automatically:

- [Python SDK](/sdks/python)
- [TypeScript SDK](/sdks/typescript)
- [Go SDK](/sdks/go)
- [CLI](/sdks/cli)

## Examples

### Create a Trace with Observations

```bash
# Create trace
TRACE=$(curl -s -X POST "https://api.agenttrace.io/v1/traces" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "code-review",
    "input": {"file": "main.py"},
    "metadata": {"repository": "my-app"}
  }')

TRACE_ID=$(echo $TRACE | jq -r '.id')

# Add a generation observation
curl -X POST "https://api.agenttrace.io/v1/observations" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"traceId\": \"$TRACE_ID\",
    \"type\": \"generation\",
    \"name\": \"analyze-code\",
    \"model\": \"claude-3-sonnet\",
    \"input\": [{\"role\": \"user\", \"content\": \"Review this code...\"}],
    \"output\": \"The code looks good...\",
    \"usage\": {
      \"inputTokens\": 150,
      \"outputTokens\": 200
    }
  }"

# Add a score
curl -X POST "https://api.agenttrace.io/v1/scores" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"traceId\": \"$TRACE_ID\",
    \"name\": \"quality\",
    \"value\": 0.95,
    \"comment\": \"High-quality review\"
  }"
```

### Fetch a Prompt and Use It

```bash
# Get the latest production prompt
PROMPT=$(curl -s "https://api.agenttrace.io/v1/prompts/code-review?label=production" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY")

# Compile with variables
COMPILED=$(curl -s -X POST "https://api.agenttrace.io/v1/prompts/code-review/compile" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "language": "Python",
      "code": "def hello(): print(\"world\")"
    }
  }')

echo $COMPILED | jq '.compiledPrompt'
```
