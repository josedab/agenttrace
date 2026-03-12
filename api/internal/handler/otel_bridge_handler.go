package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OTelBridgeHandler handles OpenTelemetry-native trace bridge HTTP requests
type OTelBridgeHandler struct {
	service *service.OTelBridgeService
	logger  *zap.Logger
}

// NewOTelBridgeHandler creates a new OTel bridge handler
func NewOTelBridgeHandler(svc *service.OTelBridgeService, logger *zap.Logger) *OTelBridgeHandler {
	return &OTelBridgeHandler{
		service: svc,
		logger:  logger,
	}
}

// GetConfig handles GET /otel-bridge/config
func (h *OTelBridgeHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get OTel bridge config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get OTel bridge config"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /otel-bridge/config
func (h *OTelBridgeHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.OTelBridgeConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to update OTel bridge config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update OTel bridge config"})
	}

	return c.JSON(config)
}

// ListDestinations handles GET /otel-bridge/destinations
func (h *OTelBridgeHandler) ListDestinations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.ListDestinations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list OTel destinations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list OTel destinations"})
	}

	return c.JSON(result)
}

// AddDestination handles POST /otel-bridge/destinations
func (h *OTelBridgeHandler) AddDestination(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.OTelDestinationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" || input.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and endpoint are required"})
	}

	result, err := h.service.AddDestination(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to add OTel destination", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add OTel destination"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// RemoveDestination handles DELETE /otel-bridge/destinations/:destId
func (h *OTelBridgeHandler) RemoveDestination(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	destIDStr := c.Params("destId")
	if destIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Destination ID is required"})
	}

	destID, err := uuid.Parse(destIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid destination ID"})
	}

	err = h.service.RemoveDestination(c.Context(), destID)
	if err != nil {
		h.logger.Error("failed to remove OTel destination", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove OTel destination"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ImportSpans handles POST /otel-bridge/import
func (h *OTelBridgeHandler) ImportSpans(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.OTelImportRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.ImportSpans(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to import OTel spans", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to import OTel spans"})
	}

	return c.JSON(result)
}

// GetStats handles GET /otel-bridge/stats
func (h *OTelBridgeHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get OTel bridge stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get OTel bridge stats"})
	}

	return c.JSON(stats)
}
