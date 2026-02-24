package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CostAlertingHandler handles cost alerting HTTP requests
type CostAlertingHandler struct {
	logger  *zap.Logger
	service *service.CostAlertingService
}

// NewCostAlertingHandler creates a new cost alerting handler
func NewCostAlertingHandler(svc *service.CostAlertingService, logger *zap.Logger) *CostAlertingHandler {
	return &CostAlertingHandler{
		logger:  logger,
		service: svc,
	}
}

// CreateAlertRule handles POST /api/public/cost-alerting/rules
// @Summary Create alert rule
// @Description Create a new cost alert rule
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Param rule body domain.CostAlertRule true "Alert rule"
// @Success 201 {object} domain.CostAlertRule
// @Failure 400 {object} map[string]string
// @Router /api/public/cost-alerting/rules [post]
func (h *CostAlertingHandler) CreateAlertRule(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var rule domain.CostAlertRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.CreateRule(c.Context(), projectID, rule)
	if err != nil {
		h.logger.Error("failed to create alert rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create alert rule"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ListAlertRules handles GET /api/public/cost-alerting/rules
// @Summary List alert rules
// @Description List all cost alert rules for a project
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Success 200 {array} domain.CostAlertRule
// @Failure 401 {object} map[string]string
// @Router /api/public/cost-alerting/rules [get]
func (h *CostAlertingHandler) ListAlertRules(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	rules, err := h.service.ListRules(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list alert rules", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list alert rules"})
	}

	return c.JSON(rules)
}

// DeleteAlertRule handles DELETE /api/public/cost-alerting/rules/:ruleId
// @Summary Delete alert rule
// @Description Delete a cost alert rule
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Param ruleId path string true "Rule ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/public/cost-alerting/rules/{ruleId} [delete]
func (h *CostAlertingHandler) DeleteAlertRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	if err := h.service.DeleteRule(c.Context(), ruleID); err != nil {
		h.logger.Error("failed to delete alert rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete alert rule"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListCostAlerts handles GET /api/public/cost-alerting/alerts
// @Summary List cost alerts
// @Description List all cost alerts for a project
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(100)
// @Success 200 {array} domain.CostAlert
// @Failure 401 {object} map[string]string
// @Router /api/public/cost-alerting/alerts [get]
func (h *CostAlertingHandler) ListCostAlerts(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	limit := c.QueryInt("limit", 100)

	alerts, err := h.service.ListAlerts(c.Context(), projectID, limit)
	if err != nil {
		h.logger.Error("failed to list cost alerts", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list cost alerts"})
	}

	return c.JSON(alerts)
}

// AcknowledgeCostAlert handles POST /api/public/cost-alerting/alerts/:alertId/acknowledge
// @Summary Acknowledge cost alert
// @Description Acknowledge a cost alert
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Param alertId path string true "Alert ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/cost-alerting/alerts/{alertId}/acknowledge [post]
func (h *CostAlertingHandler) AcknowledgeCostAlert(c *fiber.Ctx) error {
	alertID, err := uuid.Parse(c.Params("alertId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid alert ID"})
	}

	if err := h.service.AcknowledgeAlert(c.Context(), alertID); err != nil {
		h.logger.Error("failed to acknowledge cost alert", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to acknowledge alert"})
	}

	return c.JSON(fiber.Map{"status": "acknowledged"})
}

// GetCircuitBreakerConfig handles GET /api/public/cost-alerting/circuit-breaker
// @Summary Get circuit breaker config
// @Description Get the circuit breaker configuration for a project
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Success 200 {object} domain.CircuitBreakerConfig
// @Failure 401 {object} map[string]string
// @Router /api/public/cost-alerting/circuit-breaker [get]
func (h *CostAlertingHandler) GetCircuitBreakerConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetCircuitBreaker(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get circuit breaker config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get circuit breaker config"})
	}

	return c.JSON(config)
}

// UpdateCircuitBreakerConfig handles PUT /api/public/cost-alerting/circuit-breaker
// @Summary Update circuit breaker config
// @Description Update the circuit breaker configuration for a project
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Param config body domain.CircuitBreakerConfig true "Circuit breaker configuration"
// @Success 200 {object} domain.CircuitBreakerConfig
// @Failure 400 {object} map[string]string
// @Router /api/public/cost-alerting/circuit-breaker [put]
func (h *CostAlertingHandler) UpdateCircuitBreakerConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var config domain.CircuitBreakerConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.UpdateCircuitBreaker(c.Context(), projectID, config)
	if err != nil {
		h.logger.Error("failed to update circuit breaker config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update circuit breaker config"})
	}

	return c.JSON(result)
}

// CostAlertingCheckRequest represents the request to check cost
type CostAlertingCheckRequest struct {
	Cost  float64 `json:"cost"`
	Model string  `json:"model"`
}

// CheckCost handles POST /api/public/cost-alerting/check
// @Summary Check cost
// @Description Check cost against alert rules and trigger alerts if needed
// @Tags cost-alerting
// @Accept json
// @Produce json
// @Param body body CostAlertingCheckRequest true "Cost check parameters"
// @Success 200 {object} domain.CostAlert
// @Failure 400 {object} map[string]string
// @Router /api/public/cost-alerting/check [post]
func (h *CostAlertingHandler) CheckCost(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req CostAlertingCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	alert, err := h.service.CheckAndAlert(c.Context(), projectID, req.Cost, req.Model)
	if err != nil {
		h.logger.Error("failed to check cost", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check cost"})
	}

	return c.JSON(alert)
}
