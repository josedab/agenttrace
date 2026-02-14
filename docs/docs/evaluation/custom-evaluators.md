---
sidebar_position: 4
title: Custom Evaluators
description: Build custom evaluators in AgentTrace using regex matching, JSON schema validation, Python functions, and API-based scoring for deterministic evaluation.
---

# Custom Evaluators

Custom evaluators let you build deterministic, rule-based evaluation logic using regex patterns, JSON schema validation, Python functions, or any external scoring system. They run fast, cost nothing in LLM API calls, and are ideal for format validation and compliance checks.

## When to Use Custom Evaluators

| Use Case | Example |
|----------|---------|
| **Format validation** | Output must be valid JSON |
| **Keyword detection** | Response must include specific terms |
| **Schema compliance** | Output must match a JSON schema |
| **Length checks** | Response must be between 50–500 words |
| **Latency thresholds** | Trace must complete under 5 seconds |
| **Custom business logic** | Domain-specific scoring rules |

## Regex Matching

Use regular expressions to check outputs for patterns:

```python
import re
from agenttrace import AgentTrace

client = AgentTrace()

def evaluate_format(trace_id: str, output: str):
    """Check that the output contains a properly formatted code block."""
    has_code_block = bool(re.search(r'```[\w]*\n.+?\n```', output, re.DOTALL))

    client.score(
        trace_id=trace_id,
        name="has-code-block",
        value=1 if has_code_block else 0,
        comment="Checked for fenced code block in output"
    )
```

## JSON Schema Validation

Validate that outputs conform to an expected structure:

```python
import json
import jsonschema
from agenttrace import AgentTrace

client = AgentTrace()

expected_schema = {
    "type": "object",
    "required": ["summary", "key_points"],
    "properties": {
        "summary": {"type": "string", "minLength": 10},
        "key_points": {
            "type": "array",
            "items": {"type": "string"},
            "minItems": 1
        }
    }
}

def evaluate_schema(trace_id: str, output: str):
    """Validate output against JSON schema."""
    try:
        data = json.loads(output)
        jsonschema.validate(data, expected_schema)
        client.score(
            trace_id=trace_id,
            name="schema-valid",
            value=1,
            comment="Output matches expected schema"
        )
    except (json.JSONDecodeError, jsonschema.ValidationError) as e:
        client.score(
            trace_id=trace_id,
            name="schema-valid",
            value=0,
            comment=f"Schema validation failed: {str(e)[:200]}"
        )
```

## Python Function Evaluators

Build any custom scoring logic as a Python function:

```python
from agenttrace import AgentTrace

client = AgentTrace()

def evaluate_completeness(trace_id: str, output: dict, expected: dict):
    """Score based on how many expected keys are present in the output."""
    expected_keys = set(expected.keys())
    actual_keys = set(output.keys())
    overlap = expected_keys & actual_keys

    score = len(overlap) / len(expected_keys) if expected_keys else 1.0

    client.score(
        trace_id=trace_id,
        name="completeness",
        value=round(score, 2),
        comment=f"Found {len(overlap)}/{len(expected_keys)} expected keys"
    )


def evaluate_word_count(trace_id: str, output: str, min_words: int = 50, max_words: int = 500):
    """Check that the output is within the expected word count range."""
    word_count = len(output.split())
    in_range = min_words <= word_count <= max_words

    client.score(
        trace_id=trace_id,
        name="word-count-ok",
        value=1 if in_range else 0,
        comment=f"Word count: {word_count} (expected {min_words}-{max_words})"
    )
```

## API-Based Scoring

Submit scores from any language or system using the REST API:

```bash
curl -X POST "https://api.agenttrace.io/v1/scores" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "trace-abc-123",
    "name": "custom-check",
    "value": 0.95,
    "comment": "Passed all custom validation rules"
  }'
```

### Batch Scoring

Submit multiple scores in a single request:

```bash
curl -X POST "https://api.agenttrace.io/v1/scores/batch" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "scores": [
      { "traceId": "trace-1", "name": "format-valid", "value": 1 },
      { "traceId": "trace-1", "name": "word-count-ok", "value": 1 },
      { "traceId": "trace-2", "name": "format-valid", "value": 0 }
    ]
  }'
```

## TypeScript Example

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });

async function evaluateJsonOutput(traceId: string, output: string) {
  try {
    const parsed = JSON.parse(output);
    const hasRequired = 'summary' in parsed && 'key_points' in parsed;

    await client.score({
      traceId,
      name: 'json-structure',
      value: hasRequired ? 1 : 0,
      comment: hasRequired ? 'Valid structure' : 'Missing required fields',
    });
  } catch {
    await client.score({
      traceId,
      name: 'json-structure',
      value: 0,
      comment: 'Output is not valid JSON',
    });
  }
}
```

## Running Custom Evaluators in CI/CD

Integrate custom evaluators into your deployment pipeline:

```python
from agenttrace import AgentTrace

client = AgentTrace()

def run_evaluation_suite(trace_id: str, output: str):
    """Run all custom evaluators on a trace."""
    evaluate_format(trace_id, output)
    evaluate_schema(trace_id, output)
    evaluate_word_count(trace_id, output)

    scores = client.get_scores(trace_id=trace_id)
    failed = [s for s in scores if s.value == 0]

    if failed:
        print(f"❌ {len(failed)} check(s) failed")
        for s in failed:
            print(f"  - {s.name}: {s.comment}")
        exit(1)
    else:
        print("✅ All checks passed")
```

## Best Practices

- **Combine with LLM-as-Judge** — use custom evaluators for deterministic checks and LLM-as-Judge for subjective quality.
- **Keep evaluators focused** — each evaluator should check one specific criterion.
- **Use descriptive score names** — names like `schema-valid` and `word-count-ok` are clearer than `check-1`.
- **Include comments** — always add a comment explaining why a score was assigned to aid debugging.
- **Test your evaluators** — validate your scoring logic with known good and bad outputs before deploying.
- **Use batch scoring** — when evaluating many traces, use the batch API to reduce HTTP overhead.
