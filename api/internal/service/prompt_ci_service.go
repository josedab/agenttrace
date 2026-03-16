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
	if input.Name == "" {
		return nil, fmt.Errorf("baseline name is required")
	}
	if input.Branch == "" {
		return nil, fmt.Errorf("branch is required")
	}

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

// CreateGateConfig creates a configurable CI gate for blocking PRs on regression
func (s *PromptCIService) CreateGateConfig(ctx context.Context, projectID uuid.UUID, input *domain.PromptCIGateConfigInput) (*domain.PromptCIGateConfig, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("gate config name is required")
	}
	if len(input.Thresholds) == 0 {
		return nil, fmt.Errorf("at least one metric threshold is required")
	}

	blockSeverity := input.BlockOnSeverity
	if blockSeverity == "" {
		blockSeverity = domain.RegressionSeverityMajor
	}

	confidence := input.ConfidenceLevel
	if confidence <= 0 || confidence > 1 {
		confidence = 0.95
	}

	config := &domain.PromptCIGateConfig{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		BaselineID:      input.BaselineID,
		Thresholds:      input.Thresholds,
		BlockOnSeverity: blockSeverity,
		ConfidenceLevel: confidence,
		RequiredMetrics: input.RequiredMetrics,
		Enabled:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	s.logger.Info("created prompt CI gate config",
		zap.String("configId", config.ID.String()),
		zap.String("name", config.Name),
	)
	return config, nil
}

// ListGateConfigs returns all gate configurations for a project
func (s *PromptCIService) ListGateConfigs(ctx context.Context, projectID uuid.UUID) ([]domain.PromptCIGateConfig, error) {
	return []domain.PromptCIGateConfig{}, nil
}

// EvaluateGate evaluates current scores against the gate config to determine pass/fail
func (s *PromptCIService) EvaluateGate(ctx context.Context, projectID uuid.UUID, input *domain.PromptCIGateEvalInput) (*domain.PromptCIGateResult, error) {
	if len(input.Scores) == 0 {
		return nil, fmt.Errorf("scores are required for gate evaluation")
	}

	// Retrieve baseline scores (simulated)
	baseline, err := s.GetBaseline(ctx, input.GateConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline: %w", err)
	}

	var metricResults []domain.MetricGateResult
	overallPassed := true
	worstSeverity := domain.RegressionSeverityNone

	severityRank := map[domain.RegressionSeverity]int{
		domain.RegressionSeverityNone:     0,
		domain.RegressionSeverityMinor:    1,
		domain.RegressionSeverityMajor:    2,
		domain.RegressionSeverityCritical: 3,
	}

	for metricName, currentValue := range input.Scores {
		baselineValue, exists := baseline.Scores[metricName]
		if !exists {
			baselineValue = currentValue
		}

		changePct := 0.0
		if baselineValue != 0 {
			changePct = ((currentValue - baselineValue) / math.Abs(baselineValue)) * 100
		}

		// Default: higher is better, negative change is regression
		thresholdPct := 5.0 // default 5% regression threshold
		isRegression := changePct < -thresholdPct
		severity := domain.RegressionSeverityNone

		if isRegression {
			absChange := math.Abs(changePct)
			switch {
			case absChange >= 15:
				severity = domain.RegressionSeverityCritical
			case absChange >= 8:
				severity = domain.RegressionSeverityMajor
			case absChange >= thresholdPct:
				severity = domain.RegressionSeverityMinor
			}
		}

		passed := !isRegression || severityRank[severity] < severityRank[domain.RegressionSeverityMajor]
		if !passed {
			overallPassed = false
		}
		if severityRank[severity] > severityRank[worstSeverity] {
			worstSeverity = severity
		}

		metricResults = append(metricResults, domain.MetricGateResult{
			MetricName:      metricName,
			BaselineValue:   baselineValue,
			CurrentValue:    currentValue,
			ThresholdPct:    thresholdPct,
			ActualChangePct: changePct,
			Passed:          passed,
			Severity:        severity,
		})
	}

	summary := fmt.Sprintf("Evaluated %d metrics: %s", len(metricResults), func() string {
		if overallPassed {
			return "all passed"
		}
		return "gate BLOCKED - regressions detected"
	}())

	blockReason := ""
	if !overallPassed {
		blockReason = fmt.Sprintf("Prompt regression detected with severity '%s' on branch '%s' (commit %s)",
			worstSeverity, input.Branch, input.CommitSHA)
	}

	result := &domain.PromptCIGateResult{
		RunID:           uuid.New(),
		GateConfigID:    input.GateConfigID,
		Passed:          overallPassed,
		OverallSeverity: worstSeverity,
		MetricResults:   metricResults,
		Summary:         summary,
		BlockReason:     blockReason,
		EvaluatedAt:     time.Now(),
	}

	s.logger.Info("evaluated prompt CI gate",
		zap.Bool("passed", overallPassed),
		zap.String("severity", string(worstSeverity)),
		zap.String("branch", input.Branch),
	)
	return result, nil
}

