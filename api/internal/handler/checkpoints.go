package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CheckpointsHandler handles checkpoint endpoints
type CheckpointsHandler struct {
	checkpointService *service.CheckpointService
	logger            *zap.Logger
}

// NewCheckpointsHandler creates a new checkpoints handler
func NewCheckpointsHandler(checkpointService *service.CheckpointService, logger *zap.Logger) *CheckpointsHandler {
	return &CheckpointsHandler{
		checkpointService: checkpointService,
		logger:            logger,
	}
}

// ListCheckpoints handles GET /v1/checkpoints
func (h *CheckpointsHandler) ListCheckpoints(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseCheckpointFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.checkpointService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list checkpoints", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list checkpoints",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.Checkpoints,
		"totalCount": list.TotalCount,
		"hasMore":    list.HasMore,
	})
}

// GetCheckpoint handles GET /v1/checkpoints/:checkpointId
func (h *CheckpointsHandler) GetCheckpoint(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	checkpointIDStr := c.Params("checkpointId")
	if checkpointIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Checkpoint ID required",
		})
	}

	checkpointID, err := uuid.Parse(checkpointIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid checkpoint ID format",
		})
	}

	checkpoint, err := h.checkpointService.Get(c.Context(), projectID, checkpointID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Checkpoint not found",
			})
		}
		h.logger.Error("failed to get checkpoint", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get checkpoint",
		})
	}

	return c.JSON(checkpoint)
}

// CreateCheckpoint handles POST /v1/checkpoints
func (h *CheckpointsHandler) CreateCheckpoint(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.CheckpointInput
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

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	checkpoint, err := h.checkpointService.Create(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create checkpoint", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create checkpoint",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(checkpoint)
}

// RestoreCheckpoint handles POST /v1/checkpoints/:checkpointId/restore
func (h *CheckpointsHandler) RestoreCheckpoint(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	checkpointIDStr := c.Params("checkpointId")
	if checkpointIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Checkpoint ID required",
		})
	}

	checkpointID, err := uuid.Parse(checkpointIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid checkpoint ID format",
		})
	}

	var request struct {
		TraceID string `json:"traceId"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	input := &domain.RestoreCheckpointInput{
		CheckpointID: checkpointID,
		TraceID:      request.TraceID,
	}

	rollback, err := h.checkpointService.Restore(c.Context(), projectID, input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Checkpoint not found",
			})
		}
		h.logger.Error("failed to restore checkpoint", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to restore checkpoint",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(rollback)
}

// GetTraceCheckpoints handles GET /v1/traces/:traceId/checkpoints
func (h *CheckpointsHandler) GetTraceCheckpoints(c *fiber.Ctx) error {
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

	checkpoints, err := h.checkpointService.GetByTraceID(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get trace checkpoints", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace checkpoints",
		})
	}

	return c.JSON(fiber.Map{
		"data": checkpoints,
	})
}

// parseCheckpointFilter parses checkpoint filter from query params
func (h *CheckpointsHandler) parseCheckpointFilter(c *fiber.Ctx) *domain.CheckpointFilter {
	filter := &domain.CheckpointFilter{}

	if traceID := c.Query("traceId"); traceID != "" {
		filter.TraceID = &traceID
	}

	if checkpointType := c.Query("type"); checkpointType != "" {
		ct := domain.CheckpointType(checkpointType)
		filter.Type = &ct
	}

	if gitBranch := c.Query("gitBranch"); gitBranch != "" {
		filter.GitBranch = &gitBranch
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

// RegisterRoutes registers checkpoint routes
func (h *CheckpointsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/checkpoints", h.ListCheckpoints)
	v1.Get("/checkpoints/:checkpointId", h.GetCheckpoint)
	v1.Post("/checkpoints", h.CreateCheckpoint)
	v1.Post("/checkpoints/:checkpointId/restore", h.RestoreCheckpoint)

	v1.Get("/traces/:traceId/checkpoints", h.GetTraceCheckpoints)
}
