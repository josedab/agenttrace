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

func TestCloudOnboardingCreateOnboarding(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCloudOnboardingService(logger)
	ctx := context.Background()
	tenantID := uuid.New()

	onboarding, err := svc.CreateOnboarding(ctx, tenantID)
	require.NoError(t, err)
	assert.Equal(t, tenantID, onboarding.TenantID)
	assert.Equal(t, domain.OnboardingStepProjectCreated, onboarding.CurrentStep)
	assert.Len(t, onboarding.Steps, 5)
	// First step should be completed
	assert.True(t, onboarding.Steps[0].Completed)
	assert.NotNil(t, onboarding.Steps[0].CompletedAt)
	// Remaining steps should not be completed
	for _, step := range onboarding.Steps[1:] {
		assert.False(t, step.Completed)
	}
}

func TestCloudOnboardingCompleteStep(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCloudOnboardingService(logger)
	ctx := context.Background()
	tenantID := uuid.New()

	err := svc.CompleteStep(ctx, tenantID, domain.OnboardingStepSDKInstalled)
	assert.NoError(t, err)

	// Invalid step should fail
	err = svc.CompleteStep(ctx, tenantID, domain.OnboardingStep("invalid_step"))
	assert.Error(t, err)
}

func TestCloudOnboardingGenerateQuickstart(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCloudOnboardingService(logger)
	ctx := context.Background()

	// Valid language
	config, err := svc.GenerateQuickstart(ctx, "python", "langchain", "sk-test-key", "proj-123", "")
	require.NoError(t, err)
	assert.Equal(t, "python", config.Language)
	assert.Equal(t, "langchain", config.Framework)
	assert.Equal(t, "https://cloud.agenttrace.io", config.Host)

	// Invalid language
	_, err = svc.GenerateQuickstart(ctx, "rust", "langchain", "sk-test-key", "proj-123", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported language")

	// Missing API key
	_, err = svc.GenerateQuickstart(ctx, "python", "langchain", "", "proj-123", "")
	assert.Error(t, err)
}

func TestCloudOnboardingCheckQuota(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCloudOnboardingService(logger)
	ctx := context.Background()
	tenantID := uuid.New()

	// Small quantity should be allowed (mock usage has plenty of room)
	allowed, err := svc.CheckQuota(ctx, tenantID, "traces", 100)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Very large quantity should exceed quota
	allowed, err = svc.CheckQuota(ctx, tenantID, "traces", 999999)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Invalid metric
	_, err = svc.CheckQuota(ctx, tenantID, "invalid_metric", 1)
	assert.Error(t, err)
}
