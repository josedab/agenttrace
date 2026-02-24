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

func TestNewAgentVersionService(t *testing.T) {
	svc := NewAgentVersionService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestAgentVersionService_CreateVersion(t *testing.T) {
	svc := NewAgentVersionService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("auto increments version number", func(t *testing.T) {
		v1, err := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{
			AgentName:  "my-agent",
			Config:     domain.AgentConfig{Model: "gpt-4o", Temperature: 0.7, MaxTokens: 4096},
			ChangeNote: "initial version",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, v1.Version)
		assert.True(t, v1.IsActive)
		assert.Equal(t, "my-agent", v1.AgentName)

		v2, err := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{
			AgentName:  "my-agent",
			Config:     domain.AgentConfig{Model: "gpt-4o", Temperature: 0.5, MaxTokens: 8192},
			ChangeNote: "tuned temperature",
		})
		require.NoError(t, err)
		assert.Equal(t, 2, v2.Version)
		assert.True(t, v2.IsActive)
	})

	t.Run("deactivates previous versions", func(t *testing.T) {
		pid := uuid.New()
		v1, _ := svc.CreateVersion(ctx, pid, &domain.CreateVersionInput{
			AgentName: "agent-x",
			Config:    domain.AgentConfig{Model: "gpt-4o"},
		})
		_, _ = svc.CreateVersion(ctx, pid, &domain.CreateVersionInput{
			AgentName: "agent-x",
			Config:    domain.AgentConfig{Model: "claude-3.5-sonnet"},
		})

		got, err := svc.GetVersion(ctx, v1.ID)
		require.NoError(t, err)
		assert.False(t, got.IsActive)
	})
}

func TestAgentVersionService_ListVersions(t *testing.T) {
	svc := NewAgentVersionService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	_, _ = svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "agent-a", Config: domain.AgentConfig{Model: "gpt-4o"}})
	_, _ = svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "agent-a", Config: domain.AgentConfig{Model: "gpt-4o-mini"}})
	_, _ = svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "agent-b", Config: domain.AgentConfig{Model: "claude-3.5-sonnet"}})

	t.Run("filters by agent name", func(t *testing.T) {
		versions, err := svc.ListVersions(ctx, projectID, "agent-a")
		require.NoError(t, err)
		assert.Len(t, versions, 2)
	})

	t.Run("empty filter returns all project versions", func(t *testing.T) {
		versions, err := svc.ListVersions(ctx, projectID, "")
		require.NoError(t, err)
		assert.Len(t, versions, 3)
	})

	t.Run("empty result", func(t *testing.T) {
		versions, err := svc.ListVersions(ctx, uuid.New(), "")
		require.NoError(t, err)
		assert.Empty(t, versions)
	})
}

func TestAgentVersionService_GetActiveVersion(t *testing.T) {
	svc := NewAgentVersionService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	_, _ = svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "active-agent", Config: domain.AgentConfig{Model: "gpt-4o"}})
	v2, _ := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "active-agent", Config: domain.AgentConfig{Model: "gpt-4o-mini"}})

	t.Run("returns latest active", func(t *testing.T) {
		active, err := svc.GetActiveVersion(ctx, projectID, "active-agent")
		require.NoError(t, err)
		assert.Equal(t, v2.ID, active.ID)
		assert.True(t, active.IsActive)
	})

	t.Run("not found for unknown agent", func(t *testing.T) {
		_, err := svc.GetActiveVersion(ctx, projectID, "nonexistent")
		assert.Error(t, err)
	})
}

func TestAgentVersionService_Rollback(t *testing.T) {
	svc := NewAgentVersionService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	v1, _ := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "rollback-agent", Config: domain.AgentConfig{Model: "gpt-4o"}})
	v2, _ := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{AgentName: "rollback-agent", Config: domain.AgentConfig{Model: "gpt-4o-mini"}})

	t.Run("changes active version", func(t *testing.T) {
		rolled, err := svc.Rollback(ctx, projectID, v1.ID)
		require.NoError(t, err)
		assert.True(t, rolled.IsActive)
		assert.Equal(t, v1.ID, rolled.ID)

		// v2 should no longer be active
		got, _ := svc.GetVersion(ctx, v2.ID)
		assert.False(t, got.IsActive)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Rollback(ctx, projectID, uuid.New())
		assert.Error(t, err)
	})
}

func TestAgentVersionService_DiffVersions(t *testing.T) {
	svc := NewAgentVersionService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	v1, _ := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{
		AgentName: "diff-agent",
		Config:    domain.AgentConfig{Model: "gpt-4o", Temperature: 0.7, MaxTokens: 4096, SystemPrompt: "You are helpful."},
	})
	v2, _ := svc.CreateVersion(ctx, projectID, &domain.CreateVersionInput{
		AgentName: "diff-agent",
		Config:    domain.AgentConfig{Model: "claude-3.5-sonnet", Temperature: 0.5, MaxTokens: 4096, SystemPrompt: "You are helpful."},
	})

	t.Run("detects config changes", func(t *testing.T) {
		diff, err := svc.DiffVersions(ctx, v1.ID, v2.ID)
		require.NoError(t, err)
		assert.Equal(t, v1.ID, diff.VersionA.ID)
		assert.Equal(t, v2.ID, diff.VersionB.ID)

		// model and temperature differ
		fieldNames := make([]string, 0, len(diff.Changes))
		for _, c := range diff.Changes {
			fieldNames = append(fieldNames, c.Field)
		}
		assert.Contains(t, fieldNames, "model")
		assert.Contains(t, fieldNames, "temperature")
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.DiffVersions(ctx, uuid.New(), v2.ID)
		assert.Error(t, err)
	})
}
