package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user role in the system
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleViewer    Role = "viewer"
	RoleAuditor   Role = "auditor"
)

// Permission represents a granular permission
type Permission string

const (
	PermTraceRead       Permission = "trace:read"
	PermTraceWrite      Permission = "trace:write"
	PermPromptManage    Permission = "prompt:manage"
	PermEvalManage      Permission = "eval:manage"
	PermSettingsManage  Permission = "settings:manage"
	PermBillingManage   Permission = "billing:manage"
	PermAuditRead       Permission = "audit:read"
)

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermTraceRead, PermTraceWrite, PermPromptManage,
		PermEvalManage, PermSettingsManage, PermBillingManage, PermAuditRead,
	},
	RoleDeveloper: {
		PermTraceRead, PermTraceWrite, PermPromptManage, PermEvalManage,
	},
	RoleViewer: {
		PermTraceRead,
	},
	RoleAuditor: {
		PermTraceRead, PermAuditRead,
	},
}

// RoleAssignment represents a role assignment for a user
type RoleAssignment struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ProjectID uuid.UUID `json:"projectId"`
	Role      Role      `json:"role"`
	GrantedBy uuid.UUID `json:"grantedBy"`
	GrantedAt time.Time `json:"grantedAt"`
}

// SSOConfig represents SSO configuration for an organization
type SSOConfig struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"orgId"`
	Provider      string    `json:"provider"` // "saml" or "oidc"
	IssuerURL     string    `json:"issuerUrl"`
	ClientID      string    `json:"clientId"`
	ClientSecret  string    `json:"clientSecret,omitempty"`
	Enabled       bool      `json:"enabled"`
	AutoProvision bool      `json:"autoProvision"`
	DefaultRole   Role      `json:"defaultRole"`
	CreatedAt     time.Time `json:"createdAt"`
}

// APIKeyScope represents scoped permissions for an API key
type APIKeyScope struct {
	ID            uuid.UUID    `json:"id"`
	APIKeyID      uuid.UUID    `json:"apiKeyId"`
	Permissions   []Permission `json:"permissions"`
	ResourceTypes []string     `json:"resourceTypes"`
	CreatedAt     time.Time    `json:"createdAt"`
}

// SSOConfigInput represents input for configuring SSO
type SSOConfigInput struct {
	Provider      string `json:"provider"`
	IssuerURL     string `json:"issuerUrl"`
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	Enabled       bool   `json:"enabled"`
	AutoProvision bool   `json:"autoProvision"`
	DefaultRole   Role   `json:"defaultRole"`
}

// RoleAssignmentInput represents input for assigning a role
type RoleAssignmentInput struct {
	UserID    uuid.UUID `json:"userId"`
	ProjectID uuid.UUID `json:"projectId"`
	Role      Role      `json:"role"`
}

// APIKeyScopeInput represents input for scoping an API key
type APIKeyScopeInput struct {
	APIKeyID      uuid.UUID    `json:"apiKeyId"`
	Permissions   []Permission `json:"permissions"`
	ResourceTypes []string     `json:"resourceTypes"`
}
