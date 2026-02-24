package domain

import (
	"time"
)

// HandoffPriority represents the priority level of a handoff
type HandoffPriority string

const (
	HandoffPriorityLow      HandoffPriority = "low"
	HandoffPriorityNormal   HandoffPriority = "normal"
	HandoffPriorityHigh     HandoffPriority = "high"
	HandoffPriorityCritical HandoffPriority = "critical"
)

// HandoffStatus represents the status of a handoff
type HandoffStatus string

const (
	HandoffStatusInitiated HandoffStatus = "initiated"
	HandoffStatusAccepted  HandoffStatus = "accepted"
	HandoffStatusCompleted HandoffStatus = "completed"
	HandoffStatusFailed    HandoffStatus = "failed"
)

// Handoff represents an agent-to-agent handoff
type Handoff struct {
	ID                    string          `json:"id"`
	ProjectID             string          `json:"projectId"`
	FromAgent             string          `json:"fromAgent"`
	ToAgent               string          `json:"toAgent"`
	TaskDescription       string          `json:"taskDescription"`
	Context               string          `json:"context"`
	Priority              HandoffPriority `json:"priority"`
	Status                HandoffStatus   `json:"status"`
	ContextPreservationPct float64        `json:"contextPreservationPct"`
	ResolutionTimeMs      int64           `json:"resolutionTimeMs"`
	CreatedAt             time.Time       `json:"createdAt"`
	CompletedAt           *time.Time      `json:"completedAt,omitempty"`
}

// HandoffChain represents a chain of handoffs in a trace
type HandoffChain struct {
	ID              string    `json:"id"`
	Handoffs        []Handoff `json:"handoffs"`
	TotalAgents     int       `json:"totalAgents"`
	AvgPreservation float64   `json:"avgPreservation"`
	AvgResolutionMs int64     `json:"avgResolutionMs"`
	Failures        int       `json:"failures"`
}

// HandoffInput represents input for initiating a handoff
type HandoffInput struct {
	FromAgent       string          `json:"fromAgent" validate:"required"`
	ToAgent         string          `json:"toAgent" validate:"required"`
	TaskDescription string          `json:"taskDescription" validate:"required"`
	Context         string          `json:"context"`
	Priority        HandoffPriority `json:"priority"`
}

// HandoffStats provides statistics about handoffs for a project
type HandoffStats struct {
	TotalHandoffs  int            `json:"totalHandoffs"`
	SuccessRate    float64        `json:"successRate"`
	AvgPreservation float64      `json:"avgPreservation"`
	AvgResolutionMs int64        `json:"avgResolutionMs"`
	ByPriority     map[string]int `json:"byPriority"`
}
