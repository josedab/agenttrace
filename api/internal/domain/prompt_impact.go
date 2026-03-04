package domain

import (
	"time"

	"github.com/google/uuid"
)

// PromptImpactStatus represents the status of an impact analysis
type PromptImpactStatus string

const (
	PromptImpactPending   PromptImpactStatus = "pending"
	PromptImpactRunning   PromptImpactStatus = "running"
	PromptImpactCompleted PromptImpactStatus = "completed"
	PromptImpactFailed    PromptImpactStatus = "failed"
)

// PromptVersionImpactAnalysis represents an impact analysis comparing prompt versions
type PromptVersionImpactAnalysis struct {
	ID             uuid.UUID          `json:"id"`
	ProjectID      uuid.UUID          `json:"projectId"`
	PromptName     string             `json:"promptName"`
	VersionBefore  string             `json:"versionBefore"`
	VersionAfter   string             `json:"versionAfter"`
	Status         PromptImpactStatus `json:"status"`
	Dimensions     ImpactDimensions   `json:"dimensions"`
	StatTests      StatisticalTests   `json:"statTests"`
	Recommendation string             `json:"recommendation"` // "keep", "revert", "monitor"
	SampleSize     int                `json:"sampleSize"`
	CreatedBy      uuid.UUID          `json:"createdBy"`
	CreatedAt      time.Time          `json:"createdAt"`
	CompletedAt    *time.Time         `json:"completedAt,omitempty"`
}

// ImpactDimensions represents the comparison across multiple dimensions
type ImpactDimensions struct {
	Cost     DimensionComparison `json:"cost"`
	Latency  DimensionComparison `json:"latency"`
	Quality  DimensionComparison `json:"quality"`
	ErrorRate DimensionComparison `json:"errorRate"`
	Satisfaction DimensionComparison `json:"satisfaction,omitempty"`
}

// DimensionComparison represents a before/after comparison for a single dimension
type DimensionComparison struct {
	Before      float64 `json:"before"`
	After       float64 `json:"after"`
	Change      float64 `json:"change"`      // absolute change
	ChangePercent float64 `json:"changePercent"` // percentage change
	Direction   string  `json:"direction"`   // "improved", "degraded", "unchanged"
	Significant bool    `json:"significant"` // statistically significant
}

// StatisticalTests represents the results of statistical significance tests
type StatisticalTests struct {
	MannWhitneyU MannWhitneyResult `json:"mannWhitneyU"`
	ChiSquared   ChiSquaredResult  `json:"chiSquared"`
	ConfidenceLevel float64        `json:"confidenceLevel"` // e.g., 0.95
}

// MannWhitneyResult represents Mann-Whitney U test results
type MannWhitneyResult struct {
	UStatistic float64 `json:"uStatistic"`
	PValue     float64 `json:"pValue"`
	Significant bool   `json:"significant"`
}

// ChiSquaredResult represents chi-squared test results
type ChiSquaredResult struct {
	ChiSquared float64 `json:"chiSquared"`
	PValue     float64 `json:"pValue"`
	DegreesOfFreedom int `json:"degreesOfFreedom"`
	Significant bool   `json:"significant"`
}

// PromptImpactInput represents input for creating an impact analysis
type PromptImpactInput struct {
	PromptName    string `json:"promptName" validate:"required"`
	VersionBefore string `json:"versionBefore" validate:"required"`
	VersionAfter  string `json:"versionAfter" validate:"required"`
	SampleSize    int    `json:"sampleSize,omitempty"`
}

// PromptCompareInput represents input for comparing two prompt versions
type PromptCompareInput struct {
	PromptName    string `json:"promptName" validate:"required"`
	VersionA      string `json:"versionA" validate:"required"`
	VersionB      string `json:"versionB" validate:"required"`
	Dimensions    []string `json:"dimensions,omitempty"` // filter specific dimensions
}

// PromptImpactReport represents a detailed impact report
type PromptImpactReport struct {
	Analysis       PromptVersionImpactAnalysis `json:"analysis"`
	DetailedCost   TimeSeriesData       `json:"detailedCost,omitempty"`
	DetailedLatency TimeSeriesData      `json:"detailedLatency,omitempty"`
	DetailedQuality TimeSeriesData      `json:"detailedQuality,omitempty"`
	Warnings       []string             `json:"warnings,omitempty"`
}

// TimeSeriesData represents time-series data for a dimension
type TimeSeriesData struct {
	Labels []string  `json:"labels"`
	Before []float64 `json:"before"`
	After  []float64 `json:"after"`
}
