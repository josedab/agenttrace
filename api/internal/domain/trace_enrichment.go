package domain

import (
	"time"

	"github.com/google/uuid"
)

// EnrichmentRule represents a webhook trace enrichment rule
type EnrichmentRule struct {
	ID           uuid.UUID              `json:"id"`
	ProjectID    uuid.UUID              `json:"projectId"`
	Name         string                 `json:"name"`
	Enabled      bool                   `json:"enabled"`
	TriggerEvent string                 `json:"triggerEvent"`
	SourceType   string                 `json:"sourceType"`
	SourceConfig map[string]string      `json:"sourceConfig"`
	Condition    EnrichmentCondition    `json:"condition"`
	Transform    EnrichmentTransform    `json:"transform"`
	Priority     int                    `json:"priority"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// EnrichmentCondition represents a condition for triggering an enrichment rule
type EnrichmentCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// EnrichmentTransform represents the transformation to apply during enrichment
type EnrichmentTransform struct {
	TargetField  string `json:"targetField"`
	Expression   string `json:"expression"`
	DefaultValue string `json:"defaultValue"`
}

// EnrichmentSource represents a source for trace enrichment data
type EnrichmentSource struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	RequiredConfig []string `json:"requiredConfig"`
	SamplePayload  string   `json:"samplePayload"`
}

// EnrichmentExecution represents a single execution of an enrichment rule
type EnrichmentExecution struct {
	ID         uuid.UUID `json:"id"`
	RuleID     uuid.UUID `json:"ruleId"`
	TraceID    string    `json:"traceId"`
	Status     string    `json:"status"`
	Input      string    `json:"input"`
	Output     string    `json:"output"`
	Error      string    `json:"error"`
	DurationMs int64     `json:"durationMs"`
	ExecutedAt time.Time `json:"executedAt"`
}

// EnrichmentRuleInput represents input for creating or updating an enrichment rule
type EnrichmentRuleInput struct {
	Name         string                `json:"name" validate:"required"`
	TriggerEvent string                `json:"triggerEvent" validate:"required"`
	SourceType   string                `json:"sourceType" validate:"required"`
	SourceConfig map[string]string     `json:"sourceConfig,omitempty"`
	Condition    *EnrichmentCondition  `json:"condition,omitempty"`
	Transform    EnrichmentTransform   `json:"transform"`
	Priority     *int                  `json:"priority,omitempty"`
}

// EnrichmentTestInput represents input for testing an enrichment rule
type EnrichmentTestInput struct {
	RuleID  *uuid.UUID `json:"ruleId,omitempty"`
	TraceID string     `json:"traceId" validate:"required"`
	DryRun  bool       `json:"dryRun"`
}
