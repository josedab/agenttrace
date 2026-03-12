package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TraceComparisonHandler handles trace comparison HTTP requests
type TraceComparisonHandler struct {
	logger     *zap.Logger
	compareSvc *service.TraceComparisonService
}

// NewTraceComparisonHandler creates a new trace comparison handler
func NewTraceComparisonHandler(logger *zap.Logger, compareSvc *service.TraceComparisonService) *TraceComparisonHandler {
	return &TraceComparisonHandler{
		logger:     logger,
		compareSvc: compareSvc,
	}
}

// CompareTraces handles POST /api/v1/traces/compare
// @Summary Compare multiple traces
// @Description Create a comparison matrix for 2-N traces
// @Tags traces
// @Accept json
// @Produce json
// @Param body body domain.TraceComparisonInput true "Trace IDs to compare"
// @Success 200 {object} domain.TraceComparisonMatrix
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/traces/compare [post]
func (h *TraceComparisonHandler) CompareTraces(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	userID, err := RequireUserID(c)
	if err != nil {
		return err
	}

	var input domain.TraceComparisonInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if len(input.TraceIDs) < 2 {
		return errorResponse(c, fiber.StatusBadRequest, "At least 2 trace IDs are required")
	}

	if len(input.TraceIDs) > 10 {
		return errorResponse(c, fiber.StatusBadRequest, "Maximum 10 traces can be compared")
	}

	matrix, err := h.compareSvc.CompareTraces(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to compare traces", zap.Error(err))
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to compare traces")
	}

	return c.JSON(matrix)
}

// GetComparison handles GET /api/v1/traces/compare/:id
// @Summary Get a saved comparison
// @Description Retrieve a previously saved trace comparison
// @Tags traces
// @Produce json
// @Param id path string true "Comparison ID"
// @Success 200 {object} domain.TraceComparisonMatrix
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/traces/compare/{id} [get]
func (h *TraceComparisonHandler) GetComparison(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid comparison ID")
	}

	matrix, err := h.compareSvc.GetComparison(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "Comparison not found")
	}

	return c.JSON(matrix)
}

// GetSharedComparison handles GET /api/v1/traces/compare/shared/:token
// @Summary Get a shared comparison
// @Description Retrieve a comparison by its share token
// @Tags traces
// @Produce json
// @Param token path string true "Share token"
// @Success 200 {object} domain.TraceComparisonMatrix
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/traces/compare/shared/{token} [get]
func (h *TraceComparisonHandler) GetSharedComparison(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Share token is required")
	}

	matrix, err := h.compareSvc.GetComparisonByShareToken(c.Context(), token)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "Comparison not found")
	}

	return c.JSON(matrix)
}

// ExportComparison handles GET /api/v1/traces/compare/:id/export
// @Summary Export a comparison
// @Description Export a trace comparison as CSV or JSON
// @Tags traces
// @Produce json
// @Param id path string true "Comparison ID"
// @Param format query string true "Export format (csv, json)"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/traces/compare/{id}/export [get]
func (h *TraceComparisonHandler) ExportComparison(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid comparison ID")
	}

	format := c.Query("format", "csv")
	if format != "csv" && format != "json" {
		return errorResponse(c, fiber.StatusBadRequest, "Format must be 'csv' or 'json'")
	}

	export, err := h.compareSvc.ExportComparison(c.Context(), id, format)
	if err != nil {
		h.logger.Error("failed to export comparison", zap.Error(err))
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to export comparison")
	}

	contentType := "text/csv"
	if format == "json" {
		contentType = "application/json"
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename="+export.Filename)
	return c.Send(export.Data)
}
