package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostForecastPlan represents a cost forecast plan for a project
type CostForecastPlan struct {
	ID           uuid.UUID           `json:"id"`
	ProjectID    uuid.UUID           `json:"projectId"`
	Period       string              `json:"period"`
	ForecastDays int                 `json:"forecastDays"`
	DataPoints   []ForecastDataPoint `json:"dataPoints"`
	Confidence   float64             `json:"confidence"`
	Model        string              `json:"model"`
	GeneratedAt  time.Time           `json:"generatedAt"`
}

// ForecastDataPoint represents a single data point in a cost forecast
type ForecastDataPoint struct {
	Date          time.Time `json:"date"`
	PredictedCost float64  `json:"predictedCost"`
	LowerBound    float64  `json:"lowerBound"`
	UpperBound    float64  `json:"upperBound"`
	ActualCost    *float64 `json:"actualCost,omitempty"`
}

// WhatIfScenario represents a what-if cost simulation scenario
type WhatIfScenario struct {
	ID             uuid.UUID            `json:"id"`
	ProjectID      uuid.UUID            `json:"projectId"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	Changes        []ModelRoutingChange `json:"changes"`
	BaselineCost   float64              `json:"baselineCost"`
	ProjectedCost  float64              `json:"projectedCost"`
	Savings        float64              `json:"savings"`
	SavingsPercent float64              `json:"savingsPercent"`
	QualityImpact  float64              `json:"qualityImpact"`
	CreatedAt      time.Time            `json:"createdAt"`
}

// ModelRoutingChange represents a change in model routing for a what-if scenario
type ModelRoutingChange struct {
	FromModel               string  `json:"fromModel"`
	ToModel                 string  `json:"toModel"`
	TrafficPercent          float64 `json:"trafficPercent"`
	EstimatedCostPerRequest float64 `json:"estimatedCostPerRequest"`
	EstimatedQualityDelta   float64 `json:"estimatedQualityDelta"`
}

// WhatIfInput represents input for creating a what-if scenario
type WhatIfInput struct {
	Name       string               `json:"name"`
	Changes    []ModelRoutingChange `json:"changes"`
	PeriodDays int                  `json:"periodDays"`
}

// BudgetPlan represents a budget plan for a project
type BudgetPlan struct {
	ID               uuid.UUID          `json:"id"`
	ProjectID        uuid.UUID          `json:"projectId"`
	Name             string             `json:"name"`
	MonthlyBudget    float64            `json:"monthlyBudget"`
	AlertThresholds  []float64          `json:"alertThresholds"`
	ModelAllocations map[string]float64 `json:"modelAllocations"`
	Status           string             `json:"status"`
	StartDate        time.Time          `json:"startDate"`
	EndDate          time.Time          `json:"endDate"`
	CreatedAt        time.Time          `json:"createdAt"`
}

// BudgetPlanInput represents input for creating or updating a budget plan
type BudgetPlanInput struct {
	Name             string             `json:"name" validate:"required"`
	MonthlyBudget    float64            `json:"monthlyBudget" validate:"required"`
	AlertThresholds  []float64          `json:"alertThresholds,omitempty"`
	ModelAllocations map[string]float64 `json:"modelAllocations,omitempty"`
	StartDate        time.Time          `json:"startDate"`
	EndDate          time.Time          `json:"endDate"`
}

// CostHistory represents historical cost data for a project
type CostHistory struct {
	Period       string             `json:"period"`
	DataPoints   []CostHistoryPoint `json:"dataPoints"`
	TotalCost    float64            `json:"totalCost"`
	AvgDailyCost float64            `json:"avgDailyCost"`
}

// CostHistoryPoint represents a single data point in cost history
type CostHistoryPoint struct {
	Date         time.Time `json:"date"`
	Cost         float64   `json:"cost"`
	Tokens       int64     `json:"tokens"`
	RequestCount int       `json:"requestCount"`
	TopModel     string    `json:"topModel"`
}
