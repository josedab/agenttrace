package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewEmbedService(t *testing.T) {
	svc := NewEmbedService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestEmbedService_CreateConfig(t *testing.T) {
	svc := NewEmbedService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	config, err := svc.CreateConfig(ctx, projectID, &domain.EmbedConfigInput{
		Theme: domain.EmbedTheme{PrimaryColor: "#3B82F6"},
		Features: domain.EmbedFeatures{
			TraceViewer:   true,
			CostDashboard: true,
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, config.ID)
	assert.Equal(t, projectID, config.ProjectID)
	assert.True(t, config.Enabled)
	assert.Equal(t, "#3B82F6", config.Theme.PrimaryColor)
	assert.True(t, config.Features.TraceViewer)
	assert.NotNil(t, config.AllowedOrigins)
}

func TestEmbedService_GetConfig(t *testing.T) {
	svc := NewEmbedService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("found", func(t *testing.T) {
		_, _ = svc.CreateConfig(ctx, projectID, &domain.EmbedConfigInput{})
		got, err := svc.GetConfig(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, projectID, got.ProjectID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetConfig(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestEmbedService_UpdateConfig(t *testing.T) {
	svc := NewEmbedService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	_, _ = svc.CreateConfig(ctx, projectID, &domain.EmbedConfigInput{
		Theme:    domain.EmbedTheme{PrimaryColor: "#000"},
		Features: domain.EmbedFeatures{TraceViewer: true},
	})

	updated, err := svc.UpdateConfig(ctx, projectID, &domain.EmbedConfigInput{
		Theme:          domain.EmbedTheme{PrimaryColor: "#FFF", FontFamily: "Inter"},
		AllowedOrigins: []string{"https://example.com"},
		Features:       domain.EmbedFeatures{TraceViewer: true, CostDashboard: true},
	})
	require.NoError(t, err)
	assert.Equal(t, "#FFF", updated.Theme.PrimaryColor)
	assert.Equal(t, "Inter", updated.Theme.FontFamily)
	assert.True(t, updated.Features.CostDashboard)
	assert.Contains(t, updated.AllowedOrigins, "https://example.com")

	t.Run("not found project", func(t *testing.T) {
		_, err := svc.UpdateConfig(ctx, uuid.New(), &domain.EmbedConfigInput{})
		assert.Error(t, err)
	})
}

func TestEmbedService_GenerateToken(t *testing.T) {
	svc := NewEmbedService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("generates valid token", func(t *testing.T) {
		_, _ = svc.CreateConfig(ctx, projectID, &domain.EmbedConfigInput{})
		token, err := svc.GenerateToken(ctx, projectID)
		require.NoError(t, err)
		assert.NotEmpty(t, token.Token)
		assert.True(t, strings.HasPrefix(token.Token, "at_embed_"))
		assert.True(t, token.ExpiresAt.After(time.Now()))
		assert.NotEqual(t, uuid.Nil, token.ConfigID)
	})

	t.Run("no config returns error", func(t *testing.T) {
		_, err := svc.GenerateToken(ctx, uuid.New())
		assert.Error(t, err)
	})
}
