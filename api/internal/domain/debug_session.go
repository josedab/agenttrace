package domain

import (
	"time"

	"github.com/google/uuid"
)

// DebugSessionStatus represents the status of a debug session
type DebugSessionStatus string

const (
	DebugSessionActive   DebugSessionStatus = "ACTIVE"
	DebugSessionPaused   DebugSessionStatus = "PAUSED"
	DebugSessionComplete DebugSessionStatus = "COMPLETE"
)

// DebugSession represents an interactive debugging session on a trace
type DebugSession struct {
	ID          uuid.UUID          `json:"id"`
	ProjectID   uuid.UUID          `json:"projectId"`
	TraceID     string             `json:"traceId"`
	UserID      uuid.UUID          `json:"userId"`
	Status      DebugSessionStatus `json:"status"`
	CurrentStep int                `json:"currentStep"`
	TotalSteps  int                `json:"totalSteps"`
	Breakpoints []Breakpoint       `json:"breakpoints,omitempty"`
	Annotations []DebugAnnotation  `json:"annotations,omitempty"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// Breakpoint represents a debug breakpoint on a replay event
type Breakpoint struct {
	EventID   string          `json:"eventId"`
	EventType ReplayEventType `json:"eventType"`
	Condition string          `json:"condition,omitempty"`
	Enabled   bool            `json:"enabled"`
}

// DebugAnnotation represents a developer note on a replay event
type DebugAnnotation struct {
	ID        uuid.UUID `json:"id"`
	EventID   string    `json:"eventId"`
	UserID    uuid.UUID `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// DebugStepState represents the reconstructed state at a specific step
type DebugStepState struct {
	StepIndex    int               `json:"stepIndex"`
	Event        ReplayEvent       `json:"event"`
	FileTree     []FileTreeEntry   `json:"fileTree,omitempty"`
	ModifiedFiles []FileDiffEntry  `json:"modifiedFiles,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	CostSoFar    float64           `json:"costSoFar"`
	TokensSoFar  int               `json:"tokensSoFar"`
	ElapsedMs    int64             `json:"elapsedMs"`
}

// FileTreeEntry represents a file in the workspace at a point in time
type FileTreeEntry struct {
	Path     string `json:"path"`
	Size     int64  `json:"size,omitempty"`
	Modified bool   `json:"modified"`
	Created  bool   `json:"created"`
	Deleted  bool   `json:"deleted"`
}

// FileDiffEntry represents a file diff between steps
type FileDiffEntry struct {
	Path   string `json:"path"`
	Diff   string `json:"diff"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// DebugCompareResult represents a comparison between two trace replays
type DebugCompareResult struct {
	TraceA       ReplayTimeline    `json:"traceA"`
	TraceB       ReplayTimeline    `json:"traceB"`
	Differences  []ComparisonDiff  `json:"differences"`
	CostDelta    float64           `json:"costDelta"`
	LatencyDelta int64             `json:"latencyDeltaMs"`
	TokenDelta   int               `json:"tokenDelta"`
}

// ComparisonDiff represents a difference between two trace replays
type ComparisonDiff struct {
	Type     string `json:"type"` // added, removed, changed
	StepA    *int   `json:"stepA,omitempty"`
	StepB    *int   `json:"stepB,omitempty"`
	Summary  string `json:"summary"`
}

// CreateDebugSessionInput represents input for creating a debug session
type CreateDebugSessionInput struct {
	TraceID     string       `json:"traceId" validate:"required"`
	Breakpoints []Breakpoint `json:"breakpoints,omitempty"`
}
