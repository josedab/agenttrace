package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentBuilderHandler handles agent builder HTTP requests
type AgentBuilderHandler struct {
	service *service.AgentBuilderService
	logger  *zap.Logger
}

// NewAgentBuilderHandler creates a new agent builder handler
func NewAgentBuilderHandler(svc *service.AgentBuilderService, logger *zap.Logger) *AgentBuilderHandler {
	return &AgentBuilderHandler{
		service: svc,
		logger:  logger,
	}
}

// Generate handles POST /api/public/agent-builder/generate
func (h *AgentBuilderHandler) Generate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.BuilderInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TaskDescription == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Task description is required"})
	}

	blueprint, err := h.service.GenerateBlueprint(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to generate blueprint", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate blueprint"})
	}

	return c.Status(fiber.StatusCreated).JSON(blueprint)
}

// Get handles GET /api/public/agent-builder/blueprints/:blueprintId
func (h *AgentBuilderHandler) Get(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("blueprintId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid blueprint ID"})
	}

	blueprint, err := h.service.GetBlueprint(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Blueprint not found"})
	}

	return c.JSON(blueprint)
}

// List handles GET /api/public/agent-builder/blueprints
func (h *AgentBuilderHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	blueprints, err := h.service.ListBlueprints(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list blueprints", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list blueprints"})
	}

	if blueprints == nil {
		blueprints = []domain.AgentBlueprint{}
	}
	return c.JSON(blueprints)
}

// Deploy handles POST /api/public/agent-builder/blueprints/:blueprintId/deploy
func (h *AgentBuilderHandler) Deploy(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("blueprintId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid blueprint ID"})
	}

	blueprint, err := h.service.DeployBlueprint(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Blueprint not found"})
	}

	return c.JSON(blueprint)
}
