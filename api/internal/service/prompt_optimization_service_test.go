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

func TestPromptOptStartOptimization(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptOptimizationService(logger)
	ctx := context.Background()
	projectID := uuid.New()
	promptID := uuid.New()

	opt, err := svc.StartOptimization(ctx, projectID, promptID, 1)
	require.NoError(t, err)
	assert.Equal(t, projectID, opt.ProjectID)
	assert.Equal(t, promptID, opt.PromptID)
	assert.Equal(t, 1, opt.PromptVersion)
	assert.Equal(t, domain.OptimizationStatusAnalyzing, opt.Status)
	// Should have failure patterns
	assert.NotEmpty(t, opt.FailurePatterns)
	assert.Len(t, opt.FailurePatterns, 3)
	// Should have generated variants
	assert.NotEmpty(t, opt.Variants)
	assert.Len(t, opt.Variants, 2)
	for _, v := range opt.Variants {
		assert.Equal(t, domain.VariantStatusCandidate, v.Status)
		assert.NotEmpty(t, v.Content)
		assert.NotEmpty(t, v.Rationale)
	}
	assert.NotNil(t, opt.StartedAt)
}

func TestPromptOptApproveVariant(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptOptimizationService(logger)
	ctx := context.Background()

	err := svc.ApproveVariant(ctx, uuid.New())
	assert.NoError(t, err)
}

func TestPromptOptRejectVariant(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptOptimizationService(logger)
	ctx := context.Background()

	err := svc.RejectVariant(ctx, uuid.New())
	assert.NoError(t, err)
}

func TestPromptOptGetConfig(t *testing.T) {
	logger := zap.NewNop()
	svc := NewPromptOptimizationService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	config, err := svc.GetConfig(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, projectID, config.ProjectID)
	assert.True(t, config.Enabled)
	assert.Greater(t, config.MinSamplesForAnalysis, 0)
	assert.Greater(t, config.MinSamplesForPromotion, 0)
	assert.Greater(t, config.PValueThreshold, 0.0)
	assert.Greater(t, config.MaxVariantsPerRound, 0)
}
