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

func TestNewPredictiveCostService(t *testing.T) {
	svc := NewPredictiveCostService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestPredictiveCostService_PredictCost(t *testing.T) {
	svc := NewPredictiveCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("returns valid prediction", func(t *testing.T) {
		pred, err := svc.PredictCost(ctx, projectID, &domain.PredictionInput{
			TaskDescription: "Implement a complex authentication middleware with JWT validation",
		})
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, pred.ID)
		assert.Equal(t, projectID, pred.ProjectID)
		assert.Greater(t, pred.PredictedCost, 0.0)
		assert.Greater(t, pred.PredictedLatencyMs, int64(0))
		assert.Greater(t, pred.PredictedQuality, 0.0)
		assert.LessOrEqual(t, pred.PredictedQuality, 100.0)
		assert.GreaterOrEqual(t, pred.ConfidenceLevel, 0.7)
		assert.LessOrEqual(t, pred.ConfidenceLevel, 1.0)
		assert.NotEmpty(t, pred.RecommendedModel)
		assert.Greater(t, pred.PredictedTokens, 0)
	})

	t.Run("budget status within", func(t *testing.T) {
		budget := 100.0
		pred, err := svc.PredictCost(ctx, projectID, &domain.PredictionInput{
			TaskDescription: "simple format task",
			MaxBudget:       &budget,
		})
		require.NoError(t, err)
		assert.Equal(t, "within", pred.BudgetStatus)
	})

	t.Run("budget status exceeded", func(t *testing.T) {
		budget := 0.0001
		pred, err := svc.PredictCost(ctx, projectID, &domain.PredictionInput{
			TaskDescription: "Implement a complex analysis system",
			MaxBudget:       &budget,
		})
		require.NoError(t, err)
		assert.Equal(t, "exceeded", pred.BudgetStatus)
	})

	t.Run("uses specified model", func(t *testing.T) {
		pred, err := svc.PredictCost(ctx, projectID, &domain.PredictionInput{
			TaskDescription: "summarize this text",
			Model:           "gpt-4o-mini",
		})
		require.NoError(t, err)
		assert.Equal(t, "gpt-4o-mini", pred.RecommendedModel)
	})
}

func TestPredictiveCostService_ListPredictions(t *testing.T) {
	svc := NewPredictiveCostService(zap.NewNop())
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	_, _ = svc.PredictCost(ctx, projectA, &domain.PredictionInput{TaskDescription: "task 1"})
	_, _ = svc.PredictCost(ctx, projectA, &domain.PredictionInput{TaskDescription: "task 2"})
	_, _ = svc.PredictCost(ctx, projectB, &domain.PredictionInput{TaskDescription: "task 3"})

	t.Run("filters by project", func(t *testing.T) {
		predsA, err := svc.ListPredictions(ctx, projectA)
		require.NoError(t, err)
		assert.Len(t, predsA, 2)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		preds, err := svc.ListPredictions(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, preds)
	})
}

func TestPredictiveCostService_ApprovalLifecycle(t *testing.T) {
	svc := NewPredictiveCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	pred, _ := svc.PredictCost(ctx, projectID, &domain.PredictionInput{TaskDescription: "build auth"})

	t.Run("request approval", func(t *testing.T) {
		approval, err := svc.RequestApproval(ctx, projectID, pred.ID)
		require.NoError(t, err)
		assert.Equal(t, "pending", approval.Status)
		assert.Equal(t, pred.ID, approval.PredictionID)
		assert.Nil(t, approval.DecidedAt)

		t.Run("approve", func(t *testing.T) {
			decided, err := svc.DecideApproval(ctx, approval.ID, &domain.ApprovalDecisionInput{
				Status: "approved",
				Note:   "looks good",
			})
			require.NoError(t, err)
			assert.Equal(t, "approved", decided.Status)
			assert.NotNil(t, decided.DecidedAt)
			assert.Equal(t, "looks good", decided.Note)
		})

		t.Run("cannot decide twice", func(t *testing.T) {
			_, err := svc.DecideApproval(ctx, approval.ID, &domain.ApprovalDecisionInput{Status: "rejected"})
			assert.Error(t, err)
		})
	})

	t.Run("request approval for unknown prediction", func(t *testing.T) {
		_, err := svc.RequestApproval(ctx, projectID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("decide unknown approval", func(t *testing.T) {
		_, err := svc.DecideApproval(ctx, uuid.New(), &domain.ApprovalDecisionInput{Status: "approved"})
		assert.Error(t, err)
	})
}
