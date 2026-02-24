package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewSkillProfileService(t *testing.T) {
	svc := NewSkillProfileService(zap.NewNop(), nil)
	assert.NotNil(t, svc)
}

func TestSkillProfileService_GetProfile(t *testing.T) {
	svc := NewSkillProfileService(zap.NewNop(), nil)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("returns profile for agent", func(t *testing.T) {
		profile, err := svc.GetProfile(ctx, projectID, "claude-code")
		require.NoError(t, err)
		require.NotNil(t, profile)
		assert.Equal(t, "claude-code", profile.AgentName)
		assert.Equal(t, projectID, profile.ProjectID)
		assert.Greater(t, len(profile.Skills), 0)
		assert.Greater(t, profile.TotalTraces, 0)
		assert.Greater(t, profile.SuccessRate, 0.0)

		for _, skill := range profile.Skills {
			assert.GreaterOrEqual(t, skill.Score, 0.0)
			assert.LessOrEqual(t, skill.Score, 100.0)
			assert.GreaterOrEqual(t, skill.Confidence, 0.0)
			assert.LessOrEqual(t, skill.Confidence, 1.0)
		}
	})

	t.Run("has language stats", func(t *testing.T) {
		profile, err := svc.GetProfile(ctx, projectID, "test-agent")
		require.NoError(t, err)
		assert.Greater(t, len(profile.LanguageStats), 0)
	})

	t.Run("has model stats", func(t *testing.T) {
		profile, err := svc.GetProfile(ctx, projectID, "test-agent")
		require.NoError(t, err)
		assert.Greater(t, len(profile.ModelStats), 0)
	})
}

func TestSkillProfileService_ListProfiles(t *testing.T) {
	svc := NewSkillProfileService(zap.NewNop(), nil)
	ctx := context.Background()

	profiles, err := svc.ListProfiles(ctx, uuid.New())
	require.NoError(t, err)
	assert.Greater(t, len(profiles), 0)
}

func TestSkillProfileService_CompareAgents(t *testing.T) {
	svc := NewSkillProfileService(zap.NewNop(), nil)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("compares multiple agents", func(t *testing.T) {
		comparison, err := svc.CompareAgents(ctx, projectID, []string{"agent-a", "agent-b"})
		require.NoError(t, err)
		assert.Len(t, comparison.Agents, 2)
		assert.Greater(t, len(comparison.BestAgent), 0)
	})

	t.Run("handles single agent", func(t *testing.T) {
		comparison, err := svc.CompareAgents(ctx, projectID, []string{"agent-a"})
		require.NoError(t, err)
		assert.Len(t, comparison.Agents, 1)
	})
}
