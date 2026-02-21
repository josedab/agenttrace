package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PredictionHandler handles predictive health monitoring HTTP requests
type PredictionHandler struct {
	predictionService *service.PredictionService
	logger            *zap.Logger
}

// NewPredictionHandler creates a new prediction handler
func NewPredictionHandler(predictionService *service.PredictionService, logger *zap.Logger) *PredictionHandler {
	return &PredictionHandler{
		predictionService: predictionService,
		logger:            logger,
	}
}

// AnalyzeHealth handles GET /health/analyze
func (h *PredictionHandler) AnalyzeHealth(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	dashboard, err := h.predictionService.AnalyzeHealth(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to analyze health",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to analyze health",
		})
	}

	return c.JSON(dashboard)
}

// GetPredictions handles GET /health/predictions
func (h *PredictionHandler) GetPredictions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	predictions, err := h.predictionService.GetPredictions(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get predictions",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get predictions",
		})
	}

	return c.JSON(predictions)
}

// GetTrend handles GET /health/trends/:metricName?days=7
func (h *PredictionHandler) GetTrend(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	metricName := c.Params("metricName")
	if metricName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Metric name is required",
		})
	}

	days := parseIntParam(c, "days", 7)

	trend, err := h.predictionService.GetTrend(c.Context(), projectID, metricName, days)
	if err != nil {
		h.logger.Error("failed to get trend",
			zap.String("projectId", projectID.String()),
			zap.String("metricName", metricName),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trend",
		})
	}

	return c.JSON(trend)
}
