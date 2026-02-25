package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// MultiAgentGraphHandler handles multi-agent collaboration graph HTTP requests
type MultiAgentGraphHandler struct {
	logger                *zap.Logger
	multiAgentGraphService *service.MultiAgentGraphService
}

// NewMultiAgentGraphHandler creates a new multi-agent graph handler
func NewMultiAgentGraphHandler(
	multiAgentGraphService *service.MultiAgentGraphService,
	logger *zap.Logger,
) *MultiAgentGraphHandler {
	return &MultiAgentGraphHandler{
		logger:                logger,
		multiAgentGraphService: multiAgentGraphService,
	}
}

// AnalyzeSessionRequest represents the request to analyze a multi-agent session
type AnalyzeSessionRequest struct {
	TraceID string `json:"traceId"`
}

// AnalyzeSession analyzes a trace to build a multi-agent collaboration graph
// @Summary Analyze multi-agent session
// @Description Analyze a trace to extract multi-agent collaboration patterns and build a graph
// @Tags multi-agent-graph
// @Accept json
// @Produce json
// @Param body body AnalyzeSessionRequest true "Session analysis request"
// @Success 201 {object} domain.MultiAgentSession
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/multi-agent-graph/sessions/analyze [post]
func (h *MultiAgentGraphHandler) AnalyzeSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req AnalyzeSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	traceID, err := uuid.Parse(req.TraceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	session, err := h.multiAgentGraphService.AnalyzeSession(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to analyze multi-agent session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to analyze session"})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// ListSessions returns all multi-agent sessions for a project
// @Summary List multi-agent sessions
// @Description Get all multi-agent collaboration sessions for a project
// @Tags multi-agent-graph
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.MultiAgentSession
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/multi-agent-graph/sessions [get]
func (h *MultiAgentGraphHandler) ListSessions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessions, err := h.multiAgentGraphService.ListSessions(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list multi-agent sessions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list sessions"})
	}

	return c.JSON(sessions)
}

// GetSession returns a specific multi-agent session
// @Summary Get multi-agent session
// @Description Get a specific multi-agent collaboration session by ID
// @Tags multi-agent-graph
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} domain.MultiAgentSession
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/multi-agent-graph/sessions/{sessionId} [get]
func (h *MultiAgentGraphHandler) GetSession(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	session, err := h.multiAgentGraphService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get multi-agent session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	return c.JSON(session)
}

// GetTopologyGraph handles GET /api/public/multi-agent-graph/sessions/:sessionId/topology
func (h *MultiAgentGraphHandler) GetTopologyGraph(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	graph, err := h.multiAgentGraphService.GetTopologyGraph(c.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get topology graph", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get topology graph"})
	}

	return c.JSON(graph)
}