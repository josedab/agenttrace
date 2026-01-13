package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRFMiddleware(t *testing.T) {
	t.Run("allows GET requests without token", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("sets CSRF cookie on GET request", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Check for CSRF cookie
		cookies := resp.Cookies()
		var csrfCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == CSRFCookieName {
				csrfCookie = cookie
				break
			}
		}
		assert.NotNil(t, csrfCookie)
		assert.NotEmpty(t, csrfCookie.Value)
	})

	t.Run("blocks POST request without token", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("POST", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "CSRF")
	})

	t.Run("allows POST request with valid token", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		// First, get a CSRF token via GET
		getReq := httptest.NewRequest("GET", "/test", nil)
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})
		getResp, err := app.Test(getReq)
		require.NoError(t, err)

		var csrfToken string
		for _, cookie := range getResp.Cookies() {
			if cookie.Name == CSRFCookieName {
				csrfToken = cookie.Value
				break
			}
		}
		require.NotEmpty(t, csrfToken)

		// Now make POST request with token
		postReq := httptest.NewRequest("POST", "/test", nil)
		postReq.Header.Set("Cookie", CSRFCookieName+"="+csrfToken)
		postReq.Header.Set(CSRFHeaderName, csrfToken)

		postResp, err := app.Test(postReq)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, postResp.StatusCode)
	})

	t.Run("blocks POST request with mismatched token", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Cookie", CSRFCookieName+"=valid-token")
		req.Header.Set(CSRFHeaderName, "different-token")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("skips validation for API key auth", func(t *testing.T) {
		app := fiber.New()

		// Middleware to set API key auth type
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)
			return c.Next()
		})

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("POST", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Should pass without CSRF token when using API key auth
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestCSRFMiddlewareDisabled(t *testing.T) {
	t.Run("passes all requests when disabled", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddlewareWithConfig(CSRFConfig{
			Enabled: false,
		})
		app.Use(csrf.Handler())
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("POST", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestCSRFGetToken(t *testing.T) {
	t.Run("returns CSRF token in JSON", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Get("/csrf", csrf.GetToken())

		req := httptest.NewRequest("GET", "/csrf", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "csrfToken")
	})
}

func TestDefaultCSRFConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := DefaultCSRFConfig()

		assert.Equal(t, CSRFCookieName, config.CookieName)
		assert.Equal(t, "/", config.CookiePath)
		assert.True(t, config.CookieSecure)
		assert.Equal(t, "Strict", config.CookieSameSite)
		assert.True(t, config.CookieHTTPOnly)
		assert.True(t, config.Enabled)
		assert.Contains(t, config.SkipMethods, "GET")
		assert.Contains(t, config.SkipMethods, "HEAD")
		assert.Contains(t, config.SkipMethods, "OPTIONS")
		assert.Contains(t, config.SkipMethods, "TRACE")
	})
}

func TestNewCSRFMiddlewareWithConfig(t *testing.T) {
	t.Run("applies custom config", func(t *testing.T) {
		config := CSRFConfig{
			CookieName:     "custom_csrf",
			CookiePath:     "/api",
			CookieSecure:   false,
			CookieSameSite: "Lax",
			Enabled:        true,
		}

		csrf := NewCSRFMiddlewareWithConfig(config)
		assert.NotNil(t, csrf)
	})

	t.Run("sets defaults for empty fields", func(t *testing.T) {
		config := CSRFConfig{
			Enabled: true,
			// Leave other fields empty
		}

		csrf := NewCSRFMiddlewareWithConfig(config)
		assert.NotNil(t, csrf)
		// Defaults should be set
		assert.Equal(t, CSRFCookieName, csrf.config.CookieName)
		assert.Equal(t, "/", csrf.config.CookiePath)
	})
}

func TestGetCSRFToken(t *testing.T) {
	t.Run("returns token from context", func(t *testing.T) {
		app := fiber.New()

		var extractedToken string
		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(CSRFContextKey, "test-token")
			extractedToken = GetCSRFToken(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, "test-token", extractedToken)
	})

	t.Run("returns empty string when not in context", func(t *testing.T) {
		app := fiber.New()

		var extractedToken string
		app.Get("/test", func(c *fiber.Ctx) error {
			extractedToken = GetCSRFToken(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.Empty(t, extractedToken)
	})
}

func TestCSRFConstants(t *testing.T) {
	t.Run("constant values", func(t *testing.T) {
		assert.Equal(t, 32, CSRFTokenLength)
		assert.Equal(t, "_csrf", CSRFCookieName)
		assert.Equal(t, "X-CSRF-Token", CSRFHeaderName)
		assert.Equal(t, "csrf_token", CSRFFormFieldName)
		assert.Equal(t, "csrfToken", CSRFContextKey)
	})
}

func TestCSRFFormField(t *testing.T) {
	t.Run("accepts token from form field", func(t *testing.T) {
		app := fiber.New()

		csrf := NewCSRFMiddleware()
		app.Use(csrf.Handler())
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		// First, get a CSRF token
		getReq := httptest.NewRequest("GET", "/test", nil)
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})
		getResp, err := app.Test(getReq)
		require.NoError(t, err)

		var csrfToken string
		for _, cookie := range getResp.Cookies() {
			if cookie.Name == CSRFCookieName {
				csrfToken = cookie.Value
				break
			}
		}
		require.NotEmpty(t, csrfToken)

		// Now make POST request with token in form field
		formData := strings.NewReader(CSRFFormFieldName + "=" + csrfToken)
		postReq := httptest.NewRequest("POST", "/test", formData)
		postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		postReq.Header.Set("Cookie", CSRFCookieName+"="+csrfToken)

		postResp, err := app.Test(postReq)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, postResp.StatusCode)
	})
}

func TestCSRFSkipMethods(t *testing.T) {
	methods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}

	for _, method := range methods {
		t.Run("skips "+method+" request", func(t *testing.T) {
			app := fiber.New()

			csrf := NewCSRFMiddleware()
			app.Use(csrf.Handler())
			app.Add(method, "/test", func(c *fiber.Ctx) error {
				return c.SendStatus(200)
			})

			req := httptest.NewRequest(method, "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}
}

func TestCSRFStateChangingMethods(t *testing.T) {
	methods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run("validates "+method+" request", func(t *testing.T) {
			app := fiber.New()

			csrf := NewCSRFMiddleware()
			app.Use(csrf.Handler())
			app.Add(method, "/test", func(c *fiber.Ctx) error {
				return c.SendStatus(200)
			})

			req := httptest.NewRequest(method, "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			// Without token, should be forbidden
			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}
}
