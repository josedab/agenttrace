package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostRecommendationStatus represents the status of a cost recommendation
type CostRecommendationStatus string

const (
	CostRecommendationPending   CostRecommendationStatus = "PENDING"
	CostRecommendationApplied   CostRecommendationStatus = "APPLIED"
	CostRecommendationDismissed CostRecommendationStatus = "DISMISSED"
)

// CostRecommendation represents a suggestion to switch to a cheaper model
type CostRecommendation struct {
	ID                      uuid.UUID                `json:"id"`
	ProjectID               uuid.UUID                `json:"projectId"`
	CurrentModel            string                   `json:"currentModel"`
	RecommendedModel        string                   `json:"recommendedModel"`
	TraceCount              int                      `json:"traceCount"`
	EstimatedSavingsPerMonth float64                 `json:"estimatedSavingsPerMonth"`
	QualityImpactEstimate   float64                  `json:"qualityImpactEstimate"`
	Confidence              float64                  `json:"confidence"`
	Status                  CostRecommendationStatus `json:"status"`
	CreatedAt               time.Time                `json:"createdAt"`
}

// CostAnalysis represents a complete cost analysis for a project
type CostAnalysis struct {
	ProjectID         uuid.UUID            `json:"projectId"`
	TotalCostPeriod   float64              `json:"totalCostPeriod"`
	ModelBreakdown    []ModelCostEntry     `json:"modelBreakdown"`
	Recommendations   []CostRecommendation `json:"recommendations"`
	PotentialSavings  float64              `json:"potentialSavings"`
}

// ModelCostEntry represents per-model cost statistics
type ModelCostEntry struct {
	Model           string  `json:"model"`
	TraceCount      int     `json:"traceCount"`
	TotalCost       float64 `json:"totalCost"`
	AvgCostPerTrace float64 `json:"avgCostPerTrace"`
}

// DateRange represents a time range for analysis queries
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
