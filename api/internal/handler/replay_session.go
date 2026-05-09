package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
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

	userID, ok := replayActorID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Actor ID not found"})
	}
	session, err := h.service.CreateSession(c.Context(), projectID, userID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
	}
	return c.Status(fiber.StatusCreated).JSON(session)
}

// GetSession handles GET /api/public/replay-sessions/:sessionId
func (h *ReplaySessionHandler) GetSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	session, err := h.service.GetSession(c.Context(), projectID, sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
	}
	return c.JSON(session)
}

// GetTimeline handles GET /api/public/replay-sessions/:sessionId/timeline
func (h *ReplaySessionHandler) GetTimeline(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	timeline, err := h.service.GetTimeline(c.Context(), projectID, sessionID)
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

	userID, ok := replayActorID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Actor ID not found"})
	}
	branch, err := h.service.BranchSession(c.Context(), projectID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to branch session"})
	}
	return c.Status(fiber.StatusCreated).JSON(branch)
}

// GetPlaybackState handles GET /api/public/replay-sessions/:sessionId/playback
func (h *ReplaySessionHandler) GetPlaybackState(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	state, err := h.service.GetPlaybackState(c.Context(), projectID, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get playback state"})
	}
	return c.JSON(state)
}

// ShareSession handles POST /api/public/replay-sessions/:sessionId/share
func (h *ReplaySessionHandler) ShareSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	shareURL, err := h.service.ShareSession(c.Context(), projectID, sessionID)
	if err != nil {
		return replaySessionServiceError(c, err, "Failed to share session")
	}

	return c.JSON(fiber.Map{"shareUrl": shareURL, "isPublic": true})
}

// RecordEvents handles POST /api/public/replay-sessions/:sessionId/events
func (h *ReplaySessionHandler) RecordEvents(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	var inputs []domain.AgentReplayRecordEventInput
	if err := c.BodyParser(&inputs); err != nil {
		// Try single event
		var single domain.AgentReplayRecordEventInput
		if err := c.BodyParser(&single); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}
		inputs = []domain.AgentReplayRecordEventInput{single}
	}

	if len(inputs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one event is required"})
	}

	events, err := h.service.RecordEvents(c.Context(), projectID, sessionID, inputs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record events"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"events": events, "count": len(events)})
}

// ControlPlayback handles POST /api/public/replay-sessions/:sessionId/control
func (h *ReplaySessionHandler) ControlPlayback(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	var cmd domain.ReplayControlCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if cmd.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Action is required"})
	}

	state, err := h.service.ControlPlayback(c.Context(), projectID, sessionID, &cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}
	return c.JSON(state)
}

// GetFileState handles GET /api/public/replay-sessions/:sessionId/files?eventIndex=N
func (h *ReplaySessionHandler) GetFileState(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	eventIndex := c.QueryInt("eventIndex", 0)
	if eventIndex < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event index"})
	}

	snapshot, err := h.service.GetFileStateAt(c.Context(), projectID, sessionID, eventIndex)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get file state"})
	}
	return c.JSON(snapshot)
}

// CompleteSession handles POST /api/public/replay-sessions/:sessionId/complete
func (h *ReplaySessionHandler) CompleteSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	if err := h.service.CompleteSession(c.Context(), projectID, sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to complete session"})
	}
	return c.JSON(fiber.Map{"status": "completed"})
}

// GetUnifiedTimeline handles GET /api/public/replay-sessions/:sessionId/unified-timeline
func (h *ReplaySessionHandler) GetUnifiedTimeline(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	events, err := h.service.BuildUnifiedTimeline(c.Context(), projectID, sessionID)
	if err != nil {
		h.logger.Error("failed to build unified timeline", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build timeline"})
	}

	return c.JSON(fiber.Map{"events": events, "totalEvents": len(events)})
}

// GetSnapshot handles GET /api/public/replay-sessions/:sessionId/snapshot
func (h *ReplaySessionHandler) GetSnapshot(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	eventIndex := c.QueryInt("eventIndex", 0)

	snapshot, err := h.service.GetReplaySnapshot(c.Context(), projectID, sessionID, eventIndex)
	if err != nil {
		h.logger.Error("failed to get snapshot", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get snapshot"})
	}

	return c.JSON(snapshot)
}

// AddReplayAnnotation handles POST /api/public/replay-sessions/:sessionId/annotations
func (h *ReplaySessionHandler) AddReplayAnnotation(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID"})
	}

	var input domain.ReplayAnnotationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, ok := replayActorID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Actor ID not found"})
	}
	annotation, err := h.service.AddAnnotation(
		c.Context(),
		projectID,
		sessionID,
		userID,
		&input,
	)
	if err != nil {
		h.logger.Error("failed to add annotation", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(annotation)
}

// replayActorID resolves the user that owns replay records. Replay sessions and
// annotations are user-foreign-key constrained, so unowned API keys must not be
// attributed by their UUID; roadmapActorID enforces that.
func replayActorID(c *fiber.Ctx) (uuid.UUID, bool) {
	return roadmapActorID(c)
}

func replaySessionServiceError(c *fiber.Ctx, err error, fallback string) error {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		return c.Status(appErr.StatusCode).JSON(fiber.Map{"error": appErr.Message})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fallback})
}
