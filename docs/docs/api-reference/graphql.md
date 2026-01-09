---
sidebar_position: 10
---

# GraphQL API

AgentTrace provides a GraphQL API alongside the REST API, offering flexible querying capabilities with type safety and efficient data fetching.

## Endpoint

**Playground (Development)**: `http://localhost:8080/graphql`

**Production**: GraphQL playground is disabled in production for security. Use the endpoint directly.

## Authentication

Include your API key in the `Authorization` header:

```bash
curl -X POST "https://api.agenttrace.io/graphql" \
  -H "Authorization: Bearer sk-at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"query": "{ traces(input: {limit: 10}) { edges { node { id name } } } }"}'
```

## Schema Overview

The GraphQL API provides queries and mutations for all AgentTrace resources:

- **Traces & Observations**: Query and create traces, spans, generations, events
- **Scores**: Create and query evaluation scores
- **Prompts**: Manage versioned prompts with labels
- **Datasets**: Create datasets, items, and experiment runs
- **Evaluators**: Configure automated evaluations
- **Organizations & Projects**: Manage workspace structure

## Queries

### Traces

```graphql
# Get a single trace by ID
query GetTrace($id: String!) {
  trace(id: $id) {
    id
    name
    input
    output
    metadata
    startTime
    endTime
    latency
    totalCost
    totalTokens
    tags
    userId
    sessionId
    release
    version
  }
}

# List traces with filtering and pagination
query ListTraces($input: TracesInput!) {
  traces(input: $input) {
    edges {
      node {
        id
        name
        startTime
        latency
        totalCost
      }
      cursor
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    totalCount
  }
}
```

**TracesInput parameters:**

| Field | Type | Description |
|-------|------|-------------|
| `limit` | Int | Max results (default: 50, max: 100) |
| `cursor` | String | Pagination cursor |
| `userId` | String | Filter by user ID |
| `sessionId` | String | Filter by session ID |
| `name` | String | Filter by trace name |
| `tags` | [String] | Filter by tags |
| `fromTimestamp` | DateTime | Start time filter |
| `toTimestamp` | DateTime | End time filter |
| `version` | String | Filter by version |
| `release` | String | Filter by release |
| `orderBy` | String | Sort field |
| `order` | String | Sort direction (asc/desc) |

### Observations

```graphql
# Get a single observation
query GetObservation($id: String!) {
  observation(id: $id) {
    id
    traceId
    parentObservationId
    type
    name
    startTime
    endTime
    model
    input
    output
    metadata
    level
    statusMessage
    promptTokens
    completionTokens
    totalTokens
    calculatedCost
  }
}

# List observations for a trace
query ListObservations($input: ObservationsInput!) {
  observations(input: $input) {
    edges {
      node {
        id
        type
        name
        model
        latency
        totalTokens
      }
      cursor
    }
    totalCount
  }
}
```

**ObservationsInput parameters:**

| Field | Type | Description |
|-------|------|-------------|
| `traceId` | String | Filter by trace ID |
| `parentObservationId` | String | Filter by parent |
| `type` | ObservationType | SPAN, GENERATION, or EVENT |
| `name` | String | Filter by name |
| `limit` | Int | Max results |
| `cursor` | String | Pagination cursor |

### Scores

```graphql
# Get scores for a trace
query GetScores($input: ScoresInput!) {
  scores(input: $input) {
    edges {
      node {
        id
        traceId
        observationId
        name
        value
        stringValue
        dataType
        source
        comment
        createdAt
      }
    }
    totalCount
  }
}
```

### Prompts

```graphql
# Get a prompt by name (latest version)
query GetPrompt($name: String!) {
  prompt(name: $name) {
    id
    name
    description
    type
    version
    content
    config
    labels
    createdAt
  }
}

# Get a specific version
query GetPromptVersion($name: String!, $version: Int!) {
  prompt(name: $name, version: $version) {
    id
    version
    content
  }
}

# Get a labeled version (e.g., "production")
query GetPromptByLabel($name: String!, $label: String!) {
  prompt(name: $name, label: $label) {
    id
    version
    content
    labels
  }
}

# List all prompts
query ListPrompts($input: PromptsInput!) {
  prompts(input: $input) {
    edges {
      node {
        id
        name
        version
        labels
      }
    }
    totalCount
  }
}
```

### Datasets

```graphql
# Get a dataset by ID
query GetDataset($id: ID!) {
  dataset(id: $id) {
    id
    name
    description
    itemCount
    metadata
    createdAt
  }
}

# Get dataset by name
query GetDatasetByName($name: String!) {
  datasetByName(name: $name) {
    id
    name
    itemCount
  }
}

# List datasets
query ListDatasets($input: DatasetsInput!) {
  datasets(input: $input) {
    edges {
      node {
        id
        name
        itemCount
      }
    }
    totalCount
  }
}
```

### Evaluators

