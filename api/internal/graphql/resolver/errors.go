package resolver

import "errors"

// Resolver errors
var (
	// ErrProjectIDNotFound is returned when project ID is not found in context
	ErrProjectIDNotFound = errors.New("project ID not found in context")
	// ErrUserIDNotFound is returned when user ID is not found in context
	ErrUserIDNotFound = errors.New("user ID not found in context")
)
