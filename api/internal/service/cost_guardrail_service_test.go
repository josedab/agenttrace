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

func TestCostGuardrailGetDashboard(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	dashboard, err := svc.GetDashboard(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, dashboard)

	assert.Equal(t, 0, dashboard.ActivePolicies)
	assert.Empty(t, dashboard.RecentViolations)
	assert.Empty(t, dashboard.TopSpenders)
	assert.NotNil(t, dashboard.Forecast)
	assert.Empty(t, dashboard.Forecast.Recommendations)
	assert.Empty(t, dashboard.Forecast.ByModel)
}

func TestCostGuardrailCreatePolicy(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	input := domain.CostGuardrailPolicyInput{
		Name:             "Monthly Budget",
		Type:             domain.GuardrailPolicyTypePerProject,
		Action:           domain.GuardrailActionWarn,
		BudgetLimit:      500.0,
		BudgetPeriod:     "monthly",
		ThresholdPercent: 80.0,
	}

	policy, err := svc.CreatePolicy(ctx, projectID, input)
	require.NoError(t, err)
	require.NotNil(t, policy)

	assert.NotEqual(t, uuid.Nil, policy.ID)
	assert.Equal(t, projectID, policy.ProjectID)
	assert.Equal(t, "Monthly Budget", policy.Name)
	assert.Equal(t, domain.GuardrailPolicyTypePerProject, policy.Type)
	assert.Equal(t, domain.GuardrailActionWarn, policy.Action)
	assert.True(t, policy.Enabled)
	assert.Equal(t, 500.0, policy.BudgetLimit)
	assert.Equal(t, "monthly", policy.BudgetPeriod)
	assert.Equal(t, 80.0, policy.ThresholdPercent)
	assert.Equal(t, 30, policy.CooldownMinutes)
	assert.False(t, policy.CreatedAt.IsZero())
}

func TestCostGuardrailListPolicies(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	policies, err := svc.ListPolicies(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, policies)
}

func TestCostGuardrailCheckBudgetAllowed(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	// Small cost against 1000 budget with 0 current spend -> allowed
	result, err := svc.CheckBudget(ctx, uuid.New(), uuid.New(), 10.0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Allowed)
	assert.Equal(t, domain.GuardrailActionNotify, result.Action)
	assert.Equal(t, "within budget", result.Reason)
}

func TestCostGuardrailCheckBudgetWarn(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	// Cost of 850 against 1000 budget with 0 spend = 85% usage -> warn
	result, err := svc.CheckBudget(ctx, uuid.New(), uuid.New(), 850.0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Allowed)
	assert.Equal(t, domain.GuardrailActionWarn, result.Action)
	assert.Contains(t, result.Reason, "approaching limit")
}

func TestCostGuardrailCheckBudgetThrottle(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	// Cost of 950 against 1000 budget with 0 spend = 95% usage -> throttle
	result, err := svc.CheckBudget(ctx, uuid.New(), uuid.New(), 950.0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Allowed)
	assert.Equal(t, domain.GuardrailActionThrottle, result.Action)
	assert.Contains(t, result.Reason, "throttling")
	assert.Equal(t, "gpt-3.5-turbo", result.SuggestedModel)
}

func TestCostGuardrailCheckBudgetPause(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	// Cost of 1100 against 1000 budget with 0 spend = 110% usage -> pause
	result, err := svc.CheckBudget(ctx, uuid.New(), uuid.New(), 1100.0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Allowed)
	assert.Equal(t, domain.GuardrailActionPause, result.Action)
	assert.Contains(t, result.Reason, "budget exhausted")
}

func TestCostGuardrailGetForecast(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	forecast, err := svc.GetForecast(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, projectID, forecast.ProjectID)
	assert.Equal(t, float64(0), forecast.CurrentSpend)
	assert.Equal(t, 1000.0, forecast.BudgetRemaining)
	assert.Equal(t, 30, forecast.DaysUntilBudgetExhausted)
	assert.NotEmpty(t, forecast.Recommendations)
	assert.Empty(t, forecast.ByModel)
}

func TestCostGuardrailListViolations(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostGuardrailService(logger)
	ctx := context.Background()

	violations, err := svc.ListViolations(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, violations)
}
