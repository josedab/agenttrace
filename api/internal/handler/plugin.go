package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PluginHandler handles plugin system HTTP requests
type PluginHandler struct {
	service *service.PluginService
	logger  *zap.Logger
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(svc *service.PluginService, logger *zap.Logger) *PluginHandler {
	return &PluginHandler{
		service: svc,
		logger:  logger,
	}
}

// Install handles POST /api/public/plugins
func (h *PluginHandler) Install(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.PluginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Manifest.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Plugin name is required"})
	}

	plugin, err := h.service.InstallPlugin(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to install plugin", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to install plugin"})
	}

	return c.Status(fiber.StatusCreated).JSON(plugin)
}

// List handles GET /api/public/plugins
func (h *PluginHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	registry, err := h.service.ListPlugins(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list plugins", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list plugins"})
	}

	return c.JSON(registry)
}

// Get handles GET /api/public/plugins/:pluginId
func (h *PluginHandler) Get(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("pluginId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid plugin ID"})
	}

	plugin, err := h.service.GetPlugin(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Plugin not found"})
	}

	return c.JSON(plugin)
}

// Activate handles POST /api/public/plugins/:pluginId/activate
func (h *PluginHandler) Activate(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("pluginId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid plugin ID"})
	}

	plugin, err := h.service.ActivatePlugin(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Plugin not found"})
	}

	return c.JSON(plugin)
}

// Disable handles POST /api/public/plugins/:pluginId/disable
func (h *PluginHandler) Disable(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("pluginId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid plugin ID"})
	}

	plugin, err := h.service.DisablePlugin(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Plugin not found"})
	}

	return c.JSON(plugin)
}

// Execute handles POST /api/public/plugins/:pluginId/execute
func (h *PluginHandler) Execute(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("pluginId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid plugin ID"})
	}

	var body struct {
		Input string `json:"input"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	exec, err := h.service.ExecutePlugin(c.Context(), id, body.Input)
	if err != nil {
		h.logger.Error("failed to execute plugin", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(exec)
}

// Uninstall handles DELETE /api/public/plugins/:pluginId
func (h *PluginHandler) Uninstall(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("pluginId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid plugin ID"})
	}

	if err := h.service.UninstallPlugin(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Plugin not found"})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ListAdapters handles GET /api/public/adapters
func (h *PluginHandler) ListAdapters(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	registry, err := h.service.ListAdapters(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list adapters"})
	}
	return c.JSON(registry)
}

// InstallAdapter handles POST /api/public/adapters
func (h *PluginHandler) InstallAdapter(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AdapterInstallInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	adapter, err := h.service.InstallAdapter(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(adapter)
}

// IngestAdapterEvent handles POST /api/public/adapters/events
func (h *PluginHandler) IngestAdapterEvent(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AdapterEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.IngestAdapterEvent(c.Context(), projectID, &input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "accepted"})
}
