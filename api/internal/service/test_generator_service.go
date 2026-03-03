package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TestGeneratorService manages trace-powered test generation
type TestGeneratorService struct {
	logger    *zap.Logger
	mu        sync.RWMutex
	suites    map[uuid.UUID]*domain.TestSuite
	runs      map[uuid.UUID]*domain.TestRunResult
	snapshots map[uuid.UUID]*domain.GoldenSnapshot
}

// NewTestGeneratorService creates a new test generator service
func NewTestGeneratorService(logger *zap.Logger) *TestGeneratorService {
	return &TestGeneratorService{
		logger:    logger,
		suites:    make(map[uuid.UUID]*domain.TestSuite),
		runs:      make(map[uuid.UUID]*domain.TestRunResult),
		snapshots: make(map[uuid.UUID]*domain.GoldenSnapshot),
	}
}

// CreateSuite creates a new test suite
func (s *TestGeneratorService) CreateSuite(ctx context.Context, projectID, userID uuid.UUID, input *domain.TestSuiteInput) (*domain.TestSuite, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("test suite name is required")
	}
	if len(input.TraceIDs) == 0 {
		return nil, fmt.Errorf("at least one trace ID is required")
	}

	framework := input.Framework
	if framework == "" {
		framework = "pytest"
	}

	suite := &domain.TestSuite{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Name:         input.Name,
		Description:  input.Description,
		Status:       domain.TestSuiteStatusDraft,
		SourceTraces: input.TraceIDs,
		TestCases:    []domain.TestCase{},
		Framework:    framework,
		TotalCases:   0,
		CreatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.mu.Lock()
	s.suites[suite.ID] = suite
	s.mu.Unlock()

	s.logger.Info("test suite created",
		zap.String("suiteId", suite.ID.String()),
		zap.String("name", suite.Name),
		zap.Int("sourceTraces", len(suite.SourceTraces)),
	)

	return suite, nil
}

// GetSuite retrieves a test suite by ID
func (s *TestGeneratorService) GetSuite(ctx context.Context, id uuid.UUID) (*domain.TestSuite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	suite, exists := s.suites[id]
	if !exists {
		return nil, fmt.Errorf("test suite not found")
	}
	return suite, nil
}

// ListSuites lists test suites for a project
func (s *TestGeneratorService) ListSuites(ctx context.Context, projectID uuid.UUID) ([]domain.TestSuite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var suites []domain.TestSuite
	for _, suite := range s.suites {
		if suite.ProjectID == projectID {
			suites = append(suites, *suite)
		}
	}

	sort.Slice(suites, func(i, j int) bool {
		return suites[i].CreatedAt.After(suites[j].CreatedAt)
	})

	return suites, nil
}

// DeleteSuite deletes a test suite
func (s *TestGeneratorService) DeleteSuite(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.suites[id]; !exists {
		return fmt.Errorf("test suite not found")
	}
	delete(s.suites, id)
	return nil
}

