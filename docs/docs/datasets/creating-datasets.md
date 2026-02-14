---
sidebar_position: 2
title: Creating Datasets
description: How to create datasets and add items via the AgentTrace UI, REST API, and SDKs. Includes CSV/JSON import and trace-based item creation.
---

# Creating Datasets

AgentTrace provides multiple ways to create datasets and populate them with test items — through the dashboard UI, the REST API, or the Python and TypeScript SDKs.

## Creating a Dataset via the UI

1. Navigate to **Datasets** in the left sidebar.
2. Click **+ New Dataset**.
3. Enter a **name** and optional **description**.
4. Click **Create**.

The dataset is now ready to accept items.

## Creating a Dataset via the API

```bash
curl -X POST "https://api.agenttrace.io/v1/datasets" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "summarization-tests",
    "description": "Test cases for document summarization agent"
  }'
```

## Adding Items

### Manual Item Creation

Add individual test cases with an input, expected output, and optional metadata:

```bash
curl -X POST "https://api.agenttrace.io/v1/datasets/{datasetId}/items" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "input": { "document": "AgentTrace is an observability platform..." },
    "expectedOutput": { "summary": "AgentTrace provides LLM observability." },
    "metadata": { "category": "short-doc", "difficulty": "easy" }
  }'
```

### From Existing Traces

Turn production interactions into test cases by creating items from traces:

```bash
curl -X POST "https://api.agenttrace.io/v1/datasets/{datasetId}/items/from-trace" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc-123",
    "observationId": "gen-xyz-789"
  }'
```

This extracts the input and output from the trace and stores a `sourceTraceId` link for traceability.

### Batch Import

Import multiple items at once:

```bash
curl -X POST "https://api.agenttrace.io/v1/datasets/{datasetId}/items/batch" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      { "input": { "document": "First doc..." }, "expectedOutput": { "summary": "First summary." } },
      { "input": { "document": "Second doc..." }, "expectedOutput": { "summary": "Second summary." } }
    ]
  }'
```

### Importing from CSV

In the dashboard UI:

1. Open your dataset and click **Import**.
2. Select **CSV** format.
3. Upload your file. The CSV should have columns mapping to `input`, `expectedOutput`, and optionally `metadata`.
4. Map columns in the preview screen and click **Import**.

CSV example:

```csv
input_document,expected_summary,category
"AgentTrace is an observability platform...","AgentTrace provides LLM observability.","short-doc"
"Machine learning models require...","ML models need training data.","medium-doc"
```

### Importing from JSON

Upload a JSON file with an array of items:

```json
[
  {
    "input": { "document": "AgentTrace is an observability platform..." },
    "expectedOutput": { "summary": "AgentTrace provides LLM observability." },
    "metadata": { "category": "short-doc" }
  }
]
```

## SDK Examples

### Python

```python
from agenttrace import AgentTrace

client = AgentTrace()

# Create dataset
dataset = client.create_dataset(
    name="summarization-tests",
    description="Test cases for document summarization"
)

# Add an item
client.create_dataset_item(
    dataset_id=dataset.id,
    input={"document": "AgentTrace is an observability platform..."},
    expected_output={"summary": "AgentTrace provides LLM observability."},
    metadata={"category": "short-doc"}
)

# Add item from a production trace
client.create_dataset_item_from_trace(
    dataset_id=dataset.id,
    trace_id="trace-abc-123"
)
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });

// Create dataset
const dataset = await client.createDataset({
  name: 'summarization-tests',
  description: 'Test cases for document summarization',
});

// Add an item
await client.createDatasetItem({
  datasetId: dataset.id,
  input: { document: 'AgentTrace is an observability platform...' },
  expectedOutput: { summary: 'AgentTrace provides LLM observability.' },
  metadata: { category: 'short-doc' },
});
```

## Best Practices

- **Start from production traces** — real-world inputs make the most representative test cases.
- **Use metadata for categorization** — tag items with difficulty, category, or source for filtered analysis.
- **Keep expected outputs focused** — define the key properties you care about, not the exact wording.
- **Version your datasets** — export as JSON and commit alongside your code for reproducibility.
- **Aim for diversity** — include common cases, edge cases, and known failure modes.
