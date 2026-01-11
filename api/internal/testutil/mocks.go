// Package testutil provides shared test utilities for the AgentTrace API.
package testutil

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// TestProjectMiddleware creates a middleware that sets the project ID in context.
// Use this in tests to simulate authenticated requests.
func TestProjectMiddleware(projectID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), projectID)
		return c.Next()
	}
}

// TestUserMiddleware creates a middleware that sets the user ID in context.
// Use this in tests to simulate authenticated requests.
func TestUserMiddleware(userID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyUserID), userID)
		return c.Next()
	}
}

// TestAuthMiddleware creates a middleware that sets both project and user IDs in context.
// Use this in tests to simulate fully authenticated requests.
func TestAuthMiddleware(projectID, userID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), projectID)
		c.Locals(string(middleware.ContextKeyUserID), userID)
		return c.Next()
	}
}

// TestAPIKeyMiddleware creates a middleware that sets the API key ID in context.
// Use this in tests to simulate API key authenticated requests.
func TestAPIKeyMiddleware(apiKeyID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyAPIKeyID), apiKeyID)
		return c.Next()
	}
}
