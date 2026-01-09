package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a project within an organization
type Project struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	Settings       string    `json:"settings,omitempty"`
	RetentionDays  int       `json:"retentionDays"`
	RateLimitPerMin *int     `json:"rateLimitPerMinute,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

	// Related data (populated by resolver)
	Organization *Organization    `json:"organization,omitempty"`
	Members      []ProjectMember  `json:"members,omitempty"`
	APIKeys      []APIKey         `json:"apiKeys,omitempty"`
}

// ProjectInput represents input for creating a project
type ProjectInput struct {
	Name           string  `json:"name" validate:"required,min=2,max=100"`
	Slug           string  `json:"slug,omitempty" validate:"omitempty,min=2,max=100"`
	Description    *string `json:"description,omitempty"`
	RetentionDays  *int    `json:"retentionDays,omitempty" validate:"omitempty,min=1,max=365"`
	RateLimitPerMin *int   `json:"rateLimitPerMinute,omitempty" validate:"omitempty,min=1"`
}

// ProjectUpdateInput represents input for updating a project
type ProjectUpdateInput struct {
	Name           *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description    *string `json:"description,omitempty"`
	RetentionDays  *int    `json:"retentionDays,omitempty" validate:"omitempty,min=1,max=365"`
	RateLimitPerMin *int   `json:"rateLimitPerMinute,omitempty" validate:"omitempty,min=1"`
	Settings       any     `json:"settings,omitempty"`
}

// ProjectMember represents a member with project-level role override
type ProjectMember struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	UserID    uuid.UUID `json:"userId"`
	Role      OrgRole   `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Related data (populated by resolver)
	User    *User    `json:"user,omitempty"`
	Project *Project `json:"project,omitempty"`
}

// ProjectMemberInput represents input for adding/updating a project member
type ProjectMemberInput struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	Role   OrgRole   `json:"role" validate:"required"`
}

// ProjectFilter represents filter options for querying projects
type ProjectFilter struct {
	OrganizationID uuid.UUID
	UserID         *uuid.UUID
}

// ProjectList represents a paginated list of projects
type ProjectList struct {
	Projects   []Project `json:"projects"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}

// ProjectStats represents project statistics
type ProjectStats struct {
	TraceCount       int64   `json:"traceCount"`
	ObservationCount int64   `json:"observationCount"`
	TotalCost        float64 `json:"totalCost"`
	TotalTokens      int64   `json:"totalTokens"`
	UniqueUsers      int64   `json:"uniqueUsers"`
	UniqueSessions   int64   `json:"uniqueSessions"`
}

// ProjectSettings represents project-specific settings
type ProjectSettings struct {
	// Feature flags
	EnableEvaluations bool `json:"enableEvaluations"`
	EnablePrompts     bool `json:"enablePrompts"`
	EnableDatasets    bool `json:"enableDatasets"`
	EnableExports     bool `json:"enableExports"`

	// Default values
	DefaultRetentionDays int `json:"defaultRetentionDays"`

	// Integrations
	SlackWebhookURL    string `json:"slackWebhookUrl,omitempty"`
	PagerDutyAPIKey    string `json:"pagerDutyApiKey,omitempty"`
}
