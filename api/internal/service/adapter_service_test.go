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

func TestNewAdapterService(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestAdapterService_RegisterAdapter(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		svc := NewAdapterService(zap.NewNop())
		ctx := context.Background()
		projectID := uuid.New()

		adapter, err := svc.RegisterAdapter(ctx, projectID, &domain.AdapterInput{
			Name:      "my-langchain-adapter",
			Framework: domain.AdapterFrameworkLangChain,
		})
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, adapter.ID)
		assert.Equal(t, projectID, adapter.ProjectID)
		assert.Equal(t, "my-langchain-adapter", adapter.Name)
		assert.Equal(t, domain.AdapterFrameworkLangChain, adapter.Framework)
		assert.Equal(t, domain.AdapterStatusRegistered, adapter.Status)
		assert.Equal(t, "1.0.0", adapter.Version)
		assert.True(t, adapter.Config.AutoInstrument)
		assert.True(t, adapter.Config.CaptureIO)
		assert.Equal(t, 1.0, adapter.Config.SamplingRate)
		assert.Contains(t, adapter.Capabilities, "trace_capture")
		assert.Len(t, adapter.LifecycleHooks, 3)
	})

	t.Run("custom config", func(t *testing.T) {
		svc := NewAdapterService(zap.NewNop())
		ctx := context.Background()

		cfg := &domain.AdapterConfig{
			AutoInstrument: false,
			CaptureIO:      false,
			MaxSpanDepth:   10,
			SamplingRate:   0.5,
		}
		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "custom-adapter",
			Framework: domain.AdapterFrameworkCrewAI,
			Version:   "2.0.0",
			Config:    cfg,
		})
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", adapter.Version)
		assert.False(t, adapter.Config.AutoInstrument)
		assert.Equal(t, 0.5, adapter.Config.SamplingRate)
	})

	t.Run("missing name", func(t *testing.T) {
		svc := NewAdapterService(zap.NewNop())
		ctx := context.Background()

		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "",
			Framework: domain.AdapterFrameworkLangChain,
		})
		assert.Nil(t, adapter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "adapter name is required")
	})

	t.Run("invalid framework", func(t *testing.T) {
		svc := NewAdapterService(zap.NewNop())
		ctx := context.Background()

		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "bad-framework",
			Framework: domain.AdapterFramework("nonexistent"),
		})
		assert.Nil(t, adapter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid framework")
	})
}

func TestAdapterService_GetAdapter(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()

	t.Run("existing adapter", func(t *testing.T) {
		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "test-adapter",
			Framework: domain.AdapterFrameworkAutoGen,
		})
		require.NoError(t, err)

		got, err := svc.GetAdapter(ctx, adapter.ID)
		require.NoError(t, err)
		assert.Equal(t, adapter.ID, got.ID)
		assert.Equal(t, "test-adapter", got.Name)
	})

	t.Run("non-existent adapter", func(t *testing.T) {
		got, err := svc.GetAdapter(ctx, uuid.New())
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "adapter not found")
	})
}

func TestAdapterService_ListAdapters(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("empty project", func(t *testing.T) {
		adapters, err := svc.ListAdapters(ctx, projectID)
		require.NoError(t, err)
		assert.Empty(t, adapters)
	})

	t.Run("after registering adapters", func(t *testing.T) {
		_, err := svc.RegisterAdapter(ctx, projectID, &domain.AdapterInput{
			Name:      "adapter-1",
			Framework: domain.AdapterFrameworkLangChain,
		})
		require.NoError(t, err)
		_, err = svc.RegisterAdapter(ctx, projectID, &domain.AdapterInput{
			Name:      "adapter-2",
			Framework: domain.AdapterFrameworkCrewAI,
		})
		require.NoError(t, err)

		// Register in a different project — should not appear
		_, err = svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "other-project-adapter",
			Framework: domain.AdapterFrameworkCustom,
		})
		require.NoError(t, err)

		adapters, err := svc.ListAdapters(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, adapters, 2)
	})
}

func TestAdapterService_UpdateAdapter(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()

	adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
		Name:      "updatable-adapter",
		Framework: domain.AdapterFrameworkLangGraph,
	})
	require.NoError(t, err)

	t.Run("partial update name", func(t *testing.T) {
		newName := "renamed-adapter"
		updated, err := svc.UpdateAdapter(ctx, adapter.ID, &domain.AdapterUpdateInput{
			Name: &newName,
		})
		require.NoError(t, err)
		assert.Equal(t, "renamed-adapter", updated.Name)
		assert.True(t, updated.UpdatedAt.After(adapter.CreatedAt))
	})

	t.Run("status change", func(t *testing.T) {
		newStatus := domain.AdapterStatusInactive
		updated, err := svc.UpdateAdapter(ctx, adapter.ID, &domain.AdapterUpdateInput{
			Status: &newStatus,
		})
		require.NoError(t, err)
		assert.Equal(t, domain.AdapterStatusInactive, updated.Status)
	})

	t.Run("non-existent adapter", func(t *testing.T) {
		newName := "nope"
		updated, err := svc.UpdateAdapter(ctx, uuid.New(), &domain.AdapterUpdateInput{
			Name: &newName,
		})
		assert.Nil(t, updated)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "adapter not found")
	})
}

