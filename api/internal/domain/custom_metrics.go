package domain

import (
	"time"

	"github.com/google/uuid"
)

// MetricAggregation represents how a metric is aggregated
type MetricAggregation string

const (
	MetricAggAvg   MetricAggregation = "avg"
	MetricAggSum   MetricAggregation = "sum"
	MetricAggCount MetricAggregation = "count"
	MetricAggP50   MetricAggregation = "p50"
	MetricAggP95   MetricAggregation = "p95"
	MetricAggP99   MetricAggregation = "p99"
)

// MetricAlertCondition represents the condition for a metric alert
type MetricAlertCondition string

const (
	MetricAlertConditionGT MetricAlertCondition = "gt"
	MetricAlertConditionLT MetricAlertCondition = "lt"
	MetricAlertConditionEQ MetricAlertCondition = "eq"
)

// ChartType represents the type of chart for a dashboard widget
type ChartType string

const (
	ChartTypeLine   ChartType = "line"
	ChartTypeBar    ChartType = "bar"
	ChartTypeGauge  ChartType = "gauge"
	ChartTypeNumber ChartType = "number"
)

// CustomMetric represents a user-defined metric
type CustomMetric struct {
	ID              uuid.UUID         `json:"id"`
	ProjectID       uuid.UUID         `json:"projectId"`
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	Query           string            `json:"query"`
	Unit            string            `json:"unit,omitempty"`
	Aggregation     MetricAggregation `json:"aggregation"`
	RefreshInterval int               `json:"refreshInterval"`
	Enabled         bool              `json:"enabled"`
	CreatedAt       time.Time         `json:"createdAt"`
}

// CustomMetricValue represents a single data point for a custom metric
type CustomMetricValue struct {
	MetricID  uuid.UUID `json:"metricId"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// DashboardWidget represents a widget on a metric dashboard
type DashboardWidget struct {
	ID       uuid.UUID `json:"id"`
	MetricID uuid.UUID `json:"metricId"`
	ChartType ChartType `json:"chartType"`
	Position int       `json:"position"`
	Size     int       `json:"size"`
}

// MetricDashboard represents a collection of metric widgets
type MetricDashboard struct {
	ID        uuid.UUID         `json:"id"`
	ProjectID uuid.UUID         `json:"projectId"`
	Name      string            `json:"name"`
	Widgets   []DashboardWidget `json:"widgets"`
	CreatedAt time.Time         `json:"createdAt"`
}

// MetricAlert represents an alert rule for a custom metric
type MetricAlert struct {
	ID        uuid.UUID      `json:"id"`
	MetricID  uuid.UUID      `json:"metricId"`
	Condition AlertCondition `json:"condition"`
	Threshold float64        `json:"threshold"`
	Channel   string         `json:"channel"`
	Enabled   bool           `json:"enabled"`
}

// CustomMetricInput is the input for creating a custom metric
type CustomMetricInput struct {
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	Query           string            `json:"query"`
	Unit            string            `json:"unit,omitempty"`
	Aggregation     MetricAggregation `json:"aggregation"`
	RefreshInterval int               `json:"refreshInterval"`
}

// MetricAlertInput is the input for creating a metric alert
type MetricAlertInput struct {
	MetricID  uuid.UUID      `json:"metricId"`
	Condition AlertCondition `json:"condition"`
	Threshold float64        `json:"threshold"`
	Channel   string         `json:"channel"`
}
