// Package dto contains Data Transfer Objects for HTTP request/response handling.
//
// DTOs provide:
//   - Type-safe request parsing with struct tags
//   - Declarative validation using go-playground/validator
//   - Separation between API contracts and domain types
//
// # Usage
//
// Use dto.ParseAndValidate() in handlers to parse and validate requests:
//
//	var req dto.LoginRequest
//	if err := dto.ParseAndValidate(c, &req); err != nil {
//	    return err
//	}
//
// # Validation Tags
//
// Common validation tags:
//   - required: Field must be present and non-empty
//   - email: Must be valid email format
//   - min=N: Minimum length/value
//   - max=N: Maximum length/value
//   - uuid: Must be valid UUID format
//   - oneof=A B C: Must be one of the specified values
package dto
