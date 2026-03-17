package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// GuardrailsHandler handles guardrail rule and violation HTTP requests
type GuardrailsHandler struct {
	guardrailService *service.GuardrailService
	logger           *zap.Logger
}

// NewGuardrailsHandler creates a new guardrails handler
func NewGuardrailsHandler(guardrailService *service.GuardrailService, logger *zap.Logger) *GuardrailsHandler {
	return &GuardrailsHandler{
		guardrailService: guardrailService,
		logger:           logger,
	}
}

// CreateRule handles POST /guardrails
func (h *GuardrailsHandler) CreateRule(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.GuardRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	rule, err := h.guardrailService.CreateRule(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create guard rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create guard rule",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(rule)
}

// UpdateRule handles PUT /guardrails/:ruleId
func (h *GuardrailsHandler) UpdateRule(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid rule ID",
		})
	}

	var input domain.GuardRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	rule, err := h.guardrailService.UpdateRule(c.Context(), ruleID, &input)
	if err != nil {
		h.logger.Error("failed to update guard rule",
			zap.String("ruleId", ruleID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update guard rule",
		})
	}

	return c.JSON(rule)
}

// DeleteRule handles DELETE /guardrails/:ruleId
func (h *GuardrailsHandler) DeleteRule(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid rule ID",
		})
	}

	if err := h.guardrailService.DeleteRule(c.Context(), ruleID); err != nil {
		h.logger.Error("failed to delete guard rule",
			zap.String("ruleId", ruleID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete guard rule",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListRules handles GET /guardrails
func (h *GuardrailsHandler) ListRules(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	rules, err := h.guardrailService.ListRules(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list guard rules", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list guard rules",
		})
	}

	return c.JSON(rules)
}

// ListViolations handles GET /guardrails/violations
func (h *GuardrailsHandler) ListViolations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := &domain.GuardViolationFilter{}

	violations, err := h.guardrailService.ListViolations(c.Context(), projectID, filter)
	if err != nil {
		h.logger.Error("failed to list violations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list violations",
		})
	}

	return c.JSON(violations)
}

// GetPlaybookTemplates handles GET /api/public/guardrails/templates
func (h *GuardrailsHandler) GetPlaybookTemplates(c *fiber.Ctx) error {
	templates := h.guardrailService.GetPlaybookTemplates()
	return c.JSON(fiber.Map{"templates": templates})
}

// CreatePlaybook handles POST /api/public/guardrails/playbooks
func (h *GuardrailsHandler) CreatePlaybook(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.GuardPlaybookInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Playbook name is required"})
	}

	playbook, err := h.guardrailService.CreatePlaybook(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create playbook", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create playbook"})
	}

	return c.Status(fiber.StatusCreated).JSON(playbook)
}

// GetViolationStats handles GET /guardrails/violations/stats
func (h *GuardrailsHandler) GetViolationStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	stats, err := h.guardrailService.GetViolationStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get violation stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get violation stats",
		})
	}

	return c.JSON(stats)
}

// CreateSelfHealingPolicy handles POST /api/public/guardrails/policies
func (h *GuardrailsHandler) CreateSelfHealingPolicy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.SelfHealingPolicyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	policy, err := h.guardrailService.CreateSelfHealingPolicy(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create self-healing policy", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(policy)
}

// ListSelfHealingPolicies handles GET /api/public/guardrails/policies
func (h *GuardrailsHandler) ListSelfHealingPolicies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	policies, err := h.guardrailService.ListSelfHealingPolicies(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list policies"})
	}

	return c.JSON(fiber.Map{"policies": policies})
}

// EvaluatePipeline handles POST /api/public/guardrails/evaluate
func (h *GuardrailsHandler) EvaluatePipeline(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EvalPipelineInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.guardrailService.EvaluatePipeline(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to evaluate pipeline", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to evaluate"})
	}

	return c.JSON(result)
}

// GetDashboardStats handles GET /api/public/guardrails/dashboard
func (h *GuardrailsHandler) GetDashboardStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.guardrailService.GetGuardrailDashboardStats(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(stats)
}

// GetAuditTrail handles GET /api/public/guardrails/policies/:policyId/audit
func (h *GuardrailsHandler) GetAuditTrail(c *fiber.Ctx) error {
	policyID, err := uuid.Parse(c.Params("policyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid policy ID"})
	}

	trail, err := h.guardrailService.GetPolicyAuditTrail(c.Context(), policyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get audit trail"})
	}

	return c.JSON(fiber.Map{"auditTrail": trail})
}
