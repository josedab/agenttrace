package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// FederatedLearningHandler handles federated learning HTTP requests
type FederatedLearningHandler struct {
	service *service.FederatedLearningService
	logger  *zap.Logger
}

// NewFederatedLearningHandler creates a new federated learning handler
func NewFederatedLearningHandler(svc *service.FederatedLearningService, logger *zap.Logger) *FederatedLearningHandler {
	return &FederatedLearningHandler{
		service: svc,
		logger:  logger,
	}
}

// JoinRing handles POST /api/public/federated/rings/join
func (h *FederatedLearningHandler) JoinRing(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.FederationJoinInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ring, err := h.service.JoinRing(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to join federation ring", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to join ring"})
	}

	return c.Status(fiber.StatusCreated).JSON(ring)
}

// ListRings handles GET /api/public/federated/rings
func (h *FederatedLearningHandler) ListRings(c *fiber.Ctx) error {
	rings, err := h.service.ListRings(c.Context())
	if err != nil {
		h.logger.Error("failed to list federation rings", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list rings"})
	}

	return c.JSON(rings)
}

// GetInsights handles GET /api/public/federated/rings/:ringId/insights
func (h *FederatedLearningHandler) GetInsights(c *fiber.Ctx) error {
	ringID := c.Params("ringId")

	insights, err := h.service.GetInsights(c.Context(), ringID)
	if err != nil {
		h.logger.Error("failed to get federation insights", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Ring not found"})
	}

	return c.JSON(insights)
}

// GetConfig handles GET /api/public/federated/config
func (h *FederatedLearningHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get federation config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get config"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /api/public/federated/config
func (h *FederatedLearningHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var config domain.FederationConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.UpdateConfig(c.Context(), projectID.String(), &config)
	if err != nil {
		h.logger.Error("failed to update federation config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update config"})
	}

	return c.JSON(result)
}
