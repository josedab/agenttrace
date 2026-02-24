package domain

import (
	"time"

	"github.com/google/uuid"
)

// AgentVersion represents a versioned snapshot of an agent configuration
type AgentVersion struct {
	ID                 uuid.UUID       `json:"id"`
	ProjectID          uuid.UUID       `json:"projectId"`
	AgentName          string          `json:"agentName"`
	Version            int             `json:"version"`
	Tag                string          `json:"tag,omitempty"`
	Config             AgentConfig     `json:"config"`
	PerformanceMetrics *VersionMetrics `json:"performanceMetrics,omitempty"`
	CreatedBy          string          `json:"createdBy,omitempty"`
	ChangeNote         string          `json:"changeNote,omitempty"`
	IsActive           bool            `json:"isActive"`
	CreatedAt          time.Time       `json:"createdAt"`
}

// AgentConfig represents the configuration of an agent
type AgentConfig struct {
	Model        string         `json:"model"`
	SystemPrompt string         `json:"systemPrompt"`
	Temperature  float64        `json:"temperature"`
	MaxTokens    int            `json:"maxTokens"`
	Tools        []string       `json:"tools,omitempty"`
	GuardrailIDs []uuid.UUID    `json:"guardrailIds,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
}

// VersionMetrics represents performance metrics for a version
type VersionMetrics struct {
	TraceCount   int     `json:"traceCount"`
	SuccessRate  float64 `json:"successRate"`
	AvgCost      float64 `json:"avgCost"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	AvgQuality   float64 `json:"avgQuality"`
}

// VersionDiff represents the difference between two versions
type VersionDiff struct {
	VersionA AgentVersion   `json:"versionA"`
	VersionB AgentVersion   `json:"versionB"`
	Changes  []ConfigChange `json:"changes"`
}

// ConfigChange represents a single change between two configs
type ConfigChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"oldValue"`
	NewValue any    `json:"newValue"`
}

// CreateVersionInput represents the input for creating a new version
type CreateVersionInput struct {
	AgentName  string      `json:"agentName" validate:"required"`
	Config     AgentConfig `json:"config" validate:"required"`
	Tag        string      `json:"tag,omitempty"`
	ChangeNote string      `json:"changeNote,omitempty"`
}

// RollbackInput represents the input for rolling back to a version
type RollbackInput struct {
	VersionID uuid.UUID `json:"versionId" validate:"required"`
}
