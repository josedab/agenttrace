package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentVersionHandler handles agent versioning HTTP requests
type AgentVersionHandler struct {
	service *service.AgentVersionService
	logger  *zap.Logger
}

// NewAgentVersionHandler creates a new agent version handler
func NewAgentVersionHandler(svc *service.AgentVersionService, logger *zap.Logger) *AgentVersionHandler {
	return &AgentVersionHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateVersion handles POST /api/public/agent-versions
func (h *AgentVersionHandler) CreateVersion(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CreateVersionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.AgentName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agentName is required"})
	}

	version, err := h.service.CreateVersion(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create agent version", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create version"})
	}

	return c.Status(fiber.StatusCreated).JSON(version)
}

// GetVersion handles GET /api/public/agent-versions/:versionId
func (h *AgentVersionHandler) GetVersion(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	versionID, err := uuid.Parse(c.Params("versionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid version ID"})
	}

	version, err := h.service.GetVersion(c.Context(), versionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Version not found"})
	}

	return c.JSON(version)
}

// ListVersions handles GET /api/public/agent-versions
func (h *AgentVersionHandler) ListVersions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	agentName := c.Query("agentName")
	versions, err := h.service.ListVersions(c.Context(), projectID, agentName)
	if err != nil {
		h.logger.Error("failed to list agent versions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list versions"})
	}

	return c.JSON(fiber.Map{"versions": versions, "count": len(versions)})
}

// GetActive handles GET /api/public/agent-versions/active
func (h *AgentVersionHandler) GetActive(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	agentName := c.Query("agentName")
	if agentName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agentName query parameter is required"})
	}

	version, err := h.service.GetActiveVersion(c.Context(), projectID, agentName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(version)
}

// Rollback handles POST /api/public/agent-versions/:versionId/rollback
func (h *AgentVersionHandler) Rollback(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	versionID, err := uuid.Parse(c.Params("versionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid version ID"})
	}

	version, err := h.service.Rollback(c.Context(), projectID, versionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(version)
}

// DiffVersions handles POST /api/public/agent-versions/diff
func (h *AgentVersionHandler) DiffVersions(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input struct {
		VersionA uuid.UUID `json:"versionA"`
		VersionB uuid.UUID `json:"versionB"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.VersionA == uuid.Nil || input.VersionB == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "versionA and versionB are required"})
	}

	diff, err := h.service.DiffVersions(c.Context(), input.VersionA, input.VersionB)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(diff)
}
