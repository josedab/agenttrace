package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PromptCIHandler handles prompt regression testing CI HTTP requests
type PromptCIHandler struct {
	logger          *zap.Logger
	promptCIService *service.PromptCIService
}

// NewPromptCIHandler creates a new prompt CI handler
func NewPromptCIHandler(
	promptCIService *service.PromptCIService,
	logger *zap.Logger,
) *PromptCIHandler {
	return &PromptCIHandler{
		logger:          logger,
		promptCIService: promptCIService,
	}
}

// CreateBaseline creates a new prompt regression baseline
// @Summary Create prompt baseline
// @Description Create a new prompt regression testing baseline for a project
// @Tags prompt-ci
// @Accept json
// @Produce json
// @Param baseline body domain.PromptBaselineInput true "Baseline configuration"
// @Success 201 {object} domain.PromptBaseline
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/prompt-ci/baselines [post]
func (h *PromptCIHandler) CreateBaseline(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.PromptBaselineInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	baseline, err := h.promptCIService.CreateBaseline(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create prompt baseline", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create baseline"})
	}

	return c.Status(fiber.StatusCreated).JSON(baseline)
}

// ListBaselines returns all prompt baselines for a project
// @Summary List prompt baselines
// @Description Get all prompt regression testing baselines for a project
// @Tags prompt-ci
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.PromptBaseline
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/prompt-ci/baselines [get]
func (h *PromptCIHandler) ListBaselines(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	baselines, err := h.promptCIService.ListBaselines(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list prompt baselines", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list baselines"})
	}

	return c.JSON(baselines)
}

// GetBaseline returns a specific prompt baseline
// @Summary Get prompt baseline
// @Description Get a specific prompt regression testing baseline by ID
// @Tags prompt-ci
// @Accept json
// @Produce json
// @Param baselineId path string true "Baseline ID"
// @Success 200 {object} domain.PromptBaseline
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/prompt-ci/baselines/{baselineId} [get]
func (h *PromptCIHandler) GetBaseline(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	baselineID, err := uuid.Parse(c.Params("baselineId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid baseline ID"})
	}

	baseline, err := h.promptCIService.GetBaseline(c.Context(), baselineID)
	if err != nil {
		h.logger.Error("failed to get prompt baseline", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get baseline"})
	}

	return c.JSON(baseline)
}

// RunComparisonRequest represents the request to run a prompt comparison
type RunComparisonRequest struct {
	BaselineID string `json:"baselineId"`
	Branch     string `json:"branch"`
	CommitSHA  string `json:"commitSHA"`
}

// RunComparison runs a prompt regression comparison against a baseline
// @Summary Run prompt comparison
// @Description Run a prompt regression comparison against a baseline for a specific branch and commit
// @Tags prompt-ci
// @Accept json
// @Produce json
// @Param body body RunComparisonRequest true "Comparison request"
// @Success 201 {object} domain.PromptComparisonRun
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/prompt-ci/runs [post]
func (h *PromptCIHandler) RunComparison(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req RunComparisonRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	baselineID, err := uuid.Parse(req.BaselineID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid baseline ID"})
	}

	run, err := h.promptCIService.RunComparison(c.Context(), projectID, baselineID, req.Branch, req.CommitSHA)
	if err != nil {
		h.logger.Error("failed to run prompt comparison", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run comparison"})
	}

	return c.Status(fiber.StatusCreated).JSON(run)
}

// ListRuns returns all prompt comparison runs for a project
// @Summary List prompt comparison runs
// @Description Get all prompt regression comparison runs for a project
// @Tags prompt-ci
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.PromptComparisonRun
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/prompt-ci/runs [get]
func (h *PromptCIHandler) ListRuns(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	runs, err := h.promptCIService.ListRuns(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list prompt comparison runs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list runs"})
	}

	return c.JSON(runs)
}
