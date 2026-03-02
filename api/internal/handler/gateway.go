package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// GatewayHandler handles LLM gateway HTTP requests
type GatewayHandler struct {
	service *service.GatewayService
	logger  *zap.Logger
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(svc *service.GatewayService, logger *zap.Logger) *GatewayHandler {
	return &GatewayHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateConfig handles POST /api/public/gateway/configs
func (h *GatewayHandler) CreateConfig(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.GatewayConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.Strategy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Strategy is required"})
	}
	if len(input.Providers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one provider is required"})
	}

	config, err := h.service.CreateConfig(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create gateway config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create gateway config"})
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

// GetConfig handles GET /api/public/gateway/configs/:configId
func (h *GatewayHandler) GetConfig(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	config, err := h.service.GetConfig(c.Context(), configID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Config not found"})
	}

	return c.JSON(config)
}

// ListConfigs handles GET /api/public/gateway/configs
func (h *GatewayHandler) ListConfigs(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	configs, err := h.service.ListConfigs(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list gateway configs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list configs"})
	}

	if configs == nil {
		configs = []domain.GatewayConfig{}
	}

	return c.JSON(fiber.Map{"configs": configs, "count": len(configs)})
}

// UpdateConfig handles PUT /api/public/gateway/configs/:configId
func (h *GatewayHandler) UpdateConfig(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	var input domain.GatewayConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), configID, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(config)
}

// DeleteConfig handles DELETE /api/public/gateway/configs/:configId
func (h *GatewayHandler) DeleteConfig(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	if err := h.service.DeleteConfig(c.Context(), configID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// ProxyRequest handles POST /api/public/gateway/proxy/:configId
func (h *GatewayHandler) ProxyRequest(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	var req domain.GatewayRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Messages) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Messages are required"})
	}

	response, err := h.service.ProxyRequest(c.Context(), configID, &req)
	if err != nil {
		h.logger.Error("gateway proxy failed", zap.Error(err))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(response)
}

// AddRoutingRule handles POST /api/public/gateway/configs/:configId/rules
func (h *GatewayHandler) AddRoutingRule(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	var rule domain.RoutingRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if rule.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rule name is required"})
	}

	created, err := h.service.AddRoutingRule(c.Context(), configID, &rule)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

// ListRoutingRules handles GET /api/public/gateway/configs/:configId/rules
func (h *GatewayHandler) ListRoutingRules(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	rules, err := h.service.ListRoutingRules(c.Context(), configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"rules": rules, "count": len(rules)})
}

// DeleteRoutingRule handles DELETE /api/public/gateway/configs/:configId/rules/:ruleId
func (h *GatewayHandler) DeleteRoutingRule(c *fiber.Ctx) error {
	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	ruleID, err := uuid.Parse(c.Params("ruleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	if err := h.service.DeleteRoutingRule(c.Context(), configID, ruleID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// GetStats handles GET /api/public/gateway/stats
func (h *GatewayHandler) GetStats(c *fiber.Ctx) error {
	stats := h.service.GetStats(c.Context())
	return c.JSON(stats)
}
