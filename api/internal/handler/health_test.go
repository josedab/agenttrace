package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthStatus(t *testing.T) {
	t.Run("structure validation", func(t *testing.T) {
		status := HealthStatus{
			Status:    "healthy",
			Version:   "1.0.0",
			Uptime:    "1h30m",
			Timestamp: "2024-01-15T10:30:00Z",
			Checks: map[string]string{
				"postgres":   "healthy",
				"clickhouse": "healthy",
				"redis":      "healthy",
			},
		}

		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, "1h30m", status.Uptime)
		assert.Len(t, status.Checks, 3)
	})
}

func TestNewHealthHandler(t *testing.T) {
	t.Run("creates handler with correct initialization", func(t *testing.T) {
		handler := NewHealthHandler(nil, nil, nil, "1.2.3")

		require.NotNil(t, handler)
		assert.Equal(t, "1.2.3", handler.version)
		assert.False(t, handler.startTime.IsZero())
	})

	t.Run("start time is set to creation time", func(t *testing.T) {
		before := time.Now()
		handler := NewHealthHandler(nil, nil, nil, "1.0.0")
		after := time.Now()

		assert.True(t, handler.startTime.After(before) || handler.startTime.Equal(before))
		assert.True(t, handler.startTime.Before(after) || handler.startTime.Equal(after))
	})
}

func TestHealthHandler_Liveness(t *testing.T) {
	t.Run("returns alive status", func(t *testing.T) {
		app := fiber.New()
		handler := NewHealthHandler(nil, nil, nil, "1.0.0")

		app.Get("/livez", handler.Liveness)

		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "alive", result["status"])
	})
}

func TestHealthHandler_Version(t *testing.T) {
	t.Run("returns version and uptime", func(t *testing.T) {
		app := fiber.New()
		handler := NewHealthHandler(nil, nil, nil, "2.1.0")

		// Wait a bit to have measurable uptime
		time.Sleep(10 * time.Millisecond)

		app.Get("/version", handler.Version)

		req := httptest.NewRequest(http.MethodGet, "/version", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "2.1.0", result["version"])
		assert.NotEmpty(t, result["uptime"])
	})
}

func TestHealthHandler_RegisterRoutes(t *testing.T) {
	t.Run("registers all health routes", func(t *testing.T) {
		app := fiber.New()
		handler := NewHealthHandler(nil, nil, nil, "1.0.0")

		handler.RegisterRoutes(app)

		routes := app.GetRoutes()
		routePaths := make(map[string]bool)
		for _, route := range routes {
			if route.Method == "GET" {
				routePaths[route.Path] = true
			}
		}

		expectedRoutes := []string{
			"/health",
			"/healthz",
			"/livez",
			"/live",
			"/readyz",
			"/ready",
			"/version",
		}

		for _, path := range expectedRoutes {
			assert.True(t, routePaths[path], "Route %s should be registered", path)
		}
	})
}
