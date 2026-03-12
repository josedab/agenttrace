package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TraceEnrichmentHandler handles webhook trace enrichment pipeline HTTP requests
type TraceEnrichmentHandler struct {
	service *service.TraceEnrichmentService
	logger  *zap.Logger
}

// NewTraceEnrichmentHandler creates a new trace enrichment handler
func NewTraceEnrichmentHandler(svc *service.TraceEnrichmentService, logger *zap.Logger) *TraceEnrichmentHandler {
	return &TraceEnrichmentHandler{
		service: svc,
		logger:  logger,
	}
}

// ListRules handles GET /enrichment/rules
func (h *TraceEnrichmentHandler) ListRules(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.ListRules(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list enrichment rules", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list enrichment rules"})
	}

	return c.JSON(result)
}

// CreateRule handles POST /enrichment/rules
func (h *TraceEnrichmentHandler) CreateRule(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EnrichmentRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	result, err := h.service.CreateRule(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create enrichment rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create enrichment rule"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateRule handles PUT /enrichment/rules/:ruleId
func (h *TraceEnrichmentHandler) UpdateRule(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	ruleIDStr := c.Params("ruleId")
	if ruleIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rule ID is required"})
	}
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	var input domain.EnrichmentRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.UpdateRule(c.Context(), ruleID, &input)
	if err != nil {
		h.logger.Error("failed to update enrichment rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update enrichment rule"})
	}

	return c.JSON(result)
}

// DeleteRule handles DELETE /enrichment/rules/:ruleId
func (h *TraceEnrichmentHandler) DeleteRule(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	ruleIDStr := c.Params("ruleId")
	if ruleIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rule ID is required"})
	}
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	err = h.service.DeleteRule(c.Context(), ruleID)
	if err != nil {
		h.logger.Error("failed to delete enrichment rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete enrichment rule"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListSources handles GET /enrichment/sources
func (h *TraceEnrichmentHandler) ListSources(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	_ = projectID
	result, err := h.service.ListSources(c.Context())
	if err != nil {
		h.logger.Error("failed to list enrichment sources", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list enrichment sources"})
	}

	return c.JSON(result)
}

// TestRule handles POST /enrichment/test
func (h *TraceEnrichmentHandler) TestRule(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EnrichmentTestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	_ = projectID
	result, err := h.service.TestRule(c.Context(), &input)
	if err != nil {
		h.logger.Error("failed to test enrichment rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to test enrichment rule"})
	}

	return c.JSON(result)
}
