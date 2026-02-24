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

func TestAIDebuggerDebugTrace(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAIDebuggerService(logger)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New()

	result, err := svc.DebugTrace(ctx, projectID, traceID, "Why did this trace fail?", domain.DebugQueryTypeRootCause)
	require.NoError(t, err)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, traceID, result.TraceID)
	assert.Equal(t, domain.DebugQueryTypeRootCause, result.QueryType)
	// Root cause query should have root causes
	assert.NotEmpty(t, result.Response.RootCauses)
	assert.NotEmpty(t, result.Response.SuggestedFixes)
	assert.Greater(t, result.Response.Confidence, 0.0)

	// Empty query should fail
	_, err = svc.DebugTrace(ctx, projectID, traceID, "", domain.DebugQueryTypeRootCause)
	assert.Error(t, err)

	// Invalid query type should fail
	_, err = svc.DebugTrace(ctx, projectID, traceID, "query", domain.DebugQueryType("invalid"))
	assert.Error(t, err)
}

func TestAIDebuggerBuildContext(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAIDebuggerService(logger)
	ctx := context.Background()
	traceID := uuid.New()

	debugCtx, err := svc.BuildContext(ctx, traceID)
	require.NoError(t, err)
	assert.NotEmpty(t, debugCtx.TraceSummary)
	assert.NotEmpty(t, debugCtx.Observations)
	assert.Len(t, debugCtx.Observations, 3)
	assert.NotEmpty(t, debugCtx.Errors)
	assert.NotNil(t, debugCtx.GitContext)
	assert.Greater(t, debugCtx.CostTotal, 0.0)
	assert.Greater(t, debugCtx.DurationMs, int64(0))
	assert.NotEmpty(t, debugCtx.FileOperations)
}

func TestAIDebuggerGetDebugHistoryEmpty(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAIDebuggerService(logger)
	ctx := context.Background()

	history, err := svc.GetDebugHistory(ctx, uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, history)
}
