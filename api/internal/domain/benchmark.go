package domain

import (
	"time"

	"github.com/google/uuid"
)

// BenchmarkCategory represents the category of a benchmark
type BenchmarkCategory string

const (
	BenchmarkCategoryCodeGeneration BenchmarkCategory = "code_generation"
	BenchmarkCategoryBugFixing      BenchmarkCategory = "bug_fixing"
	BenchmarkCategoryQA             BenchmarkCategory = "qa"
	BenchmarkCategoryReasoning      BenchmarkCategory = "reasoning"
	BenchmarkCategoryCustom         BenchmarkCategory = "custom"
)

// IsValid checks if the benchmark category is valid
func (c BenchmarkCategory) IsValid() bool {
	switch c {
	case BenchmarkCategoryCodeGeneration, BenchmarkCategoryBugFixing, BenchmarkCategoryQA, BenchmarkCategoryReasoning, BenchmarkCategoryCustom:
		return true
	}
	return false
}

// Benchmark represents a curated benchmark for agent evaluation
type Benchmark struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Category     BenchmarkCategory `json:"category"`
	DatasetID    uuid.UUID         `json:"datasetId"`
	EvaluatorIDs []uuid.UUID       `json:"evaluatorIds"`
	Metrics      []BenchmarkMetric `json:"metrics"`
	IsPublic     bool              `json:"isPublic"`
	CreatedAt    time.Time         `json:"createdAt"`
}

// BenchmarkMetric defines a metric used in a benchmark
type BenchmarkMetric struct {
	Name          string  `json:"name"`
	Weight        float64 `json:"weight"`
	HigherIsBetter bool   `json:"higherIsBetter"`
}

// BenchmarkSubmission represents an agent's submission to a benchmark
type BenchmarkSubmission struct {
	ID           uuid.UUID          `json:"id"`
	BenchmarkID  uuid.UUID          `json:"benchmarkId"`
	ProjectID    uuid.UUID          `json:"projectId"`
	AgentName    string             `json:"agentName"`
	AgentVersion string             `json:"agentVersion"`
	Scores       map[string]float64 `json:"scores"`
	OverallScore float64            `json:"overallScore"`
	Rank         int                `json:"rank"`
	Metadata     string             `json:"metadata,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
}

// BenchmarkLeaderboard represents the leaderboard for a benchmark
type BenchmarkLeaderboard struct {
	BenchmarkID uuid.UUID             `json:"benchmarkId"`
	Submissions []BenchmarkSubmission  `json:"submissions"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

// SubmitBenchmarkInput represents input for submitting to a benchmark
type SubmitBenchmarkInput struct {
	BenchmarkID  uuid.UUID `json:"benchmarkId" validate:"required"`
	AgentName    string    `json:"agentName"`
	AgentVersion string    `json:"agentVersion"`
}

// BenchmarkComparison compares two agent submissions
type BenchmarkComparison struct {
	BenchmarkID  uuid.UUID              `json:"benchmarkId"`
	SubmissionA  BenchmarkSubmission    `json:"submissionA"`
	SubmissionB  BenchmarkSubmission    `json:"submissionB"`
	MetricDeltas map[string]MetricDelta `json:"metricDeltas"`
	Winner       string                 `json:"winner"` // a, b, tie
	Summary      string                 `json:"summary"`
}

// MetricDelta represents the difference in a metric between two submissions
type MetricDelta struct {
	MetricName string  `json:"metricName"`
	ValueA     float64 `json:"valueA"`
	ValueB     float64 `json:"valueB"`
	Delta      float64 `json:"delta"`
	DeltaPct   float64 `json:"deltaPercent"`
	IsBetter   string  `json:"isBetter"` // a, b, tie
}

// BenchmarkStats provides aggregate statistics
type BenchmarkStats struct {
	BenchmarkID      uuid.UUID              `json:"benchmarkId"`
	TotalSubmissions int                    `json:"totalSubmissions"`
	UniqueAgents     int                    `json:"uniqueAgents"`
	AverageScore     float64                `json:"averageScore"`
	BestScore        float64                `json:"bestScore"`
	MetricStats      map[string]MetricStat  `json:"metricStats"`
	LastSubmission   time.Time              `json:"lastSubmission"`
}

// MetricStat provides statistics for a single metric
type MetricStat struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"stdDev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	P90    float64 `json:"p90"`
}

// CompareInput for comparing submissions
type CompareInput struct {
	SubmissionIDA uuid.UUID `json:"submissionIdA" validate:"required"`
	SubmissionIDB uuid.UUID `json:"submissionIdB" validate:"required"`
}

// CreateBenchmarkInput for creating benchmarks
type CreateBenchmarkInput struct {
	Name         string            `json:"name" validate:"required"`
	Description  string            `json:"description,omitempty"`
	Category     BenchmarkCategory `json:"category" validate:"required"`
	DatasetID    uuid.UUID         `json:"datasetId" validate:"required"`
	EvaluatorIDs []uuid.UUID       `json:"evaluatorIds"`
	Metrics      []BenchmarkMetric `json:"metrics"`
	IsPublic     bool              `json:"isPublic"`
}