func TestAdapterService_DeleteAdapter(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()

	t.Run("existing adapter", func(t *testing.T) {
		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "delete-me",
			Framework: domain.AdapterFrameworkCrewAI,
		})
		require.NoError(t, err)

		err = svc.DeleteAdapter(ctx, adapter.ID)
		require.NoError(t, err)

		// Verify it's gone
		got, err := svc.GetAdapter(ctx, adapter.ID)
		assert.Nil(t, got)
		assert.Error(t, err)
	})

	t.Run("non-existent adapter", func(t *testing.T) {
		err := svc.DeleteAdapter(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "adapter not found")
	})
}

func TestAdapterService_IngestEvent(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()

	adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
		Name:      "ingest-adapter",
		Framework: domain.AdapterFrameworkLangChain,
	})
	require.NoError(t, err)

	t.Run("valid event", func(t *testing.T) {
		err := svc.IngestEvent(ctx, adapter.ID, &domain.AdapterEvent{
			EventType: "trace_start",
			Name:      "agent-run",
		})
		require.NoError(t, err)

		// Verify stats were updated
		got, err := svc.GetAdapter(ctx, adapter.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.Stats.TotalTraces)
		assert.Equal(t, int64(1), got.Stats.TotalSpans)
		assert.NotNil(t, got.Stats.LastActiveAt)
		assert.Equal(t, domain.AdapterStatusActive, got.Status)
	})

	t.Run("span event increments spans only", func(t *testing.T) {
		err := svc.IngestEvent(ctx, adapter.ID, &domain.AdapterEvent{
			EventType: "span_start",
			Name:      "tool-call",
		})
		require.NoError(t, err)

		got, err := svc.GetAdapter(ctx, adapter.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.Stats.TotalTraces)
		assert.Equal(t, int64(2), got.Stats.TotalSpans)
	})

	t.Run("missing event type", func(t *testing.T) {
		err := svc.IngestEvent(ctx, adapter.ID, &domain.AdapterEvent{
			Name: "no-type",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event type is required")
	})

	t.Run("invalid adapter", func(t *testing.T) {
		err := svc.IngestEvent(ctx, uuid.New(), &domain.AdapterEvent{
			EventType: "trace_start",
			Name:      "orphan-event",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "adapter not found")
	})
}

func TestAdapterService_TestAdapter(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()

	t.Run("returns test results", func(t *testing.T) {
		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "testable-adapter",
			Framework: domain.AdapterFrameworkAutoGen,
		})
		require.NoError(t, err)

		result, err := svc.TestAdapter(ctx, adapter.ID)
		require.NoError(t, err)
		assert.Equal(t, adapter.ID, result.AdapterID)
		assert.Equal(t, domain.AdapterFrameworkAutoGen, result.Framework)
		assert.True(t, result.Passed)
		assert.Len(t, result.TestResults, 4)
		assert.Equal(t, "All tests passed", result.Summary)
	})

	t.Run("non-existent adapter", func(t *testing.T) {
		result, err := svc.TestAdapter(ctx, uuid.New())
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "adapter not found")
	})

	t.Run("no lifecycle hooks fails test", func(t *testing.T) {
		adapter, err := svc.RegisterAdapter(ctx, uuid.New(), &domain.AdapterInput{
			Name:      "no-hooks-adapter",
			Framework: domain.AdapterFrameworkCustom,
		})
		require.NoError(t, err)

		// Clear lifecycle hooks via update to trigger the failed test case
		updated, err := svc.UpdateAdapter(ctx, adapter.ID, &domain.AdapterUpdateInput{
			LifecycleHooks: []domain.LifecycleHook{},
		})
		require.NoError(t, err)
		assert.Empty(t, updated.LifecycleHooks)

		result, err := svc.TestAdapter(ctx, adapter.ID)
		require.NoError(t, err)
		assert.False(t, result.Passed)
		assert.Contains(t, result.Summary, "Some tests failed")
	})
}

func TestAdapterService_GetTemplates(t *testing.T) {
	svc := NewAdapterService(zap.NewNop())
	ctx := context.Background()

	templates := svc.GetTemplates(ctx)
	assert.Len(t, templates, 6)

	frameworks := make(map[domain.AdapterFramework]bool)
	for _, tmpl := range templates {
		frameworks[tmpl.Framework] = true
		assert.NotEmpty(t, tmpl.Name)
		assert.NotEmpty(t, tmpl.Description)
		assert.NotEmpty(t, tmpl.SetupCode)
		assert.NotEmpty(t, tmpl.Language)
		assert.NotEmpty(t, tmpl.Dependencies)
	}

	assert.True(t, frameworks[domain.AdapterFrameworkLangChain])
	assert.True(t, frameworks[domain.AdapterFrameworkCrewAI])
	assert.True(t, frameworks[domain.AdapterFrameworkAutoGen])
	assert.True(t, frameworks[domain.AdapterFrameworkLangGraph])
	assert.True(t, frameworks[domain.AdapterFrameworkOpenAIAgents])
	assert.True(t, frameworks[domain.AdapterFrameworkMCP])
}
