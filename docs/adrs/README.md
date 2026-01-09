# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the AgentTrace platform. ADRs document significant architectural decisions made during the project's development.

## What is an ADR?

An Architecture Decision Record captures an important architectural decision along with its context and consequences. They help team members understand why the system is built the way it is.

## ADR Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-001](./ADR-001-go-fiber-backend.md) | Go + Fiber Backend Architecture | Accepted | 2024-01 |
| [ADR-002](./ADR-002-dual-database-strategy.md) | Dual Database: PostgreSQL + ClickHouse | Accepted | 2024-01 |
| [ADR-003](./ADR-003-agent-first-schema.md) | Agent-First Schema Design | Accepted | 2024-01 |
| [ADR-004](./ADR-004-multi-sdk-strategy.md) | Multi-SDK Strategy with Idiomatic Patterns | Accepted | 2024-01 |
| [ADR-005](./ADR-005-asynq-background-jobs.md) | Asynq for Background Job Processing | Accepted | 2024-01 |
| [ADR-006](./ADR-006-layered-architecture.md) | Layered Architecture Pattern | Accepted | 2024-01 |
| [ADR-007](./ADR-007-hybrid-graphql-rest.md) | Hybrid GraphQL + REST API | Accepted | 2024-01 |
| [ADR-008](./ADR-008-dual-authentication.md) | API Key + JWT Dual Authentication | Accepted | 2024-01 |
| [ADR-009](./ADR-009-async-cost-calculation.md) | Asynchronous Cost Calculation | Accepted | 2024-01 |
| [ADR-010](./ADR-010-nextjs-bff.md) | Next.js BFF with Server Components | Accepted | 2024-01 |

## ADR Template

When creating new ADRs, use the following template:

```markdown
# ADR-NNNN: Title

## Status

Accepted | Superseded | Deprecated

## Context

What prompted this decision? What problem were we solving?

## Decision

What was decided?

## Consequences

### Positive
- Benefits and improvements enabled

### Negative
- Tradeoffs and limitations introduced

### Neutral
- Other implications to be aware of
```

## Contributing

When making significant architectural changes:

1. Create a new ADR with the next sequential number
2. Document the context, decision, and consequences
3. Update the index in this README
4. Include the ADR in your pull request
