package domain

import (
	"time"

	"github.com/google/uuid"
)

// GuardrailPolicyType represents the scope of a cost guardrail policy
type GuardrailPolicyType string

const (
	GuardrailPolicyTypePerProject GuardrailPolicyType = "per_project"
	GuardrailPolicyTypePerUser    GuardrailPolicyType = "per_user"
	GuardrailPolicyTypePerModel   GuardrailPolicyType = "per_model"
	GuardrailPolicyTypePerSession GuardrailPolicyType = "per_session"
)

// GuardrailAction represents the action taken when a guardrail is triggered
type GuardrailAction string

const (
	GuardrailActionWarn           GuardrailAction = "warn"
	GuardrailActionThrottle       GuardrailAction = "throttle"
	GuardrailActionPause          GuardrailAction = "pause"
	GuardrailActionNotify         GuardrailAction = "notify"
	GuardrailActionDowngradeModel GuardrailAction = "downgrade_model"
)

// GuardrailStatus represents the status of a guardrail policy
type GuardrailStatus string

const (
	GuardrailStatusActive    GuardrailStatus = "active"
	GuardrailStatusTriggered GuardrailStatus = "triggered"
	GuardrailStatusPaused    GuardrailStatus = "paused"
	GuardrailStatusDisabled  GuardrailStatus = "disabled"
)

// CostGuardrailPolicy defines a cost guardrail policy for a project
type CostGuardrailPolicy struct {
	ID               uuid.UUID           `json:"id"`
	ProjectID        uuid.UUID           `json:"projectId"`
	Name             string              `json:"name"`
	Type             GuardrailPolicyType `json:"type"`
	Action           GuardrailAction     `json:"action"`
	Enabled          bool                `json:"enabled"`
	BudgetLimit      float64             `json:"budgetLimit"`
	BudgetPeriod     string              `json:"budgetPeriod"`
	CurrentSpend     float64             `json:"currentSpend"`
	ThresholdPercent float64             `json:"thresholdPercent"`
	ModelDowngradeMap map[string]string   `json:"modelDowngradeMap,omitempty"`
	NotifyChannels   []uuid.UUID         `json:"notifyChannels"`
	CooldownMinutes  int                 `json:"cooldownMinutes"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
}

// CostGuardrailViolation represents a guardrail policy violation event
type CostGuardrailViolation struct {
	ID                uuid.UUID       `json:"id"`
	PolicyID          uuid.UUID       `json:"policyId"`
	ProjectID         uuid.UUID       `json:"projectId"`
	TraceID           uuid.UUID       `json:"traceId"`
	UserID            uuid.UUID       `json:"userId"`
	Action            GuardrailAction `json:"action"`
	AmountAtViolation float64         `json:"amountAtViolation"`
	BudgetLimit       float64         `json:"budgetLimit"`
	Timestamp         time.Time       `json:"timestamp"`
}

// GuardrailCostForecast represents a cost forecast for guardrail evaluation
type GuardrailCostForecast struct {
	ProjectID                uuid.UUID           `json:"projectId"`
	Period                   TimeWindow          `json:"period"`
	CurrentSpend             float64             `json:"currentSpend"`
	ProjectedSpend           float64             `json:"projectedSpend"`
	BudgetRemaining          float64             `json:"budgetRemaining"`
	DaysUntilBudgetExhausted int                 `json:"daysUntilBudgetExhausted"`
	Recommendations          []string            `json:"recommendations"`
	ByModel                  []ModelCostForecast `json:"byModel"`
}

// ModelCostForecast represents cost forecast details for a specific model
type ModelCostForecast struct {
	Model          string  `json:"model"`
	CurrentSpend   float64 `json:"currentSpend"`
	ProjectedSpend float64 `json:"projectedSpend"`
	DailyTrend     float64 `json:"dailyTrend"`
}

// CostGuardrailPolicyInput represents input for creating/updating a cost guardrail policy
type CostGuardrailPolicyInput struct {
	Name              string              `json:"name" validate:"required,min=1,max=100"`
	Type              GuardrailPolicyType `json:"type" validate:"required"`
	Action            GuardrailAction     `json:"action" validate:"required"`
	Enabled           *bool               `json:"enabled,omitempty"`
	BudgetLimit       float64             `json:"budgetLimit" validate:"required"`
	BudgetPeriod      string              `json:"budgetPeriod" validate:"required"`
	ThresholdPercent  float64             `json:"thresholdPercent,omitempty"`
	ModelDowngradeMap map[string]string   `json:"modelDowngradeMap,omitempty"`
	NotifyChannels    []uuid.UUID         `json:"notifyChannels,omitempty"`
	CooldownMinutes   *int                `json:"cooldownMinutes,omitempty"`
}

// CostGuardrailFilter represents filter options for querying cost guardrail policies
type CostGuardrailFilter struct {
	ProjectID uuid.UUID
	Type      *GuardrailPolicyType
	Action    *GuardrailAction
	Status    *GuardrailStatus
	Enabled   *bool
}

// CostGuardrailPolicyList represents a paginated list of cost guardrail policies
type CostGuardrailPolicyList struct {
	Policies   []CostGuardrailPolicy `json:"policies"`
	TotalCount int64                 `json:"totalCount"`
	HasMore    bool                  `json:"hasMore"`
}

// CostGuardrailViolationList represents a paginated list of cost guardrail violations
type CostGuardrailViolationList struct {
	Violations []CostGuardrailViolation `json:"violations"`
	TotalCount int64                    `json:"totalCount"`
	HasMore    bool                     `json:"hasMore"`
}

// SpenderInfo represents spend information for a user
type SpenderInfo struct {
	UserID uuid.UUID `json:"userId"`
	Name   string    `json:"name"`
	Spend  float64   `json:"spend"`
}

// CostGuardrailDashboard provides a complete dashboard view for cost guardrails
type CostGuardrailDashboard struct {
	ActivePolicies   int                      `json:"activePolicies"`
	RecentViolations []CostGuardrailViolation `json:"recentViolations"`
	Forecast         GuardrailCostForecast    `json:"forecast"`
	TopSpenders      []SpenderInfo            `json:"topSpenders"`
}
