package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CostBudgetRepository defines repository operations for cost budget management
type CostBudgetRepository interface {
	Save(ctx context.Context, budget *domain.CostBudget) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CostBudget, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.CostBudget, error)
	Update(ctx context.Context, budget *domain.CostBudget) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CostBudgetService manages cost budgets and forecasting, including budget
// creation, threshold evaluation, and spend projections.
type CostBudgetService struct {
	logger       *zap.Logger
	budgetRepo   CostBudgetRepository
	queryService *QueryService
	costService  *CostService
}

// NewCostBudgetService creates a new cost budget service
func NewCostBudgetService(
	logger *zap.Logger,
	budgetRepo CostBudgetRepository,
	queryService *QueryService,
	costService *CostService,
) *CostBudgetService {
	return &CostBudgetService{
		logger:       logger,
		budgetRepo:   budgetRepo,
		queryService: queryService,
		costService:  costService,
	}
}

// CreateBudget creates a new cost budget for a project.
func (s *CostBudgetService) CreateBudget(ctx context.Context, projectID uuid.UUID, input domain.CostBudgetInput) (*domain.CostBudget, error) {
	now := time.Now()
	budget := &domain.CostBudget{
		ID:                projectID,
		ProjectID:         projectID,
		Name:              input.Name,
		MonthlyLimitCents: input.MonthlyLimitCents,
		AlertThresholds:   input.AlertThresholds,
		AutoAction:        input.AutoAction,
		Enabled:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	budget.ID = uuid.New()

	if err := s.budgetRepo.Save(ctx, budget); err != nil {
		return nil, fmt.Errorf("failed to save cost budget: %w", err)
	}

	s.logger.Info("created cost budget",
		zap.String("projectId", projectID.String()),
		zap.String("budgetId", budget.ID.String()),
		zap.Int64("monthlyLimitCents", input.MonthlyLimitCents),
	)

	return budget, nil
}

// GetBudget retrieves a cost budget by ID.
func (s *CostBudgetService) GetBudget(ctx context.Context, budgetID uuid.UUID) (*domain.CostBudget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost budget: %w", err)
	}
	return budget, nil
}

// ListBudgets retrieves all cost budgets for a project.
func (s *CostBudgetService) ListBudgets(ctx context.Context, projectID uuid.UUID) ([]domain.CostBudget, error) {
	budgets, err := s.budgetRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost budgets: %w", err)
	}
	return budgets, nil
}

// UpdateBudget updates an existing cost budget.
func (s *CostBudgetService) UpdateBudget(ctx context.Context, budgetID uuid.UUID, input domain.CostBudgetInput) (*domain.CostBudget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost budget for update: %w", err)
	}

	budget.Name = input.Name
	budget.MonthlyLimitCents = input.MonthlyLimitCents
	budget.AlertThresholds = input.AlertThresholds
	budget.AutoAction = input.AutoAction
	budget.UpdatedAt = time.Now()

	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, fmt.Errorf("failed to update cost budget: %w", err)
	}

	s.logger.Info("updated cost budget",
		zap.String("budgetId", budgetID.String()),
	)

	return budget, nil
}

// DeleteBudget deletes a cost budget by ID.
func (s *CostBudgetService) DeleteBudget(ctx context.Context, budgetID uuid.UUID) error {
	if err := s.budgetRepo.Delete(ctx, budgetID); err != nil {
		return fmt.Errorf("failed to delete cost budget: %w", err)
	}

	s.logger.Info("deleted cost budget",
		zap.String("budgetId", budgetID.String()),
	)

	return nil
}

