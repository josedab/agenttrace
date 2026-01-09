package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization
type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Related data (populated by resolver)
	Members  []OrganizationMember `json:"members,omitempty"`
	Projects []Project            `json:"projects,omitempty"`
}

// OrganizationInput represents input for creating an organization
type OrganizationInput struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug,omitempty" validate:"omitempty,min=2,max=100,alphanumunicode"`
}

// OrganizationUpdateInput represents input for updating an organization
type OrganizationUpdateInput struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	UserID         uuid.UUID `json:"userId"`
	Role           OrgRole   `json:"role"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

	// Related data (populated by resolver)
	User         *User         `json:"user,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
}

// OrganizationMemberInput represents input for adding/updating a member
type OrganizationMemberInput struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	Role   OrgRole   `json:"role" validate:"required"`
}

// OrganizationInvitation represents an invitation to join an organization
type OrganizationInvitation struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	Email          string     `json:"email"`
	Role           OrgRole    `json:"role"`
	InvitedBy      uuid.UUID  `json:"invitedBy"`
	Token          string     `json:"-"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`

	// Related data (populated by resolver)
	Organization *Organization `json:"organization,omitempty"`
	InvitedByUser *User        `json:"invitedByUser,omitempty"`
}

// OrganizationInvitationInput represents input for creating an invitation
type OrganizationInvitationInput struct {
	Email string  `json:"email" validate:"required,email"`
	Role  OrgRole `json:"role" validate:"required"`
}

// OrganizationFilter represents filter options for querying organizations
type OrganizationFilter struct {
	UserID uuid.UUID
}

// OrganizationList represents a paginated list of organizations
type OrganizationList struct {
	Organizations []Organization `json:"organizations"`
	TotalCount    int64          `json:"totalCount"`
	HasMore       bool           `json:"hasMore"`
}

// OrganizationStats represents organization statistics
type OrganizationStats struct {
	TotalProjects int64   `json:"totalProjects"`
	TotalMembers  int64   `json:"totalMembers"`
	TotalTraces   int64   `json:"totalTraces"`
	TotalCost     float64 `json:"totalCost"`
}

// GenerateSlug generates a URL-safe slug from a name
func GenerateSlug(name string) string {
	// Simple slug generation - replace spaces with hyphens, lowercase
	slug := ""
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			slug += string(r)
		} else if r >= 'A' && r <= 'Z' {
			slug += string(r + 32) // lowercase
		} else if r >= '0' && r <= '9' {
			slug += string(r)
		} else if r == ' ' || r == '-' || r == '_' {
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug += "-"
			}
		}
	}
	// Trim trailing hyphens
	for len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}
	return slug
}
