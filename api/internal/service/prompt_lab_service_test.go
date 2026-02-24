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

func TestNewPromptLabService(t *testing.T) {
	svc := NewPromptLabService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestPromptLabService_ExperimentLifecycle(t *testing.T) {
	svc := NewPromptLabService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("creates experiment with variants", func(t *testing.T) {
		input := &domain.PromptExperimentInput{
			Name:       "test-experiment",
			PromptName: "system-prompt",
			Variants: []domain.PromptVariantInput{
				{Name: "control", PromptContent: "You are a helpful assistant", IsControl: true, TrafficWeight: 0.5},
				{Name: "variant-a", PromptContent: "You are a coding expert", TrafficWeight: 0.5},
			},
		}

		exp, err := svc.CreateExperiment(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "test-experiment", exp.Name)
		assert.Equal(t, domain.PromptExpStatusDraft, exp.Status)
		assert.Len(t, exp.Variants, 2)

		totalWeight := 0.0
		for _, v := range exp.Variants {
			totalWeight += v.TrafficWeight
		}
		assert.InDelta(t, 1.0, totalWeight, 0.01)
	})

	t.Run("starts experiment", func(t *testing.T) {
		input := &domain.PromptExperimentInput{
			Name: "start-test", PromptName: "test",
			Variants: []domain.PromptVariantInput{
				{Name: "a", PromptContent: "a"},
				{Name: "b", PromptContent: "b"},
			},
		}
		exp, _ := svc.CreateExperiment(ctx, projectID, input)

		started, err := svc.StartExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PromptExpStatusRunning, started.Status)
	})

	t.Run("completes experiment", func(t *testing.T) {
		input := &domain.PromptExperimentInput{
			Name: "complete-test", PromptName: "test",
			Variants: []domain.PromptVariantInput{
				{Name: "a", PromptContent: "a"},
				{Name: "b", PromptContent: "b"},
			},
		}
		exp, _ := svc.CreateExperiment(ctx, projectID, input)

		completed, err := svc.CompleteExperiment(ctx, exp.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PromptExpStatusCompleted, completed.Status)
		assert.NotNil(t, completed.CompletedAt)
	})

	t.Run("lists experiments for project", func(t *testing.T) {
		exps, err := svc.ListExperiments(ctx, projectID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(exps), 3)
	})

	t.Run("returns error for non-existent experiment", func(t *testing.T) {
		_, err := svc.GetExperiment(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestPromptLabService_Suggestions(t *testing.T) {
	svc := NewPromptLabService(zap.NewNop())
	ctx := context.Background()

	suggestions, err := svc.GetOptimizationSuggestions(ctx, uuid.New(), "test-prompt")
	require.NoError(t, err)
	assert.Greater(t, len(suggestions), 0)

	for _, s := range suggestions {
		assert.NotEmpty(t, s.Technique)
		assert.Greater(t, s.SavingsPercent, 0.0)
		assert.Greater(t, s.Confidence, 0.0)
	}
}
