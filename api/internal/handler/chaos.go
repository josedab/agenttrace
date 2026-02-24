package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ChaosHandler handles chaos testing HTTP requests
type ChaosHandler struct {
	service *service.ChaosService
	logger  *zap.Logger
}

// NewChaosHandler creates a new chaos handler
func NewChaosHandler(svc *service.ChaosService, logger *zap.Logger) *ChaosHandler {
	return &ChaosHandler{
		service: svc,
		logger:  logger,
	}
}

// Create handles POST /api/public/chaos/experiments
func (h *ChaosHandler) Create(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.ChaosExperimentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	experiment, err := h.service.CreateExperiment(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create experiment", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create experiment"})
	}

	return c.Status(fiber.StatusCreated).JSON(experiment)
}

// Run handles POST /api/public/chaos/experiments/:experimentId/run
func (h *ChaosHandler) Run(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	experimentID, err := uuid.Parse(c.Params("experimentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid experiment ID"})
	}

	experiment, err := h.service.RunExperiment(c.Context(), projectID, experimentID)
	if err != nil {
		h.logger.Error("failed to run experiment", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run experiment"})
	}

	return c.JSON(experiment)
}

// Get handles GET /api/public/chaos/experiments/:experimentId
func (h *ChaosHandler) Get(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	experimentID, err := uuid.Parse(c.Params("experimentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid experiment ID"})
	}

	experiment, err := h.service.GetExperiment(c.Context(), projectID, experimentID)
	if err != nil {
		h.logger.Error("failed to get experiment", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Experiment not found"})
	}

	return c.JSON(experiment)
}

// List handles GET /api/public/chaos/experiments
func (h *ChaosHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	experiments, err := h.service.ListExperiments(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list experiments", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list experiments"})
	}

	return c.JSON(experiments)
}

// GetScorecard handles GET /api/public/chaos/scorecard/:agentName
func (h *ChaosHandler) GetScorecard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	agentName := c.Params("agentName")
	if agentName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Agent name is required"})
	}

	scorecard, err := h.service.GetResilienceScorecard(c.Context(), projectID, agentName)
	if err != nil {
		h.logger.Error("failed to get scorecard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get scorecard"})
	}

	return c.JSON(scorecard)
}
