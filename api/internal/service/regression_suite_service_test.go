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

func TestRegressionCreateGoldenDataset(t *testing.T) {
	logger := zap.NewNop()
	svc := NewRegressionSuiteService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	input := domain.GoldenDataset{
		Name:     "Bug Fix Dataset",
		Category: domain.GoldenDatasetCategoryBugFix,
		Language: "python",
	}

	dataset, err := svc.CreateGoldenDataset(ctx, projectID, input)
	require.NoError(t, err)
	assert.Equal(t, "Bug Fix Dataset", dataset.Name)
	assert.Equal(t, projectID, dataset.ProjectID)
	assert.Equal(t, domain.GoldenDatasetCategoryBugFix, dataset.Category)
	// Bug fix category has 2 templates
	assert.Len(t, dataset.Items, 2)
	assert.Equal(t, 2, dataset.ItemCount)
	for _, item := range dataset.Items {
		assert.NotEmpty(t, item.Input)
		assert.NotEmpty(t, item.ExpectedBehavior)
		assert.Contains(t, item.Tags, "python")
	}

	// Empty name should fail
	_, err = svc.CreateGoldenDataset(ctx, projectID, domain.GoldenDataset{Name: "", Category: domain.GoldenDatasetCategoryBugFix})
	assert.Error(t, err)

	// Invalid category should fail
	_, err = svc.CreateGoldenDataset(ctx, projectID, domain.GoldenDataset{Name: "X", Category: domain.GoldenDatasetCategory("invalid")})
	assert.Error(t, err)
}

func TestRegressionRunRegression(t *testing.T) {
	logger := zap.NewNop()
	svc := NewRegressionSuiteService(logger)
	ctx := context.Background()
	projectID := uuid.New()
	suiteID := uuid.New()

	run, err := svc.RunRegression(ctx, projectID, suiteID, "model=gpt-4")
	require.NoError(t, err)
	assert.Equal(t, projectID, run.ProjectID)
	assert.Equal(t, suiteID, run.SuiteID)
	// Should have results (5 simulated items since GetGoldenDataset returns empty)
	assert.Len(t, run.Results, 5)
	assert.Equal(t, 5, run.TotalTests)
	assert.Equal(t, run.Passed+run.Failed, run.TotalTests)
	// Pass rate should be between 0 and 100
	assert.GreaterOrEqual(t, run.PassRate, 0.0)
	assert.LessOrEqual(t, run.PassRate, 100.0)
	// Status should be passed or failed
	assert.True(t, run.Status == domain.RegressionRunStatusPassed || run.Status == domain.RegressionRunStatusFailed)
	assert.NotNil(t, run.BaselineComparison)
	assert.NotNil(t, run.StartedAt)
	assert.NotNil(t, run.CompletedAt)

	// Empty agent config should fail
	_, err = svc.RunRegression(ctx, projectID, suiteID, "")
	assert.Error(t, err)
}

func TestRegressionListGoldenDatasets(t *testing.T) {
	logger := zap.NewNop()
	svc := NewRegressionSuiteService(logger)
	ctx := context.Background()

	datasets, err := svc.ListGoldenDatasets(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, datasets)
}
