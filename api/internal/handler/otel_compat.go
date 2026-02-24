package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OTelCompatHandler handles OpenTelemetry compatibility HTTP requests
type OTelCompatHandler struct {
	logger  *zap.Logger
	service *service.OTelCompatService
}

// NewOTelCompatHandler creates a new OTel compatibility handler
func NewOTelCompatHandler(svc *service.OTelCompatService, logger *zap.Logger) *OTelCompatHandler {
	return &OTelCompatHandler{
		logger:  logger,
		service: svc,
	}
}

// CreateExportDestination handles POST /api/public/otel-compat/destinations
// @Summary Create export destination
// @Description Create a new OTel export destination
// @Tags otel-compat
// @Accept json
// @Produce json
// @Param destination body domain.OTelExportDestination true "Export destination"
// @Success 201 {object} domain.OTelExportDestination
// @Failure 400 {object} map[string]string
// @Router /api/public/otel-compat/destinations [post]
func (h *OTelCompatHandler) CreateExportDestination(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.OTelExportDestination
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	destination, err := h.service.CreateDestination(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create export destination", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create export destination"})
	}

	return c.Status(fiber.StatusCreated).JSON(destination)
}

// ListExportDestinations handles GET /api/public/otel-compat/destinations
// @Summary List export destinations
// @Description List all OTel export destinations for a project
// @Tags otel-compat
// @Accept json
// @Produce json
// @Success 200 {array} domain.OTelExportDestination
// @Failure 401 {object} map[string]string
// @Router /api/public/otel-compat/destinations [get]
func (h *OTelCompatHandler) ListExportDestinations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	destinations, err := h.service.ListDestinations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list export destinations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list export destinations"})
	}

	return c.JSON(destinations)
}

// DeleteExportDestination handles DELETE /api/public/otel-compat/destinations/:destinationId
// @Summary Delete export destination
// @Description Delete an OTel export destination
// @Tags otel-compat
// @Accept json
// @Produce json
// @Param destinationId path string true "Destination ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/public/otel-compat/destinations/{destinationId} [delete]
func (h *OTelCompatHandler) DeleteExportDestination(c *fiber.Ctx) error {
	destinationID, err := uuid.Parse(c.Params("destinationId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid destination ID"})
	}

	if err := h.service.DeleteDestination(c.Context(), destinationID); err != nil {
		h.logger.Error("failed to delete export destination", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete export destination"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetOTelMappings handles GET /api/public/otel-compat/mappings
// @Summary Get OTel mappings
// @Description Get the OTel attribute mappings
// @Tags otel-compat
// @Accept json
// @Produce json
// @Success 200 {array} domain.OTelMapping
// @Router /api/public/otel-compat/mappings [get]
func (h *OTelCompatHandler) GetOTelMappings(c *fiber.Ctx) error {
	mappings, err := h.service.GetMappings()
	if err != nil {
		h.logger.Error("failed to get OTel mappings", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get OTel mappings"})
	}

	return c.JSON(mappings)
}

// GetOTelDashboard handles GET /api/public/otel-compat/dashboard
// @Summary Get OTel dashboard
// @Description Get the OTel compatibility dashboard for a project
// @Tags otel-compat
// @Accept json
// @Produce json
// @Success 200 {object} domain.OTelCompatDashboard
// @Failure 401 {object} map[string]string
// @Router /api/public/otel-compat/dashboard [get]
func (h *OTelCompatHandler) GetOTelDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get OTel dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get OTel dashboard"})
	}

	return c.JSON(dashboard)
}

// GenerateCollectorConfig handles POST /api/public/otel-compat/collector-config
// @Summary Generate collector config
// @Description Generate an OTel collector configuration for a project
// @Tags otel-compat
// @Accept json
// @Produce json
// @Success 200 {object} domain.OTelCollectorConfig
// @Failure 401 {object} map[string]string
// @Router /api/public/otel-compat/collector-config [post]
func (h *OTelCompatHandler) GenerateCollectorConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GenerateCollectorConfig(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to generate collector config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate collector config"})
	}

	return c.JSON(config)
}
