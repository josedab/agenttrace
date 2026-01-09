---
sidebar_position: 1
---

# SDKs Overview

AgentTrace provides official SDKs for Python, TypeScript, Go, and a CLI wrapper for any language.

## Quick Comparison

| Feature | Python | TypeScript | Go | CLI |
|---------|--------|------------|-----|-----|
| Automatic tracing | ✅ `@observe` | ✅ `observe()` | ✅ Context-based | ✅ Wrapper |
| Async support | ✅ | ✅ | ✅ | N/A |
| Batching | ✅ Auto | ✅ Auto | ✅ Auto | ✅ |
| Prompt fetching | ✅ | ✅ | ✅ | ✅ |
| LLM integrations | ✅ OpenAI, Anthropic | ✅ OpenAI, Anthropic | Manual | Auto-detect |
| Checkpoints | ✅ | ✅ | ✅ | ✅ Auto |
| Git linking | ✅ | ✅ | ✅ | ✅ Auto |

## Installation

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="python" label="Python" default>

```bash
pip install agenttrace
```

**Requirements**: Python 3.9+

</TabItem>
<TabItem value="typescript" label="TypeScript">

```bash
npm install agenttrace
# or
yarn add agenttrace
# or
pnpm add agenttrace
```

**Requirements**: Node.js 18+

</TabItem>
<TabItem value="go" label="Go">

```bash
go get github.com/agenttrace/agenttrace/sdk/go
```

**Requirements**: Go 1.21+

</TabItem>
<TabItem value="cli" label="CLI">

```bash
# Install via script
curl -sSL https://get.agenttrace.io/cli | sh

# Or via Go
go install github.com/agenttrace/agenttrace/sdk/cli@latest

# Or via npm (global)
npm install -g @agenttrace/cli
```

</TabItem>
</Tabs>

## Configuration

All SDKs can be configured via environment variables:

```bash
export AGENTTRACE_API_KEY="sk-at-..."
export AGENTTRACE_PROJECT_ID="your-project-id"  # optional
export AGENTTRACE_API_URL="https://api.agenttrace.io"  # for self-hosted
export AGENTTRACE_DEBUG="true"  # enable debug logging
```

Or via code:

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace

at = AgentTrace(
    api_key="sk-at-...",
    project_id="your-project-id",
    api_url="https://api.agenttrace.io",  # optional
    debug=True,  # optional
    flush_interval=5.0,  # seconds, default 5
    batch_size=100,  # default 100
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';

const at = new AgentTrace({
    apiKey: 'sk-at-...',
    projectId: 'your-project-id',
    apiUrl: 'https://api.agenttrace.io',  // optional
    debug: true,  // optional
    flushInterval: 5000,  // ms, default 5000
    batchSize: 100,  // default 100
});
```

</TabItem>
<TabItem value="go" label="Go">

```go
import "github.com/agenttrace/agenttrace/sdk/go"

at := agenttrace.New(
    agenttrace.WithAPIKey("sk-at-..."),
    agenttrace.WithProjectID("your-project-id"),
    agenttrace.WithAPIURL("https://api.agenttrace.io"),
    agenttrace.WithDebug(true),
    agenttrace.WithFlushInterval(5 * time.Second),
    agenttrace.WithBatchSize(100),
)
```

</TabItem>
</Tabs>

## Core Concepts

### Traces

A trace represents a complete unit of work, like a single agent task or user request.

<Tabs>
<TabItem value="python" label="Python" default>

```python
# Using decorator
@observe()
def my_task(input: str) -> str:
    return process(input)

# Using context manager
with at.trace("my-task", input={"query": "hello"}) as trace:
    result = my_task("hello")
    trace.output = result
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
// Using wrapper
const myTask = at.observe(async (input: string) => {
    return await process(input);
}, { name: 'my-task' });

// Using manual trace
const trace = at.startTrace('my-task', { input: { query: 'hello' } });
const result = await myTask('hello');
trace.end({ output: result });
```

</TabItem>
<TabItem value="go" label="Go">

```go
// Start a trace
trace, ctx := at.StartTrace(ctx, "my-task", agenttrace.WithInput(map[string]any{"query": "hello"}))
defer trace.End()

result := myTask(ctx, "hello")
trace.SetOutput(result)
```

</TabItem>
</Tabs>

### Observations

Observations are spans within a trace (LLM calls, tool uses, etc.)

<Tabs>
<TabItem value="python" label="Python" default>

```python
with at.trace("my-task") as trace:
    # Generation observation
    with trace.generation(
        name="llm-call",
        model="claude-3-sonnet",
        input=[{"role": "user", "content": "Hello"}]
    ) as gen:
        response = call_llm(...)
        gen.output = response
        gen.usage = {"inputTokens": 10, "outputTokens": 20}

    # Span observation
    with trace.span(name="process-result") as span:
        processed = process(response)
        span.output = processed
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
const trace = at.startTrace('my-task');

// Generation observation
const gen = trace.startGeneration({
    name: 'llm-call',
    model: 'claude-3-sonnet',
    input: [{ role: 'user', content: 'Hello' }]
});
const response = await callLLM();
gen.end({ output: response, usage: { inputTokens: 10, outputTokens: 20 } });

// Span observation
const span = trace.startSpan({ name: 'process-result' });
const processed = await process(response);
span.end({ output: processed });

trace.end();
```

</TabItem>
</Tabs>

### Prompts

Fetch and use versioned prompts:

<Tabs>
<TabItem value="python" label="Python" default>

```python
# Get latest production prompt
prompt = at.get_prompt("code-review", label="production")

# Compile with variables
compiled = prompt.compile(
    language="Python",
    code="def hello(): print('world')"
)

# Use in a trace
with at.trace("review") as trace:
    trace.prompt = prompt  # Links prompt version to trace
    response = call_llm(compiled)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
// Get latest production prompt
const prompt = await at.getPrompt('code-review', { label: 'production' });

// Compile with variables
const compiled = prompt.compile({
    language: 'Python',
    code: "def hello(): print('world')"
});

// Use in a trace
const trace = at.startTrace('review', { prompt });
const response = await callLLM(compiled);
trace.end({ output: response });
```

</TabItem>
</Tabs>

### Scores

Score trace outputs:

<Tabs>
<TabItem value="python" label="Python" default>

```python
# Score a trace
at.score(
    trace_id="trace-123",
    name="quality",
    value=0.95,
    comment="Excellent code review"
)

# Or score within a trace context
with at.trace("my-task") as trace:
    result = process()
    trace.score(name="quality", value=0.95)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
// Score a trace
await at.score({
    traceId: 'trace-123',
    name: 'quality',
    value: 0.95,
    comment: 'Excellent code review'
});

// Or score within a trace
const trace = at.startTrace('my-task');
const result = await process();
trace.score({ name: 'quality', value: 0.95 });
trace.end();
```

</TabItem>
</Tabs>

## Next Steps

- [Python SDK Reference](/sdks/python) - Full Python SDK documentation
- [TypeScript SDK Reference](/sdks/typescript) - Full TypeScript SDK documentation
- [Go SDK Reference](/sdks/go) - Full Go SDK documentation
- [CLI Reference](/sdks/cli) - CLI wrapper documentation
