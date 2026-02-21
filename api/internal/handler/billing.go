package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// BillingHandler handles billing HTTP requests
type BillingHandler struct {
	billingService *service.BillingService
	logger         *zap.Logger
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(billingService *service.BillingService, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		logger:         logger,
	}
}

// GetPlans handles GET /billing/plans
func (h *BillingHandler) GetPlans(c *fiber.Ctx) error {
	plans := h.billingService.GetPlans()
	return c.JSON(plans)
}

// GetSubscription handles GET /billing/subscription
func (h *BillingHandler) GetSubscription(c *fiber.Ctx) error {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Tenant ID not found",
		})
	}

	sub, err := h.billingService.GetSubscription(c.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get subscription",
			zap.String("tenantId", tenantID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get subscription",
		})
	}

	return c.JSON(sub)
}

// UpgradePlan handles POST /billing/upgrade
func (h *BillingHandler) UpgradePlan(c *fiber.Ctx) error {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Tenant ID not found",
		})
	}

	var input domain.PlanUpgradeInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	sub, err := h.billingService.UpgradePlan(c.Context(), tenantID, input)
	if err != nil {
		h.logger.Error("failed to upgrade plan",
			zap.String("tenantId", tenantID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to upgrade plan",
		})
	}

	return c.JSON(sub)
}

// CancelSubscription handles POST /billing/cancel
func (h *BillingHandler) CancelSubscription(c *fiber.Ctx) error {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Tenant ID not found",
		})
	}

	if err := h.billingService.CancelSubscription(c.Context(), tenantID); err != nil {
		h.logger.Error("failed to cancel subscription",
			zap.String("tenantId", tenantID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to cancel subscription",
		})
	}

	return c.JSON(fiber.Map{"status": "cancelled"})
}

// GetInvoices handles GET /billing/invoices
func (h *BillingHandler) GetInvoices(c *fiber.Ctx) error {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Tenant ID not found",
		})
	}

	limit := parseIntParam(c, "limit", 50)

	invoices, err := h.billingService.GetInvoices(c.Context(), tenantID, limit)
	if err != nil {
		h.logger.Error("failed to get invoices",
			zap.String("tenantId", tenantID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get invoices",
		})
	}

	return c.JSON(invoices)
}

// HandleWebhook handles POST /billing/webhook (no auth — Stripe sends this)
func (h *BillingHandler) HandleWebhook(c *fiber.Ctx) error {
	var payload map[string]any
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid webhook payload",
		})
	}

	eventType, _ := payload["type"].(string)
	if eventType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Missing event type",
		})
	}

	if err := h.billingService.HandleWebhook(c.Context(), eventType, payload); err != nil {
		h.logger.Error("failed to handle billing webhook",
			zap.String("eventType", eventType),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to process webhook",
		})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
