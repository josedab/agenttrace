package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PredictiveCostHandler handles predictive cost modeling HTTP requests
type PredictiveCostHandler struct {
	service *service.PredictiveCostService
	logger  *zap.Logger
}

// NewPredictiveCostHandler creates a new predictive cost handler
func NewPredictiveCostHandler(svc *service.PredictiveCostService, logger *zap.Logger) *PredictiveCostHandler {
	return &PredictiveCostHandler{
		service: svc,
		logger:  logger,
	}
}

// Predict handles POST /api/public/predictions/cost
func (h *PredictiveCostHandler) Predict(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.PredictionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TaskDescription == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "taskDescription is required"})
	}

	prediction, err := h.service.PredictCost(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to predict cost", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to predict cost"})
	}

	return c.Status(fiber.StatusCreated).JSON(prediction)
}

// ListPredictions handles GET /api/public/predictions
func (h *PredictiveCostHandler) ListPredictions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	predictions, err := h.service.ListPredictions(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list predictions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list predictions"})
	}

	return c.JSON(fiber.Map{"predictions": predictions, "count": len(predictions)})
}

// RequestApproval handles POST /api/public/predictions/:predictionId/approve
func (h *PredictiveCostHandler) RequestApproval(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	predictionID, err := uuid.Parse(c.Params("predictionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid prediction ID"})
	}

	approval, err := h.service.RequestApproval(c.Context(), projectID, predictionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.Status(fiber.StatusCreated).JSON(approval)
}

// DecideApproval handles POST /api/public/approvals/:approvalId/decide
func (h *PredictiveCostHandler) DecideApproval(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	approvalID, err := uuid.Parse(c.Params("approvalId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid approval ID"})
	}

	var input domain.ApprovalDecisionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Status != "approved" && input.Status != "rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be 'approved' or 'rejected'"})
	}

	approval, err := h.service.DecideApproval(c.Context(), approvalID, &input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(approval)
}
