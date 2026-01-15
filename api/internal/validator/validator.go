package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// V is the singleton validator instance
var V *validator.Validate

func init() {
	V = validator.New()

	// Register custom validations as needed
	// Example: V.RegisterValidation("custom", customValidationFunc)
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// Validate validates a struct and returns ValidationErrors if invalid
func Validate(v any) error {
	if err := V.Struct(v); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors converts validator errors to ValidationErrors
func formatValidationErrors(err error) ValidationErrors {
	var validationErrors ValidationErrors

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   toJSONFieldName(e.Field()),
				Message: getErrorMessage(e),
			})
		}
	}

	return validationErrors
}

// toJSONFieldName converts struct field name to JSON field name (camelCase)
func toJSONFieldName(field string) string {
	if len(field) == 0 {
		return field
	}
	// Convert first character to lowercase for camelCase
	return strings.ToLower(field[:1]) + field[1:]
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		if e.Type().Kind().String() == "string" {
			return fmt.Sprintf("must be at least %s characters", e.Param())
		}
		return fmt.Sprintf("must be at least %s", e.Param())
	case "max":
		if e.Type().Kind().String() == "string" {
			return fmt.Sprintf("must be at most %s characters", e.Param())
		}
		return fmt.Sprintf("must be at most %s", e.Param())
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", e.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", e.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", e.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", e.Param())
	default:
		return fmt.Sprintf("failed validation: %s", e.Tag())
	}
}

// IsValidationError checks if an error is a ValidationErrors
func IsValidationError(err error) bool {
	_, ok := err.(ValidationErrors)
	return ok
}
