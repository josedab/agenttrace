# Community Prompt Library

The AgentTrace Prompt Library enables teams to share, discover, and benchmark prompt templates. Build on proven prompts from the community or contribute your own.

## Overview

The library provides:

- **Prompt Sharing** - Share prompts publicly or within your organization
- **Versioning** - Track changes with semantic versioning
- **Forking** - Build on existing prompts with attribution
- **Benchmarking** - Compare prompt performance across models
- **Variables** - Reusable templates with variable substitution
- **Categories & Tags** - Easy discovery and organization

## Creating a Prompt

### Via API

```bash
curl -X POST https://your-agenttrace.com/api/public/library/prompts \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Document Summarizer",
    "description": "Summarizes documents with configurable length and style",
    "category": "summarization",
    "visibility": "public",
    "tags": ["summarization", "documents", "concise"],
    "template": "Summarize the following document in {{length}} sentences.\n\nStyle: {{style}}\n\nDocument:\n{{document}}\n\nSummary:",
    "variables": [
      {
        "name": "document",
        "type": "string",
        "description": "The document to summarize",
        "required": true
      },
      {
        "name": "length",
        "type": "number",
        "description": "Number of sentences in summary",
        "required": false,
        "default": 3
      },
      {
        "name": "style",
        "type": "string",
        "description": "Writing style (formal, casual, technical)",
        "required": false,
        "default": "formal"
      }
    ],
    "examples": [
      {
        "name": "Technical paper",
        "variables": {
          "document": "This paper presents a novel approach to...",
          "length": 5,
          "style": "technical"
        },
        "expected": "A 5-sentence technical summary"
      }
    ],
    "recommendedModels": ["claude-3-5-sonnet-20241022", "gpt-4-turbo"],
    "modelParams": {
      "temperature": 0.3,
      "max_tokens": 500
    }
  }'
```

### Prompt Configuration

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Human-readable prompt name |
| `description` | string | Detailed description of use case |
| `category` | enum | Prompt category (see below) |
| `visibility` | enum | `private`, `organization`, `public` |
| `tags` | array | Searchable tags |
| `template` | string | Prompt template with `{{variables}}` |
| `variables` | array | Variable definitions |
| `examples` | array | Usage examples |
| `recommendedModels` | array | Suggested models |
| `modelParams` | object | Default model parameters |

### Categories

| Category | Description |
|----------|-------------|
| `agent` | Autonomous agent prompts |
| `chat` | Conversational prompts |
| `completion` | Text completion |
| `summarization` | Document summarization |
| `extraction` | Information extraction |
| `classification` | Text classification |
| `code_generation` | Code generation |
| `translation` | Language translation |
| `custom` | Other types |

## Variable Syntax

Variables use double curly braces: `{{variable_name}}`

### Variable Types

```json
{
  "variables": [
    {
      "name": "query",
      "type": "string",
      "description": "User's search query",
      "required": true
    },
    {
      "name": "max_results",
      "type": "number",
      "description": "Maximum results to return",
      "required": false,
      "default": 10
    },
    {
      "name": "include_metadata",
      "type": "boolean",
      "description": "Include result metadata",
      "required": false,
      "default": false
    },
    {
      "name": "categories",
      "type": "array",
      "description": "Categories to filter by",
      "required": false,
      "example": ["news", "blog"]
    }
  ]
}
```

### Rendering Templates

```bash
curl -X POST https://your-agenttrace.com/api/public/library/prompts/{id}/render \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "document": "The quarterly results show...",
      "length": 3,
      "style": "formal"
    }
  }'
```

Response:

```json
{
  "renderedPrompt": "Summarize the following document in 3 sentences.\n\nStyle: formal\n\nDocument:\nThe quarterly results show...\n\nSummary:",
  "promptId": "...",
  "version": 1
}
```

## Discovering Prompts

### Browse Library

```bash
# Popular prompts
curl "https://your-agenttrace.com/api/public/library/prompts?sortBy=popular"

# By category
curl "https://your-agenttrace.com/api/public/library/prompts?category=agent"

# By tags
curl "https://your-agenttrace.com/api/public/library/prompts?tags=code,python"

# Search
curl "https://your-agenttrace.com/api/public/library/prompts?search=summarization"
```

### Get Categories

```bash
curl https://your-agenttrace.com/api/public/library/categories
```

### Popular Tags

```bash
curl https://your-agenttrace.com/api/public/library/tags?limit=20
```

## Starring and Forking

### Star a Prompt

```bash
# Add star
curl -X POST https://your-agenttrace.com/api/public/library/prompts/{id}/star \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"

# Remove star
curl -X DELETE https://your-agenttrace.com/api/public/library/prompts/{id}/star \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

### Fork a Prompt

```bash
curl -X POST https://your-agenttrace.com/api/public/library/prompts/{id}/fork \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Custom Summarizer",
    "visibility": "private"
  }'
```

The fork includes:
- Original template and variables
- Link to source prompt
- Attribution in version notes

## Version History

### Get Versions

```bash
curl https://your-agenttrace.com/api/public/library/prompts/{id}/versions
```

### Get Specific Version

```bash
curl https://your-agenttrace.com/api/public/library/prompts/{id}/versions/2
```

### Create New Version

When updating a prompt, set `bumpVersion: true`:

```bash
curl -X PATCH https://your-agenttrace.com/api/public/library/prompts/{id} \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "Updated template with {{new_variable}}...",
    "variables": [...],
    "bumpVersion": true,
    "versionNotes": "Added new_variable for better control"
  }'
