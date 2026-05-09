package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// FederatedAggregationHandler handles federated trace aggregation HTTP requests
type FederatedAggregationHandler struct {
	logger                      *zap.Logger
	federatedAggregationService *service.FederatedAggregationService
}

// NewFederatedAggregationHandler creates a new federated aggregation handler
func NewFederatedAggregationHandler(
	federatedAggregationService *service.FederatedAggregationService,
	logger *zap.Logger,
) *FederatedAggregationHandler {
	return &FederatedAggregationHandler{
		logger:                      logger,
		federatedAggregationService: federatedAggregationService,
	}
}

// GetDashboard returns the federated aggregation dashboard
// @Summary Get federated aggregation dashboard
// @Description Get the federated trace aggregation dashboard with cross-instance metrics
// @Tags federated-aggregation
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.FederatedDashboard
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/federated-aggregation/dashboard [get]
func (h *FederatedAggregationHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.federatedAggregationService.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get federated aggregation dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}

// RegisterInstance registers a new instance for federated aggregation
// @Summary Register federated instance
// @Description Register a new instance to participate in federated trace aggregation
// @Tags federated-aggregation
// @Accept json
// @Produce json
// @Param instance body domain.FederatedInstanceInput true "Instance configuration"
// @Success 201 {object} domain.FederatedInstance
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/federated-aggregation/instances [post]
func (h *FederatedAggregationHandler) RegisterInstance(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.FederatedInstanceInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	instance, err := h.federatedAggregationService.RegisterInstance(c.Context(), input)
	if err != nil {
		h.logger.Error("failed to register federated instance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register instance"})
	}

	return c.Status(fiber.StatusCreated).JSON(instance)
}

// ListInstances returns all registered federated instances
// @Summary List federated instances
// @Description Get all registered federated instances for a project
// @Tags federated-aggregation
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.FederatedInstance
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/federated-aggregation/instances [get]
func (h *FederatedAggregationHandler) ListInstances(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	instances, err := h.federatedAggregationService.ListInstances(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list federated instances", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list instances"})
	}

	return c.JSON(instances)
}

// SubmitMetricsRequest represents the request to submit federated metrics
type SubmitMetricsRequest struct {
	InstanceID string                   `json:"instanceId"`
	Metrics    []domain.FederatedMetric `json:"metrics"`
}

// SubmitMetrics submits metrics from a federated instance
// @Summary Submit federated metrics
// @Description Submit aggregated metrics from a federated instance
// @Tags federated-aggregation
// @Accept json
// @Produce json
// @Param metrics body SubmitMetricsRequest true "Metrics data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/federated-aggregation/metrics [post]
func (h *FederatedAggregationHandler) SubmitMetrics(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req SubmitMetricsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	instanceID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid instance ID"})
	}

	err = h.federatedAggregationService.SubmitMetrics(c.Context(), instanceID, req.Metrics)
	if err != nil {
		h.logger.Error("failed to submit federated metrics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit metrics"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "accepted"})
}

// GetBenchmarks returns federated benchmarks
// @Summary Get federated benchmarks
// @Description Get cross-instance benchmarks filtered by metric type
// @Tags federated-aggregation
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param metricType query string false "Filter by metric type"
// @Success 200 {array} domain.FederatedBenchmark
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/federated-aggregation/benchmarks [get]
func (h *FederatedAggregationHandler) GetBenchmarks(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	metricType := domain.FederatedMetricType(c.Query("metricType"))

	benchmarks, err := h.federatedAggregationService.GetBenchmarks(c.Context(), projectID, metricType)
	if err != nil {
		h.logger.Error("failed to get federated benchmarks", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get benchmarks"})
	}

	return c.JSON(benchmarks)
}

// GetInsights returns federated aggregation insights
// @Summary Get federated insights
// @Description Get cross-instance insights and recommendations from federated aggregation
// @Tags federated-aggregation
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.FederatedInsights
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/federated-aggregation/insights [get]
func (h *FederatedAggregationHandler) GetInsights(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	insights, err := h.federatedAggregationService.GetInsights(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get federated insights", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get insights"})
	}

	return c.JSON(insights)
}

// SubmitAnonymizedBenchmark handles POST /api/public/federated-aggregation/anonymized-benchmark
func (h *FederatedAggregationHandler) SubmitAnonymizedBenchmark(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input struct {
		Metrics       []domain.AnonymizedMetric         `json:"metrics"`
		PrivacyConfig *domain.DifferentialPrivacyConfig `json:"privacyConfig,omitempty"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(input.Metrics) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one metric is required"})
	}

	submission, err := h.federatedAggregationService.SubmitAnonymizedBenchmark(
		c.Context(), projectID, input.Metrics, input.PrivacyConfig,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit benchmark"})
	}

	return c.Status(fiber.StatusCreated).JSON(submission)
}

// GetIndustryBaselines handles GET /api/public/federated-aggregation/baselines
func (h *FederatedAggregationHandler) GetIndustryBaselines(c *fiber.Ctx) error {
	baselines, err := h.federatedAggregationService.GetIndustryBaselines(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get baselines"})
	}
	return c.JSON(fiber.Map{"baselines": baselines})
}

// GetMeshStatus handles GET /api/public/federated-aggregation/mesh-status
func (h *FederatedAggregationHandler) GetMeshStatus(c *fiber.Ctx) error {
	status, err := h.federatedAggregationService.GetMeshStatus(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get mesh status"})
	}
	return c.JSON(status)
}

// GetFederatedAnalyticsDashboard handles GET /api/public/federated/dashboard
func (h *FederatedAggregationHandler) GetFederatedAnalyticsDashboard(c *fiber.Ctx) error {
	instanceIDStr := c.Query("instanceId")
	var instanceID uuid.UUID
	if instanceIDStr != "" {
		var err error
		instanceID, err = uuid.Parse(instanceIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid instance ID"})
		}
	} else {
		instanceID = uuid.New()
	}

	dashboard, err := h.federatedAggregationService.GetFederatedAnalyticsDashboard(c.Context(), instanceID)
	if err != nil {
		h.logger.Error("failed to get federated analytics dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}

// RunFederatedQuery handles POST /api/public/federated/query
func (h *FederatedAggregationHandler) RunFederatedQuery(c *fiber.Ctx) error {
	var input domain.FederatedQueryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(input.Metrics) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one metric is required"})
	}

	instanceID := uuid.New()
	results, err := h.federatedAggregationService.RunPrivacyPreservingQuery(c.Context(), instanceID, &input)
	if err != nil {
		h.logger.Error("failed to run federated query", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run query"})
	}

	return c.JSON(fiber.Map{"comparisons": results})
}
