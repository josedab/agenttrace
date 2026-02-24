package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type SandboxHandler struct {
	sandboxService *service.SandboxService
	logger         *zap.Logger
}

func NewSandboxHandler(svc *service.SandboxService, logger *zap.Logger) *SandboxHandler {
	return &SandboxHandler{sandboxService: svc, logger: logger}
}

func (h *SandboxHandler) SubmitReview(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	var input domain.SandboxReviewInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	review, err := h.sandboxService.SubmitForReview(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(review)
}

func (h *SandboxHandler) GetReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}
	review, err := h.sandboxService.GetReview(c.Context(), reviewID)
	if err != nil || review == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Review not found"})
	}
	return c.JSON(review)
}

func (h *SandboxHandler) ListPending(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	reviews, err := h.sandboxService.ListPendingReviews(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"reviews": reviews, "count": len(reviews)})
}

func (h *SandboxHandler) Decide(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("reviewId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid review ID"})
	}
	var decision domain.SandboxDecision
	if err := c.BodyParser(&decision); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	review, err := h.sandboxService.ReviewDecision(c.Context(), reviewID, &decision)
	if err != nil || review == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Review not found"})
	}
	return c.JSON(review)
}

func (h *SandboxHandler) CreatePolicy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	var input domain.SandboxPolicyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	policy, err := h.sandboxService.CreatePolicy(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(policy)
}

func (h *SandboxHandler) ListPolicies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	policies, err := h.sandboxService.ListPolicies(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"policies": policies})
}

func (h *SandboxHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	stats, err := h.sandboxService.GetStats(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}
