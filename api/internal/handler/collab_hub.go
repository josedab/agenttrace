package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CollabHubHandler handles collaboration hub HTTP requests
type CollabHubHandler struct {
	logger  *zap.Logger
	service *service.CollabHubService
}

// NewCollabHubHandler creates a new collaboration hub handler
func NewCollabHubHandler(svc *service.CollabHubService, logger *zap.Logger) *CollabHubHandler {
	return &CollabHubHandler{
		logger:  logger,
		service: svc,
	}
}

// CollabHubCreateQueueRequest represents the request to create a review queue
type CollabHubCreateQueueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateReviewQueue handles POST /api/public/collab-hub/queues
// @Summary Create review queue
// @Description Create a new review queue
// @Tags collab-hub
// @Accept json
// @Produce json
// @Param body body CollabHubCreateQueueRequest true "Queue details"
// @Success 201 {object} domain.ReviewQueue
// @Failure 400 {object} map[string]string
// @Router /api/public/collab-hub/queues [post]
func (h *CollabHubHandler) CreateReviewQueue(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req CollabHubCreateQueueRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	queue, err := h.service.CreateReviewQueue(c.Context(), projectID, req.Name, req.Description)
	if err != nil {
		h.logger.Error("failed to create review queue", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create review queue"})
	}

	return c.Status(fiber.StatusCreated).JSON(queue)
}

// ListReviewQueues handles GET /api/public/collab-hub/queues
// @Summary List review queues
// @Description List all review queues for a project
// @Tags collab-hub
// @Accept json
// @Produce json
// @Success 200 {array} domain.ReviewQueue
// @Failure 401 {object} map[string]string
// @Router /api/public/collab-hub/queues [get]
func (h *CollabHubHandler) ListReviewQueues(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	queues, err := h.service.ListQueues(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list review queues", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list review queues"})
	}

	return c.JSON(queues)
}

// CollabHubAssignReviewRequest represents the request to assign a review
type CollabHubAssignReviewRequest struct {
	QueueID  string `json:"queueId"`
	TraceID  string `json:"traceId"`
	AssignTo string `json:"assignTo"`
}

// AssignReview handles POST /api/public/collab-hub/assignments
// @Summary Assign review
// @Description Assign a trace for review
// @Tags collab-hub
// @Accept json
// @Produce json
// @Param body body CollabHubAssignReviewRequest true "Assignment details"
// @Success 201 {object} domain.ReviewAssignment
// @Failure 400 {object} map[string]string
// @Router /api/public/collab-hub/assignments [post]
func (h *CollabHubHandler) AssignReview(c *fiber.Ctx) error {
	var req CollabHubAssignReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	queueID, err := uuid.Parse(req.QueueID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid queue ID"})
	}

	traceID, err := uuid.Parse(req.TraceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	assignTo, err := uuid.Parse(req.AssignTo)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignTo ID"})
	}

	assignment, err := h.service.AssignReview(c.Context(), queueID, traceID, assignTo)
	if err != nil {
		h.logger.Error("failed to assign review", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign review"})
	}

	return c.Status(fiber.StatusCreated).JSON(assignment)
}

// CollabHubCompleteReviewRequest represents the request to complete a review
type CollabHubCompleteReviewRequest struct {
	Status   string   `json:"status"`
	Feedback string   `json:"feedback"`
	Score    *float64 `json:"score,omitempty"`
}

// CompleteReview handles POST /api/public/collab-hub/assignments/:assignmentId/complete
// @Summary Complete review
// @Description Complete a review assignment
// @Tags collab-hub
// @Accept json
// @Produce json
// @Param assignmentId path string true "Assignment ID"
// @Param body body CollabHubCompleteReviewRequest true "Review completion details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/collab-hub/assignments/{assignmentId}/complete [post]
func (h *CollabHubHandler) CompleteReview(c *fiber.Ctx) error {
	assignmentID, err := uuid.Parse(c.Params("assignmentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID"})
	}

	var req CollabHubCompleteReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.CompleteReview(c.Context(), assignmentID, req.Status, req.Feedback, req.Score); err != nil {
		h.logger.Error("failed to complete review", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to complete review"})
	}

	return c.JSON(fiber.Map{"status": "completed"})
}

// CollabHubCreateStandardRequest represents the request to create a quality standard
type CollabHubCreateStandardRequest struct {
	Name  string               `json:"name"`
	Rules []domain.QualityRule `json:"rules"`
}

// CreateQualityStandard handles POST /api/public/collab-hub/standards
// @Summary Create quality standard
// @Description Create a new quality standard
// @Tags collab-hub
// @Accept json
// @Produce json
// @Param body body CollabHubCreateStandardRequest true "Quality standard details"
// @Success 201 {object} domain.QualityStandard
// @Failure 400 {object} map[string]string
// @Router /api/public/collab-hub/standards [post]
func (h *CollabHubHandler) CreateQualityStandard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req CollabHubCreateStandardRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	standard, err := h.service.CreateQualityStandard(c.Context(), projectID, req.Name, req.Rules)
	if err != nil {
		h.logger.Error("failed to create quality standard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create quality standard"})
	}

	return c.Status(fiber.StatusCreated).JSON(standard)
}

// ListQualityStandards handles GET /api/public/collab-hub/standards
// @Summary List quality standards
// @Description List all quality standards for a project
// @Tags collab-hub
// @Accept json
// @Produce json
// @Success 200 {array} domain.QualityStandard
// @Failure 401 {object} map[string]string
// @Router /api/public/collab-hub/standards [get]
func (h *CollabHubHandler) ListQualityStandards(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	standards, err := h.service.ListStandards(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list quality standards", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list quality standards"})
	}

	return c.JSON(standards)
}

// GetActivityFeed handles GET /api/public/collab-hub/activity
// @Summary Get activity feed
// @Description Get the activity feed for a project
// @Tags collab-hub
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(50)
// @Success 200 {array} domain.ActivityFeedItem
// @Failure 401 {object} map[string]string
// @Router /api/public/collab-hub/activity [get]
func (h *CollabHubHandler) GetActivityFeed(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	limit := c.QueryInt("limit", 50)

	feed, err := h.service.GetActivityFeed(c.Context(), projectID, limit)
	if err != nil {
		h.logger.Error("failed to get activity feed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get activity feed"})
	}

	return c.JSON(feed)
}
