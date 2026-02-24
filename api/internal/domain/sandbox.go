package domain

import (
	"time"

	"github.com/google/uuid"
)

type SandboxStatus string

const (
	SandboxStatusPending   SandboxStatus = "pending"
	SandboxStatusReviewing SandboxStatus = "reviewing"
	SandboxStatusApproved  SandboxStatus = "approved"
	SandboxStatusRejected  SandboxStatus = "rejected"
	SandboxStatusExpired   SandboxStatus = "expired"
)

func (s SandboxStatus) IsValid() bool {
	switch s {
	case SandboxStatusPending, SandboxStatusReviewing, SandboxStatusApproved, SandboxStatusRejected, SandboxStatusExpired:
		return true
	}
	return false
}

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

type SandboxAction struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // file_write, file_delete, command_exec, network_request, env_access
	Target      string `json:"target"` // file path, command, URL
	Description string `json:"description"`
	RiskLevel   string `json:"riskLevel"`
	Diff        string `json:"diff,omitempty"`
	Approved    *bool  `json:"approved,omitempty"`
}

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

type SandboxReviewInput struct {
	TraceID         uuid.UUID       `json:"traceId" validate:"required"`
	ProposedActions []SandboxAction `json:"proposedActions" validate:"required"`
}

type SandboxDecision struct {
	Action    string   `json:"action"` // approve, reject, approve_partial
	Note      string   `json:"note,omitempty"`
	ActionIDs []string `json:"actionIds,omitempty"` // for partial approval
}

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

type SandboxStats struct {
	TotalReviews    int                `json:"totalReviews"`
	PendingReviews  int                `json:"pendingReviews"`
	ApprovalRate    float64            `json:"approvalRate"`
	AvgReviewTimeMs int64              `json:"avgReviewTimeMs"`
	ByRiskLevel     map[string]int     `json:"byRiskLevel"`
	ByStatus        map[string]int     `json:"byStatus"`
	TopBlockedPaths []PathBlockCount   `json:"topBlockedPaths"`
}

type PathBlockCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}
