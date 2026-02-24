package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// WebhookOrchestrationHandler handles webhook orchestration HTTP requests
type WebhookOrchestrationHandler struct {
	service *service.WebhookOrchestrationService
	logger  *zap.Logger
}

// NewWebhookOrchestrationHandler creates a new webhook orchestration handler
func NewWebhookOrchestrationHandler(svc *service.WebhookOrchestrationService, logger *zap.Logger) *WebhookOrchestrationHandler {
	return &WebhookOrchestrationHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateRule handles POST /api/public/webhook-rules
func (h *WebhookOrchestrationHandler) CreateRule(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.WebhookRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.Trigger == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trigger is required"})
	}
	if input.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Action is required"})
	}

	rule, err := h.service.CreateRule(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create webhook rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create rule"})
	}

	return c.Status(fiber.StatusCreated).JSON(rule)
}

// ListRules handles GET /api/public/webhook-rules
func (h *WebhookOrchestrationHandler) ListRules(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	rules := h.service.ListRules(c.Context(), projectID)
	if rules == nil {
		rules = []domain.WebhookRule{}
	}

	return c.JSON(fiber.Map{"rules": rules, "count": len(rules)})
}

// DeleteRule handles DELETE /api/public/webhook-rules/:ruleId
func (h *WebhookOrchestrationHandler) DeleteRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	if err := h.service.DeleteRule(c.Context(), ruleID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rule not found"})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// GetTemplates handles GET /api/public/webhook-rules/templates
func (h *WebhookOrchestrationHandler) GetTemplates(c *fiber.Ctx) error {
	templates := h.service.GetTemplates(c.Context())
	return c.JSON(fiber.Map{"templates": templates, "count": len(templates)})
}

// ListDeliveries handles GET /api/public/webhook-rules/deliveries
func (h *WebhookOrchestrationHandler) ListDeliveries(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	deliveries := h.service.ListDeliveries(c.Context(), projectID)
	if deliveries == nil {
		deliveries = []domain.WebhookRuleDelivery{}
	}

	return c.JSON(fiber.Map{"deliveries": deliveries, "count": len(deliveries)})
}

// TestRule handles POST /api/public/webhook-rules/:ruleId/test
func (h *WebhookOrchestrationHandler) TestRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	delivery, err := h.service.TestRule(c.Context(), ruleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rule not found"})
	}

	return c.JSON(delivery)
}
