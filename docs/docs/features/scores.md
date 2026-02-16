---
sidebar_position: 4
title: Scores
description: Evaluate and score traces and observations in AgentTrace with numeric, boolean, and categorical scores from APIs, evaluators, or human annotation.
---

# Scores

Scores let you attach quality metrics to [traces](./tracing.md) and [observations](./observations.md). Use them to track accuracy, relevance, user satisfaction, or any custom evaluation criterion.

## Score Types

| Type | Value | Example |
|------|-------|---------|
| **NUMERIC** | Float between 0 and 1 | `accuracy: 0.92` |
| **BOOLEAN** | `true` or `false` | `hallucination: false` |
| **CATEGORICAL** | String label | `sentiment: "positive"` |

## Score Sources

Scores can originate from three sources:

- **API** — programmatic scoring from your application code or CI pipeline.
- **Evaluator** — automated evaluation functions that run server-side (e.g., LLM-as-judge, regex checks).
- **Human annotation** — manual review via the AgentTrace dashboard.

## Adding Scores via SDK

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

### Numeric Score

<Tabs>
<TabItem value="python" label="Python" default>

```python
from agenttrace import AgentTrace

at = AgentTrace()

# Score a trace
at.score(
    trace_id="trace-abc-123",
    name="relevance",
    value=0.87,
    data_type="NUMERIC",
    comment="Answer was relevant but missed one key detail"
)

# Score a specific observation within a trace
at.score(
    trace_id="trace-abc-123",
    observation_id="gen-456",
    name="faithfulness",
    value=0.95,
    data_type="NUMERIC"
)
```

</TabItem>
<TabItem value="typescript" label="TypeScript">

```typescript
import { AgentTrace } from 'agenttrace';

const at = new AgentTrace();

// Score a trace
at.score({
  traceId: 'trace-abc-123',
  name: 'relevance',
  value: 0.87,
  dataType: 'NUMERIC',
  comment: 'Answer was relevant but missed one key detail',
});

// Score a specific observation within a trace
at.score({
  traceId: 'trace-abc-123',
  observationId: 'gen-456',
  name: 'faithfulness',
  value: 0.95,
  dataType: 'NUMERIC',
});
```

</TabItem>
</Tabs>

### Boolean Score

```python
at.score(
    trace_id="trace-abc-123",
    name="hallucination",
    value=False,
    data_type="BOOLEAN",
    comment="Verified against source documents"
)
```

### Categorical Score

```python
at.score(
    trace_id="trace-abc-123",
    name="tone",
    value="professional",
    data_type="CATEGORICAL",
    comment="Appropriate for enterprise context"
)
```

### Inline Scoring on a Trace

You can also score directly from a trace object without supplying the trace ID:

```python
with at.trace("my-task") as trace:
    result = run_task()
    trace.output = result

    # Score inline
    trace.score(name="accuracy", value=0.91, data_type="NUMERIC")
    trace.score(name="contains-pii", value=False, data_type="BOOLEAN")
```

## Score Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `traceId` | string | ✅ | The trace to score |
| `observationId` | string | — | Optional observation within the trace |
| `name` | string | ✅ | Score name (e.g., `relevance`, `accuracy`) |
| `value` | number / boolean / string | ✅ | The score value |
| `dataType` | string | — | `NUMERIC`, `BOOLEAN`, or `CATEGORICAL` |
| `comment` | string | — | Free-text explanation |

## Human Annotation

In the dashboard, open any trace and click **Add Score** to manually annotate:

1. Select or create a score name.
2. Choose the data type.
3. Enter a value and optional comment.
4. Click **Save**.

Human annotations appear alongside API and evaluator scores with a `HUMAN` source label so you can filter by origin.

## Viewing Scores in the Dashboard

- **Trace list** — Score columns can be added to the trace table for at-a-glance quality metrics.
- **Trace detail** — All scores attached to a trace and its observations are shown in the **Scores** tab.
- **Analytics** — Aggregate scores over time to monitor quality trends. Filter by score name, source, or data type.

## Best Practices

1. **Use consistent naming** — standardize score names across your project (e.g., always `relevance`, not `relevance_score`).
2. **Score at the right level** — score the trace for end-to-end quality; score individual observations for component-level evaluation.
3. **Combine sources** — use automated evaluators for scale and human annotation for calibration.
4. **Track over time** — monitor score distributions to catch quality regressions early.
