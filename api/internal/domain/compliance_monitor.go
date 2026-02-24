package domain

import (
	"time"

	"github.com/google/uuid"
)

// CompliancePolicy represents an automated compliance policy
type CompliancePolicy struct {
	ID        uuid.UUID        `json:"id"`
	ProjectID uuid.UUID        `json:"projectId"`
	Name      string           `json:"name"`
	Framework string           `json:"framework"` // eu_ai_act, soc2, iso27001, custom
	Rules     []ComplianceRule `json:"rules"`
	Enabled   bool             `json:"enabled"`
	CreatedAt time.Time        `json:"createdAt"`
}

// ComplianceRule represents a single compliance rule within a policy
type ComplianceRule struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Check       string  `json:"check"` // guardrails_enabled, pii_redaction, audit_logging, cost_limits, access_control
	Required    bool    `json:"required"`
	Weight      float64 `json:"weight"`
}

// ComplianceScore represents the compliance evaluation score for a project
type ComplianceScore struct {
	ProjectID    uuid.UUID    `json:"projectId"`
	Framework    string       `json:"framework"`
	OverallScore float64      `json:"overallScore"` // 0-100
	RuleResults  []RuleResult `json:"ruleResults"`
	LastChecked  time.Time    `json:"lastChecked"`
	Trend        string       `json:"trend"`
}

// RuleResult represents the result of evaluating a single compliance rule
type RuleResult struct {
	RuleID    string    `json:"ruleId"`
	RuleName  string    `json:"ruleName"`
	Compliant bool      `json:"compliant"`
	Evidence  string    `json:"evidence"`
	CheckedAt time.Time `json:"checkedAt"`
}

// CompliancePolicyInput represents input for creating a compliance policy
type CompliancePolicyInput struct {
	Name      string           `json:"name"`
	Framework string           `json:"framework"`
	Rules     []ComplianceRule `json:"rules"`
	Enabled   bool             `json:"enabled"`
}

// ContinuousMonitorConfig represents configuration for continuous compliance monitoring
type ContinuousMonitorConfig struct {
	ProjectID     uuid.UUID `json:"projectId"`
	Enabled       bool      `json:"enabled"`
	CheckInterval int       `json:"checkInterval"` // minutes
	AlertOnDrop   bool      `json:"alertOnDrop"`
	MinScore      float64   `json:"minScore"`
}
