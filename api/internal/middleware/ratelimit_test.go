package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRateLimitTest(t *testing.T, cfg RateLimitConfig) (*fiber.App, *RateLimitMiddleware, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimitMiddleware(client, cfg)

	app := fiber.New()
	return app, rl, mr
}

func TestRateLimitMiddleware_RequestsUnderLimit(t *testing.T) {
	cfg := RateLimitConfig{
		Max:    5,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "test-key"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limited"})
		},
	}
	app, rl, _ := setupRateLimitTest(t, cfg)
	app.Get("/test", rl.Handler(), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Make 3 requests (under limit of 5)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRateLimitMiddleware_RequestAtMaxBlocked(t *testing.T) {
	cfg := RateLimitConfig{
		Max:    3,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "test-key"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limited"})
		},
	}
	app, rl, _ := setupRateLimitTest(t, cfg)
	app.Get("/test", rl.Handler(), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Use up the limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Next request should be blocked
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	cfg := RateLimitConfig{
		Max:    10,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "test-key"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).SendString("limited")
		},
	}
	app, rl, _ := setupRateLimitTest(t, cfg)
	app.Get("/test", rl.Handler(), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check X-RateLimit-Remaining header
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	assert.Equal(t, "9", remaining) // 10 max - 1 used = 9

	// Check X-RateLimit-Limit header
	limit := resp.Header.Get("X-RateLimit-Limit")
	assert.Equal(t, "10", limit)

	// Check X-RateLimit-Reset is set and parseable
	resetStr := resp.Header.Get("X-RateLimit-Reset")
	assert.NotEmpty(t, resetStr)
	resetVal, err := strconv.ParseInt(resetStr, 10, 64)
	require.NoError(t, err)
	assert.True(t, resetVal > time.Now().Unix())
}

func TestRateLimitMiddleware_SkipFunction(t *testing.T) {
	cfg := RateLimitConfig{
		Max:    1,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "test-key"
		},
		Skip: func(c *fiber.Ctx) bool {
			return c.Path() == "/skip"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).SendString("limited")
		},
	}
	app, rl, _ := setupRateLimitTest(t, cfg)
	handler := rl.Handler()
	app.Get("/test", handler, func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/skip", handler, func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Use up the limit on /test
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	resp.Body.Close()

	// /skip should bypass limit
	for i := 0; i < 5; i++ {
		req = httptest.NewRequest(http.MethodGet, "/skip", nil)
		resp, err = app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRateLimitMiddleware_CustomKeyGenerator(t *testing.T) {
	cfg := RateLimitConfig{
		Max:    2,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("X-User-ID", "anonymous")
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).SendString("limited")
		},
	}
	app, rl, _ := setupRateLimitTest(t, cfg)
	app.Get("/test", rl.Handler(), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// User A uses up limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "user-a")
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		resp.Body.Close()
	}

	// User A is blocked
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user-a")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	resp.Body.Close()

	// User B can still make requests
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user-b")
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestRateLimitMiddleware_RedisUnavailable_FailOpen(t *testing.T) {
	// Create middleware with bad Redis connection
	client := redis.NewClient(&redis.Options{Addr: "localhost:1"}) // invalid port
	rl := NewRateLimitMiddleware(client, RateLimitConfig{
		Max:    1,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "test-key"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).SendString("limited")
		},
	})

	app := fiber.New()
	app.Get("/test", rl.Handler(), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// When Redis is unavailable, the middleware fails open (allows request)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}

func TestRateLimitMiddleware_BurstRateLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimitMiddleware(client)

	app := fiber.New()
	app.Get("/test", rl.BurstRateLimit(3, 1.0), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Should allow 3 burst requests (initial tokens = maxTokens)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "request %d should pass", i+1)
		resp.Body.Close()
	}

	// 4th request should be blocked (no tokens left, insufficient refill time)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	resp.Body.Close()
}
