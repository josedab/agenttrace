---
sidebar_position: 6
title: Migration from Langfuse
description: Guide for migrating from Langfuse to AgentTrace — SDK mapping, API differences, and step-by-step instructions.
---

# Migration from Langfuse

This guide covers the documented compatibility subset. AgentTrace does not claim complete Langfuse API parity.

## Import Existing Data

Export JSON from Langfuse, then validate locally:

```bash
agenttrace migrate validate \
  --source langfuse \
  --source-file ./langfuse-export.json
```

Preview without writing:

```bash
agenttrace migrate \
  --source langfuse \
  --source-file ./langfuse-export.json \
  --dry-run
```

Run the import:

```bash
agenttrace migrate \
  --source langfuse \
  --source-file ./langfuse-export.json
```

The supported export arrays are:

- `traces`
- `observations` (`SPAN`, `GENERATION`, and `EVENT`)
- `scores`
- `prompts`

Datasets, media, annotation queues, and Langfuse-specific organization settings are not imported.

The CLI fingerprints the file and derives a stable job ID. Batches are capped at 500 records, and each source type/ID is recorded in an idempotency ledger. Rerunning the same command resumes safely and skips records that the ledger already reports as imported. Errors are redacted before persistence or display.

### Duplicate safety

Trace, observation, and score imports derive deterministic identifiers from their source IDs; scores previously received random identifiers. Prompt imports instead reuse an existing project prompt version when its name and content match. Together these rules close the window between writing a record and writing its ledger entry:

- if the ledger write fails, the retry rewrites the **same** row instead of creating a duplicate
- a ledger entry that records an import without an identifier is treated as incomplete and imported again
- an identical prompt version is reused instead of appending a second version
- concurrent duplicate batches for deterministic record types converge on the same rows, and the ledger upsert keeps one monotonic entry per source item
- a successful retry clears the earlier error, including a ledger error, so a resumed job is not reported as failed

Repeating a batch that is already recorded performs no writes and counts the records as skipped.

Server-side database DSNs are intentionally unsupported. This avoids sending source credentials to the AgentTrace API and works with no-egress deployments because the CLI reads the export locally. In no-egress mode the API refuses migrations whose source DSN is anything other than the local `json-export` mode with `422 Unprocessable Entity`.

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

## Switch Ongoing Instrumentation

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
