package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AIDebuggerHandler handles AI debugger HTTP requests
type AIDebuggerHandler struct {
	logger  *zap.Logger
	service *service.AIDebuggerService
}

// NewAIDebuggerHandler creates a new AI debugger handler
func NewAIDebuggerHandler(svc *service.AIDebuggerService, logger *zap.Logger) *AIDebuggerHandler {
	return &AIDebuggerHandler{
		logger:  logger,
		service: svc,
	}
}

// AIDebugRequest represents the request to debug a trace
type AIDebugRequest struct {
	TraceID   string                `json:"traceId"`
	Query     string                `json:"query"`
	QueryType domain.DebugQueryType `json:"queryType"`
}

// DebugTrace handles POST /api/public/ai-debugger/debug
// @Summary Debug trace
// @Description Submit an AI debugging query for a trace
// @Tags ai-debugger
// @Accept json
// @Produce json
// @Param body body AIDebugRequest true "Debug request"
// @Success 200 {object} domain.DebugQuery
// @Failure 400 {object} map[string]string
// @Router /api/public/ai-debugger/debug [post]
func (h *AIDebuggerHandler) DebugTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req AIDebugRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	traceID, err := uuid.Parse(req.TraceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}

	result, err := h.service.DebugTrace(c.Context(), projectID, traceID, req.Query, req.QueryType)
	if err != nil {
		h.logger.Error("failed to debug trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to debug trace"})
	}

	return c.JSON(result)
}

// GetDebugHistory handles GET /api/public/ai-debugger/traces/:traceId/debug-history
// @Summary Get debug history
// @Description Get the debug query history for a trace
// @Tags ai-debugger
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Success 200 {array} domain.DebugQuery
// @Failure 400 {object} map[string]string
// @Router /api/public/ai-debugger/traces/{traceId}/debug-history [get]
func (h *AIDebuggerHandler) GetDebugHistory(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	history, err := h.service.GetDebugHistory(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get debug history", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get debug history"})
	}

	return c.JSON(history)
}

// BuildContext handles GET /api/public/ai-debugger/traces/:traceId/debug-context
// @Summary Build debug context
// @Description Build the debug context for a trace
// @Tags ai-debugger
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Success 200 {object} domain.DebugContext
// @Failure 400 {object} map[string]string
// @Router /api/public/ai-debugger/traces/{traceId}/debug-context [get]
func (h *AIDebuggerHandler) BuildContext(c *fiber.Ctx) error {
	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	ctx, err := h.service.BuildContext(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to build debug context", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build debug context"})
	}

	return c.JSON(ctx)
}
