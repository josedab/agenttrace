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

func TestFederatedAggregationDashboard(t *testing.T) {
	svc := NewFederatedAggregationService(zap.NewNop())
	ctx := context.Background()
	instanceID := uuid.New()

	dashboard, err := svc.GetFederatedAnalyticsDashboard(ctx, instanceID)
	require.NoError(t, err)
	require.NotNil(t, dashboard)

	// Verify mesh status
	assert.Greater(t, dashboard.MeshStatus.TotalInstances, 0)
	assert.Greater(t, dashboard.MeshStatus.ActiveInstances, 0)

	// Verify privacy budget
	assert.Equal(t, instanceID, dashboard.PrivacyBudget.InstanceID)
	assert.Greater(t, dashboard.PrivacyBudget.TotalEpsilon, 0.0)
	assert.GreaterOrEqual(t, dashboard.PrivacyBudget.RemainingEpsilon, 0.0)
	assert.LessOrEqual(t, dashboard.PrivacyBudget.UsedEpsilon, dashboard.PrivacyBudget.TotalEpsilon)

	// Verify comparisons
	assert.NotEmpty(t, dashboard.Comparisons)
	for _, c := range dashboard.Comparisons {
		assert.NotEmpty(t, c.MetricName)
		assert.Greater(t, c.ParticipantCount, 0)
		assert.NotEmpty(t, c.Trend)
		// P25 <= Median <= P75 <= P90
		assert.LessOrEqual(t, c.IndustryP25, c.IndustryMedian)
		assert.LessOrEqual(t, c.IndustryMedian, c.IndustryP75)
		assert.LessOrEqual(t, c.IndustryP75, c.IndustryP90)
	}
}

func TestFederatedAggregationPrivacyQuery(t *testing.T) {
	svc := NewFederatedAggregationService(zap.NewNop())
	ctx := context.Background()
	instanceID := uuid.New()

	t.Run("valid query", func(t *testing.T) {
		input := &domain.FederatedQueryInput{
			Metrics: []string{"avg_latency_ms", "error_rate"},
		}
		results, err := svc.RunPrivacyPreservingQuery(ctx, instanceID, input)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
}

func TestFederatedInsightGeneration(t *testing.T) {
	svc := NewFederatedAggregationService(zap.NewNop())
	ctx := context.Background()
	instanceID := uuid.New()

	dashboard, err := svc.GetFederatedAnalyticsDashboard(ctx, instanceID)
	require.NoError(t, err)

	// Insights should be generated based on percentile position
	for _, insight := range dashboard.Insights {
		assert.NotEqual(t, uuid.Nil, insight.ID)
		assert.NotEmpty(t, insight.Category)
		assert.NotEmpty(t, insight.Title)
		assert.NotEmpty(t, insight.Impact)
		assert.NotEmpty(t, insight.Recommendation)
	}
}
