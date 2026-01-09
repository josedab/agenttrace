package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ScoresHandler handles score endpoints
type ScoresHandler struct {
	scoreService *service.ScoreService
	logger       *zap.Logger
}

// NewScoresHandler creates a new scores handler
func NewScoresHandler(scoreService *service.ScoreService, logger *zap.Logger) *ScoresHandler {
	return &ScoresHandler{
		scoreService: scoreService,
		logger:       logger,
	}
}

// ListScores handles GET /v1/scores
func (h *ScoresHandler) ListScores(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseScoreFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.scoreService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list scores", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list scores",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.Scores,
		"totalCount": list.TotalCount,
	})
}

// GetScore handles GET /v1/scores/:scoreId
func (h *ScoresHandler) GetScore(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	scoreID := c.Params("scoreId")
	if scoreID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Score ID required",
		})
	}

	score, err := h.scoreService.Get(c.Context(), projectID, scoreID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Score not found",
			})
		}
		h.logger.Error("failed to get score", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get score",
		})
	}

	return c.JSON(score)
}

// CreateScore handles POST /v1/scores
func (h *ScoresHandler) CreateScore(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ScoreInput
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

	if input.Value == nil && input.StringValue == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "value or stringValue is required",
		})
	}

	score, err := h.scoreService.Create(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create score", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create score",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(score)
}

// UpdateScore handles PATCH /v1/scores/:scoreId
func (h *ScoresHandler) UpdateScore(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	scoreID := c.Params("scoreId")
	if scoreID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Score ID required",
		})
	}

	var input domain.ScoreInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	score, err := h.scoreService.Update(c.Context(), projectID, scoreID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Score not found",
			})
		}
		h.logger.Error("failed to update score", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update score",
		})
	}

	return c.JSON(score)
}


// DeleteScore handles DELETE /v1/scores/:scoreId
func (h *ScoresHandler) DeleteScore(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	scoreID := c.Params("scoreId")
	if scoreID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Score ID required",
		})
	}

	if err := h.scoreService.Delete(c.Context(), projectID, scoreID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Score not found",
			})
		}
		h.logger.Error("failed to delete score", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete score",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetTraceScores handles GET /v1/traces/:traceId/scores
func (h *ScoresHandler) GetTraceScores(c *fiber.Ctx) error {
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

	scores, err := h.scoreService.GetByTraceID(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get trace scores", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace scores",
		})
	}

	return c.JSON(fiber.Map{
		"data": scores,
	})
}

// BatchCreateScores handles POST /v1/scores/batch
func (h *ScoresHandler) BatchCreateScores(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request struct {
		Scores []*domain.ScoreInput `json:"scores"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if len(request.Scores) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "At least one score required",
		})
	}

	if len(request.Scores) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Maximum 100 scores per batch",
		})
	}

	results, err := h.scoreService.CreateBatch(c.Context(), projectID, request.Scores)
	if err != nil {
		h.logger.Error("failed to batch create scores", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create scores",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": results,
	})
}

// SubmitFeedback handles POST /v1/feedback
func (h *ScoresHandler) SubmitFeedback(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request struct {
		TraceID  string                 `json:"traceId"`
		Name     string                 `json:"name"`
		Value    *float64               `json:"value"`
		DataType domain.ScoreDataType   `json:"dataType"`
		Comment  *string                `json:"comment"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if request.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	feedback := &service.FeedbackInput{
		Name:     request.Name,
		Value:    request.Value,
		DataType: request.DataType,
		Comment:  request.Comment,
	}

	score, err := h.scoreService.SubmitFeedback(c.Context(), projectID, request.TraceID, feedback)
	if err != nil {
		h.logger.Error("failed to submit feedback", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to submit feedback",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(score)
}

// GetScoreStats handles GET /v1/scores/stats
func (h *ScoresHandler) GetScoreStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	scoreName := c.Query("name")
	if scoreName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Score name required",
		})
	}

	stats, err := h.scoreService.GetStats(c.Context(), projectID, scoreName)
	if err != nil {
		h.logger.Error("failed to get score stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get score stats",
		})
	}

	return c.JSON(stats)
}

// parseScoreFilter parses score filter from query params
func (h *ScoresHandler) parseScoreFilter(c *fiber.Ctx) *domain.ScoreFilter {
	filter := &domain.ScoreFilter{}

	if traceID := c.Query("traceId"); traceID != "" {
		filter.TraceID = &traceID
	}

	if obsID := c.Query("observationId"); obsID != "" {
		filter.ObservationID = &obsID
	}

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	if source := c.Query("source"); source != "" {
		s := domain.ScoreSource(source)
		filter.Source = &s
	}

	if dataType := c.Query("dataType"); dataType != "" {
		dt := domain.ScoreDataType(dataType)
		filter.DataType = &dt
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

// RegisterRoutes registers score routes
func (h *ScoresHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/scores", h.ListScores)
	v1.Get("/scores/stats", h.GetScoreStats)
	v1.Get("/scores/:scoreId", h.GetScore)
	v1.Post("/scores", h.CreateScore)
	v1.Post("/scores/batch", h.BatchCreateScores)
	v1.Patch("/scores/:scoreId", h.UpdateScore)
	v1.Delete("/scores/:scoreId", h.DeleteScore)

	v1.Get("/traces/:traceId/scores", h.GetTraceScores)
	v1.Post("/feedback", h.SubmitFeedback)
}