```graphql
# Get an evaluator
query GetEvaluator($id: ID!) {
  evaluator(id: $id) {
    id
    name
    type
    description
    config
    enabled
    samplingRate
    createdAt
  }
}

# List evaluators
query ListEvaluators($input: EvaluatorsInput!) {
  evaluators(input: $input) {
    edges {
      node {
        id
        name
        type
        enabled
      }
    }
    totalCount
  }
}

# Get evaluator templates
query GetEvaluatorTemplates {
  evaluatorTemplates {
    id
    name
    description
    type
    defaultConfig
  }
}
```

### Metrics

```graphql
# Get aggregated metrics
query GetMetrics($input: MetricsInput!) {
  metrics(input: $input) {
    traceCount
    observationCount
    totalCost
    totalTokens
    avgLatency
    p50Latency
    p95Latency
    p99Latency
    modelUsage {
      model
      count
      promptTokens
      completionTokens
      cost
    }
  }
}

# Get daily cost breakdown
query GetDailyCosts($input: DailyCostsInput!) {
  dailyCosts(input: $input) {
    date
    totalCost
    traceCount
    modelCosts {
      model
      cost
      count
    }
  }
}
```

### User & Organization

```graphql
# Get current user
query Me {
  me {
    id
    email
    name
    avatarUrl
  }
}

# List organizations
query ListOrganizations {
  organizations {
    id
    name
    createdAt
  }
}

# List projects
query ListProjects($organizationId: ID) {
  projects(organizationId: $organizationId) {
    id
    name
    description
  }
}
```

## Mutations

### Traces

```graphql
# Create a trace
mutation CreateTrace($input: CreateTraceInput!) {
  createTrace(input: $input) {
    id
    name
  }
}

# Update a trace
mutation UpdateTrace($id: String!, $input: UpdateTraceInput!) {
  updateTrace(id: $id, input: $input) {
    id
    name
    output
  }
}
```

**CreateTraceInput:**

```graphql
input CreateTraceInput {
  id: String
  name: String
  userId: String
  sessionId: String
  input: JSON
  output: JSON
  metadata: JSON
  tags: [String!]
  release: String
  version: String
  public: Boolean
  timestamp: DateTime
}
```

### Observations

```graphql
# Create a span
mutation CreateSpan($input: CreateObservationInput!) {
  createSpan(input: $input) {
    id
    traceId
    name
  }
}

# Create a generation (LLM call)
mutation CreateGeneration($input: CreateGenerationInput!) {
  createGeneration(input: $input) {
    id
    traceId
    name
    model
    totalTokens
    calculatedCost
  }
}

# Create an event
mutation CreateEvent($input: CreateObservationInput!) {
  createEvent(input: $input) {
    id
    name
  }
}

# Update an observation
mutation UpdateObservation($id: String!, $input: UpdateObservationInput!) {
  updateObservation(id: $id, input: $input) {
    id
    output
    endTime
  }
}
```

**CreateGenerationInput:**

```graphql
input CreateGenerationInput {
  id: String
  traceId: String!
  parentObservationId: String
  name: String
  model: String
  modelParameters: JSON
  input: JSON
  output: JSON
  metadata: JSON
  usage: UsageInput
  startTime: DateTime
  endTime: DateTime
  level: String
  statusMessage: String
  promptId: ID
  version: String
}

input UsageInput {
  promptTokens: Int
  completionTokens: Int
  totalTokens: Int
}
```

### Scores

```graphql
# Create a score
mutation CreateScore($input: CreateScoreInput!) {
  createScore(input: $input) {
    id
    name
    value
  }
}

# Update a score
mutation UpdateScore($id: String!, $input: UpdateScoreInput!) {
  updateScore(id: $id, input: $input) {
    id
    value
    comment
  }
}
```

**CreateScoreInput:**

```graphql
input CreateScoreInput {
  traceId: String!
  observationId: String
  name: String!
  value: Float
  stringValue: String
  dataType: ScoreDataType
  source: ScoreSource
  comment: String
}
```

### Prompts

```graphql
# Create a prompt
mutation CreatePrompt($input: CreatePromptInput!) {
  createPrompt(input: $input) {
    id
    name
    version
  }
}

# Update a prompt (creates new version)
mutation UpdatePrompt($name: String!, $input: UpdatePromptInput!) {
  updatePrompt(name: $name, input: $input) {
    id
    name
    version
  }
}

# Set a label on a prompt version
mutation SetPromptLabel($name: String!, $version: Int!, $label: String!) {
  setPromptLabel(name: $name, version: $version, label: $label)
}

# Remove a label
mutation RemovePromptLabel($name: String!, $label: String!) {
  removePromptLabel(name: $name, label: $label)
}

# Delete a prompt
mutation DeletePrompt($name: String!) {
  deletePrompt(name: $name)
}
```

### Datasets

```graphql
# Create a dataset
mutation CreateDataset($input: CreateDatasetInput!) {
  createDataset(input: $input) {
    id
    name
  }
}

# Add item to dataset
mutation CreateDatasetItem($datasetId: ID!, $input: CreateDatasetItemInput!) {
  createDatasetItem(datasetId: $datasetId, input: $input) {
    id
    input
    expectedOutput
  }
}

# Create experiment run
mutation CreateDatasetRun($datasetId: ID!, $input: CreateDatasetRunInput!) {
  createDatasetRun(datasetId: $datasetId, input: $input) {
    id
    name
  }
}

# Add result to run
mutation AddDatasetRunItem($runId: ID!, $input: AddDatasetRunItemInput!) {
  addDatasetRunItem(runId: $runId, input: $input) {
    id
    datasetItemId
    observationId
  }
}
```

