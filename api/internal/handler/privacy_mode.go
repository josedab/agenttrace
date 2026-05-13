package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PrivacyModeHandler exposes actual runtime privacy capabilities.
type PrivacyModeHandler struct {
	policy *service.EgressPolicy
}

// NewPrivacyModeHandler creates a privacy capability handler.
func NewPrivacyModeHandler(policy *service.EgressPolicy) *PrivacyModeHandler {
	return &PrivacyModeHandler{policy: policy}
}

// GetCapabilities handles GET /privacy/capabilities.
func (h *PrivacyModeHandler) GetCapabilities(c *fiber.Ctx) error {
	if _, ok := middleware.GetProjectID(c); !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}
	return c.JSON(h.policy.Capabilities())
}
