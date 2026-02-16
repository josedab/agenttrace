---
sidebar_position: 5
title: Cost Tracking
description: Automatically track and analyze LLM costs across 400+ models with AgentTrace's built-in pricing engine and dashboard views.
---

# Cost Tracking

AgentTrace automatically calculates costs for LLM generations across 400+ models. Every generation that reports token usage is matched against a built-in pricing table, giving you real-time cost visibility without any manual configuration.

## How It Works

1. **Report token usage** — your SDK calls include `usage` data (input tokens, output tokens) when ending a generation.
2. **Model matching** — AgentTrace maps the `model` field to its pricing table (updated regularly with new models and price changes).
3. **Async cost calculation** — a background worker calculates costs shortly after ingestion, so trace data is available immediately while costs populate asynchronously.
4. **Aggregation** — costs roll up from individual generations → observations → traces → sessions, giving you totals at every level.

## Reporting Token Usage

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="python" label="Python" default>

```python
generation = trace.generation(
    name="answer-question",
    model="gpt-4o",
    model_parameters={"temperature": 0.2},
    input={"messages": messages}
)
response = call_openai(messages)
generation.end(
    output=response.content,
    usage={
        "input_tokens": response.usage.prompt_tokens,
        "output_tokens": response.usage.completion_tokens,
    }
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
const generation = trace.generation({
  name: 'answer-question',
  model: 'gpt-4o',
  modelParameters: { temperature: 0.2 },
  input: { messages },
});
const response = await callOpenAI(messages);
generation.end({
  output: response.content,
  usage: {
    inputTokens: response.usage.prompt_tokens,
    outputTokens: response.usage.completion_tokens,
  },
});
```

</TabItem>
</Tabs>

:::tip
If you use an AgentTrace integration (OpenAI, Anthropic, LangChain, etc.), token usage is captured automatically — no manual `usage` parameter needed.
:::

## Supported Models

AgentTrace maintains a pricing table for 400+ models, including:

| Provider | Example Models |
|----------|---------------|
| OpenAI | GPT-4o, GPT-4 Turbo, GPT-3.5 Turbo, o1, o1-mini |
| Anthropic | Claude 4 Sonnet, Claude 3.5 Haiku, Claude 3 Opus |
| Google | Gemini 2.0 Flash, Gemini 1.5 Pro |
| Meta | Llama 3.1 (via hosted providers) |
| Mistral | Mistral Large, Codestral |
| Cohere | Command R+ |

Pricing is defined per million input and output tokens. The table is updated with each AgentTrace release; self-hosted users can customize model pricing via the admin settings.

## Dashboard Cost Views

### Trace List

The trace table includes a **Cost** column showing the total cost per trace. Sort or filter by cost to find expensive requests quickly.

### Trace Detail

Open any trace to see cost broken down by observation:

- Each generation shows its individual cost and token counts.
- The trace header shows the aggregated total.

### Cost Analytics

Navigate to **Dashboard → Analytics** to access cost dashboards:

| View | Description |
|------|-------------|
| **Cost over time** | Daily / weekly cost trends with breakdown by model |
| **Cost by model** | Compare spend across different LLM providers and models |
| **Cost by user** | Per-user cost attribution |
| **Cost by trace name** | Identify which operations are most expensive |

## Custom Model Pricing

For self-hosted deployments or models not yet in the default table, you can define custom pricing:

```bash
# Via the Admin API
curl -X POST https://your-instance.example.com/api/admin/model-pricing \
  -H "Authorization: Bearer $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "my-fine-tuned-model",
    "inputPricePerMillion": 3.00,
    "outputPricePerMillion": 6.00
  }'
```

## Best Practices

1. **Always report usage** — even if cost calculation is automatic, it relies on accurate token counts from your SDK calls.
2. **Use integrations** — AgentTrace integrations capture usage automatically, reducing the chance of missing data.
3. **Set cost alerts** — combine cost tracking with [anomaly detection](./anomaly-detection.md) to get notified about unexpected spend.
4. **Review by model** — periodically check cost-by-model analytics to identify opportunities to switch to cheaper models without quality loss.