### Evaluators

```graphql
# Create an evaluator
mutation CreateEvaluator($input: CreateEvaluatorInput!) {
  createEvaluator(input: $input) {
    id
    name
    type
  }
}

# Update an evaluator
mutation UpdateEvaluator($id: ID!, $input: UpdateEvaluatorInput!) {
  updateEvaluator(id: $id, input: $input) {
    id
    enabled
  }
}

# Delete an evaluator
mutation DeleteEvaluator($id: ID!) {
  deleteEvaluator(id: $id)
}
```

### Organizations & Projects

```graphql
# Create organization
mutation CreateOrganization($name: String!) {
  createOrganization(name: $name) {
    id
    name
  }
}

# Create project
mutation CreateProject($input: CreateProjectInput!) {
  createProject(input: $input) {
    id
    name
  }
}

# Create API key
mutation CreateAPIKey($input: CreateAPIKeyInput!) {
  createAPIKey(input: $input) {
    id
    name
    key  # Only returned once!
    displayKey
    scopes
    expiresAt
  }
}
```

## Subscriptions

AgentTrace supports GraphQL subscriptions for real-time updates:

```graphql
# Subscribe to new traces
subscription OnTraceCreated {
  traceCreated {
    id
    name
    startTime
  }
}

# Subscribe to trace updates
subscription OnTraceUpdated($traceId: String!) {
  traceUpdated(traceId: $traceId) {
    id
    output
    endTime
    totalCost
  }
}

# Subscribe to new observations on a trace
subscription OnObservationCreated($traceId: String!) {
  observationCreated(traceId: $traceId) {
    id
    type
    name
    startTime
  }
}
```

## Types

### Enums

```graphql
enum ObservationType {
  SPAN
  GENERATION
  EVENT
}

enum ScoreDataType {
  NUMERIC
  BOOLEAN
  CATEGORICAL
}

enum ScoreSource {
  API
  ANNOTATION
  EVAL
}

enum EvaluatorType {
  LLM_AS_JUDGE
  CUSTOM
  MANUAL
}

enum PromptType {
  TEXT
  CHAT
}
```

### Scalars

```graphql
scalar DateTime  # ISO 8601 format
scalar JSON      # Arbitrary JSON object
scalar ID        # UUID
```

## Example: Complete Tracing Flow

```graphql
# 1. Create a trace
mutation {
  createTrace(input: {
    name: "code-review"
    userId: "user-123"
    input: {file: "main.py", language: "python"}
    metadata: {repository: "my-app"}
  }) {
    id
  }
}

# 2. Add a generation
mutation {
  createGeneration(input: {
    traceId: "trace-id-from-step-1"
    name: "analyze-code"
    model: "claude-3-sonnet"
    input: [{role: "user", content: "Review this code..."}]
    modelParameters: {temperature: 0.7}
  }) {
    id
  }
}

# 3. Update generation with output
mutation {
  updateObservation(
    id: "generation-id-from-step-2"
    input: {
      output: {role: "assistant", content: "The code looks good..."}
      usage: {promptTokens: 150, completionTokens: 200}
      endTime: "2024-01-15T10:30:00Z"
    }
  ) {
    id
    totalTokens
    calculatedCost
  }
}

# 4. Add a score
mutation {
  createScore(input: {
    traceId: "trace-id-from-step-1"
    name: "quality"
    value: 0.95
    comment: "Thorough review with actionable suggestions"
  }) {
    id
  }
}

# 5. Update trace with output
mutation {
  updateTrace(
    id: "trace-id-from-step-1"
    input: {
      output: {review: "Approved with suggestions"}
    }
  ) {
    id
    totalCost
    latency
  }
}
```

## Error Handling

GraphQL errors are returned in the standard format:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Trace not found: abc123",
      "path": ["trace"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ]
}
```

Common error codes:

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Invalid or missing API key |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource doesn't exist |
| `VALIDATION_ERROR` | Invalid input |
| `INTERNAL_ERROR` | Server error |

## Best Practices

1. **Use fragments** for reusable field selections
2. **Batch queries** to reduce round trips
3. **Use variables** instead of string interpolation
4. **Request only needed fields** to minimize response size
5. **Handle errors gracefully** with proper error checking

## Rate Limiting

GraphQL requests count toward your API rate limit. Complex queries may count as multiple requests based on complexity scoring.

## SDKs

The official SDKs use the REST API by default. For GraphQL, use any standard GraphQL client:

```typescript
import { GraphQLClient } from 'graphql-request';

const client = new GraphQLClient('https://api.agenttrace.io/graphql', {
  headers: {
    Authorization: `Bearer ${process.env.AGENTTRACE_API_KEY}`,
  },
});

const data = await client.request(query, variables);
```
