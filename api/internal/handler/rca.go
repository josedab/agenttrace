package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RCAHandler handles root cause analysis HTTP requests
type RCAHandler struct {
	service *service.RCAService
	logger  *zap.Logger
}

// NewRCAHandler creates a new RCA handler
func NewRCAHandler(svc *service.RCAService, logger *zap.Logger) *RCAHandler {
	return &RCAHandler{
		service: svc,
		logger:  logger,
	}
}

// Analyze handles POST /api/public/rca/analyze
func (h *RCAHandler) Analyze(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.RCAInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TraceID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "traceId is required"})
	}

	report, err := h.service.AnalyzeTrace(c.Context(), projectID, input.TraceID)
	if err != nil {
		h.logger.Error("failed to analyze trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to analyze trace"})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetReport handles GET /api/public/rca/reports/:reportId
func (h *RCAHandler) GetReport(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	reportID, err := uuid.Parse(c.Params("reportId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid report ID"})
	}

	report, err := h.service.GetReport(c.Context(), reportID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
	}

	return c.JSON(report)
}

// ListReports handles GET /api/public/rca/reports
func (h *RCAHandler) ListReports(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	reports, err := h.service.ListReports(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list RCA reports", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list reports"})
	}

	return c.JSON(fiber.Map{"reports": reports, "count": len(reports)})
}
