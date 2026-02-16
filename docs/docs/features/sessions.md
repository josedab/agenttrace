---
sidebar_position: 3
title: Sessions
description: Group related traces into sessions to track multi-turn conversations, workflows, and user journeys in AgentTrace.
---

# Sessions

Sessions group related [traces](./tracing.md) together — for example, all turns in a chat conversation or steps in a multi-stage workflow. This lets you analyze user journeys end-to-end rather than inspecting individual requests in isolation.

## How Sessions Work

A session is created implicitly when you pass a `session_id` to a trace. All traces that share the same `session_id` are grouped together in the dashboard.

- Sessions are **not** created explicitly — they emerge from traces that reference the same ID.
- A single session can contain any number of traces.
- Session IDs are arbitrary strings you control (UUIDs, conversation IDs, ticket numbers, etc.).

## Creating Traces with Sessions

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace
import uuid

at = AgentTrace()

# Start a session for a chat conversation
session_id = str(uuid.uuid4())

# First turn
with at.trace("chat-turn", session_id=session_id, user_id="user-42") as trace:
    trace.input = {"message": "What is AgentTrace?"}
    response = generate_response(trace.input["message"])
    trace.output = response

# Second turn — same session
with at.trace("chat-turn", session_id=session_id, user_id="user-42") as trace:
    trace.input = {"message": "How do I set it up?"}
    response = generate_response(trace.input["message"])
    trace.output = response
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';
import { randomUUID } from 'crypto';

const at = new AgentTrace();

// Start a session for a chat conversation
const sessionId = randomUUID();

// First turn
const trace1 = at.startTrace('chat-turn', {
  sessionId,
  userId: 'user-42',
  input: { message: 'What is AgentTrace?' },
});
const response1 = await generateResponse('What is AgentTrace?');
trace1.end({ output: response1 });

// Second turn — same session
const trace2 = at.startTrace('chat-turn', {
  sessionId,
  userId: 'user-42',
  input: { message: 'How do I set it up?' },
});
const response2 = await generateResponse('How do I set it up?');
trace2.end({ output: response2 });
```

</TabItem>
</Tabs>

## Updating the Session ID

You can assign or change the session ID after a trace is created:

```python
trace = at.trace("my-task")
# ... later, once the session context is known
trace.update(session_id="session-abc-123")
```

## Session Analytics in the Dashboard

Navigate to **Sessions** in the sidebar to access the session list. Each session row shows:

| Metric | Description |
|--------|-------------|
| **Trace count** | Number of traces in the session |
| **Total duration** | Time from first trace start to last trace end |
| **Total cost** | Aggregated cost across all traces |
| **Total tokens** | Sum of input + output tokens |
| **User** | The `userId` associated with the session |

### Session Detail View

Click a session to see:

1. **Trace timeline** — All traces displayed chronologically, showing how the conversation or workflow progressed.
2. **I/O pairs** — Input and output for each trace, making it easy to follow multi-turn exchanges.
3. **Scores** — Any [scores](./scores.md) attached to traces within the session.
4. **Cost & latency** — Per-trace and aggregate cost and latency breakdowns.

## Common Session Patterns

### Chat conversations

Group every user message → assistant response cycle under one session:

```python
session_id = f"chat-{user_id}-{conversation_id}"
```

### Multi-step agent workflows

Track an agent that performs research, planning, and execution as a single session:

```python
session_id = f"agent-run-{task_id}"
```

### A/B testing

Use session metadata to compare variants:

```python
with at.trace("chat-turn", session_id=session_id, metadata={"variant": "B"}) as trace:
    ...
```

## Best Practices

1. **Use stable, meaningful IDs** — derive session IDs from your domain (conversation ID, ticket number) rather than random UUIDs when possible.
2. **Include `user_id`** — this enables per-user session analytics.
3. **Keep sessions focused** — one session per logical user journey. Avoid catch-all sessions.
4. **Add metadata** — tag sessions with environment, feature flag, or experiment variant for filtering.
