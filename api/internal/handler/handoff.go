package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// HandoffHandler handles agent handoff HTTP requests
type HandoffHandler struct {
	service *service.HandoffService
	logger  *zap.Logger
}

// NewHandoffHandler creates a new handoff handler
func NewHandoffHandler(svc *service.HandoffService, logger *zap.Logger) *HandoffHandler {
	return &HandoffHandler{
		service: svc,
		logger:  logger,
	}
}

// Initiate handles POST /api/public/handoffs
func (h *HandoffHandler) Initiate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.HandoffInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.FromAgent == "" || input.ToAgent == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "fromAgent and toAgent are required"})
	}

	handoff, err := h.service.InitiateHandoff(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to initiate handoff", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to initiate handoff"})
	}

	return c.Status(fiber.StatusCreated).JSON(handoff)
}

// Accept handles POST /api/public/handoffs/:handoffId/accept
func (h *HandoffHandler) Accept(c *fiber.Ctx) error {
	handoffID := c.Params("handoffId")
	if handoffID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Handoff ID is required"})
	}

	handoff, err := h.service.AcceptHandoff(c.Context(), handoffID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(handoff)
}

// Complete handles POST /api/public/handoffs/:handoffId/complete
func (h *HandoffHandler) Complete(c *fiber.Ctx) error {
	handoffID := c.Params("handoffId")
	if handoffID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Handoff ID is required"})
	}

	handoff, err := h.service.CompleteHandoff(c.Context(), handoffID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(handoff)
}

// GetChain handles GET /api/public/handoffs/chain/:traceId
func (h *HandoffHandler) GetChain(c *fiber.Ctx) error {
	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	chain, err := h.service.GetChain(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to get handoff chain", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get handoff chain"})
	}

	return c.JSON(chain)
}

// GetStats handles GET /api/public/handoffs/stats
func (h *HandoffHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetStats(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get handoff stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get handoff stats"})
	}

	return c.JSON(stats)
}
