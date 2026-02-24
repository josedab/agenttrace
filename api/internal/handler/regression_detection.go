package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RegressionDetectionHandler handles regression detection HTTP requests
type RegressionDetectionHandler struct {
	logger  *zap.Logger
	service *service.RegressionDetectionService
}

// NewRegressionDetectionHandler creates a new regression detection handler
func NewRegressionDetectionHandler(
	service *service.RegressionDetectionService,
	logger *zap.Logger,
) *RegressionDetectionHandler {
	return &RegressionDetectionHandler{
		logger:  logger,
		service: service,
	}
}

// CreateConfig creates a new regression detection configuration
// @Summary Create regression detection config
// @Description Create a new regression detection configuration
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param config body domain.RegressionDetectionInput true "Config configuration"
// @Success 201 {object} domain.RegressionDetectionConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/configs [post]
func (h *RegressionDetectionHandler) CreateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.RegressionDetectionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.CreateConfig(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create regression detection config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create config"})
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

// ListConfigs lists regression detection configurations
// @Summary List regression detection configs
// @Description Get all regression detection configurations for a project
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.RegressionDetectionConfigList
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/configs [get]
func (h *RegressionDetectionHandler) ListConfigs(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	configs, err := h.service.ListConfigs(c.Context(), projectID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list regression detection configs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list configs"})
	}

	return c.JSON(configs)
}

// GetConfig returns a specific regression detection configuration
// @Summary Get regression detection config
// @Description Get a specific regression detection configuration by ID
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param configId path string true "Config ID"
// @Success 200 {object} domain.RegressionDetectionConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/configs/{configId} [get]
func (h *RegressionDetectionHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID, configID)
	if err != nil {
		h.logger.Error("failed to get regression detection config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get config"})
	}

	return c.JSON(config)
}

// UpdateConfig updates an existing regression detection configuration
// @Summary Update regression detection config
// @Description Update an existing regression detection configuration
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param configId path string true "Config ID"
// @Param config body domain.RegressionDetectionInput true "Updated config"
// @Success 200 {object} domain.RegressionDetectionConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/configs/{configId} [put]
func (h *RegressionDetectionHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	var input domain.RegressionDetectionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID, configID, input)
	if err != nil {
		h.logger.Error("failed to update regression detection config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update config"})
	}

	return c.JSON(config)
}

// DeleteConfig deletes a regression detection configuration
// @Summary Delete regression detection config
// @Description Delete a regression detection configuration by ID
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param configId path string true "Config ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/configs/{configId} [delete]
func (h *RegressionDetectionHandler) DeleteConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	if err := h.service.DeleteConfig(c.Context(), projectID, configID); err != nil {
		h.logger.Error("failed to delete regression detection config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete config"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RunDetection executes a regression detection analysis for a config
// @Summary Run regression detection
// @Description Execute a regression detection analysis for the given configuration
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param configId path string true "Config ID"
// @Success 201 {object} domain.RegressionDetectionResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/configs/{configId}/run [post]
func (h *RegressionDetectionHandler) RunDetection(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config ID"})
	}

	result, err := h.service.RunDetection(c.Context(), projectID, configID)
	if err != nil {
		h.logger.Error("failed to run regression detection", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run detection"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ListDetections lists recent regression detection results
// @Summary List regression detections
// @Description Get regression detection results with optional severity and status filters
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param severity query string false "Filter by severity"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.RegressionDetectionResultList
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/detections [get]
func (h *RegressionDetectionHandler) ListDetections(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	severity := c.Query("severity")
	status := c.Query("status")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	results, err := h.service.ListDetections(c.Context(), projectID, severity, status, limit, offset)
	if err != nil {
		h.logger.Error("failed to list regression detections", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list detections"})
	}

	return c.JSON(results)
}

// AcknowledgeDetection marks a regression detection as acknowledged
// @Summary Acknowledge regression detection
// @Description Mark a regression detection result as acknowledged by the current user
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param detectionId path string true "Detection ID"
// @Success 200 {object} domain.RegressionDetectionResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/detections/{detectionId}/acknowledge [post]
func (h *RegressionDetectionHandler) AcknowledgeDetection(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	detectionID, err := uuid.Parse(c.Params("detectionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid detection ID"})
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found"})
	}

	result, err := h.service.AcknowledgeDetection(c.Context(), projectID, detectionID, userID)
	if err != nil {
		h.logger.Error("failed to acknowledge regression detection", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to acknowledge detection"})
	}

	return c.JSON(result)
}

// ResolveDetectionRequest represents the request body for resolving a detection
type ResolveDetectionRequest struct {
	FalsePositive bool `json:"falsePositive"`
}

// ResolveDetection marks a regression detection as resolved
// @Summary Resolve regression detection
// @Description Mark a regression detection result as resolved, optionally flagging as false positive
// @Tags regression-detection
// @Accept json
// @Produce json
// @Param detectionId path string true "Detection ID"
// @Param body body ResolveDetectionRequest true "Resolution details"
// @Success 200 {object} domain.RegressionDetectionResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/detections/{detectionId}/resolve [post]
func (h *RegressionDetectionHandler) ResolveDetection(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	detectionID, err := uuid.Parse(c.Params("detectionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid detection ID"})
	}

	var req ResolveDetectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.ResolveDetection(c.Context(), projectID, detectionID, req.FalsePositive)
	if err != nil {
		h.logger.Error("failed to resolve regression detection", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to resolve detection"})
	}

	return c.JSON(result)
}

// GetDashboard returns the regression detection dashboard overview
// @Summary Get regression detection dashboard
// @Description Get an overview dashboard for regression detection including metric health and recent detections
// @Tags regression-detection
// @Accept json
// @Produce json
// @Success 200 {object} domain.RegressionDetectionDashboard
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/regression-detection/dashboard [get]
func (h *RegressionDetectionHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get regression detection dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}
