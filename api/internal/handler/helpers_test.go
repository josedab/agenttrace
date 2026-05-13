package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

func TestRequireProjectIDStopsHandlerWithoutProjectContext(t *testing.T) {
	app := fiber.New()
	executed := false
	app.Get("/resource", func(c *fiber.Ctx) error {
		projectID, err := RequireProjectID(c)
		if err != nil {
			return err
		}
		executed = true
		return c.JSON(fiber.Map{"projectId": projectID.String()})
	})

	response, err := app.Test(
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/resource", nil),
		-1,
	)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	// The body must not run: it would otherwise query and mutate data with an
	// empty project ID.
	assert.False(t, executed)
}

func TestRequireProjectIDReturnsRequestProject(t *testing.T) {
	app := fiber.New()
	projectID := uuid.New()
	app.Get("/resource", func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), projectID)
		resolved, err := RequireProjectID(c)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"projectId": resolved.String()})
	})

	response, err := app.Test(
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/resource", nil),
		-1,
	)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestRequireUserIDStopsHandlerWithoutUserContext(t *testing.T) {
	app := fiber.New()
	executed := false
	app.Get("/resource", func(c *fiber.Ctx) error {
		if _, err := RequireUserID(c); err != nil {
			return err
		}
		executed = true
		return c.SendStatus(fiber.StatusOK)
	})

	response, err := app.Test(
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/resource", nil),
		-1,
	)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	assert.False(t, executed)
}
