// Package errors provides application error types for AgentTrace.
//
// This package defines:
//   - AppError type with error classification
//   - Error constructors for common error types
//   - Error type checking helpers
//   - HTTP status code mapping
//
// # Error Types
//
//   - NotFound: Resource does not exist (404)
//   - Validation: Invalid input data (400)
//   - Unauthorized: Authentication required (401)
//   - Forbidden: Insufficient permissions (403)
//   - Internal: Unexpected server error (500)
//
// # Usage
//
// Create errors using constructor functions:
//
//	return apperrors.NotFound("user", userID)
//	return apperrors.Validation("email is required")
//
// Check error types:
//
//	if apperrors.IsNotFound(err) {
//	    // Handle not found
//	}
//
// # Error Wrapping
//
// Errors support wrapping with fmt.Errorf:
//
//	return fmt.Errorf("operation failed: %w", apperrors.NotFound("item", id))
package errors
