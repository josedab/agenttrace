package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// MobileHandler handles mobile companion app HTTP requests
type MobileHandler struct {
	service *service.MobileService
	logger  *zap.Logger
}

// NewMobileHandler creates a new mobile handler
func NewMobileHandler(svc *service.MobileService, logger *zap.Logger) *MobileHandler {
	return &MobileHandler{
		service: svc,
		logger:  logger,
	}
}

// RegisterDevice handles POST /api/public/mobile/devices
func (h *MobileHandler) RegisterDevice(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.MobileDeviceInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Platform == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Platform is required"})
	}
	if input.PushToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Push token is required"})
	}

	device, err := h.service.RegisterDevice(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to register device", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register device"})
	}

	return c.Status(fiber.StatusCreated).JSON(device)
}

// GetDashboard handles GET /api/public/mobile/dashboard
func (h *MobileHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get mobile dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}

// ListNotifications handles GET /api/public/mobile/notifications
func (h *MobileHandler) ListNotifications(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	notifications, err := h.service.ListNotifications(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list notifications", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list notifications"})
	}

	if notifications == nil {
		notifications = []domain.PushNotification{}
	}
	return c.JSON(notifications)
}
