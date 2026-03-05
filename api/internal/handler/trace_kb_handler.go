package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TraceKBHandler handles trace annotation knowledge base HTTP requests
type TraceKBHandler struct {
	service *service.TraceKBService
	logger  *zap.Logger
}

// NewTraceKBHandler creates a new trace knowledge base handler
func NewTraceKBHandler(svc *service.TraceKBService, logger *zap.Logger) *TraceKBHandler {
	return &TraceKBHandler{
		service: svc,
		logger:  logger,
	}
}

// ListEntries handles GET /knowledge-base/entries
func (h *TraceKBHandler) ListEntries(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.ListEntries(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to list knowledge base entries", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list knowledge base entries"})
	}

	return c.JSON(result)
}

// CreateEntry handles POST /knowledge-base/entries
func (h *TraceKBHandler) CreateEntry(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.KBEntryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}

	result, err := h.service.CreateEntry(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to create knowledge base entry", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create knowledge base entry"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetEntry handles GET /knowledge-base/entries/:entryId
func (h *TraceKBHandler) GetEntry(c *fiber.Ctx) error {
	entryID := c.Params("entryId")
	if entryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Entry ID is required"})
	}

	result, err := h.service.GetEntry(c.Context(), entryID)
	if err != nil {
		h.logger.Error("failed to get knowledge base entry", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Knowledge base entry not found"})
	}

	return c.JSON(result)
}

// Search handles GET /knowledge-base/search
func (h *TraceKBHandler) Search(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query parameter is required"})
	}

	result, err := h.service.Search(c.Context(), projectID.String(), query)
	if err != nil {
		h.logger.Error("failed to search knowledge base", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to search knowledge base"})
	}

	return c.JSON(result)
}

// GetSuggestions handles GET /knowledge-base/suggestions
func (h *TraceKBHandler) GetSuggestions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID := c.Query("traceId")

	result, err := h.service.GetSuggestions(c.Context(), projectID.String(), traceID)
	if err != nil {
		h.logger.Error("failed to get suggestions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get suggestions"})
	}

	return c.JSON(result)
}
