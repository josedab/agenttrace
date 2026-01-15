// Package validator provides struct validation for AgentTrace.
//
// This package wraps go-playground/validator to provide:
//   - Consistent validation across all handlers
//   - Human-readable error messages
//   - Structured validation error responses
//
// # Usage
//
// Use validator.Validate() directly or through dto.ParseAndValidate():
//
//	if err := validator.Validate(myStruct); err != nil {
//	    // err is a validator.ValidationErrors
//	}
//
// # Custom Validations
//
// Custom validations can be registered in the init() function.
// The validator instance is package-level and thread-safe.
package validator
