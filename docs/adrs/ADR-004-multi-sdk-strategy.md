# ADR-004: Multi-SDK Strategy with Idiomatic Patterns

## Status

Accepted

## Context

AgentTrace users work in diverse technology stacks:

- **Python**: ML/AI engineers, data scientists, most LLM frameworks (LangChain, LlamaIndex, CrewAI)
- **TypeScript**: Frontend developers, full-stack teams, browser-based agents
- **Go**: Backend engineers, infrastructure teams, CLI tool developers
- **CLI**: Autonomous agents (Claude Code, Aider, Cursor) that wrap arbitrary commands

Each language has different idioms for instrumentation:
- Python: Decorators, context managers
- TypeScript: Higher-order functions, async/await patterns
- Go: Context propagation, middleware patterns

A "lowest common denominator" API would feel unnatural in each language and reduce adoption.

### Alternatives Considered

1. **Single SDK (Python only)**
   - Pros: Easier to maintain, focused effort
   - Cons: Excludes TypeScript/Go users, limits adoption

2. **Universal API across all SDKs**
   - Pros: Consistent documentation, easier to switch languages
   - Cons: Feels unnatural, ignores language strengths

3. **Language-idiomatic SDKs** (chosen)
   - Pros: Natural feel in each language, leverages language strengths
   - Cons: Higher maintenance burden, potential feature drift

4. **OpenTelemetry SDK only**
   - Pros: Standard protocol, no custom SDKs
   - Cons: Missing LLM-specific features, complex setup

## Decision

We provide **four separate SDKs**, each following its language's idioms while maintaining feature parity:

### Python SDK (`agenttrace-python`)

```python
from agenttrace import observe, generation

# Decorator pattern (most Pythonic)
@observe(name="process-request")
def process_request(data):
    return analyze(data)

# Context manager pattern
with generation(name="llm-call", model="gpt-4") as gen:
    response = openai.chat.completions.create(...)
    gen.update(output=response, usage=response.usage)

# Auto-instrumentation
from agenttrace.integrations import openai as openai_integration
openai_integration.instrument()
```

### TypeScript SDK (`agenttrace-js`)

```typescript
import { observe, generation, AgentTrace } from 'agenttrace';

// Higher-order function pattern
const processRequest = observe({ name: 'process-request' }, async (data) => {
  return await analyze(data);
});

// Fluent builder pattern
const gen = generation({ name: 'llm-call', model: 'gpt-4' });
const response = await openai.chat.completions.create(...);
gen.end({ output: response, usage: response.usage });

// Callback pattern for compatibility
AgentTrace.trace('my-trace', async (trace) => {
  const span = trace.span('my-span');
  // ...
  span.end();
});
```

### Go SDK (`agenttrace-go`)

```go
import "github.com/agenttrace/agenttrace-go"

// Context-based pattern (idiomatic Go)
func ProcessRequest(ctx context.Context, data Data) error {
    ctx, span := agenttrace.StartSpan(ctx, "process-request")
    defer span.End()

    return analyze(ctx, data)
}

// Middleware pattern for HTTP
router.Use(agenttrace.Middleware())

// Generation helper
gen := agenttrace.StartGeneration(ctx, "llm-call", agenttrace.WithModel("gpt-4"))
response, err := client.CreateChatCompletion(ctx, req)
gen.End(agenttrace.WithOutput(response), agenttrace.WithUsage(response.Usage))
```

### CLI Wrapper (`agenttrace-cli`)

```bash
# Wrap any command with automatic tracing
agenttrace wrap -- python my_script.py

# With explicit trace naming
agenttrace wrap --name "data-pipeline" -- ./run_pipeline.sh

# Capture environment context
agenttrace wrap --capture-env --capture-git -- claude-code task.md
```

## Consequences

### Positive

- **Natural feel**: Each SDK feels like it belongs in its ecosystem
- **Adoption**: Lower friction for developers already familiar with language patterns
- **Leverage strengths**: Python decorators, Go contexts, TypeScript async/await
- **Auto-instrumentation**: Framework-specific integrations (LangChain, OpenAI, etc.)
- **CLI coverage**: Agents without SDK access can still be traced

### Negative

- **Maintenance burden**: Four codebases to maintain, test, and release
- **Feature drift risk**: Features may be implemented in one SDK before others
- **Documentation overhead**: Four sets of documentation, examples, and tutorials
- **Testing complexity**: Matrix of language versions × SDK versions × OS
- **Release coordination**: Synchronized releases across all SDKs

### Neutral

- Each SDK has its own versioning (semver)
- Bug fixes may need to be ported across SDKs
- Some features may be language-specific (e.g., Python decorators)

## Feature Parity Matrix

| Feature | Python | TypeScript | Go | CLI |
|---------|--------|------------|-----|-----|
| Trace creation | ✓ | ✓ | ✓ | ✓ |
| Span/Generation | ✓ | ✓ | ✓ | ✓ |
| Scores | ✓ | ✓ | ✓ | - |
| Git linking | ✓ | ✓ | ✓ | ✓ |
| Checkpoints | ✓ | ✓ | ✓ | - |
| File ops | ✓ | ✓ | ✓ | ✓ |
| Terminal cmds | ✓ | ✓ | ✓ | ✓ |
| Auto-instrument | ✓ | ✓ | - | - |
| Async support | ✓ | ✓ | ✓ | N/A |

## CI/CD Strategy

```yaml
# Each SDK tested independently in CI
jobs:
  python-sdk:
    strategy:
      matrix:
        python-version: ['3.9', '3.10', '3.11', '3.12']

  typescript-sdk:
    strategy:
      matrix:
        node-version: ['18', '20', '22']

  go-sdk:
    strategy:
      matrix:
        go-version: ['1.22', '1.23', '1.24']
```

## References

- [OpenTelemetry SDK Design](https://opentelemetry.io/docs/specs/otel/overview/)
- [Sentry SDK Strategy](https://docs.sentry.io/platforms/)
- [Langfuse SDK Design](https://langfuse.com/docs/sdk)
