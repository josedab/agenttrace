package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// FleetHandler handles fleet management HTTP requests
type FleetHandler struct {
	service *service.FleetService
	logger  *zap.Logger
}

// NewFleetHandler creates a new fleet handler
func NewFleetHandler(svc *service.FleetService, logger *zap.Logger) *FleetHandler {
	return &FleetHandler{
		service: svc,
		logger:  logger,
	}
}

// GetDashboard handles GET /api/public/fleet/dashboard
func (h *FleetHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get fleet dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}

// ListAgents handles GET /api/public/fleet/agents
func (h *FleetHandler) ListAgents(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	agents, err := h.service.ListAgents(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list fleet agents", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list agents"})
	}

	return c.JSON(agents)
}

// CreatePolicy handles POST /api/public/fleet/policies
func (h *FleetHandler) CreatePolicy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.FleetPolicyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Policy name is required"})
	}

	policy, err := h.service.CreatePolicy(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create fleet policy", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create policy"})
	}

	return c.Status(fiber.StatusCreated).JSON(policy)
}

// ListPolicies handles GET /api/public/fleet/policies
func (h *FleetHandler) ListPolicies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	policies, err := h.service.ListPolicies(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list fleet policies", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list policies"})
	}

	if policies == nil {
		policies = []domain.FleetPolicy{}
	}
	return c.JSON(policies)
}

// BulkUpdate handles POST /api/public/fleet/bulk-update
func (h *FleetHandler) BulkUpdate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var update domain.BulkConfigUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(update.AgentNames) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one agent name is required"})
	}

	count, err := h.service.BulkUpdate(c.Context(), projectID, &update)
	if err != nil {
		h.logger.Error("failed to apply bulk update", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to apply bulk update"})
	}

	return c.JSON(fiber.Map{"updatedAgents": count})
}

// GetScalingRecommendations handles GET /api/public/fleet/scaling
func (h *FleetHandler) GetScalingRecommendations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	recs, err := h.service.GetScalingRecommendations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get scaling recommendations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get recommendations"})
	}

	return c.JSON(recs)
}
