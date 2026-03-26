package domain

import (
	"time"

	"github.com/google/uuid"
)

// IntentDeclaration records an agent's declared intent and its verification
// result, tracking alignment between what the agent said it would do and
// what it actually did.
type IntentDeclaration struct {
	ID              uuid.UUID           `json:"id"`
	ProjectID       uuid.UUID           `json:"projectId"`
	TraceID         string              `json:"traceId"`
	AgentName       string              `json:"agentName"`
	DeclaredIntent  string              `json:"declaredIntent"`
	DeclaredActions []string            `json:"declaredActions"`
	Status          string              `json:"status"` // pending, verified, misaligned
	ActualActions   []string            `json:"actualActions,omitempty"`
	AlignmentScore  float64             `json:"alignmentScore"`
	Misalignments   []IntentMisalignment `json:"misalignments,omitempty"`
	DeclaredAt      time.Time           `json:"declaredAt"`
	VerifiedAt      *time.Time          `json:"verifiedAt,omitempty"`
}

// IntentMisalignment describes a specific discrepancy between a declared
// action and the action actually performed by the agent.
type IntentMisalignment struct {
	DeclaredAction string `json:"declaredAction"`
	ActualAction   string `json:"actualAction"`
	Severity       string `json:"severity"` // minor, major, critical
	Description    string `json:"description"`
}

// IntentVerificationStats holds aggregate intent verification statistics for a project.
type IntentVerificationStats struct {
	ProjectID          uuid.UUID      `json:"projectId"`
	TotalVerifications int            `json:"totalVerifications"`
	AlignmentRate      float64        `json:"alignmentRate"`
	MisalignmentsByAgent map[string]int `json:"misalignmentsByAgent"`
}

// IntentInput is the input for declaring an agent's intent before execution.
type IntentInput struct {
	TraceID         string   `json:"traceId" validate:"required"`
	AgentName       string   `json:"agentName" validate:"required"`
	DeclaredIntent  string   `json:"declaredIntent" validate:"required"`
	DeclaredActions []string `json:"declaredActions" validate:"required"`
}
