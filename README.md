# AgentTrace

[![CI](https://github.com/agenttrace/agenttrace/actions/workflows/ci.yml/badge.svg)](https://github.com/agenttrace/agenttrace/actions/workflows/ci.yml)
[![Security](https://github.com/agenttrace/agenttrace/actions/workflows/security.yml/badge.svg)](https://github.com/agenttrace/agenttrace/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/agenttrace/agenttrace/branch/main/graph/badge.svg)](https://codecov.io/gh/agenttrace/agenttrace)
[![Go Version](https://img.shields.io/github/go-mod/go-version/agenttrace/agenttrace?filename=api%2Fgo.mod)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Open-source observability platform for AI coding agents. LLM traces, evaluations, prompt management, and dataset experiments with Langfuse feature parity.

<p align="center">
  <img src="docs/static/img/dashboard-screenshot.png" alt="AgentTrace Dashboard" width="800" />
  <br />
  <em>AgentTrace dashboard showing real-time trace exploration with agent graph views</em>
</p>

## Features

- **Trace Exploration**: Real-time trace visualization with parent-child relationships, latency waterfall, and agent graph views
- **Prompt Management**: Version-controlled prompts with A/B testing, playground, and SDK integration
- **Evaluators & Scores**: Built-in evaluators (LLM-as-Judge, regex, JSON schema) and custom scoring
- **Datasets & Experiments**: Create datasets, run experiments, and compare results
- **Multi-language SDKs**: Python, TypeScript, Go SDKs with auto-instrumentation
- **CLI Wrapper**: Wrap any CLI tool for automatic tracing
- **Cost Tracking**: Automatic cost calculation for 400+ LLM models

## Quick Start

### Deploy with One Click

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/agenttrace)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/agenttrace/agenttrace)

Or deploy locally with Docker Compose:

### 1. Start the Services

```bash
# Clone the repository
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace

# Start with Docker Compose
cd deploy
cp .env.example .env
# Edit .env with your credentials
docker compose up -d
```

### 2. Access the Dashboard

Open [http://localhost:3000](http://localhost:3000) in your browser.

### 3. Create an API Key

Navigate to Settings > API Keys and create a new key.

### 4. Install the SDK

**Python:**
```bash
pip install agenttrace
```

**TypeScript:**
```bash
npm install agenttrace
```

**Go:**
```bash
go get github.com/agenttrace/agenttrace/sdk/go
```

### 5. Start Tracing

**Python:**
```python
from agenttrace import AgentTrace, observe

client = AgentTrace(
    api_key="your-api-key",
    host="http://localhost:8080"
)

@observe()
def my_llm_function(query: str) -> str:
    # Your LLM call here
    return "response"

result = my_llm_function("Hello, world!")
client.flush()
```

**TypeScript:**
```typescript
import { AgentTrace, observe } from 'agenttrace';

const client = new AgentTrace({
  apiKey: 'your-api-key',
  host: 'http://localhost:8080'
});

const myLLMFunction = observe(async (query: string) => {
  // Your LLM call here
  return 'response';
});

const result = await myLLMFunction('Hello, world!');
await client.flush();
```

**Go:**
```go
package main

import (
    "context"
    agenttrace "github.com/agenttrace/agenttrace/sdk/go"
)

func main() {
    client := agenttrace.New(agenttrace.Config{
        APIKey: "your-api-key",
        Host:   "http://localhost:8080",
    })
    defer client.Shutdown()

    ctx := context.Background()
    trace := client.Trace(ctx, agenttrace.TraceOptions{
        Name: "my-trace",
    })

    gen := trace.Generation(agenttrace.GenerationOptions{
        Name:  "llm-call",
        Model: "gpt-4",
        Input: map[string]any{"query": "Hello"},
    })
    gen.End(agenttrace.GenerationEndOptions{
        Output: "Hi there!",
    })

    trace.End(nil)
}
```

## Architecture

```
                    ┌─────────────────────────────────────────┐
                    │           Next.js Frontend              │
                    │   Dashboard, Traces, Prompts, Evals     │
                    └─────────────────────────────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────────┐
                    │            Go Backend (Fiber)           │
                    │   REST API, GraphQL, OTLP Ingestion     │
                    └─────────────────────────────────────────┘
                         │              │              │
                         ▼              ▼              ▼
                    ┌──────────┐  ┌──────────┐  ┌──────────┐
                    │ClickHouse│  │PostgreSQL│  │  Redis   │
                    │ (Traces) │  │(Metadata)│  │ (Queue)  │
                    └──────────┘  └──────────┘  └──────────┘
```

## Project Structure

```
agenttrace/
├── api/                    # Go backend
│   ├── cmd/               # Entry points (server, worker)
│   ├── internal/          # Internal packages
│   │   ├── domain/        # Domain models
│   │   ├── repository/    # Data access
│   │   ├── service/       # Business logic
│   │   ├── handler/       # HTTP handlers
│   │   └── graphql/       # GraphQL resolvers
│   ├── migrations/        # Database migrations
│   └── schema/            # GraphQL schema
├── web/                   # Next.js frontend
│   ├── app/              # App router pages
│   ├── components/       # React components
│   ├── hooks/            # React hooks
│   └── lib/              # Utilities
├── sdk/                   # Language SDKs
│   ├── python/           # Python SDK
│   ├── typescript/       # TypeScript SDK
│   ├── go/               # Go SDK
│   └── cli/              # CLI wrapper
├── deploy/               # Deployment configs
│   ├── docker-compose.yml
│   └── .env.example
├── docs/                  # Documentation (Docusaurus)
├── examples/              # Example projects
├── extensions/            # IDE extensions (VS Code, JetBrains)
├── actions/               # GitHub Actions (reusable workflows)
├── ci/                    # CI pipeline configuration
├── scripts/               # Development and build scripts
└── mobile/                # Mobile app (experimental)
```

## SDK Features

### Python SDK

```python
from agenttrace import AgentTrace, observe, generation

# Initialize client
client = AgentTrace(api_key="...")

# Decorator-based tracing
@observe()
def my_function():
    pass

# Manual generation tracking
with generation(name="llm-call", model="gpt-4") as gen:
    response = call_llm()
    gen.update(output=response)

# Prompt management
from agenttrace import Prompt
prompt = await Prompt.get(name="my-prompt")
compiled = prompt.compile(name="Alice")

# Auto-instrumentation
from agenttrace.integrations.openai import OpenAIInstrumentation
OpenAIInstrumentation.enable()  # Automatically traces OpenAI calls
```

### TypeScript SDK

```typescript
import { AgentTrace, observe, startGeneration, Prompt } from 'agenttrace';

// Initialize client
const client = new AgentTrace({ apiKey: '...' });

// Function wrapper
const myFunction = observe(async () => {
  return 'result';
});

// Manual generation tracking
const trace = client.trace({ name: 'my-trace' });
const gen = trace.generation({ name: 'llm-call', model: 'gpt-4' });
gen.end({ output: 'response' });

// Prompt management
const prompt = await Prompt.get({ name: 'my-prompt' });
const compiled = prompt.compile({ name: 'Alice' });
```

### Go SDK

```go
import agenttrace "github.com/agenttrace/agenttrace/sdk/go"

// Initialize client
client := agenttrace.New(agenttrace.Config{APIKey: "..."})
defer client.Shutdown()

// Context-based tracing
ctx := context.Background()
trace := client.Trace(ctx, agenttrace.TraceOptions{Name: "my-trace"})

// Generation tracking
gen := trace.Generation(agenttrace.GenerationOptions{
    Name:  "llm-call",
    Model: "gpt-4",
})
gen.End(agenttrace.GenerationEndOptions{Output: "response"})

// Prompt management
prompt, _ := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name: "my-prompt",
})
compiled := prompt.Compile(map[string]any{"name": "Alice"})
```

### CLI Wrapper

```bash
# Install
go install github.com/agenttrace/agenttrace/sdk/cli@latest

# Wrap any command
agenttrace wrap --name "my-agent" -- python my_agent.py

# With git correlation
agenttrace wrap --name "coding-agent" --git-link -- npm run dev
```

## API Reference

### REST API

The full API covers 64 endpoints across these categories:

| Category | Endpoints | Description |
|----------|-----------|-------------|
| Health | `GET /health`, `/livez`, `/readyz`, `/version` | Service health checks |
| Auth | `POST /api/auth/login`, `register`, `refresh`, `logout` | Authentication |
| Ingestion | `POST /api/public/ingestion` | Batch event ingestion (Langfuse-compatible) |
| Traces | `GET/POST /api/public/traces`, `GET/DELETE/PATCH …/:traceId` | Trace CRUD + nested observations/scores |
| Observations | `POST /api/public/spans`, `generations`, `events` | Create spans, generations, events |
| Sessions | `GET /api/public/sessions`, `…/:sessionId` | Session listing and details |
| Scores | `GET/POST /api/public/scores`, batch, stats, `…/:scoreId` | Scoring and stats |
| Prompts | `GET/POST /api/public/prompts`, versions, labels, compile | Prompt management |
| Datasets | `GET/POST /api/public/datasets`, items, runs | Dataset and experiment management |
| Evaluators | `GET/POST /api/public/evaluators`, execute, templates | Evaluator configuration and execution |
| Checkpoints | `GET/POST /v1/checkpoints`, restore | Agent checkpoints |
| Git Links | `GET/POST /v1/git-links`, timeline | Git commit correlation |
| CI Runs | `GET/POST /v1/ci-runs`, `…/:ciRunId` | CI pipeline tracking |
| File Ops | `GET/POST /v1/file-operations` | File operation tracking |
| Terminal | `GET/POST /v1/terminal-commands` | Terminal command tracking |
| Orgs/Projects | `GET /api/v1/me`, organizations, projects, API keys | Account management |
| SSO | `/auth/sso/login`, callback, SAML, config | Enterprise SSO |
| Audit Logs | `GET /v1/organizations/:orgId/audit-logs`, export | Audit trail |
| Feedback | `POST /api/public/feedback` | User feedback |

📄 **Full OpenAPI specification**: [`api/docs/openapi.yaml`](api/docs/openapi.yaml)

### GraphQL API

```graphql
query GetTrace($id: ID!) {
  trace(id: $id) {
    id
    name
    observations {
      id
      type
      name
      startTime
      endTime
    }
  }
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| **Database** | | |
| `POSTGRES_USER` | PostgreSQL username | `agenttrace` |
| `POSTGRES_PASSWORD` | PostgreSQL password | - |
| `POSTGRES_DB` | PostgreSQL database | `agenttrace` |
| `CLICKHOUSE_USER` | ClickHouse username | `agenttrace` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | - |
| `CLICKHOUSE_DB` | ClickHouse database | `agenttrace` |
| `REDIS_PASSWORD` | Redis password | - |
| `MINIO_ROOT_USER` | MinIO root username | `agenttrace` |
| `MINIO_ROOT_PASSWORD` | MinIO root password | - |
| **Security** | | |
| `JWT_SECRET` | JWT signing secret (generate with `openssl rand -base64 32`) | - |
| `ENCRYPTION_KEY` | Encryption key for sensitive data at rest | - |
| `NEXTAUTH_URL` | Public URL for NextAuth (e.g., `https://your-domain.com`) | - |
| `NEXTAUTH_SECRET` | NextAuth session secret | - |
| **OAuth Providers** | | |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | - |
| **API** | | |
| `API_HOST` | API server bind address | `0.0.0.0` |
| `API_PORT` | API server port | `8080` |
| `NEXT_PUBLIC_API_URL` | Public API URL used by the web frontend | `http://api:8080` |
| **External Services** | | |
| `OPENAI_API_KEY` | OpenAI API key (required for LLM-as-Judge evaluators) | - |
| **Deployment** | | |
| `VERSION` | Docker image version tag | `latest` |
| `WEB_PORT` | Web frontend port | `3000` |
| `LOG_LEVEL` | API server log level (`debug`, `info`, `warn`, `error`) | `debug` |

See [`deploy/.env.example`](deploy/.env.example) for a copy-ready template.

## Development

### Prerequisites

- Go 1.24+
- Node.js 18+
- Docker & Docker Compose

> **Tip**: Run `make doctor` to verify all prerequisites are installed, then `make setup` for automated environment setup including database migrations and dependency installation.

### Local Development

```bash
# Start databases
docker compose -f deploy/docker-compose.dev.yml up -d

# Start API
cd api
go run cmd/server/main.go

# Start web (in another terminal)
cd web
npm install
npm run dev
```

### Running Tests

```bash
# Go backend
cd api && go test ./...

# TypeScript SDK
cd sdk/typescript && npm test

# Python SDK
cd sdk/python && pytest

# Go SDK
cd sdk/go && go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Langfuse](https://langfuse.com) - Inspiration for the API design
- [ClickHouse](https://clickhouse.com) - High-performance analytics database
- [Next.js](https://nextjs.org) - React framework for the frontend
