# ADR-007: Hybrid GraphQL + REST API Design

## Status

Accepted

## Context

AgentTrace serves diverse API consumers with different needs:

1. **SDKs (Python, TypeScript, Go)**: Need simple, fast endpoints for trace ingestion
2. **Web dashboard**: Needs flexible queries for complex nested data (traces with observations, scores)
3. **Integrations**: Third-party tools expect REST conventions
4. **Power users**: Want to build custom queries without waiting for new endpoints

Different query patterns have different optimal API styles:
- Simple CRUD and ingestion → REST
- Complex nested queries → GraphQL
- Real-time subscriptions → GraphQL/WebSocket

### Alternatives Considered

1. **REST only**
   - Pros: Simple, cacheable, widely understood, great tooling
   - Cons: Over-fetching, N+1 on nested resources, endpoint proliferation

2. **GraphQL only**
   - Pros: Flexible queries, single endpoint, strong typing
   - Cons: Caching complexity, learning curve, overkill for simple operations

3. **gRPC**
   - Pros: High performance, strong typing, streaming
   - Cons: Browser support requires proxy, steeper learning curve

4. **Hybrid REST + GraphQL** (chosen)
   - Pros: Right tool for each job, incremental adoption, satisfies all consumers
   - Cons: Two API surfaces to maintain, documentation complexity

## Decision

We provide **both REST and GraphQL APIs**, each optimized for its use case:

### REST API (`/api/public/`)

Used for:
- Trace/observation ingestion (high throughput)
- Simple CRUD operations
- SDK operations
- Health checks and metrics
- Webhook callbacks

```
POST   /api/public/traces              # Ingest trace
POST   /api/public/observations        # Ingest observation
POST   /api/public/scores              # Add score
GET    /api/public/prompts/:name       # Get prompt by name
POST   /api/public/prompts/:name/compile  # Compile prompt
GET    /api/public/health              # Health check
GET    /api/public/metrics             # Prometheus metrics
```

### GraphQL API (`/graphql`)

Used for:
- Complex nested queries (dashboard)
- Filtering and pagination
- Custom client queries
- Batch operations

```graphql
query GetTraceWithDetails($id: ID!) {
  trace(id: $id) {
    id
    name
    startTime
    endTime
    latencyMs
    totalCost
    observations {
      id
      type
      name
      model
      tokens {
        input
        output
      }
      scores {
        name
        value
      }
    }
    gitLinks {
      commitSha
      repository
    }
  }
}

query ListTracesWithFilters($filter: TraceFilter!, $pagination: Pagination!) {
  traces(filter: $filter, pagination: $pagination) {
    edges {
      node {
        id
        name
        totalCost
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

### API Routing

```go
// routes.go
func SetupRoutes(app *fiber.App) {
    // REST API
    api := app.Group("/api/public")
    api.Post("/traces", traceHandler.Create)
    api.Post("/observations", observationHandler.Create)
    api.Get("/prompts/:name", promptHandler.GetByName)
    // ...

    // GraphQL API
    app.All("/graphql", gqlHandler)
    app.Get("/graphql/playground", playgroundHandler)
}
```

## Consequences

### Positive

- **Optimized for use case**: REST for ingestion speed, GraphQL for query flexibility
- **Incremental adoption**: Teams can use REST first, adopt GraphQL as needed
- **Dashboard efficiency**: GraphQL reduces round trips for complex views
- **SDK simplicity**: SDKs use straightforward REST endpoints
- **Schema documentation**: GraphQL introspection provides self-documenting API

### Negative

- **Dual maintenance**: Two API surfaces to maintain, test, and version
- **Documentation overhead**: Need to document both APIs clearly
- **Consistency risk**: Must keep REST and GraphQL responses aligned
- **Tooling complexity**: Different testing approaches for each API

### Neutral

- GraphQL implemented via gqlgen (code generation)
- REST follows OpenAPI specification
- Both APIs share the same service layer

## Implementation Details

### GraphQL with gqlgen

```go
// gqlgen.yml
schema:
  - graph/*.graphqls

exec:
  filename: graph/generated.go
  package: graph

model:
  filename: graph/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: graph/resolver
```

### Dataloaders for N+1 Prevention

```go
// Prevents N+1 queries on nested resources
type Loaders struct {
    ObservationsByTraceID *dataloader.Loader[uuid.UUID, []domain.Observation]
    ScoresByObservationID *dataloader.Loader[uuid.UUID, []domain.Score]
}

func (r *traceResolver) Observations(ctx context.Context, obj *domain.Trace) ([]domain.Observation, error) {
    return r.loaders.ObservationsByTraceID.Load(ctx, obj.ID)
}
```

### REST OpenAPI Specification

```yaml
openapi: 3.0.3
info:
  title: AgentTrace API
  version: 1.0.0

paths:
  /api/public/traces:
    post:
      summary: Ingest a trace
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TraceInput'
      responses:
        '201':
          description: Trace created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Trace'
```

## When to Use Which API

| Use Case | API | Rationale |
|----------|-----|-----------|
| Trace ingestion | REST | Speed, simplicity |
| SDK operations | REST | Minimal dependencies |
| Dashboard queries | GraphQL | Flexible nested data |
| Custom analytics | GraphQL | Ad-hoc queries |
| Webhook callbacks | REST | Standard convention |
| Health/metrics | REST | Prometheus compatibility |
| Batch mutations | GraphQL | Single request |

## References

- [gqlgen Documentation](https://gqlgen.com/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [REST vs GraphQL](https://www.apollographql.com/blog/graphql/basics/graphql-vs-rest/)
