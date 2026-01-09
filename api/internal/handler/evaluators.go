package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// EvaluatorsHandler handles evaluator endpoints
type EvaluatorsHandler struct {
	evalService *service.EvalService
	logger      *zap.Logger
}

// NewEvaluatorsHandler creates a new evaluators handler
func NewEvaluatorsHandler(evalService *service.EvalService, logger *zap.Logger) *EvaluatorsHandler {
	return &EvaluatorsHandler{
		evalService: evalService,
		logger:      logger,
	}
}

// ListEvaluators handles GET /v1/evaluators
func (h *EvaluatorsHandler) ListEvaluators(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := &domain.EvaluatorFilter{
		ProjectID: projectID,
	}

	if evalType := c.Query("type"); evalType != "" {
		t := domain.EvaluatorType(evalType)
		filter.Type = &t
	}

	if enabledStr := c.Query("enabled"); enabledStr != "" {
		enabled := enabledStr == "true"
		filter.Enabled = &enabled
	}

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.evalService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list evaluators", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list evaluators",
		})
	}

	return c.JSON(list)
}

// GetEvaluator handles GET /v1/evaluators/:evaluatorId
func (h *EvaluatorsHandler) GetEvaluator(c *fiber.Ctx) error {
	evaluatorID, err := uuid.Parse(c.Params("evaluatorId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid evaluator ID",
		})
	}

	evaluator, err := h.evalService.Get(c.Context(), evaluatorID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Evaluator not found",
			})
		}
		h.logger.Error("failed to get evaluator", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get evaluator",
		})
	}

	return c.JSON(evaluator)
}

// CreateEvaluator handles POST /v1/evaluators
func (h *EvaluatorsHandler) CreateEvaluator(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.EvaluatorInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	if input.ScoreName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "scoreName is required",
		})
	}

	// Get user ID if available
	var userID uuid.UUID
	if uid, ok := middleware.GetUserID(c); ok {
		userID = uid
	}

	evaluator, err := h.evalService.Create(c.Context(), projectID, &input, userID)
	if err != nil {
		if apperrors.IsValidation(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}
		h.logger.Error("failed to create evaluator", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create evaluator",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(evaluator)
}

// UpdateEvaluator handles PATCH /v1/evaluators/:evaluatorId
func (h *EvaluatorsHandler) UpdateEvaluator(c *fiber.Ctx) error {
	evaluatorID, err := uuid.Parse(c.Params("evaluatorId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid evaluator ID",
		})
	}

	var input domain.EvaluatorUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	evaluator, err := h.evalService.Update(c.Context(), evaluatorID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Evaluator not found",
			})
		}
		if apperrors.IsValidation(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}
		h.logger.Error("failed to update evaluator", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update evaluator",
		})
	}

	return c.JSON(evaluator)
}

// DeleteEvaluator handles DELETE /v1/evaluators/:evaluatorId
func (h *EvaluatorsHandler) DeleteEvaluator(c *fiber.Ctx) error {
	evaluatorID, err := uuid.Parse(c.Params("evaluatorId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid evaluator ID",
		})
	}

	if err := h.evalService.Delete(c.Context(), evaluatorID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Evaluator not found",
			})
		}
		h.logger.Error("failed to delete evaluator", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete evaluator",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ExecuteEvaluator handles POST /v1/evaluators/:evaluatorId/execute
func (h *EvaluatorsHandler) ExecuteEvaluator(c *fiber.Ctx) error {
	evaluatorID, err := uuid.Parse(c.Params("evaluatorId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid evaluator ID",
		})
	}

	var input service.ExecuteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	result, err := h.evalService.Execute(c.Context(), evaluatorID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Evaluator not found",
			})
		}
		h.logger.Error("failed to execute evaluator", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to execute evaluator",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(result)
}

// GetJobStatus handles GET /v1/evaluators/jobs/:jobId
func (h *EvaluatorsHandler) GetJobStatus(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid job ID",
		})
	}

	job, err := h.evalService.GetJobStatus(c.Context(), jobID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Job not found",
			})
		}
		h.logger.Error("failed to get job status", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get job status",
		})
	}

	return c.JSON(job)
}

