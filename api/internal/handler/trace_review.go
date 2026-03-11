package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TraceReviewHandler handles trace review HTTP requests
type TraceReviewHandler struct {
	service *service.TraceReviewService
	logger  *zap.Logger
}

// NewTraceReviewHandler creates a new trace review handler
func NewTraceReviewHandler(svc *service.TraceReviewService, logger *zap.Logger) *TraceReviewHandler {
	return &TraceReviewHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateReview handles POST /api/public/reviews
func (h *TraceReviewHandler) CreateReview(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.TraceReviewInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}
	if input.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}

	userID := uuid.New()
	review, err := h.service.CreateReview(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to create review", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(review)
}

// GetReview handles GET /api/public/reviews/:reviewId
func (h *TraceReviewHandler) GetReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}

	review, err := h.service.GetReview(c.Context(), reviewID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Review not found"})
	}

	return c.JSON(review)
}

// ListReviews handles GET /api/public/reviews
func (h *TraceReviewHandler) ListReviews(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var status *domain.TraceReviewStatus
	if s := c.Query("status"); s != "" {
		st := domain.TraceReviewStatus(s)
		status = &st
	}

	reviews, err := h.service.ListReviews(c.Context(), projectID, status)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if reviews == nil {
		reviews = []domain.TraceReview{}
	}

	return c.JSON(fiber.Map{"reviews": reviews, "count": len(reviews)})
}

// UpdateReview handles PUT /api/public/reviews/:reviewId
func (h *TraceReviewHandler) UpdateReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}

	var input domain.TraceReviewInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	review, err := h.service.UpdateReview(c.Context(), reviewID, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(review)
}

// AddComment handles POST /api/public/reviews/:reviewId/comments
func (h *TraceReviewHandler) AddComment(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}

	var input domain.ReviewCommentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Comment content is required"})
	}

	userID := uuid.New()
	comment, err := h.service.AddComment(c.Context(), reviewID, userID, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// Approve handles POST /api/public/reviews/:reviewId/approve
func (h *TraceReviewHandler) Approve(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}

	userID := uuid.New()
	review, err := h.service.Approve(c.Context(), reviewID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(review)
}

// Reject handles POST /api/public/reviews/:reviewId/reject
func (h *TraceReviewHandler) Reject(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}

	var input struct {
		Reason string `json:"reason"`
	}
	_ = c.BodyParser(&input)

	userID := uuid.New()
	review, err := h.service.Reject(c.Context(), reviewID, userID, input.Reason)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(review)
}

// GetQueue handles GET /api/public/reviews/queue
func (h *TraceReviewHandler) GetQueue(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	queue, err := h.service.GetQueue(c.Context(), projectID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(queue)
}
