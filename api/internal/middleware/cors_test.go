package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionCORSConfig(t *testing.T) {
	config := ProductionCORSConfig([]string{"https://app.example.com"})

	assert.Equal(t, []string{"https://app.example.com"}, config.AllowOrigins)
	assert.True(t, config.AllowCredentials)
	assert.Contains(t, config.AllowHeaders, "X-Project-ID")
}

func TestCORSMiddlewareAllowsProjectContextHeader(t *testing.T) {
	app := fiber.New()
	app.Use(NewCORSMiddleware(
		ProductionCORSConfig([]string{"https://app.example.com"}),
	).Handler())
	app.Get("/api/public/traces", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	request := httptest.NewRequestWithContext(
		context.Background(),
		fiber.MethodOptions,
		"/api/public/traces",
		nil,
	)
	request.Header.Set("Origin", "https://app.example.com")
	request.Header.Set("Access-Control-Request-Method", fiber.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "X-Project-ID")

	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()
	assert.Equal(t, fiber.StatusNoContent, response.StatusCode)
	assert.Contains(t, response.Header.Get("Access-Control-Allow-Headers"), "X-Project-ID")
}
