package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// MultiModalHandler handles multi-modal trace HTTP requests
type MultiModalHandler struct {
	service *service.MultiModalService
	logger  *zap.Logger
}

// NewMultiModalHandler creates a new multi-modal handler
func NewMultiModalHandler(svc *service.MultiModalService, logger *zap.Logger) *MultiModalHandler {
	return &MultiModalHandler{
		service: svc,
		logger:  logger,
	}
}

// Register handles POST /api/public/multimodal/attachments
func (h *MultiModalHandler) Register(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AttachmentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	attachment, err := h.service.RegisterAttachment(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to register attachment", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register attachment"})
	}

	return c.Status(fiber.StatusCreated).JSON(attachment)
}

// GetTraceAttachments handles GET /api/public/multimodal/traces/:traceId
func (h *MultiModalHandler) GetTraceAttachments(c *fiber.Ctx) error {
	traceID := c.Params("traceId")

	result, err := h.service.GetTraceAttachments(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to get trace attachments", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get trace attachments"})
	}

	return c.JSON(result)
}

// GetAttachment handles GET /api/public/multimodal/attachments/:attachmentId
func (h *MultiModalHandler) GetAttachment(c *fiber.Ctx) error {
	attachmentID := c.Params("attachmentId")

	attachment, err := h.service.GetAttachment(c.Context(), attachmentID)
	if err != nil {
		h.logger.Error("failed to get attachment", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Attachment not found"})
	}

	return c.JSON(attachment)
}

// GetSummary handles GET /api/public/multimodal/traces/:traceId/summary
func (h *MultiModalHandler) GetSummary(c *fiber.Ctx) error {
	traceID := c.Params("traceId")

	summary, err := h.service.GetSummary(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to get summary", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get summary"})
	}

	return c.JSON(summary)
}

// List handles GET /api/public/multimodal/attachments
func (h *MultiModalHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	filter := &domain.AttachmentFilter{
		TraceID: c.Query("traceId"),
		Type:    c.Query("type"),
	}

	attachments, err := h.service.ListAttachments(c.Context(), projectID.String(), filter)
	if err != nil {
		h.logger.Error("failed to list attachments", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list attachments"})
	}

	return c.JSON(attachments)
}
