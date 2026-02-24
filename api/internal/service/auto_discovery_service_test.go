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

func TestAutoDiscoveryScanProject(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAutoDiscoveryService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	dashboard, err := svc.ScanProject(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, dashboard)

	// Should discover LangChain and OpenAI
	assert.Len(t, dashboard.Frameworks, 2)
	frameworkNames := make(map[domain.FrameworkType]bool)
	for _, fw := range dashboard.Frameworks {
		frameworkNames[fw.Framework] = true
		assert.Equal(t, projectID, fw.ProjectID)
	}
	assert.True(t, frameworkNames[domain.FrameworkTypeLangChain])
	assert.True(t, frameworkNames[domain.FrameworkTypeOpenAI])
	assert.Equal(t, 5, dashboard.TotalComponents)
	assert.Equal(t, 2, dashboard.InstrumentedComponents)
	assert.NotNil(t, dashboard.LastScanAt)
}

func TestAutoDiscoveryToggleInstrumentation(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAutoDiscoveryService(logger)
	ctx := context.Background()

	err := svc.ToggleInstrumentation(ctx, uuid.New(), true)
	assert.NoError(t, err)

	err = svc.ToggleInstrumentation(ctx, uuid.New(), false)
	assert.NoError(t, err)
}

func TestAutoDiscoveryUpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAutoDiscoveryService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	// Valid config
	err := svc.UpdateConfig(ctx, projectID, domain.DiscoveryConfig{
		Enabled:      true,
		SamplingRate: 0.5,
		MaxDepth:     3,
	})
	assert.NoError(t, err)

	// Invalid sampling rate
	err = svc.UpdateConfig(ctx, projectID, domain.DiscoveryConfig{
		Enabled:      true,
		SamplingRate: 1.5,
		MaxDepth:     3,
	})
	assert.Error(t, err)

	// Invalid max depth
	err = svc.UpdateConfig(ctx, projectID, domain.DiscoveryConfig{
		Enabled:      true,
		SamplingRate: 0.5,
		MaxDepth:     0,
	})
	assert.Error(t, err)
}
