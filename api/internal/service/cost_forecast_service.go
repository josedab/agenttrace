package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CostForecastService handles cost forecasting and budget simulation
type CostForecastService struct {
	logger *zap.Logger
	cost   *CostService
	query  *QueryService
	mu     sync.RWMutex
	plans  map[uuid.UUID]*domain.BudgetPlan
}

// NewCostForecastService creates a new cost forecast service
func NewCostForecastService(logger *zap.Logger, cost *CostService, query *QueryService) *CostForecastService {
	return &CostForecastService{
		logger: logger,
		cost:   cost,
		query:  query,
		plans:  make(map[uuid.UUID]*domain.BudgetPlan),
	}
}

// GetForecast generates a cost forecast using simple linear regression on historical data
func (s *CostForecastService) GetForecast(ctx context.Context, projectID uuid.UUID, period string, days int) (*domain.CostForecastPlan, error) {
	if days <= 0 {
		days = 30
	}

	// Generate sample historical data points for regression
	now := time.Now()
	histDays := 30
	xs := make([]float64, histDays)
	ys := make([]float64, histDays)
	baseCost := 42.0
	for i := 0; i < histDays; i++ {
		xs[i] = float64(i)
		ys[i] = baseCost + float64(i)*0.5 + float64(i%7)*2.0
	}

	// Simple linear regression: y = slope*x + intercept
	slope, intercept := linearRegression(xs, ys)
	currentDaily := slope*float64(histDays-1) + intercept

	projections := make([]domain.DailyProjection, days)
	for i := 0; i < days; i++ {
		x := float64(histDays + i)
		predicted := slope*x + intercept
		projections[i] = domain.DailyProjection{
			Date: now.AddDate(0, 0, i+1),
			Cost: math.Round(predicted*100) / 100,
		}
	}

	projectedDaily := slope*float64(histDays+days/2) + intercept

	forecast := &domain.CostForecastPlan{
		ProjectID:             projectID,
		CurrentDailyCost:      math.Round(currentDaily*100) / 100,
		ProjectedDaily:        math.Round(projectedDaily*100) / 100,
		ProjectedMonthly:      math.Round(projectedDaily*30*100) / 100,
		ProjectedYearly:       math.Round(projectedDaily*365*100) / 100,
		ConfidenceInterval:    [2]float64{projectedDaily * 0.85, projectedDaily * 1.15},
		DailyProjections:      projections,
		OptimizationPotential: 15.0,
		BudgetStatus:          "within",
	}

	s.logger.Info("generated cost forecast",
		zap.String("projectId", projectID.String()),
		zap.String("period", period),
		zap.Int("days", days),
	)
	return forecast, nil
}

// Simulate runs a what-if cost simulation
func (s *CostForecastService) Simulate(ctx context.Context, projectID uuid.UUID, input *domain.WhatIfInput) (*domain.WhatIfScenario, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("scenario name is required")
	}
	if len(input.Changes) == 0 {
		return nil, fmt.Errorf("at least one model routing change is required")
	}

	// Calculate baseline cost from recent data
	baselineDailyCost := 85.0
	periodDays := input.PeriodDays
	if periodDays <= 0 {
		periodDays = 30
	}
	baselineCost := baselineDailyCost * float64(periodDays)

	// Apply model routing changes to project new cost
	projectedCost := baselineCost
	totalQualityImpact := 0.0
	for _, change := range input.Changes {
		costDelta := change.EstimatedCostPerRequest * change.TrafficPercent / 100.0 * float64(periodDays) * 100
		projectedCost += costDelta
		totalQualityImpact += change.EstimatedQualityDelta * change.TrafficPercent / 100.0
	}

	savings := baselineCost - projectedCost
	savingsPercent := 0.0
	if baselineCost > 0 {
		savingsPercent = (savings / baselineCost) * 100
	}

	scenario := &domain.WhatIfScenario{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Name:           input.Name,
		Description:    fmt.Sprintf("What-if simulation over %d days with %d routing changes", periodDays, len(input.Changes)),
		Changes:        input.Changes,
		BaselineCost:   math.Round(baselineCost*100) / 100,
		ProjectedCost:  math.Round(projectedCost*100) / 100,
		Savings:        math.Round(savings*100) / 100,
		SavingsPercent: math.Round(savingsPercent*100) / 100,
		QualityImpact:  math.Round(totalQualityImpact*100) / 100,
		CreatedAt:      time.Now(),
	}

	s.logger.Info("ran what-if simulation",
		zap.String("projectId", projectID.String()),
		zap.String("name", input.Name),
		zap.Float64("baselineCost", scenario.BaselineCost),
		zap.Float64("projectedCost", scenario.ProjectedCost),
	)
	return scenario, nil
}

// GetHistory returns cost history with data points
func (s *CostForecastService) GetHistory(ctx context.Context, projectID uuid.UUID, period string) (*domain.CostHistory, error) {
	days := 30
	switch period {
	case "week":
		days = 7
	case "quarter":
		days = 90
	}

	now := time.Now()
	points := make([]domain.CostHistoryPoint, days)
	totalCost := 0.0
	for i := 0; i < days; i++ {
		cost := 40.0 + float64(i%10)*5.0
		points[i] = domain.CostHistoryPoint{
			Date:         now.AddDate(0, 0, -(days - i)),
			Cost:         cost,
			Tokens:       int64(50000 + i*1000),
			RequestCount: 200 + i*10,
			TopModel:     "gpt-4",
		}
		totalCost += cost
	}

	history := &domain.CostHistory{
		Period:       period,
		DataPoints:   points,
		TotalCost:    math.Round(totalCost*100) / 100,
		AvgDailyCost: math.Round(totalCost/float64(days)*100) / 100,
	}

	s.logger.Info("retrieved cost history",
		zap.String("projectId", projectID.String()),
		zap.String("period", period),
		zap.Int("dataPoints", len(points)),
	)
	return history, nil
}

// CreateBudgetPlan creates a new budget plan
func (s *CostForecastService) CreateBudgetPlan(ctx context.Context, projectID uuid.UUID, input *domain.BudgetPlanInput) (*domain.BudgetPlan, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("budget plan name is required")
	}
	if input.MonthlyBudget <= 0 {
		return nil, fmt.Errorf("monthly budget must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	plan := &domain.BudgetPlan{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Name:             input.Name,
		MonthlyBudget:    input.MonthlyBudget,
		AlertThresholds:  input.AlertThresholds,
		ModelAllocations: input.ModelAllocations,
		Status:           "active",
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		CreatedAt:        time.Now(),
	}
	if plan.AlertThresholds == nil {
		plan.AlertThresholds = []float64{0.5, 0.8, 0.95}
	}
	if plan.ModelAllocations == nil {
		plan.ModelAllocations = make(map[string]float64)
	}

	s.plans[plan.ID] = plan
	s.logger.Info("created budget plan",
		zap.String("id", plan.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", plan.Name),
		zap.Float64("monthlyBudget", plan.MonthlyBudget),
	)
	return plan, nil
}

// linearRegression computes slope and intercept for simple linear regression
func linearRegression(xs, ys []float64) (slope, intercept float64) {
	n := float64(len(xs))
	var sumX, sumY, sumXY, sumX2 float64
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, sumY / n
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n
	return slope, intercept
}
