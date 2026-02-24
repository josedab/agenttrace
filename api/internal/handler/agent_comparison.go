package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentComparisonHandler handles agent comparison HTTP requests
type AgentComparisonHandler struct {
	logger                 *zap.Logger
	agentComparisonService *service.AgentComparisonService
}

// NewAgentComparisonHandler creates a new agent comparison handler
func NewAgentComparisonHandler(
	agentComparisonService *service.AgentComparisonService,
	logger *zap.Logger,
) *AgentComparisonHandler {
	return &AgentComparisonHandler{
		logger:                 logger,
		agentComparisonService: agentComparisonService,
	}
}

// CreateProfile creates a new agent profile
// @Summary Create agent profile
// @Description Create a new agent profile for comparison
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param profile body domain.AgentProfileInput true "Agent profile configuration"
// @Success 201 {object} domain.AgentProfile
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/profiles [post]
func (h *AgentComparisonHandler) CreateProfile(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AgentProfileInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	profile, err := h.agentComparisonService.CreateProfile(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create agent profile", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create profile"})
	}

	return c.Status(fiber.StatusCreated).JSON(profile)
}

// ListProfiles returns all agent profiles for a project
// @Summary List agent profiles
// @Description Get all agent profiles for a project
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.AgentProfileList
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/profiles [get]
func (h *AgentComparisonHandler) ListProfiles(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	profiles, err := h.agentComparisonService.ListProfiles(c.Context(), projectID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list agent profiles", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list profiles"})
	}

	return c.Status(fiber.StatusOK).JSON(profiles)
}

// GetProfile returns a specific agent profile
// @Summary Get agent profile
// @Description Get a specific agent profile by ID
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param profileId path string true "Profile ID"
// @Success 200 {object} domain.AgentProfile
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/profiles/{profileId} [get]
func (h *AgentComparisonHandler) GetProfile(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	profileID, err := uuid.Parse(c.Params("profileId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	profile, err := h.agentComparisonService.GetProfile(c.Context(), projectID, profileID)
	if err != nil {
		h.logger.Error("failed to get agent profile", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get profile"})
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}

// RunComparison runs a comparison between agents
// @Summary Run agent comparison
// @Description Run a comparison between multiple agents
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param comparison body domain.AgentComparisonInput true "Comparison configuration"
// @Success 201 {object} domain.AgentComparisonRun
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/compare [post]
func (h *AgentComparisonHandler) RunComparison(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AgentComparisonInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	comparison, err := h.agentComparisonService.RunComparison(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to run comparison", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run comparison"})
	}

	return c.Status(fiber.StatusCreated).JSON(comparison)
}

// ListComparisons returns all comparison runs for a project
// @Summary List comparisons
// @Description Get all agent comparison runs for a project
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.AgentComparisonList
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/comparisons [get]
func (h *AgentComparisonHandler) ListComparisons(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	comparisons, err := h.agentComparisonService.ListComparisons(c.Context(), projectID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list comparisons", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list comparisons"})
	}

	return c.Status(fiber.StatusOK).JSON(comparisons)
}

// GetComparison returns a specific comparison run
// @Summary Get comparison
// @Description Get a specific agent comparison run by ID
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param comparisonId path string true "Comparison ID"
// @Success 200 {object} domain.AgentComparisonRun
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/comparisons/{comparisonId} [get]
func (h *AgentComparisonHandler) GetComparison(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	comparisonID, err := uuid.Parse(c.Params("comparisonId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid comparison ID"})
	}

	comparison, err := h.agentComparisonService.GetComparison(c.Context(), projectID, comparisonID)
	if err != nil {
		h.logger.Error("failed to get comparison", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get comparison"})
	}

	return c.Status(fiber.StatusOK).JSON(comparison)
}

// GetTrends returns trend data for agent metrics
// @Summary Get agent trends
// @Description Get trend data for agent metrics over time
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Param agentIds query string false "Comma-separated agent IDs"
// @Param metricType query string false "Metric type" default(quality)
// @Param days query int false "Number of days" default(30)
// @Success 200 {array} domain.AgentTrendPoint
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/trends [get]
func (h *AgentComparisonHandler) GetTrends(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var agentIDs []uuid.UUID
	if agentIDsStr := c.Query("agentIds", ""); agentIDsStr != "" {
		for _, idStr := range strings.Split(agentIDsStr, ",") {
			id, err := uuid.Parse(strings.TrimSpace(idStr))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid agent ID: " + idStr})
			}
			agentIDs = append(agentIDs, id)
		}
	}

	metricType := domain.ComparisonMetricType(c.Query("metricType", "quality"))
	days := c.QueryInt("days", 30)

	trends, err := h.agentComparisonService.GetTrends(c.Context(), projectID, agentIDs, metricType, days)
	if err != nil {
		h.logger.Error("failed to get trends", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get trends"})
	}

	return c.Status(fiber.StatusOK).JSON(trends)
}

// GetDashboardSummary returns the agent comparison dashboard summary
// @Summary Get dashboard summary
// @Description Get the agent comparison dashboard summary with top performers
// @Tags agent-comparison
// @Accept json
// @Produce json
// @Success 200 {object} domain.AgentComparisonSummary
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-comparison/summary [get]
func (h *AgentComparisonHandler) GetDashboardSummary(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	summary, err := h.agentComparisonService.GetDashboardSummary(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get dashboard summary", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard summary"})
	}

	return c.Status(fiber.StatusOK).JSON(summary)
}
