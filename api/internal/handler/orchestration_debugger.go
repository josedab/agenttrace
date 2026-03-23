package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OrchestrationDebuggerHandler handles orchestration debugger HTTP requests
type OrchestrationDebuggerHandler struct {
	service *service.OrchestrationDebuggerService
	logger  *zap.Logger
}

// NewOrchestrationDebuggerHandler creates a new orchestration debugger handler
func NewOrchestrationDebuggerHandler(svc *service.OrchestrationDebuggerService, logger *zap.Logger) *OrchestrationDebuggerHandler {
	return &OrchestrationDebuggerHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateSession handles POST /api/public/orchestration/sessions
func (h *OrchestrationDebuggerHandler) CreateSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.OrchestrationSessionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TraceID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "traceId is required"})
	}

	session, err := h.service.CreateSession(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create orchestration session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// GetSession handles GET /api/public/orchestration/sessions/:sessionId
func (h *OrchestrationDebuggerHandler) GetSession(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	session, err := h.service.GetSession(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
	}

	return c.JSON(session)
}

// ExecuteCommand handles POST /api/public/orchestration/sessions/:sessionId/command
func (h *OrchestrationDebuggerHandler) ExecuteCommand(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	var cmd domain.DebugCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if cmd.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "action is required"})
	}

	session, err := h.service.ExecuteCommand(c.Context(), sessionID, &cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(session)
}

// AddBreakpoint handles POST /api/public/orchestration/sessions/:sessionId/breakpoints
func (h *OrchestrationDebuggerHandler) AddBreakpoint(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	var bp domain.AgentBreakpoint
	if err := c.BodyParser(&bp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	session, err := h.service.AddBreakpoint(c.Context(), sessionID, &bp)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// ListSessions handles GET /api/public/orchestration/sessions
func (h *OrchestrationDebuggerHandler) ListSessions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessions, err := h.service.ListSessions(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list orchestration sessions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list sessions"})
	}

	return c.JSON(fiber.Map{"sessions": sessions, "count": len(sessions)})
}
