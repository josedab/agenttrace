package domain

import (
	"time"

	"github.com/google/uuid"
)

// TrendDirection represents the direction of a metric trend
type TrendDirection string

const (
	TrendDirectionImproving TrendDirection = "improving"
	TrendDirectionStable    TrendDirection = "stable"
	TrendDirectionDegrading TrendDirection = "degrading"
)

// IsValid checks if the trend direction is valid
func (d TrendDirection) IsValid() bool {
	switch d {
	case TrendDirectionImproving, TrendDirectionStable, TrendDirectionDegrading:
		return true
	}
	return false
}

// PredictionAlertLevel represents the alert severity for a health prediction
type PredictionAlertLevel string

const (
	PredictionAlertLevelNone     PredictionAlertLevel = "none"
	PredictionAlertLevelWarning  PredictionAlertLevel = "warning"
	PredictionAlertLevelCritical PredictionAlertLevel = "critical"
)

// IsValid checks if the prediction alert level is valid
func (l PredictionAlertLevel) IsValid() bool {
	switch l {
	case PredictionAlertLevelNone, PredictionAlertLevelWarning, PredictionAlertLevelCritical:
		return true
	}
	return false
}

// PredictionRootCauseType represents the type of root cause for a prediction anomaly
type PredictionRootCauseType string

const (
	PredictionRootCausePromptChange PredictionRootCauseType = "prompt_change"
	PredictionRootCauseModelChange  PredictionRootCauseType = "model_change"
	PredictionRootCauseDataDrift    PredictionRootCauseType = "data_drift"
	PredictionRootCauseTrafficSpike PredictionRootCauseType = "traffic_spike"
)

// IsValid checks if the prediction root cause type is valid
func (t PredictionRootCauseType) IsValid() bool {
	switch t {
	case PredictionRootCausePromptChange, PredictionRootCauseModelChange, PredictionRootCauseDataDrift, PredictionRootCauseTrafficSpike:
		return true
	}
	return false
}

// OverallHealthStatus represents the overall health status of a project
type OverallHealthStatus string

const (
	OverallHealthStatusHealthy  OverallHealthStatus = "healthy"
	OverallHealthStatusWarning  OverallHealthStatus = "warning"
	OverallHealthStatusCritical OverallHealthStatus = "critical"
)

// IsValid checks if the overall health status is valid
func (s OverallHealthStatus) IsValid() bool {
	switch s {
	case OverallHealthStatusHealthy, OverallHealthStatusWarning, OverallHealthStatusCritical:
		return true
	}
	return false
}

// HealthPrediction represents a predictive health assessment for an agent metric
type HealthPrediction struct {
	ID               uuid.UUID            `json:"id"`
	ProjectID        uuid.UUID            `json:"projectId"`
	MetricName       string               `json:"metricName"` // accuracy, latency, cost, error_rate
	CurrentValue     float64              `json:"currentValue"`
	PredictedValue   float64              `json:"predictedValue"`
	TrendDirection   TrendDirection       `json:"trendDirection"`
	ConfidenceLevel  float64              `json:"confidenceLevel"`
	TimeHorizonHours int                  `json:"timeHorizonHours"`
	AlertLevel       PredictionAlertLevel `json:"alertLevel"`
	RootCause        *PredictionRootCause `json:"rootCause,omitempty"`
	CreatedAt        time.Time            `json:"createdAt"`
}

// PredictionRootCause describes the identified root cause of a prediction anomaly
type PredictionRootCause struct {
	Type            PredictionRootCauseType `json:"type"`
	Description     string                  `json:"description"`
	RelatedEntityID string                  `json:"relatedEntityId"`
	DetectedAt      time.Time               `json:"detectedAt"`
}

// TrendDataPoint represents a single data point in a time series trend
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// HealthTrend represents a metric trend with historical data and forecast
type HealthTrend struct {
	MetricName  string           `json:"metricName"`
	DataPoints  []TrendDataPoint `json:"dataPoints"`
	TrendSlope  float64          `json:"trendSlope"`
	IsAnomalous bool             `json:"isAnomalous"`
	Forecast    []TrendDataPoint `json:"forecast"`
}

// HealthDashboard provides an aggregate health view for a project
type HealthDashboard struct {
	ProjectID     uuid.UUID          `json:"projectId"`
	Predictions   []HealthPrediction `json:"predictions"`
	OverallHealth OverallHealthStatus `json:"overallHealth"`
	LastUpdated   time.Time          `json:"lastUpdated"`
}
