package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// GuardrailRepository defines repository operations for guard rules and violations
type GuardrailRepository interface {
	SaveRule(ctx context.Context, rule *domain.GuardRule) error
	GetRuleByID(ctx context.Context, id uuid.UUID) (*domain.GuardRule, error)
	UpdateRule(ctx context.Context, rule *domain.GuardRule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error
	ListRules(ctx context.Context, projectID uuid.UUID) ([]domain.GuardRule, error)
	ListEnabledRules(ctx context.Context, projectID uuid.UUID) ([]domain.GuardRule, error)
	SaveViolation(ctx context.Context, violation *domain.GuardViolation) error
	ListViolations(ctx context.Context, filter *domain.GuardViolationFilter, limit, offset int) ([]domain.GuardViolation, int64, error)
}

// GuardrailService enforces safety rules during trace ingestion
type GuardrailService struct {
	logger          *zap.Logger
	guardrailRepo   GuardrailRepository
	notificationSvc *NotificationService
}

// NewGuardrailService creates a new guardrail service
func NewGuardrailService(
	logger *zap.Logger,
	guardrailRepo GuardrailRepository,
	notificationSvc *NotificationService,
) *GuardrailService {
	return &GuardrailService{
		logger:          logger,
		guardrailRepo:   guardrailRepo,
		notificationSvc: notificationSvc,
	}
}

// CreateRule creates a new guard rule
func (s *GuardrailService) CreateRule(ctx context.Context, projectID uuid.UUID, input *domain.GuardRuleInput) (*domain.GuardRule, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	rule := &domain.GuardRule{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Type:        input.Type,
		Config:      input.Config,
		Action:      input.Action,
		Enabled:     enabled,
		CreatedAt:   time.Now(),
	}

	if err := s.guardrailRepo.SaveRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to save guard rule: %w", err)
	}

	s.logger.Info("created guard rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("name", rule.Name),
		zap.String("type", string(rule.Type)),
	)

	return rule, nil
}

// UpdateRule updates an existing guard rule
func (s *GuardrailService) UpdateRule(ctx context.Context, ruleID uuid.UUID, input *domain.GuardRuleInput) (*domain.GuardRule, error) {
	rule, err := s.guardrailRepo.GetRuleByID(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guard rule: %w", err)
	}

	rule.Name = input.Name
	rule.Description = input.Description
	rule.Type = input.Type
	rule.Config = input.Config
	rule.Action = input.Action
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}

	if err := s.guardrailRepo.UpdateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update guard rule: %w", err)
	}

	return rule, nil
}

// DeleteRule deletes a guard rule
func (s *GuardrailService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	if err := s.guardrailRepo.DeleteRule(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete guard rule: %w", err)
	}
	return nil
}

// ListRules retrieves all guard rules for a project
func (s *GuardrailService) ListRules(ctx context.Context, projectID uuid.UUID) ([]domain.GuardRule, error) {
	rules, err := s.guardrailRepo.ListRules(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list guard rules: %w", err)
	}
	return rules, nil
}

// Evaluate checks all enabled rules for a project against a trace and its observations
func (s *GuardrailService) Evaluate(ctx context.Context, projectID uuid.UUID, trace *domain.Trace, observations []domain.Observation) (*domain.GuardEvalResult, error) {
	rules, err := s.guardrailRepo.ListEnabledRules(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled rules: %w", err)
	}

	result := &domain.GuardEvalResult{
		Passed:     true,
		Violations: []domain.GuardViolation{},
	}

	for _, rule := range rules {
		violations := s.evaluateRule(&rule, trace, observations)
		for _, v := range violations {
			v.ID = uuid.New()
			v.ProjectID = projectID
			v.RuleID = rule.ID
			v.CreatedAt = time.Now()

			result.Violations = append(result.Violations, v)

			if rule.Action == domain.GuardActionBlock {
				result.Passed = false
			}

			// Persist violation
			if err := s.guardrailRepo.SaveViolation(ctx, &v); err != nil {
				s.logger.Warn("failed to save violation",
					zap.String("ruleId", rule.ID.String()),
					zap.Error(err),
				)
			}
		}
	}

	return result, nil
}

