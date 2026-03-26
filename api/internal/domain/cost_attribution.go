package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostAttribution links a trace's incurred cost to the issue or task it resolved,
// enabling ROI calculation for AI agent usage.
type CostAttribution struct {
	ID                 uuid.UUID `json:"id"`
	ProjectID          uuid.UUID `json:"projectId"`
	TraceID            string    `json:"traceId"`
	IssueRef           string    `json:"issueRef"`
	IssueTitle         string    `json:"issueTitle"`
	CostIncurred       float64   `json:"costIncurred"`
	EstimatedValueSaved float64  `json:"estimatedValueSaved"`
	HoursSaved         float64   `json:"hoursSaved"`
	HourlyRate         float64   `json:"hourlyRate"`
	ROI                float64   `json:"roi"`
	Category           string    `json:"category"` // bug_fix, feature, refactor, review, test
	CreatedAt          time.Time `json:"createdAt"`
}

// AttributionDateRange defines a time period for filtering cost attribution reports.
type AttributionDateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// AttributionReport aggregates cost attributions for a project over a given period,
// including per-category ROI breakdowns.
type AttributionReport struct {
	ProjectID       uuid.UUID              `json:"projectId"`
	Period          AttributionDateRange   `json:"period"`
	TotalCost       float64                `json:"totalCost"`
	TotalValueSaved float64                `json:"totalValueSaved"`
	OverallROI      float64                `json:"overallROI"`
	Attributions    []CostAttribution      `json:"attributions"`
	ByCategory      map[string]CategoryROI `json:"byCategory"`
}

// CategoryROI summarizes cost and value metrics for a single attribution category.
type CategoryROI struct {
	Category   string  `json:"category"`
	Cost       float64 `json:"cost"`
	Value      float64 `json:"value"`
	ROI        float64 `json:"roi"`
	TraceCount int     `json:"traceCount"`
}

// AttributionInput is the input for creating a new cost attribution.
type AttributionInput struct {
	TraceID    string  `json:"traceId" validate:"required"`
	IssueRef   string  `json:"issueRef" validate:"required"`
	IssueTitle string  `json:"issueTitle"`
	Category   string  `json:"category" validate:"required"`
	HoursSaved float64 `json:"hoursSaved" validate:"required"`
	HourlyRate float64 `json:"hourlyRate" validate:"required"`
}
