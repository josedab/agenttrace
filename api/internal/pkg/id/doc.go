// Package id provides identifier generation for AgentTrace.
//
// This package generates:
//   - W3C-compliant trace IDs (32 hex characters)
//   - W3C-compliant span IDs (16 hex characters)
//   - UUID v4 identifiers
//   - API keys (public and secret)
//
// # Performance
//
// ID generation uses sync.Pool to minimize allocations in hot paths.
// All functions are safe for concurrent use.
//
// # Validation
//
// The package includes validators for all ID formats:
//
//	if !id.ValidateTraceID(traceID) {
//	    return errors.New("invalid trace ID")
//	}
//
// # API Keys
//
// API keys use prefixes for easy identification:
//   - pk-at-* : Public keys (client-safe)
//   - sk-at-* : Secret keys (server-only)
package id
