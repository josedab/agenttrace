package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewMarketplaceService(t *testing.T) {
	svc := NewMarketplaceService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestMarketplaceService_SeededPackages(t *testing.T) {
	svc := NewMarketplaceService(zap.NewNop())
	ctx := context.Background()

	t.Run("has featured packages", func(t *testing.T) {
		featured := svc.GetFeatured(ctx)
		assert.Greater(t, len(featured), 0)
	})

	t.Run("search returns results", func(t *testing.T) {
		results, total := svc.SearchPackages(ctx, &domain.MarketplaceSearch{Limit: 10})
		assert.Greater(t, len(results), 0)
		assert.Greater(t, total, 0)
	})
}

func TestMarketplaceService_PublishAndInstall(t *testing.T) {
	svc := NewMarketplaceService(zap.NewNop())
	ctx := context.Background()

	t.Run("publishes package", func(t *testing.T) {
		input := &domain.PackagePublishInput{
			Name: "test-prompt-template", Type: domain.PackagePrompt,
			Description: "A test prompt", Content: "prompt: test", Tags: []string{"test"},
		}
		pkg, err := svc.PublishPackage(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, "test-prompt-template", pkg.Name)
		assert.Equal(t, domain.PackagePrompt, pkg.Type)
		assert.Equal(t, 0, pkg.Downloads)
	})

	t.Run("installs package increments downloads", func(t *testing.T) {
		results, _ := svc.SearchPackages(ctx, &domain.MarketplaceSearch{Limit: 1})
		require.Greater(t, len(results), 0)

		pkg, err := svc.InstallPackage(ctx, results[0].ID)
		require.NoError(t, err)
		assert.Greater(t, pkg.Downloads, 0)
	})
}

func TestMarketplaceService_Rating(t *testing.T) {
	svc := NewMarketplaceService(zap.NewNop())
	ctx := context.Background()

	results, _ := svc.SearchPackages(ctx, &domain.MarketplaceSearch{Limit: 1})
	require.Greater(t, len(results), 0)
	pkgID := results[0].ID

	t.Run("rates package", func(t *testing.T) {
		pkg, err := svc.RatePackage(ctx, pkgID, 5, "Excellent!")
		require.NoError(t, err)
		assert.Greater(t, pkg.RatingCount, 0)
	})

	t.Run("gets package by ID", func(t *testing.T) {
		pkg, err := svc.GetPackage(ctx, pkgID)
		require.NoError(t, err)
		assert.NotNil(t, pkg)
		assert.Equal(t, pkgID, pkg.ID)
	})
}
