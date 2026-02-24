package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIDETraceViewGetFileMapping(t *testing.T) {
	logger := zap.NewNop()
	svc := NewIDETraceViewService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	filePath := "src/auth/handler.go"

	mapping, err := svc.GetFileMapping(ctx, projectID, filePath)
	require.NoError(t, err)
	require.NotNil(t, mapping)

	assert.Equal(t, filePath, mapping.FilePath)
	assert.Equal(t, projectID, mapping.ProjectID)
	assert.NotNil(t, mapping.Annotations)
	assert.Empty(t, mapping.Annotations)
	assert.NotNil(t, mapping.Summary.TopAgents)
	assert.Empty(t, mapping.Summary.TopAgents)
}

func TestIDETraceViewGetTraceContext(t *testing.T) {
	logger := zap.NewNop()
	svc := NewIDETraceViewService(logger)
	ctx := context.Background()

	traceID := uuid.New()
	traceCtx, err := svc.GetTraceContext(ctx, traceID)
	require.NoError(t, err)
	require.NotNil(t, traceCtx)

	assert.Equal(t, traceID, traceCtx.TraceID)
	assert.NotNil(t, traceCtx.FileChanges)
	assert.Empty(t, traceCtx.FileChanges)
}

func TestIDETraceViewGetBatchMappings(t *testing.T) {
	logger := zap.NewNop()
	svc := NewIDETraceViewService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	filePaths := []string{
		"src/auth/handler.go",
		"src/api/router.go",
		"src/models/user.go",
	}

	mappings, err := svc.GetBatchMappings(ctx, projectID, filePaths)
	require.NoError(t, err)
	require.Len(t, mappings, 3)

	for i, mapping := range mappings {
		assert.Equal(t, filePaths[i], mapping.FilePath)
		assert.Equal(t, projectID, mapping.ProjectID)
		assert.NotNil(t, mapping.Annotations)
	}
}
