package domain

import (
	"time"

	"github.com/google/uuid"
)

// GoldenDatasetCategory represents the category of a golden dataset
type GoldenDatasetCategory string

const (
	GoldenDatasetCategoryBugFix      GoldenDatasetCategory = "bug_fix"
	GoldenDatasetCategoryRefactoring GoldenDatasetCategory = "refactoring"
	GoldenDatasetCategoryTestWriting GoldenDatasetCategory = "test_writing"
	GoldenDatasetCategoryFeatureImpl GoldenDatasetCategory = "feature_impl"
	GoldenDatasetCategoryCodeReview  GoldenDatasetCategory = "code_review"
)

// IsValid checks if the golden dataset category is valid
func (c GoldenDatasetCategory) IsValid() bool {
	switch c {
	case GoldenDatasetCategoryBugFix, GoldenDatasetCategoryRefactoring, GoldenDatasetCategoryTestWriting, GoldenDatasetCategoryFeatureImpl, GoldenDatasetCategoryCodeReview:
		return true
	}
	return false
}

// RegressionRunStatus represents the status of a regression test run
type RegressionRunStatus string

const (
	RegressionRunStatusPending RegressionRunStatus = "pending"
	RegressionRunStatusRunning RegressionRunStatus = "running"
	RegressionRunStatusPassed  RegressionRunStatus = "passed"
	RegressionRunStatusFailed  RegressionRunStatus = "failed"
	RegressionRunStatusError   RegressionRunStatus = "error"
)

// IsValid checks if the regression run status is valid
func (s RegressionRunStatus) IsValid() bool {
	switch s {
	case RegressionRunStatusPending, RegressionRunStatusRunning, RegressionRunStatusPassed, RegressionRunStatusFailed, RegressionRunStatusError:
		return true
	}
	return false
}

// GoldenDataset represents a curated set of test cases for agent regression testing
type GoldenDataset struct {
	ID          uuid.UUID             `json:"id"`
	ProjectID   uuid.UUID             `json:"projectId"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Category    GoldenDatasetCategory `json:"category"`
	Language    string                `json:"language"`
	Items       []GoldenDatasetItem   `json:"items"`
	ItemCount   int                   `json:"itemCount"`
	CreatedAt   time.Time             `json:"createdAt"`
	CreatedBy   uuid.UUID             `json:"createdBy"`
}

// GoldenDatasetItem represents a single test case within a golden dataset
type GoldenDatasetItem struct {
	ID               uuid.UUID          `json:"id"`
	Input            string             `json:"input"`
	ExpectedBehavior string             `json:"expectedBehavior"`
	EvalCriteria     map[string]float64 `json:"evalCriteria"`
	Difficulty       string             `json:"difficulty"`
	Tags             []string           `json:"tags,omitempty"`
}

// RegressionRun represents a single regression test suite run
type RegressionRun struct {
	ID                 uuid.UUID           `json:"id"`
	ProjectID          uuid.UUID           `json:"projectId"`
	SuiteID            uuid.UUID           `json:"suiteId"`
	AgentConfig        string              `json:"agentConfig"`
	Status             RegressionRunStatus `json:"status"`
	Results            []RegressionRunResult `json:"results"`
	PassRate           float64             `json:"passRate"`
	TotalTests         int                 `json:"totalTests"`
	Passed             int                 `json:"passed"`
	Failed             int                 `json:"failed"`
	BaselineComparison *BaselineComparison `json:"baselineComparison,omitempty"`
	StartedAt          *time.Time          `json:"startedAt,omitempty"`
	CompletedAt        *time.Time          `json:"completedAt,omitempty"`
}

// RegressionRunResult represents the outcome of a single test case in a regression run
type RegressionRunResult struct {
	ItemID      uuid.UUID `json:"itemId"`
	Passed      bool      `json:"passed"`
	Score       float64   `json:"score"`
	ActualOutput string   `json:"actualOutput"`
	Explanation string    `json:"explanation,omitempty"`
	LatencyMs   int64     `json:"latencyMs"`
	CostUSD     float64   `json:"costUsd"`
}

// BaselineComparison represents a comparison between current and baseline regression results
type BaselineComparison struct {
	BaselinePassRate float64 `json:"baselinePassRate"`
	CurrentPassRate  float64 `json:"currentPassRate"`
	ScoreDelta       float64 `json:"scoreDelta"`
	StatSignificant  bool    `json:"statSignificant"`
	PValue           float64 `json:"pValue"`
}
