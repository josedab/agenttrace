package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PromptOptimizationHandler handles prompt optimization HTTP requests
type PromptOptimizationHandler struct {
	logger  *zap.Logger
	service *service.PromptOptimizationService
}

// NewPromptOptimizationHandler creates a new prompt optimization handler
func NewPromptOptimizationHandler(svc *service.PromptOptimizationService, logger *zap.Logger) *PromptOptimizationHandler {
	return &PromptOptimizationHandler{
		logger:  logger,
		service: svc,
	}
}

// PromptOptimizationStartRequest represents the request to start an optimization
type PromptOptimizationStartRequest struct {
	PromptID      string `json:"promptId"`
	PromptVersion int    `json:"promptVersion"`
}

// StartOptimization handles POST /api/public/prompt-optimization/optimizations
// @Summary Start optimization
// @Description Start a new prompt optimization
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Param body body PromptOptimizationStartRequest true "Optimization parameters"
// @Success 201 {object} domain.ContinuousPromptOptimization
// @Failure 400 {object} map[string]string
// @Router /api/public/prompt-optimization/optimizations [post]
func (h *PromptOptimizationHandler) StartOptimization(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req PromptOptimizationStartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	promptID, err := uuid.Parse(req.PromptID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid prompt ID"})
	}

	optimization, err := h.service.StartOptimization(c.Context(), projectID, promptID, req.PromptVersion)
	if err != nil {
		h.logger.Error("failed to start optimization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start optimization"})
	}

	return c.Status(fiber.StatusCreated).JSON(optimization)
}

// GetOptimization handles GET /api/public/prompt-optimization/optimizations/:optimizationId
// @Summary Get optimization
// @Description Get a specific optimization by ID
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Param optimizationId path string true "Optimization ID"
// @Success 200 {object} domain.ContinuousPromptOptimization
// @Failure 400 {object} map[string]string
// @Router /api/public/prompt-optimization/optimizations/{optimizationId} [get]
func (h *PromptOptimizationHandler) GetOptimization(c *fiber.Ctx) error {
	optimizationID, err := uuid.Parse(c.Params("optimizationId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid optimization ID"})
	}

	optimization, err := h.service.GetOptimization(c.Context(), optimizationID)
	if err != nil {
		h.logger.Error("failed to get optimization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get optimization"})
	}

	return c.JSON(optimization)
}

// ListOptimizations handles GET /api/public/prompt-optimization/optimizations
// @Summary List optimizations
// @Description List all optimizations for a project
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Success 200 {array} domain.ContinuousPromptOptimization
// @Failure 401 {object} map[string]string
// @Router /api/public/prompt-optimization/optimizations [get]
func (h *PromptOptimizationHandler) ListOptimizations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	optimizations, err := h.service.ListOptimizations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list optimizations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list optimizations"})
	}

	return c.JSON(optimizations)
}

// GetOptConfig handles GET /api/public/prompt-optimization/config
// @Summary Get optimization config
// @Description Get the optimization configuration for a project
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Success 200 {object} domain.OptimizationConfig
// @Failure 401 {object} map[string]string
// @Router /api/public/prompt-optimization/config [get]
func (h *PromptOptimizationHandler) GetOptConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get optimization config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get config"})
	}

	return c.JSON(config)
}

// UpdateOptConfig handles PUT /api/public/prompt-optimization/config
// @Summary Update optimization config
// @Description Update the optimization configuration for a project
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Param config body domain.OptimizationConfig true "Optimization configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/prompt-optimization/config [put]
func (h *PromptOptimizationHandler) UpdateOptConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var config domain.OptimizationConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.UpdateConfig(c.Context(), projectID, config); err != nil {
		h.logger.Error("failed to update optimization config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update config"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// ApproveVariant handles POST /api/public/prompt-optimization/variants/:variantId/approve
// @Summary Approve variant
// @Description Approve an optimization variant
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Param variantId path string true "Variant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/prompt-optimization/variants/{variantId}/approve [post]
func (h *PromptOptimizationHandler) ApproveVariant(c *fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("variantId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid variant ID"})
	}

	if err := h.service.ApproveVariant(c.Context(), variantID); err != nil {
		h.logger.Error("failed to approve variant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to approve variant"})
	}

	return c.JSON(fiber.Map{"status": "approved"})
}

// RejectVariant handles POST /api/public/prompt-optimization/variants/:variantId/reject
// @Summary Reject variant
// @Description Reject an optimization variant
// @Tags prompt-optimization
// @Accept json
// @Produce json
// @Param variantId path string true "Variant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/prompt-optimization/variants/{variantId}/reject [post]
func (h *PromptOptimizationHandler) RejectVariant(c *fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("variantId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid variant ID"})
	}

	if err := h.service.RejectVariant(c.Context(), variantID); err != nil {
		h.logger.Error("failed to reject variant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reject variant"})
	}

	return c.JSON(fiber.Map{"status": "rejected"})
}
