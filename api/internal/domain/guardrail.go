package domain

import (
	"time"

	"github.com/google/uuid"
)

// GuardRuleType represents the type of guard rule
type GuardRuleType string

const (
	GuardRuleTypeCostLimit       GuardRuleType = "cost_limit"
	GuardRuleTypeLatencyLimit    GuardRuleType = "latency_limit"
	GuardRuleTypeFileRestriction GuardRuleType = "file_restriction"
	GuardRuleTypePatternBlock    GuardRuleType = "pattern_block"
	GuardRuleTypeCustom          GuardRuleType = "custom"
)

// IsValid checks if the guard rule type is valid
func (t GuardRuleType) IsValid() bool {
	switch t {
	case GuardRuleTypeCostLimit, GuardRuleTypeLatencyLimit, GuardRuleTypeFileRestriction, GuardRuleTypePatternBlock, GuardRuleTypeCustom:
		return true
	}
	return false
}

// GuardAction represents the action to take when a rule is violated
type GuardAction string

const (
	GuardActionBlock GuardAction = "block"
	GuardActionAlert GuardAction = "alert"
	GuardActionLog   GuardAction = "log"
)

// IsValid checks if the guard action is valid
func (a GuardAction) IsValid() bool {
	switch a {
	case GuardActionBlock, GuardActionAlert, GuardActionLog:
		return true
	}
	return false
}

// GuardViolationSeverity represents the severity of a guard violation
type GuardViolationSeverity string

const (
	GuardViolationSeverityCritical GuardViolationSeverity = "critical"
	GuardViolationSeverityWarning  GuardViolationSeverity = "warning"
	GuardViolationSeverityInfo     GuardViolationSeverity = "info"
)

// IsValid checks if the guard violation severity is valid
func (s GuardViolationSeverity) IsValid() bool {
	switch s {
	case GuardViolationSeverityCritical, GuardViolationSeverityWarning, GuardViolationSeverityInfo:
		return true
	}
	return false
}

// GuardRule defines a safety rule enforced during trace ingestion
type GuardRule struct {
	ID          uuid.UUID     `json:"id"`
	ProjectID   uuid.UUID     `json:"projectId"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Type        GuardRuleType `json:"type"`
	Config      GuardRuleConfig `json:"config"`
	Action      GuardAction   `json:"action"`
	Enabled     bool          `json:"enabled"`
	CreatedAt   time.Time     `json:"createdAt"`
}

// GuardRuleConfig contains type-specific configuration for a guard rule
type GuardRuleConfig struct {
	MaxCostPerTrace  *float64 `json:"maxCostPerTrace,omitempty"`
	MaxLatencyMs     *int64   `json:"maxLatencyMs,omitempty"`
	RestrictedPaths  []string `json:"restrictedPaths,omitempty"`
	BlockedPatterns  []string `json:"blockedPatterns,omitempty"`
	CustomExpression string   `json:"customExpression,omitempty"`
}

// GuardViolation represents a detected violation of a guard rule
type GuardViolation struct {
	ID        uuid.UUID              `json:"id"`
	ProjectID uuid.UUID              `json:"projectId"`
	RuleID    uuid.UUID              `json:"ruleId"`
	TraceID   string                 `json:"traceId"`
	Severity  GuardViolationSeverity `json:"severity"`
	Details   string                 `json:"details"`
	Action    GuardAction            `json:"action"`
	CreatedAt time.Time              `json:"createdAt"`
}

// GuardEvalResult represents the result of evaluating guard rules against a trace
type GuardEvalResult struct {
	Passed     bool             `json:"passed"`
	Violations []GuardViolation `json:"violations"`
}

// GuardRuleInput represents input for creating or updating a guard rule
type GuardRuleInput struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description,omitempty"`
	Type        GuardRuleType   `json:"type" validate:"required"`
	Config      GuardRuleConfig `json:"config"`
	Action      GuardAction     `json:"action" validate:"required"`
	Enabled     *bool           `json:"enabled,omitempty"`
}

// GuardViolationFilter represents filter options for querying violations
type GuardViolationFilter struct {
	ProjectID uuid.UUID
	RuleID    *uuid.UUID
	Severity  *GuardViolationSeverity
	TraceID   *string
	StartTime *time.Time
	EndTime   *time.Time
}

// GuardViolationStats represents aggregated violation statistics
type GuardViolationStats struct {
	ProjectID      uuid.UUID                     `json:"projectId"`
	TotalCount     int64                         `json:"totalCount"`
	BySeverity     map[GuardViolationSeverity]int `json:"bySeverity"`
	ByRule         map[string]int                `json:"byRule"`
	ByAction       map[GuardAction]int           `json:"byAction"`
}
