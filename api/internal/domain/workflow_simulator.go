package domain

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowNodeType represents the type of a node in an agent workflow
type WorkflowNodeType string

const (
	WorkflowNodeTypeLLMCall     WorkflowNodeType = "llm_call"
	WorkflowNodeTypeToolCall    WorkflowNodeType = "tool_call"
	WorkflowNodeTypeCondition   WorkflowNodeType = "condition"
	WorkflowNodeTypeLoop        WorkflowNodeType = "loop"
	WorkflowNodeTypeParallel    WorkflowNodeType = "parallel"
	WorkflowNodeTypeHumanReview WorkflowNodeType = "human_review"
)

// IsValid checks if the workflow node type is valid
func (t WorkflowNodeType) IsValid() bool {
	switch t {
	case WorkflowNodeTypeLLMCall, WorkflowNodeTypeToolCall, WorkflowNodeTypeCondition, WorkflowNodeTypeLoop, WorkflowNodeTypeParallel, WorkflowNodeTypeHumanReview:
		return true
	}
	return false
}

// WorkflowStatus represents the status of a workflow definition
type WorkflowStatus string

const (
	WorkflowStatusDraft      WorkflowStatus = "draft"
	WorkflowStatusActive     WorkflowStatus = "active"
	WorkflowStatusSimulating WorkflowStatus = "simulating"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusArchived   WorkflowStatus = "archived"
)

// IsValid checks if the workflow status is valid
func (s WorkflowStatus) IsValid() bool {
	switch s {
	case WorkflowStatusDraft, WorkflowStatusActive, WorkflowStatusSimulating, WorkflowStatusCompleted, WorkflowStatusArchived:
		return true
	}
	return false
}

// WorkflowDefinition represents a visual agent workflow definition
type WorkflowDefinition struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Status      WorkflowStatus `json:"status"`
	Nodes       []WorkflowNode `json:"nodes"`
	Edges       []WorkflowEdge `json:"edges"`
	Version     int            `json:"version"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	CreatedBy   uuid.UUID      `json:"createdBy"`
}

// WorkflowNode represents a single node in the workflow graph
type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     WorkflowNodeType       `json:"type"`
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Position WorkflowPosition       `json:"position"`
}

// WorkflowPosition represents the X/Y position of a node in the visual editor
type WorkflowPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorkflowEdge represents a directed edge between two workflow nodes
type WorkflowEdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Label     string `json:"label,omitempty"`
	Condition string `json:"condition,omitempty"`
}

// WorkflowSimulation represents a simulation run against a workflow definition
type WorkflowSimulation struct {
	ID                     uuid.UUID              `json:"id"`
	WorkflowID             uuid.UUID              `json:"workflowId"`
	ProjectID              uuid.UUID              `json:"projectId"`
	Name                   string                 `json:"name"`
	Status                 string                 `json:"status"`
	PredictedCostUSD       float64                `json:"predictedCostUsd"`
	PredictedLatencyMs     float64                `json:"predictedLatencyMs"`
	PredictedQualityScore  float64                `json:"predictedQualityScore"`
	TraceDataUsed          int                    `json:"traceDataUsed"`
	Results                []SimulationStepResult `json:"results"`
	StartedAt              *time.Time             `json:"startedAt,omitempty"`
	CompletedAt            *time.Time             `json:"completedAt,omitempty"`
	CreatedAt              time.Time              `json:"createdAt"`
}

// SimulationStepResult represents the predicted outcome for a single workflow node
type SimulationStepResult struct {
	NodeID           string  `json:"nodeId"`
	NodeName         string  `json:"nodeName"`
	PredictedCostUSD float64 `json:"predictedCostUsd"`
	PredictedLatencyMs float64 `json:"predictedLatencyMs"`
	TokensEstimated  int     `json:"tokensEstimated"`
	Confidence       float64 `json:"confidence"`
	BasedOnTraces    int     `json:"basedOnTraces"`
}

// WorkflowDefinitionInput represents input for creating or updating a workflow definition
type WorkflowDefinitionInput struct {
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description,omitempty"`
	Nodes       []WorkflowNode `json:"nodes" validate:"required"`
	Edges       []WorkflowEdge `json:"edges"`
}

// SimulationInput represents input for running a workflow simulation
type SimulationInput struct {
	WorkflowID        uuid.UUID              `json:"workflowId" validate:"required"`
	Name              string                 `json:"name" validate:"required"`
	ScenarioOverrides map[string]interface{} `json:"scenarioOverrides,omitempty"`
}

// WorkflowList represents a paginated list of workflow definitions
type WorkflowList struct {
	Workflows  []WorkflowDefinition `json:"workflows"`
	TotalCount int64                `json:"totalCount"`
	HasMore    bool                 `json:"hasMore"`
}

// SimulationList represents a paginated list of workflow simulations
type SimulationList struct {
	Simulations []WorkflowSimulation `json:"simulations"`
	TotalCount  int64                `json:"totalCount"`
	HasMore     bool                 `json:"hasMore"`
}
