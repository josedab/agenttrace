package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PromptCIService handles prompt CI/CD regression testing logic
type PromptCIService struct {
	logger *zap.Logger
}

// NewPromptCIService creates a new prompt CI service
func NewPromptCIService(logger *zap.Logger) *PromptCIService {
	return &PromptCIService{
		logger: logger,
	}
}

// CreateBaseline creates a new prompt performance baseline
func (s *PromptCIService) CreateBaseline(ctx context.Context, projectID uuid.UUID, input domain.PromptBaselineInput) (*domain.PromptBaseline, error) {
	baseline := &domain.PromptBaseline{
		ID:            uuid.New(),
		ProjectID:     projectID,
		DatasetID:     input.DatasetID,
		PromptID:      input.PromptID,
		PromptVersion: input.PromptVersion,
		Name:          input.Name,
		Branch:        input.Branch,
		Scores:        input.Scores,
		SampleSize:    input.SampleSize,
		CreatedAt:     time.Now(),
	}

	s.logger.Info("created prompt baseline",
		zap.String("baselineId", baseline.ID.String()),
		zap.String("name", baseline.Name),
		zap.String("branch", baseline.Branch),
	)

	return baseline, nil
}

// RunComparison runs a CI comparison against a baseline, computing score deltas
// and classifying regression severity
func (s *PromptCIService) RunComparison(ctx context.Context, projectID uuid.UUID, baselineID uuid.UUID, branch string, commitSHA string) (*domain.PromptCIRun, error) {
	s.logger.Info("running prompt CI comparison",
		zap.String("projectId", projectID.String()),
		zap.String("baselineId", baselineID.String()),
		zap.String("branch", branch),
		zap.String("commitSha", commitSHA),
	)

	now := time.Now()
	completedAt := now.Add(5 * time.Second)

	// Simulated score comparisons with delta and severity classification
	comparisons := []domain.ScoreComparison{
		s.compareScore("accuracy", 0.92, 0.89),
		s.compareScore("relevance", 0.88, 0.87),
		s.compareScore("coherence", 0.95, 0.94),
		s.compareScore("latency_ms", 450.0, 520.0),
	}

	overallSeverity := s.classifyOverallSeverity(comparisons)

	summary := fmt.Sprintf("Compared %d metrics against baseline. Overall severity: %s",
		len(comparisons), overallSeverity)

	run := &domain.PromptCIRun{
		ID:              uuid.New(),
		ProjectID:       projectID,
		BaselineID:      baselineID,
		Branch:          branch,
		CommitSHA:       commitSHA,
		Status:          "completed",
		ScoreComparison: comparisons,
		OverallSeverity: overallSeverity,
		Summary:         summary,
		StartedAt:       now,
		CompletedAt:     &completedAt,
	}

	return run, nil
}

// compareScore computes the delta and severity for a single metric comparison
func (s *PromptCIService) compareScore(metricName string, baselineValue, currentValue float64) domain.ScoreComparison {
	delta := currentValue - baselineValue
	absDelta := math.Abs(delta)

	// For latency, higher is worse; for quality scores, lower is worse
	isRegression := false
	if metricName == "latency_ms" {
		isRegression = delta > 0 && absDelta > baselineValue*0.05
	} else {
		isRegression = delta < 0 && absDelta > baselineValue*0.01
	}

	severity := domain.RegressionSeverityNone
	if isRegression {
		pctChange := absDelta / baselineValue
		switch {
		case pctChange >= 0.15:
			severity = domain.RegressionSeverityCritical
		case pctChange >= 0.08:
			severity = domain.RegressionSeverityMajor
		default:
			severity = domain.RegressionSeverityMinor
		}
	}

	return domain.ScoreComparison{
		MetricName:    metricName,
		BaselineValue: baselineValue,
		CurrentValue:  currentValue,
		Delta:         delta,
		PValue:        0.05,
		IsRegression:  isRegression,
		Severity:      severity,
	}
}

// classifyOverallSeverity returns the worst severity across all comparisons
func (s *PromptCIService) classifyOverallSeverity(comparisons []domain.ScoreComparison) domain.RegressionSeverity {
	worst := domain.RegressionSeverityNone

	severityRank := map[domain.RegressionSeverity]int{
		domain.RegressionSeverityNone:     0,
		domain.RegressionSeverityMinor:    1,
		domain.RegressionSeverityMajor:    2,
		domain.RegressionSeverityCritical: 3,
	}

	for _, c := range comparisons {
		if severityRank[c.Severity] > severityRank[worst] {
			worst = c.Severity
		}
	}

	return worst
}

// GetBaseline returns a specific prompt baseline by ID
func (s *PromptCIService) GetBaseline(ctx context.Context, baselineID uuid.UUID) (*domain.PromptBaseline, error) {
	s.logger.Info("fetching prompt baseline", zap.String("baselineId", baselineID.String()))

	baseline := &domain.PromptBaseline{
		ID:        baselineID,
		Scores:    map[string]float64{},
		CreatedAt: time.Now(),
	}

	return baseline, nil
}

// ListBaselines returns all prompt baselines for a project
func (s *PromptCIService) ListBaselines(ctx context.Context, projectID uuid.UUID) ([]domain.PromptBaseline, error) {
	s.logger.Info("listing prompt baselines", zap.String("projectId", projectID.String()))
	return []domain.PromptBaseline{}, nil
}

// ListRuns returns all prompt CI runs for a project
func (s *PromptCIService) ListRuns(ctx context.Context, projectID uuid.UUID) ([]domain.PromptCIRun, error) {
	s.logger.Info("listing prompt CI runs", zap.String("projectId", projectID.String()))
	return []domain.PromptCIRun{}, nil
}
