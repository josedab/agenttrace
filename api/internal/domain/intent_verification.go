package domain

import (
	"time"

	"github.com/google/uuid"
)

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

type IntentMisalignment struct {
	DeclaredAction string `json:"declaredAction"`
	ActualAction   string `json:"actualAction"`
	Severity       string `json:"severity"` // minor, major, critical
	Description    string `json:"description"`
}

type IntentVerificationStats struct {
	ProjectID          uuid.UUID      `json:"projectId"`
	TotalVerifications int            `json:"totalVerifications"`
	AlignmentRate      float64        `json:"alignmentRate"`
	MisalignmentsByAgent map[string]int `json:"misalignmentsByAgent"`
}

type IntentInput struct {
	TraceID         string   `json:"traceId" validate:"required"`
	AgentName       string   `json:"agentName" validate:"required"`
	DeclaredIntent  string   `json:"declaredIntent" validate:"required"`
	DeclaredActions []string `json:"declaredActions" validate:"required"`
}
