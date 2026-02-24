package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CopilotHandler handles observability copilot HTTP requests
type CopilotHandler struct {
	service *service.CopilotService
	logger  *zap.Logger
}

// NewCopilotHandler creates a new copilot handler
func NewCopilotHandler(svc *service.CopilotService, logger *zap.Logger) *CopilotHandler {
	return &CopilotHandler{
		service: svc,
		logger:  logger,
	}
}

// Ask handles POST /api/public/copilot/ask
func (h *CopilotHandler) Ask(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CopilotQueryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query, err := h.service.AskQuestion(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to process copilot query", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process query"})
	}

	return c.JSON(query)
}

// GetSuggestions handles GET /api/public/copilot/suggestions
func (h *CopilotHandler) GetSuggestions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	suggestions, err := h.service.GetSuggestions(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get copilot suggestions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get suggestions"})
	}

	return c.JSON(suggestions)
}

// GetInsights handles GET /api/public/copilot/insights
func (h *CopilotHandler) GetInsights(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	insights, err := h.service.GetProactiveInsights(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get copilot insights", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get insights"})
	}

	return c.JSON(insights)
}
