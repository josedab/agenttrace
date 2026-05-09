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

// CustomMetricsHandler handles custom metrics HTTP requests
type CustomMetricsHandler struct {
	service *service.CustomMetricsService
	logger  *zap.Logger
}

// NewCustomMetricsHandler creates a new custom metrics handler
func NewCustomMetricsHandler(svc *service.CustomMetricsService, logger *zap.Logger) *CustomMetricsHandler {
	return &CustomMetricsHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateMetric handles POST /api/public/custom-metrics
func (h *CustomMetricsHandler) CreateMetric(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CustomMetricInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	metric, err := h.service.CreateMetric(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create metric", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create metric"})
	}

	return c.Status(fiber.StatusCreated).JSON(metric)
}

// ListMetrics handles GET /api/public/custom-metrics
func (h *CustomMetricsHandler) ListMetrics(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	metrics, err := h.service.ListMetrics(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list metrics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list metrics"})
	}

	return c.JSON(metrics)
}

// GetValues handles GET /api/public/custom-metrics/:metricId/values
func (h *CustomMetricsHandler) GetValues(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	metricID, err := uuid.Parse(c.Params("metricId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid metric ID"})
	}

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	values, err := h.service.GetMetricValues(c.Context(), metricID, from, to)
	if err != nil {
		h.logger.Error("failed to get metric values", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get metric values"})
	}

	return c.JSON(values)
}

// CreateDashboard handles POST /api/public/custom-metrics/dashboards
func (h *CustomMetricsHandler) CreateDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input struct {
		Name    string                   `json:"name"`
		Widgets []domain.DashboardWidget `json:"widgets"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	dashboard, err := h.service.CreateDashboard(c.Context(), projectID, input.Name, input.Widgets)
	if err != nil {
		h.logger.Error("failed to create dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create dashboard"})
	}

	return c.Status(fiber.StatusCreated).JSON(dashboard)
}

// ListDashboards handles GET /api/public/custom-metrics/dashboards
func (h *CustomMetricsHandler) ListDashboards(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboards, err := h.service.ListDashboards(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list dashboards", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list dashboards"})
	}

	return c.JSON(dashboards)
}

// CreateAlert handles POST /api/public/custom-metrics/alerts
func (h *CustomMetricsHandler) CreateAlert(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.MetricAlertInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	alert, err := h.service.CreateAlert(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create alert", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create alert"})
	}

	return c.Status(fiber.StatusCreated).JSON(alert)
}

// ListAlerts handles GET /api/public/custom-metrics/alerts
func (h *CustomMetricsHandler) ListAlerts(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	alerts, err := h.service.ListAlerts(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list alerts", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list alerts"})
	}

	return c.JSON(alerts)
}