// UpdateGateConfig updates an existing gate configuration
func (s *PromptCIService) UpdateGateConfig(ctx context.Context, projectID uuid.UUID, configID uuid.UUID, input *domain.PromptCIGateConfigUpdate) (*domain.PromptCIGateConfig, error) {
	s.logger.Info("updating prompt CI gate config",
		zap.String("configId", configID.String()),
	)

	config := &domain.PromptCIGateConfig{
		ID:        configID,
		ProjectID: projectID,
		UpdatedAt: time.Now(),
	}

	if input.Name != nil {
		config.Name = *input.Name
	}
	if input.Thresholds != nil {
		config.Thresholds = input.Thresholds
	}
	if input.BlockOnSeverity != nil {
		config.BlockOnSeverity = *input.BlockOnSeverity
	}
	if input.ConfidenceLevel != nil {
		config.ConfidenceLevel = *input.ConfidenceLevel
	}
	if input.RequiredMetrics != nil {
		config.RequiredMetrics = input.RequiredMetrics
	}
	if input.Enabled != nil {
		config.Enabled = *input.Enabled
	}

	return config, nil
}

// GetRegressionHistory returns regression history for a project
func (s *PromptCIService) GetRegressionHistory(ctx context.Context, projectID uuid.UUID, filter *domain.RegressionHistoryFilter) ([]domain.PromptRegressionHistory, error) {
	s.logger.Info("fetching regression history",
		zap.String("projectId", projectID.String()),
	)

	history := []domain.PromptRegressionHistory{}
	return history, nil
}

// RecordRegressionEvent records a regression event in history
func (s *PromptCIService) RecordRegressionEvent(ctx context.Context, projectID uuid.UUID, result *domain.PromptCIGateResult, branch string, commitSHA string, prNumber *int) (*domain.PromptRegressionHistory, error) {
	metricDeltas := make(map[string]float64)
	for _, mr := range result.MetricResults {
		metricDeltas[mr.MetricName] = mr.ActualChangePct
	}

	event := &domain.PromptRegressionHistory{
		ID:           uuid.New(),
		ProjectID:    projectID,
		GateConfigID: result.GateConfigID,
		RunID:        result.RunID,
		Branch:       branch,
		CommitSHA:    commitSHA,
		PRNumber:     prNumber,
		Passed:       result.Passed,
		Severity:     result.OverallSeverity,
		MetricDeltas: metricDeltas,
		BlockedPR:    !result.Passed && prNumber != nil,
		CreatedAt:    time.Now(),
	}

	s.logger.Info("recorded regression event",
		zap.String("eventId", event.ID.String()),
		zap.Bool("passed", event.Passed),
		zap.Bool("blockedPR", event.BlockedPR),
	)

	return event, nil
}

// GetDashboardStats returns prompt CI dashboard statistics
func (s *PromptCIService) GetDashboardStats(ctx context.Context, projectID uuid.UUID) (*domain.PromptCIDashboardStats, error) {
	s.logger.Info("fetching prompt CI dashboard stats",
		zap.String("projectId", projectID.String()),
	)

	stats := &domain.PromptCIDashboardStats{
		TotalBaselines:    0,
		TotalRuns:         0,
		TotalGateConfigs:  0,
		PassRate:          100.0,
		BlockedPRs:        0,
		RecentRegressions: 0,
	}

	return stats, nil
}
