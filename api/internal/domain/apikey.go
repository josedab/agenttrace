package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key for a project
type APIKey struct {
	ID               uuid.UUID  `json:"id"`
	ProjectID        uuid.UUID  `json:"projectId"`
	Name             string     `json:"name"`
	PublicKey        string     `json:"publicKey"`
	SecretKeyHash    string     `json:"-"`
	SecretKeyPreview string     `json:"secretKeyPreview"`
	Scopes           []string   `json:"scopes"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt       *time.Time `json:"lastUsedAt,omitempty"`
	CreatedBy        *uuid.UUID `json:"createdBy,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// APIKeyInput represents input for creating an API key
type APIKeyInput struct {
	Name      string     `json:"name" validate:"required,min=1,max=100"`
	Scopes    []string   `json:"scopes,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// APIKeyCreateResult represents the result of creating an API key
type APIKeyCreateResult struct {
	APIKey    *APIKey `json:"apiKey"`
	SecretKey string  `json:"secretKey"`
}

// APIKeyFilter represents filter options for querying API keys
type APIKeyFilter struct {
	ProjectID uuid.UUID
	CreatedBy *uuid.UUID
}

// APIKeyList represents a paginated list of API keys
type APIKeyList struct {
	APIKeys    []APIKey `json:"apiKeys"`
	TotalCount int64    `json:"totalCount"`
	HasMore    bool     `json:"hasMore"`
}

// DefaultScopes returns the default API key scopes
func DefaultScopes() []string {
	return []string{
		"traces:write",
		"traces:read",
		"observations:write",
		"observations:read",
		"scores:write",
		"scores:read",
		"prompts:read",
	}
}

// AllScopes returns all available API key scopes
func AllScopes() []string {
	return []string{
		// Trace operations
		"traces:write",
		"traces:read",
		"traces:delete",

		// Observation operations
		"observations:write",
		"observations:read",

		// Score operations
		"scores:write",
		"scores:read",
		"scores:delete",

		// Prompt operations
		"prompts:read",
		"prompts:write",
		"prompts:delete",

		// Dataset operations
		"datasets:read",
		"datasets:write",
		"datasets:delete",

		// Evaluator operations
		"evaluators:read",
		"evaluators:write",
		"evaluators:delete",

		// Metrics operations
		"metrics:read",

		// Export operations
		"exports:read",
		"exports:write",

		// Admin operations
		"admin:read",
		"admin:write",
	}
}

// HasScope checks if the API key has a specific scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "admin:write" {
			return true
		}
		// Check for wildcard scope
		if len(s) > 1 && s[len(s)-1] == '*' {
			prefix := s[:len(s)-1]
			if len(scope) >= len(prefix) && scope[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// HasAnyScope checks if the API key has any of the specified scopes
func (k *APIKey) HasAnyScope(scopes ...string) bool {
	for _, scope := range scopes {
		if k.HasScope(scope) {
			return true
		}
	}
	return false
}

// IsExpired checks if the API key has expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// APIKeyContext represents the context extracted from an API key
type APIKeyContext struct {
	APIKeyID  uuid.UUID
	ProjectID uuid.UUID
	Scopes    []string
}

// HasScope checks if the context has a specific scope
func (c *APIKeyContext) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope || s == "admin:write" {
			return true
		}
	}
	return false
}
