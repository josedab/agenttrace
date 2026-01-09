# ADR-001: Go + Fiber Backend Architecture

## Status

Accepted

## Context

AgentTrace is an observability platform for AI coding agents that must handle high-throughput trace ingestion from potentially thousands of concurrent agent sessions. The platform needs to:

1. Ingest traces with minimal latency to avoid impacting agent performance
2. Handle burst traffic during peak development hours
3. Maintain low memory footprint for cost-effective deployment
4. Provide simple deployment for self-hosted installations

The existing landscape includes Langfuse (Node.js/tRPC) and LangSmith (Python/FastAPI). We needed to choose a backend technology stack that would differentiate AgentTrace on performance while maintaining developer productivity.

### Alternatives Considered

1. **Node.js + tRPC** (like Langfuse)
   - Pros: Large ecosystem, TypeScript throughout, team familiarity
   - Cons: Single-threaded, higher memory per request, GC pauses

2. **Python + FastAPI** (like LangSmith)
   - Pros: ML ecosystem alignment, easy prototyping
   - Cons: GIL limitations, higher latency, complex deployment (ASGI servers)

3. **Rust + Axum**
   - Pros: Maximum performance, memory safety
   - Cons: Slower development velocity, smaller talent pool

4. **Go + Fiber** (chosen)
   - Pros: Excellent concurrency, low memory, single binary, fast compilation
   - Cons: Smaller web ecosystem, less frontend developer familiarity

## Decision

We chose **Go 1.24 with the Fiber web framework** as the backend stack for AgentTrace.

Key factors:
- **10x throughput improvement** over Node.js for trace ingestion workloads
- **Native concurrency** via goroutines handles thousands of concurrent connections efficiently
- **Single binary deployment** simplifies Docker images and self-hosted installations
- **Low memory footprint** (~10-50MB per instance vs ~100-500MB for Node.js)
- **Fiber framework** provides Express-like ergonomics for Go, reducing learning curve

## Consequences

### Positive

- **Performance**: Benchmarks show 10x throughput improvement over equivalent Node.js implementations for trace ingestion
- **Operational simplicity**: Single statically-linked binary with no runtime dependencies
- **Cost efficiency**: Lower memory usage allows running more instances per host
- **Reliability**: No GC pauses during high-throughput periods
- **Fast builds**: Go compilation is fast, enabling quick CI/CD cycles

### Negative

- **Ecosystem gap**: Fewer ready-made libraries compared to Node.js/Python
- **Team composition**: Requires Go expertise, which may limit contributor pool
- **Full-stack friction**: Frontend developers (TypeScript) need to context-switch to Go for backend changes
- **Fiber-specific patterns**: Some Fiber middleware differs from standard `net/http`, requiring adaptation

### Neutral

- GraphQL support via gqlgen works well but requires code generation
- Error handling patterns differ from exception-based languages
- Testing patterns require learning Go-specific tooling (testify, gomock)

## References

- [Fiber Framework](https://gofiber.io/)
- [Go vs Node.js Performance Benchmarks](https://www.techempower.com/benchmarks/)
- [Langfuse Architecture](https://langfuse.com/docs/deployment/self-host)
