package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostPrediction represents a predicted cost for a task
type CostPrediction struct {
	ID                 uuid.UUID `json:"id"`
	ProjectID          uuid.UUID `json:"projectId"`
	TaskDescription    string    `json:"taskDescription"`
	PredictedCost      float64   `json:"predictedCost"`
	PredictedLatencyMs int64     `json:"predictedLatencyMs"`
	PredictedQuality   float64   `json:"predictedQuality"` // 0-100
	PredictedTokens    int       `json:"predictedTokens"`
	ConfidenceLevel    float64   `json:"confidenceLevel"` // 0-1
	RecommendedModel   string    `json:"recommendedModel"`
	BudgetStatus       string    `json:"budgetStatus"` // within, warning, exceeded
	SimilarTraces      int       `json:"similarTraces"`
	CreatedAt          time.Time `json:"createdAt"`
}

// PredictionInput represents the input for a cost prediction
type PredictionInput struct {
	TaskDescription string   `json:"taskDescription" validate:"required"`
	Model           string   `json:"model,omitempty"`
	MaxBudget       *float64 `json:"maxBudget,omitempty"`
}

// BudgetApproval represents a budget approval request
type BudgetApproval struct {
	ID           uuid.UUID  `json:"id"`
	ProjectID    uuid.UUID  `json:"projectId"`
	PredictionID uuid.UUID  `json:"predictionId"`
	Status       string     `json:"status"` // pending, approved, rejected
	ApproverID   *uuid.UUID `json:"approverId,omitempty"`
	Note         string     `json:"note,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	DecidedAt    *time.Time `json:"decidedAt,omitempty"`
}

// ApprovalDecisionInput represents the input for an approval decision
type ApprovalDecisionInput struct {
	Status string `json:"status"` // approved, rejected
	Note   string `json:"note,omitempty"`
}
