package domain

import (
	"time"
)

// SLO represents a Service Level Objective for an agent
type SLO struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	AgentName string    `json:"agentName"`
	Name      string    `json:"name"`
	Metric    string    `json:"metric"` // latency_p99, success_rate, cost_per_trace, uptime
	Target    float64   `json:"target"`
	Window    string    `json:"window"` // 1h, 24h, 7d, 30d
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}

// SLOStatus represents the current compliance status of an SLO
type SLOStatus struct {
	SLOID                string    `json:"sloId"`
	CurrentValue         float64   `json:"currentValue"`
	Target               float64   `json:"target"`
	Compliant            bool      `json:"compliant"`
	ErrorBudgetRemaining float64   `json:"errorBudgetRemaining"`
	BurnRate             float64   `json:"burnRate"`
	ViolationCount       int       `json:"violationCount"`
	LastChecked          time.Time `json:"lastChecked"`
}

// SLOReport represents an overall SLO compliance report for a project
type SLOReport struct {
	ProjectID         string      `json:"projectId"`
	SLOs              []SLOStatus `json:"slos"`
	OverallCompliance float64     `json:"overallCompliance"`
	AtRisk            []string    `json:"atRisk"`
}

// SLOInput represents input for creating an SLO
type SLOInput struct {
	AgentName string  `json:"agentName" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Metric    string  `json:"metric" validate:"required"`
	Target    float64 `json:"target" validate:"required"`
	Window    string  `json:"window" validate:"required"`
}

// SLOHistoryPoint represents a point in SLO compliance history
type SLOHistoryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Compliant bool      `json:"compliant"`
}
