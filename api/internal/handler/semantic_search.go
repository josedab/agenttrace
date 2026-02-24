package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// SemanticSearchHandler handles semantic search HTTP requests
type SemanticSearchHandler struct {
	logger  *zap.Logger
	service *service.SemanticSearchService
}

// NewSemanticSearchHandler creates a new semantic search handler
func NewSemanticSearchHandler(logger *zap.Logger, svc *service.SemanticSearchService) *SemanticSearchHandler {
	return &SemanticSearchHandler{
		logger:  logger,
		service: svc,
	}
}

// Search handles POST /api/public/search
func (h *SemanticSearchHandler) Search(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var query domain.SemanticSearchQuery
	if err := c.BodyParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if query.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Search query is required"})
	}

	result, err := h.service.Search(c.Context(), projectID, query)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Search failed"})
	}

	return c.JSON(result)
}

// GetSuggestions handles GET /api/public/search/suggestions
func (h *SemanticSearchHandler) GetSuggestions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	prefix := c.Query("prefix", "")

	suggestions, err := h.service.GetSuggestions(c.Context(), projectID, prefix)
	if err != nil {
		h.logger.Error("failed to get suggestions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get suggestions"})
	}

	return c.JSON(fiber.Map{"suggestions": suggestions})
}
