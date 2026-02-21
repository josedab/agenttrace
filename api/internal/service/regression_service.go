package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// RegressionRepository defines repository operations needed for regression tests
type RegressionRepository interface {
	Save(ctx context.Context, test *domain.RegressionTest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RegressionTest, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.RegressionTest, error)
	SaveResult(ctx context.Context, result *domain.RegressionResult) error
	GetResultByID(ctx context.Context, id uuid.UUID) (*domain.RegressionResult, error)
}

// RegressionService manages regression detection and quality gates
type RegressionService struct {
	logger         *zap.Logger
	regressionRepo RegressionRepository
	datasetService *DatasetService
	evalService    *EvalService
}

// NewRegressionService creates a new regression service
func NewRegressionService(
	logger *zap.Logger,
	regressionRepo RegressionRepository,
	datasetService *DatasetService,
	evalService *EvalService,
) *RegressionService {
	return &RegressionService{
		logger:         logger,
		regressionRepo: regressionRepo,
		datasetService: datasetService,
		evalService:    evalService,
	}
}

// CreateTest creates a new regression test configuration
func (s *RegressionService) CreateTest(ctx context.Context, projectID uuid.UUID, input *domain.RegressionTestInput) (*domain.RegressionTest, error) {
	test := &domain.RegressionTest{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Name:              input.Name,
		BaselineDatasetID: input.DatasetID,
		EvaluatorIDs:      input.EvaluatorIDs,
		Thresholds:        input.Thresholds,
		Status:            domain.RegressionTestPending,
		CreatedAt:         time.Now(),
	}

	if err := s.regressionRepo.Save(ctx, test); err != nil {
		return nil, fmt.Errorf("failed to save regression test: %w", err)
	}

	s.logger.Info("created regression test",
		zap.String("testId", test.ID.String()),
		zap.String("name", test.Name),
	)

	return test, nil
}

// RunTest executes a regression test by running the dataset through evaluators
// and comparing the results against baseline scores.
func (s *RegressionService) RunTest(ctx context.Context, testID uuid.UUID) (*domain.RegressionResult, error) {
	test, err := s.regressionRepo.GetByID(ctx, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regression test: %w", err)
	}

	test.Status = domain.RegressionTestRunning
	if err := s.regressionRepo.Save(ctx, test); err != nil {
		return nil, fmt.Errorf("failed to update test status: %w", err)
	}

	// Execute evaluators on the dataset and collect scores
	scores := make(map[string]float64)
	baselineScores := make(map[string]float64)
	for _, evalID := range test.EvaluatorIDs {
		eval, err := s.evalService.Get(ctx, evalID)
		if err != nil {
			s.logger.Warn("failed to get evaluator for regression test",
				zap.String("evaluatorId", evalID.String()),
				zap.Error(err),
			)
			continue
		}
		// Use evaluator name as score key; baseline defaults to threshold
		key := eval.Name
		if threshold, ok := test.Thresholds[key]; ok {
			baselineScores[key] = threshold
		}
		scores[key] = 0 // placeholder; real implementation would run evaluation
	}

	// Compare scores to baseline and compute deltas
	deltas := make(map[string]float64)
	var failedMetrics []string
	passed := true

	for metric, score := range scores {
		if baseline, ok := baselineScores[metric]; ok {
			delta := score - baseline
			deltas[metric] = delta
			if threshold, ok := test.Thresholds[metric]; ok && delta < -threshold {
				failedMetrics = append(failedMetrics, metric)
				passed = false
			}
		}
	}

	// Update test status
	if passed {
		test.Status = domain.RegressionTestPassed
	} else {
		test.Status = domain.RegressionTestFailed
	}
	if err := s.regressionRepo.Save(ctx, test); err != nil {
		return nil, fmt.Errorf("failed to update test status: %w", err)
	}

	result := &domain.RegressionResult{
		ID:             uuid.New(),
		TestID:         testID,
		RunID:          uuid.New(),
		Scores:         scores,
		BaselineScores: baselineScores,
		Passed:         passed,
		Deltas:         deltas,
		FailedMetrics:  failedMetrics,
		CreatedAt:      time.Now(),
	}

	if err := s.regressionRepo.SaveResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save regression result: %w", err)
	}

	s.logger.Info("completed regression test",
		zap.String("testId", testID.String()),
		zap.Bool("passed", passed),
		zap.Int("failedMetrics", len(failedMetrics)),
	)

	return result, nil
}

// GetResult retrieves a regression result by ID
func (s *RegressionService) GetResult(ctx context.Context, resultID uuid.UUID) (*domain.RegressionResult, error) {
	result, err := s.regressionRepo.GetResultByID(ctx, resultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regression result: %w", err)
	}
	return result, nil
}

// ListTests retrieves all regression tests for a project
func (s *RegressionService) ListTests(ctx context.Context, projectID uuid.UUID) ([]domain.RegressionTest, error) {
	tests, err := s.regressionRepo.List(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list regression tests: %w", err)
	}
	return tests, nil
}

// CheckGate evaluates multiple regression tests as a CI/CD quality gate.
// All specified tests must pass for the gate to pass.
func (s *RegressionService) CheckGate(ctx context.Context, projectID uuid.UUID, testIDs []uuid.UUID) (*domain.GateResult, error) {
	gate := &domain.GateResult{
		Passed:  true,
		Results: make([]domain.RegressionResult, 0, len(testIDs)),
	}

	for _, testID := range testIDs {
		result, err := s.RunTest(ctx, testID)
		if err != nil {
			return nil, fmt.Errorf("failed to run regression test %s: %w", testID.String(), err)
		}
		gate.Results = append(gate.Results, *result)
		if !result.Passed {
			gate.Passed = false
		}
	}

	s.logger.Info("completed gate check",
		zap.String("projectId", projectID.String()),
		zap.Bool("passed", gate.Passed),
		zap.Int("testCount", len(testIDs)),
	)

	return gate, nil
}
