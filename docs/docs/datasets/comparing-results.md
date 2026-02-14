---
sidebar_position: 4
title: Comparing Results
description: Compare experiment runs side-by-side in AgentTrace. Analyze score distributions, detect regressions, and use dashboard views for data-driven decisions.
---

# Comparing Results

After running multiple experiments, AgentTrace lets you compare results side-by-side to understand how changes impact quality across your entire dataset.

## Viewing Run Results

### Dashboard

1. Navigate to **Datasets** and select your dataset.
2. Click the **Runs** tab to see all experiment runs.
3. Click a run to view its results, including per-item match status and similarity scores.

Each run summary shows:

| Metric | Description |
|--------|-------------|
| **Match Rate** | Percentage of items where actual output matched expected output. |
| **Avg Similarity** | Mean semantic similarity across all items (0–1). |
| **Completed** | Number of items processed out of the total. |

### API

```bash
curl "https://api.agenttrace.io/v1/datasets/runs/{runId}/results" \
  -H "Authorization: Bearer at-your-api-key"
```

Response includes per-item results and an aggregated summary:

```json
{
  "summary": {
    "matchCount": 20,
    "matchRate": 0.80,
    "avgSimilarity": 0.92
  },
  "results": [
    {
      "datasetItemId": "item-1",
      "match": true,
      "similarity": 1.0,
      "traceId": "trace-abc"
    },
    {
      "datasetItemId": "item-2",
      "match": false,
      "similarity": 0.65,
      "traceId": "trace-def"
    }
  ]
}
```

## Side-by-Side Comparison

### Dashboard

1. From the **Runs** tab, select two or more runs using the checkboxes.
2. Click **Compare Selected**.
3. The comparison view shows each dataset item as a row, with columns for each run's output, match status, and similarity score.

The comparison dashboard highlights:

- **Improvements** — items that went from failing to passing (green).
- **Regressions** — items that went from passing to failing (red).
- **Unchanged** — items with the same result across runs (gray).

### API

```bash
curl "https://api.agenttrace.io/v1/datasets/{datasetId}/runs/compare?runIds=run-1,run-2" \
  -H "Authorization: Bearer at-your-api-key"
```

```json
{
  "runs": [
    { "id": "run-1", "name": "gpt4-v2.0", "matchRate": 0.75 },
    { "id": "run-2", "name": "gpt4-v2.1", "matchRate": 0.85 }
  ],
  "itemComparison": [
    {
      "datasetItemId": "item-1",
      "results": {
        "run-1": { "match": false, "similarity": 0.60 },
        "run-2": { "match": true, "similarity": 0.95 }
      }
    }
  ]
}
```

## Score Distributions

The comparison dashboard includes distribution charts for each run:

- **Similarity histogram** — shows the spread of similarity scores across items.
- **Score box plots** — compare median, quartiles, and outliers between runs.
- **Per-category breakdown** — if items have metadata categories, view match rates per category.

These charts help you identify whether improvements are uniform or concentrated in specific categories.

## Regression Detection

AgentTrace flags regressions automatically when comparing runs:

1. **Item-level regressions** — items that previously matched but no longer do.
2. **Score drops** — items where the similarity score decreased by more than a configurable threshold.
3. **Category regressions** — categories where the overall match rate declined.

In the dashboard, regressions appear with a red indicator. You can click through to the trace to investigate the root cause.

### Filtering Regressions

Use the comparison filters to focus on:

- **Regressions only** — show only items that got worse.
- **Improvements only** — show only items that got better.
- **By category** — filter by metadata category.
- **By similarity delta** — filter by the magnitude of change.

## SDK Examples

### Python

```python
from agenttrace import AgentTrace

client = AgentTrace()

# Get results for a single run
results = client.get_run_results("run-uuid")
print(f"Match rate: {results.summary.match_rate}")

# Compare two runs
for item in results.results:
    if not item.match:
        print(f"Failed: {item.dataset_item_id} (similarity: {item.similarity})")
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });

const results = await client.getRunResults('run-uuid');
console.log(`Match rate: ${results.summary.matchRate}`);

results.results
  .filter((r) => !r.match)
  .forEach((r) => console.log(`Failed: ${r.datasetItemId} (${r.similarity})`));
```

## Best Practices

- **Always compare against a baseline** — establish a known-good run and compare new runs to it.
- **Investigate regressions before deploying** — drill into traces for items that went from passing to failing.
- **Track trends over time** — use match rate and similarity trends to monitor overall quality direction.
- **Use categories for targeted analysis** — metadata categories help you identify which types of inputs are most affected by changes.
- **Export comparison data** — download results as CSV for offline analysis or sharing with stakeholders.
