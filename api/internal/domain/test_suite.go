package domain

import (
	"time"

	"github.com/google/uuid"
)

// TestSuiteStatus represents the status of a test suite
type TestSuiteStatus string

const (
	TestSuiteStatusDraft    TestSuiteStatus = "draft"
	TestSuiteStatusActive   TestSuiteStatus = "active"
	TestSuiteStatusArchived TestSuiteStatus = "archived"
)

// TestCaseStatus represents the status of a test case result
type TestCaseStatus string

const (
	TestCaseStatusPassed  TestCaseStatus = "passed"
	TestCaseStatusFailed  TestCaseStatus = "failed"
	TestCaseStatusSkipped TestCaseStatus = "skipped"
	TestCaseStatusError   TestCaseStatus = "error"
)

// TestSuite represents a collection of generated test cases
type TestSuite struct {
	ID          uuid.UUID       `json:"id"`
	ProjectID   uuid.UUID       `json:"projectId"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Status      TestSuiteStatus `json:"status"`
	SourceTraces []string       `json:"sourceTraces"`
	TestCases   []TestCase      `json:"testCases"`
	Framework   string          `json:"framework"` // "pytest", "jest", "custom"
	TotalCases  int             `json:"totalCases"`
	CreatedBy   uuid.UUID       `json:"createdBy"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// TestCase represents a single test case generated from a trace
type TestCase struct {
	ID          uuid.UUID `json:"id"`
	SuiteID     uuid.UUID `json:"suiteId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	TraceID     string    `json:"traceId"`
	Input       string    `json:"input"`
	ExpectedOutput string `json:"expectedOutput"`
	Assertions  []TestAssertion `json:"assertions"`
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TestAssertion represents a single assertion in a test case
type TestAssertion struct {
	Type     string `json:"type"`     // "exact_match", "contains", "json_path", "regex", "similarity"
	Path     string `json:"path,omitempty"`
	Expected string `json:"expected"`
	Operator string `json:"operator,omitempty"` // "eq", "neq", "gt", "lt", "contains"
}

// TestRunResult represents the result of running a test suite
type TestRunResult struct {
	ID         uuid.UUID              `json:"id"`
	SuiteID    uuid.UUID              `json:"suiteId"`
	Status     string                 `json:"status"` // "running", "completed", "failed"
	TotalTests int                    `json:"totalTests"`
	Passed     int                    `json:"passed"`
	Failed     int                    `json:"failed"`
	Skipped    int                    `json:"skipped"`
	Errors     int                    `json:"errors"`
	Duration   int64                  `json:"durationMs"`
	Results    []TestCaseResult       `json:"results"`
	StartedAt  time.Time              `json:"startedAt"`
	CompletedAt *time.Time            `json:"completedAt,omitempty"`
}

// TestCaseResult represents the result of a single test case
type TestCaseResult struct {
	TestCaseID uuid.UUID      `json:"testCaseId"`
	Name       string         `json:"name"`
	Status     TestCaseStatus `json:"status"`
	ActualOutput string       `json:"actualOutput,omitempty"`
	Error      string         `json:"error,omitempty"`
	Duration   int64          `json:"durationMs"`
	Assertions []AssertionResult `json:"assertions"`
}

// AssertionResult represents the result of a single assertion
type AssertionResult struct {
	Type    string `json:"type"`
	Passed  bool   `json:"passed"`
	Expected string `json:"expected"`
	Actual  string `json:"actual,omitempty"`
	Message string `json:"message,omitempty"`
}

// GoldenSnapshot represents a golden trace snapshot for acceptance testing
type GoldenSnapshot struct {
	ID        uuid.UUID `json:"id"`
	SuiteID   uuid.UUID `json:"suiteId"`
	Name      string    `json:"name"`
	TraceData string    `json:"traceData"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
}

// TestSuiteInput represents input for creating a test suite
type TestSuiteInput struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description,omitempty"`
	Framework   string   `json:"framework,omitempty"`
	TraceIDs    []string `json:"traceIds" validate:"required"`
}

// TestGenerateInput represents input for generating tests from traces
type TestGenerateInput struct {
	SuiteID       *uuid.UUID `json:"suiteId,omitempty"`
	TraceIDs      []string   `json:"traceIds" validate:"required"`
	AssertionTypes []string  `json:"assertionTypes,omitempty"`
	Framework     string     `json:"framework,omitempty"`
}
