package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// RegressionSuiteService handles golden dataset management and regression test runs
type RegressionSuiteService struct {
	logger *zap.Logger
}

// NewRegressionSuiteService creates a new regression suite service
func NewRegressionSuiteService(logger *zap.Logger) *RegressionSuiteService {
	return &RegressionSuiteService{
		logger: logger,
	}
}

// categoryTemplates defines sample golden dataset items per category
var categoryTemplates = map[domain.GoldenDatasetCategory][]struct {
	input    string
	expected string
	criteria map[string]float64
	diff     string
}{
	domain.GoldenDatasetCategoryBugFix: {
		{input: "Fix the null pointer exception in user authentication when email is empty", expected: "Add nil check before accessing email field", criteria: map[string]float64{"correctness": 0.9, "completeness": 0.8}, diff: "medium"},
		{input: "Resolve race condition in concurrent cache updates", expected: "Implement mutex locking around cache write operations", criteria: map[string]float64{"correctness": 0.95, "safety": 0.9}, diff: "hard"},
	},
	domain.GoldenDatasetCategoryRefactoring: {
		{input: "Extract the payment processing logic into a separate service", expected: "Create PaymentService with clean interface boundaries", criteria: map[string]float64{"code_quality": 0.85, "maintainability": 0.9}, diff: "medium"},
		{input: "Replace callback-based async code with async/await pattern", expected: "Convert all .then() chains to async/await with proper error handling", criteria: map[string]float64{"readability": 0.9, "correctness": 0.85}, diff: "easy"},
	},
	domain.GoldenDatasetCategoryTestWriting: {
		{input: "Write unit tests for the OrderService.calculateTotal method", expected: "Cover edge cases: empty cart, discounts, tax calculations, currency rounding", criteria: map[string]float64{"coverage": 0.9, "edge_cases": 0.85}, diff: "easy"},
		{input: "Create integration tests for the REST API authentication flow", expected: "Test login, token refresh, logout, and invalid credential scenarios", criteria: map[string]float64{"coverage": 0.85, "completeness": 0.9}, diff: "medium"},
	},
	domain.GoldenDatasetCategoryFeatureImpl: {
		{input: "Implement a rate limiter using the token bucket algorithm", expected: "Token bucket with configurable rate, burst size, and per-client tracking", criteria: map[string]float64{"correctness": 0.9, "performance": 0.85}, diff: "hard"},
	},
	domain.GoldenDatasetCategoryCodeReview: {
		{input: "Review this PR for security vulnerabilities and suggest improvements", expected: "Identify SQL injection risks, missing input validation, and hardcoded secrets", criteria: map[string]float64{"thoroughness": 0.9, "actionability": 0.85}, diff: "medium"},
	},
}

// CreateGoldenDataset creates a new golden dataset with sample items based on category
func (s *RegressionSuiteService) CreateGoldenDataset(ctx context.Context, projectID uuid.UUID, input domain.GoldenDataset) (*domain.GoldenDataset, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("dataset name is required")
	}
	if !input.Category.IsValid() {
		return nil, fmt.Errorf("invalid category: %s", input.Category)
	}

	now := time.Now()
	dataset := &domain.GoldenDataset{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Language:    input.Language,
		CreatedAt:   now,
	}

	// Generate sample items from category templates
	templates, ok := categoryTemplates[input.Category]
	if ok {
		for _, tmpl := range templates {
			item := domain.GoldenDatasetItem{
				ID:               uuid.New(),
				Input:            tmpl.input,
				ExpectedBehavior: tmpl.expected,
				EvalCriteria:     tmpl.criteria,
				Difficulty:       tmpl.diff,
				Tags:             []string{string(input.Category), input.Language},
			}
			dataset.Items = append(dataset.Items, item)
		}
	}
	dataset.ItemCount = len(dataset.Items)

	s.logger.Info("golden dataset created",
		zap.String("id", dataset.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", dataset.Name),
		zap.String("category", string(dataset.Category)),
		zap.Int("itemCount", dataset.ItemCount),
	)
	return dataset, nil
}

// GetGoldenDataset retrieves a golden dataset by ID
func (s *RegressionSuiteService) GetGoldenDataset(ctx context.Context, id uuid.UUID) (*domain.GoldenDataset, error) {
	s.logger.Debug("fetching golden dataset", zap.String("id", id.String()))

	return &domain.GoldenDataset{
		ID:        id,
		Name:      "Sample Dataset",
		Category:  domain.GoldenDatasetCategoryBugFix,
		Items:     []domain.GoldenDatasetItem{},
		ItemCount: 0,
		CreatedAt: time.Now(),
	}, nil
}

