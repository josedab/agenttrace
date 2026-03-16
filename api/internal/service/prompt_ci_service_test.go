package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestPromptCICreateBaseline(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	input := domain.PromptBaselineInput{
		Name:          "v1 Baseline",
		DatasetID:     uuid.New(),
		PromptID:      uuid.New(),
		PromptVersion: 1,
		Branch:        "main",
		Scores: map[string]float64{
			"accuracy":  0.92,
			"relevance": 0.88,
		},
		SampleSize: 100,
	}

	baseline, err := svc.CreateBaseline(ctx, projectID, input)
	require.NoError(t, err)
	require.NotNil(t, baseline)

	assert.NotEqual(t, uuid.Nil, baseline.ID)
	assert.Equal(t, projectID, baseline.ProjectID)
	assert.Equal(t, "v1 Baseline", baseline.Name)
	assert.Equal(t, input.DatasetID, baseline.DatasetID)
	assert.Equal(t, input.PromptID, baseline.PromptID)
	assert.Equal(t, 1, baseline.PromptVersion)
	assert.Equal(t, "main", baseline.Branch)
	assert.Equal(t, 0.92, baseline.Scores["accuracy"])
	assert.Equal(t, 0.88, baseline.Scores["relevance"])
	assert.Equal(t, 100, baseline.SampleSize)
	assert.False(t, baseline.CreatedAt.IsZero())
}

