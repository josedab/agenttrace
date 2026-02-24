package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostAlertSeverity represents the severity level of a cost alert
type CostAlertSeverity string

const (
	CostAlertSeverityInfo      CostAlertSeverity = "info"
	CostAlertSeverityWarning   CostAlertSeverity = "warning"
	CostAlertSeverityCritical  CostAlertSeverity = "critical"
	CostAlertSeverityEmergency CostAlertSeverity = "emergency"
)

// IsValid checks if the cost alert severity is valid
func (s CostAlertSeverity) IsValid() bool {
	switch s {
	case CostAlertSeverityInfo, CostAlertSeverityWarning, CostAlertSeverityCritical, CostAlertSeverityEmergency:
		return true
	}
	return false
}

// CostAlertAction represents an action to take when a cost alert is triggered
type CostAlertAction string

const (
	CostAlertActionNotify        CostAlertAction = "notify"
	CostAlertActionThrottle      CostAlertAction = "throttle"
	CostAlertActionPauseAgent    CostAlertAction = "pause_agent"
	CostAlertActionFallbackModel CostAlertAction = "fallback_model"
)

// IsValid checks if the cost alert action is valid
func (a CostAlertAction) IsValid() bool {
	switch a {
	case CostAlertActionNotify, CostAlertActionThrottle, CostAlertActionPauseAgent, CostAlertActionFallbackModel:
		return true
	}
	return false
}

// CircuitBreakerState represents the state of a cost circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerStateClosed   CircuitBreakerState = "closed"
	CircuitBreakerStateHalfOpen CircuitBreakerState = "half_open"
	CircuitBreakerStateOpen     CircuitBreakerState = "open"
)

// IsValid checks if the circuit breaker state is valid
func (s CircuitBreakerState) IsValid() bool {
	switch s {
	case CircuitBreakerStateClosed, CircuitBreakerStateHalfOpen, CircuitBreakerStateOpen:
		return true
	}
	return false
}

// CostAlert represents a real-time cost anomaly alert
type CostAlert struct {
	ID              uuid.UUID         `json:"id"`
	ProjectID       uuid.UUID         `json:"projectId"`
	Severity        CostAlertSeverity `json:"severity"`
	Action          CostAlertAction   `json:"action"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	CurrentCost     float64           `json:"currentCost"`
	ThresholdCost   float64           `json:"thresholdCost"`
	AffectedTraceID *uuid.UUID        `json:"affectedTraceId,omitempty"`
	AffectedModel   string            `json:"affectedModel,omitempty"`
	Channels        []string          `json:"channels,omitempty"`
	SentAt          *time.Time        `json:"sentAt,omitempty"`
	AcknowledgedAt  *time.Time        `json:"acknowledgedAt,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
}

// CircuitBreakerConfig represents the configuration for a cost circuit breaker
type CircuitBreakerConfig struct {
	ID                 uuid.UUID           `json:"id"`
	ProjectID          uuid.UUID           `json:"projectId"`
	Enabled            bool                `json:"enabled"`
	State              CircuitBreakerState `json:"state"`
	MaxCostPerMinute   float64             `json:"maxCostPerMinute"`
	MaxCostPerHour     float64             `json:"maxCostPerHour"`
	FallbackModelChain []string            `json:"fallbackModelChain,omitempty"`
	CooldownSeconds    int                 `json:"cooldownSeconds"`
	LastTrippedAt      *time.Time          `json:"lastTrippedAt,omitempty"`
	ResetAfterSeconds  int                 `json:"resetAfterSeconds"`
}

// CostAlertRule represents a configurable cost alert rule
type CostAlertRule struct {
	ID        uuid.UUID           `json:"id"`
	ProjectID uuid.UUID           `json:"projectId"`
	Name      string              `json:"name"`
	Enabled   bool                `json:"enabled"`
	Severity  CostAlertSeverity   `json:"severity"`
	Actions   []CostAlertAction   `json:"actions"`
	Condition CostAlertCondition  `json:"condition"`
	Channels  []uuid.UUID         `json:"channels,omitempty"`
	Cooldown  int                 `json:"cooldown"`
	CreatedAt time.Time           `json:"createdAt"`
}

// CostAlertCondition represents the condition that triggers a cost alert
type CostAlertCondition struct {
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
	Window    string  `json:"window"`
}
