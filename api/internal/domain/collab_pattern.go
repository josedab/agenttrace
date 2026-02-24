package domain

import (
	"time"

	"github.com/google/uuid"
)

// CollabPattern represents an agent collaboration pattern
type CollabPattern struct {
	ID             uuid.UUID        `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Type           string           `json:"type"` // coordinator, pipeline, voting, debate, swarm, hierarchical
	AgentRoles     []PatternRole    `json:"agentRoles"`
	MessageFlow    []PatternMessage `json:"messageFlow"`
	Complexity     string           `json:"complexity"` // simple, moderate, complex
	UseCases       []string         `json:"useCases"`
	DeployCount    int              `json:"deployCount"`
	AvgPerformance float64          `json:"avgPerformance"`
	CreatedAt      time.Time        `json:"createdAt"`
}

// PatternRole represents a role within a collaboration pattern
type PatternRole struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Responsibilities []string `json:"responsibilities"`
	Model            string   `json:"model"`
}

// PatternMessage represents a message flow between roles in a pattern
type PatternMessage struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Type        string `json:"type"` // request, response, broadcast, vote
	Description string `json:"description"`
}

// PatternDeployment represents a deployment of a collaboration pattern
type PatternDeployment struct {
	ID         uuid.UUID      `json:"id"`
	PatternID  uuid.UUID      `json:"patternId"`
	ProjectID  uuid.UUID      `json:"projectId"`
	Config     map[string]any `json:"config"`
	Status     string         `json:"status"` // deployed, active, stopped
	DeployedAt time.Time      `json:"deployedAt"`
}

// DeployPatternInput represents input for deploying a pattern
type DeployPatternInput struct {
	PatternID uuid.UUID      `json:"patternId"`
	Config    map[string]any `json:"config"`
}
