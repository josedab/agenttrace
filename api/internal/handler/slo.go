package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// SLOHandler handles agent performance SLO HTTP requests
type SLOHandler struct {
	service *service.SLOService
	logger  *zap.Logger
}

// NewSLOHandler creates a new SLO handler
func NewSLOHandler(svc *service.SLOService, logger *zap.Logger) *SLOHandler {
	return &SLOHandler{
		service: svc,
		logger:  logger,
	}
}

// Create handles POST /api/public/slos
func (h *SLOHandler) Create(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.SLOInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" || input.Metric == "" || input.AgentName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agentName, name, and metric are required"})
	}

	slo, err := h.service.CreateSLO(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to create SLO", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create SLO"})
	}

	return c.Status(fiber.StatusCreated).JSON(slo)
}

// List handles GET /api/public/slos
func (h *SLOHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	slos, err := h.service.ListSLOs(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to list SLOs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list SLOs"})
	}

	return c.JSON(slos)
}

// GetStatus handles GET /api/public/slos/:sloId/status
func (h *SLOHandler) GetStatus(c *fiber.Ctx) error {
	sloID := c.Params("sloId")
	if sloID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SLO ID is required"})
	}

	status, err := h.service.GetSLOStatus(c.Context(), sloID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(status)
}

// GetReport handles GET /api/public/slos/report
func (h *SLOHandler) GetReport(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	report, err := h.service.GetReport(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get SLO report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get SLO report"})
	}

	return c.JSON(report)
}

// GetHistory handles GET /api/public/slos/:sloId/history
func (h *SLOHandler) GetHistory(c *fiber.Ctx) error {
	sloID := c.Params("sloId")
	if sloID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SLO ID is required"})
	}

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	history, err := h.service.GetHistory(c.Context(), sloID, from, to)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(history)
}
