package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ComplianceMonitorHandler handles compliance monitoring HTTP requests
type ComplianceMonitorHandler struct {
	service *service.ComplianceMonitorService
	logger  *zap.Logger
}

// NewComplianceMonitorHandler creates a new compliance monitor handler
func NewComplianceMonitorHandler(svc *service.ComplianceMonitorService, logger *zap.Logger) *ComplianceMonitorHandler {
	return &ComplianceMonitorHandler{
		service: svc,
		logger:  logger,
	}
}

// CreatePolicy handles POST /api/public/compliance-monitor/policies
func (h *ComplianceMonitorHandler) CreatePolicy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CompliancePolicyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	policy, err := h.service.CreatePolicy(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to create compliance policy", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create policy"})
	}

	return c.Status(fiber.StatusCreated).JSON(policy)
}

// Evaluate handles POST /api/public/compliance-monitor/evaluate
func (h *ComplianceMonitorHandler) Evaluate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	framework := c.Query("framework", "eu_ai_act")

	score, err := h.service.EvaluateCompliance(c.Context(), projectID.String(), framework)
	if err != nil {
		h.logger.Error("failed to evaluate compliance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to evaluate compliance"})
	}

	return c.JSON(score)
}

// GetScore handles GET /api/public/compliance-monitor/score/:framework
func (h *ComplianceMonitorHandler) GetScore(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	framework := c.Params("framework")

	score, err := h.service.GetScore(c.Context(), projectID.String(), framework)
	if err != nil {
		h.logger.Error("failed to get compliance score", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get score"})
	}

	return c.JSON(score)
}

// ListPolicies handles GET /api/public/compliance-monitor/policies
func (h *ComplianceMonitorHandler) ListPolicies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	policies, err := h.service.ListPolicies(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to list compliance policies", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list policies"})
	}

	return c.JSON(policies)
}

// Configure handles POST /api/public/compliance-monitor/configure
func (h *ComplianceMonitorHandler) Configure(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var config domain.ContinuousMonitorConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.ConfigureMonitor(c.Context(), projectID.String(), &config)
	if err != nil {
		h.logger.Error("failed to configure compliance monitor", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to configure monitor"})
	}

	return c.JSON(result)
}
