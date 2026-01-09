package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// FileOperationsHandler handles file operation endpoints
type FileOperationsHandler struct {
	fileOpService *service.FileOperationService
	logger        *zap.Logger
}

// NewFileOperationsHandler creates a new file operations handler
func NewFileOperationsHandler(fileOpService *service.FileOperationService, logger *zap.Logger) *FileOperationsHandler {
	return &FileOperationsHandler{
		fileOpService: fileOpService,
		logger:        logger,
	}
}

// ListFileOperations handles GET /v1/file-operations
func (h *FileOperationsHandler) ListFileOperations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseFileOperationFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.fileOpService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list file operations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list file operations",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.FileOperations,
		"totalCount": list.TotalCount,
		"hasMore":    list.HasMore,
	})
}

// CreateFileOperation handles POST /v1/file-operations
func (h *FileOperationsHandler) CreateFileOperation(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.FileOperationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	if input.FilePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "filePath is required",
		})
	}

	if input.Operation == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "operation is required",
		})
	}

	fileOp, err := h.fileOpService.Track(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create file operation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create file operation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fileOp)
}

// BatchCreateFileOperations handles POST /v1/file-operations/batch
func (h *FileOperationsHandler) BatchCreateFileOperations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request struct {
		Operations []*domain.FileOperationInput `json:"operations"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if len(request.Operations) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "At least one operation required",
		})
	}

	if len(request.Operations) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Maximum 100 operations per batch",
		})
	}

	results, err := h.fileOpService.TrackBatch(c.Context(), projectID, request.Operations)
	if err != nil {
		h.logger.Error("failed to batch create file operations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create file operations",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": results,
	})
}

// GetTraceFileOperations handles GET /v1/traces/:traceId/file-operations
func (h *FileOperationsHandler) GetTraceFileOperations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Trace ID required",
		})
	}

	fileOps, err := h.fileOpService.GetByTraceID(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get trace file operations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace file operations",
		})
	}

	return c.JSON(fiber.Map{
		"data": fileOps,
	})
}

// GetFileOperationStats handles GET /v1/file-operations/stats
func (h *FileOperationsHandler) GetFileOperationStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var traceID *string
	if tid := c.Query("traceId"); tid != "" {
		traceID = &tid
	}

	stats, err := h.fileOpService.GetStats(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get file operation stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get file operation stats",
		})
	}

	return c.JSON(stats)
}

// parseFileOperationFilter parses file operation filter from query params
func (h *FileOperationsHandler) parseFileOperationFilter(c *fiber.Ctx) *domain.FileOperationFilter {
	filter := &domain.FileOperationFilter{}

	if traceID := c.Query("traceId"); traceID != "" {
		filter.TraceID = &traceID
	}

	if obsID := c.Query("observationId"); obsID != "" {
		filter.ObservationID = &obsID
	}

	if operation := c.Query("operation"); operation != "" {
		op := domain.FileOperationType(operation)
		filter.Operation = &op
	}

	if filePath := c.Query("filePath"); filePath != "" {
		filter.FilePath = &filePath
	}

	if success := c.Query("success"); success != "" {
		s := success == "true"
		filter.Success = &s
	}

	if from := c.Query("fromTimestamp"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.FromTime = &t
		}
	}

	if to := c.Query("toTimestamp"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.ToTime = &t
		}
	}

	return filter
}

// RegisterRoutes registers file operation routes
func (h *FileOperationsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/file-operations", h.ListFileOperations)
	v1.Get("/file-operations/stats", h.GetFileOperationStats)
	v1.Post("/file-operations", h.CreateFileOperation)
	v1.Post("/file-operations/batch", h.BatchCreateFileOperations)

	v1.Get("/traces/:traceId/file-operations", h.GetTraceFileOperations)
}
