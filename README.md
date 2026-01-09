# AgentTrace

Open-source observability platform for AI coding agents. LLM traces, evaluations, prompt management, and dataset experiments with Langfuse feature parity.

## Features

- **Trace Exploration**: Real-time trace visualization with parent-child relationships, latency waterfall, and agent graph views
- **Prompt Management**: Version-controlled prompts with A/B testing, playground, and SDK integration
- **Evaluators & Scores**: Built-in evaluators (LLM-as-Judge, regex, JSON schema) and custom scoring
- **Datasets & Experiments**: Create datasets, run experiments, and compare results
- **Multi-language SDKs**: Python, TypeScript, Go SDKs with auto-instrumentation
- **CLI Wrapper**: Wrap any CLI tool for automatic tracing
- **Cost Tracking**: Automatic cost calculation for 400+ LLM models

## Quick Start

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
└── deploy/               # Deployment configs
    ├── docker-compose.yml
    └── .env.example
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

- `POST /api/public/ingestion` - Batch event ingestion
- `GET /api/public/traces` - List traces
- `GET /api/public/traces/:id` - Get trace by ID
- `GET /api/public/prompts` - Get prompt by name
- `POST /api/public/scores` - Create score

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
| `POSTGRES_USER` | PostgreSQL username | `agenttrace` |
| `POSTGRES_PASSWORD` | PostgreSQL password | - |
| `POSTGRES_DB` | PostgreSQL database | `agenttrace` |
| `CLICKHOUSE_USER` | ClickHouse username | `default` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | - |
| `REDIS_PASSWORD` | Redis password | - |
| `JWT_SECRET` | JWT signing secret | - |
| `NEXTAUTH_SECRET` | NextAuth secret | - |
| `NEXTAUTH_URL` | NextAuth URL | - |

See `deploy/.env.example` for full list.

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

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