// evaluateRule checks a single rule against a trace
func (s *GuardrailService) evaluateRule(rule *domain.GuardRule, trace *domain.Trace, observations []domain.Observation) []domain.GuardViolation {
	switch rule.Type {
	case domain.GuardRuleTypeCostLimit:
		return s.evaluateCostLimit(rule, trace)
	case domain.GuardRuleTypeLatencyLimit:
		return s.evaluateLatencyLimit(rule, trace)
	case domain.GuardRuleTypePatternBlock:
		return s.evaluatePatternBlock(rule, trace, observations)
	default:
		return nil
	}
}

// evaluateCostLimit checks if trace cost exceeds the limit
func (s *GuardrailService) evaluateCostLimit(rule *domain.GuardRule, trace *domain.Trace) []domain.GuardViolation {
	if rule.Config.MaxCostPerTrace == nil {
		return nil
	}

	if trace.TotalCost > *rule.Config.MaxCostPerTrace {
		return []domain.GuardViolation{
			{
				TraceID:  trace.ID,
				Severity: domain.GuardViolationSeverityWarning,
				Details:  fmt.Sprintf("trace cost $%.4f exceeds limit $%.4f", trace.TotalCost, *rule.Config.MaxCostPerTrace),
				Action:   rule.Action,
			},
		}
	}
	return nil
}

// evaluateLatencyLimit checks if trace latency exceeds the limit
func (s *GuardrailService) evaluateLatencyLimit(rule *domain.GuardRule, trace *domain.Trace) []domain.GuardViolation {
	if rule.Config.MaxLatencyMs == nil {
		return nil
	}

	if trace.DurationMs > float64(*rule.Config.MaxLatencyMs) {
		return []domain.GuardViolation{
			{
				TraceID:  trace.ID,
				Severity: domain.GuardViolationSeverityWarning,
				Details:  fmt.Sprintf("trace latency %.0fms exceeds limit %dms", trace.DurationMs, *rule.Config.MaxLatencyMs),
				Action:   rule.Action,
			},
		}
	}
	return nil
}

// evaluatePatternBlock checks trace data for blocked patterns
func (s *GuardrailService) evaluatePatternBlock(rule *domain.GuardRule, trace *domain.Trace, observations []domain.Observation) []domain.GuardViolation {
	var violations []domain.GuardViolation

	for _, pattern := range rule.Config.BlockedPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		for _, obs := range observations {
			if re.MatchString(obs.Input) || re.MatchString(obs.Output) {
				violations = append(violations, domain.GuardViolation{
					TraceID:  trace.ID,
					Severity: domain.GuardViolationSeverityCritical,
					Details:  fmt.Sprintf("blocked pattern %q matched in observation %s", pattern, obs.ID),
					Action:   rule.Action,
				})
			}
		}
	}

	return violations
}

// ListViolations retrieves violations matching the given filter
func (s *GuardrailService) ListViolations(ctx context.Context, projectID uuid.UUID, filter *domain.GuardViolationFilter) ([]domain.GuardViolation, error) {
	filter.ProjectID = projectID
	violations, _, err := s.guardrailRepo.ListViolations(ctx, filter, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list violations: %w", err)
	}
	return violations, nil
}

// GetViolationStats computes aggregated violation statistics for a project
func (s *GuardrailService) GetViolationStats(ctx context.Context, projectID uuid.UUID) (*domain.GuardViolationStats, error) {
	filter := &domain.GuardViolationFilter{ProjectID: projectID}
	violations, totalCount, err := s.guardrailRepo.ListViolations(ctx, filter, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list violations for stats: %w", err)
	}

	stats := &domain.GuardViolationStats{
		ProjectID:  projectID,
		TotalCount: totalCount,
		BySeverity: make(map[domain.GuardViolationSeverity]int),
		ByRule:     make(map[string]int),
		ByAction:   make(map[domain.GuardAction]int),
	}

	for _, v := range violations {
		stats.BySeverity[v.Severity]++
		stats.ByRule[v.RuleID.String()]++
		stats.ByAction[v.Action]++
	}

	return stats, nil
}
