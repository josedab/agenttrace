package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// StreamingDashboardHandler handles real-time streaming dashboard HTTP requests
type StreamingDashboardHandler struct {
	service *service.StreamingDashboardService
	logger  *zap.Logger
}

// NewStreamingDashboardHandler creates a new streaming dashboard handler
func NewStreamingDashboardHandler(svc *service.StreamingDashboardService, logger *zap.Logger) *StreamingDashboardHandler {
	return &StreamingDashboardHandler{
		service: svc,
		logger:  logger,
	}
}

// GetDashboard handles GET /streaming-dashboard
func (h *StreamingDashboardHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.GetDashboard(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(result)
}

// WebSocketHandler handles GET /streaming-dashboard/ws
func (h *StreamingDashboardHandler) WebSocketHandler(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	err := h.service.HandleWebSocket(c.Context(), projectID.String(), c)
	if err != nil {
		h.logger.Error("websocket handler error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "WebSocket connection failed"})
	}

	return nil
}

// GetConfig handles GET /streaming-dashboard/config
func (h *StreamingDashboardHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get dashboard config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard config"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /streaming-dashboard/config
func (h *StreamingDashboardHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.DashboardConfig
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to update dashboard config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update dashboard config"})
	}

	return c.JSON(config)
}
