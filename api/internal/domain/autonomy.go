package domain

import (
	"time"

	"github.com/google/uuid"
)

// AutonomyLevel represents the degree of autonomy granted to an agent.
type AutonomyLevel string

const (
	AutonomyFullAuto    AutonomyLevel = "full_auto"
	AutonomyHumanGuided AutonomyLevel = "human_guided"
	AutonomySupervised  AutonomyLevel = "supervised"
	AutonomyManual      AutonomyLevel = "manual"
)

// IsValid reports whether a is a recognized autonomy level value.
func (a AutonomyLevel) IsValid() bool {
	switch a {
	case AutonomyFullAuto, AutonomyHumanGuided, AutonomySupervised, AutonomyManual:
		return true
	}
	return false
}

// AutonomyConfig defines the autonomy level, permissions, and trust score
// for a specific agent within a project.
type AutonomyConfig struct {
	ID          uuid.UUID           `json:"id"`
	ProjectID   uuid.UUID           `json:"projectId"`
	AgentName   string              `json:"agentName"`
	Level       AutonomyLevel       `json:"level"`
	Permissions AutonomyPermissions `json:"permissions"`
	TrustScore  float64             `json:"trustScore"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

// AutonomyPermissions specifies which operations an agent is allowed to perform.
type AutonomyPermissions struct {
	CanWriteFiles      bool    `json:"canWriteFiles"`
	CanDeleteFiles     bool    `json:"canDeleteFiles"`
	CanExecuteCommands bool    `json:"canExecuteCommands"`
	CanAccessNetwork   bool    `json:"canAccessNetwork"`
	CanModifyConfig    bool    `json:"canModifyConfig"`
	RequiresApproval   bool    `json:"requiresApproval"`
	MaxCostPerRun      float64 `json:"maxCostPerRun"`
}

// TrustEvolution tracks the historical evolution of an agent's trust score.
type TrustEvolution struct {
	AgentName string           `json:"agentName"`
	History   []TrustDataPoint `json:"history"`
	Current   float64          `json:"currentScore"`
	Trend     string           `json:"trend"`
}

// TrustDataPoint is a single data point in an agent's trust score history.
type TrustDataPoint struct {
	Timestamp  time.Time     `json:"timestamp"`
	TrustScore float64       `json:"trustScore"`
	Level      AutonomyLevel `json:"level"`
	Reason     string        `json:"reason,omitempty"`
}

// AutonomyDashboard provides an overview of autonomy configurations and trust
// distribution across all agents in a project.
type AutonomyDashboard struct {
	ProjectID    uuid.UUID                `json:"projectId"`
	Agents       []AutonomyConfig         `json:"agents"`
	Distribution map[AutonomyLevel]int    `json:"distribution"`
	AvgTrust     float64                  `json:"avgTrustScore"`
}

// AutonomyConfigInput is the input for creating or updating an autonomy configuration.
type AutonomyConfigInput struct {
	AgentName   string               `json:"agentName" validate:"required"`
	Level       AutonomyLevel        `json:"level" validate:"required"`
	Permissions *AutonomyPermissions `json:"permissions,omitempty"`
}
