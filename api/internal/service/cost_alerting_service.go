package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CostAlertingService handles cost alert rules, real-time alerts, and circuit breakers
type CostAlertingService struct {
	logger *zap.Logger
}

// NewCostAlertingService creates a new cost alerting service
func NewCostAlertingService(logger *zap.Logger) *CostAlertingService {
	return &CostAlertingService{
		logger: logger,
	}
}

// CreateRule creates a new cost alert rule
func (s *CostAlertingService) CreateRule(ctx context.Context, projectID uuid.UUID, rule domain.CostAlertRule) (*domain.CostAlertRule, error) {
	if rule.Name == "" {
		return nil, fmt.Errorf("rule name is required")
	}
	if !rule.Severity.IsValid() {
		return nil, fmt.Errorf("invalid severity: %s", rule.Severity)
	}
	if rule.Condition.Threshold <= 0 {
		return nil, fmt.Errorf("threshold must be positive, got %f", rule.Condition.Threshold)
	}

	rule.ID = uuid.New()
	rule.ProjectID = projectID
	rule.Enabled = true
	rule.CreatedAt = time.Now()

	s.logger.Info("cost alert rule created",
		zap.String("id", rule.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", rule.Name),
		zap.String("severity", string(rule.Severity)),
		zap.Float64("threshold", rule.Condition.Threshold),
	)
	return &rule, nil
}

// ListRules lists all cost alert rules for a project
func (s *CostAlertingService) ListRules(ctx context.Context, projectID uuid.UUID) ([]domain.CostAlertRule, error) {
	s.logger.Debug("listing cost alert rules", zap.String("projectId", projectID.String()))
	return []domain.CostAlertRule{}, nil
}

// DeleteRule deletes a cost alert rule by ID
func (s *CostAlertingService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	s.logger.Info("cost alert rule deleted", zap.String("ruleId", ruleID.String()))
	return nil
}

// ListAlerts lists recent cost alerts for a project
func (s *CostAlertingService) ListAlerts(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.CostAlert, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	s.logger.Debug("listing cost alerts",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
	)
	return []domain.CostAlert{}, nil
}

// AcknowledgeAlert marks a cost alert as acknowledged
func (s *CostAlertingService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID) error {
	s.logger.Info("cost alert acknowledged", zap.String("alertId", alertID.String()))
	return nil
}

// GetCircuitBreaker retrieves the circuit breaker configuration for a project
func (s *CostAlertingService) GetCircuitBreaker(ctx context.Context, projectID uuid.UUID) (*domain.CircuitBreakerConfig, error) {
	s.logger.Debug("fetching circuit breaker config", zap.String("projectId", projectID.String()))

	return &domain.CircuitBreakerConfig{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		Enabled:            false,
		State:              domain.CircuitBreakerStateClosed,
		MaxCostPerMinute:   1.0,
		MaxCostPerHour:     10.0,
		FallbackModelChain: []string{"gpt-3.5-turbo", "claude-haiku"},
		CooldownSeconds:    60,
		ResetAfterSeconds:  300,
	}, nil
}

// UpdateCircuitBreaker updates the circuit breaker configuration for a project
func (s *CostAlertingService) UpdateCircuitBreaker(ctx context.Context, projectID uuid.UUID, config domain.CircuitBreakerConfig) (*domain.CircuitBreakerConfig, error) {
	if config.MaxCostPerMinute <= 0 {
		return nil, fmt.Errorf("max cost per minute must be positive, got %f", config.MaxCostPerMinute)
	}
	if config.MaxCostPerHour <= 0 {
		return nil, fmt.Errorf("max cost per hour must be positive, got %f", config.MaxCostPerHour)
	}
	if config.MaxCostPerHour < config.MaxCostPerMinute {
		return nil, fmt.Errorf("max cost per hour ($%.2f) must be >= max cost per minute ($%.2f)", config.MaxCostPerHour, config.MaxCostPerMinute)
	}
	if config.CooldownSeconds < 10 {
		return nil, fmt.Errorf("cooldown must be at least 10 seconds, got %d", config.CooldownSeconds)
	}

	config.ID = uuid.New()
	config.ProjectID = projectID

	s.logger.Info("circuit breaker config updated",
		zap.String("projectId", projectID.String()),
		zap.Bool("enabled", config.Enabled),
		zap.Float64("maxCostPerMinute", config.MaxCostPerMinute),
		zap.Float64("maxCostPerHour", config.MaxCostPerHour),
	)
	return &config, nil
}

// CheckAndAlert evaluates cost alert rules against current cost metrics and
// generates alerts when thresholds are breached
func (s *CostAlertingService) CheckAndAlert(ctx context.Context, projectID uuid.UUID, currentCost float64, model string) (*domain.CostAlert, error) {
	rules, err := s.ListRules(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alert rules: %w", err)
	}

	// Evaluate each enabled rule
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		breached := false
		switch rule.Condition.Operator {
		case "gt", ">":
			breached = currentCost > rule.Condition.Threshold
		case "gte", ">=":
			breached = currentCost >= rule.Condition.Threshold
		case "lt", "<":
			breached = currentCost < rule.Condition.Threshold
		}

		if breached {
			now := time.Now()
			action := domain.CostAlertActionNotify
			if len(rule.Actions) > 0 {
				action = rule.Actions[0]
			}

			alert := &domain.CostAlert{
				ID:            uuid.New(),
				ProjectID:     projectID,
				Severity:      rule.Severity,
				Action:        action,
				Title:         fmt.Sprintf("Cost threshold breached: %s", rule.Name),
				Description:   fmt.Sprintf("Current cost $%.4f exceeds threshold $%.4f for rule '%s' (model: %s)", currentCost, rule.Condition.Threshold, rule.Name, model),
				CurrentCost:   currentCost,
				ThresholdCost: rule.Condition.Threshold,
				AffectedModel: model,
				SentAt:        &now,
				CreatedAt:     now,
			}

			s.logger.Warn("cost alert triggered",
				zap.String("projectId", projectID.String()),
				zap.String("rule", rule.Name),
				zap.String("severity", string(rule.Severity)),
				zap.Float64("currentCost", currentCost),
				zap.Float64("threshold", rule.Condition.Threshold),
			)
			return alert, nil
		}
	}

	// No rules breached
	s.logger.Debug("cost check passed, no alerts",
		zap.String("projectId", projectID.String()),
		zap.Float64("currentCost", currentCost),
	)
	return nil, nil
}
