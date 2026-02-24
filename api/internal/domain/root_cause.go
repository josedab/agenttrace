package domain

import (
	"time"

	"github.com/google/uuid"
)

// FailureCategory represents the category of a failure
type FailureCategory string

const (
	FailureCategoryPrompt         FailureCategory = "prompt_issue"
	FailureCategoryModel          FailureCategory = "model_limitation"
	FailureCategoryContext         FailureCategory = "context_overflow"
	FailureCategoryTool           FailureCategory = "tool_failure"
	FailureCategoryData           FailureCategory = "data_quality"
	FailureCategoryTimeout        FailureCategory = "timeout"
	FailureCategoryRateLimit      FailureCategory = "rate_limit"
	FailureCategoryInfrastructure FailureCategory = "infrastructure"
)

// RCAReport represents a root cause analysis report
type RCAReport struct {
	ID                  uuid.UUID            `json:"id"`
	ProjectID           uuid.UUID            `json:"projectId"`
	TraceID             uuid.UUID            `json:"traceId"`
	PrimaryCategory     FailureCategory      `json:"primaryCategory"`
	Confidence          float64              `json:"confidence"`
	Summary             string               `json:"summary"`
	DetailedAnalysis    string               `json:"detailedAnalysis"`
	ContributingFactors []ContributingFactor `json:"contributingFactors"`
	Remediations        []Remediation        `json:"remediations"`
	SimilarIncidents    []uuid.UUID          `json:"similarIncidents,omitempty"`
	AnalyzedAt          time.Time            `json:"analyzedAt"`
}

// ContributingFactor represents a factor that contributed to a failure
type ContributingFactor struct {
	Category    FailureCategory `json:"category"`
	Description string          `json:"description"`
	Evidence    string          `json:"evidence"`
	Impact      float64         `json:"impact"` // 0-1
}

// Remediation represents a recommended action to fix a failure
type Remediation struct {
	Priority    int    `json:"priority"` // 1=highest
	Action      string `json:"action"`
	Description string `json:"description"`
	Automated   bool   `json:"automated"`
}

// RCAInput represents the input for triggering root cause analysis
type RCAInput struct {
	TraceID uuid.UUID `json:"traceId" validate:"required"`
}
