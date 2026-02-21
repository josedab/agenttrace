package domain

import (
	"time"

	"github.com/google/uuid"
)

// RegressionTestStatus represents the status of a regression test
type RegressionTestStatus string

const (
	RegressionTestPending RegressionTestStatus = "PENDING"
	RegressionTestRunning RegressionTestStatus = "RUNNING"
	RegressionTestPassed  RegressionTestStatus = "PASSED"
	RegressionTestFailed  RegressionTestStatus = "FAILED"
)

// RegressionTest defines a regression test configuration
type RegressionTest struct {
	ID                uuid.UUID            `json:"id"`
	ProjectID         uuid.UUID            `json:"projectId"`
	Name              string               `json:"name"`
	BaselineDatasetID uuid.UUID            `json:"baselineDatasetId"`
	EvaluatorIDs      []uuid.UUID          `json:"evaluatorIds"`
	Thresholds        map[string]float64   `json:"thresholds"`
	Status            RegressionTestStatus `json:"status"`
	CreatedAt         time.Time            `json:"createdAt"`
}

// RegressionResult represents the outcome of running a regression test
type RegressionResult struct {
	ID             uuid.UUID          `json:"id"`
	TestID         uuid.UUID          `json:"testId"`
	RunID          uuid.UUID          `json:"runId"`
	Scores         map[string]float64 `json:"scores"`
	BaselineScores map[string]float64 `json:"baselineScores"`
	Passed         bool               `json:"passed"`
	Deltas         map[string]float64 `json:"deltas"`
	FailedMetrics  []string           `json:"failedMetrics"`
	CreatedAt      time.Time          `json:"createdAt"`
}

// RegressionTestInput represents input for creating a regression test
type RegressionTestInput struct {
	Name         string             `json:"name" validate:"required"`
	DatasetID    uuid.UUID          `json:"datasetId" validate:"required"`
	EvaluatorIDs []uuid.UUID        `json:"evaluatorIds" validate:"required,min=1"`
	Thresholds   map[string]float64 `json:"thresholds" validate:"required"`
}

// GateResult represents the outcome of a CI/CD quality gate check
type GateResult struct {
	Passed  bool               `json:"passed"`
	Results []RegressionResult `json:"results"`
}
