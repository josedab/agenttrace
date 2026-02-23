package domain

import (
	"time"

	"github.com/google/uuid"
)

// StreamEventType represents types of streaming events
type StreamEventType string

const (
	StreamEventTraceActivity    StreamEventType = "trace.activity"
	StreamEventObservationStart StreamEventType = "observation.start"
	StreamEventObservationEnd   StreamEventType = "observation.end"
	StreamEventFileChange       StreamEventType = "file.change"
	StreamEventTerminalOutput   StreamEventType = "terminal.output"
	StreamEventCostUpdate       StreamEventType = "cost.update"
	StreamEventErrorOccurred    StreamEventType = "error.occurred"
	StreamEventMetricsSnapshot  StreamEventType = "metrics.snapshot"
	StreamEventIntervention     StreamEventType = "intervention.requested"
)

// LiveMetrics represents real-time aggregated metrics for a trace
type LiveMetrics struct {
	TraceID          uuid.UUID `json:"traceId"`
	ActiveSpans      int       `json:"activeSpans"`
	CompletedSpans   int       `json:"completedSpans"`
	TotalTokens      int       `json:"totalTokens"`
	TotalCost        float64   `json:"totalCost"`
	ErrorCount       int       `json:"errorCount"`
	ElapsedMs        int64     `json:"elapsedMs"`
	TokensPerSecond  float64   `json:"tokensPerSecond"`
	CostPerMinute    float64   `json:"costPerMinute"`
	FilesModified    int       `json:"filesModified"`
	TerminalCommands int       `json:"terminalCommands"`
	LastUpdated      time.Time `json:"lastUpdated"`
}

// StreamActivity represents an individual activity entry in the live feed
type StreamActivity struct {
	ID          string          `json:"id"`
	TraceID     uuid.UUID       `json:"traceId"`
	Type        StreamEventType `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
	DurationMs  *int64          `json:"durationMs,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	Status      string          `json:"status"` // running, completed, error
}

// InterventionAction represents a control action on a running agent
type InterventionAction string

const (
	InterventionPause   InterventionAction = "pause"
	InterventionResume  InterventionAction = "resume"
	InterventionCancel  InterventionAction = "cancel"
	InterventionMessage InterventionAction = "message"
)

// InterventionRequest represents a request to intervene in an agent's execution
type InterventionRequest struct {
	ID        uuid.UUID          `json:"id"`
	TraceID   uuid.UUID          `json:"traceId"`
	ProjectID uuid.UUID          `json:"projectId"`
	Action    InterventionAction `json:"action"`
	Message   string             `json:"message,omitempty"`
	UserID    uuid.UUID          `json:"userId"`
	CreatedAt time.Time          `json:"createdAt"`
	Status    string             `json:"status"` // pending, delivered, acknowledged
}

// StreamSubscription represents a client's subscription preferences
type StreamSubscription struct {
	TraceID    *uuid.UUID        `json:"traceId,omitempty"`
	EventTypes []StreamEventType `json:"eventTypes,omitempty"`
	FollowMode bool              `json:"followMode"`
}
