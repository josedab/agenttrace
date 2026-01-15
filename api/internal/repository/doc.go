// Package repository contains data access implementations for AgentTrace.
//
// Repositories provide persistence operations for domain entities,
// abstracting the underlying data stores (PostgreSQL, ClickHouse, Redis).
//
// # Architecture
//
// Repository interfaces are defined at the service layer (consumer-defined
// interfaces) following Go's dependency inversion best practices.
// This package contains the concrete implementations.
//
// # Data Stores
//
// The system uses multiple specialized data stores:
//   - PostgreSQL: Transactional data (users, projects, organizations)
//   - ClickHouse: Analytics and time-series data (traces, observations)
//   - Redis: Caching and session storage
//
// # Thread Safety
//
// All repository implementations are safe for concurrent use.
// Connection pools are managed at the database layer.
package repository
