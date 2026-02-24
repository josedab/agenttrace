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

func TestFederatedAggregationGetDashboard(t *testing.T) {
	logger := zap.NewNop()
	svc := NewFederatedAggregationService(logger)
	ctx := context.Background()

	dashboard, err := svc.GetDashboard(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, dashboard)

	assert.NotNil(t, dashboard.Instances)
	assert.Empty(t, dashboard.Instances)
	assert.NotNil(t, dashboard.Benchmarks)
	assert.Empty(t, dashboard.Benchmarks)
	assert.NotNil(t, dashboard.Insights)
	assert.Empty(t, dashboard.Insights)
	assert.Equal(t, 0, dashboard.ParticipantCount)
}

func TestFederatedAggregationRegisterInstance(t *testing.T) {
	logger := zap.NewNop()
	svc := NewFederatedAggregationService(logger)
	ctx := context.Background()

	input := domain.FederatedInstanceInput{
		Name:         "Production Instance",
		Endpoint:     "https://prod.example.com/api",
		APIKey:       "test-api-key-123",
		PrivacyLevel: domain.PrivacyLevelAggregatedOnly,
	}

	instance, err := svc.RegisterInstance(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, instance)

	assert.NotEqual(t, uuid.Nil, instance.ID)
	assert.Equal(t, "Production Instance", instance.Name)
	assert.Equal(t, "https://prod.example.com/api", instance.Endpoint)
	assert.Equal(t, "test-api-key-123", instance.APIKey)
	assert.Equal(t, domain.PrivacyLevelAggregatedOnly, instance.PrivacyLevel)
	assert.Equal(t, "active", instance.Status)
	assert.Equal(t, int64(0), instance.MetricsCount)
	assert.False(t, instance.CreatedAt.IsZero())
}

func TestFederatedAggregationListInstances(t *testing.T) {
	logger := zap.NewNop()
	svc := NewFederatedAggregationService(logger)
	ctx := context.Background()

	instances, err := svc.ListInstances(ctx, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, instances)
	assert.Empty(t, instances)
}

func TestFederatedAggregationGetBenchmarks(t *testing.T) {
	logger := zap.NewNop()
	svc := NewFederatedAggregationService(logger)
	ctx := context.Background()

	benchmarks, err := svc.GetBenchmarks(ctx, uuid.New(), domain.FederatedMetricTypeLatency)
	require.NoError(t, err)
	assert.NotNil(t, benchmarks)
	assert.Empty(t, benchmarks)
}

func TestFederatedAggregationGetInsights(t *testing.T) {
	logger := zap.NewNop()
	svc := NewFederatedAggregationService(logger)
	ctx := context.Background()

	insights, err := svc.GetInsights(ctx, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, insights)
	assert.Empty(t, insights)
}
