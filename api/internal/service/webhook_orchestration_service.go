package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// WebhookOrchestrationService manages webhook orchestration rules
type WebhookOrchestrationService struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	rules      map[uuid.UUID]*domain.WebhookRule
	deliveries map[uuid.UUID]*domain.WebhookRuleDelivery
}

// NewWebhookOrchestrationService creates a new webhook orchestration service
func NewWebhookOrchestrationService(logger *zap.Logger) *WebhookOrchestrationService {
	return &WebhookOrchestrationService{
		logger:     logger,
		rules:      make(map[uuid.UUID]*domain.WebhookRule),
		deliveries: make(map[uuid.UUID]*domain.WebhookRuleDelivery),
	}
}

// CreateRule creates a new webhook rule
func (s *WebhookOrchestrationService) CreateRule(ctx context.Context, projectID uuid.UUID, input *domain.WebhookRuleInput) (*domain.WebhookRule, error) {
	cooldown := 15
	if input.Cooldown != nil {
		cooldown = *input.Cooldown
	}

	rule := &domain.WebhookRule{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Name:         input.Name,
		Enabled:      true,
		Trigger:      input.Trigger,
		Condition:    input.Condition,
		Action:       input.Action,
		ActionConfig: input.ActionConfig,
		Cooldown:     cooldown,
		CreatedAt:    time.Now(),
		FireCount:    0,
	}

	s.mu.Lock()
	s.rules[rule.ID] = rule
	s.mu.Unlock()

	s.logger.Info("created webhook rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("name", rule.Name),
		zap.String("trigger", string(rule.Trigger)),
		zap.String("action", string(rule.Action)),
	)

	return rule, nil
}

// ListRules returns all webhook rules for a project
func (s *WebhookOrchestrationService) ListRules(ctx context.Context, projectID uuid.UUID) []domain.WebhookRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.WebhookRule
	for _, rule := range s.rules {
		if rule.ProjectID == projectID {
			result = append(result, *rule)
		}
	}
	return result
}

// DeleteRule removes a webhook rule
func (s *WebhookOrchestrationService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found")
	}
	delete(s.rules, ruleID)

	s.logger.Info("deleted webhook rule", zap.String("ruleId", ruleID.String()))
	return nil
}

// GetTemplates returns pre-built webhook rule templates
func (s *WebhookOrchestrationService) GetTemplates(ctx context.Context) []domain.WebhookRuleTemplate {
	costThreshold := 10.0
	evalThreshold := 0.7

	return []domain.WebhookRuleTemplate{
		{
			Name:        "Cost Alert → Slack",
			Description: "Notify Slack when trace cost exceeds threshold",
			Trigger:     domain.TriggerCostExceeded,
			Action:      domain.ActionSlack,
			Condition: domain.RuleCondition{
				Field:     "cost",
				Operator:  "gt",
				Threshold: &costThreshold,
			},
		},
		{
			Name:        "Error → PagerDuty",
			Description: "Page on-call when agent errors are detected",
			Trigger:     domain.TriggerErrorDetected,
			Action:      domain.ActionPagerDuty,
			Condition: domain.RuleCondition{
				Field:    "error_type",
				Operator: "contains",
				Value:    "critical",
			},
		},
		{
			Name:        "Guardrail Violation → Jira",
			Description: "Create Jira ticket for guardrail violations",
			Trigger:     domain.TriggerGuardrailViolation,
			Action:      domain.ActionJira,
			Condition: domain.RuleCondition{
				Field:    "severity",
				Operator: "eq",
				Value:    "high",
			},
		},
		{
			Name:        "Low Eval Score → Email",
			Description: "Email team when evaluation scores drop",
			Trigger:     domain.TriggerEvalScoreLow,
			Action:      domain.ActionEmail,
			Condition: domain.RuleCondition{
				Field:     "score",
				Operator:  "lt",
				Threshold: &evalThreshold,
			},
		},
	}
}

// ListDeliveries returns webhook delivery history for a project
func (s *WebhookOrchestrationService) ListDeliveries(ctx context.Context, projectID uuid.UUID) []domain.WebhookRuleDelivery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.WebhookRuleDelivery
	for _, delivery := range s.deliveries {
		// Filter deliveries by rules belonging to this project
		if rule, ok := s.rules[delivery.RuleID]; ok && rule.ProjectID == projectID {
			result = append(result, *delivery)
		}
	}
	return result
}

// TestRule simulates firing a webhook rule and records a delivery
func (s *WebhookOrchestrationService) TestRule(ctx context.Context, ruleID uuid.UUID) (*domain.WebhookRuleDelivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found")
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"test":    true,
		"ruleId":  ruleID.String(),
		"trigger": rule.Trigger,
		"action":  rule.Action,
	})

	delivery := &domain.WebhookRuleDelivery{
		ID:        uuid.New(),
		RuleID:    ruleID,
		Status:    "success",
		Payload:   string(payload),
		Response:  `{"ok": true}`,
		Attempts:  1,
		CreatedAt: time.Now(),
	}

	s.deliveries[delivery.ID] = delivery

	now := time.Now()
	rule.LastFiredAt = &now
	rule.FireCount++

	s.logger.Info("tested webhook rule",
		zap.String("ruleId", ruleID.String()),
		zap.String("deliveryId", delivery.ID.String()),
	)

	return delivery, nil
}
