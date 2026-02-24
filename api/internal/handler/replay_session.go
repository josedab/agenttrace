package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ReplaySessionHandler handles replay session HTTP requests
type ReplaySessionHandler struct {
	logger  *zap.Logger
	service *service.ReplaySessionService
}

// NewReplaySessionHandler creates a new replay session handler
func NewReplaySessionHandler(
	service *service.ReplaySessionService,
	logger *zap.Logger,
) *ReplaySessionHandler {
	return &ReplaySessionHandler{logger: logger, service: service}
}

// ListSessions handles GET /api/public/replay-sessions
func (h *ReplaySessionHandler) ListSessions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	filter := domain.AgentReplaySessionFilter{ProjectID: projectID}

	if traceIDStr := c.Query("traceId"); traceIDStr != "" {
		traceID, err := uuid.Parse(traceIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
		}
		filter.TraceID = &traceID
	}

	result, err := h.service.ListSessions(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list sessions"})
	}
	return c.JSON(result)
}

// CreateSession handles POST /api/public/replay-sessions
func (h *ReplaySessionHandler) CreateSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AgentReplaySessionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	userID := uuid.New()
	session, err := h.service.CreateSession(c.Context(), projectID, userID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
	}
	return c.Status(fiber.StatusCreated).JSON(session)
}

// GetSession handles GET /api/public/replay-sessions/:sessionId
func (h *ReplaySessionHandler) GetSession(c *fiber.Ctx) error {
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

// GetTimeline handles GET /api/public/replay-sessions/:sessionId/timeline
func (h *ReplaySessionHandler) GetTimeline(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	timeline, err := h.service.GetTimeline(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get timeline"})
	}
	return c.JSON(timeline)
}

// BranchSession handles POST /api/public/replay-sessions/:sessionId/branch
func (h *ReplaySessionHandler) BranchSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	var req domain.AgentReplayBranchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.SessionID = sessionID

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Branch name is required"})
	}

	userID := uuid.New()
	branch, err := h.service.BranchSession(c.Context(), projectID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to branch session"})
	}
	return c.Status(fiber.StatusCreated).JSON(branch)
}

// GetPlaybackState handles GET /api/public/replay-sessions/:sessionId/playback
func (h *ReplaySessionHandler) GetPlaybackState(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	state, err := h.service.GetPlaybackState(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get playback state"})
	}
	return c.JSON(state)
}

// ShareSession handles POST /api/public/replay-sessions/:sessionId/share
func (h *ReplaySessionHandler) ShareSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	shareURL, err := h.service.ShareSession(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to share session"})
	}

	return c.JSON(fiber.Map{"shareUrl": shareURL, "isPublic": true})
}
