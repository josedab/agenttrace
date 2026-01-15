package dto

import "github.com/agenttrace/agenttrace/api/internal/domain"

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	OrganizationID  string                  `json:"organizationId" validate:"required,uuid"`
	Name            string                  `json:"name" validate:"required"`
	Description     string                  `json:"description,omitempty"`
	Settings        *domain.ProjectSettings `json:"settings,omitempty"`
	RetentionDays   *int                    `json:"retentionDays,omitempty"`
	RateLimitPerMin *int                    `json:"rateLimitPerMin,omitempty"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name            string                  `json:"name,omitempty"`
	Description     string                  `json:"description,omitempty"`
	Settings        *domain.ProjectSettings `json:"settings,omitempty"`
	RetentionDays   *int                    `json:"retentionDays,omitempty"`
	RateLimitPerMin *int                    `json:"rateLimitPerMin,omitempty"`
}

// AddMemberRequest represents the request to add a member to a project
type AddMemberRequest struct {
	UserID string         `json:"userId" validate:"required,uuid"`
	Role   domain.OrgRole `json:"role" validate:"required"`
}
