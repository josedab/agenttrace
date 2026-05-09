package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CostGuardrailHandler handles cost guardrail HTTP requests
type CostGuardrailHandler struct {
	logger               *zap.Logger
	costGuardrailService *service.CostGuardrailService
}

// NewCostGuardrailHandler creates a new cost guardrail handler
func NewCostGuardrailHandler(
	costGuardrailService *service.CostGuardrailService,
	logger *zap.Logger,
) *CostGuardrailHandler {
	return &CostGuardrailHandler{
		logger:               logger,
		costGuardrailService: costGuardrailService,
	}
}

// GetDashboard returns the cost guardrail dashboard for a project
// @Summary Get cost guardrail dashboard
// @Description Get the cost guardrail dashboard with budget overview and alerts
// @Tags cost-guardrail
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.CostGuardrailDashboard
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/cost-guardrail/dashboard [get]
func (h *CostGuardrailHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.costGuardrailService.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get cost guardrail dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}

// CreatePolicy creates a new cost guardrail policy
// @Summary Create cost guardrail policy
// @Description Create a new cost guardrail policy for a project
// @Tags cost-guardrail
// @Accept json
// @Produce json
// @Param policy body domain.CostGuardrailPolicyInput true "Policy configuration"
// @Success 201 {object} domain.CostGuardrailPolicy
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/cost-guardrail/policies [post]
func (h *CostGuardrailHandler) CreatePolicy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CostGuardrailPolicyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	policy, err := h.costGuardrailService.CreatePolicy(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create cost guardrail policy", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create policy"})
	}

	return c.Status(fiber.StatusCreated).JSON(policy)
}

// ListPolicies returns all cost guardrail policies for a project
// @Summary List cost guardrail policies
// @Description Get all cost guardrail policies for a project
// @Tags cost-guardrail
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.CostGuardrailPolicy
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/cost-guardrail/policies [get]
func (h *CostGuardrailHandler) ListPolicies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	policies, err := h.costGuardrailService.ListPolicies(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list cost guardrail policies", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list policies"})
	}

	return c.JSON(policies)
}

// GuardrailCheckBudgetRequest represents the request to check budget for a trace
type GuardrailCheckBudgetRequest struct {
	TraceID       string  `json:"traceId"`
	EstimatedCost float64 `json:"estimatedCost"`
}

// CheckBudget checks if a trace's estimated cost is within budget
// @Summary Check budget for a trace
// @Description Check if a trace's estimated cost is within the project's budget guardrails
// @Tags cost-guardrail
// @Accept json
// @Produce json
// @Param body body CheckBudgetRequest true "Budget check request"
// @Success 200 {object} domain.BudgetCheckResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/cost-guardrail/check-budget [post]
func (h *CostGuardrailHandler) CheckBudget(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req GuardrailCheckBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	traceID, err := uuid.Parse(req.TraceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	result, err := h.costGuardrailService.CheckBudget(c.Context(), projectID, traceID, req.EstimatedCost)
	if err != nil {
		h.logger.Error("failed to check budget", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check budget"})
	}

	return c.JSON(result)
}

// GetForecast returns cost forecast for a project
// @Summary Get cost forecast
// @Description Get cost forecast based on current usage patterns
// @Tags cost-guardrail
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.CostForecast
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/cost-guardrail/forecast [get]
func (h *CostGuardrailHandler) GetForecast(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	forecast, err := h.costGuardrailService.GetForecast(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get cost forecast", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get forecast"})
	}

	return c.JSON(forecast)
}

// ListViolations returns cost guardrail violations for a project
// @Summary List cost guardrail violations
// @Description Get all cost guardrail violations for a project
// @Tags cost-guardrail
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.CostGuardrailViolation
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/cost-guardrail/violations [get]
func (h *CostGuardrailHandler) ListViolations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	violations, err := h.costGuardrailService.ListViolations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list cost guardrail violations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list violations"})
	}

	return c.JSON(violations)
}
