package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRCAService(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestRCAService_AnalyzeTrace(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New()

	report, err := svc.AnalyzeTrace(ctx, projectID, traceID)
	require.NoError(t, err)

	t.Run("report has valid fields", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, report.ID)
		assert.Equal(t, projectID, report.ProjectID)
		assert.Equal(t, traceID, report.TraceID)
	})

	t.Run("has category and confidence", func(t *testing.T) {
		assert.NotEmpty(t, string(report.PrimaryCategory))
		assert.GreaterOrEqual(t, report.Confidence, 0.75)
		assert.LessOrEqual(t, report.Confidence, 1.0)
	})

	t.Run("has summary and analysis", func(t *testing.T) {
		assert.NotEmpty(t, report.Summary)
		assert.NotEmpty(t, report.DetailedAnalysis)
	})

	t.Run("has contributing factors", func(t *testing.T) {
		assert.GreaterOrEqual(t, len(report.ContributingFactors), 2)
		for _, f := range report.ContributingFactors {
			assert.NotEmpty(t, string(f.Category))
			assert.NotEmpty(t, f.Description)
			assert.Greater(t, f.Impact, 0.0)
		}
	})

	t.Run("has remediations", func(t *testing.T) {
		assert.Greater(t, len(report.Remediations), 0)
		for _, r := range report.Remediations {
			assert.NotEmpty(t, r.Action)
			assert.NotEmpty(t, r.Description)
			assert.Greater(t, r.Priority, 0)
		}
	})

	t.Run("has similar incidents", func(t *testing.T) {
		assert.Len(t, report.SimilarIncidents, 2)
	})
}

func TestRCAService_GetReport(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()

	t.Run("existing report", func(t *testing.T) {
		report, _ := svc.AnalyzeTrace(ctx, uuid.New(), uuid.New())
		got, err := svc.GetReport(ctx, report.ID)
		require.NoError(t, err)
		assert.Equal(t, report.ID, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetReport(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestRCAService_ListReports(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	_, _ = svc.AnalyzeTrace(ctx, projectA, uuid.New())
	_, _ = svc.AnalyzeTrace(ctx, projectA, uuid.New())
	_, _ = svc.AnalyzeTrace(ctx, projectB, uuid.New())

	t.Run("filters by project", func(t *testing.T) {
		reportsA, err := svc.ListReports(ctx, projectA)
		require.NoError(t, err)
		assert.Len(t, reportsA, 2)

		reportsB, err := svc.ListReports(ctx, projectB)
		require.NoError(t, err)
		assert.Len(t, reportsB, 1)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		reports, err := svc.ListReports(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, reports)
	})
}
