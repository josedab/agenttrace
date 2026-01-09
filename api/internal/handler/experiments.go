package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ExperimentHandler handles experiment-related HTTP requests
type ExperimentHandler struct {
	logger            *zap.Logger
	experimentService *service.ExperimentService
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(
	logger *zap.Logger,
	experimentService *service.ExperimentService,
) *ExperimentHandler {
	return &ExperimentHandler{
		logger:            logger,
		experimentService: experimentService,
	}
}

// ListExperiments returns all experiments for a project
// @Summary List experiments
// @Description Get all experiments for a project
// @Tags experiments
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name"
// @Success 200 {object} domain.ExperimentList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/experiments [get]
func (h *ExperimentHandler) ListExperiments(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	filter := domain.ExperimentFilter{
		ProjectID: projectID,
	}

	if status := c.Query("status"); status != "" {
		s := domain.ExperimentStatus(status)
		filter.Status = &s
	}

	filter.Search = c.Query("search")

	// For now, return empty list - actual implementation would query database
	result := domain.ExperimentList{
		Experiments: []domain.Experiment{},
		TotalCount:  0,
		HasMore:     false,
	}

	return c.JSON(result)
}

// GetExperiment returns a specific experiment
// @Summary Get experiment
// @Description Get a specific experiment by ID
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} domain.Experiment
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id} [get]
func (h *ExperimentHandler) GetExperiment(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	h.logger.Debug("Get experiment", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// CreateExperiment creates a new experiment
// @Summary Create experiment
// @Description Create a new A/B test experiment
// @Tags experiments
// @Accept json
// @Produce json
// @Param experiment body domain.ExperimentInput true "Experiment configuration"
// @Success 201 {object} domain.Experiment
// @Failure 400 {object} ErrorResponse
// @Router /api/public/experiments [post]
func (h *ExperimentHandler) CreateExperiment(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	userID := uuid.New() // In real implementation, get from auth context

	var input domain.ExperimentInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate
	if input.Name == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Name is required", nil)
	}
	if len(input.Variants) < 2 {
		return errorResponse(c, fiber.StatusBadRequest, "At least 2 variants are required", nil)
	}
	if input.TargetMetric == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Target metric is required", nil)
	}

	experiment, err := h.experimentService.CreateExperiment(c.Context(), projectID, userID, &input)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return c.Status(fiber.StatusCreated).JSON(experiment)
}

// UpdateExperiment updates an experiment
// @Summary Update experiment
// @Description Update an experiment's configuration
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Param experiment body domain.ExperimentUpdateInput true "Updated configuration"
// @Success 200 {object} domain.Experiment
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id} [patch]
func (h *ExperimentHandler) UpdateExperiment(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	var input domain.ExperimentUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Debug("Update experiment", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// StartExperiment starts an experiment
// @Summary Start experiment
// @Description Start running an experiment
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} domain.Experiment
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id}/start [post]
func (h *ExperimentHandler) StartExperiment(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	h.logger.Info("Starting experiment", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// PauseExperiment pauses a running experiment
// @Summary Pause experiment
// @Description Pause a running experiment
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} domain.Experiment
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id}/pause [post]
func (h *ExperimentHandler) PauseExperiment(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	h.logger.Info("Pausing experiment", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// CompleteExperiment completes an experiment and calculates results
// @Summary Complete experiment
// @Description Complete an experiment and calculate final results
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} domain.Experiment
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id}/complete [post]
func (h *ExperimentHandler) CompleteExperiment(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	h.logger.Info("Completing experiment", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// GetResults returns the analysis results for an experiment
// @Summary Get experiment results
// @Description Get statistical analysis results for an experiment
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} domain.ExperimentResults
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id}/results [get]
func (h *ExperimentHandler) GetResults(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	h.logger.Debug("Get experiment results", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// AssignVariant assigns a trace to an experiment variant
// @Summary Assign variant
// @Description Assign a trace to a variant for an experiment
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Param body body AssignVariantRequest true "Assignment request"
// @Success 200 {object} domain.ExperimentAssignment
// @Failure 400 {object} ErrorResponse
// @Router /api/public/experiments/{id}/assign [post]
func (h *ExperimentHandler) AssignVariant(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	var req AssignVariantRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Debug("Assign variant",
		zap.String("experimentId", experimentID.String()),
		zap.String("traceId", req.TraceID.String()),
	)

	// In real implementation, would fetch experiment and call service
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}

// AssignVariantRequest represents the request to assign a variant
type AssignVariantRequest struct {
	TraceID uuid.UUID `json:"traceId"`
	UserID  string    `json:"userId,omitempty"`
}

// RecordMetric records a metric value for an experiment
// @Summary Record metric
// @Description Record a metric value for experiment analysis
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Param body body RecordMetricRequest true "Metric data"
// @Success 201 "Created"
// @Failure 400 {object} ErrorResponse
// @Router /api/public/experiments/{id}/metrics [post]
func (h *ExperimentHandler) RecordMetric(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	var req RecordMetricRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Debug("Record metric",
		zap.String("experimentId", experimentID.String()),
		zap.String("traceId", req.TraceID.String()),
		zap.Float64("value", req.Value),
	)

	return c.SendStatus(fiber.StatusCreated)
}

// RecordMetricRequest represents the request to record a metric
type RecordMetricRequest struct {
	TraceID   uuid.UUID `json:"traceId"`
	VariantID uuid.UUID `json:"variantId"`
	Value     float64   `json:"value"`
}

// DeleteExperiment deletes an experiment
// @Summary Delete experiment
// @Description Delete an experiment (must be in draft or archived status)
// @Tags experiments
// @Accept json
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/experiments/{id} [delete]
func (h *ExperimentHandler) DeleteExperiment(c *fiber.Ctx) error {
	experimentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid experiment ID", err)
	}

	h.logger.Info("Delete experiment", zap.String("experimentId", experimentID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Experiment not found", nil)
}
