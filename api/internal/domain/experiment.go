package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	ExperimentStatusDraft    ExperimentStatus = "draft"
	ExperimentStatusRunning  ExperimentStatus = "running"
	ExperimentStatusPaused   ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusArchived ExperimentStatus = "archived"
)

// Experiment represents an A/B test or experiment
type Experiment struct {
	ID          uuid.UUID        `json:"id"`
	ProjectID   uuid.UUID        `json:"projectId"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Status      ExperimentStatus `json:"status"`

	// Configuration
	Variants      []ExperimentVariant `json:"variants"`
	TargetMetric  string              `json:"targetMetric"` // e.g., "latency", "cost", "score:quality"
	TargetGoal    string              `json:"targetGoal"`   // "minimize" or "maximize"

	// Traffic allocation
	TrafficPercent float64 `json:"trafficPercent"` // 0-100, percentage of traffic included

	// Filters
	TraceNameFilter   string            `json:"traceNameFilter,omitempty"`
	UserIDFilter      []string          `json:"userIdFilter,omitempty"`
	MetadataFilters   map[string]string `json:"metadataFilters,omitempty"`

	// Duration
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty"`
	MinDuration *int       `json:"minDurationHours,omitempty"` // Minimum experiment duration
	MinSamples  *int       `json:"minSamplesPerVariant,omitempty"`

	// Results
	WinningVariant    *uuid.UUID           `json:"winningVariant,omitempty"`
	Results           *ExperimentResults   `json:"results,omitempty"`
	StatisticalPower  float64              `json:"statisticalPower,omitempty"` // 0-1

	// Audit
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   uuid.UUID `json:"createdBy"`
}

// ExperimentVariant represents a variant in an experiment
type ExperimentVariant struct {
	ID           uuid.UUID   `json:"id"`
	ExperimentID uuid.UUID   `json:"experimentId"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	Weight       float64     `json:"weight"` // 0-100, traffic allocation percentage
	IsControl    bool        `json:"isControl"`

	// Configuration applied to this variant
	Config VariantConfig `json:"config"`

	// Computed metrics
	SampleCount  int     `json:"sampleCount"`
	MetricMean   float64 `json:"metricMean,omitempty"`
	MetricStdDev float64 `json:"metricStdDev,omitempty"`
	MetricMin    float64 `json:"metricMin,omitempty"`
	MetricMax    float64 `json:"metricMax,omitempty"`
}

// VariantConfig represents the configuration for a variant
type VariantConfig struct {
	// Prompt configuration
	PromptName    string `json:"promptName,omitempty"`
	PromptVersion *int   `json:"promptVersion,omitempty"`

	// Model configuration
	Model       string             `json:"model,omitempty"`
	ModelParams map[string]any     `json:"modelParams,omitempty"` // temperature, max_tokens, etc.

	// Custom configuration
	CustomConfig map[string]any `json:"customConfig,omitempty"`
}

// ExperimentResults contains the statistical analysis results
type ExperimentResults struct {
	AnalyzedAt       time.Time                  `json:"analyzedAt"`
	TotalSamples     int                        `json:"totalSamples"`
	VariantResults   []VariantResult            `json:"variantResults"`
	Comparisons      []VariantComparison        `json:"comparisons"`
	RecommendedAction string                    `json:"recommendedAction"`
	Confidence       float64                    `json:"confidence"` // 0-1
}

// VariantResult contains results for a single variant
type VariantResult struct {
	VariantID    uuid.UUID `json:"variantId"`
	VariantName  string    `json:"variantName"`
	SampleCount  int       `json:"sampleCount"`
	Mean         float64   `json:"mean"`
	StdDev       float64   `json:"stdDev"`
	Median       float64   `json:"median"`
	P95          float64   `json:"p95"`
	P99          float64   `json:"p99"`
	Min          float64   `json:"min"`
	Max          float64   `json:"max"`
	ErrorRate    float64   `json:"errorRate"`
}

// VariantComparison contains statistical comparison between two variants
type VariantComparison struct {
	VariantA        uuid.UUID `json:"variantA"`
	VariantB        uuid.UUID `json:"variantB"`
	VariantAName    string    `json:"variantAName"`
	VariantBName    string    `json:"variantBName"`
	MeanDifference  float64   `json:"meanDifference"`
	PercentChange   float64   `json:"percentChange"`
	PValue          float64   `json:"pValue"`
	IsSignificant   bool      `json:"isSignificant"` // p < 0.05
	ConfidenceLevel float64   `json:"confidenceLevel"`
	Winner          *string   `json:"winner,omitempty"` // "A", "B", or nil for no winner
}

// ExperimentInput represents input for creating an experiment
type ExperimentInput struct {
	Name           string              `json:"name" validate:"required,min=1,max=100"`
	Description    string              `json:"description,omitempty"`
	Variants       []VariantInput      `json:"variants" validate:"required,min=2,max=10"`
	TargetMetric   string              `json:"targetMetric" validate:"required"`
	TargetGoal     string              `json:"targetGoal" validate:"required,oneof=minimize maximize"`
	TrafficPercent float64             `json:"trafficPercent" validate:"min=0,max=100"`
	TraceNameFilter string             `json:"traceNameFilter,omitempty"`
	UserIDFilter    []string           `json:"userIdFilter,omitempty"`
	MetadataFilters map[string]string  `json:"metadataFilters,omitempty"`
	MinDuration     *int               `json:"minDurationHours,omitempty"`
	MinSamples      *int               `json:"minSamplesPerVariant,omitempty"`
}

// VariantInput represents input for creating a variant
type VariantInput struct {
	Name        string        `json:"name" validate:"required,min=1,max=100"`
	Description string        `json:"description,omitempty"`
	Weight      float64       `json:"weight" validate:"min=0,max=100"`
	IsControl   bool          `json:"isControl"`
	Config      VariantConfig `json:"config"`
}

// ExperimentUpdateInput represents input for updating an experiment
type ExperimentUpdateInput struct {
	Name           *string            `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description    *string            `json:"description,omitempty"`
	Status         *ExperimentStatus  `json:"status,omitempty"`
	TrafficPercent *float64           `json:"trafficPercent,omitempty" validate:"omitempty,min=0,max=100"`
}

// ExperimentFilter represents filter options for querying experiments
type ExperimentFilter struct {
	ProjectID uuid.UUID
	Status    *ExperimentStatus
	Search    string
}

// ExperimentList represents a paginated list of experiments
type ExperimentList struct {
	Experiments []Experiment `json:"experiments"`
	TotalCount  int64        `json:"totalCount"`
	HasMore     bool         `json:"hasMore"`
}

// ExperimentAssignment represents which variant a trace is assigned to
type ExperimentAssignment struct {
	ExperimentID uuid.UUID `json:"experimentId"`
	VariantID    uuid.UUID `json:"variantId"`
	TraceID      uuid.UUID `json:"traceId"`
	AssignedAt   time.Time `json:"assignedAt"`

	// Variant configuration at time of assignment (for reproducibility)
	VariantConfig VariantConfig `json:"variantConfig"`
}

// ExperimentMetric represents a metric value for analysis
type ExperimentMetric struct {
	ExperimentID uuid.UUID `json:"experimentId"`
	VariantID    uuid.UUID `json:"variantId"`
	TraceID      uuid.UUID `json:"traceId"`
	MetricName   string    `json:"metricName"`
	MetricValue  float64   `json:"metricValue"`
	RecordedAt   time.Time `json:"recordedAt"`
}
