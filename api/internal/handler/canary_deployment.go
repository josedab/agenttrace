package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CanaryDeploymentHandler handles canary deployment HTTP requests
type CanaryDeploymentHandler struct {
	service *service.CanaryDeploymentService
	logger  *zap.Logger
}

// NewCanaryDeploymentHandler creates a new canary deployment handler
func NewCanaryDeploymentHandler(svc *service.CanaryDeploymentService, logger *zap.Logger) *CanaryDeploymentHandler {
	return &CanaryDeploymentHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateDeployment handles POST /api/public/canary/deployments
func (h *CanaryDeploymentHandler) CreateDeployment(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.CanaryDeploymentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.BaselineVersion == "" || input.CanaryVersion == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Both baseline and canary versions are required"})
	}

	deployment, err := h.service.CreateDeployment(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create canary deployment", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(deployment)
}

// GetDeployment handles GET /api/public/canary/deployments/:deploymentId
func (h *CanaryDeploymentHandler) GetDeployment(c *fiber.Ctx) error {
	depID, err := uuid.Parse(c.Params("deploymentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid deployment ID"})
	}

	dep, err := h.service.GetDeployment(c.Context(), depID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Deployment not found"})
	}

	return c.JSON(dep)
}

// ListDeployments handles GET /api/public/canary/deployments
func (h *CanaryDeploymentHandler) ListDeployments(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	list, err := h.service.ListDeployments(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(list)
}

// Promote handles POST /api/public/canary/deployments/:deploymentId/promote
func (h *CanaryDeploymentHandler) Promote(c *fiber.Ctx) error {
	depID, err := uuid.Parse(c.Params("deploymentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid deployment ID"})
	}

	dep, err := h.service.Promote(c.Context(), depID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dep)
}

// Rollback handles POST /api/public/canary/deployments/:deploymentId/rollback
func (h *CanaryDeploymentHandler) Rollback(c *fiber.Ctx) error {
	depID, err := uuid.Parse(c.Params("deploymentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid deployment ID"})
	}

	dep, err := h.service.Rollback(c.Context(), depID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dep)
}

// GetMetrics handles GET /api/public/canary/deployments/:deploymentId/metrics
func (h *CanaryDeploymentHandler) GetMetrics(c *fiber.Ctx) error {
	depID, err := uuid.Parse(c.Params("deploymentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid deployment ID"})
	}

	metrics, err := h.service.GetMetrics(c.Context(), depID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(metrics)
}

// GetActiveVersion handles GET /api/public/canary/active-version
func (h *CanaryDeploymentHandler) GetActiveVersion(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	version, err := h.service.GetActiveVersion(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(version)
}
