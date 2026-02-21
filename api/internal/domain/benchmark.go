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
