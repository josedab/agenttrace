package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// EvalPlaygroundHandler handles evaluation playground HTTP endpoints
type EvalPlaygroundHandler struct {
	service *service.EvalPlaygroundService
	logger  *zap.Logger
}

// NewEvalPlaygroundHandler creates a new eval playground handler
func NewEvalPlaygroundHandler(svc *service.EvalPlaygroundService, logger *zap.Logger) *EvalPlaygroundHandler {
	return &EvalPlaygroundHandler{service: svc, logger: logger}
}

// CreateSession creates a new playground session
func (h *EvalPlaygroundHandler) CreateSession(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Get("X-Project-ID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project ID"})
	}
	userID, _ := uuid.Parse(c.Get("X-User-ID"))

	var input domain.PlaygroundCreateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	session, err := h.service.CreateSession(c.Context(), projectID, userID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// GetSession retrieves a playground session
func (h *EvalPlaygroundHandler) GetSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session ID"})
	}

	session, err := h.service.GetSession(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(session)
}

// Execute runs evaluator code against trace data
func (h *EvalPlaygroundHandler) Execute(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Get("X-Project-ID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project ID"})
	}

	var input domain.PlaygroundExecuteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.Code == "" || len(input.TraceIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code and traceIds are required"})
	}

	results, err := h.service.Execute(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"results": results})
}

// ListTemplates returns built-in evaluator templates
func (h *EvalPlaygroundHandler) ListTemplates(c *fiber.Ctx) error {
	templates := h.service.ListTemplates(c.Context())
	return c.JSON(fiber.Map{"templates": templates})
}

// ShareSession makes a session publicly accessible
func (h *EvalPlaygroundHandler) ShareSession(c *fiber.Ctx) error {
	var input domain.PlaygroundShareInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	session, err := h.service.ShareSession(c.Context(), input.SessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"shareUrl": "/eval-playground/shared/" + session.ShareToken, "shareToken": session.ShareToken})
}

// GetSharedSession retrieves a shared session by token
func (h *EvalPlaygroundHandler) GetSharedSession(c *fiber.Ctx) error {
	token := c.Params("shareToken")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "share token is required"})
	}

	session, err := h.service.GetSharedSession(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(session)
}
