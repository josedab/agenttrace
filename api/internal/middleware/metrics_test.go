package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPathNormalizer(t *testing.T) {
	t.Parallel()
	assert.Equal(
		t,
		unmatchedMetricsPath,
		DefaultPathNormalizer(
			"/api/public/datasets/01234567-89ab-cdef-0123-456789abcdef/items",
		),
	)
	assert.Equal(
		t,
		unmatchedMetricsPath,
		DefaultPathNormalizer("/arbitrary/high-cardinality/value"),
	)
}

func TestMetricsPathUsesRegisteredRoutePattern(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	var path string
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		path = metricsPath(c, DefaultPathNormalizer)
		return err
	})
	app.Get("/api/public/datasets/:datasetId/items/:itemId", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/api/public/datasets/01234567-89ab-cdef-0123-456789abcdef/items/fedcba98-7654-3210-fedc-ba9876543210",
		nil,
	))
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	assert.Equal(t, fiber.StatusNoContent, response.StatusCode)
	assert.Equal(t, "/api/public/datasets/:datasetId/items/:itemId", path)
}

func TestMetricsPathCollapsesUnmatchedRequests(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	var path string
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		path = metricsPath(c, DefaultPathNormalizer)
		return err
	})

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		fiber.MethodGet,
		"/attacker-controlled/01234567-89ab-cdef-0123-456789abcdef",
		nil,
	))
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	assert.Equal(t, fiber.StatusNotFound, response.StatusCode)
	assert.Equal(t, unmatchedMetricsPath, path)
}

func TestMetricsLabelsRemainStableAcrossReusedRequestBuffers(t *testing.T) {
	resetHTTPMetrics()
	t.Cleanup(resetHTTPMetrics)

	app := fiber.New()
	app.Use(NewMetricsMiddleware(DefaultMetricsConfig()).Handler())
	app.All("/resource", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	for _, method := range []string{fiber.MethodPost, fiber.MethodGet, fiber.MethodPost, fiber.MethodGet} {
		response, err := app.Test(httptest.NewRequestWithContext(
			context.Background(),
			method,
			"/resource",
			nil,
		))
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		assert.Equal(t, fiber.StatusCreated, response.StatusCode)
	}

	metrics, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	for _, family := range metrics {
		if family.GetName() != "agenttrace_http_requests_total" {
			continue
		}
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "method" {
					assert.Contains(t, []string{fiber.MethodGet, fiber.MethodPost}, label.GetValue())
				}
			}
		}
	}
}

func resetHTTPMetrics() {
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()
	httpRequestSize.Reset()
	httpResponseSize.Reset()
	httpActiveRequests.Reset()
}
