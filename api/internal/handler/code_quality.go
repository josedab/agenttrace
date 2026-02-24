package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CodeQualityHandler handles code quality HTTP requests
type CodeQualityHandler struct {
	logger             *zap.Logger
	codeQualityService *service.CodeQualityService
}

// NewCodeQualityHandler creates a new code quality handler
func NewCodeQualityHandler(
	logger *zap.Logger,
	codeQualityService *service.CodeQualityService,
) *CodeQualityHandler {
	return &CodeQualityHandler{
		logger:             logger,
		codeQualityService: codeQualityService,
	}
}

// CreateConfig creates a new code quality configuration
// @Summary Create code quality config
// @Description Create a new code quality analysis configuration for a project
// @Tags code-quality
// @Accept json
// @Produce json
// @Param config body domain.CodeQualityConfigInput true "Config configuration"
// @Success 201 {object} domain.CodeQualityConfig
// @Failure 400 {object} ErrorResponse
// @Router /api/public/code-quality/configs [post]
func (h *CodeQualityHandler) CreateConfig(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	var input domain.CodeQualityConfigInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if input.Name == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Name is required")
	}

	config, err := h.codeQualityService.CreateConfig(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create code quality config",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to create code quality config")
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

// GetConfig returns a specific code quality configuration
// @Summary Get code quality config
// @Description Get a specific code quality configuration by ID
// @Tags code-quality
// @Accept json
// @Produce json
// @Param configId path string true "Config ID"
// @Success 200 {object} domain.CodeQualityConfig
// @Failure 400 {object} ErrorResponse
// @Router /api/public/code-quality/configs/{configId} [get]
func (h *CodeQualityHandler) GetConfig(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	configID, err := uuid.Parse(c.Params("configId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid config ID")
	}

	config, err := h.codeQualityService.GetConfig(c.Context(), projectID, configID)
	if err != nil {
		h.logger.Error("failed to get code quality config",
			zap.String("projectId", projectID.String()),
			zap.String("configId", configID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get code quality config")
	}

	return c.JSON(config)
}

// AnalyzeTrace runs code quality analysis on a trace
// @Summary Analyze trace for code quality
// @Description Run code quality analysis on a specific trace
// @Tags code-quality
// @Accept json
// @Produce json
// @Param input body domain.CodeQualityInput true "Analysis input"
// @Success 201 {object} domain.CodeQualityReport
// @Failure 400 {object} ErrorResponse
// @Router /api/public/code-quality/analyze [post]
func (h *CodeQualityHandler) AnalyzeTrace(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	var input domain.CodeQualityInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	report, err := h.codeQualityService.AnalyzeTrace(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to analyze trace",
			zap.String("projectId", projectID.String()),
			zap.String("traceId", input.TraceID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to analyze trace")
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// ListReports returns a paginated list of code quality reports
// @Summary List code quality reports
// @Description Get a paginated list of code quality reports for a project
// @Tags code-quality
// @Accept json
// @Produce json
// @Param traceId query string false "Filter by trace ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.CodeQualityReportList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/code-quality/reports [get]
func (h *CodeQualityHandler) ListReports(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	var traceID *uuid.UUID
	if traceIDStr := c.Query("traceId"); traceIDStr != "" {
		parsed, err := uuid.Parse(traceIDStr)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid trace ID")
		}
		traceID = &parsed
	}

	pagination := ParsePagination(c, 100)

	reports, err := h.codeQualityService.ListReports(c.Context(), projectID, traceID, pagination.Limit, pagination.Offset)
	if err != nil {
		h.logger.Error("failed to list code quality reports",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to list code quality reports")
	}

	return c.JSON(reports)
}

// GetReport returns a specific code quality report
// @Summary Get code quality report
// @Description Get a specific code quality report by ID
// @Tags code-quality
// @Accept json
// @Produce json
// @Param reportId path string true "Report ID"
// @Success 200 {object} domain.CodeQualityReport
// @Failure 400 {object} ErrorResponse
// @Router /api/public/code-quality/reports/{reportId} [get]
func (h *CodeQualityHandler) GetReport(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	reportID, err := uuid.Parse(c.Params("reportId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid report ID")
	}

	report, err := h.codeQualityService.GetReport(c.Context(), projectID, reportID)
	if err != nil {
		h.logger.Error("failed to get code quality report",
			zap.String("projectId", projectID.String()),
			zap.String("reportId", reportID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get code quality report")
	}

	return c.JSON(report)
}

// GetDashboard returns aggregated code quality metrics
// @Summary Get code quality dashboard
// @Description Get aggregated code quality metrics for a project
// @Tags code-quality
// @Accept json
// @Produce json
// @Success 200 {object} domain.CodeQualityDashboard
// @Failure 400 {object} ErrorResponse
// @Router /api/public/code-quality/dashboard [get]
func (h *CodeQualityHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	dashboard, err := h.codeQualityService.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get code quality dashboard",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get code quality dashboard")
	}

	return c.JSON(dashboard)
}
