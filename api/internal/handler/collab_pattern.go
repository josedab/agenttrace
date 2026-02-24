package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CollabPatternHandler handles collaboration pattern HTTP requests
type CollabPatternHandler struct {
	service *service.CollabPatternService
	logger  *zap.Logger
}

// NewCollabPatternHandler creates a new collaboration pattern handler
func NewCollabPatternHandler(svc *service.CollabPatternService, logger *zap.Logger) *CollabPatternHandler {
	return &CollabPatternHandler{
		service: svc,
		logger:  logger,
	}
}

// List handles GET /api/public/collab-patterns
func (h *CollabPatternHandler) List(c *fiber.Ctx) error {
	patterns, err := h.service.ListPatterns(c.Context())
	if err != nil {
		h.logger.Error("failed to list patterns", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list patterns"})
	}

	return c.JSON(patterns)
}

// Get handles GET /api/public/collab-patterns/:patternId
func (h *CollabPatternHandler) Get(c *fiber.Ctx) error {
	patternID := c.Params("patternId")

	pattern, err := h.service.GetPattern(c.Context(), patternID)
	if err != nil {
		h.logger.Error("failed to get pattern", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pattern not found"})
	}

	return c.JSON(pattern)
}

// Deploy handles POST /api/public/collab-patterns/:patternId/deploy
func (h *CollabPatternHandler) Deploy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.DeployPatternInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	patternID := c.Params("patternId")
	_ = patternID // patternID from URL; input.PatternID used by service

	deployment, err := h.service.DeployPattern(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to deploy pattern", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to deploy pattern"})
	}

	return c.Status(fiber.StatusCreated).JSON(deployment)
}

// GetDeployments handles GET /api/public/collab-patterns/deployments
func (h *CollabPatternHandler) GetDeployments(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	deployments, err := h.service.GetDeployments(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get deployments", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get deployments"})
	}

	return c.JSON(deployments)
}

// GetAnalytics handles GET /api/public/collab-patterns/:patternId/analytics
func (h *CollabPatternHandler) GetAnalytics(c *fiber.Ctx) error {
	patternID := c.Params("patternId")

	analytics, err := h.service.GetPatternAnalytics(c.Context(), patternID)
	if err != nil {
		h.logger.Error("failed to get pattern analytics", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pattern not found"})
	}

	return c.JSON(analytics)
}
