package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAgentKnowledgeGraphBuildGraph(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentKnowledgeGraphService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	graph, err := svc.BuildGraph(ctx, projectID, "authentication flow")
	require.NoError(t, err)
	require.NotNil(t, graph)

	assert.Equal(t, projectID, graph.ProjectID)
	assert.NotNil(t, graph.Nodes)
	assert.Empty(t, graph.Nodes)
	assert.NotNil(t, graph.Edges)
	assert.Empty(t, graph.Edges)
	assert.NotNil(t, graph.Stats.MostConnected)
	assert.False(t, graph.GeneratedAt.IsZero())
}

func TestAgentKnowledgeGraphGetEvolution(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentKnowledgeGraphService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	evolution, err := svc.GetEvolution(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, evolution)

	assert.Equal(t, projectID, evolution.ProjectID)
	assert.NotNil(t, evolution.Snapshots)
	assert.Empty(t, evolution.Snapshots)
}

func TestAgentKnowledgeGraphQueryGraph(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentKnowledgeGraphService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	result, err := svc.QueryGraph(ctx, projectID, "database queries")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, projectID, result.ProjectID)
	assert.NotNil(t, result.Nodes)
	assert.NotNil(t, result.Edges)
	assert.False(t, result.GeneratedAt.IsZero())
}

func TestAgentKnowledgeGraphGetStats(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentKnowledgeGraphService(logger)
	ctx := context.Background()

	stats, err := svc.GetStats(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, 0, stats.TotalNodes)
	assert.Equal(t, 0, stats.TotalEdges)
	assert.Equal(t, 0, stats.FilesCovered)
	assert.Equal(t, 0, stats.FunctionsCovered)
	assert.Equal(t, 0, stats.AvgDepth)
	assert.NotNil(t, stats.MostConnected)
	assert.Empty(t, stats.MostConnected)
}
