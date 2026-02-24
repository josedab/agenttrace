package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type AutonomyHandler struct {
	service *service.AutonomyService
	logger  *zap.Logger
}

func NewAutonomyHandler(svc *service.AutonomyService, logger *zap.Logger) *AutonomyHandler {
	return &AutonomyHandler{service: svc, logger: logger}
}

func (h *AutonomyHandler) SetAutonomy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AutonomyConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.SetAutonomy(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to set autonomy", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

func (h *AutonomyHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	agentName := c.Params("agentName")
	config, err := h.service.GetConfig(c.Context(), projectID, agentName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(config)
}

func (h *AutonomyHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get autonomy dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}

func (h *AutonomyHandler) GetTrustEvolution(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	agentName := c.Params("agentName")
	evolution, err := h.service.GetTrustEvolution(c.Context(), projectID, agentName)
	if err != nil {
		h.logger.Error("failed to get trust evolution", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get trust evolution"})
	}

	return c.JSON(evolution)
}
