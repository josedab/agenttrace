package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CarbonHandler handles energy and carbon tracking HTTP requests
type CarbonHandler struct {
	service *service.CarbonService
	logger  *zap.Logger
}

// NewCarbonHandler creates a new carbon handler
func NewCarbonHandler(svc *service.CarbonService, logger *zap.Logger) *CarbonHandler {
	return &CarbonHandler{
		service: svc,
		logger:  logger,
	}
}

// GetFootprint handles GET /api/public/carbon/footprint
func (h *CarbonHandler) GetFootprint(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	period := domain.CarbonDateRange{
		From: c.Query("from", "2024-01-01"),
		To:   c.Query("to", "2024-12-31"),
	}

	footprint, err := h.service.GetFootprint(c.Context(), projectID.String(), period)
	if err != nil {
		h.logger.Error("failed to get carbon footprint", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get carbon footprint"})
	}

	return c.JSON(footprint)
}

// GetConfig handles GET /api/public/carbon/config
func (h *CarbonHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get carbon config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get carbon config"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /api/public/carbon/config
func (h *CarbonHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CarbonConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to update carbon config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update carbon config"})
	}

	return c.JSON(config)
}

// GetSuggestions handles GET /api/public/carbon/suggestions
func (h *CarbonHandler) GetSuggestions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	suggestions, err := h.service.GetSuggestions(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get carbon suggestions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get suggestions"})
	}

	return c.JSON(suggestions)
}
