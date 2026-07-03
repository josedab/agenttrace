# AgentTrace

[![CI](https://github.com/agenttrace/agenttrace/actions/workflows/ci.yml/badge.svg)](https://github.com/agenttrace/agenttrace/actions/workflows/ci.yml)
[![Security](https://github.com/agenttrace/agenttrace/actions/workflows/security.yml/badge.svg)](https://github.com/agenttrace/agenttrace/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/agenttrace/agenttrace/branch/main/graph/badge.svg)](https://codecov.io/gh/agenttrace/agenttrace)
[![Go Version](https://img.shields.io/github/go-mod/go-version/agenttrace/agenttrace?filename=api%2Fgo.mod)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Open-source outcome observability for AI coding agents. AgentTrace connects recorded runs to replay, evaluations, prompts, costs, git, and CI without inventing missing metrics.

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
- **Outcome Analytics**: Trace-to-commit-to-CI success, regression signals, and cost per successful outcome
- **Safe Replay Debugger**: Checkpoint-aware timeline inspection and deterministic generation replay without host code execution
- **Eval Hub**: Versioned private, organization, and public evaluation packages with provenance
- **Privacy Controls**: Redacted revocable links and enforceable no-egress mode

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
docker compose up -d --build
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

For an existing project, the CLI can detect the runtime and create a secret-free config:

```bash
agenttrace init
export AGENTTRACE_API_KEY="sk-at-..."
```

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

# Detect the project and create .agenttrace.yaml without secrets
agenttrace init

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
| Outcomes | `GET /api/public/outcomes`, digest and delivery routes | Trace-to-git-to-CI analytics |
| Replay | replay capabilities, plans, execution, comparison | Safe time-travel debugging |
| Eval Hub | package publish, fork, and run routes | Versioned evaluation assets |
| Sharing | expiring trace/replay share links | Server-redacted read-only views |
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
| `MINIO_ENDPOINT` | Optional external S3-compatible endpoint | - |
| `MINIO_ACCESS_KEY` | Optional object-storage access key | - |
| `MINIO_SECRET_KEY` | Optional object-storage secret key | - |
| **Security** | | |
| `JWT_SECRET` | JWT signing secret (generate with `openssl rand -base64 32`) | - |
| `NEXTAUTH_URL` | Public URL for NextAuth (e.g., `https://your-domain.com`) | - |
| `NEXTAUTH_SECRET` | NextAuth session secret | - |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins allowed by the API | - |
| **OAuth Providers** | | |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | - |
| **API** | | |
| `API_HOST` | API server bind address | `0.0.0.0` |
| `API_PORT` | API server port | `8080` |
| `NEXT_PUBLIC_API_URL` | Public API URL embedded into the web image at build time | - |
| **External Services** | | |
| `OPENAI_API_KEY` | OpenAI API key (required for LLM-as-Judge evaluators) | - |
| **Deployment** | | |
| `VERSION` | Immutable API and web image version tag | - |
| `WEB_PORT` | Web frontend port | `3000` |
| `LOG_LEVEL` | API server log level (`debug`, `info`, `warn`, `error`) | `debug` |

See [`deploy/.env.example`](deploy/.env.example) for a copy-ready template.

## Development

### Prerequisites

- Go 1.25.12+
- Node.js 20+
- Docker with Docker Compose v2
- Make

Python is only required when working on the Python SDK.

### Choose the Smallest Workflow

| Workflow | Deployed services | Commands |
|----------|-------------------|----------|
| Unit and component tests | None | `make test` |
| Core application development | PostgreSQL, ClickHouse | `make setup && make dev` |
| Workers, async exports, distributed rate limiting | PostgreSQL, ClickHouse, Redis, MinIO | `make setup-full && make dev-full` |

### Local Development

```bash
# Verify local tools
make doctor

# First-time minimal setup
make setup

# Start API and web
make dev
```

The minimal workflow intentionally disables Redis-backed job queues, distributed rate limiting, and MinIO-backed exports. PostgreSQL and ClickHouse remain required because they are the application's primary metadata and trace stores.

Use `make dev-api` or `make dev-web` to run one application component. See the [local development guide](docs/docs/getting-started/development.md) for full-stack and devcontainer workflows.

### Running Tests

```bash
# All default test suites; no database or cache services required
make test

# Individual components
make test-api
make test-web
make test-sdk-ts
make test-sdk-py
make test-sdk-go
```

Database integration and end-to-end tests are opt-in and may start or require their documented services.

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
