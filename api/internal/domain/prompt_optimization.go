package domain

import (
	"time"

	"github.com/google/uuid"
)

// OptimizationStatus represents the status of a prompt optimization run
type OptimizationStatus string

const (
	OptimizationStatusAnalyzing  OptimizationStatus = "analyzing"
	OptimizationStatusGenerating OptimizationStatus = "generating"
	OptimizationStatusTesting    OptimizationStatus = "testing"
	OptimizationStatusPromoting  OptimizationStatus = "promoting"
	OptimizationStatusCompleted  OptimizationStatus = "completed"
	OptimizationStatusFailed     OptimizationStatus = "failed"
)

// IsValid checks if the optimization status is valid
func (s OptimizationStatus) IsValid() bool {
	switch s {
	case OptimizationStatusAnalyzing, OptimizationStatusGenerating, OptimizationStatusTesting, OptimizationStatusPromoting, OptimizationStatusCompleted, OptimizationStatusFailed:
		return true
	}
	return false
}

// VariantStatus represents the status of a prompt variant
type VariantStatus string

const (
	VariantStatusCandidate VariantStatus = "candidate"
	VariantStatusTesting   VariantStatus = "testing"
	VariantStatusWinning   VariantStatus = "winning"
	VariantStatusRejected  VariantStatus = "rejected"
	VariantStatusPromoted  VariantStatus = "promoted"
)

// IsValid checks if the variant status is valid
func (s VariantStatus) IsValid() bool {
	switch s {
	case VariantStatusCandidate, VariantStatusTesting, VariantStatusWinning, VariantStatusRejected, VariantStatusPromoted:
		return true
	}
	return false
}

// ContinuousPromptOptimization represents a continuous prompt optimization run
type ContinuousPromptOptimization struct {
	ID              uuid.UUID          `json:"id"`
	ProjectID       uuid.UUID          `json:"projectId"`
	PromptID        uuid.UUID          `json:"promptId"`
	PromptVersion   int                `json:"promptVersion"`
	Status          OptimizationStatus `json:"status"`
	FailurePatterns []OptimizationFailurePattern `json:"failurePatterns,omitempty"`
	Variants        []OptimizationVariant `json:"variants,omitempty"`
	BestVariantID   *uuid.UUID         `json:"bestVariantId,omitempty"`
	ImprovementPct  float64            `json:"improvementPct"`
	StartedAt       *time.Time         `json:"startedAt,omitempty"`
	CompletedAt     *time.Time         `json:"completedAt,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
}

// OptimizationFailurePattern represents a detected pattern of prompt failures
type OptimizationFailurePattern struct {
	Pattern       string      `json:"pattern"`
	Frequency     int         `json:"frequency"`
	ExampleTraceIDs []uuid.UUID `json:"exampleTraceIds,omitempty"`
	Category      string      `json:"category"`
	AvgScore      float64     `json:"avgScore"`
}

// OptimizationVariant represents a generated prompt variant for A/B testing
type OptimizationVariant struct {
	ID               uuid.UUID     `json:"id"`
	OptimizationID   uuid.UUID     `json:"optimizationId"`
	Content          string        `json:"content"`
	Rationale        string        `json:"rationale"`
	Status           VariantStatus `json:"status"`
	SampleSize       int           `json:"sampleSize"`
	AvgScore         float64       `json:"avgScore"`
	BaselineAvgScore float64       `json:"baselineAvgScore"`
	PValue           float64       `json:"pValue"`
	CreatedAt        time.Time     `json:"createdAt"`
}

// OptimizationConfig represents configuration for the prompt optimization engine
type OptimizationConfig struct {
	ID                      uuid.UUID `json:"id"`
	ProjectID               uuid.UUID `json:"projectId"`
	Enabled                 bool      `json:"enabled"`
	MinSamplesForAnalysis   int       `json:"minSamplesForAnalysis"`
	MinSamplesForPromotion  int       `json:"minSamplesForPromotion"`
	PValueThreshold         float64   `json:"pValueThreshold"`
	RequireApproval         bool      `json:"requireApproval"`
	MaxVariantsPerRound     int       `json:"maxVariantsPerRound"`
	ScheduleCron            string    `json:"scheduleCron,omitempty"`
}
