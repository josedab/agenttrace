---
sidebar_position: 6
title: Migration from Langfuse
description: Guide for migrating from Langfuse to AgentTrace — SDK mapping, API differences, and step-by-step instructions.
---

# Migration from Langfuse

This guide helps you migrate from Langfuse to AgentTrace. The APIs are intentionally similar, so migration is straightforward.

## SDK Mapping

### Python

| Langfuse | AgentTrace | Notes |
|----------|------------|-------|
| `from langfuse import Langfuse` | `from agenttrace import AgentTrace` | Drop-in replacement |
| `from langfuse.decorators import observe` | `from agenttrace import observe` | Same decorator pattern |
| `langfuse.generation()` | `client.generation()` | Same API |
| `langfuse.span()` | `client.span()` | Same API |
| `langfuse.flush()` | `client.flush()` | Same API |
| `langfuse.shutdown()` | `client.shutdown()` | Same API |

**Before (Langfuse):**
```python
from langfuse import Langfuse
from langfuse.decorators import observe

client = Langfuse(
    public_key="pk-...",
    secret_key="sk-...",
    host="https://cloud.langfuse.com"
)

@observe()
def my_function(query: str) -> str:
    return call_llm(query)
```

**After (AgentTrace):**
```python
from agenttrace import AgentTrace, observe

client = AgentTrace(
    api_key="your-api-key",
    host="http://localhost:8080"
)

@observe()
def my_function(query: str) -> str:
    return call_llm(query)
```

### TypeScript

| Langfuse | AgentTrace | Notes |
|----------|------------|-------|
| `new Langfuse({ publicKey, secretKey })` | `new AgentTrace({ apiKey })` | Single API key |
| `langfuse.trace()` | `client.trace()` | Same API |
| `trace.generation()` | `trace.generation()` | Same API |
| `langfuse.shutdownAsync()` | `client.shutdown()` | Same API |

### Go

AgentTrace provides a **native Go SDK** — something Langfuse doesn't offer. If you were using Langfuse's REST API from Go, you can now use the idiomatic SDK:

```go
import agenttrace "github.com/agenttrace/agenttrace/sdk/go"

client := agenttrace.New(agenttrace.Config{
    APIKey: "your-api-key",
    Host:   "http://localhost:8080",
})
defer client.Shutdown()
```

## API Differences

### Authentication

| Langfuse | AgentTrace |
|----------|------------|
| Basic Auth with `publicKey:secretKey` | Single API key via `X-API-Key` header |

### Endpoints

| Operation | Langfuse | AgentTrace |
|-----------|----------|------------|
| Batch ingestion | `POST /api/public/ingestion` | `POST /api/public/ingestion` |
| Get traces | `GET /api/public/traces` | `GET /api/public/traces` |
| Get trace | `GET /api/public/traces/:id` | `GET /api/public/traces/:id` |
| Get prompts | `GET /api/public/v2/prompts/:name` | `GET /api/public/prompts?name=...` |
| Create score | `POST /api/public/scores` | `POST /api/public/scores` |

Most endpoints are compatible. Key differences:
- **Authentication**: Use `X-API-Key` header instead of Basic Auth
- **Prompt API**: Slightly different URL structure
- **Additional endpoints**: AgentTrace adds `/api/public/checkpoints`, `/api/public/git-links`, `/api/public/ci-runs`

## Environment Variables

| Langfuse | AgentTrace | Notes |
|----------|------------|-------|
| `LANGFUSE_PUBLIC_KEY` | `AGENTTRACE_API_KEY` | Single key |
| `LANGFUSE_SECRET_KEY` | *(not needed)* | Simplified auth |
| `LANGFUSE_HOST` | `AGENTTRACE_HOST` | Same purpose |

## Step-by-Step Migration

### 1. Deploy AgentTrace

```bash
cd deploy
cp .env.example .env
# Edit .env with your credentials
docker compose up -d
```

### 2. Create an API Key

Navigate to Settings → API Keys in the AgentTrace dashboard.

### 3. Update SDK Dependencies

```bash
# Python
pip uninstall langfuse
pip install agenttrace

# TypeScript
npm uninstall langfuse
npm install agenttrace
```

### 4. Update Imports and Configuration

Find-and-replace in your codebase:
- `langfuse` → `agenttrace`
- `Langfuse` → `AgentTrace`
- `LANGFUSE_PUBLIC_KEY` → `AGENTTRACE_API_KEY`
- `LANGFUSE_HOST` → `AGENTTRACE_HOST`

### 5. Remove Secret Key References

AgentTrace uses a single API key instead of public/secret key pairs. Remove any `secret_key` or `secretKey` parameters.

### 6. Verify

Run your application and check the AgentTrace dashboard for incoming traces.

## What You Gain

After migrating, you get access to AgentTrace-exclusive features:
- **Git linking** — Correlate traces with git commits
- **Code checkpoints** — Snapshot file state during agent execution
- **File operation tracking** — See every file read/write
- **Terminal command logging** — Capture subprocess execution
- **CLI wrapper** — Trace any command-line tool with zero code changes
- **ClickHouse analytics** — Sub-second queries on billions of traces
