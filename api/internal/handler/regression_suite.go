package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RegressionSuiteHandler handles regression suite HTTP requests
type RegressionSuiteHandler struct {
	logger  *zap.Logger
	service *service.RegressionSuiteService
}

// NewRegressionSuiteHandler creates a new regression suite handler
func NewRegressionSuiteHandler(svc *service.RegressionSuiteService, logger *zap.Logger) *RegressionSuiteHandler {
	return &RegressionSuiteHandler{
		logger:  logger,
		service: svc,
	}
}

// CreateGoldenDataset handles POST /api/public/regression-suite/datasets
// @Summary Create golden dataset
// @Description Create a new golden dataset for regression testing
// @Tags regression-suite
// @Accept json
// @Produce json
// @Param dataset body domain.GoldenDataset true "Golden dataset"
// @Success 201 {object} domain.GoldenDataset
// @Failure 400 {object} map[string]string
// @Router /api/public/regression-suite/datasets [post]
func (h *RegressionSuiteHandler) CreateGoldenDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.GoldenDataset
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dataset, err := h.service.CreateGoldenDataset(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create golden dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create golden dataset"})
	}

	return c.Status(fiber.StatusCreated).JSON(dataset)
}

// GetGoldenDataset handles GET /api/public/regression-suite/datasets/:datasetId
// @Summary Get golden dataset
// @Description Get a specific golden dataset by ID
// @Tags regression-suite
// @Accept json
// @Produce json
// @Param datasetId path string true "Dataset ID"
// @Success 200 {object} domain.GoldenDataset
// @Failure 400 {object} map[string]string
// @Router /api/public/regression-suite/datasets/{datasetId} [get]
func (h *RegressionSuiteHandler) GetGoldenDataset(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dataset ID"})
	}

	dataset, err := h.service.GetGoldenDataset(c.Context(), datasetID)
	if err != nil {
		h.logger.Error("failed to get golden dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get golden dataset"})
	}

	return c.JSON(dataset)
}

// ListGoldenDatasets handles GET /api/public/regression-suite/datasets
// @Summary List golden datasets
// @Description List all golden datasets for a project
// @Tags regression-suite
// @Accept json
// @Produce json
// @Success 200 {array} domain.GoldenDataset
// @Failure 401 {object} map[string]string
// @Router /api/public/regression-suite/datasets [get]
func (h *RegressionSuiteHandler) ListGoldenDatasets(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasets, err := h.service.ListGoldenDatasets(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list golden datasets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list golden datasets"})
	}

	return c.JSON(datasets)
}

// RegressionSuiteRunRequest represents the request to run a regression
type RegressionSuiteRunRequest struct {
	SuiteID     string `json:"suiteId"`
	AgentConfig string `json:"agentConfig"`
}

// RunRegression handles POST /api/public/regression-suite/runs
// @Summary Run regression
// @Description Run a regression test suite
// @Tags regression-suite
// @Accept json
// @Produce json
// @Param body body RegressionSuiteRunRequest true "Regression run parameters"
// @Success 201 {object} domain.RegressionRun
// @Failure 400 {object} map[string]string
// @Router /api/public/regression-suite/runs [post]
func (h *RegressionSuiteHandler) RunRegression(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req RegressionSuiteRunRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	suiteID, err := uuid.Parse(req.SuiteID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	run, err := h.service.RunRegression(c.Context(), projectID, suiteID, req.AgentConfig)
	if err != nil {
		h.logger.Error("failed to run regression", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run regression"})
	}

	return c.Status(fiber.StatusCreated).JSON(run)
}

// GetRegressionRun handles GET /api/public/regression-suite/runs/:runId
// @Summary Get regression run
// @Description Get a specific regression run by ID
// @Tags regression-suite
// @Accept json
// @Produce json
// @Param runId path string true "Run ID"
// @Success 200 {object} domain.RegressionRun
// @Failure 400 {object} map[string]string
// @Router /api/public/regression-suite/runs/{runId} [get]
func (h *RegressionSuiteHandler) GetRegressionRun(c *fiber.Ctx) error {
	runID, err := uuid.Parse(c.Params("runId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid run ID"})
	}

	run, err := h.service.GetRun(c.Context(), runID)
	if err != nil {
		h.logger.Error("failed to get regression run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get regression run"})
	}

	return c.JSON(run)
}

// ListRegressionRuns handles GET /api/public/regression-suite/runs
// @Summary List regression runs
// @Description List all regression runs for a project
// @Tags regression-suite
// @Accept json
// @Produce json
// @Success 200 {array} domain.RegressionRun
// @Failure 401 {object} map[string]string
// @Router /api/public/regression-suite/runs [get]
func (h *RegressionSuiteHandler) ListRegressionRuns(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	runs, err := h.service.ListRuns(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list regression runs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list regression runs"})
	}

	return c.JSON(runs)
}
