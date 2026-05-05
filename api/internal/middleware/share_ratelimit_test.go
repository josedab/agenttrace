package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShareRateLimiterLimitsAndResets(t *testing.T) {
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	limiter := NewShareRateLimiter(2, time.Minute)
	limiter.clock = func() time.Time { return now }

	app := fiber.New()
	app.Get("/share", limiter.Handler(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for range 2 {
		response, err := app.Test(httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/share",
			nil,
		))
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, response.StatusCode)
		require.NoError(t, response.Body.Close())
	}
	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/share",
		nil,
	))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, response.StatusCode)
	require.NoError(t, response.Body.Close())

	now = now.Add(time.Minute)
	response, err = app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/share",
		nil,
	))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	require.NoError(t, response.Body.Close())
}

func TestShareRateLimiterBoundsActiveClientEntries(t *testing.T) {
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	limiter := NewShareRateLimiter(2, time.Minute)
	limiter.maxEntries = 2
	limiter.clock = func() time.Time { return now }

	app := fiber.New(fiber.Config{ProxyHeader: "X-Forwarded-For"})
	app.Get("/share", limiter.Handler(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	request := func(ip string) *http.Response {
		t.Helper()
		httpRequest := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/share",
			nil,
		)
		httpRequest.Header.Set("X-Forwarded-For", ip)
		response, err := app.Test(httpRequest)
		require.NoError(t, err)
		return response
	}

	for _, ip := range []string{"192.0.2.1", "192.0.2.2"} {
		response := request(ip)
		assert.Equal(t, fiber.StatusOK, response.StatusCode)
		require.NoError(t, response.Body.Close())
	}

	response := request("192.0.2.3")
	assert.Equal(t, fiber.StatusTooManyRequests, response.StatusCode)
	require.NoError(t, response.Body.Close())
	assert.Len(t, limiter.entries, 2)

	now = now.Add(time.Minute)
	response = request("192.0.2.3")
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	require.NoError(t, response.Body.Close())
	assert.Len(t, limiter.entries, 1)
}
