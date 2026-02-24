package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AutoDiscoveryHandler handles auto-discovery HTTP requests
type AutoDiscoveryHandler struct {
	logger  *zap.Logger
	service *service.AutoDiscoveryService
}

// NewAutoDiscoveryHandler creates a new auto-discovery handler
func NewAutoDiscoveryHandler(svc *service.AutoDiscoveryService, logger *zap.Logger) *AutoDiscoveryHandler {
	return &AutoDiscoveryHandler{
		logger:  logger,
		service: svc,
	}
}

// ScanProject handles POST /api/public/auto-discovery/scan
// @Summary Scan project
// @Description Scan a project to discover frameworks and instrumentation points
// @Tags auto-discovery
// @Accept json
// @Produce json
// @Success 200 {object} domain.DiscoveryDashboard
// @Failure 401 {object} map[string]string
// @Router /api/public/auto-discovery/scan [post]
func (h *AutoDiscoveryHandler) ScanProject(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.ScanProject(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to scan project", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan project"})
	}

	return c.JSON(dashboard)
}

// GetFramework handles GET /api/public/auto-discovery/frameworks/:frameworkId
// @Summary Get framework
// @Description Get a discovered framework by ID
// @Tags auto-discovery
// @Accept json
// @Produce json
// @Param frameworkId path string true "Framework ID"
// @Success 200 {object} domain.DiscoveredFramework
// @Failure 400 {object} map[string]string
// @Router /api/public/auto-discovery/frameworks/{frameworkId} [get]
func (h *AutoDiscoveryHandler) GetFramework(c *fiber.Ctx) error {
	frameworkID, err := uuid.Parse(c.Params("frameworkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid framework ID"})
	}

	framework, err := h.service.GetFramework(c.Context(), frameworkID)
	if err != nil {
		h.logger.Error("failed to get framework", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get framework"})
	}

	return c.JSON(framework)
}

// UpdateConfig handles PUT /api/public/auto-discovery/config
// @Summary Update discovery config
// @Description Update the auto-discovery configuration for a project
// @Tags auto-discovery
// @Accept json
// @Produce json
// @Param config body domain.DiscoveryConfig true "Discovery configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/auto-discovery/config [put]
func (h *AutoDiscoveryHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var config domain.DiscoveryConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.UpdateConfig(c.Context(), projectID, config); err != nil {
		h.logger.Error("failed to update discovery config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update config"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// AutoDiscoveryToggleRequest represents the request to toggle instrumentation
type AutoDiscoveryToggleRequest struct {
	Enabled bool `json:"enabled"`
}

// ToggleInstrumentation handles POST /api/public/auto-discovery/frameworks/:frameworkId/toggle
// @Summary Toggle instrumentation
// @Description Enable or disable instrumentation for a discovered framework
// @Tags auto-discovery
// @Accept json
// @Produce json
// @Param frameworkId path string true "Framework ID"
// @Param body body AutoDiscoveryToggleRequest true "Toggle state"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/auto-discovery/frameworks/{frameworkId}/toggle [post]
func (h *AutoDiscoveryHandler) ToggleInstrumentation(c *fiber.Ctx) error {
	frameworkID, err := uuid.Parse(c.Params("frameworkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid framework ID"})
	}

	var req AutoDiscoveryToggleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.ToggleInstrumentation(c.Context(), frameworkID, req.Enabled); err != nil {
		h.logger.Error("failed to toggle instrumentation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to toggle instrumentation"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
