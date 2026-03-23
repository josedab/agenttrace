package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AnnotationHandler handles collaborative annotation HTTP requests
type AnnotationHandler struct {
	service *service.AnnotationService
	logger  *zap.Logger
}

// NewAnnotationHandler creates a new annotation handler
func NewAnnotationHandler(svc *service.AnnotationService, logger *zap.Logger) *AnnotationHandler {
	return &AnnotationHandler{
		service: svc,
		logger:  logger,
	}
}

// Create handles POST /api/public/annotations
func (h *AnnotationHandler) Create(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AnnotationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Content == "" || input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "traceId and content are required"})
	}

	// Use a mock user for now
	userID := "user-1"
	userName := "Developer"

	annotation, err := h.service.CreateAnnotation(c.Context(), projectID.String(), userID, userName, &input)
	if err != nil {
		h.logger.Error("failed to create annotation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create annotation"})
	}

	return c.Status(fiber.StatusCreated).JSON(annotation)
}

// List handles GET /api/public/annotations/traces/:traceId
func (h *AnnotationHandler) List(c *fiber.Ctx) error {
	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	annotations, err := h.service.ListAnnotations(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to list annotations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list annotations"})
	}

	return c.JSON(annotations)
}

// Reply handles POST /api/public/annotations/:annotationId/reply
func (h *AnnotationHandler) Reply(c *fiber.Ctx) error {
	annotationID := c.Params("annotationId")
	if annotationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Annotation ID is required"})
	}

	var input domain.ReplyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	// Use a mock user for now
	userID := "user-1"
	userName := "Developer"

	annotation, err := h.service.AddReply(c.Context(), annotationID, userID, userName, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(annotation)
}

// Resolve handles POST /api/public/annotations/:annotationId/resolve
func (h *AnnotationHandler) Resolve(c *fiber.Ctx) error {
	annotationID := c.Params("annotationId")
	if annotationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Annotation ID is required"})
	}

	annotation, err := h.service.ResolveAnnotation(c.Context(), annotationID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(annotation)
}

// GetPresence handles GET /api/public/annotations/presence/:traceId
func (h *AnnotationHandler) GetPresence(c *fiber.Ctx) error {
	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	presence, err := h.service.GetPresence(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to get presence", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get presence"})
	}

	return c.JSON(presence)
}
