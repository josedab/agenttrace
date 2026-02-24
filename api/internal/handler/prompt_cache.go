package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PromptCacheHandler handles prompt caching HTTP requests
type PromptCacheHandler struct {
	service *service.PromptCacheService
	logger  *zap.Logger
}

// NewPromptCacheHandler creates a new prompt cache handler
func NewPromptCacheHandler(svc *service.PromptCacheService, logger *zap.Logger) *PromptCacheHandler {
	return &PromptCacheHandler{
		service: svc,
		logger:  logger,
	}
}

// Analyze handles GET /api/public/prompt-cache/analyze
func (h *PromptCacheHandler) Analyze(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	analysis, err := h.service.AnalyzeCache(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to analyze cache", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to analyze cache"})
	}

	return c.JSON(analysis)
}

// GetConfig handles GET /api/public/prompt-cache/config
func (h *PromptCacheHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get cache config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get cache config"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /api/public/prompt-cache/config
func (h *PromptCacheHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CacheConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to update cache config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cache config"})
	}

	return c.JSON(config)
}

// GetStats handles GET /api/public/prompt-cache/stats
func (h *PromptCacheHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get cache stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get cache stats"})
	}

	return c.JSON(stats)
}

// Invalidate handles POST /api/public/prompt-cache/invalidate
func (h *PromptCacheHandler) Invalidate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.InvalidateCache(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to invalidate cache", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to invalidate cache"})
	}

	return c.JSON(result)
}
