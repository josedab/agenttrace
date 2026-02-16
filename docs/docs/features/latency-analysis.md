---
sidebar_position: 6
title: Latency Analysis
description: Visualize and analyze latency across traces and observations with waterfall views, percentile metrics, and slow-span identification in AgentTrace.
---

# Latency Analysis

AgentTrace captures timing data for every trace and observation automatically. Use the latency tools in the dashboard to identify bottlenecks, track performance trends, and set latency targets.

## How Latency Is Captured

Latency is measured implicitly — no extra SDK configuration is needed:

- **Start time** is recorded when a trace or observation is created.
- **End time** is recorded when `.end()` is called (or the context manager exits).
- **Latency** = `endTime - startTime`, computed server-side in milliseconds.

```python
# Latency is tracked automatically
span = trace.span(name="retrieval")     # startTime recorded
docs = fetch_documents(query)
span.end(output=docs)                   # endTime recorded → latency computed
```

## Waterfall Visualization

The **Trace Detail** view includes a latency waterfall that shows every observation on a horizontal timeline:

```
┌─ trace: rag-pipeline ─────────────────────────────────── 1,240 ms ─┐
│  ├─ span: embedding          ███░░░░░░░░░░░░░░░░░░░  120 ms       │
│  ├─ span: vector-search      ░░░████░░░░░░░░░░░░░░░  180 ms       │
│  ├─ span: rerank             ░░░░░░░██░░░░░░░░░░░░░   90 ms       │
│  └─ generation: llm-answer   ░░░░░░░░░███████████░░  850 ms       │
└────────────────────────────────────────────────────────────────────┘
```

The waterfall makes it easy to see:

- **Sequential vs. parallel** operations.
- **Which observation dominates** total trace latency.
- **Gaps** between observations (idle time, network overhead).

## Percentile Metrics

Navigate to **Dashboard → Analytics** for aggregate latency metrics:

| Metric | Description |
|--------|-------------|
| **P50** | Median latency — the typical user experience |
| **P95** | 95th percentile — captures most outliers |
| **P99** | 99th percentile — worst-case performance |
| **Mean** | Average latency across all traces |
| **Max** | Single slowest trace in the time window |

### Filtering

Slice latency data by:

- **Trace name** — compare latency across different operations.
- **Model** — see how model choice affects response time.
- **Time range** — spot performance regressions after deployments.
- **Tags / metadata** — isolate latency for specific users, environments, or features.

## Identifying Slow Spans

AgentTrace highlights observations that consume a disproportionate share of trace latency:

1. **Latency contribution %** — each observation in the waterfall shows what percentage of total trace time it represents.
2. **Slow span indicator** — observations exceeding a configurable threshold (e.g., P95 for that span name) are flagged.
3. **Comparison view** — select two time ranges to compare latency distributions before and after a change.

### Example: Finding the Bottleneck

In the trace list, sort by **Latency (desc)** to surface the slowest traces. Open one and inspect the waterfall:

- If a single generation dominates, consider switching to a faster model or reducing token count.
- If a retrieval span is slow, investigate your vector database performance.
- If there are large gaps between observations, look for network or queuing delays in your application.

## Latency Over Time

The **Latency** chart on the Analytics page plots P50, P95, and P99 over time:

- **Spikes** may correlate with model provider outages or traffic surges.
- **Gradual increases** can indicate growing prompt sizes or degraded infrastructure.
- Hover over any data point to see the exact value and drill into the contributing traces.

## Setting Latency Targets

Use latency data to define SLOs (Service Level Objectives) for your AI features:

1. Establish a baseline — review P50 and P95 over a representative period.
2. Set targets — for example, P95 < 2 seconds for chat responses.
3. Monitor — combine latency analytics with [anomaly detection](./anomaly-detection.md) to alert when targets are breached.

## Best Practices

1. **Always call `.end()`** — unterminated observations will lack an `endTime` and won't appear in latency metrics.
2. **Name observations consistently** — this enables meaningful aggregation (e.g., all `vector-search` spans across traces).
3. **Use nested spans** — break complex operations into sub-spans so the waterfall reveals exactly where time is spent.
4. **Compare before/after** — when optimizing, use the comparison view to validate that latency actually improved.
5. **Monitor P95, not just P50** — median latency hides tail-latency issues that affect a meaningful portion of users.
