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

func TestNewPluginService(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestPluginService_InstallPlugin(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	plugin, err := svc.InstallPlugin(ctx, projectID, &domain.PluginInput{
		Manifest: domain.PluginManifest{
			Name:        "cost-evaluator",
			Version:     "1.0.0",
			Type:        domain.PluginTypeEvaluator,
			Description: "Evaluates agent cost efficiency",
			Author:      "test-author",
			EntryPoint:  "main.wasm",
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, plugin.ID)
	assert.Equal(t, "cost-evaluator", plugin.Name)
	assert.Equal(t, domain.PluginStatusInstalled, plugin.Status)
	assert.Equal(t, projectID, plugin.ProjectID)
	assert.Equal(t, "1.0.0", plugin.Version)
}

func TestPluginService_ListPlugins(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	_, _ = svc.InstallPlugin(ctx, projectA, &domain.PluginInput{Manifest: domain.PluginManifest{Name: "p1", Version: "1.0", Type: domain.PluginTypeEvaluator}})
	_, _ = svc.InstallPlugin(ctx, projectA, &domain.PluginInput{Manifest: domain.PluginManifest{Name: "p2", Version: "1.0", Type: domain.PluginTypeTraceProcessor}})
	_, _ = svc.InstallPlugin(ctx, projectB, &domain.PluginInput{Manifest: domain.PluginManifest{Name: "p3", Version: "1.0", Type: domain.PluginTypeDashboardWidget}})

	t.Run("filters by project", func(t *testing.T) {
		reg, err := svc.ListPlugins(ctx, projectA)
		require.NoError(t, err)
		assert.Equal(t, 2, reg.TotalCount)
		assert.Len(t, reg.Plugins, 2)
	})

	t.Run("empty project", func(t *testing.T) {
		reg, err := svc.ListPlugins(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 0, reg.TotalCount)
	})
}

func TestPluginService_ActivatePlugin(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	ctx := context.Background()

	plugin, _ := svc.InstallPlugin(ctx, uuid.New(), &domain.PluginInput{
		Manifest: domain.PluginManifest{Name: "activate-me", Version: "1.0", Type: domain.PluginTypeEvaluator},
	})

	t.Run("changes status to active", func(t *testing.T) {
		activated, err := svc.ActivatePlugin(ctx, plugin.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PluginStatusActive, activated.Status)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.ActivatePlugin(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestPluginService_DisablePlugin(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	ctx := context.Background()

	plugin, _ := svc.InstallPlugin(ctx, uuid.New(), &domain.PluginInput{
		Manifest: domain.PluginManifest{Name: "disable-me", Version: "1.0", Type: domain.PluginTypeEvaluator},
	})
	_, _ = svc.ActivatePlugin(ctx, plugin.ID)

	disabled, err := svc.DisablePlugin(ctx, plugin.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PluginStatusDisabled, disabled.Status)
}

func TestPluginService_ExecutePlugin(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	ctx := context.Background()

	plugin, _ := svc.InstallPlugin(ctx, uuid.New(), &domain.PluginInput{
		Manifest: domain.PluginManifest{Name: "exec-plugin", Version: "1.0", Type: domain.PluginTypeEvaluator},
	})

	t.Run("fails if not active", func(t *testing.T) {
		_, err := svc.ExecutePlugin(ctx, plugin.ID, "test input")
		assert.Error(t, err)
	})

	t.Run("succeeds when active", func(t *testing.T) {
		_, _ = svc.ActivatePlugin(ctx, plugin.ID)
		exec, err := svc.ExecutePlugin(ctx, plugin.ID, "test input data")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, exec.ID)
		assert.Equal(t, plugin.ID, exec.PluginID)
		assert.Equal(t, domain.PluginExecSuccess, exec.Status)
		assert.NotEmpty(t, exec.Output)
		assert.Greater(t, exec.DurationMs, int64(0))
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.ExecutePlugin(ctx, uuid.New(), "input")
		assert.Error(t, err)
	})
}

func TestPluginService_UninstallPlugin(t *testing.T) {
	svc := NewPluginService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	plugin, _ := svc.InstallPlugin(ctx, projectID, &domain.PluginInput{
		Manifest: domain.PluginManifest{Name: "uninstall-me", Version: "1.0", Type: domain.PluginTypeEvaluator},
	})

	t.Run("removes from registry", func(t *testing.T) {
		err := svc.UninstallPlugin(ctx, plugin.ID)
		require.NoError(t, err)

		_, err = svc.GetPlugin(ctx, plugin.ID)
		assert.Error(t, err)

		reg, _ := svc.ListPlugins(ctx, projectID)
		assert.Equal(t, 0, reg.TotalCount)
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.UninstallPlugin(ctx, uuid.New())
		assert.Error(t, err)
	})
}
