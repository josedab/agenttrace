package domain

import (
	"time"

	"github.com/google/uuid"
)

// SandboxStatus represents the lifecycle state of a sandbox review.
type SandboxStatus string

const (
	SandboxStatusPending   SandboxStatus = "pending"
	SandboxStatusReviewing SandboxStatus = "reviewing"
	SandboxStatusApproved  SandboxStatus = "approved"
	SandboxStatusRejected  SandboxStatus = "rejected"
	SandboxStatusExpired   SandboxStatus = "expired"
)

// IsValid reports whether s is a recognized sandbox status value.
func (s SandboxStatus) IsValid() bool {
	switch s {
	case SandboxStatusPending, SandboxStatusReviewing, SandboxStatusApproved, SandboxStatusRejected, SandboxStatusExpired:
		return true
	}
	return false
}

// SandboxReview represents an agent action review, including proposed actions,
// risk assessment, and approval state.
type SandboxReview struct {
	ID              uuid.UUID       `json:"id"`
	ProjectID       uuid.UUID       `json:"projectId"`
	TraceID         uuid.UUID       `json:"traceId"`
	Status          SandboxStatus   `json:"status"`
	ProposedActions []SandboxAction `json:"proposedActions"`
	RiskLevel       string          `json:"riskLevel"` // low, medium, high, critical
	RiskScore       float64         `json:"riskScore"` // 0-100
	ReviewerID      *uuid.UUID      `json:"reviewerId,omitempty"`
	ReviewNote      string          `json:"reviewNote,omitempty"`
	Policy          SandboxPolicy   `json:"policy"`
	CreatedAt       time.Time       `json:"createdAt"`
	ReviewedAt      *time.Time      `json:"reviewedAt,omitempty"`
	ExpiresAt       time.Time       `json:"expiresAt"`
}

// SandboxAction describes a single proposed agent action subject to sandbox review.
type SandboxAction struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // file_write, file_delete, command_exec, network_request, env_access
	Target      string `json:"target"` // file path, command, URL
	Description string `json:"description"`
	RiskLevel   string `json:"riskLevel"`
	Diff        string `json:"diff,omitempty"`
	Approved    *bool  `json:"approved,omitempty"`
}

// SandboxPolicy defines the security policy governing which agent actions
// are allowed, blocked, or require human review.
type SandboxPolicy struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"projectId"`
	Name            string    `json:"name"`
	AllowedPaths    []string  `json:"allowedPaths"`
	BlockedPaths    []string  `json:"blockedPaths"`
	AllowedCommands []string  `json:"allowedCommands"`
	BlockedCommands []string  `json:"blockedCommands"`
	AllowNetwork    bool      `json:"allowNetwork"`
	AllowEnvAccess  bool      `json:"allowEnvAccess"`
	RequireReview   string    `json:"requireReview"` // always, high_risk, never
	MaxFileSize     int64     `json:"maxFileSizeBytes"`
	CreatedAt       time.Time `json:"createdAt"`
}

// SandboxReviewInput is the input for creating a new sandbox review.
type SandboxReviewInput struct {
	TraceID         uuid.UUID       `json:"traceId" validate:"required"`
	ProposedActions []SandboxAction `json:"proposedActions" validate:"required"`
}

// SandboxDecision represents a reviewer's decision on a sandbox review,
// including full approval, rejection, or partial approval of specific actions.
type SandboxDecision struct {
	Action    string   `json:"action"` // approve, reject, approve_partial
	Note      string   `json:"note,omitempty"`
	ActionIDs []string `json:"actionIds,omitempty"` // for partial approval
}

// SandboxPolicyInput is the input for creating or updating a sandbox policy.
type SandboxPolicyInput struct {
	Name            string   `json:"name" validate:"required"`
	AllowedPaths    []string `json:"allowedPaths"`
	BlockedPaths    []string `json:"blockedPaths"`
	AllowedCommands []string `json:"allowedCommands"`
	BlockedCommands []string `json:"blockedCommands"`
	AllowNetwork    *bool    `json:"allowNetwork"`
	AllowEnvAccess  *bool    `json:"allowEnvAccess"`
	RequireReview   string   `json:"requireReview"`
	MaxFileSize     *int64   `json:"maxFileSizeBytes"`
}

// SandboxStats contains aggregate statistics for sandbox reviews within a project.
type SandboxStats struct {
	TotalReviews    int                `json:"totalReviews"`
	PendingReviews  int                `json:"pendingReviews"`
	ApprovalRate    float64            `json:"approvalRate"`
	AvgReviewTimeMs int64              `json:"avgReviewTimeMs"`
	ByRiskLevel     map[string]int     `json:"byRiskLevel"`
	ByStatus        map[string]int     `json:"byStatus"`
	TopBlockedPaths []PathBlockCount   `json:"topBlockedPaths"`
}

// PathBlockCount tracks how many times a specific path has been blocked by sandbox policy.
type PathBlockCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// CloudSandboxSession represents a browser-based demo environment
type CloudSandboxSession struct {
	ID            uuid.UUID            `json:"id"`
	Status        CloudSandboxStatus   `json:"status"`
	URL           string               `json:"url"`
	DashboardURL  string               `json:"dashboardUrl"`
	APIURL        string               `json:"apiUrl"`
	APIKey        string               `json:"apiKey"`
	ExpiresAt     time.Time            `json:"expiresAt"`
	SampleData    CloudSandboxData     `json:"sampleData"`
	CreatedAt     time.Time            `json:"createdAt"`
}

// CloudSandboxStatus represents the status of a cloud sandbox session
type CloudSandboxStatus string

const (
	CloudSandboxProvisioning CloudSandboxStatus = "provisioning"
	CloudSandboxReady        CloudSandboxStatus = "ready"
	CloudSandboxExpired      CloudSandboxStatus = "expired"
	CloudSandboxError        CloudSandboxStatus = "error"
)

// CloudSandboxData describes the pre-loaded sample data
type CloudSandboxData struct {
	TraceCount       int      `json:"traceCount"`
	AgentCount       int      `json:"agentCount"`
	PromptCount      int      `json:"promptCount"`
	DatasetCount     int      `json:"datasetCount"`
	Features         []string `json:"features"`
	SampleAgents     []string `json:"sampleAgents"`
}

// CloudSandboxInput represents input for creating a cloud sandbox
type CloudSandboxInput struct {
	Email       string `json:"email,omitempty"`
	UseCase     string `json:"useCase,omitempty"`
	PreloadData bool   `json:"preloadData"`
}
