package domain

import (
	"time"

	"github.com/google/uuid"
)

// ShareResourceType identifies the read-only resource exposed by a token.
type ShareResourceType string

// Shareable resource types.
const (
	ShareResourceTrace      ShareResourceType = "trace"
	ShareResourceReplayPlan ShareResourceType = "replay_plan"
)

// IsValid reports whether a share resource is supported.
func (t ShareResourceType) IsValid() bool {
	return t == ShareResourceTrace || t == ShareResourceReplayPlan
}

// ShareLink stores only a token hash; the raw token is returned once at creation.
type ShareLink struct {
	ID               uuid.UUID         `json:"id"`
	ProjectID        uuid.UUID         `json:"-"`
	ResourceType     ShareResourceType `json:"resourceType"`
	ResourceID       string            `json:"resourceId"`
	TokenHash        []byte            `json:"-"`
	RedactionVersion int               `json:"redactionVersion"`
	ExpiresAt        time.Time         `json:"expiresAt"`
	RevokedAt        *time.Time        `json:"revokedAt,omitempty"`
	CreatedBy        uuid.UUID         `json:"createdBy"`
	CreatedAt        time.Time         `json:"createdAt"`
}

// ShareLinkInput creates an expiring project-scoped share token.
type ShareLinkInput struct {
	ResourceType     ShareResourceType `json:"resourceType"`
	ResourceID       string            `json:"resourceId"`
	ExpiresInSeconds int64             `json:"expiresInSeconds,omitempty"`
}

// ShareLinkCreated is returned once and includes the unpersisted raw token.
type ShareLinkCreated struct {
	ShareLink
	Token string `json:"token"`
	URL   string `json:"url"`
}

// SharedReplayEvent is a source-free replay event view.
type SharedReplayEvent struct {
	Type       ReplayEventType `json:"type"`
	Timestamp  time.Time       `json:"timestamp"`
	DurationMs int64           `json:"durationMs,omitempty"`
	Title      string          `json:"title"`
	Status     string          `json:"status"`
	Model      string          `json:"model,omitempty"`
	Tokens     int             `json:"tokens,omitempty"`
}

// SharedTraceView exposes timeline metadata without prompts, outputs, commands, paths, or diffs.
type SharedTraceView struct {
	TraceID    string              `json:"traceId"`
	Name       string              `json:"name"`
	StartTime  time.Time           `json:"startTime"`
	EndTime    *time.Time          `json:"endTime,omitempty"`
	DurationMs int64               `json:"durationMs"`
	Summary    ReplaySummary       `json:"summary"`
	Events     []SharedReplayEvent `json:"events"`
}

// SharedReplayPlanView exposes capabilities and comparison without source content.
type SharedReplayPlanView struct {
	PlanID       uuid.UUID              `json:"planId"`
	TraceID      string                 `json:"traceId"`
	Status       ReplayPlanStatus       `json:"status"`
	Capabilities ReplayCapabilityReport `json:"capabilities"`
	Comparison   *ReplayPlanComparison  `json:"comparison,omitempty"`
}

// SharedResourceView is the unauthenticated read-only response.
type SharedResourceView struct {
	ResourceType ShareResourceType     `json:"resourceType"`
	ExpiresAt    time.Time             `json:"expiresAt"`
	Trace        *SharedTraceView      `json:"trace,omitempty"`
	ReplayPlan   *SharedReplayPlanView `json:"replayPlan,omitempty"`
}
