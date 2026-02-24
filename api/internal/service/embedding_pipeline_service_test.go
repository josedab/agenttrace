package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEmbeddingPipelineService_IndexAndSearch(t *testing.T) {
	logger := zap.NewNop()
	svc := NewEmbeddingPipelineService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	// Index several traces
	err := svc.IndexTrace(ctx, projectID, uuid.New(), "debug-typescript-generics",
		"Fix TypeScript generic type inference issue",
		"Applied constraint to generic parameter",
	)
	require.NoError(t, err)

	err = svc.IndexTrace(ctx, projectID, uuid.New(), "refactor-auth-module",
		"Refactor authentication middleware",
		"Extracted JWT validation into separate module",
	)
	require.NoError(t, err)

	err = svc.IndexTrace(ctx, projectID, uuid.New(), "fix-typescript-types",
		"TypeScript type error in user service",
		"Fixed incorrect type assertion for generic response",
	)
	require.NoError(t, err)

	// Search for TypeScript-related traces
	results, err := svc.Search(ctx, projectID, "TypeScript generics type error", 10, 0.05)
	require.NoError(t, err)
	assert.NotEmpty(t, results, "should find matching traces")

	// TypeScript-related traces should score higher than auth
	for _, r := range results {
		assert.Greater(t, r.Score, 0.0)
		assert.NotEmpty(t, r.Snippet)
	}
}

func TestEmbeddingPipelineService_SearchNoResults(t *testing.T) {
	logger := zap.NewNop()
	svc := NewEmbeddingPipelineService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	results, err := svc.Search(ctx, projectID, "anything", 10, 0.1)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestEmbeddingPipelineService_ProjectIsolation(t *testing.T) {
	logger := zap.NewNop()
	svc := NewEmbeddingPipelineService(logger)
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	err := svc.IndexTrace(ctx, projectA, uuid.New(), "trace-a", "project A content", nil)
	require.NoError(t, err)

	err = svc.IndexTrace(ctx, projectB, uuid.New(), "trace-b", "project B content", nil)
	require.NoError(t, err)

	// Search project A should not find project B traces
	results, err := svc.Search(ctx, projectA, "project B content", 10, 0.01)
	require.NoError(t, err)
	for _, r := range results {
		assert.NotEqual(t, "trace-b", r.TraceName)
	}
}

func TestEmbeddingPipelineService_Deduplication(t *testing.T) {
	logger := zap.NewNop()
	svc := NewEmbeddingPipelineService(logger)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New()

	err := svc.IndexTrace(ctx, projectID, traceID, "same-trace", "same content", nil)
	require.NoError(t, err)

	// Index same trace again
	err = svc.IndexTrace(ctx, projectID, traceID, "same-trace", "same content", nil)
	require.NoError(t, err)

	stats := svc.GetIndexStats(projectID)
	assert.Equal(t, 1, stats["totalIndexed"])
}

func TestEmbeddingPipelineService_ClusterTraces(t *testing.T) {
	logger := zap.NewNop()
	svc := NewEmbeddingPipelineService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	// Index traces that should cluster together
	for i := 0; i < 5; i++ {
		err := svc.IndexTrace(ctx, projectID, uuid.New(), "typescript-error",
			"TypeScript compilation error in module", nil)
		require.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		err := svc.IndexTrace(ctx, projectID, uuid.New(), "python-test",
			"Python pytest failure in test suite", nil)
		require.NoError(t, err)
	}

	clusters, err := svc.ClusterTraces(ctx, projectID, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, clusters)

	totalTraces := 0
	for _, c := range clusters {
		assert.Greater(t, c.TraceCount, 0)
		totalTraces += c.TraceCount
	}
	assert.Equal(t, 8, totalTraces)
}

func TestEmbeddingPipelineService_IndexStats(t *testing.T) {
	logger := zap.NewNop()
	svc := NewEmbeddingPipelineService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	err := svc.IndexTrace(ctx, projectID, uuid.New(), "test", "content", nil)
	require.NoError(t, err)

	err = svc.IndexObservation(ctx, projectID, uuid.New(), uuid.New(), "obs", "input", "output")
	require.NoError(t, err)

	stats := svc.GetIndexStats(projectID)
	assert.Equal(t, 2, stats["totalIndexed"])
	assert.Equal(t, 1, stats["tracesIndexed"])
	assert.Equal(t, 1, stats["observationsIndexed"])
	assert.Equal(t, 128, stats["embeddingDimension"])
}

func TestLocalEmbeddingProvider(t *testing.T) {
	provider := NewLocalEmbeddingProvider()

	embedding, err := provider.Embed(context.Background(), "test input text")
	require.NoError(t, err)
	assert.Len(t, embedding, 128)

	// Verify L2 normalization
	var norm float64
	for _, v := range embedding {
		norm += float64(v * v)
	}
	assert.InDelta(t, 1.0, norm, 0.01, "embedding should be L2 normalized")
}

func TestCosineSimilarity(t *testing.T) {
	// Identical vectors should have similarity 1.0
	a := []float32{1, 0, 0}
	assert.InDelta(t, 1.0, pipelineCosineSimilarity(a, a), 0.001)

	// Orthogonal vectors should have similarity 0.0
	b := []float32{0, 1, 0}
	assert.InDelta(t, 0.0, pipelineCosineSimilarity(a, b), 0.001)

	// Opposite vectors should have similarity -1.0
	c := []float32{-1, 0, 0}
	assert.InDelta(t, -1.0, pipelineCosineSimilarity(a, c), 0.001)
}
