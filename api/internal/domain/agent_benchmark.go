package domain

import (
	"time"

	"github.com/google/uuid"
)

// AgentBenchmarkCategory represents the category of an agent benchmark task
type AgentBenchmarkCategory string

const (
	AgentBenchmarkCategoryBugFix      AgentBenchmarkCategory = "bug_fix"
	AgentBenchmarkCategoryFeatureImpl AgentBenchmarkCategory = "feature_impl"
	AgentBenchmarkCategoryRefactoring AgentBenchmarkCategory = "refactoring"
	AgentBenchmarkCategoryCodeReview  AgentBenchmarkCategory = "code_review"
	AgentBenchmarkCategoryTestWriting AgentBenchmarkCategory = "test_writing"
)

// BenchmarkDifficulty represents the difficulty level of a benchmark task
type BenchmarkDifficulty string

const (
	BenchmarkDifficultyEasy   BenchmarkDifficulty = "easy"
	BenchmarkDifficultyMedium BenchmarkDifficulty = "medium"
	BenchmarkDifficultyHard   BenchmarkDifficulty = "hard"
	BenchmarkDifficultyExpert BenchmarkDifficulty = "expert"
)

// AgentBenchmarkSuite represents a suite of benchmark tasks for agent evaluation
type AgentBenchmarkSuite struct {
	ID          uuid.UUID         `json:"id"`
	ProjectID   uuid.UUID         `json:"projectId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    AgentBenchmarkCategory `json:"category"`
	Tasks       []BenchmarkTask   `json:"tasks"`
	CreatedAt   time.Time         `json:"createdAt"`
	CreatedBy   uuid.UUID         `json:"createdBy"`
}

// BenchmarkTask represents a single benchmark task within a suite
type BenchmarkTask struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Difficulty     BenchmarkDifficulty `json:"difficulty"`
	InputCode      string             `json:"inputCode"`
	ExpectedOutput string             `json:"expectedOutput"`
	EvalCriteria   map[string]float64 `json:"evalCriteria"`
	TimeoutSeconds int                `json:"timeoutSeconds"`
}

// BenchmarkRun represents a single run of a benchmark suite by an agent
type BenchmarkRun struct {
	ID           uuid.UUID    `json:"id"`
	SuiteID      uuid.UUID    `json:"suiteId"`
	AgentName    string       `json:"agentName"`
	ModelName    string       `json:"modelName"`
	Results      []AgentBenchmarkTaskResult `json:"results"`
	OverallScore float64      `json:"overallScore"`
	AvgLatencyMs float64      `json:"avgLatencyMs"`
	TotalCostUsd float64      `json:"totalCostUsd"`
	StartedAt    time.Time    `json:"startedAt"`
	CompletedAt  *time.Time   `json:"completedAt,omitempty"`
}

// AgentBenchmarkTaskResult represents the result of a single benchmark task execution
type AgentBenchmarkTaskResult struct {
	TaskID      uuid.UUID `json:"taskId"`
	Status      string    `json:"status"`
	Score       float64   `json:"score"`
	Output      string    `json:"output"`
	Explanation string    `json:"explanation"`
	LatencyMs   int64     `json:"latencyMs"`
	TokensUsed  int       `json:"tokensUsed"`
}

// AgentBenchmarkLeaderboard represents the leaderboard for an agent benchmark suite
type AgentBenchmarkLeaderboard struct {
	SuiteID   uuid.UUID                    `json:"suiteId"`
	Entries   []AgentBenchmarkLeaderEntry  `json:"entries"`
	UpdatedAt time.Time                    `json:"updatedAt"`
}

// AgentBenchmarkLeaderEntry represents a single entry in the agent benchmark leaderboard
type AgentBenchmarkLeaderEntry struct {
	Rank         int     `json:"rank"`
	AgentName    string  `json:"agentName"`
	ModelName    string  `json:"modelName"`
	OverallScore float64 `json:"overallScore"`
	AvgLatency   float64 `json:"avgLatency"`
	TotalCost    float64 `json:"totalCost"`
	RunCount     int     `json:"runCount"`
}

// AgentBenchmarkSuiteInput represents input for creating/updating a benchmark suite
type AgentBenchmarkSuiteInput struct {
	Name        string            `json:"name" validate:"required,min=1,max=200"`
	Description string            `json:"description,omitempty"`
	Category    AgentBenchmarkCategory `json:"category" validate:"required"`
	Tasks       []BenchmarkTask   `json:"tasks,omitempty"`
}

// BenchmarkRunInput represents input for creating a benchmark run
type BenchmarkRunInput struct {
	SuiteID   uuid.UUID `json:"suiteId" validate:"required"`
	AgentName string    `json:"agentName" validate:"required"`
	ModelName string    `json:"modelName" validate:"required"`
}

// AgentBenchmarkSuiteList represents a paginated list of benchmark suites
type AgentBenchmarkSuiteList struct {
	Suites     []AgentBenchmarkSuite `json:"suites"`
	TotalCount int64                 `json:"totalCount"`
	HasMore    bool                  `json:"hasMore"`
}

// BenchmarkRunList represents a paginated list of benchmark runs
type BenchmarkRunList struct {
	Runs       []BenchmarkRun `json:"runs"`
	TotalCount int64          `json:"totalCount"`
	HasMore    bool           `json:"hasMore"`
}
