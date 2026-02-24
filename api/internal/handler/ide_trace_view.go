package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// IDETraceViewHandler handles IDE inline trace viewer HTTP requests
type IDETraceViewHandler struct {
	logger              *zap.Logger
	ideTraceViewService *service.IDETraceViewService
}

// NewIDETraceViewHandler creates a new IDE trace view handler
func NewIDETraceViewHandler(
	ideTraceViewService *service.IDETraceViewService,
	logger *zap.Logger,
) *IDETraceViewHandler {
	return &IDETraceViewHandler{
		logger:              logger,
		ideTraceViewService: ideTraceViewService,
	}
}

// GetFileMapping returns trace mappings for a specific file
// @Summary Get file trace mapping
// @Description Get trace data mapped to specific lines in a source file for IDE inline display
// @Tags ide-trace-view
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param filePath query string true "File path to get trace mappings for"
// @Success 200 {object} domain.FileTraceMapping
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/ide-trace-view/file-mapping [get]
func (h *IDETraceViewHandler) GetFileMapping(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	filePath := c.Query("filePath")
	if filePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "filePath query parameter is required"})
	}

	mapping, err := h.ideTraceViewService.GetFileMapping(c.Context(), projectID, filePath)
	if err != nil {
		h.logger.Error("failed to get file trace mapping", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get file mapping"})
	}

	return c.JSON(mapping)
}

// GetTraceContext returns trace context details for IDE display
// @Summary Get trace context
// @Description Get detailed trace context for inline display in the IDE
// @Tags ide-trace-view
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Success 200 {object} domain.IDETraceContext
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/ide-trace-view/traces/{traceId} [get]
func (h *IDETraceViewHandler) GetTraceContext(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	traceContext, err := h.ideTraceViewService.GetTraceContext(c.Context(), traceID)
	if err != nil {
		h.logger.Error("failed to get trace context", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get trace context"})
	}

	return c.JSON(traceContext)
}

// GetBatchMappingsRequest represents the request to get batch file trace mappings
type GetBatchMappingsRequest struct {
	FilePaths []string `json:"filePaths"`
}

// GetBatchMappings returns trace mappings for multiple files
// @Summary Get batch file trace mappings
// @Description Get trace data mapped to specific lines across multiple source files for IDE inline display
// @Tags ide-trace-view
// @Accept json
// @Produce json
// @Param body body GetBatchMappingsRequest true "Batch mappings request"
// @Success 200 {array} domain.FileTraceMapping
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/ide-trace-view/batch-mappings [post]
func (h *IDETraceViewHandler) GetBatchMappings(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req GetBatchMappingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.FilePaths) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "filePaths array is required"})
	}

	mappings, err := h.ideTraceViewService.GetBatchMappings(c.Context(), projectID, req.FilePaths)
	if err != nil {
		h.logger.Error("failed to get batch file trace mappings", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get batch mappings"})
	}

	return c.JSON(mappings)
}
