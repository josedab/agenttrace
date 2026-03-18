package domain

import (
	"time"

	"github.com/google/uuid"
)

// PromptABTestStatus represents the status of a prompt A/B test experiment
type PromptABTestStatus string

const (
	PromptABTestStatusDraft     PromptABTestStatus = "draft"
	PromptABTestStatusRunning   PromptABTestStatus = "running"
	PromptABTestStatusPaused    PromptABTestStatus = "paused"
	PromptABTestStatusCompleted PromptABTestStatus = "completed"
	PromptABTestStatusCancelled PromptABTestStatus = "cancelled"
)

// PromptABTest represents a prompt A/B test experiment with traffic splitting
type PromptABTest struct {
	ID               uuid.UUID              `json:"id"`
	ProjectID        uuid.UUID              `json:"projectId"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	PromptID         uuid.UUID              `json:"promptId"`
	Status           PromptABTestStatus     `json:"status"`
	Variants         []PromptABTestVariant  `json:"variants"`
	TrafficSplit     PromptTrafficSplit     `json:"trafficSplit"`
	TargetMetric     string                 `json:"targetMetric"`
	SecondaryMetrics []string               `json:"secondaryMetrics,omitempty"`
	MinSampleSize    int                    `json:"minSampleSize"`
	ConfidenceLevel  float64                `json:"confidenceLevel"`
	WinnerID         *uuid.UUID             `json:"winnerId,omitempty"`
	AutoSelectWinner bool                   `json:"autoSelectWinner"`
	GradualRollout   *PromptGradualRollout  `json:"gradualRollout,omitempty"`
	StartedAt        *time.Time             `json:"startedAt,omitempty"`
	EndedAt          *time.Time             `json:"endedAt,omitempty"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
}

// PromptABTestVariant represents a variant in a prompt A/B test
type PromptABTestVariant struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	PromptVersionID uuid.UUID              `json:"promptVersionId"`
	TrafficPercent  float64                `json:"trafficPercent"`
	IsControl       bool                   `json:"isControl"`
	SampleCount     int                    `json:"sampleCount"`
	Metrics         ABTestVariantMetrics   `json:"metrics"`
}

// ABTestVariantMetrics holds aggregated metrics for a prompt A/B test variant
type ABTestVariantMetrics struct {
	AvgScore     float64 `json:"avgScore"`
	StdDeviation float64 `json:"stdDeviation"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	AvgCostUSD   float64 `json:"avgCostUsd"`
	ErrorRate    float64 `json:"errorRate"`
	P95LatencyMs float64 `json:"p95LatencyMs"`
	TotalTokens  int64   `json:"totalTokens"`
}

// PromptTrafficSplit defines how traffic is distributed
type PromptTrafficSplit struct {
	Method      string `json:"method"`
	StickyField string `json:"stickyField,omitempty"`
}

// PromptGradualRollout configures gradual winner rollout
type PromptGradualRollout struct {
	Enabled            bool    `json:"enabled"`
	InitialPercent     float64 `json:"initialPercent"`
	IncrementPercent   float64 `json:"incrementPercent"`
	IncrementIntervalH int     `json:"incrementIntervalHours"`
	CurrentPercent     float64 `json:"currentPercent"`
	AutoComplete       bool    `json:"autoComplete"`
}

// PromptABTestStatistics provides statistical analysis of test results
type PromptABTestStatistics struct {
	TestID          uuid.UUID                  `json:"testId"`
	IsSignificant   bool                       `json:"isSignificant"`
	PValue          float64                    `json:"pValue"`
	ConfidenceLevel float64                    `json:"confidenceLevel"`
	Effect          float64                    `json:"effect"`
	PowerAnalysis   float64                    `json:"powerAnalysis"`
	RequiredSamples int                        `json:"requiredSamples"`
	CurrentSamples  int                        `json:"currentSamples"`
	VariantStats    []PromptVariantStatistic   `json:"variantStats"`
	Recommendation  string                     `json:"recommendation"`
	AnalyzedAt      time.Time                  `json:"analyzedAt"`
}

// PromptVariantStatistic provides per-variant statistical analysis
type PromptVariantStatistic struct {
	VariantID       uuid.UUID `json:"variantId"`
	VariantName     string    `json:"variantName"`
	Mean            float64   `json:"mean"`
	StdDev          float64   `json:"stdDev"`
	ConfidenceLower float64   `json:"confidenceLower"`
	ConfidenceUpper float64   `json:"confidenceUpper"`
	SampleSize      int       `json:"sampleSize"`
	IsWinner        bool      `json:"isWinner"`
	Improvement     float64   `json:"improvement"`
}

// PromptVariantAssignment represents a traffic assignment for a specific request
type PromptVariantAssignment struct {
	TestID          uuid.UUID `json:"testId"`
	VariantID       uuid.UUID `json:"variantId"`
	VariantName     string    `json:"variantName"`
	PromptVersionID uuid.UUID `json:"promptVersionId"`
	AssignmentKey   string    `json:"assignmentKey"`
}

// PromptABTestInput represents input for creating a prompt A/B test
type PromptABTestInput struct {
	Name             string                       `json:"name" validate:"required,min=1,max=200"`
	Description      string                       `json:"description,omitempty"`
	PromptID         uuid.UUID                    `json:"promptId" validate:"required"`
	Variants         []PromptABTestVariantInput   `json:"variants" validate:"required,min=2"`
	TargetMetric     string                       `json:"targetMetric" validate:"required"`
	SecondaryMetrics []string                     `json:"secondaryMetrics,omitempty"`
	MinSampleSize    int                          `json:"minSampleSize,omitempty"`
	ConfidenceLevel  float64                      `json:"confidenceLevel,omitempty"`
	TrafficSplit     *PromptTrafficSplit          `json:"trafficSplit,omitempty"`
	AutoSelectWinner bool                         `json:"autoSelectWinner,omitempty"`
	GradualRollout   *PromptGradualRollout        `json:"gradualRollout,omitempty"`
}

// PromptABTestVariantInput represents input for a variant in a prompt A/B test
type PromptABTestVariantInput struct {
	Name            string    `json:"name" validate:"required"`
	PromptVersionID uuid.UUID `json:"promptVersionId" validate:"required"`
	TrafficPercent  float64   `json:"trafficPercent" validate:"required"`
	IsControl       bool      `json:"isControl,omitempty"`
}

// PromptABRecordResultInput represents input for recording a test result
type PromptABRecordResultInput struct {
	VariantID uuid.UUID              `json:"variantId" validate:"required"`
	Score     float64                `json:"score"`
	LatencyMs float64               `json:"latencyMs,omitempty"`
	CostUSD   float64               `json:"costUsd,omitempty"`
	Tokens    int                    `json:"tokens,omitempty"`
	IsError   bool                   `json:"isError,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