// ListGoldenDatasets lists all golden datasets for a project
func (s *RegressionSuiteService) ListGoldenDatasets(ctx context.Context, projectID uuid.UUID) ([]domain.GoldenDataset, error) {
	s.logger.Debug("listing golden datasets", zap.String("projectId", projectID.String()))
	return []domain.GoldenDataset{}, nil
}

// RunRegression executes a regression test run against a golden dataset, producing
// mock results with pass/fail outcomes and score comparisons
func (s *RegressionSuiteService) RunRegression(ctx context.Context, projectID, suiteID uuid.UUID, agentConfig string) (*domain.RegressionRun, error) {
	if agentConfig == "" {
		return nil, fmt.Errorf("agent config is required")
	}

	dataset, err := s.GetGoldenDataset(ctx, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch golden dataset: %w", err)
	}

	now := time.Now()
	rng := rand.New(rand.NewSource(now.UnixNano()))

	var results []domain.RegressionRunResult
	passed, failed := 0, 0
	totalItems := len(dataset.Items)
	if totalItems == 0 {
		totalItems = 5 // simulate with 5 items if dataset is empty
		for i := 0; i < totalItems; i++ {
			dataset.Items = append(dataset.Items, domain.GoldenDatasetItem{
				ID:    uuid.New(),
				Input: fmt.Sprintf("Test case %d", i+1),
			})
		}
	}

	for _, item := range dataset.Items {
		score := 0.3 + rng.Float64()*0.7
		itemPassed := score >= 0.6
		if itemPassed {
			passed++
		} else {
			failed++
		}

		explanation := "Output matched expected behavior within tolerance"
		if !itemPassed {
			explanation = "Output diverged significantly from expected behavior"
		}

		results = append(results, domain.RegressionRunResult{
			ItemID:       item.ID,
			Passed:       itemPassed,
			Score:        score,
			ActualOutput: fmt.Sprintf("Generated response for: %s", item.Input),
			Explanation:  explanation,
			LatencyMs:    500 + rng.Int63n(2500),
			CostUSD:      0.005 + rng.Float64()*0.04,
		})
	}

	passRate := 0.0
	if totalItems > 0 {
		passRate = float64(passed) / float64(totalItems) * 100
	}

	completedAt := time.Now()
	status := domain.RegressionRunStatusPassed
	if passRate < 80 {
		status = domain.RegressionRunStatusFailed
	}

	baselinePassRate := 75.0 + rng.Float64()*20
	run := &domain.RegressionRun{
		ID:          uuid.New(),
		ProjectID:   projectID,
		SuiteID:     suiteID,
		AgentConfig: agentConfig,
		Status:      status,
		Results:     results,
		PassRate:    passRate,
		TotalTests:  totalItems,
		Passed:      passed,
		Failed:      failed,
		BaselineComparison: &domain.BaselineComparison{
			BaselinePassRate: baselinePassRate,
			CurrentPassRate:  passRate,
			ScoreDelta:       passRate - baselinePassRate,
			StatSignificant:  passed+failed >= 30,
			PValue:           0.01 + rng.Float64()*0.1,
		},
		StartedAt:   &now,
		CompletedAt: &completedAt,
	}

	s.logger.Info("regression run completed",
		zap.String("id", run.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("status", string(status)),
		zap.Float64("passRate", passRate),
		zap.Int("passed", passed),
		zap.Int("failed", failed),
	)
	return run, nil
}

// GetRun retrieves a regression run by ID
func (s *RegressionSuiteService) GetRun(ctx context.Context, id uuid.UUID) (*domain.RegressionRun, error) {
	s.logger.Debug("fetching regression run", zap.String("id", id.String()))

	return &domain.RegressionRun{
		ID:     id,
		Status: domain.RegressionRunStatusPassed,
	}, nil
}

// ListRuns lists all regression runs for a project
func (s *RegressionSuiteService) ListRuns(ctx context.Context, projectID uuid.UUID) ([]domain.RegressionRun, error) {
	s.logger.Debug("listing regression runs", zap.String("projectId", projectID.String()))
	return []domain.RegressionRun{}, nil
}
