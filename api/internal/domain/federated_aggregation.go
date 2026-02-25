package domain

import (
	"time"

	"github.com/google/uuid"
)

// FederatedMetricType represents the type of a federated metric
type FederatedMetricType string

const (
	FederatedMetricTypeLatency    FederatedMetricType = "latency"
	FederatedMetricTypeCost       FederatedMetricType = "cost"
	FederatedMetricTypeErrorRate  FederatedMetricType = "error_rate"
	FederatedMetricTypeThroughput FederatedMetricType = "throughput"
	FederatedMetricTypeTokenUsage FederatedMetricType = "token_usage"
)

// PrivacyLevel represents the privacy level for federated data sharing
type PrivacyLevel string

const (
	PrivacyLevelFull                PrivacyLevel = "full"
	PrivacyLevelAggregatedOnly      PrivacyLevel = "aggregated_only"
	PrivacyLevelDifferentialPrivacy PrivacyLevel = "differential_privacy"
)

// FederatedInstance represents a participating instance in the federation
type FederatedInstance struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Endpoint     string       `json:"endpoint"`
	APIKey       string       `json:"apiKey"`
	PrivacyLevel PrivacyLevel `json:"privacyLevel"`
	LastSyncAt   *time.Time   `json:"lastSyncAt,omitempty"`
	Status       string       `json:"status"`
	MetricsCount int64        `json:"metricsCount"`
	CreatedAt    time.Time    `json:"createdAt"`
}

// FederatedMetric represents an aggregated metric from the federation
type FederatedMetric struct {
	InstanceID        uuid.UUID          `json:"instanceId"`
	MetricType        FederatedMetricType `json:"metricType"`
	Period            TimeWindow         `json:"period"`
	Value             float64            `json:"value"`
	P50               float64            `json:"p50"`
	P95               float64            `json:"p95"`
	P99               float64            `json:"p99"`
	StdDev            float64            `json:"stdDev"`
	SampleCount       int64              `json:"sampleCount"`
	ModelDistribution map[string]float64 `json:"modelDistribution,omitempty"`
}

// FederatedBenchmark represents a benchmark comparison against the federation
type FederatedBenchmark struct {
	MetricType        FederatedMetricType `json:"metricType"`
	YourValue         float64             `json:"yourValue"`
	IndustryP25       float64             `json:"industryP25"`
	IndustryP50       float64             `json:"industryP50"`
	IndustryP75       float64             `json:"industryP75"`
	Percentile        int                 `json:"percentile"`
	TotalParticipants int                 `json:"totalParticipants"`
	Recommendations   []string            `json:"recommendations"`
}

// FederatedAggInsight represents an insight derived from federated aggregation data
type FederatedAggInsight struct {
	ID              uuid.UUID `json:"id"`
	Type            string    `json:"type"`
	Description     string    `json:"description"`
	Confidence      float64   `json:"confidence"`
	AffectedMetrics []string  `json:"affectedMetrics"`
	Evidence        string    `json:"evidence"`
	GeneratedAt     time.Time `json:"generatedAt"`
}

// FederatedInstanceInput represents input for creating/updating a federated instance
type FederatedInstanceInput struct {
	Name         string       `json:"name" validate:"required,min=1,max=200"`
	Endpoint     string       `json:"endpoint" validate:"required"`
	APIKey       string       `json:"apiKey" validate:"required"`
	PrivacyLevel PrivacyLevel `json:"privacyLevel" validate:"required"`
}

// FederatedDashboard provides a complete dashboard view for federated aggregation
type FederatedDashboard struct {
	Instances        []FederatedInstance  `json:"instances"`
	Benchmarks       []FederatedBenchmark `json:"benchmarks"`
	Insights         []FederatedAggInsight `json:"insights"`
	ParticipantCount int                  `json:"participantCount"`
	LastAggregation  *time.Time           `json:"lastAggregation,omitempty"`
}

// DifferentialPrivacyConfig configures the differential privacy parameters
type DifferentialPrivacyConfig struct {
	Epsilon      float64 `json:"epsilon"`      // Privacy budget (lower = more private)
	Delta        float64 `json:"delta"`         // Probability of privacy breach
	Sensitivity  float64 `json:"sensitivity"`   // Max influence of single record
	NoiseType    string  `json:"noiseType"`     // laplacian, gaussian
}

// AnonymizedBenchmarkSubmission represents anonymized metrics submitted to the mesh
type AnonymizedBenchmarkSubmission struct {
	ID               uuid.UUID           `json:"id"`
	InstanceHash     string              `json:"instanceHash"` // Hashed instance ID
	Metrics          []AnonymizedMetric  `json:"metrics"`
	PrivacyConfig    DifferentialPrivacyConfig `json:"privacyConfig"`
	SubmittedAt      time.Time           `json:"submittedAt"`
}

// AnonymizedMetric represents a single metric with noise applied
type AnonymizedMetric struct {
	MetricType  FederatedMetricType `json:"metricType"`
	Value       float64             `json:"value"`       // Noise-injected value
	NoiseAdded  float64             `json:"noiseAdded"`  // Amount of noise (for transparency)
	SampleSize  int64               `json:"sampleSize"`
	Period      string              `json:"period"`      // daily, weekly, monthly
}

// IndustryBaseline represents aggregated industry baseline statistics
type IndustryBaseline struct {
	MetricType    FederatedMetricType `json:"metricType"`
	P10           float64             `json:"p10"`
	P25           float64             `json:"p25"`
	P50           float64             `json:"p50"`
	P75           float64             `json:"p75"`
	P90           float64             `json:"p90"`
	Mean          float64             `json:"mean"`
	StdDev        float64             `json:"stdDev"`
	Participants  int                 `json:"participants"`
	LastUpdated   time.Time           `json:"lastUpdated"`
}

// MeshStatus represents the health and status of the federated mesh
type MeshStatus struct {
	TotalInstances   int        `json:"totalInstances"`
	ActiveInstances  int        `json:"activeInstances"`
	TotalMetrics     int64      `json:"totalMetrics"`
	AvgPrivacyLevel  string     `json:"avgPrivacyLevel"`
	MeshHealth       string     `json:"meshHealth"` // healthy, degraded, unhealthy
	LastSync         *time.Time `json:"lastSync,omitempty"`
}
