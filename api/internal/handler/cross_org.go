package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type CrossOrgHandler struct {
	service *service.CrossOrgService
	logger  *zap.Logger
}

func NewCrossOrgHandler(svc *service.CrossOrgService, logger *zap.Logger) *CrossOrgHandler {
	return &CrossOrgHandler{service: svc, logger: logger}
}

func (h *CrossOrgHandler) Submit(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CrossOrgSubmissionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	submission, err := h.service.Submit(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to submit cross-org benchmark", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit"})
	}

	return c.Status(fiber.StatusCreated).JSON(submission)
}

func (h *CrossOrgHandler) GetReport(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	report, err := h.service.GetReport(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get cross-org report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get report"})
	}

	return c.JSON(report)
}

func (h *CrossOrgHandler) GetIndustryStats(c *fiber.Ctx) error {
	category := c.Params("category")
	stats, err := h.service.GetIndustryStats(c.Context(), category)
	if err != nil {
		h.logger.Error("failed to get industry stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get industry stats"})
	}

	return c.JSON(stats)
}
