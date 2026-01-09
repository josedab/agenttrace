---
sidebar_position: 1
---

# Quickstart

Get AgentTrace up and running in under 5 minutes.

## Prerequisites

- An AgentTrace account (or self-hosted instance)
- Python 3.9+, Node.js 18+, or Go 1.21+

## Step 1: Get Your API Key

1. Log in to your [AgentTrace dashboard](https://app.agenttrace.io)
2. Go to **Settings > API Keys**
3. Click **Create API Key**
4. Copy your key (it starts with `sk-at-`)

## Step 2: Install the SDK

Choose your preferred language:

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="python" label="Python" default>

```bash
pip install agenttrace
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```bash
npm install agenttrace
# or
yarn add agenttrace
# or
pnpm add agenttrace
```

</TabItem>
<TabItem value="go" label="Go">

```bash
go get github.com/agenttrace/agenttrace/sdk/go
```

</TabItem>
<TabItem value="cli" label="CLI">

```bash
curl -sSL https://get.agenttrace.io/cli | sh
```

</TabItem>
</Tabs>

## Step 3: Initialize AgentTrace

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace

# Initialize with your API key
at = AgentTrace(
    api_key="sk-at-...",  # or set AGENTTRACE_API_KEY env var
    project_id="your-project-id"  # optional, uses default project
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';

const at = new AgentTrace({
    apiKey: 'sk-at-...',  // or set AGENTTRACE_API_KEY env var
    projectId: 'your-project-id'  // optional
});
```

</TabItem>
<TabItem value="go" label="Go">

```go
import "github.com/agenttrace/agenttrace/sdk/go"

at := agenttrace.New(
    agenttrace.WithAPIKey("sk-at-..."),
    agenttrace.WithProjectID("your-project-id"),
)
```

</TabItem>
</Tabs>

## Step 4: Create Your First Trace

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace, observe

at = AgentTrace()

@observe()
def my_agent_task(prompt: str) -> str:
    # Your agent logic here
    response = call_llm(prompt)
    return response

# Or use the context manager
with at.trace("my-task") as trace:
    # Automatically captures timing, inputs, outputs
    result = my_agent_task("Write a function to sort a list")
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';

const at = new AgentTrace();

// Using the observe wrapper
const myAgentTask = at.observe(async (prompt: string) => {
    const response = await callLLM(prompt);
    return response;
}, { name: 'my-agent-task' });

// Or using trace directly
const trace = at.startTrace('my-task');
try {
    const result = await myAgentTask('Write a function to sort a list');
    trace.end({ output: result });
} catch (error) {
    trace.end({ error });
}
```

</TabItem>
<TabItem value="go" label="Go">

```go
ctx := context.Background()

// Start a trace
trace, ctx := at.StartTrace(ctx, "my-task")
defer trace.End()

// Your agent logic
result, err := myAgentTask(ctx, "Write a function to sort a list")
if err != nil {
    trace.SetError(err)
    return
}
trace.SetOutput(result)
```

</TabItem>
<TabItem value="cli" label="CLI">

```bash
# Wrap any command with AgentTrace
agenttrace wrap -- python my_agent.py

# Or wrap with explicit trace name
agenttrace wrap --name "code-review" -- ./review-pr.sh
```

</TabItem>
</Tabs>

## Step 5: View Your Traces

1. Go to your [AgentTrace dashboard](https://app.agenttrace.io)
2. Navigate to **Traces**
3. You'll see your trace with timing, inputs, and outputs

![Trace View](./img/trace-view.png)

## What's Next?

Now that you have basic tracing set up, explore more features:

- [Track LLM Generations](/features/observations) - Capture model, tokens, and costs
- [Manage Prompts](/prompts/overview) - Version and test your prompts
- [Score Outputs](/features/scores) - Evaluate agent performance
- [Run Experiments](/datasets/overview) - Test prompts against datasets

## Troubleshooting

### Traces not appearing?

1. **Check your API key** - Ensure it's valid and has the correct permissions
2. **Check the project ID** - Make sure it matches your dashboard
3. **Check network** - AgentTrace uses `api.agenttrace.io` (or your self-hosted URL)
4. **Enable debug logging**:

```python
import logging
logging.getLogger('agenttrace').setLevel(logging.DEBUG)
```

### Need help?

- Join our [Discord](https://discord.gg/agenttrace) for community support
- Check our [GitHub Issues](https://github.com/agenttrace/agenttrace/issues)
- Email [support@agenttrace.io](mailto:support@agenttrace.io)
