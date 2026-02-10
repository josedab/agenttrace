# AgentTrace Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/agenttrace/agenttrace/sdk/go.svg)](https://pkg.go.dev/github.com/agenttrace/agenttrace/sdk/go)

Go SDK for [AgentTrace](https://github.com/agenttrace/agenttrace) — observability for AI coding agents.

## Installation

```bash
go get github.com/agenttrace/agenttrace/sdk/go
```

## Quick Start

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

## Features

- **Context-based tracing** — Idiomatic Go context propagation
- **Thread-safe** — Mutex-protected queue with concurrent access support
- **Batch processing** — Automatic event batching with configurable thresholds
- **Race-condition tested** — Extensive concurrent testing with `-race` flag
- **Benchmark suite** — Performance regression detection

## Documentation

See the [full documentation](https://docs.agenttrace.io/sdks/go).

## License

MIT
