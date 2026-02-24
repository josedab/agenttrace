package domain

import (
	"time"

	"github.com/google/uuid"
)

// AgentType represents the type of AI coding agent
type AgentType string

const (
	AgentTypeClaudeCode AgentType = "claude_code"
	AgentTypeCopilot    AgentType = "copilot"
	AgentTypeCursor     AgentType = "cursor"
	AgentTypeAider      AgentType = "aider"
	AgentTypeCustom     AgentType = "custom"
)

// ComparisonMetricType represents the type of metric used in agent comparisons
type ComparisonMetricType string

const (
	ComparisonMetricCost            ComparisonMetricType = "cost"
	ComparisonMetricQuality         ComparisonMetricType = "quality"
	ComparisonMetricSpeed           ComparisonMetricType = "speed"
	ComparisonMetricTokenEfficiency ComparisonMetricType = "token_efficiency"
	ComparisonMetricErrorRate       ComparisonMetricType = "error_rate"
)

// AgentProfile represents a configured agent profile for comparison tracking
type AgentProfile struct {
	ID             uuid.UUID              `json:"id"`
	ProjectID      uuid.UUID              `json:"projectId"`
	Name           string                 `json:"name"`
	AgentType      AgentType              `json:"agentType"`
	ModelName      string                 `json:"modelName"`
	Description    string                 `json:"description,omitempty"`
	Config         map[string]any         `json:"config,omitempty"`
	AverageMetrics *AgentMetricsSummary   `json:"averageMetrics,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// AgentMetricsSummary represents aggregated metrics for an agent
type AgentMetricsSummary struct {
	TotalTraces      int     `json:"totalTraces"`
	AvgCostPerTrace  float64 `json:"avgCostPerTrace"`
	AvgLatencyMs     float64 `json:"avgLatencyMs"`
	AvgTokensPerTrace float64 `json:"avgTokensPerTrace"`
	AvgQualityScore  float64 `json:"avgQualityScore"`
	ErrorRate        float64 `json:"errorRate"`
	P50LatencyMs     float64 `json:"p50LatencyMs"`
	P95LatencyMs     float64 `json:"p95LatencyMs"`
	P99LatencyMs     float64 `json:"p99LatencyMs"`
}

// AgentComparisonRun represents a comparison run between multiple agents
type AgentComparisonRun struct {
	ID               uuid.UUID                       `json:"id"`
	ProjectID        uuid.UUID                       `json:"projectId"`
	Name             string                          `json:"name"`
	AgentIDs         []uuid.UUID                     `json:"agentIds"`
	DateRange        *ComparisonDateRange            `json:"dateRange,omitempty"`
	Metrics          []ComparisonMetric              `json:"metrics"`
	NormalizedScores map[string]map[string]float64   `json:"normalizedScores,omitempty"`
	WinnerByMetric   map[string]uuid.UUID            `json:"winnerByMetric,omitempty"`
	OverallWinner    *uuid.UUID                      `json:"overallWinner,omitempty"`
	CreatedAt        time.Time                       `json:"createdAt"`
}

// ComparisonDateRange represents a time range for comparison filtering
type ComparisonDateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ComparisonMetric represents a per-agent metric in a comparison
type ComparisonMetric struct {
	AgentID         uuid.UUID            `json:"agentId"`
	AgentName       string               `json:"agentName"`
	AgentType       AgentType            `json:"agentType"`
	MetricType      ComparisonMetricType `json:"metricType"`
	RawValue        float64              `json:"rawValue"`
	NormalizedValue float64              `json:"normalizedValue"`
	Rank            int                  `json:"rank"`
}

// AgentComparisonInput represents input for creating a comparison
type AgentComparisonInput struct {
	Name        string                 `json:"name" validate:"required"`
	AgentIDs    []uuid.UUID            `json:"agentIds" validate:"required,min=2"`
	FromDate    *time.Time             `json:"fromDate,omitempty"`
	ToDate      *time.Time             `json:"toDate,omitempty"`
	MetricTypes []ComparisonMetricType `json:"metricTypes,omitempty"`
}

// AgentProfileInput represents input for creating an agent profile
type AgentProfileInput struct {
	Name        string         `json:"name" validate:"required,min=1,max=200"`
	AgentType   AgentType      `json:"agentType" validate:"required"`
	ModelName   string         `json:"modelName" validate:"required"`
	Description string         `json:"description,omitempty"`
	Config      map[string]any `json:"config,omitempty"`
}

// AgentComparisonList represents a paginated list of agent comparisons
type AgentComparisonList struct {
	Comparisons []AgentComparisonRun `json:"comparisons"`
	TotalCount  int64                `json:"totalCount"`
	HasMore     bool                 `json:"hasMore"`
}

// AgentProfileList represents a paginated list of agent profiles
type AgentProfileList struct {
	Profiles   []AgentProfile `json:"profiles"`
	TotalCount int64          `json:"totalCount"`
	HasMore    bool           `json:"hasMore"`
}

// AgentTrendPoint represents a time-series data point for agent metrics
type AgentTrendPoint struct {
	Timestamp  time.Time            `json:"timestamp"`
	AgentID    uuid.UUID            `json:"agentId"`
	MetricType ComparisonMetricType `json:"metricType"`
	Value      float64              `json:"value"`
}

// AgentComparisonSummary represents a dashboard summary for agent comparisons
type AgentComparisonSummary struct {
	TotalProfiles     int               `json:"totalProfiles"`
	TotalComparisons  int               `json:"totalComparisons"`
	TopAgent          *AgentProfile     `json:"topAgent,omitempty"`
	CostLeader        *AgentProfile     `json:"costLeader,omitempty"`
	SpeedLeader       *AgentProfile     `json:"speedLeader,omitempty"`
	QualityLeader     *AgentProfile     `json:"qualityLeader,omitempty"`
	RecentComparisons []AgentComparisonRun `json:"recentComparisons"`
}