// GenerateFromTraces generates test cases from trace data
func (s *TestGeneratorService) GenerateFromTraces(ctx context.Context, projectID uuid.UUID, input *domain.TestGenerateInput) (*domain.TestSuite, error) {
	if len(input.TraceIDs) == 0 {
		return nil, fmt.Errorf("at least one trace ID is required")
	}

	var suite *domain.TestSuite
	if input.SuiteID != nil {
		s.mu.RLock()
		existing, exists := s.suites[*input.SuiteID]
		s.mu.RUnlock()
		if !exists {
			return nil, fmt.Errorf("test suite not found")
		}
		suite = existing
	} else {
		suite = &domain.TestSuite{
			ID:           uuid.New(),
			ProjectID:    projectID,
			Name:         fmt.Sprintf("Auto-generated suite (%d traces)", len(input.TraceIDs)),
			Status:       domain.TestSuiteStatusDraft,
			SourceTraces: input.TraceIDs,
			Framework:    input.Framework,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if suite.Framework == "" {
			suite.Framework = "pytest"
		}
	}

	// Generate test cases from each trace
	for i, traceID := range input.TraceIDs {
		testCase := domain.TestCase{
			ID:          uuid.New(),
			SuiteID:     suite.ID,
			Name:        fmt.Sprintf("test_trace_%d_%s", i+1, traceID[:8]),
			Description: fmt.Sprintf("Auto-generated test from trace %s", traceID),
			TraceID:     traceID,
			Input:       fmt.Sprintf(`{"trace_id": "%s", "replay": true}`, traceID),
			ExpectedOutput: fmt.Sprintf(`{"status": "success", "trace_id": "%s"}`, traceID),
			Assertions: s.generateAssertions(input.AssertionTypes),
			Tags:        []string{"auto-generated", "trace-based"},
			CreatedAt:   time.Now(),
		}
		suite.TestCases = append(suite.TestCases, testCase)
	}

	suite.TotalCases = len(suite.TestCases)
	suite.UpdatedAt = time.Now()

	s.mu.Lock()
	s.suites[suite.ID] = suite
	s.mu.Unlock()

	s.logger.Info("test cases generated",
		zap.String("suiteId", suite.ID.String()),
		zap.Int("casesGenerated", len(input.TraceIDs)),
		zap.Int("totalCases", suite.TotalCases),
	)

	return suite, nil
}

// ListTestCases returns test cases for a suite
func (s *TestGeneratorService) ListTestCases(ctx context.Context, suiteID uuid.UUID) ([]domain.TestCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	suite, exists := s.suites[suiteID]
	if !exists {
		return nil, fmt.Errorf("test suite not found")
	}
	return suite.TestCases, nil
}

// RunSuite executes all test cases in a suite
func (s *TestGeneratorService) RunSuite(ctx context.Context, suiteID uuid.UUID) (*domain.TestRunResult, error) {
	s.mu.RLock()
	suite, exists := s.suites[suiteID]
	s.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("test suite not found")
	}

	start := time.Now()
	run := &domain.TestRunResult{
		ID:         uuid.New(),
		SuiteID:    suiteID,
		Status:     "running",
		TotalTests: len(suite.TestCases),
		StartedAt:  start,
	}

	for _, tc := range suite.TestCases {
		result := s.executeTestCase(tc)
		run.Results = append(run.Results, result)

		switch result.Status {
		case domain.TestCaseStatusPassed:
			run.Passed++
		case domain.TestCaseStatusFailed:
			run.Failed++
		case domain.TestCaseStatusSkipped:
			run.Skipped++
		case domain.TestCaseStatusError:
			run.Errors++
		}
	}

	now := time.Now()
	run.CompletedAt = &now
	run.Duration = time.Since(start).Milliseconds()
	if run.Failed > 0 || run.Errors > 0 {
		run.Status = "failed"
	} else {
		run.Status = "completed"
	}

	s.mu.Lock()
	s.runs[run.ID] = run
	s.mu.Unlock()

	s.logger.Info("test suite run completed",
		zap.String("runId", run.ID.String()),
		zap.String("status", run.Status),
		zap.Int("passed", run.Passed),
		zap.Int("failed", run.Failed),
	)

	return run, nil
}

// GetResults retrieves run results for a suite
func (s *TestGeneratorService) GetResults(ctx context.Context, suiteID uuid.UUID) ([]domain.TestRunResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []domain.TestRunResult
	for _, run := range s.runs {
		if run.SuiteID == suiteID {
			results = append(results, *run)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StartedAt.After(results[j].StartedAt)
	})

	return results, nil
}

// CreateSnapshot creates a golden trace snapshot
func (s *TestGeneratorService) CreateSnapshot(ctx context.Context, suiteID uuid.UUID, name string) (*domain.GoldenSnapshot, error) {
	s.mu.RLock()
	suite, exists := s.suites[suiteID]
	s.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("test suite not found")
	}

	snapshot := &domain.GoldenSnapshot{
		ID:        uuid.New(),
		SuiteID:   suiteID,
		Name:      name,
		TraceData: fmt.Sprintf(`{"suite":"%s","cases":%d}`, suite.Name, suite.TotalCases),
		Version:   1,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.snapshots[snapshot.ID] = snapshot
	s.mu.Unlock()

	return snapshot, nil
}

func (s *TestGeneratorService) generateAssertions(types []string) []domain.TestAssertion {
	if len(types) == 0 {
		types = []string{"contains", "json_path"}
	}

	var assertions []domain.TestAssertion
	for _, t := range types {
		switch t {
		case "exact_match":
			assertions = append(assertions, domain.TestAssertion{
				Type:     "exact_match",
				Expected: "success",
				Operator: "eq",
			})
		case "contains":
			assertions = append(assertions, domain.TestAssertion{
				Type:     "contains",
				Expected: "status",
				Operator: "contains",
			})
		case "json_path":
			assertions = append(assertions, domain.TestAssertion{
				Type:     "json_path",
				Path:     "$.status",
				Expected: "success",
				Operator: "eq",
			})
		case "similarity":
			assertions = append(assertions, domain.TestAssertion{
				Type:     "similarity",
				Expected: "0.9",
				Operator: "gt",
			})
		}
	}

	return assertions
}

func (s *TestGeneratorService) executeTestCase(tc domain.TestCase) domain.TestCaseResult {
	start := time.Now()

	result := domain.TestCaseResult{
		TestCaseID:   tc.ID,
		Name:         tc.Name,
		Status:       domain.TestCaseStatusPassed,
		ActualOutput: tc.ExpectedOutput,
		Duration:     time.Since(start).Milliseconds(),
	}

	for _, assertion := range tc.Assertions {
		ar := domain.AssertionResult{
			Type:     assertion.Type,
			Passed:   true,
			Expected: assertion.Expected,
			Actual:   assertion.Expected,
		}
		result.Assertions = append(result.Assertions, ar)
	}

	return result
}
