package domain

import (
	"time"

	"github.com/google/uuid"
)

// Additional agent node types for orchestration debugging
const (
	AgentNodeCoordinator AgentNodeType = "coordinator"
	AgentNodeHuman       AgentNodeType = "human"
)

// MessageDirection represents the direction of a message between agents
type MessageDirection string

const (
	MessageRequest   MessageDirection = "request"
	MessageResponse  MessageDirection = "response"
	MessageBroadcast MessageDirection = "broadcast"
)

// OrchestrationSession represents a multi-agent debugging session
type OrchestrationSession struct {
	ID          uuid.UUID           `json:"id"`
	ProjectID   uuid.UUID           `json:"projectId"`
	TraceID     uuid.UUID           `json:"traceId"`
	Agents      []OrchestratorAgent `json:"agents"`
	Messages    []AgentMessage      `json:"messages"`
	Breakpoints []AgentBreakpoint   `json:"breakpoints"`
	Status      string              `json:"status"` // running, paused, completed
	CurrentStep int                 `json:"currentStep"`
	TotalSteps  int                 `json:"totalSteps"`
	CreatedAt   time.Time           `json:"createdAt"`
}

// OrchestratorAgent represents an agent participating in the orchestration
type OrchestratorAgent struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       AgentNodeType  `json:"type"`
	Model      string         `json:"model,omitempty"`
	Status     string         `json:"status"` // idle, active, waiting, error
	State      map[string]any `json:"state,omitempty"`
	TokensUsed int            `json:"tokensUsed"`
	Cost       float64        `json:"cost"`
}

// AgentMessage represents a message exchanged between agents
type AgentMessage struct {
	ID         string           `json:"id"`
	FromAgent  string           `json:"fromAgent"`
	ToAgent    string           `json:"toAgent"`
	Direction  MessageDirection `json:"direction"`
	Content    string           `json:"content"`
	Timestamp  time.Time        `json:"timestamp"`
	LatencyMs  int64            `json:"latencyMs,omitempty"`
	TokenCount int              `json:"tokenCount,omitempty"`
	StepIndex  int              `json:"stepIndex"`
}

// AgentBreakpoint represents a breakpoint set in an orchestration session
type AgentBreakpoint struct {
	ID        string `json:"id"`
	AgentID   string `json:"agentId"`
	Condition string `json:"condition"` // on_message, on_error, on_state_change, always
	Enabled   bool   `json:"enabled"`
}

// DebugCommand represents a debugging command to execute
type DebugCommand struct {
	Action    string `json:"action"` // step, continue, step_over, inspect, set_breakpoint
	AgentID   string `json:"agentId,omitempty"`
	StepCount int    `json:"stepCount,omitempty"`
	Condition string `json:"condition,omitempty"`
}

// OrchestrationSessionInput represents the input for creating a session
type OrchestrationSessionInput struct {
	TraceID uuid.UUID `json:"traceId" validate:"required"`
}
