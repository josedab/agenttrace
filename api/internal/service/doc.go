// Package service contains the business logic layer for AgentTrace.
//
// Services coordinate between handlers and repositories, implementing
// domain rules and orchestrating operations across multiple repositories.
//
// Services depend on repository interfaces defined in this package,
// following the dependency inversion principle. Each service typically
// handles a specific domain area (traces, scores, prompts, etc.).
//
// # Architecture
//
// The service layer sits between:
//   - HTTP handlers (presentation layer)
//   - Repository implementations (data access layer)
//
// Services are responsible for:
//   - Business logic and validation
//   - Orchestrating multiple repository calls
//   - Transaction coordination where needed
//   - Domain event publishing
//
// # Thread Safety
//
// All services are designed to be safe for concurrent use from
// multiple goroutines.
package service
