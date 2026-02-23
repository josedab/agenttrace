package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CollaborationHandler handles collaboration HTTP requests
type CollaborationHandler struct {
	collaborationService *service.CollaborationService
	logger               *zap.Logger
}

// NewCollaborationHandler creates a new collaboration handler
func NewCollaborationHandler(collaborationService *service.CollaborationService, logger *zap.Logger) *CollaborationHandler {
	return &CollaborationHandler{
		collaborationService: collaborationService,
		logger:               logger,
	}
}

// GetPresence handles GET /collaboration/traces/:traceId/presence
func (h *CollaborationHandler) GetPresence(c *fiber.Ctx) error {
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

	presence, err := h.collaborationService.GetPresence(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get presence",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get presence",
		})
	}

	return c.JSON(presence)
}

// AddAnnotation handles POST /collaboration/traces/:traceId/annotations
func (h *CollaborationHandler) AddAnnotation(c *fiber.Ctx) error {
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

	var annotation domain.TraceAnnotation
	if err := c.BodyParser(&annotation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}
	annotation.TraceID = traceID

	userID, _ := middleware.GetUserID(c)
	annotation.UserID = userID

	result, err := h.collaborationService.AddAnnotation(c.Context(), projectID, &annotation)
	if err != nil {
		h.logger.Error("failed to add annotation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to add annotation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ListAnnotations handles GET /collaboration/traces/:traceId/annotations
func (h *CollaborationHandler) ListAnnotations(c *fiber.Ctx) error {
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

	annotations, err := h.collaborationService.ListAnnotations(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to list annotations",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list annotations",
		})
	}

	return c.JSON(annotations)
}

// ResolveAnnotation handles POST /collaboration/annotations/:annotationId/resolve
func (h *CollaborationHandler) ResolveAnnotation(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	annotationID, err := uuid.Parse(c.Params("annotationId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid annotation ID",
		})
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.collaborationService.ResolveAnnotation(c.Context(), annotationID, userID); err != nil {
		h.logger.Error("failed to resolve annotation",
			zap.String("annotationId", annotationID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to resolve annotation",
		})
	}

	return c.JSON(fiber.Map{"status": "resolved"})
}

// CreateDiscussion handles POST /api/public/collaboration/discussions
func (h *CollaborationHandler) CreateDiscussion(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.DiscussionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Title == "" || input.InitialMessage == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title and initial message are required"})
	}

	// Use project ID as placeholder for user ID since we have API key auth
	thread, err := h.collaborationService.CreateDiscussionThread(c.Context(), projectID, projectID, "API User", &input)
	if err != nil {
		h.logger.Error("failed to create discussion", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create discussion"})
	}

	return c.Status(fiber.StatusCreated).JSON(thread)
}

// AddMessage handles POST /api/public/collaboration/discussions/:threadId/messages
func (h *CollaborationHandler) AddMessage(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	threadID, err := uuid.Parse(c.Params("threadId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid thread ID"})
	}

	var input struct {
		Content  string      `json:"content"`
		Mentions []uuid.UUID `json:"mentions"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	msg, err := h.collaborationService.AddThreadMessage(c.Context(), threadID, uuid.New(), "API User", input.Content, input.Mentions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add message"})
	}

	return c.Status(fiber.StatusCreated).JSON(msg)
}

// CreateEvalQueue handles POST /api/public/collaboration/eval-queues
func (h *CollaborationHandler) CreateEvalQueue(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EvalQueueInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	queue, err := h.collaborationService.CreateEvalQueue(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create eval queue"})
	}

	return c.Status(fiber.StatusCreated).JSON(queue)
}

// CreateSharedSession handles POST /collaboration/sessions
func (h *CollaborationHandler) CreateSharedSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var session domain.SharedSession
	if err := c.BodyParser(&session); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	userID, _ := middleware.GetUserID(c)
	session.CreatedBy = userID

	result, err := h.collaborationService.CreateSharedSession(c.Context(), projectID, &session)
	if err != nil {
		h.logger.Error("failed to create shared session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create shared session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}
