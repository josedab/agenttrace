package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	t.Run("generates request ID when not present", func(t *testing.T) {
		app := fiber.New()

		app.Use(RequestID())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Check response header contains request ID
		requestID := resp.Header.Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})

	t.Run("preserves existing request ID from header", func(t *testing.T) {
		app := fiber.New()

		app.Use(RequestID())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		existingID := "existing-request-id-12345"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)

		resp, err := app.Test(req)
		require.NoError(t, err)

		// Check response header contains the original request ID
		requestID := resp.Header.Get("X-Request-ID")
		assert.Equal(t, existingID, requestID)
	})

	t.Run("stores request ID in locals", func(t *testing.T) {
		app := fiber.New()

		var localRequestID string
		app.Use(RequestID())
		app.Get("/test", func(c *fiber.Ctx) error {
			localRequestID = c.Locals("requestID").(string)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.NotEmpty(t, localRequestID)
	})
}

func TestRequestIDWithConfig(t *testing.T) {
	t.Run("uses custom header name", func(t *testing.T) {
		app := fiber.New()

		config := RequestIDConfig{
			Header: "X-Custom-Request-ID",
			Generator: func() string {
				return "custom-generated-id"
			},
		}
		app.Use(RequestID(config))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Check custom header
		requestID := resp.Header.Get("X-Custom-Request-ID")
		assert.Equal(t, "custom-generated-id", requestID)
	})

	t.Run("uses custom generator", func(t *testing.T) {
		app := fiber.New()

		callCount := 0
		config := RequestIDConfig{
			Header: "X-Request-ID",
			Generator: func() string {
				callCount++
				return "generated-id"
			},
		}
		app.Use(RequestID(config))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 1, callCount)
	})

	t.Run("does not call generator when ID exists", func(t *testing.T) {
		app := fiber.New()

		callCount := 0
		config := RequestIDConfig{
			Header: "X-Request-ID",
			Generator: func() string {
				callCount++
				return "generated-id"
			},
		}
		app.Use(RequestID(config))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "existing-id")
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 0, callCount)
	})
}

func TestDefaultRequestIDConfig(t *testing.T) {
	t.Run("default config has correct header", func(t *testing.T) {
		config := DefaultRequestIDConfig()
		assert.Equal(t, "X-Request-ID", config.Header)
	})

	t.Run("default config generator produces UUID format", func(t *testing.T) {
		config := DefaultRequestIDConfig()
		id := config.Generator()
		// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		assert.Len(t, id, 36)
		assert.Contains(t, id, "-")
	})
}

func TestTraceRequestID(t *testing.T) {
	t.Run("generates trace ID format", func(t *testing.T) {
		app := fiber.New()

		app.Use(TraceRequestID())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		requestID := resp.Header.Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		// Trace ID format should be 32 characters
		assert.Len(t, requestID, 32)
	})
}