func TestPromptCICreateBaselineValidation(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()

	t.Run("empty name fails", func(t *testing.T) {
		input := domain.PromptBaselineInput{
			Name:   "",
			Branch: "main",
			Scores: map[string]float64{"accuracy": 0.9},
		}
		_, err := svc.CreateBaseline(ctx, uuid.New(), input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("empty branch fails", func(t *testing.T) {
		input := domain.PromptBaselineInput{
			Name:   "test",
			Branch: "",
		}
		_, err := svc.CreateBaseline(ctx, uuid.New(), input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "branch is required")
	})
}

func TestPromptCIListBaselines(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()

	baselines, err := svc.ListBaselines(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, baselines)
}

func TestPromptCIRunComparison(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	baselineID := uuid.New()

	run, err := svc.RunComparison(ctx, projectID, baselineID, "feature-branch", "abc123def")
	require.NoError(t, err)
	require.NotNil(t, run)

	assert.NotEqual(t, uuid.Nil, run.ID)
	assert.Equal(t, projectID, run.ProjectID)
	assert.Equal(t, baselineID, run.BaselineID)
	assert.Equal(t, "feature-branch", run.Branch)
	assert.Equal(t, "abc123def", run.CommitSHA)
	assert.Equal(t, "completed", run.Status)
	assert.NotEmpty(t, run.ScoreComparison)
	assert.NotEmpty(t, run.Summary)
	assert.NotNil(t, run.CompletedAt)
	assert.False(t, run.StartedAt.IsZero())

	// Verify score comparisons have expected metrics
	metricNames := make(map[string]bool)
	for _, sc := range run.ScoreComparison {
		metricNames[sc.MetricName] = true
		assert.NotZero(t, sc.BaselineValue)
	}
	assert.True(t, metricNames["accuracy"])
	assert.True(t, metricNames["relevance"])
	assert.True(t, metricNames["latency_ms"])

	// Severity should be set
	assert.NotEmpty(t, string(run.OverallSeverity))
}

func TestPromptCIListRuns(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()

	runs, err := svc.ListRuns(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, runs)
}

func TestPromptCICreateGateConfig(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid config", func(t *testing.T) {
		input := &domain.PromptCIGateConfigInput{
			Name:       "Production Gate",
			BaselineID: uuid.New(),
			Thresholds: map[string]domain.MetricThreshold{
				"accuracy": {MaxRegressionPct: 5.0, MinAbsoluteValue: 0.85, Direction: "higher_better"},
			},
			BlockOnSeverity: domain.RegressionSeverityMajor,
			ConfidenceLevel: 0.95,
		}

		config, err := svc.CreateGateConfig(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "Production Gate", config.Name)
		assert.Equal(t, domain.RegressionSeverityMajor, config.BlockOnSeverity)
		assert.Equal(t, 0.95, config.ConfidenceLevel)
		assert.True(t, config.Enabled)
		assert.NotEqual(t, uuid.Nil, config.ID)
	})

	t.Run("empty name fails", func(t *testing.T) {
		input := &domain.PromptCIGateConfigInput{
			Name:       "",
			BaselineID: uuid.New(),
			Thresholds: map[string]domain.MetricThreshold{"x": {}},
		}
		_, err := svc.CreateGateConfig(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("empty thresholds fails", func(t *testing.T) {
		input := &domain.PromptCIGateConfigInput{
			Name:       "test",
			BaselineID: uuid.New(),
			Thresholds: map[string]domain.MetricThreshold{},
		}
		_, err := svc.CreateGateConfig(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "threshold")
	})

	t.Run("defaults applied for missing optional fields", func(t *testing.T) {
		input := &domain.PromptCIGateConfigInput{
			Name:       "DefaultGate",
			BaselineID: uuid.New(),
			Thresholds: map[string]domain.MetricThreshold{"latency": {MaxRegressionPct: 10}},
		}
		config, err := svc.CreateGateConfig(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, domain.RegressionSeverityMajor, config.BlockOnSeverity)
		assert.Equal(t, 0.95, config.ConfidenceLevel)
	})

	t.Run("invalid confidence level gets default", func(t *testing.T) {
		input := &domain.PromptCIGateConfigInput{
			Name:            "BadConfidence",
			BaselineID:      uuid.New(),
			Thresholds:      map[string]domain.MetricThreshold{"x": {}},
			ConfidenceLevel: 1.5,
		}
		config, err := svc.CreateGateConfig(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, 0.95, config.ConfidenceLevel)
	})
}

func TestPromptCIEvaluateGate(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("empty scores fails", func(t *testing.T) {
		input := &domain.PromptCIGateEvalInput{
			GateConfigID: uuid.New(),
			Branch:       "main",
			CommitSHA:    "abc123",
			Scores:       map[string]float64{},
		}
		_, err := svc.EvaluateGate(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scores are required")
	})

	t.Run("passing scores return passed", func(t *testing.T) {
		input := &domain.PromptCIGateEvalInput{
			GateConfigID: uuid.New(),
			Branch:       "main",
			CommitSHA:    "abc123",
			Scores:       map[string]float64{"accuracy": 0.95, "relevance": 0.90},
		}
		result, err := svc.EvaluateGate(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEqual(t, uuid.Nil, result.RunID)
		assert.NotEmpty(t, result.Summary)
		assert.False(t, result.EvaluatedAt.IsZero())
	})
}

func TestPromptCIUpdateGateConfig(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()
	projectID := uuid.New()
	configID := uuid.New()

	t.Run("partial update", func(t *testing.T) {
		newName := "Updated Gate"
		input := &domain.PromptCIGateConfigUpdate{
			Name: &newName,
		}
		config, err := svc.UpdateGateConfig(ctx, projectID, configID, input)
		require.NoError(t, err)
		assert.Equal(t, configID, config.ID)
		assert.Equal(t, projectID, config.ProjectID)
		assert.Equal(t, "Updated Gate", config.Name)
		assert.False(t, config.UpdatedAt.IsZero())
	})

	t.Run("toggle enabled", func(t *testing.T) {
		enabled := false
		input := &domain.PromptCIGateConfigUpdate{
			Enabled: &enabled,
		}
		config, err := svc.UpdateGateConfig(ctx, projectID, configID, input)
		require.NoError(t, err)
		assert.False(t, config.Enabled)
	})
}

func TestPromptCIRecordRegressionEvent(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("records failed gate with PR", func(t *testing.T) {
		prNum := 42
		result := &domain.PromptCIGateResult{
			RunID:           uuid.New(),
			GateConfigID:    uuid.New(),
			Passed:          false,
			OverallSeverity: domain.RegressionSeverityCritical,
			MetricResults: []domain.MetricGateResult{
				{MetricName: "accuracy", ActualChangePct: -12.5, Passed: false},
			},
		}

		event, err := svc.RecordRegressionEvent(ctx, projectID, result, "feature/new", "deadbeef", &prNum)
		require.NoError(t, err)
		assert.False(t, event.Passed)
		assert.True(t, event.BlockedPR)
		assert.Equal(t, &prNum, event.PRNumber)
		assert.Equal(t, domain.RegressionSeverityCritical, event.Severity)
		assert.Contains(t, event.MetricDeltas, "accuracy")
		assert.Equal(t, -12.5, event.MetricDeltas["accuracy"])
	})

	t.Run("passing result without PR", func(t *testing.T) {
		result := &domain.PromptCIGateResult{
			RunID:           uuid.New(),
			GateConfigID:    uuid.New(),
			Passed:          true,
			OverallSeverity: domain.RegressionSeverityNone,
			MetricResults:   []domain.MetricGateResult{},
		}
		event, err := svc.RecordRegressionEvent(ctx, projectID, result, "main", "abc123", nil)
		require.NoError(t, err)
		assert.True(t, event.Passed)
		assert.False(t, event.BlockedPR)
		assert.Nil(t, event.PRNumber)
	})
}

func TestPromptCIGetDashboardStats(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)
	ctx := context.Background()

	stats, err := svc.GetDashboardStats(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, 100.0, stats.PassRate)
}

func TestPromptCICompareScore(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)

	t.Run("quality metric regression", func(t *testing.T) {
		// 0.92 -> 0.75 = significant regression
		sc := svc.compareScore("accuracy", 0.92, 0.75)
		assert.True(t, sc.IsRegression)
		assert.Equal(t, domain.RegressionSeverityCritical, sc.Severity)
		assert.InDelta(t, -0.17, sc.Delta, 0.001)
	})

	t.Run("quality metric minor regression", func(t *testing.T) {
		// 0.92 -> 0.90 = minor regression (delta ~2.2%)
		sc := svc.compareScore("accuracy", 0.92, 0.90)
		assert.True(t, sc.IsRegression)
		assert.Equal(t, domain.RegressionSeverityMinor, sc.Severity)
	})

	t.Run("quality metric no regression", func(t *testing.T) {
		// 0.92 -> 0.93 = improvement
		sc := svc.compareScore("accuracy", 0.92, 0.93)
		assert.False(t, sc.IsRegression)
		assert.Equal(t, domain.RegressionSeverityNone, sc.Severity)
	})

	t.Run("latency regression", func(t *testing.T) {
		// 450ms -> 600ms = significant latency regression
		sc := svc.compareScore("latency_ms", 450.0, 600.0)
		assert.True(t, sc.IsRegression)
		assert.True(t, sc.Severity != domain.RegressionSeverityNone)
	})

	t.Run("latency improvement", func(t *testing.T) {
		// 450ms -> 400ms = improvement (latency decreased)
		sc := svc.compareScore("latency_ms", 450.0, 400.0)
		assert.False(t, sc.IsRegression)
	})

	t.Run("latency within threshold", func(t *testing.T) {
		// 450ms -> 460ms = within 5% threshold
		sc := svc.compareScore("latency_ms", 450.0, 460.0)
		assert.False(t, sc.IsRegression)
	})
}

func TestPromptCIClassifyOverallSeverity(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptCIService(logger)

	t.Run("no regressions", func(t *testing.T) {
		comps := []domain.ScoreComparison{
			{Severity: domain.RegressionSeverityNone},
			{Severity: domain.RegressionSeverityNone},
		}
		assert.Equal(t, domain.RegressionSeverityNone, svc.classifyOverallSeverity(comps))
	})

	t.Run("worst wins", func(t *testing.T) {
		comps := []domain.ScoreComparison{
			{Severity: domain.RegressionSeverityMinor},
			{Severity: domain.RegressionSeverityCritical},
			{Severity: domain.RegressionSeverityMajor},
		}
		assert.Equal(t, domain.RegressionSeverityCritical, svc.classifyOverallSeverity(comps))
	})

	t.Run("empty comparisons", func(t *testing.T) {
		assert.Equal(t, domain.RegressionSeverityNone, svc.classifyOverallSeverity(nil))
	})
}
