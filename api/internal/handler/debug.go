package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// DebugHandler handles debug session HTTP requests
type DebugHandler struct {
	debugService *service.DebugService
	logger       *zap.Logger
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler(debugService *service.DebugService, logger *zap.Logger) *DebugHandler {
	return &DebugHandler{
		debugService: debugService,
		logger:       logger,
	}
}

// CreateSession handles POST /debug
func (h *DebugHandler) CreateSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	userID, _ := middleware.GetUserID(c)

	var input domain.CreateDebugSessionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	session, err := h.debugService.CreateSession(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to create debug session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create debug session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

// GetSession handles GET /debug/:sessionId
func (h *DebugHandler) GetSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid session ID",
		})
	}

	h.logger.Debug("getting debug session",
		zap.String("sessionId", sessionID.String()),
		zap.String("projectId", projectID.String()),
	)

	session, err := h.debugService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get debug session", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Debug session not found",
		})
	}

	return c.JSON(session)
}

// GetStepState handles GET /debug/traces/:traceId/debug/step/:stepIndex
func (h *DebugHandler) GetStepState(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Trace ID is required",
		})
	}

	stepIndex, err := strconv.Atoi(c.Params("stepIndex"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid step index",
		})
	}

	state, err := h.debugService.GetStepState(c.Context(), projectID, traceID, stepIndex)
	if err != nil {
		h.logger.Error("failed to get step state",
			zap.String("traceId", traceID),
			zap.Int("stepIndex", stepIndex),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get step state",
		})
	}

	return c.JSON(state)
}

// AddAnnotation handles POST /debug/:sessionId/annotations
func (h *DebugHandler) AddAnnotation(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	userID, _ := middleware.GetUserID(c)

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid session ID",
		})
	}

	var body struct {
		Content string `json:"content"`
		EventID string `json:"eventId"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if body.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "content is required",
		})
	}

	annotation, err := h.debugService.AddAnnotation(c.Context(), sessionID, body.EventID, userID, body.Content)
	if err != nil {
		h.logger.Error("failed to add annotation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to add annotation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(annotation)
}
