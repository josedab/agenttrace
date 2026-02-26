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

// ELORating represents an agent's ELO-style rating for a benchmark
type ELORating struct {
	ID          uuid.UUID `json:"id"`
	BenchmarkID uuid.UUID `json:"benchmarkId"`
	AgentName   string    `json:"agentName"`
	ProjectID   uuid.UUID `json:"projectId"`
	Rating      float64   `json:"rating"`
	Wins        int       `json:"wins"`
	Losses      int       `json:"losses"`
	Draws       int       `json:"draws"`
	TotalGames  int       `json:"totalGames"`
	Confidence  float64   `json:"confidence"` // 0-1 confidence interval
	UpdatedAt   time.Time `json:"updatedAt"`
}

// BenchmarkSuite represents a collection of related benchmarks
type BenchmarkSuite struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Benchmarks  []uuid.UUID `json:"benchmarkIds"`
	ScoringRule string      `json:"scoringRule"` // "average", "weighted", "min"
	IsPublic    bool        `json:"isPublic"`
	CreatedBy   uuid.UUID   `json:"createdBy"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// CreateBenchmarkSuiteInput for creating benchmark suites
type CreateBenchmarkSuiteInput struct {
	Name         string      `json:"name" validate:"required"`
	Description  string      `json:"description,omitempty"`
	BenchmarkIDs []uuid.UUID `json:"benchmarkIds" validate:"required,min=1"`
	ScoringRule  string      `json:"scoringRule,omitempty"`
	IsPublic     bool        `json:"isPublic"`
}

// CommunitySubmission represents a submission from the community API
type CommunitySubmission struct {
	ID           uuid.UUID          `json:"id"`
	BenchmarkID  uuid.UUID          `json:"benchmarkId"`
	SubmitterID  uuid.UUID          `json:"submitterId"`
	AgentName    string             `json:"agentName"`
	AgentVersion string             `json:"agentVersion"`
	RepoURL      string             `json:"repoUrl,omitempty"`
	Scores       map[string]float64 `json:"scores"`
	OverallScore float64            `json:"overallScore"`
	Verified     bool               `json:"verified"`
	Metadata     map[string]string  `json:"metadata,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
}

// CommunitySubmissionInput for community API submissions
type CommunitySubmissionInput struct {
	BenchmarkID  uuid.UUID          `json:"benchmarkId" validate:"required"`
	AgentName    string             `json:"agentName" validate:"required"`
	AgentVersion string             `json:"agentVersion" validate:"required"`
	RepoURL      string             `json:"repoUrl,omitempty"`
	Scores       map[string]float64 `json:"scores" validate:"required"`
	Metadata     map[string]string  `json:"metadata,omitempty"`
}

// GHABenchmarkConfig represents GitHub Actions integration config
type GHABenchmarkConfig struct {
	BenchmarkID  uuid.UUID `json:"benchmarkId"`
	RepoOwner    string    `json:"repoOwner"`
	RepoName     string    `json:"repoName"`
	WorkflowFile string    `json:"workflowFile"`
	TriggerEvent string    `json:"triggerEvent"` // push, pull_request, schedule
	Schedule     string    `json:"schedule,omitempty"` // cron expression
	Enabled      bool      `json:"enabled"`
}

// LeaderboardEntry extends BenchmarkSubmission with ELO data
type LeaderboardEntry struct {
	BenchmarkSubmission
	ELORating   float64 `json:"eloRating"`
	RatingDelta float64 `json:"ratingDelta"` // Change since last submission
	Trend       string  `json:"trend"`       // "up", "down", "stable"
}
