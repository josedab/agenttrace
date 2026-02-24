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

func TestOTelCompatCreateDestination(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelCompatService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	input := domain.OTelExportDestination{
		Name:     "Jaeger Export",
		Endpoint: "http://jaeger:14268",
		Format:   domain.OTelExportFormatJaeger,
	}

	dest, err := svc.CreateDestination(ctx, projectID, input)
	require.NoError(t, err)
	assert.Equal(t, "Jaeger Export", dest.Name)
	assert.Equal(t, projectID, dest.ProjectID)
	assert.True(t, dest.Enabled)
	assert.Equal(t, int64(0), dest.ExportedCount)
	// Defaults should be applied
	assert.Equal(t, 1.0, dest.SamplingRate)
	assert.Equal(t, 512, dest.BatchSize)
	assert.Equal(t, 5000, dest.FlushIntervalMs)

	// Empty name should fail
	_, err = svc.CreateDestination(ctx, projectID, domain.OTelExportDestination{
		Name: "", Endpoint: "http://x", Format: domain.OTelExportFormatJaeger,
	})
	assert.Error(t, err)

	// Empty endpoint should fail
	_, err = svc.CreateDestination(ctx, projectID, domain.OTelExportDestination{
		Name: "X", Endpoint: "", Format: domain.OTelExportFormatJaeger,
	})
	assert.Error(t, err)
}

func TestOTelCompatGetMappings(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelCompatService(logger)

	mappings, err := svc.GetMappings()
	require.NoError(t, err)
	assert.NotEmpty(t, mappings)

	// Verify gen_ai.* mappings exist
	genAICount := 0
	for _, m := range mappings {
		if m.OTelNamespace == "gen_ai" || m.OTelNamespace == "gen_ai.content" {
			genAICount++
		}
	}
	assert.Greater(t, genAICount, 0, "should have gen_ai.* namespace mappings")

	// Verify key mappings exist
	attrMap := make(map[string]bool)
	for _, m := range mappings {
		attrMap[m.OTelAttribute] = true
	}
	assert.True(t, attrMap["gen_ai.request.model"])
	assert.True(t, attrMap["gen_ai.usage.input_tokens"])
	assert.True(t, attrMap["gen_ai.usage.output_tokens"])
}

func TestOTelCompatGenerateCollectorConfig(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelCompatService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	config, err := svc.GenerateCollectorConfig(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, projectID, config.ProjectID)
	assert.NotEmpty(t, config.Receivers)
	assert.NotEmpty(t, config.Exporters)
	assert.NotEmpty(t, config.Processors)
	assert.NotNil(t, config.PipelineConfig)
	// Should have traces pipeline
	traces, ok := config.PipelineConfig["traces"]
	assert.True(t, ok, "pipeline config should have 'traces' key")
	assert.NotNil(t, traces)
}

func TestOTelCompatGetDashboard(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelCompatService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	dashboard, err := svc.GetDashboard(ctx, projectID)
	require.NoError(t, err)
	assert.Greater(t, dashboard.ImportedTraces, int64(0))
	assert.Greater(t, dashboard.ExportedTraces, int64(0))
	assert.Equal(t, domain.OTelSemanticVersionLatest, dashboard.SemanticVersion)
	assert.Greater(t, dashboard.MappingCoverage, 0.0)
}
