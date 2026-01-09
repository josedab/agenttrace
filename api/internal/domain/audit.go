package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	// Authentication actions
	AuditActionLogin          AuditAction = "login"
	AuditActionLogout         AuditAction = "logout"
	AuditActionLoginFailed    AuditAction = "login_failed"
	AuditActionSSOLogin       AuditAction = "sso_login"
	AuditActionAPIKeyUsed     AuditAction = "api_key_used"

	// User management
	AuditActionUserCreated    AuditAction = "user_created"
	AuditActionUserUpdated    AuditAction = "user_updated"
	AuditActionUserDeleted    AuditAction = "user_deleted"
	AuditActionUserInvited    AuditAction = "user_invited"
	AuditActionUserRoleChanged AuditAction = "user_role_changed"

	// Organization management
	AuditActionOrgCreated     AuditAction = "org_created"
	AuditActionOrgUpdated     AuditAction = "org_updated"
	AuditActionOrgDeleted     AuditAction = "org_deleted"
	AuditActionMemberAdded    AuditAction = "member_added"
	AuditActionMemberRemoved  AuditAction = "member_removed"

	// Project management
	AuditActionProjectCreated AuditAction = "project_created"
	AuditActionProjectUpdated AuditAction = "project_updated"
	AuditActionProjectDeleted AuditAction = "project_deleted"

	// API Key management
	AuditActionAPIKeyCreated  AuditAction = "api_key_created"
	AuditActionAPIKeyRevoked  AuditAction = "api_key_revoked"

	// SSO configuration
	AuditActionSSOConfigured  AuditAction = "sso_configured"
	AuditActionSSOEnabled     AuditAction = "sso_enabled"
	AuditActionSSODisabled    AuditAction = "sso_disabled"

	// Data access
	AuditActionDataExported   AuditAction = "data_exported"
	AuditActionDataDeleted    AuditAction = "data_deleted"

	// Settings changes
	AuditActionSettingsChanged AuditAction = "settings_changed"

	// Prompt management
	AuditActionPromptCreated   AuditAction = "prompt_created"
	AuditActionPromptUpdated   AuditAction = "prompt_updated"
	AuditActionPromptDeleted   AuditAction = "prompt_deleted"
	AuditActionPromptPublished AuditAction = "prompt_published"

	// Dataset management
	AuditActionDatasetCreated AuditAction = "dataset_created"
	AuditActionDatasetUpdated AuditAction = "dataset_updated"
	AuditActionDatasetDeleted AuditAction = "dataset_deleted"

	// Evaluator management
	AuditActionEvaluatorCreated AuditAction = "evaluator_created"
	AuditActionEvaluatorUpdated AuditAction = "evaluator_updated"
	AuditActionEvaluatorDeleted AuditAction = "evaluator_deleted"
)

// AuditResourceType represents the type of resource being audited
type AuditResourceType string

const (
	AuditResourceUser         AuditResourceType = "user"
	AuditResourceOrganization AuditResourceType = "organization"
	AuditResourceProject      AuditResourceType = "project"
	AuditResourceAPIKey       AuditResourceType = "api_key"
	AuditResourceSSO          AuditResourceType = "sso"
	AuditResourcePrompt       AuditResourceType = "prompt"
	AuditResourceDataset      AuditResourceType = "dataset"
	AuditResourceEvaluator    AuditResourceType = "evaluator"
	AuditResourceTrace        AuditResourceType = "trace"
	AuditResourceSettings     AuditResourceType = "settings"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             uuid.UUID         `json:"id" db:"id"`
	OrganizationID uuid.UUID         `json:"organizationId" db:"organization_id"`
	ActorID        *uuid.UUID        `json:"actorId,omitempty" db:"actor_id"`        // User who performed the action
	ActorEmail     string            `json:"actorEmail" db:"actor_email"`            // Email for display (preserved even if user deleted)
	ActorType      string            `json:"actorType" db:"actor_type"`              // "user", "api_key", "system"
	Action         AuditAction       `json:"action" db:"action"`
	ResourceType   AuditResourceType `json:"resourceType" db:"resource_type"`
	ResourceID     *uuid.UUID        `json:"resourceId,omitempty" db:"resource_id"`
	ResourceName   string            `json:"resourceName,omitempty" db:"resource_name"` // Human-readable name
	Description    string            `json:"description" db:"description"`
	Metadata       map[string]any    `json:"metadata,omitempty" db:"metadata"` // Additional context
	Changes        *AuditChanges     `json:"changes,omitempty" db:"changes"`   // Before/after for updates

	// Request context
	IPAddress   string `json:"ipAddress" db:"ip_address"`
	UserAgent   string `json:"userAgent" db:"user_agent"`
	RequestID   string `json:"requestId,omitempty" db:"request_id"`
	SessionID   string `json:"sessionId,omitempty" db:"session_id"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// AuditChanges represents before/after state for update actions
type AuditChanges struct {
	Before map[string]any `json:"before,omitempty"`
	After  map[string]any `json:"after,omitempty"`
}

// AuditLogFilter represents filter options for querying audit logs
type AuditLogFilter struct {
	OrganizationID *uuid.UUID
	ActorID        *uuid.UUID
	Action         *AuditAction
	Actions        []AuditAction
	ResourceType   *AuditResourceType
	ResourceID     *uuid.UUID
	StartTime      *time.Time
	EndTime        *time.Time
	IPAddress      *string
	SearchQuery    *string // Search in description, resource name

	// Pagination
	Limit  int
	Offset int
}

// AuditLogList represents a paginated list of audit logs
type AuditLogList struct {
	Data       []AuditLog `json:"data"`
	TotalCount int        `json:"totalCount"`
	HasMore    bool       `json:"hasMore"`
}

// AuditLogInput represents input for creating an audit log entry
type AuditLogInput struct {
	OrganizationID uuid.UUID
	ActorID        *uuid.UUID
	ActorEmail     string
	ActorType      string
	Action         AuditAction
	ResourceType   AuditResourceType
	ResourceID     *uuid.UUID
	ResourceName   string
	Description    string
	Metadata       map[string]any
	Changes        *AuditChanges
	IPAddress      string
	UserAgent      string
	RequestID      string
	SessionID      string
}

// AuditRetentionPolicy represents the retention policy for audit logs
type AuditRetentionPolicy struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organizationId" db:"organization_id"`
	RetentionDays  int       `json:"retentionDays" db:"retention_days"` // 0 = forever
	Enabled        bool      `json:"enabled" db:"enabled"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// AuditExportRequest represents a request to export audit logs
type AuditExportRequest struct {
	OrganizationID uuid.UUID         `json:"organizationId"`
	Filter         AuditLogFilter    `json:"filter"`
	Format         string            `json:"format"` // "csv", "json"
	Compress       bool              `json:"compress"`
}

// AuditSummary provides aggregated audit statistics
type AuditSummary struct {
	OrganizationID   uuid.UUID              `json:"organizationId"`
	Period           string                 `json:"period"` // "day", "week", "month"
	TotalEvents      int                    `json:"totalEvents"`
	EventsByAction   map[AuditAction]int    `json:"eventsByAction"`
	EventsByResource map[AuditResourceType]int `json:"eventsByResource"`
	UniqueActors     int                    `json:"uniqueActors"`
	TopActors        []AuditActorSummary    `json:"topActors"`
}

// AuditActorSummary represents activity summary for an actor
type AuditActorSummary struct {
	ActorID    *uuid.UUID `json:"actorId,omitempty"`
	ActorEmail string     `json:"actorEmail"`
	ActorType  string     `json:"actorType"`
	EventCount int        `json:"eventCount"`
}
