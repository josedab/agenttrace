package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantMiddleware_Disabled(t *testing.T) {
	app := fiber.New()
	mw := NewTenantMiddleware(nil, nil, false)

	app.Use(mw.EnforceTraceLimit())
	app.Post("/ingest", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("POST", "/ingest", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ok", string(body))
}

func TestTenantMiddleware_NoTenantContext(t *testing.T) {
	app := fiber.New()
	mw := NewTenantMiddleware(nil, nil, true)

	app.Use(mw.EnforceTraceLimit())
	app.Post("/ingest", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("POST", "/ingest", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetTenantID(t *testing.T) {
	app := fiber.New()
	tenantID := uuid.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals(string(ContextKeyTenantID), tenantID)
		return c.Next()
	})
	app.Get("/test", func(c *fiber.Ctx) error {
		got, ok := GetTenantID(c)
		if !ok {
			return c.SendStatus(500)
		}
		return c.SendString(got.String())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, tenantID.String(), string(body))
}

func TestGetTenantID_Missing(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		_, ok := GetTenantID(c)
		if !ok {
			return c.SendString("no-tenant")
		}
		return c.SendString("has-tenant")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "no-tenant", string(body))
}
