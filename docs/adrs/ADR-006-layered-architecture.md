# ADR-006: Layered Architecture Pattern

## Status

Accepted

## Context

As AgentTrace grows in complexity, we need a consistent code organization pattern that:

1. Separates concerns clearly (HTTP handling, business logic, data access)
2. Enables unit testing at each layer independently
3. Allows swapping implementations (e.g., different databases)
4. Makes the codebase navigable for new team members
5. Prevents tight coupling between components

The Go ecosystem has various patterns: flat structures, domain-driven design, hexagonal architecture, and classic layered architecture.

### Alternatives Considered

1. **Flat structure**
   - Pros: Simple, no indirection
   - Cons: Doesn't scale, business logic mixed with HTTP handling

2. **Hexagonal architecture (ports & adapters)**
   - Pros: Maximum flexibility, easy to test
   - Cons: More abstractions, higher learning curve, verbose for CRUD operations

3. **Domain-driven design (DDD)**
   - Pros: Rich domain models, aggregates, value objects
   - Cons: Overhead for observability platform (mostly CRUD + queries)

4. **Classic layered architecture** (chosen)
   - Pros: Well-understood, clear data flow, testable, appropriate complexity
   - Cons: Can lead to anemic models, potential for layer violations

## Decision

We use a **classic three-layer architecture** with clear separation:

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer                              │
│  api/internal/handler/                                       │
│  - Fiber route handlers                                      │
│  - Request parsing, validation                               │
│  - Response formatting                                       │
│  - Authentication/authorization checks                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│  api/internal/service/                                       │
│  - Business logic                                            │
│  - Orchestration across repositories                         │
│  - Transaction management                                    │
│  - Domain validation                                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│  api/internal/repository/                                    │
│  - Data access abstractions                                  │
│  - Database-specific implementations                         │
│  - Query building                                            │
│  - Connection management                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                              │
│  api/internal/domain/                                        │
│  - Entity definitions                                        │
│  - Value objects                                             │
│  - Domain constants                                          │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
api/internal/
├── handler/              # HTTP layer
│   ├── trace_handler.go
│   ├── prompt_handler.go
│   ├── evaluation_handler.go
│   └── ...
├── service/              # Business logic layer
│   ├── trace_service.go
│   ├── prompt_service.go
│   ├── evaluation_service.go
│   └── ...
├── repository/           # Data access layer
│   ├── interfaces.go     # Repository interfaces
│   ├── postgres/         # PostgreSQL implementations
│   │   ├── prompt_repository.go
│   │   └── ...
│   └── clickhouse/       # ClickHouse implementations
│       ├── trace_repository.go
│       └── ...
├── domain/               # Domain models
│   ├── trace.go
│   ├── prompt.go
│   └── ...
├── middleware/           # Cross-cutting concerns
│   ├── auth.go
│   ├── logging.go
│   └── ...
└── pkg/                  # Shared utilities
    ├── errors/
    └── validation/
```

### Layer Responsibilities

**Handler Layer:**
```go
// trace_handler.go
func (h *TraceHandler) Create(c *fiber.Ctx) error {
    // 1. Parse request
    var input domain.TraceInput
    if err := c.BodyParser(&input); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }

    // 2. Validate
    if err := h.validator.Struct(input); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, err.Error())
    }

    // 3. Delegate to service
    trace, err := h.traceService.Create(c.Context(), input)
    if err != nil {
        return err
    }

    // 4. Format response
    return c.Status(fiber.StatusCreated).JSON(trace)
}
```

**Service Layer:**
```go
// trace_service.go
func (s *TraceService) Create(ctx context.Context, input domain.TraceInput) (*domain.Trace, error) {
    // 1. Business validation
    if err := s.validateTraceInput(input); err != nil {
        return nil, err
    }

    // 2. Enrich with computed fields
    trace := &domain.Trace{
        ID:        uuid.New(),
        ProjectID: input.ProjectID,
        Name:      input.Name,
        CreatedAt: time.Now(),
    }

    // 3. Persist via repository
    if err := s.traceRepo.Create(ctx, trace); err != nil {
        return nil, err
    }

    // 4. Side effects (async)
    s.eventPublisher.Publish(TraceCreatedEvent{Trace: trace})

    return trace, nil
}
```

**Repository Layer:**
```go
// Repository interface (in interfaces.go)
type TraceRepository interface {
    Create(ctx context.Context, trace *domain.Trace) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Trace, error)
    List(ctx context.Context, filter TraceFilter) ([]domain.Trace, error)
}

// ClickHouse implementation
func (r *clickhouseTraceRepo) Create(ctx context.Context, trace *domain.Trace) error {
    query := `INSERT INTO traces (id, project_id, name, created_at) VALUES (?, ?, ?, ?)`
    return r.conn.Exec(ctx, query, trace.ID, trace.ProjectID, trace.Name, trace.CreatedAt)
}
```

## Consequences

### Positive

- **Clear responsibilities**: Each layer has a single purpose
- **Testability**: Mock repositories for service tests, mock services for handler tests
- **Database abstraction**: Repository interfaces allow swapping implementations
- **Onboarding**: New developers understand data flow immediately
- **Maintainability**: Changes isolated to specific layers

### Negative

- **Boilerplate**: CRUD operations require touching all layers
- **Potential anemia**: Risk of thin services that just pass through to repositories
- **Layer violations**: Discipline required to prevent handlers calling repositories directly
- **Indirection**: More function calls for simple operations

### Neutral

- GraphQL resolvers map to services (same layer as REST handlers)
- Workers follow same pattern (worker → service → repository)
- Middleware applies at handler layer

## Testing Strategy

```go
// Service tests mock the repository
func TestTraceService_Create(t *testing.T) {
    mockRepo := &MockTraceRepository{}
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    service := NewTraceService(mockRepo)
    trace, err := service.Create(ctx, input)

    assert.NoError(t, err)
    assert.NotNil(t, trace)
    mockRepo.AssertExpectations(t)
}

// Handler tests mock the service
func TestTraceHandler_Create(t *testing.T) {
    mockService := &MockTraceService{}
    mockService.On("Create", mock.Anything, mock.Anything).Return(&domain.Trace{}, nil)

    handler := NewTraceHandler(mockService)
    // ... test HTTP request/response
}
```

## References

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Domain-Driven Design in Go](https://threedots.tech/post/ddd-lite-in-go-introduction/)
