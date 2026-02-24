package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TeamIntelligenceHandler handles team intelligence HTTP requests
type TeamIntelligenceHandler struct {
	logger  *zap.Logger
	service *service.TeamIntelligenceService
}

// NewTeamIntelligenceHandler creates a new team intelligence handler
func NewTeamIntelligenceHandler(logger *zap.Logger, svc *service.TeamIntelligenceService) *TeamIntelligenceHandler {
	return &TeamIntelligenceHandler{
		logger:  logger,
		service: svc,
	}
}

// GetDashboard handles GET /api/public/team/dashboard
func (h *TeamIntelligenceHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	filter := domain.TeamDashboardFilter{
		ProjectID: projectID.String(),
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID, filter)
	if err != nil {
		h.logger.Error("failed to get team dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get team dashboard"})
	}

	return c.JSON(dashboard)
}

// CalculateROI handles GET /api/public/team/roi
func (h *TeamIntelligenceHandler) CalculateROI(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	hourlyRate := 75.0
	if rateStr := c.Query("hourlyRate"); rateStr != "" {
		if parsed, err := strconv.ParseFloat(rateStr, 64); err == nil {
			hourlyRate = parsed
		}
	}

	roi, err := h.service.CalculateROI(c.Context(), projectID, hourlyRate)
	if err != nil {
		h.logger.Error("failed to calculate ROI", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate ROI"})
	}

	return c.JSON(roi)
}
