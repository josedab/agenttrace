// Package handler contains HTTP request handlers for AgentTrace.
//
// Handlers are the entry point for HTTP requests, responsible for:
//   - Request parsing and validation
//   - Authentication context extraction
//   - Calling appropriate services
//   - Response formatting
//   - Error response mapping
//
// # Route Organization
//
// Routes are organized by resource:
//   - /api/public/* - Public API routes (API key authentication)
//   - /api/v1/* - Internal API routes (JWT authentication)
//   - /api/auth/* - Authentication routes (no auth required)
//
// # Error Handling
//
// Handlers convert domain errors to appropriate HTTP status codes
// using the apperrors package for consistent error responses.
//
// # Thread Safety
//
// All handlers are safe for concurrent use.
package handler
