package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes
const (
	CodeInternal       = "INTERNAL_ERROR"
	CodeNotFound       = "NOT_FOUND"
	CodeValidation     = "VALIDATION_ERROR"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeConflict       = "CONFLICT"
	CodeRateLimited    = "RATE_LIMITED"
	CodeBadRequest     = "BAD_REQUEST"
	CodeUnprocessable  = "UNPROCESSABLE_ENTITY"
)

// AppError represents an application error with context
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	StatusCode int               `json:"-"`
	Err        error             `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail adds a detail to the error
func (e *AppError) WithDetail(key, value string) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Internal creates an internal server error
func Internal(message string) *AppError {
	return New(CodeInternal, message, http.StatusInternalServerError)
}

// NotFound creates a not found error
func NotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// Validation creates a validation error
func Validation(message string) *AppError {
	return New(CodeValidation, message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "unauthorized"
	}
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	if message == "" {
		message = "forbidden"
	}
	return New(CodeForbidden, message, http.StatusForbidden)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	return New(CodeConflict, message, http.StatusConflict)
}

// RateLimited creates a rate limited error
func RateLimited() *AppError {
	return New(CodeRateLimited, "rate limit exceeded", http.StatusTooManyRequests)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message, http.StatusBadRequest)
}

// Unprocessable creates an unprocessable entity error
func Unprocessable(message string) *AppError {
	return New(CodeUnprocessable, message, http.StatusUnprocessableEntity)
}

// Is checks if an error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As attempts to convert an error to a specific type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// IsAppError checks if the error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error if present
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == CodeNotFound
	}
	return false
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == CodeValidation
	}
	return false
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == CodeUnauthorized
	}
	return false
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == CodeForbidden
	}
	return false
}

// IsConflict checks if the error is a conflict error
func IsConflict(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == CodeConflict
	}
	return false
}

// IsRateLimited checks if the error is a rate limited error
func IsRateLimited(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == CodeRateLimited
	}
	return false
}
