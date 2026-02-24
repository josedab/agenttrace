package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// BudgetCheckResult represents the result of a budget check against guardrail policies
type BudgetCheckResult struct {
	Allowed         bool                `json:"allowed"`
	RemainingBudget float64             `json:"remainingBudget"`
	Action          domain.GuardrailAction `json:"action"`
	Reason          string              `json:"reason"`
	SuggestedModel  string              `json:"suggestedModel,omitempty"`
}

// CostGuardrailService handles cost guardrail logic
type CostGuardrailService struct {
	logger *zap.Logger
}

// NewCostGuardrailService creates a new cost guardrail service
func NewCostGuardrailService(logger *zap.Logger) *CostGuardrailService {
	return &CostGuardrailService{
		logger: logger,
	}
}

// GetDashboard returns the cost guardrail dashboard for a project
func (s *CostGuardrailService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.CostGuardrailDashboard, error) {
	s.logger.Info("fetching cost guardrail dashboard", zap.String("projectId", projectID.String()))

	dashboard := &domain.CostGuardrailDashboard{
		ActivePolicies:   0,
		RecentViolations: []domain.CostGuardrailViolation{},
		Forecast: domain.GuardrailCostForecast{
			ProjectID:       projectID,
			Recommendations: []string{},
			ByModel:         []domain.ModelCostForecast{},
		},
		TopSpenders: []domain.SpenderInfo{},
	}

	return dashboard, nil
}

// CreatePolicy creates a new cost guardrail policy
func (s *CostGuardrailService) CreatePolicy(ctx context.Context, projectID uuid.UUID, input domain.CostGuardrailPolicyInput) (*domain.CostGuardrailPolicy, error) {
	now := time.Now()

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	cooldown := 30
	if input.CooldownMinutes != nil {
		cooldown = *input.CooldownMinutes
	}

	policy := &domain.CostGuardrailPolicy{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Name:              input.Name,
		Type:              input.Type,
		Action:            input.Action,
		Enabled:           enabled,
		BudgetLimit:       input.BudgetLimit,
		BudgetPeriod:      input.BudgetPeriod,
		ThresholdPercent:  input.ThresholdPercent,
		ModelDowngradeMap: input.ModelDowngradeMap,
		NotifyChannels:    input.NotifyChannels,
		CooldownMinutes:   cooldown,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.logger.Info("created cost guardrail policy",
		zap.String("policyId", policy.ID.String()),
		zap.String("name", policy.Name),
		zap.Float64("budgetLimit", policy.BudgetLimit),
	)

	return policy, nil
}

// CheckBudget checks whether a request is allowed under current guardrail policies.
// Implements progressive enforcement: warn at 80%, throttle at 90%, pause at 100%.
func (s *CostGuardrailService) CheckBudget(ctx context.Context, projectID uuid.UUID, traceID uuid.UUID, estimatedCost float64) (*BudgetCheckResult, error) {
	s.logger.Info("checking budget",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID.String()),
		zap.Float64("estimatedCost", estimatedCost),
	)

	// Simulated current spend and budget for demonstration
	currentSpend := 0.0
	budgetLimit := 1000.0
	remaining := budgetLimit - currentSpend

	usageRatio := (currentSpend + estimatedCost) / budgetLimit

	result := &BudgetCheckResult{
		Allowed:         true,
		RemainingBudget: remaining - estimatedCost,
	}

	switch {
	case usageRatio >= 1.0:
		// At or over budget — pause
		result.Allowed = false
		result.Action = domain.GuardrailActionPause
		result.Reason = fmt.Sprintf("budget exhausted: projected spend $%.2f exceeds limit $%.2f", currentSpend+estimatedCost, budgetLimit)
	case usageRatio >= 0.9:
		// 90-100% — throttle and suggest cheaper model
		result.Allowed = true
		result.Action = domain.GuardrailActionThrottle
		result.Reason = fmt.Sprintf("budget at %.0f%%: throttling requests, consider downgrading model", usageRatio*100)
		result.SuggestedModel = "gpt-3.5-turbo"
	case usageRatio >= 0.8:
		// 80-90% — warn
		result.Allowed = true
		result.Action = domain.GuardrailActionWarn
		result.Reason = fmt.Sprintf("budget at %.0f%%: approaching limit", usageRatio*100)
	default:
		result.Action = domain.GuardrailActionNotify
		result.Reason = "within budget"
	}

	return result, nil
}

// GetForecast returns the cost forecast for guardrail evaluation
func (s *CostGuardrailService) GetForecast(ctx context.Context, projectID uuid.UUID) (*domain.GuardrailCostForecast, error) {
	s.logger.Info("computing cost forecast", zap.String("projectId", projectID.String()))

	forecast := &domain.GuardrailCostForecast{
		ProjectID:                projectID,
		CurrentSpend:             0,
		ProjectedSpend:           0,
		BudgetRemaining:          1000.0,
		DaysUntilBudgetExhausted: 30,
		Recommendations:          []string{"Current spending is within normal range"},
		ByModel:                  []domain.ModelCostForecast{},
	}

	return forecast, nil
}

// ListPolicies returns all cost guardrail policies for a project
func (s *CostGuardrailService) ListPolicies(ctx context.Context, projectID uuid.UUID) ([]domain.CostGuardrailPolicy, error) {
	s.logger.Info("listing cost guardrail policies", zap.String("projectId", projectID.String()))
	return []domain.CostGuardrailPolicy{}, nil
}

// ListViolations returns all cost guardrail violations for a project
func (s *CostGuardrailService) ListViolations(ctx context.Context, projectID uuid.UUID) ([]domain.CostGuardrailViolation, error) {
	s.logger.Info("listing cost guardrail violations", zap.String("projectId", projectID.String()))
	return []domain.CostGuardrailViolation{}, nil
}

// EnforcePolicy evaluates a policy against current spend and creates a violation if triggered
func (s *CostGuardrailService) EnforcePolicy(ctx context.Context, policy domain.CostGuardrailPolicy, currentSpend float64) (*domain.CostGuardrailViolation, error) {
	if !policy.Enabled {
		return nil, nil
	}

	threshold := policy.BudgetLimit * (policy.ThresholdPercent / 100.0)
	if currentSpend < threshold {
		return nil, nil
	}

	violation := &domain.CostGuardrailViolation{
		ID:                uuid.New(),
		PolicyID:          policy.ID,
		ProjectID:         policy.ProjectID,
		Action:            policy.Action,
		AmountAtViolation: currentSpend,
		BudgetLimit:       policy.BudgetLimit,
		Timestamp:         time.Now(),
	}

	s.logger.Warn("guardrail policy violated",
		zap.String("policyId", policy.ID.String()),
		zap.String("action", string(policy.Action)),
		zap.Float64("spend", currentSpend),
		zap.Float64("limit", policy.BudgetLimit),
	)

	return violation, nil
}
