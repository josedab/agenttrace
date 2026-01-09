---
sidebar_position: 1
---

# Tracing

Traces are the foundation of AgentTrace observability. They capture the complete lifecycle of an AI agent task.

## Creating Traces

### Using the SDK

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace, observe

at = AgentTrace()

# Using decorator
@observe()
def my_task(prompt: str) -> str:
    return process(prompt)

# Using context manager
with at.trace("my-task") as trace:
    trace.input = {"prompt": "Hello"}
    result = my_task("Hello")
    trace.output = result
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';

const at = new AgentTrace();

// Using wrapper
const myTask = at.observe(async (prompt: string) => {
    return await process(prompt);
}, { name: 'my-task' });

// Using manual trace
const trace = at.startTrace('my-task', { input: { prompt: 'Hello' } });
const result = await myTask('Hello');
trace.end({ output: result });
```

</TabItem>
</Tabs>

## Trace Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | Unique trace identifier |
| `name` | string | Human-readable name |
| `input` | any | Input data |
| `output` | any | Output/result data |
| `metadata` | object | Custom metadata |
| `tags` | string[] | Filterable labels |
| `level` | string | Log level |
| `sessionId` | string | Session grouping |
| `userId` | string | User identifier |
| `latency` | number | Duration in ms |
| `totalCost` | number | Aggregated cost |

## Viewing Traces

In the dashboard:

1. Navigate to **Traces**
2. Use filters to find specific traces
3. Click a trace to see details

### Trace Detail View

The trace detail shows:
- Input/output data
- Observation timeline
- Latency breakdown
- Cost summary
- Linked scores
- Related sessions

## Filtering Traces

Filter by:
- **Name**: Exact or partial match
- **Tags**: One or more tags
- **Session**: Related traces
- **Time range**: Custom date range
- **Level**: DEBUG, INFO, WARNING, ERROR
- **User**: Who initiated

## Best Practices

1. **Name traces descriptively**: Use names like `code-review`, `chat-response`
2. **Add relevant metadata**: Include context like repository, branch, user
3. **Use tags for filtering**: Tag with environment, feature, etc.
4. **Group with sessions**: Connect related traces
5. **Set appropriate levels**: Use ERROR for failures, DEBUG for development