// GetForecast generates a cost forecast for a project by aggregating daily
// spend for the last 30 days, computing a 7-day burn rate, and projecting
// the monthly total and budget exhaustion date.
func (s *CostBudgetService) GetForecast(ctx context.Context, projectID uuid.UUID) (*domain.BudgetForecast, error) {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Aggregate daily spend for the last 30 days
	var dailySpend []domain.DailySpendPoint
	for d := 0; d < 30; d++ {
		dayStart := thirtyDaysAgo.AddDate(0, 0, d)
		dayEnd := dayStart.AddDate(0, 0, 1)

		filter := &domain.TraceFilter{
			ProjectID: projectID,
			FromTime:  &dayStart,
			ToTime:    &dayEnd,
		}
		traces, err := s.queryService.ListTraces(ctx, filter, 10000, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list traces for forecast: %w", err)
		}

		var dayCostCents int64
		traceCount := 0
		if traces != nil {
			traceCount = len(traces.Traces)
			for _, t := range traces.Traces {
				dayCostCents += int64(t.TotalCost * 100)
			}
		}

		dailySpend = append(dailySpend, domain.DailySpendPoint{
			Date:        dayStart.Format("2006-01-02"),
			AmountCents: dayCostCents,
			TraceCount:  traceCount,
		})
	}

	// Compute daily burn rate (average of last 7 days)
	var recentTotal int64
	recentDays := 7
	if len(dailySpend) < recentDays {
		recentDays = len(dailySpend)
	}
	for i := len(dailySpend) - recentDays; i < len(dailySpend); i++ {
		recentTotal += dailySpend[i].AmountCents
	}
	var dailyRate float64
	if recentDays > 0 {
		dailyRate = float64(recentTotal) / float64(recentDays)
	}

	// Project monthly total
	projectedMonthly := dailyRate * 30

	// Calculate exhaustion date if a budget exists
	var exhaustionDate *time.Time
	budgets, err := s.budgetRepo.ListByProject(ctx, projectID)
	if err == nil && len(budgets) > 0 && dailyRate > 0 {
		budget := budgets[0]
		remaining := float64(budget.MonthlyLimitCents) - float64(budget.CurrentSpendCents)
		if remaining > 0 {
			daysLeft := remaining / dailyRate
			t := now.AddDate(0, 0, int(daysLeft))
			exhaustionDate = &t
		}
	}

	confidence := 0.5
	if len(dailySpend) >= 14 {
		confidence = 0.8
	} else if len(dailySpend) >= 7 {
		confidence = 0.65
	}

	forecast := &domain.BudgetForecast{
		ProjectID:             projectID,
		CurrentDailyRate:      dailyRate / 100, // convert cents to dollars
		ProjectedMonthlyTotal: projectedMonthly / 100,
		ExhaustionDate:        exhaustionDate,
		DataPointsDays:        len(dailySpend),
		Confidence:            confidence,
		DailySpend:            dailySpend,
	}

	s.logger.Info("generated cost forecast",
		zap.String("projectId", projectID.String()),
		zap.Float64("dailyRate", forecast.CurrentDailyRate),
		zap.Float64("projectedMonthly", forecast.ProjectedMonthlyTotal),
	)

	return forecast, nil
}

// CheckBudget checks whether adding the given cost would exceed the project's
// budget and returns the configured auto-action if the limit is breached.
func (s *CostBudgetService) CheckBudget(ctx context.Context, projectID uuid.UUID, additionalCostCents int64) (bool, domain.BudgetAutoAction, error) {
	budgets, err := s.budgetRepo.ListByProject(ctx, projectID)
	if err != nil {
		return true, domain.BudgetAutoActionNone, fmt.Errorf("failed to list budgets for check: %w", err)
	}

	// If no budgets exist, allow the action
	if len(budgets) == 0 {
		return true, domain.BudgetAutoActionNone, nil
	}

	budget := budgets[0]
	if !budget.Enabled {
		return true, domain.BudgetAutoActionNone, nil
	}

	newSpend := budget.CurrentSpendCents + additionalCostCents
	if newSpend > budget.MonthlyLimitCents {
		s.logger.Warn("budget exceeded",
			zap.String("projectId", projectID.String()),
			zap.Int64("currentSpend", budget.CurrentSpendCents),
			zap.Int64("additionalCost", additionalCostCents),
			zap.Int64("limit", budget.MonthlyLimitCents),
		)
		return false, budget.AutoAction, nil
	}

	return true, domain.BudgetAutoActionNone, nil
}

// EvaluateThresholds checks which budget alert thresholds have been crossed
// but not yet notified and returns them.
func (s *CostBudgetService) EvaluateThresholds(ctx context.Context, projectID uuid.UUID) ([]domain.BudgetThreshold, error) {
	budgets, err := s.budgetRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets for threshold evaluation: %w", err)
	}

	var crossed []domain.BudgetThreshold

	for _, budget := range budgets {
		if !budget.Enabled || budget.MonthlyLimitCents == 0 {
			continue
		}

		usagePercent := int(float64(budget.CurrentSpendCents) / float64(budget.MonthlyLimitCents) * 100)

		for i := range budget.AlertThresholds {
			threshold := &budget.AlertThresholds[i]
			if usagePercent >= threshold.Percent && !threshold.Notified {
				crossed = append(crossed, *threshold)

				// Mark as notified
				threshold.Notified = true
			}
		}

		// Persist updated thresholds
		if len(crossed) > 0 {
			budget.UpdatedAt = time.Now()
			if err := s.budgetRepo.Update(ctx, &budget); err != nil {
				s.logger.Warn("failed to update budget thresholds",
					zap.String("budgetId", budget.ID.String()),
					zap.Error(err),
				)
			}
		}
	}

	s.logger.Debug("evaluated budget thresholds",
		zap.String("projectId", projectID.String()),
		zap.Int("crossedCount", len(crossed)),
	)

	return crossed, nil
}
