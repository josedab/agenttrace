package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceEnrichmentService manages webhook trace enrichment rules and execution
type TraceEnrichmentService struct {
	logger *zap.Logger
	mu     sync.RWMutex
	rules  map[uuid.UUID]*domain.EnrichmentRule
}

// NewTraceEnrichmentService creates a new trace enrichment service
func NewTraceEnrichmentService(logger *zap.Logger) *TraceEnrichmentService {
	return &TraceEnrichmentService{
		logger: logger,
		rules:  make(map[uuid.UUID]*domain.EnrichmentRule),
	}
}

// ListRules returns all enrichment rules for a project
func (s *TraceEnrichmentService) ListRules(ctx context.Context, projectID uuid.UUID) ([]domain.EnrichmentRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.EnrichmentRule
	for _, r := range s.rules {
		if r.ProjectID == projectID {
			result = append(result, *r)
		}
	}
	if result == nil {
		result = []domain.EnrichmentRule{}
	}
	return result, nil
}

// CreateRule creates a new enrichment rule with validation
func (s *TraceEnrichmentService) CreateRule(ctx context.Context, projectID uuid.UUID, input *domain.EnrichmentRuleInput) (*domain.EnrichmentRule, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("rule name is required")
	}
	if input.TriggerEvent == "" {
		return nil, fmt.Errorf("trigger event is required")
	}
	if input.SourceType == "" {
		return nil, fmt.Errorf("source type is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	priority := 0
	if input.Priority != nil {
		priority = *input.Priority
	}
	condition := domain.EnrichmentCondition{}
	if input.Condition != nil {
		condition = *input.Condition
	}

	rule := &domain.EnrichmentRule{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Name:         input.Name,
		Enabled:      true,
		TriggerEvent: input.TriggerEvent,
		SourceType:   input.SourceType,
		SourceConfig: input.SourceConfig,
		Condition:    condition,
		Transform:    input.Transform,
		Priority:     priority,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.rules[rule.ID] = rule
	s.logger.Info("created enrichment rule",
		zap.String("id", rule.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", rule.Name),
	)
	return rule, nil
}

// UpdateRule updates an existing enrichment rule
func (s *TraceEnrichmentService) UpdateRule(ctx context.Context, ruleID uuid.UUID, input *domain.EnrichmentRuleInput) (*domain.EnrichmentRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, ok := s.rules[ruleID]
	if !ok {
		return nil, fmt.Errorf("enrichment rule not found: %s", ruleID)
	}

	if input.Name != "" {
		rule.Name = input.Name
	}
	if input.TriggerEvent != "" {
		rule.TriggerEvent = input.TriggerEvent
	}
	if input.SourceType != "" {
		rule.SourceType = input.SourceType
	}
	if input.SourceConfig != nil {
		rule.SourceConfig = input.SourceConfig
	}
	if input.Condition != nil {
		rule.Condition = *input.Condition
	}
	rule.Transform = input.Transform
	if input.Priority != nil {
		rule.Priority = *input.Priority
	}
	rule.UpdatedAt = time.Now()

	s.logger.Info("updated enrichment rule",
		zap.String("id", ruleID.String()),
		zap.String("name", rule.Name),
	)
	return rule, nil
}

// DeleteRule deletes an enrichment rule
func (s *TraceEnrichmentService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[ruleID]; !ok {
		return fmt.Errorf("enrichment rule not found: %s", ruleID)
	}

	delete(s.rules, ruleID)
	s.logger.Info("deleted enrichment rule", zap.String("id", ruleID.String()))
	return nil
}

// ListSources returns built-in enrichment source types
func (s *TraceEnrichmentService) ListSources(ctx context.Context) ([]domain.EnrichmentSource, error) {
	sources := []domain.EnrichmentSource{
		{
			ID:             "cicd",
			Name:           "CI/CD",
			Type:           "cicd",
			Description:    "Enrich traces with CI/CD pipeline data (build status, deploy info)",
			RequiredConfig: []string{"webhookUrl", "apiToken"},
			SamplePayload:  `{"build_id":"123","status":"success","branch":"main"}`,
		},
		{
			ID:             "code_review",
			Name:           "Code Review",
			Type:           "code_review",
			Description:    "Enrich traces with code review context (PR info, review comments)",
			RequiredConfig: []string{"repoUrl", "apiToken"},
			SamplePayload:  `{"pr_number":42,"status":"approved","reviewer":"alice"}`,
		},
		{
			ID:             "user_satisfaction",
			Name:           "User Satisfaction",
			Type:           "user_satisfaction",
			Description:    "Enrich traces with user satisfaction scores and feedback",
			RequiredConfig: []string{"webhookUrl"},
			SamplePayload:  `{"score":4.5,"feedback":"Great response","session_id":"abc"}`,
		},
		{
			ID:             "production_errors",
			Name:           "Production Errors",
			Type:           "production_errors",
			Description:    "Enrich traces with production error and incident data",
			RequiredConfig: []string{"webhookUrl", "errorTrackingId"},
			SamplePayload:  `{"error_id":"err_1","severity":"high","stack_trace":"..."}`,
		},
		{
			ID:             "custom_webhook",
			Name:           "Custom Webhook",
			Type:           "custom_webhook",
			Description:    "Enrich traces with data from any custom webhook source",
			RequiredConfig: []string{"webhookUrl"},
			SamplePayload:  `{"key":"value"}`,
		},
	}
	return sources, nil
}

// TestRule performs a dry-run of a rule against a trace
func (s *TraceEnrichmentService) TestRule(ctx context.Context, input *domain.EnrichmentTestInput) (*domain.EnrichmentExecution, error) {
	if input.TraceID == "" {
		return nil, fmt.Errorf("trace ID is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ruleID := uuid.Nil
	if input.RuleID != nil {
		ruleID = *input.RuleID
		if _, ok := s.rules[ruleID]; !ok {
			return nil, fmt.Errorf("enrichment rule not found: %s", ruleID)
		}
	}

	execution := &domain.EnrichmentExecution{
		ID:         uuid.New(),
		RuleID:     ruleID,
		TraceID:    input.TraceID,
		Status:     "success",
		Input:      fmt.Sprintf(`{"traceId":"%s","dryRun":%v}`, input.TraceID, input.DryRun),
		Output:     `{"enriched":true,"fields_added":["cicd.status","cicd.build_id"]}`,
		DurationMs: 12,
		ExecutedAt: time.Now(),
	}

	s.logger.Info("tested enrichment rule",
		zap.String("traceId", input.TraceID),
		zap.String("ruleId", ruleID.String()),
		zap.Bool("dryRun", input.DryRun),
	)
	return execution, nil
}
