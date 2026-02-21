package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CostOptimizerHandler handles cost optimization HTTP requests
type CostOptimizerHandler struct {
	costOptimizerService *service.CostOptimizerService
	logger               *zap.Logger
}

// NewCostOptimizerHandler creates a new cost optimizer handler
func NewCostOptimizerHandler(costOptimizerService *service.CostOptimizerService, logger *zap.Logger) *CostOptimizerHandler {
	return &CostOptimizerHandler{
		costOptimizerService: costOptimizerService,
		logger:               logger,
	}
}

// Analyze handles GET /cost-optimizer/analyze?dateRange=7d
func (h *CostOptimizerHandler) Analyze(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	dateRange := parseDateRange(c.Query("dateRange", "7d"))

	analysis, err := h.costOptimizerService.Analyze(c.Context(), projectID, dateRange)
	if err != nil {
		h.logger.Error("failed to analyze costs",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to analyze costs",
		})
	}

	return c.JSON(analysis)
}

// GetRecommendations handles GET /cost-optimizer/recommendations
func (h *CostOptimizerHandler) GetRecommendations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	recs, err := h.costOptimizerService.GetRecommendations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get recommendations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get recommendations",
		})
	}

	return c.JSON(recs)
}

// ApplyRecommendation handles POST /cost-optimizer/recommendations/:id/apply
func (h *CostOptimizerHandler) ApplyRecommendation(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	recID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid recommendation ID",
		})
	}

	if err := h.costOptimizerService.ApplyRecommendation(c.Context(), recID); err != nil {
		h.logger.Error("failed to apply recommendation",
			zap.String("recommendationId", recID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to apply recommendation",
		})
	}

	return c.JSON(fiber.Map{"status": "applied"})
}

// DismissRecommendation handles POST /cost-optimizer/recommendations/:id/dismiss
func (h *CostOptimizerHandler) DismissRecommendation(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	recID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid recommendation ID",
		})
	}

	if err := h.costOptimizerService.DismissRecommendation(c.Context(), recID); err != nil {
		h.logger.Error("failed to dismiss recommendation",
			zap.String("recommendationId", recID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to dismiss recommendation",
		})
	}

	return c.JSON(fiber.Map{"status": "dismissed"})
}

// parseDateRange converts a date range string like "7d", "30d" into a domain.DateRange
func parseDateRange(s string) domain.DateRange {
	now := time.Now()
	days := 7 // default

	if len(s) > 1 && s[len(s)-1] == 'd' {
		val := 0
		for _, ch := range s[:len(s)-1] {
			if ch >= '0' && ch <= '9' {
				val = val*10 + int(ch-'0')
			}
		}
		if val > 0 {
			days = val
		}
	}

	return domain.DateRange{
		Start: now.AddDate(0, 0, -days),
		End:   now,
	}
}
