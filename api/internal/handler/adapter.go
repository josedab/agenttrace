package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AdapterHandler handles adapter HTTP requests
type AdapterHandler struct {
	service *service.AdapterService
	logger  *zap.Logger
}

// NewAdapterHandler creates a new adapter handler
func NewAdapterHandler(svc *service.AdapterService, logger *zap.Logger) *AdapterHandler {
	return &AdapterHandler{
		service: svc,
		logger:  logger,
	}
}

// RegisterAdapter handles POST /api/public/adapters
func (h *AdapterHandler) RegisterAdapter(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AdapterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Adapter name is required"})
	}

	adapter, err := h.service.RegisterAdapter(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to register adapter", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register adapter"})
	}

	return c.Status(fiber.StatusCreated).JSON(adapter)
}

// GetAdapter handles GET /api/public/adapters/:adapterId
func (h *AdapterHandler) GetAdapter(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("adapterId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid adapter ID"})
	}

	adapter, err := h.service.GetAdapter(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Adapter not found"})
	}

	return c.JSON(adapter)
}

// ListAdapters handles GET /api/public/adapters
func (h *AdapterHandler) ListAdapters(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	adapters, err := h.service.ListAdapters(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list adapters", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list adapters"})
	}

	return c.JSON(fiber.Map{"adapters": adapters})
}

// UpdateAdapter handles PUT /api/public/adapters/:adapterId
func (h *AdapterHandler) UpdateAdapter(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("adapterId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid adapter ID"})
	}

	var input domain.AdapterUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	adapter, err := h.service.UpdateAdapter(c.Context(), id, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Adapter not found"})
	}

	return c.JSON(adapter)
}

// DeleteAdapter handles DELETE /api/public/adapters/:adapterId
func (h *AdapterHandler) DeleteAdapter(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("adapterId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid adapter ID"})
	}

	if err := h.service.DeleteAdapter(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Adapter not found"})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// IngestEvent handles POST /api/public/adapters/:adapterId/events
func (h *AdapterHandler) IngestEvent(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	adapterID, err := uuid.Parse(c.Params("adapterId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid adapter ID"})
	}

	var event domain.AdapterEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	event.AdapterID = adapterID

	if err := h.service.IngestEvent(c.Context(), adapterID, &event); err != nil {
		h.logger.Error("failed to ingest adapter event", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "accepted"})
}

// TestAdapter handles POST /api/public/adapters/:adapterId/test
func (h *AdapterHandler) TestAdapter(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("adapterId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid adapter ID"})
	}

	result, err := h.service.TestAdapter(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Adapter not found"})
	}

	return c.JSON(result)
}

// GetTemplates handles GET /api/public/adapters/templates
func (h *AdapterHandler) GetTemplates(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	templates := h.service.GetTemplates(c.Context())
	return c.JSON(fiber.Map{"templates": templates})
}
