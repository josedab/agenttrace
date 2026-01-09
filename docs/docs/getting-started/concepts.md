---
sidebar_position: 4
---

# Core Concepts

Understanding AgentTrace's core concepts will help you get the most out of the platform.

## Traces

A **trace** represents a complete unit of work, such as:
- A single AI agent task
- A user request to your application
- A CI/CD pipeline run

```
Trace: "code-review"
├── Input: { file: "main.py" }
├── Output: { review: "Looks good!" }
├── Metadata: { repository: "my-app" }
├── Latency: 5.2s
└── Cost: $0.003
```

### Trace Properties

| Property | Description |
|----------|-------------|
| `name` | Human-readable name for the trace |
| `input` | Input data (captured automatically or manually) |
| `output` | Output/result data |
| `metadata` | Custom key-value pairs |
| `tags` | Filterable labels |
| `level` | Log level (DEBUG, INFO, WARNING, ERROR) |
| `sessionId` | Groups related traces |
| `userId` | User who initiated the trace |

## Observations

**Observations** are individual operations within a trace. There are three types:

### Generations

LLM calls with model, tokens, and cost tracking:

```python
with trace.generation(
    name="analyze-code",
    model="claude-3-sonnet",
    input=[{"role": "user", "content": "..."}]
) as gen:
    response = call_llm(...)
    gen.output = response
    gen.usage = {"inputTokens": 150, "outputTokens": 200}
```

### Spans

Non-LLM operations like database queries, API calls, or processing:

```python
with trace.span(name="fetch-context") as span:
    context = fetch_from_database()
    span.output = {"records": len(context)}
```

### Events

Discrete points in time, like checkpoints or state changes:

```python
trace.event(name="checkpoint", data={"step": 1})
```

## Observation Hierarchy

Observations can be nested to show execution flow:

```
Trace: "code-review"
└── Span: "analyze"
    ├── Generation: "understand-code" (claude-3-sonnet)
    ├── Span: "fetch-context"
    │   └── Span: "database-query"
    └── Generation: "generate-review" (claude-3-sonnet)
```

## Sessions

A **session** groups related traces together, typically representing:
- A user conversation
- A multi-step workflow
- A debugging session

```python
session = at.session(id="user-123-session")

with session.trace("task-1"):
    process_task_1()

with session.trace("task-2"):
    process_task_2()
```

Sessions help you:
- Track user journeys
- Analyze multi-turn conversations
- Debug complex workflows

## Prompts

**Prompts** are versioned, managed templates for LLM inputs:

```
Prompt: "code-review"
├── Version 1 (draft)
├── Version 2 (staging)
├── Version 3 (production) ← labeled
└── Version 4 (draft)
```

### Prompt Properties

| Property | Description |
|----------|-------------|
| `name` | Unique identifier |
| `version` | Auto-incrementing version number |
| `label` | Named pointer (e.g., "production") |
| `content` | Template with `{{variables}}` |

### Prompt Compilation

```python
prompt = at.get_prompt("code-review", label="production")

compiled = prompt.compile(
    language="Python",
    code="def hello(): print('world')"
)
# "Review this Python code:\n```python\ndef hello(): print('world')\n```"
```

## Scores

**Scores** are evaluations attached to traces or observations:

| Type | Example | Use Case |
|------|---------|----------|
| Numeric | `0.95` | Quality scores, confidence |
| Boolean | `1` / `0` | Correct/incorrect |
| Categorical | `"good"` | Rating categories |

### Score Sources

- **API**: Programmatically created scores
- **LLM**: LLM-as-judge evaluations
- **Human**: Manual annotations

```python
# API score
at.score(trace_id="...", name="quality", value=0.95)

# LLM-as-judge (via evaluator)
evaluator = at.get_evaluator("hallucination-check")
evaluator.execute(trace_id="...")

# Human annotation (via dashboard)
```

## Datasets

**Datasets** are collections of input-output pairs for testing:

```
Dataset: "code-review-examples"
├── Item 1: { input: {...}, expected: {...} }
├── Item 2: { input: {...}, expected: {...} }
└── Item 3: { input: {...}, expected: {...} }
```

### Dataset Runs

Run your agent against a dataset to compare results:

```
Run: "experiment-1"
├── Item 1: { actual: {...}, scores: [...] }
├── Item 2: { actual: {...}, scores: [...] }
└── Item 3: { actual: {...}, scores: [...] }
```

## Evaluators

**Evaluators** automatically score trace outputs:

### LLM-as-Judge

Uses an LLM to evaluate quality:

```yaml
name: hallucination-check
type: llm
model: gpt-4
prompt: |
  Evaluate if the response contains hallucinations.
  Output: {"score": 0-1, "reason": "..."}
```

### Code Evaluators

Custom code for specific checks:

```python
def evaluate(trace):
    if "error" in trace.output:
        return {"score": 0, "reason": "Contains error"}
    return {"score": 1, "reason": "No errors"}
```

## Cost Tracking

AgentTrace automatically calculates costs for LLM calls:

```
Generation: "analyze-code"
├── Model: claude-3-sonnet
├── Input tokens: 1,500
├── Output tokens: 500
├── Input cost: $0.0045
├── Output cost: $0.0025
└── Total cost: $0.007
```

Costs are aggregated at:
- Observation level
- Trace level
- Session level
- Project level
- Time periods

## Projects & Organizations

```
Organization: "Acme Corp"
├── Project: "Production Agent"
│   ├── API Keys
│   ├── Prompts
│   ├── Datasets
│   └── Traces
└── Project: "Development"
    ├── API Keys
    └── Traces
```

### Access Control

| Role | Permissions |
|------|-------------|
| Owner | Full access, billing, delete org |
| Admin | Manage members, projects, settings |
| Member | View/create traces, prompts |
| Viewer | Read-only access |

## Next Steps

Now that you understand the concepts:

- [Create your first trace](/getting-started/first-trace)
- [Track LLM calls](/features/observations)
- [Manage prompts](/prompts/overview)
- [Set up evaluation](/evaluation/overview)
