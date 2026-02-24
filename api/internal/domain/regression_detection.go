package domain

import (
	"time"

	"github.com/google/uuid"
)

// RegressionDetectionMethod represents the algorithm used for regression detection
type RegressionDetectionMethod string

const (
	RegressionDetectionMethodStatistical   RegressionDetectionMethod = "statistical"    // z-score based
	RegressionDetectionMethodMLAnomaly     RegressionDetectionMethod = "ml_anomaly"     // isolation forest
	RegressionDetectionMethodTrendAnalysis RegressionDetectionMethod = "trend_analysis" // moving average
	RegressionDetectionMethodComparative   RegressionDetectionMethod = "comparative"    // A/B comparison
)

// RegressionDetectionSeverity represents the severity level of a detected regression
type RegressionDetectionSeverity string

const (
	RegressionDetectionSeverityCritical RegressionDetectionSeverity = "critical"
	RegressionDetectionSeverityHigh     RegressionDetectionSeverity = "high"
	RegressionDetectionSeverityMedium   RegressionDetectionSeverity = "medium"
	RegressionDetectionSeverityLow      RegressionDetectionSeverity = "low"
	RegressionDetectionSeverityInfo     RegressionDetectionSeverity = "info"
)

// RegressionDetectionConfig defines a regression detection configuration for a project
type RegressionDetectionConfig struct {
	ID               uuid.UUID                 `json:"id"`
	ProjectID        uuid.UUID                 `json:"projectId"`
	Name             string                    `json:"name"`
	Enabled          bool                      `json:"enabled"`
	Method           RegressionDetectionMethod `json:"method"`
	MonitoredMetrics []string                  `json:"monitoredMetrics"` // e.g. ["quality", "latency", "cost", "error_rate"]
	Thresholds       *RegressionThresholds     `json:"thresholds,omitempty"`
	BaselineWindow   int                       `json:"baselineWindow"`   // days of baseline data
	EvaluationWindow int                       `json:"evaluationWindow"` // days of recent data to evaluate
	MinSampleSize    int                       `json:"minSampleSize"`    // minimum traces needed
	AlertChannels    []uuid.UUID               `json:"alertChannels"`
	Schedule         string                    `json:"schedule"` // cron expression
	CreatedAt        time.Time                 `json:"createdAt"`
	UpdatedAt        time.Time                 `json:"updatedAt"`
}

// RegressionThresholds defines thresholds that trigger regression alerts
type RegressionThresholds struct {
	QualityDropPct       float64 `json:"qualityDropPct"`       // % drop that triggers alert
	LatencyIncreasePct   float64 `json:"latencyIncreasePct"`
	CostIncreasePct      float64 `json:"costIncreasePct"`
	ErrorRateIncreasePct float64 `json:"errorRateIncreasePct"`
	PValueThreshold      float64 `json:"pValueThreshold"` // statistical significance
	MinEffectSize        float64 `json:"minEffectSize"`   // minimum practical significance
}

// RegressionDetectionResult represents the outcome of a regression detection run
type RegressionDetectionResult struct {
	ID                  uuid.UUID                 `json:"id"`
	ProjectID           uuid.UUID                 `json:"projectId"`
	ConfigID            uuid.UUID                 `json:"configId"`
	DetectedAt          time.Time                 `json:"detectedAt"`
	Severity            RegressionDetectionSeverity `json:"severity"`
	Method              RegressionDetectionMethod `json:"method"`
	AffectedMetric      string                    `json:"affectedMetric"`
	BaselineValue       float64                   `json:"baselineValue"`
	CurrentValue        float64                   `json:"currentValue"`
	DeltaPct            float64                   `json:"deltaPct"`
	PValue              *float64                  `json:"pValue,omitempty"`
	EffectSize          *float64                  `json:"effectSize,omitempty"`
	IsRegression        bool                      `json:"isRegression"`
	Summary             string                    `json:"summary"`
	RootCauseHypothesis string                    `json:"rootCauseHypothesis"`
	RelatedChanges      []RelatedChange           `json:"relatedChanges"`
	Status              string                    `json:"status"` // detected, investigating, confirmed, resolved, false_positive
	AcknowledgedBy      *uuid.UUID                `json:"acknowledgedBy,omitempty"`
	AcknowledgedAt      *time.Time                `json:"acknowledgedAt,omitempty"`
	ResolvedAt          *time.Time                `json:"resolvedAt,omitempty"`
}

// RelatedChange represents a change potentially related to a detected regression
type RelatedChange struct {
	Type        string         `json:"type"` // model_update, prompt_change, config_change, sdk_version
	Description string         `json:"description"`
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RegressionDetectionInput represents input for creating a regression detection config
type RegressionDetectionInput struct {
	Name             string                    `json:"name" validate:"required,min=1,max=200"`
	Method           RegressionDetectionMethod `json:"method" validate:"required"`
	MonitoredMetrics []string                  `json:"monitoredMetrics" validate:"required,min=1"`
	Thresholds       *RegressionThresholds     `json:"thresholds,omitempty"`
	BaselineWindow   int                       `json:"baselineWindow"`
	EvaluationWindow int                       `json:"evaluationWindow"`
	MinSampleSize    int                       `json:"minSampleSize"`
	AlertChannels    []uuid.UUID               `json:"alertChannels"`
	Schedule         string                    `json:"schedule"`
}

// RegressionDetectionDashboard provides an overview of regression detection status
type RegressionDetectionDashboard struct {
	TotalConfigs         int                                `json:"totalConfigs"`
	ActiveConfigs        int                                `json:"activeConfigs"`
	TotalDetections      int                                `json:"totalDetections"`
	UnresolvedDetections int                                `json:"unresolvedDetections"`
	RecentDetections     []RegressionDetectionResult        `json:"recentDetections"`
	MetricHealth         map[string]MetricHealthStatus      `json:"metricHealth"`
}

// MetricHealthStatus represents the health status of a monitored metric
type MetricHealthStatus struct {
	Metric         string    `json:"metric"`
	Status         string    `json:"status"` // healthy, warning, critical
	CurrentValue   float64   `json:"currentValue"`
	BaselineValue  float64   `json:"baselineValue"`
	TrendDirection float64   `json:"trendDirection"`
	LastChecked    time.Time `json:"lastChecked"`
}

// RegressionDetectionConfigList is a paginated list of regression detection configs
type RegressionDetectionConfigList struct {
	Configs    []RegressionDetectionConfig `json:"configs"`
	TotalCount int64                       `json:"totalCount"`
	HasMore    bool                        `json:"hasMore"`
}

// RegressionDetectionResultList is a paginated list of regression detection results
type RegressionDetectionResultList struct {
	Results    []RegressionDetectionResult `json:"results"`
	TotalCount int64                       `json:"totalCount"`
	HasMore    bool                        `json:"hasMore"`
}
