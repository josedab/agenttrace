package domain

import (
	"time"

	"github.com/google/uuid"
)

// WebhookRuleTrigger defines the event that triggers a webhook rule
type WebhookRuleTrigger string

const (
	TriggerCostExceeded       WebhookRuleTrigger = "cost_exceeded"
	TriggerErrorDetected      WebhookRuleTrigger = "error_detected"
	TriggerGuardrailViolation WebhookRuleTrigger = "guardrail_violation"
	TriggerAnomalyDetected    WebhookRuleTrigger = "anomaly_detected"
	TriggerEvalScoreLow       WebhookRuleTrigger = "eval_score_low"
	TriggerTraceCompleted     WebhookRuleTrigger = "trace_completed"
)

// WebhookRuleAction defines the action taken when a rule fires
type WebhookRuleAction string

const (
	ActionSlack     WebhookRuleAction = "slack"
	ActionPagerDuty WebhookRuleAction = "pagerduty"
	ActionJira      WebhookRuleAction = "jira"
	ActionGitHub    WebhookRuleAction = "github_issue"
	ActionEmail     WebhookRuleAction = "email"
	ActionCustom    WebhookRuleAction = "custom_webhook"
)

// WebhookRule defines a webhook orchestration rule
type WebhookRule struct {
	ID           uuid.UUID          `json:"id"`
	ProjectID    uuid.UUID          `json:"projectId"`
	Name         string             `json:"name"`
	Enabled      bool               `json:"enabled"`
	Trigger      WebhookRuleTrigger `json:"trigger"`
	Condition    RuleCondition      `json:"condition"`
	Action       WebhookRuleAction  `json:"action"`
	ActionConfig map[string]string  `json:"actionConfig"`
	Cooldown     int                `json:"cooldownMinutes"`
	CreatedAt    time.Time          `json:"createdAt"`
	LastFiredAt  *time.Time         `json:"lastFiredAt,omitempty"`
	FireCount    int                `json:"fireCount"`
}

// RuleCondition defines the condition under which a rule fires
type RuleCondition struct {
	Field     string   `json:"field,omitempty"`
	Operator  string   `json:"operator,omitempty"` // gt, lt, eq, contains
	Value     string   `json:"value,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
}

// WebhookRuleDelivery tracks delivery of a webhook rule firing
type WebhookRuleDelivery struct {
	ID        uuid.UUID `json:"id"`
	RuleID    uuid.UUID `json:"ruleId"`
	Status    string    `json:"status"` // success, failed, retrying
	Payload   string    `json:"payload,omitempty"`
	Response  string    `json:"response,omitempty"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"createdAt"`
}

// WebhookRuleInput represents input for creating a webhook rule
type WebhookRuleInput struct {
	Name         string             `json:"name" validate:"required"`
	Trigger      WebhookRuleTrigger `json:"trigger" validate:"required"`
	Condition    RuleCondition      `json:"condition"`
	Action       WebhookRuleAction  `json:"action" validate:"required"`
	ActionConfig map[string]string  `json:"actionConfig"`
	Cooldown     *int               `json:"cooldownMinutes,omitempty"`
}

// WebhookRuleTemplate provides a pre-built rule template
type WebhookRuleTemplate struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Trigger     WebhookRuleTrigger `json:"trigger"`
	Action      WebhookRuleAction  `json:"action"`
	Condition   RuleCondition      `json:"condition"`
}
