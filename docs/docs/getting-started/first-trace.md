---
sidebar_position: 3
title: Create Your First Trace
description: End-to-end tutorial for creating your first trace with the Python, TypeScript, and Go SDKs.
---

# Create Your First Trace

This tutorial walks you through instrumenting an AI agent call and viewing the resulting trace in the AgentTrace dashboard. By the end you'll have a working traced function in your language of choice.

## Prerequisites

- A running AgentTrace instance ([Installation Guide](/getting-started/installation)) or a [cloud account](https://app.agenttrace.io)
- An API key (go to **Settings → API Keys** in the dashboard)

## Step 1: Install the SDK

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="python" label="Python" default>

```bash
pip install agenttrace
```

Requires Python 3.9+.

</TabItem>
<TabItem value="typescript" label="TypeScript">

```bash
npm install agenttrace
# or: yarn add agenttrace / pnpm add agenttrace
```

Requires Node.js 18+.

</TabItem>
<TabItem value="go" label="Go">

```bash
go get github.com/agenttrace/agenttrace-go
```

Requires Go 1.25+.

</TabItem>
</Tabs>

## Step 2: Configure the Client

Set your API key as an environment variable so it stays out of source code:

```bash
export AGENTTRACE_API_KEY="sk-at-..."
# For self-hosted instances, also set:
export AGENTTRACE_API_URL="http://localhost:8080"
```

Then initialize the client in your application:

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace

at = AgentTrace(
    # Reads AGENTTRACE_API_KEY from env automatically
    project_id="my-first-project",  # optional
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from "agenttrace";

const at = new AgentTrace({
  // Reads AGENTTRACE_API_KEY from env automatically
  projectId: "my-first-project", // optional
});
```

</TabItem>
<TabItem value="go" label="Go">

```go
package main

import (
    agenttrace "github.com/agenttrace/agenttrace-go"
)

func main() {
    client := agenttrace.New(agenttrace.Config{
        // Reads AGENTTRACE_API_KEY from env automatically
        ProjectID: "my-first-project", // optional
    })
    defer client.Shutdown()
}
```

</TabItem>
</Tabs>

## Step 3: Create a Traced Function

<Tabs>
<TabItem value="python" label="Python" default>

The `@observe` decorator is the simplest way to trace a function. It automatically captures inputs, outputs, timing, and errors.

```python
from agenttrace import AgentTrace, observe

at = AgentTrace()

@observe()
def summarize(text: str) -> str:
    """Summarize the input text using an LLM."""
    # Replace with your actual LLM call
    return f"Summary of: {text[:50]}..."

# Run the function — a trace is created automatically
result = summarize("AgentTrace is an open-source observability platform for LLM applications.")
print(result)

# Flush to ensure the trace is sent before the process exits
at.flush()
```

For more control, use the context manager to add metadata, tags, or nested spans:

```python
with at.trace("summarize-task") as trace:
    trace.input = {"text": "Hello, AgentTrace!"}
    trace.tags = ["tutorial", "getting-started"]

    with trace.generation(name="llm-call", model="gpt-4") as gen:
        gen.input = [{"role": "user", "content": "Summarize: Hello, AgentTrace!"}]
        response = "A greeting to AgentTrace."  # your LLM call here
        gen.output = response
        gen.usage = {"inputTokens": 12, "outputTokens": 8}

    trace.output = {"summary": response}
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

Use the `observe` wrapper to automatically trace any async function:

```typescript
import { AgentTrace, observe } from "agenttrace";

const at = new AgentTrace();

const summarize = observe(
  async (text: string): Promise<string> => {
    // Replace with your actual LLM call
    return `Summary of: ${text.slice(0, 50)}...`;
  },
  { name: "summarize" }
);

// Run the function — a trace is created automatically
const result = await summarize(
  "AgentTrace is an open-source observability platform for LLM applications."
);
console.log(result);

// Flush to ensure the trace is sent before the process exits
await at.flush();
```

For more control, manage traces explicitly:

```typescript
const trace = at.startTrace("summarize-task");
trace.input = { text: "Hello, AgentTrace!" };
trace.tags = ["tutorial", "getting-started"];

const gen = trace.generation({
  name: "llm-call",
  model: "gpt-4",
  input: [{ role: "user", content: "Summarize: Hello, AgentTrace!" }],
});
const response = "A greeting to AgentTrace."; // your LLM call here
gen.end({
  output: response,
  usage: { inputTokens: 12, outputTokens: 8 },
});

trace.end({ output: { summary: response } });
await at.flush();
```

</TabItem>
<TabItem value="go" label="Go">

In Go, use `client.Trace` and pass the context through your call chain:

```go
package main

import (
    "context"
    "fmt"

    agenttrace "github.com/agenttrace/agenttrace-go"
)

func main() {
    client := agenttrace.New(agenttrace.Config{})
    defer client.Shutdown()

    ctx := context.Background()

    // Create a trace
    trace := client.Trace(ctx, agenttrace.TraceOptions{
        Name:  "summarize-task",
        Input: map[string]any{"text": "Hello, AgentTrace!"},
        Tags:  []string{"tutorial", "getting-started"},
    })

    // Add a generation (LLM call)
    gen := trace.Generation(agenttrace.GenerationOptions{
        Name:  "llm-call",
        Model: "gpt-4",
        Input: map[string]any{"query": "Summarize: Hello, AgentTrace!"},
    })

    // Your LLM call here
    response := "A greeting to AgentTrace."

    gen.End(agenttrace.GenerationEndOptions{
        Output: response,
        Usage:  agenttrace.Usage{InputTokens: 12, OutputTokens: 8},
    })

    trace.End(agenttrace.TraceEndOptions{
        Output: map[string]any{"summary": response},
    })

    fmt.Println(response)
}
```

</TabItem>
</Tabs>

## Step 4: View the Trace in the Dashboard

1. Open the AgentTrace dashboard (default: [http://localhost:3000](http://localhost:3000))
2. Navigate to **Traces** in the sidebar
3. Click on your trace to inspect it

You'll see:

- **Timeline** — A waterfall view showing each operation and its duration
- **Input / Output** — The data passed into and returned from your function
- **Generation details** — Model name, token usage, and estimated cost
- **Metadata & Tags** — Custom attributes you attached to the trace

## Step 5: Add More Detail

Once basic tracing works, enrich your traces with spans and scores:

<Tabs>
<TabItem value="python" label="Python" default>

```python
with at.trace("code-review") as trace:
    # Span: fetch context from a database
    with trace.span(name="fetch-context") as span:
        context_docs = ["doc1", "doc2"]  # your retrieval logic
        span.output = {"count": len(context_docs)}

    # Generation: call the LLM
    with trace.generation(name="review", model="claude-3-sonnet") as gen:
        gen.input = [{"role": "user", "content": "Review this code..."}]
        review = "Looks good!"
        gen.output = review
        gen.usage = {"inputTokens": 200, "outputTokens": 50}

    trace.output = {"review": review}

    # Score the result
    trace.score(name="quality", value=0.9, comment="Accurate review")
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
const trace = at.startTrace("code-review");

const span = trace.span({ name: "fetch-context" });
const contextDocs = ["doc1", "doc2"]; // your retrieval logic
span.end({ output: { count: contextDocs.length } });

const gen = trace.generation({
  name: "review",
  model: "claude-3-sonnet",
  input: [{ role: "user", content: "Review this code..." }],
});
const review = "Looks good!";
gen.end({ output: review, usage: { inputTokens: 200, outputTokens: 50 } });

trace.end({ output: { review } });
trace.score({ name: "quality", value: 0.9, comment: "Accurate review" });
await at.flush();
```

</TabItem>
<TabItem value="go" label="Go">

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{Name: "code-review"})

span := trace.Span(agenttrace.SpanOptions{Name: "fetch-context"})
contextDocs := []string{"doc1", "doc2"}
span.End(agenttrace.SpanEndOptions{
    Output: map[string]any{"count": len(contextDocs)},
})

gen := trace.Generation(agenttrace.GenerationOptions{
    Name:  "review",
    Model: "claude-3-sonnet",
    Input: map[string]any{"query": "Review this code..."},
})
review := "Looks good!"
gen.End(agenttrace.GenerationEndOptions{
    Output: review,
    Usage:  agenttrace.Usage{InputTokens: 200, OutputTokens: 50},
})

trace.End(agenttrace.TraceEndOptions{
    Output: map[string]any{"review": review},
})
```

</TabItem>
</Tabs>

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Trace not appearing | Ensure you call `flush()` / `Shutdown()` before the process exits |
| `401 Unauthorized` | Verify your API key is correct and has not expired |
| Connection refused | Check that `AGENTTRACE_API_URL` points to your running instance |
| Debug logging | Set `AGENTTRACE_DEBUG=true` in your environment |

## What's Next?

- [Core Concepts](/getting-started/concepts) — Understand traces, spans, generations, and sessions
- [Python SDK Reference](/sdks/python) — Full API details for the Python SDK
- [TypeScript SDK Reference](/sdks/typescript) — Full API details for the TypeScript SDK
- [Go SDK Reference](/sdks/go) — Full API details for the Go SDK
- [Prompt Management](/prompts/overview) — Version and deploy prompts from the dashboard
