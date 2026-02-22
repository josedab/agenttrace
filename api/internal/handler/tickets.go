package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TicketHandler handles ticket HTTP requests
type TicketHandler struct {
	ticketService *service.TicketService
	logger        *zap.Logger
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *service.TicketService, logger *zap.Logger) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
		logger:        logger,
	}
}

// CreateTicket handles POST /tickets
func (h *TicketHandler) CreateTicket(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.TicketCreateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	result, err := h.ticketService.CreateTicket(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create ticket",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create ticket",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ListTickets handles GET /tickets?traceId=
func (h *TicketHandler) ListTickets(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Query("traceId")

	tickets, err := h.ticketService.ListTickets(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to list tickets",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list tickets",
		})
	}

	return c.JSON(tickets)
}

// ConfigureIntegration handles POST /tickets/integrations
func (h *TicketHandler) ConfigureIntegration(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var integration domain.TicketIntegration
	if err := c.BodyParser(&integration); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if err := h.ticketService.ConfigureIntegration(c.Context(), projectID, integration); err != nil {
		h.logger.Error("failed to configure ticket integration",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to configure ticket integration",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "configured"})
}

// GetIntegrations handles GET /tickets/integrations
func (h *TicketHandler) GetIntegrations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	integrations, err := h.ticketService.GetIntegrations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get ticket integrations",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get ticket integrations",
		})
	}

	return c.JSON(integrations)
}

// PreviewTicketRequest represents the request body for ticket preview
type PreviewTicketRequest struct {
	TraceID string `json:"traceId"`
}

// PreviewTicket handles POST /tickets/preview
func (h *TicketHandler) PreviewTicket(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var req PreviewTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if req.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	template, err := h.ticketService.BuildTicketBody(c.Context(), projectID, req.TraceID)
	if err != nil {
		h.logger.Error("failed to preview ticket",
			zap.String("projectId", projectID.String()),
			zap.String("traceId", req.TraceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to preview ticket",
		})
	}

	return c.JSON(template)
}
