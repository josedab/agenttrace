// Package domain contains the core business entities and types for AgentTrace.
//
// This package defines:
//   - Entity types (Trace, Observation, Score, etc.)
//   - Value objects and enums
//   - Input/output types for service operations
//   - Domain-level validation rules
//
// # Design Philosophy
//
// Domain types are persistence-agnostic and represent the core
// business concepts independent of how they are stored or transmitted.
//
// # Key Entities
//
//   - Trace: A complete execution trace, the root of the trace hierarchy
//   - Observation: Events within a trace (spans, generations, events)
//   - Score: Quality metrics attached to traces or observations
//   - Session: Groups related traces together
//   - Prompt: Versioned prompt templates
//   - Dataset: Collections of data items for evaluation
//
// # Naming Conventions
//
// Types ending in "Input" are used for create/update operations.
// Types ending in "Filter" are used for query operations.
package domain