```

## Benchmarking

Compare prompt performance across models and datasets.

### Run Benchmark

```bash
curl -X POST https://your-agenttrace.com/api/public/library/prompts/{id}/benchmark \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "datasetId": "dataset-uuid",
    "sampleCount": 100,
    "evaluators": ["accuracy", "relevance"]
  }'
```

### Get Benchmark Results

```bash
curl https://your-agenttrace.com/api/public/library/prompts/{id}/benchmarks
```

Response:

```json
[
  {
    "id": "...",
    "promptId": "...",
    "promptVersion": 2,
    "model": "claude-3-5-sonnet-20241022",
    "datasetName": "Summarization Test Set",
    "sampleCount": 100,
    "metrics": {
      "accuracy": 0.92,
      "relevance": 0.88,
      "avgLatencyMs": 450,
      "p95LatencyMs": 720,
      "avgTokens": 125,
      "totalCost": 0.15,
      "avgCostPerCall": 0.0015,
      "successRate": 0.99,
      "errorCount": 1
    },
    "runAt": "2024-01-15T10:30:00Z",
    "durationSeconds": 120
  }
]
```

### Benchmark Metrics

| Metric | Description |
|--------|-------------|
| `accuracy` | Quality score (0-1) |
| `relevance` | Response relevance (0-1) |
| `coherence` | Logical coherence (0-1) |
| `helpfulness` | Helpfulness rating (0-1) |
| `avgLatencyMs` | Average response time |
| `p95LatencyMs` | 95th percentile latency |
| `avgTokens` | Average tokens per call |
| `totalCost` | Total benchmark cost |
| `successRate` | Successful completion rate |

## Usage Tracking

Track when library prompts are used in your projects.

### Record Usage

```bash
curl -X POST https://your-agenttrace.com/api/public/library/prompts/{id}/usage \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "version": 1,
    "projectId": "project-uuid",
    "traceId": "trace-uuid"
  }'
```

Usage tracking:
- Links library prompts to traces
- Tracks which versions are actively used
- Enables impact analysis for updates

## Visibility Levels

| Level | Description |
|-------|-------------|
| `private` | Only visible to the author |
| `organization` | Visible to organization members |
| `public` | Visible to everyone in the library |

### Publishing a Prompt

Make a private prompt public:

```bash
curl -X POST https://your-agenttrace.com/api/public/library/prompts/{id}/publish \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

## SDK Integration

### Python

```python
from agenttrace import AgentTrace

client = AgentTrace()

# Use a library prompt
prompt = client.library.get("summarizer-prompt-id")
rendered = prompt.render(
    document="The quarterly results...",
    length=3
)

# Track usage
with client.trace(name="summarization") as trace:
    response = call_llm(rendered)
    prompt.record_usage(trace_id=trace.id)
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace();

// Use a library prompt
const prompt = await client.library.get('summarizer-prompt-id');
const rendered = prompt.render({
  document: 'The quarterly results...',
  length: 3
});

// Track usage
const trace = client.trace({ name: 'summarization' });
const response = await callLLM(rendered);
await prompt.recordUsage({ traceId: trace.id });
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/public/library/prompts` | GET | List library prompts |
| `/api/public/library/prompts` | POST | Create prompt |
| `/api/public/library/prompts/{id}` | GET | Get prompt |
| `/api/public/library/prompts/{id}` | PATCH | Update prompt |
| `/api/public/library/prompts/{id}` | DELETE | Delete prompt |
| `/api/public/library/prompts/slug/{slug}` | GET | Get prompt by slug |
| `/api/public/library/prompts/{id}/fork` | POST | Fork prompt |
| `/api/public/library/prompts/{id}/publish` | POST | Publish prompt |
| `/api/public/library/prompts/{id}/star` | POST | Star prompt |
| `/api/public/library/prompts/{id}/star` | DELETE | Unstar prompt |
| `/api/public/library/prompts/{id}/versions` | GET | Get versions |
| `/api/public/library/prompts/{id}/versions/{v}` | GET | Get specific version |
| `/api/public/library/prompts/{id}/render` | POST | Render template |
| `/api/public/library/prompts/{id}/benchmark` | POST | Run benchmark |
| `/api/public/library/prompts/{id}/benchmarks` | GET | Get benchmarks |
| `/api/public/library/prompts/{id}/usage` | POST | Record usage |
| `/api/public/library/prompts/starred` | GET | Get starred prompts |
| `/api/public/library/prompts/mine` | GET | Get my prompts |
| `/api/public/library/categories` | GET | Get categories |
| `/api/public/library/tags` | GET | Get popular tags |

## Best Practices

### Writing Good Prompts

1. **Clear Instructions** - Be explicit about the expected output
2. **Examples** - Include input/output examples in the template
3. **Constraints** - Specify length, format, and style constraints
4. **Variables** - Use variables for reusable, configurable prompts

### Versioning Strategy

1. **Patch versions** - Minor wording tweaks (don't bump version)
2. **Minor versions** - Add new variables or examples (bump version)
3. **Major versions** - Significant behavior changes (consider new prompt)

### Benchmarking Tips

1. **Representative datasets** - Use datasets that reflect real usage
2. **Multiple models** - Benchmark across model families
3. **Track over time** - Re-run benchmarks after model updates
4. **Document findings** - Add benchmark context to description
