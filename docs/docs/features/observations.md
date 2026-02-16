---
sidebar_position: 2
title: Observations
description: Learn about the individual units within a trace — spans, generations, and events — and how to create them with the AgentTrace SDK.
---

# Observations

Observations are the individual units of work within a [trace](./tracing.md). Each trace contains one or more observations that represent discrete operations, LLM calls, or notable events during execution.

## Observation Types

AgentTrace supports three observation types:

| Type | Purpose | Key Fields |
|------|---------|------------|
| **SPAN** | Generic operation (retrieval, processing, tool call) | `input`, `output`, `metadata` |
| **GENERATION** | LLM call with model details | `model`, `modelParameters`, `usage` (tokens & cost) |
| **EVENT** | Point-in-time marker (no duration) | `input`, `metadata`, `level` |

## Creating Observations

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

### Spans

Spans represent any generic operation — a database query, API call, retrieval step, or processing block.

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace

at = AgentTrace()

with at.trace("my-task") as trace:
    # Create a span for a retrieval step
    span = trace.span(
        name="document-retrieval",
        input={"query": "How does authentication work?"},
        metadata={"source": "vector-db"}
    )
    results = retrieve_documents("How does authentication work?")
    span.end(output={"documents": results})
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';

const at = new AgentTrace();

const trace = at.startTrace('my-task');
const span = trace.span({
  name: 'document-retrieval',
  input: { query: 'How does authentication work?' },
  metadata: { source: 'vector-db' },
});
const results = await retrieveDocuments('How does authentication work?');
span.end({ output: { documents: results } });
```

</TabItem>
</Tabs>

### Generations

Generations track LLM calls with model-specific metadata, token usage, and cost.

<Tabs>
<TabItem value="python" label="Python" default>

```python
generation = trace.generation(
    name="summarize-docs",
    model="gpt-4",
    model_parameters={"temperature": 0.3, "max_tokens": 500},
    input={"messages": [{"role": "user", "content": "Summarize these docs."}]},
)
response = call_llm(...)
generation.end(
    output=response.text,
    usage={"input_tokens": 320, "output_tokens": 150}
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
const generation = trace.generation({
  name: 'summarize-docs',
  model: 'gpt-4',
  modelParameters: { temperature: 0.3, maxTokens: 500 },
  input: { messages: [{ role: 'user', content: 'Summarize these docs.' }] },
});
const response = await callLLM(/* ... */);
generation.end({
  output: response.text,
  usage: { inputTokens: 320, outputTokens: 150 },
});
```

</TabItem>
</Tabs>

### Events

Events are point-in-time markers with no duration — useful for logging decisions, errors, or state changes.

<Tabs>
<TabItem value="python" label="Python" default>

```python
trace.event(
    name="cache-miss",
    input={"key": "user-profile-42"},
    metadata={"cache": "redis"},
    level="WARNING"
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
trace.event({
  name: 'cache-miss',
  input: { key: 'user-profile-42' },
  metadata: { cache: 'redis' },
  level: 'WARNING',
});
```

</TabItem>
</Tabs>

## Nesting Observations

Observations can be nested to capture parent-child relationships. A span can contain generations, and generations can contain child spans.

```python
with at.trace("rag-pipeline") as trace:
    retrieval = trace.span(name="retrieval")
    docs = fetch_docs()
    retrieval.end(output=docs)

    # Nest a generation under the trace
    generation = trace.generation(name="answer", model="gpt-4", input=docs)
    answer = call_llm(docs)
    generation.end(output=answer, usage={"input_tokens": 800, "output_tokens": 200})
```

## Observation Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | Auto-generated unique identifier |
| `name` | string | Human-readable label |
| `startTime` | datetime | When the observation began |
| `endTime` | datetime | When the observation ended (spans/generations only) |
| `input` | any | Input data |
| `output` | any | Output/result data |
| `metadata` | object | Arbitrary key-value metadata |
| `level` | string | `DEBUG`, `DEFAULT`, `WARNING`, `ERROR` |
| `parentObservationId` | string | ID of the parent observation for nesting |

## Viewing Observations

In the dashboard, click any trace to open the **Trace Detail** view. Observations are displayed in a timeline waterfall, showing nesting, duration, and type at a glance.
