package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/config"
)

func TestSetupGraphQLDisabledInProduction(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	cfg := &config.Config{
		Server: config.ServerConfig{Env: "production"},
	}

	setupGraphQL(app, nil, cfg)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/graphql",
		nil,
	))
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	assert.Equal(t, fiber.StatusNotFound, response.StatusCode)
}

func TestSentryRuntimeEnabledHonorsNoEgress(t *testing.T) {
	cfg := &config.Config{
		Sentry: config.SentryConfig{
			Enabled: true,
			DSN:     "https://public@example.ingest.sentry.io/1",
		},
		Privacy: config.PrivacyConfig{
			NoEgress:         true,
			RedactionEnabled: true,
		},
	}

	assert.False(t, sentryRuntimeEnabled(cfg))

	cfg.Privacy.NoEgress = false
	assert.True(t, sentryRuntimeEnabled(cfg))
}
