package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type CostAttributionHandler struct {
	service *service.CostAttributionService
	logger  *zap.Logger
}

func NewCostAttributionHandler(svc *service.CostAttributionService, logger *zap.Logger) *CostAttributionHandler {
	return &CostAttributionHandler{service: svc, logger: logger}
}

func (h *CostAttributionHandler) Attribute(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AttributionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	attr, err := h.service.Attribute(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to attribute cost", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to attribute cost"})
	}

	return c.Status(fiber.StatusCreated).JSON(attr)
}

func (h *CostAttributionHandler) GetReport(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	period := domain.AttributionDateRange{
		From: c.Query("from", "2024-01-01"),
		To:   c.Query("to", "2024-12-31"),
	}

	report, err := h.service.GetReport(c.Context(), projectID, period)
	if err != nil {
		h.logger.Error("failed to get attribution report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get report"})
	}

	return c.JSON(report)
}

func (h *CostAttributionHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	attrs, err := h.service.ListAttributions(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list attributions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list attributions"})
	}

	return c.JSON(attrs)
}
