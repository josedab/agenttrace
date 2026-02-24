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

func TestSemanticTraceSearchSearch(t *testing.T) {
	logger := zap.NewNop()
	svc := NewSemanticTraceSearchService(logger)
	ctx := context.Background()

	query := domain.SemanticTraceSearchQuery{
		Query:     "error handling in authentication",
		ProjectID: uuid.New(),
		Limit:     10,
	}

	results, err := svc.Search(ctx, query)
	require.NoError(t, err)
	assert.NotNil(t, results)
	assert.Empty(t, results)
}

func TestSemanticTraceSearchGetClusters(t *testing.T) {
	logger := zap.NewNop()
	svc := NewSemanticTraceSearchService(logger)
	ctx := context.Background()

	clusters, err := svc.GetClusters(ctx, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, clusters)
	assert.Empty(t, clusters)
}

func TestSemanticTraceSearchGetAnomalyPatterns(t *testing.T) {
	logger := zap.NewNop()
	svc := NewSemanticTraceSearchService(logger)
	ctx := context.Background()

	patterns, err := svc.GetAnomalyPatterns(ctx, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, patterns)
	assert.Empty(t, patterns)
}

func TestSemanticTraceSearchGetDashboard(t *testing.T) {
	logger := zap.NewNop()
	svc := NewSemanticTraceSearchService(logger)
	ctx := context.Background()

	dashboard, err := svc.GetDashboard(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, dashboard)

	assert.NotNil(t, dashboard.RecentSearches)
	assert.Empty(t, dashboard.RecentSearches)
	assert.NotNil(t, dashboard.TopClusters)
	assert.Empty(t, dashboard.TopClusters)
	assert.NotNil(t, dashboard.AnomalyPatterns)
	assert.Empty(t, dashboard.AnomalyPatterns)
	assert.Equal(t, int64(0), dashboard.TotalIndexedTraces)
}

func TestSemanticTraceSearchIndexTrace(t *testing.T) {
	logger := zap.NewNop()
	svc := NewSemanticTraceSearchService(logger)
	ctx := context.Background()

	err := svc.IndexTrace(ctx, uuid.New())
	require.NoError(t, err)
}
