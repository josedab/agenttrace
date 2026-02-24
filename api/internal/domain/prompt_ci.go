package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaselineStatus represents the status of a prompt baseline
type BaselineStatus string

const (
	BaselineStatusActive     BaselineStatus = "active"
	BaselineStatusSuperseded BaselineStatus = "superseded"
	BaselineStatusArchived   BaselineStatus = "archived"
)

// RegressionSeverity represents the severity of a prompt regression
type RegressionSeverity string

const (
	RegressionSeverityNone     RegressionSeverity = "none"
	RegressionSeverityMinor    RegressionSeverity = "minor"
	RegressionSeverityMajor    RegressionSeverity = "major"
	RegressionSeverityCritical RegressionSeverity = "critical"
)

// PromptBaseline represents a baseline snapshot of prompt performance
type PromptBaseline struct {
	ID            uuid.UUID          `json:"id"`
	ProjectID     uuid.UUID          `json:"projectId"`
	DatasetID     uuid.UUID          `json:"datasetId"`
	PromptID      uuid.UUID          `json:"promptId"`
	PromptVersion int                `json:"promptVersion"`
	Name          string             `json:"name"`
	Branch        string             `json:"branch"`
	Scores        map[string]float64 `json:"scores"`
	SampleSize    int                `json:"sampleSize"`
	CreatedAt     time.Time          `json:"createdAt"`
	CreatedBy     uuid.UUID          `json:"createdBy"`
}

// PromptCIRun represents a CI run that evaluates prompt regression
type PromptCIRun struct {
	ID              uuid.UUID          `json:"id"`
	ProjectID       uuid.UUID          `json:"projectId"`
	BaselineID      uuid.UUID          `json:"baselineId"`
	Branch          string             `json:"branch"`
	CommitSHA       string             `json:"commitSha"`
	PRNumber        *int               `json:"prNumber,omitempty"`
	Status          string             `json:"status"`
	ScoreComparison []ScoreComparison  `json:"scoreComparison"`
	OverallSeverity RegressionSeverity `json:"overallSeverity"`
	Summary         string             `json:"summary"`
	StartedAt       time.Time          `json:"startedAt"`
	CompletedAt     *time.Time         `json:"completedAt,omitempty"`
}

// ScoreComparison represents a comparison between baseline and current scores
type ScoreComparison struct {
	MetricName    string             `json:"metricName"`
	BaselineValue float64            `json:"baselineValue"`
	CurrentValue  float64            `json:"currentValue"`
	Delta         float64            `json:"delta"`
	PValue        float64            `json:"pValue"`
	IsRegression  bool               `json:"isRegression"`
	Severity      RegressionSeverity `json:"severity"`
}

// PromptBaselineInput represents input for creating a prompt baseline
type PromptBaselineInput struct {
	Name          string             `json:"name" validate:"required,min=1,max=200"`
	DatasetID     uuid.UUID          `json:"datasetId" validate:"required"`
	PromptID      uuid.UUID          `json:"promptId" validate:"required"`
	PromptVersion int                `json:"promptVersion" validate:"required"`
	Branch        string             `json:"branch" validate:"required"`
	Scores        map[string]float64 `json:"scores"`
	SampleSize    int                `json:"sampleSize,omitempty"`
}

// PromptCIRunInput represents input for creating a prompt CI run
type PromptCIRunInput struct {
	BaselineID uuid.UUID `json:"baselineId" validate:"required"`
	Branch     string    `json:"branch" validate:"required"`
	CommitSHA  string    `json:"commitSha" validate:"required"`
	PRNumber   *int      `json:"prNumber,omitempty"`
}

// PromptBaselineList represents a paginated list of prompt baselines
type PromptBaselineList struct {
	Baselines  []PromptBaseline `json:"baselines"`
	TotalCount int64            `json:"totalCount"`
	HasMore    bool             `json:"hasMore"`
}

// PromptCIRunList represents a paginated list of prompt CI runs
type PromptCIRunList struct {
	Runs       []PromptCIRun `json:"runs"`
	TotalCount int64         `json:"totalCount"`
	HasMore    bool          `json:"hasMore"`
}
