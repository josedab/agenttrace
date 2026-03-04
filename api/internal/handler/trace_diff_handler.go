package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TraceDiffHandler handles trace diff and bisect HTTP endpoints
type TraceDiffHandler struct {
	service *service.TraceDiffService
	logger  *zap.Logger
}

// NewTraceDiffHandler creates a new trace diff handler
func NewTraceDiffHandler(svc *service.TraceDiffService, logger *zap.Logger) *TraceDiffHandler {
	return &TraceDiffHandler{service: svc, logger: logger}
}

// DiffTraces computes a structural diff between two traces
func (h *TraceDiffHandler) DiffTraces(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Get("X-Project-ID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project ID"})
	}

	var input domain.TraceDiffInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.LeftTraceID == "" || input.RightTraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leftTraceId and rightTraceId are required"})
	}

	result, err := h.service.DiffTraces(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to diff traces", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// StartBisect starts a new regression bisect session
func (h *TraceDiffHandler) StartBisect(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Get("X-Project-ID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project ID"})
	}

	userID, _ := uuid.Parse(c.Get("X-User-ID"))

	var input domain.BisectStartInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.GoodTraceID == "" || input.BadTraceID == "" || len(input.TraceHistory) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "goodTraceId, badTraceId, and traceHistory (min 2) are required"})
	}

	session, err := h.service.StartBisect(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to start bisect", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// GetBisectSession retrieves a bisect session
func (h *TraceDiffHandler) GetBisectSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session ID"})
	}

	session, err := h.service.GetBisectSession(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(session)
}

// SubmitBisectVerdict submits a verdict for the current bisect step
func (h *TraceDiffHandler) SubmitBisectVerdict(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session ID"})
	}

	var input domain.BisectVerdictInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.Verdict != "good" && input.Verdict != "bad" && input.Verdict != "skip" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "verdict must be good, bad, or skip"})
	}

	session, err := h.service.SubmitBisectVerdict(c.Context(), sessionID, &input)
	if err != nil {
		h.logger.Error("failed to submit verdict", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(session)
}

// GetBisectResult returns the result of a completed bisect session
func (h *TraceDiffHandler) GetBisectResult(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session ID"})
	}

	result, err := h.service.GetBisectResult(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ListBisectSessions lists bisect sessions for a project
func (h *TraceDiffHandler) ListBisectSessions(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Get("X-Project-ID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project ID"})
	}

	sessions, err := h.service.ListBisectSessions(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"sessions": sessions})
}
