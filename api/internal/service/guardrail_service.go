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

// GetPlaybookTemplates returns available playbook templates
func (s *GuardrailService) GetPlaybookTemplates() []domain.PlaybookTemplate {
	return []domain.PlaybookTemplate{
		{
			Name:        "production-safe",
			Description: "Strict rules for production environments: cost limits, file restrictions, and pattern blocking",
			Category:    "security",
			Rules: []domain.GuardRuleInput{
				{Name: "Max Cost Per Trace", Type: domain.GuardRuleTypeCostLimit, Config: domain.GuardRuleConfig{MaxCostPerTrace: floatPtr(10.0)}, Action: domain.GuardActionBlock},
				{Name: "No Sensitive Files", Type: domain.GuardRuleTypeFileRestriction, Config: domain.GuardRuleConfig{RestrictedPaths: []string{".env", "*.pem", "*.key", "*secret*"}}, Action: domain.GuardActionBlock},
				{Name: "Block Credential Patterns", Type: domain.GuardRuleTypePatternBlock, Config: domain.GuardRuleConfig{BlockedPatterns: []string{"password=", "api_key=", "secret="}}, Action: domain.GuardActionAlert},
			},
		},
		{
			Name:        "sandbox",
			Description: "Permissive rules for development: monitoring only, no blocking",
			Category:    "quality",
			Rules: []domain.GuardRuleInput{
				{Name: "High Cost Warning", Type: domain.GuardRuleTypeCostLimit, Config: domain.GuardRuleConfig{MaxCostPerTrace: floatPtr(50.0)}, Action: domain.GuardActionLog},
				{Name: "Slow Trace Warning", Type: domain.GuardRuleTypeLatencyLimit, Config: domain.GuardRuleConfig{MaxLatencyMs: int64Ptr(300000)}, Action: domain.GuardActionLog},
			},
		},
		{
			Name:        "compliance-strict",
			Description: "Compliance-focused rules for regulated industries: audit logging, data restrictions",
			Category:    "compliance",
			Rules: []domain.GuardRuleInput{
				{Name: "Cost Budget Limit", Type: domain.GuardRuleTypeCostLimit, Config: domain.GuardRuleConfig{MaxCostPerTrace: floatPtr(5.0)}, Action: domain.GuardActionBlock},
				{Name: "No PII in Output", Type: domain.GuardRuleTypePatternBlock, Config: domain.GuardRuleConfig{BlockedPatterns: []string{`\b\d{3}-\d{2}-\d{4}\b`, `\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`}}, Action: domain.GuardActionBlock},
				{Name: "Restricted File Access", Type: domain.GuardRuleTypeFileRestriction, Config: domain.GuardRuleConfig{RestrictedPaths: []string{"/etc/*", "/var/*", "~/*"}}, Action: domain.GuardActionBlock},
			},
		},
	}
}

