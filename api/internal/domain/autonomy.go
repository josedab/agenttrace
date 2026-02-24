package domain

import (
	"time"

	"github.com/google/uuid"
)

type AutonomyLevel string

const (
	AutonomyFullAuto    AutonomyLevel = "full_auto"
	AutonomyHumanGuided AutonomyLevel = "human_guided"
	AutonomySupervised  AutonomyLevel = "supervised"
	AutonomyManual      AutonomyLevel = "manual"
)

func (a AutonomyLevel) IsValid() bool {
	switch a {
	case AutonomyFullAuto, AutonomyHumanGuided, AutonomySupervised, AutonomyManual:
		return true
	}
	return false
}

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

type AutonomyPermissions struct {
	CanWriteFiles      bool    `json:"canWriteFiles"`
	CanDeleteFiles     bool    `json:"canDeleteFiles"`
	CanExecuteCommands bool    `json:"canExecuteCommands"`
	CanAccessNetwork   bool    `json:"canAccessNetwork"`
	CanModifyConfig    bool    `json:"canModifyConfig"`
	RequiresApproval   bool    `json:"requiresApproval"`
	MaxCostPerRun      float64 `json:"maxCostPerRun"`
}

type TrustEvolution struct {
	AgentName string           `json:"agentName"`
	History   []TrustDataPoint `json:"history"`
	Current   float64          `json:"currentScore"`
	Trend     string           `json:"trend"`
}

type TrustDataPoint struct {
	Timestamp  time.Time     `json:"timestamp"`
	TrustScore float64       `json:"trustScore"`
	Level      AutonomyLevel `json:"level"`
	Reason     string        `json:"reason,omitempty"`
}

type AutonomyDashboard struct {
	ProjectID    uuid.UUID                `json:"projectId"`
	Agents       []AutonomyConfig         `json:"agents"`
	Distribution map[AutonomyLevel]int    `json:"distribution"`
	AvgTrust     float64                  `json:"avgTrustScore"`
}

type AutonomyConfigInput struct {
	AgentName   string               `json:"agentName" validate:"required"`
	Level       AutonomyLevel        `json:"level" validate:"required"`
	Permissions *AutonomyPermissions `json:"permissions,omitempty"`
}
