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
