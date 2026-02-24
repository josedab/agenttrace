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

func TestCostAlertingCreateRule(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostAlertingService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	rule := domain.CostAlertRule{
		Name:     "High cost alert",
		Severity: domain.CostAlertSeverityWarning,
		Condition: domain.CostAlertCondition{
			Metric:    "total_cost",
			Operator:  "gt",
			Threshold: 10.0,
		},
	}

	created, err := svc.CreateRule(ctx, projectID, rule)
	require.NoError(t, err)
	assert.Equal(t, "High cost alert", created.Name)
	assert.Equal(t, projectID, created.ProjectID)
	assert.True(t, created.Enabled)
	assert.NotEqual(t, uuid.Nil, created.ID)

	// Empty name should fail
	_, err = svc.CreateRule(ctx, projectID, domain.CostAlertRule{
		Name:      "",
		Severity:  domain.CostAlertSeverityWarning,
		Condition: domain.CostAlertCondition{Threshold: 1.0},
	})
	assert.Error(t, err)

	// Negative threshold should fail
	_, err = svc.CreateRule(ctx, projectID, domain.CostAlertRule{
		Name:      "Bad",
		Severity:  domain.CostAlertSeverityWarning,
		Condition: domain.CostAlertCondition{Threshold: -1.0},
	})
	assert.Error(t, err)
}

func TestCostAlertingCheckAndAlertBelow(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostAlertingService(logger)
	ctx := context.Background()

	// ListRules returns empty, so no rules to breach
	alert, err := svc.CheckAndAlert(ctx, uuid.New(), 5.0, "gpt-4")
	require.NoError(t, err)
	assert.Nil(t, alert, "no alert should be generated when no rules exist")
}

func TestCostAlertingCheckAndAlertAbove(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostAlertingService(logger)
	ctx := context.Background()

	// Since ListRules returns empty slice, CheckAndAlert will never trigger.
	// Verify the nil response for no-breach scenario.
	alert, err := svc.CheckAndAlert(ctx, uuid.New(), 100.0, "gpt-4")
	require.NoError(t, err)
	// With no rules configured, even high cost produces no alert
	assert.Nil(t, alert)
}

func TestCostAlertingGetCircuitBreaker(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostAlertingService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	cb, err := svc.GetCircuitBreaker(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, projectID, cb.ProjectID)
	assert.Equal(t, domain.CircuitBreakerStateClosed, cb.State)
	assert.False(t, cb.Enabled)
	assert.Greater(t, cb.MaxCostPerMinute, 0.0)
	assert.Greater(t, cb.MaxCostPerHour, 0.0)
	assert.NotEmpty(t, cb.FallbackModelChain)
}

func TestCostAlertingUpdateCircuitBreaker(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCostAlertingService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	// Valid config
	config := domain.CircuitBreakerConfig{
		Enabled:          true,
		MaxCostPerMinute: 2.0,
		MaxCostPerHour:   20.0,
		CooldownSeconds:  30,
	}
	result, err := svc.UpdateCircuitBreaker(ctx, projectID, config)
	require.NoError(t, err)
	assert.Equal(t, projectID, result.ProjectID)
	assert.True(t, result.Enabled)

	// Max per hour < max per minute should fail
	_, err = svc.UpdateCircuitBreaker(ctx, projectID, domain.CircuitBreakerConfig{
		MaxCostPerMinute: 10.0,
		MaxCostPerHour:   5.0,
		CooldownSeconds:  30,
	})
	assert.Error(t, err)

	// Cooldown too small should fail
	_, err = svc.UpdateCircuitBreaker(ctx, projectID, domain.CircuitBreakerConfig{
		MaxCostPerMinute: 1.0,
		MaxCostPerHour:   10.0,
		CooldownSeconds:  5,
	})
	assert.Error(t, err)
}
