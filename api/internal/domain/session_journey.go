package domain

import (
	"time"

	"github.com/google/uuid"
)

// SessionJourney represents a session-based trace journey
type SessionJourney struct {
	ID            uuid.UUID       `json:"id"`
	ProjectID     uuid.UUID       `json:"projectId"`
	SessionID     string          `json:"sessionId"`
	Phases        []WorkflowPhase `json:"phases"`
	TotalDuration int64           `json:"totalDuration"` // milliseconds
	TotalCost     float64         `json:"totalCost"`
	TotalTokens   int64           `json:"totalTokens"`
	TraceCount    int             `json:"traceCount"`
	Status        string          `json:"status"` // active, completed, failed
	DetectedAt    time.Time       `json:"detectedAt"`
}

// WorkflowPhase represents a phase within a session journey
type WorkflowPhase struct {
	Name       string       `json:"name"` // planning, implementation, testing, debugging, review, deployment
	StartTime  time.Time    `json:"startTime"`
	EndTime    *time.Time   `json:"endTime,omitempty"`
	DurationMs int64        `json:"durationMs"`
	TraceIDs   []string     `json:"traceIds"`
	Metrics    PhaseMetrics `json:"metrics"`
	Confidence float64      `json:"confidence"` // 0-1
}

// PhaseMetrics represents metrics for a workflow phase
type PhaseMetrics struct {
	Cost          float64 `json:"cost"`
	Tokens        int64   `json:"tokens"`
	ErrorCount    int     `json:"errorCount"`
	ToolCallCount int     `json:"toolCallCount"`
	FilesModified int     `json:"filesModified"`
}

// JourneyFilter represents filter options for querying session journeys
type JourneyFilter struct {
	ProjectID uuid.UUID  `json:"projectId"`
	SessionID *string    `json:"sessionId,omitempty"`
	Status    *string    `json:"status,omitempty"`
	FromTime  *time.Time `json:"fromTime,omitempty"`
	ToTime    *time.Time `json:"toTime,omitempty"`
}
