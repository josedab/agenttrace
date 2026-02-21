package domain

import (
	"github.com/google/uuid"
)

// AgentNodeType represents the type of agent node
type AgentNodeType string

const (
	AgentNodeOrchestrator AgentNodeType = "orchestrator"
	AgentNodeWorker       AgentNodeType = "worker"
	AgentNodeTool         AgentNodeType = "tool"
)

// AgentGraph represents a graph visualization of a multi-agent trace
type AgentGraph struct {
	TraceID         uuid.UUID   `json:"traceId"`
	ProjectID       uuid.UUID   `json:"projectId"`
	Agents          []AgentNode `json:"agents"`
	Edges           []AgentEdge `json:"edges"`
	TotalCost       float64     `json:"totalCost"`
	TotalDurationMs int64       `json:"totalDurationMs"`
}

// AgentNode represents a single agent in the graph
type AgentNode struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Type          AgentNodeType `json:"type"`
	ObservationID string        `json:"observationId"`
	Model         string        `json:"model,omitempty"`
	TokensUsed    int           `json:"tokensUsed"`
	Cost          float64       `json:"cost"`
	DurationMs    int64         `json:"durationMs"`
	Status        string        `json:"status"`
}

// AgentEdge represents a connection between two agents
type AgentEdge struct {
	SourceID     string `json:"sourceId"`
	TargetID     string `json:"targetId"`
	Label        string `json:"label,omitempty"`
	MessageCount int    `json:"messageCount"`
	TokenCount   int    `json:"tokenCount"`
}

// AgentGraphFilter represents filter options for building an agent graph
type AgentGraphFilter struct {
	TraceID         uuid.UUID `json:"traceId" validate:"required"`
	IncludeMessages bool      `json:"includeMessages"`
}

// AgentGraphComparison represents the result of comparing two agent graphs
type AgentGraphComparison struct {
	GraphA       AgentGraph `json:"graphA"`
	GraphB       AgentGraph `json:"graphB"`
	AddedNodes   []string   `json:"addedNodes"`
	RemovedNodes []string   `json:"removedNodes"`
	CostDelta    float64    `json:"costDelta"`
	LatencyDelta int64      `json:"latencyDeltaMs"`
}