// ListTemplates handles GET /v1/evaluators/templates
func (h *EvaluatorsHandler) ListTemplates(c *fiber.Ctx) error {
	templates, err := h.evalService.ListTemplates(c.Context())
	if err != nil {
		h.logger.Error("failed to list templates", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list templates",
		})
	}

	return c.JSON(fiber.Map{
		"data": templates,
	})
}

// ListAnnotationQueues handles GET /v1/annotation-queues
func (h *EvaluatorsHandler) ListAnnotationQueues(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	queues, err := h.evalService.ListAnnotationQueues(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list annotation queues", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list annotation queues",
		})
	}

	return c.JSON(fiber.Map{
		"data": queues,
	})
}

// GetAnnotationQueue handles GET /v1/annotation-queues/:queueId
func (h *EvaluatorsHandler) GetAnnotationQueue(c *fiber.Ctx) error {
	queueID, err := uuid.Parse(c.Params("queueId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid queue ID",
		})
	}

	queue, err := h.evalService.GetAnnotationQueue(c.Context(), queueID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Annotation queue not found",
			})
		}
		h.logger.Error("failed to get annotation queue", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get annotation queue",
		})
	}

	return c.JSON(queue)
}

// CreateAnnotationQueue handles POST /v1/annotation-queues
func (h *EvaluatorsHandler) CreateAnnotationQueue(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		ScoreName   string         `json:"scoreName"`
		ScoreConfig map[string]any `json:"scoreConfig,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	if input.ScoreName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "scoreName is required",
		})
	}

	queue, err := h.evalService.CreateAnnotationQueue(c.Context(), projectID, input.Name, input.Description, input.ScoreName, input.ScoreConfig)
	if err != nil {
		h.logger.Error("failed to create annotation queue", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create annotation queue",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(queue)
}

// GetNextAnnotationItem handles GET /v1/annotation-queues/:queueId/next
func (h *EvaluatorsHandler) GetNextAnnotationItem(c *fiber.Ctx) error {
	queueID, err := uuid.Parse(c.Params("queueId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid queue ID",
		})
	}

	item, err := h.evalService.GetNextAnnotationItem(c.Context(), queueID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "No items available for annotation",
			})
		}
		h.logger.Error("failed to get next annotation item", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get next annotation item",
		})
	}

	return c.JSON(item)
}

// CompleteAnnotation handles POST /v1/annotation-queues/:queueId/items/:itemId/complete
func (h *EvaluatorsHandler) CompleteAnnotation(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	queueID, err := uuid.Parse(c.Params("queueId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid queue ID",
		})
	}

	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid item ID",
		})
	}

	var input struct {
		Value       *float64 `json:"value,omitempty"`
		StringValue *string  `json:"stringValue,omitempty"`
		Comment     *string  `json:"comment,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Value == nil && input.StringValue == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "value or stringValue is required",
		})
	}

	// Get user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}

	if err := h.evalService.CompleteAnnotation(c.Context(), projectID, queueID, itemID, userID, input.Value, input.StringValue, input.Comment); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Queue or item not found",
			})
		}
		h.logger.Error("failed to complete annotation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to complete annotation",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Annotation completed successfully",
	})
}

// RegisterRoutes registers evaluator routes
func (h *EvaluatorsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	// Evaluator endpoints
	v1.Get("/evaluators/templates", h.ListTemplates)
	v1.Get("/evaluators/jobs/:jobId", h.GetJobStatus)
	v1.Get("/evaluators", h.ListEvaluators)
	v1.Get("/evaluators/:evaluatorId", h.GetEvaluator)
	v1.Post("/evaluators", h.CreateEvaluator)
	v1.Post("/evaluators/:evaluatorId/execute", h.ExecuteEvaluator)
	v1.Patch("/evaluators/:evaluatorId", h.UpdateEvaluator)
	v1.Delete("/evaluators/:evaluatorId", h.DeleteEvaluator)

	// Annotation queue endpoints
	v1.Get("/annotation-queues", h.ListAnnotationQueues)
	v1.Get("/annotation-queues/:queueId", h.GetAnnotationQueue)
	v1.Post("/annotation-queues", h.CreateAnnotationQueue)
	v1.Get("/annotation-queues/:queueId/next", h.GetNextAnnotationItem)
	v1.Post("/annotation-queues/:queueId/items/:itemId/complete", h.CompleteAnnotation)
}
