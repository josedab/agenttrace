package handler

import (
	"context"
	"testing"

	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// runInCtx executes fn inside a live Fiber request context so the *fiber.Ctx
// locals behave exactly as they do in production.
func runInCtx(t *testing.T, setup func(c *fiber.Ctx), fn func(c *fiber.Ctx)) {
	t.Helper()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		if setup != nil {
			setup(c)
		}
		fn(c)
		return c.SendStatus(200)
	})
	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func TestRoadmapActorID(t *testing.T) {
	t.Parallel()

	t.Run("resolves to the authenticated user", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		runInCtx(t, func(c *fiber.Ctx) {
			c.Locals(string(middleware.ContextKeyUserID), userID)
		}, func(c *fiber.Ctx) {
			got, ok := roadmapActorID(c)
			assert.True(t, ok)
			assert.Equal(t, userID, got)
		})
	})

	t.Run("does not fall back to the API key UUID for unowned keys", func(t *testing.T) {
		t.Parallel()
		apiKeyID := uuid.New()
		runInCtx(t, func(c *fiber.Ctx) {
			// Simulate an unowned API key: key id present, but no user context.
			c.Locals(string(middleware.ContextKeyAPIKeyID), apiKeyID)
		}, func(c *fiber.Ctx) {
			got, ok := roadmapActorID(c)
			assert.False(t, ok, "unowned key must not be treated as a user actor")
			assert.Equal(t, uuid.Nil, got)
		})
	})
}

func TestRoadmapAttributionID(t *testing.T) {
	t.Parallel()

	t.Run("prefers the authenticated user", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		apiKeyID := uuid.New()
		runInCtx(t, func(c *fiber.Ctx) {
			c.Locals(string(middleware.ContextKeyUserID), userID)
			c.Locals(string(middleware.ContextKeyAPIKeyID), apiKeyID)
		}, func(c *fiber.Ctx) {
			got, ok := roadmapAttributionID(c)
			assert.True(t, ok)
			assert.Equal(t, userID, got)
		})
	})

	t.Run("falls back to the API key identity", func(t *testing.T) {
		t.Parallel()
		apiKeyID := uuid.New()
		runInCtx(t, func(c *fiber.Ctx) {
			c.Locals(string(middleware.ContextKeyAPIKeyID), apiKeyID)
		}, func(c *fiber.Ctx) {
			got, ok := roadmapAttributionID(c)
			assert.True(t, ok, "share-link attribution may use the API key identity")
			assert.Equal(t, apiKeyID, got)
		})
	})
}
