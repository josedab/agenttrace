package domain

import (
	"time"

	"github.com/google/uuid"
)

// AgentReplaySessionStatus represents the status of an interactive replay session
type AgentReplaySessionStatus string

const (
	AgentReplayRecording AgentReplaySessionStatus = "recording"
	AgentReplayCompleted AgentReplaySessionStatus = "completed"
	AgentReplayPlaying   AgentReplaySessionStatus = "playing"
	AgentReplayPaused    AgentReplaySessionStatus = "paused"
	AgentReplayFailed    AgentReplaySessionStatus = "failed"
)

// AgentReplayFidelity controls how much data is recorded
type AgentReplayFidelity string

const (
	AgentReplayFidelityFull     AgentReplayFidelity = "full"
	AgentReplayFidelityStandard AgentReplayFidelity = "standard"
	AgentReplayFidelityMinimal  AgentReplayFidelity = "minimal"
)

// AgentReplaySession represents a recorded agent session with time-travel debugging
type AgentReplaySession struct {
	ID          uuid.UUID                `json:"id"`
	ProjectID   uuid.UUID                `json:"projectId"`
	TraceID     uuid.UUID                `json:"traceId"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Status      AgentReplaySessionStatus `json:"status"`

	// Recording metadata
	RecordingFidelity AgentReplayFidelity `json:"recordingFidelity"`
	TotalEvents       int                 `json:"totalEvents"`
	TotalDurationMs   int64               `json:"totalDurationMs"`
	FilesTracked      int                 `json:"filesTracked"`
	CheckpointCount   int                 `json:"checkpointCount"`

	// Branching support
	ParentSessionID *uuid.UUID `json:"parentSessionId,omitempty"`
	BranchPoint     int        `json:"branchPoint,omitempty"`

	// Sharing
	IsPublic bool   `json:"isPublic"`
	ShareURL string `json:"shareUrl,omitempty"`

	// Audit
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

// AgentReplayTimelineEvent represents a single recorded event with state delta
type AgentReplayTimelineEvent struct {
	ID             uuid.UUID              `json:"id"`
	SessionID      uuid.UUID              `json:"sessionId"`
	Index          int                    `json:"index"`
	Type           ReplayEventType        `json:"type"`
	Timestamp      time.Time              `json:"timestamp"`
	Data           map[string]interface{} `json:"data"`
	Input          interface{}            `json:"input,omitempty"`
	Output         interface{}            `json:"output,omitempty"`
	DurationMs     int64                  `json:"durationMs,omitempty"`
	ObservationID  *uuid.UUID             `json:"observationId,omitempty"`
	FileDelta      *ReplayFileDelta       `json:"fileDelta,omitempty"`
}

// ReplayFileDelta represents a file system change at a point in time
type ReplayFileDelta struct {
	Path      string `json:"path"`
	Operation string `json:"operation"`
	Before    string `json:"before,omitempty"`
	After     string `json:"after,omitempty"`
	DiffPatch string `json:"diffPatch,omitempty"`
}

// AgentReplayPlaybackState represents the current playback position
type AgentReplayPlaybackState struct {
	SessionID    uuid.UUID `json:"sessionId"`
	CurrentIndex int       `json:"currentIndex"`
	TotalEvents  int       `json:"totalEvents"`
	IsPlaying    bool      `json:"isPlaying"`
	Speed        float64   `json:"speed"`
	ElapsedMs    int64     `json:"elapsedMs"`
	TotalMs      int64     `json:"totalMs"`
}

// AgentReplayBranch represents a branch point in the replay
type AgentReplayBranch struct {
	SessionID  uuid.UUID `json:"sessionId"`
	EventIndex int       `json:"eventIndex"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
}

// AgentReplayMilestone represents an important moment in the replay
type AgentReplayMilestone struct {
	EventIndex  int    `json:"eventIndex"`
	Label       string `json:"label"`
	Type        string `json:"type"` // checkpoint, error, completion, decision_point
	Description string `json:"description,omitempty"`
}

// AgentReplayFullTimeline represents the complete interactive timeline
type AgentReplayFullTimeline struct {
	Session    AgentReplaySession          `json:"session"`
	Events     []AgentReplayTimelineEvent  `json:"events"`
	Branches   []AgentReplayBranch         `json:"branches"`
	Milestones []AgentReplayMilestone      `json:"milestones"`
}

// AgentReplayBranchRequest represents a request to branch at a specific point
type AgentReplayBranchRequest struct {
	SessionID  uuid.UUID `json:"sessionId" validate:"required"`
	EventIndex int       `json:"eventIndex" validate:"required,min=0"`
	Name       string    `json:"name" validate:"required,min=1,max=100"`
}

// AgentReplaySessionInput represents input for creating a replay session
type AgentReplaySessionInput struct {
	TraceID           uuid.UUID           `json:"traceId" validate:"required"`
	Name              string              `json:"name" validate:"required,min=1,max=200"`
	Description       string              `json:"description,omitempty"`
	RecordingFidelity AgentReplayFidelity `json:"recordingFidelity,omitempty"`
}

// AgentReplaySessionFilter represents filter options
type AgentReplaySessionFilter struct {
	ProjectID uuid.UUID
	TraceID   *uuid.UUID
	Status    *AgentReplaySessionStatus
	IsPublic  *bool
}

// AgentReplaySessionList represents a paginated list
type AgentReplaySessionList struct {
	Sessions   []AgentReplaySession `json:"sessions"`
	TotalCount int64                `json:"totalCount"`
	HasMore    bool                 `json:"hasMore"`
}
