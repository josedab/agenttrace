package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// SessionJourneyHandler handles session-based trace journey HTTP requests
type SessionJourneyHandler struct {
	service *service.SessionJourneyService
	logger  *zap.Logger
}

// NewSessionJourneyHandler creates a new session journey handler
func NewSessionJourneyHandler(svc *service.SessionJourneyService, logger *zap.Logger) *SessionJourneyHandler {
	return &SessionJourneyHandler{
		service: svc,
		logger:  logger,
	}
}

// GetJourney handles GET /sessions/:sessionId/journey
func (h *SessionJourneyHandler) GetJourney(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Session ID is required"})
	}

	result, err := h.service.GetJourney(c.Context(), projectID, sessionID)
	if err != nil {
		h.logger.Error("failed to get journey", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get journey"})
	}

	return c.JSON(result)
}

// GetPhases handles GET /sessions/:sessionId/phases
func (h *SessionJourneyHandler) GetPhases(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Session ID is required"})
	}

	result, err := h.service.GetPhases(c.Context(), projectID, sessionID)
	if err != nil {
		h.logger.Error("failed to get phases", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get phases"})
	}

	return c.JSON(result)
}

// ListRecentJourneys handles GET /session-journeys/recent
func (h *SessionJourneyHandler) ListRecentJourneys(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.ListRecentJourneys(c.Context(), projectID, 20)
	if err != nil {
		h.logger.Error("failed to list recent journeys", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list recent journeys"})
	}

	return c.JSON(result)
}