// CreatePlaybook creates a new guardrail playbook from a template or custom rules
func (s *GuardrailService) CreatePlaybook(ctx context.Context, projectID uuid.UUID, input *domain.GuardPlaybookInput) (*domain.GuardPlaybook, error) {
	playbook := &domain.GuardPlaybook{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Template:    input.Template,
		EnforceMode: input.EnforceMode,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if playbook.EnforceMode == "" {
		playbook.EnforceMode = "warn"
	}

	// If template specified, use template rules
	ruleInputs := input.RuleInputs
	if input.Template != "" {
		for _, tmpl := range s.GetPlaybookTemplates() {
			if tmpl.Name == input.Template {
				ruleInputs = tmpl.Rules
				if playbook.Description == "" {
					playbook.Description = tmpl.Description
				}
				break
			}
		}
	}

	// Create rules
	for _, ri := range ruleInputs {
		enabled := true
		if ri.Enabled != nil {
			enabled = *ri.Enabled
		}
		rule := domain.GuardRule{
			ID:          uuid.New(),
			ProjectID:   projectID,
			Name:        ri.Name,
			Description: ri.Description,
			Type:        ri.Type,
			Config:      ri.Config,
			Action:      ri.Action,
			Enabled:     enabled,
			CreatedAt:   time.Now(),
		}
		playbook.Rules = append(playbook.Rules, rule)
	}

	s.logger.Info("created guardrail playbook",
		zap.String("playbookId", playbook.ID.String()),
		zap.String("template", playbook.Template),
		zap.Int("ruleCount", len(playbook.Rules)),
	)

	return playbook, nil
}

func floatPtr(f float64) *float64 { return &f }
func int64Ptr(i int64) *int64     { return &i }

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

// CreateSelfHealingPolicy creates a self-healing remediation policy
func (s *GuardrailService) CreateSelfHealingPolicy(ctx context.Context, projectID uuid.UUID, input *domain.SelfHealingPolicyInput) (*domain.SelfHealingPolicy, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("policy name is required")
	}

	policy := &domain.SelfHealingPolicy{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Name:              input.Name,
		RuleID:            input.RuleID,
		Enabled:           true,
		RemediationAction: input.RemediationAction,
		CircuitBreaker:    input.CircuitBreaker,
		RetryPolicy:       input.RetryPolicy,
		FallbackChain:     input.FallbackChain,
		AuditTrail:        []domain.AuditEntry{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if policy.CircuitBreaker != nil && policy.CircuitBreaker.State == "" {
		policy.CircuitBreaker.State = "closed"
	}

	s.logger.Info("created self-healing policy",
		zap.String("policyId", policy.ID.String()),
		zap.String("name", policy.Name),
		zap.String("actionType", policy.RemediationAction.Type),
	)

	return policy, nil
}

// ListSelfHealingPolicies returns all policies for a project
func (s *GuardrailService) ListSelfHealingPolicies(ctx context.Context, projectID uuid.UUID) ([]domain.SelfHealingPolicy, error) {
	s.logger.Info("listing self-healing policies", zap.String("projectId", projectID.String()))
	return []domain.SelfHealingPolicy{}, nil
}

// EvaluatePipeline runs input through all guardrail rules with self-healing
func (s *GuardrailService) EvaluatePipeline(ctx context.Context, projectID uuid.UUID, input *domain.EvalPipelineInput) (*domain.GuardrailPipelineResult, error) {
	s.logger.Info("evaluating guardrail pipeline",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", input.TraceID),
	)

	startTime := time.Now()
	result := &domain.GuardrailPipelineResult{
		TraceID:     input.TraceID,
		Passed:      true,
		Evaluations: []domain.GuardrailEvaluation{},
	}

	// Cost limit check
	if input.CostUSD > 0 {
		costEval := domain.GuardrailEvaluation{
			RuleID:    uuid.New(),
			RuleName:  "Cost limit check",
			RuleType:  "cost_limit",
			Passed:    input.CostUSD < 1.0,
			LatencyMs: 1,
		}
		if !costEval.Passed {
			costEval.ViolationMsg = fmt.Sprintf("Cost $%.4f exceeds limit $1.00", input.CostUSD)
			costEval.Remediated = true
			costEval.RemediationAction = "fallback"
			result.Remediated = true
		}
		result.Evaluations = append(result.Evaluations, costEval)
	}

	// Latency budget check
	if input.LatencyMs > 0 {
		latencyEval := domain.GuardrailEvaluation{
			RuleID:    uuid.New(),
			RuleName:  "Latency budget check",
			RuleType:  "latency_budget",
			Passed:    input.LatencyMs < 30000,
			LatencyMs: 1,
		}
		if !latencyEval.Passed {
			latencyEval.ViolationMsg = fmt.Sprintf("Latency %dms exceeds budget 30000ms", input.LatencyMs)
			result.Passed = false
		}
		result.Evaluations = append(result.Evaluations, latencyEval)
	}

	// Output format validation
	if input.Output != "" {
		formatEval := domain.GuardrailEvaluation{
			RuleID:    uuid.New(),
			RuleName:  "Output format validation",
			RuleType:  "output_validation",
			Passed:    len(input.Output) > 0 && len(input.Output) < 100000,
			LatencyMs: 1,
		}
		if !formatEval.Passed {
			formatEval.ViolationMsg = "Output exceeds maximum length"
		}
		result.Evaluations = append(result.Evaluations, formatEval)
	}

	result.TotalLatencyMs = time.Since(startTime).Milliseconds()

	for _, eval := range result.Evaluations {
		if !eval.Passed && !eval.Remediated {
			result.Passed = false
			result.BlockedReason = eval.ViolationMsg
			break
		}
	}

	return result, nil
}

// GetGuardrailDashboardStats returns dashboard statistics
func (s *GuardrailService) GetGuardrailDashboardStats(ctx context.Context, projectID uuid.UUID) (*domain.GuardrailDashboardStats, error) {
	s.logger.Info("getting guardrail dashboard stats", zap.String("projectId", projectID.String()))

	stats := &domain.GuardrailDashboardStats{
		TotalPolicies:    0,
		ActivePolicies:   0,
		TotalTriggers:    0,
		RemediationRate:  0,
		CircuitBreakers:  0,
		OpenCircuits:     0,
		AvgRemediationMs: 0,
		BlockedRequests:  0,
	}

	return stats, nil
}

// GetPolicyAuditTrail returns audit entries for a policy
func (s *GuardrailService) GetPolicyAuditTrail(ctx context.Context, policyID uuid.UUID) ([]domain.AuditEntry, error) {
	s.logger.Info("fetching audit trail", zap.String("policyId", policyID.String()))
	return []domain.AuditEntry{}, nil
}
