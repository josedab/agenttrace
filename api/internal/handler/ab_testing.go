package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ABTestingHandler handles A/B testing HTTP requests
type ABTestingHandler struct {
	service *service.ABTestingService
	logger  *zap.Logger
}

// NewABTestingHandler creates a new A/B testing handler
func NewABTestingHandler(svc *service.ABTestingService, logger *zap.Logger) *ABTestingHandler {
	return &ABTestingHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateTest handles POST /api/public/ab-tests
func (h *ABTestingHandler) CreateTest(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.PromptABTestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	test, err := h.service.CreateTest(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create A/B test", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(test)
}

// GetTest handles GET /api/public/ab-tests/:testId
func (h *ABTestingHandler) GetTest(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	test, err := h.service.GetTest(c.Context(), testID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "A/B test not found"})
	}

	return c.JSON(test)
}

// ListTests handles GET /api/public/ab-tests
func (h *ABTestingHandler) ListTests(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	tests, err := h.service.ListTests(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list A/B tests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list tests"})
	}

	return c.JSON(fiber.Map{"tests": tests})
}

// StartTest handles POST /api/public/ab-tests/:testId/start
func (h *ABTestingHandler) StartTest(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	test, err := h.service.StartTest(c.Context(), testID)
	if err != nil {
		h.logger.Error("failed to start A/B test", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(test)
}

// PauseTest handles POST /api/public/ab-tests/:testId/pause
func (h *ABTestingHandler) PauseTest(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	test, err := h.service.PauseTest(c.Context(), testID)
	if err != nil {
		h.logger.Error("failed to pause A/B test", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(test)
}

// StopTest handles POST /api/public/ab-tests/:testId/stop
func (h *ABTestingHandler) StopTest(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	test, err := h.service.StopTest(c.Context(), testID)
	if err != nil {
		h.logger.Error("failed to stop A/B test", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(test)
}

// AssignVariant handles POST /api/public/ab-tests/:testId/assign
func (h *ABTestingHandler) AssignVariant(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	var body struct {
		AssignmentKey string `json:"assignmentKey"`
	}
	if err := c.BodyParser(&body); err != nil || body.AssignmentKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "assignmentKey is required"})
	}

	assignment, err := h.service.AssignVariant(c.Context(), testID, body.AssignmentKey)
	if err != nil {
		h.logger.Error("failed to assign variant", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(assignment)
}

// RecordResult handles POST /api/public/ab-tests/:testId/results
func (h *ABTestingHandler) RecordResult(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	var input domain.PromptABRecordResultInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.RecordResult(c.Context(), testID, &input); err != nil {
		h.logger.Error("failed to record result", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "recorded"})
}

// GetStatistics handles GET /api/public/ab-tests/:testId/statistics
func (h *ABTestingHandler) GetStatistics(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	stats, err := h.service.GetStatistics(c.Context(), testID)
	if err != nil {
		h.logger.Error("failed to get statistics", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(stats)
}

// SelectWinner handles POST /api/public/ab-tests/:testId/select-winner
func (h *ABTestingHandler) SelectWinner(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	var body struct {
		VariantID uuid.UUID `json:"variantId"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	test, err := h.service.SelectWinner(c.Context(), testID, body.VariantID)
	if err != nil {
		h.logger.Error("failed to select winner", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(test)
}

// StartGradualRollout handles POST /api/public/ab-tests/:testId/rollout
func (h *ABTestingHandler) StartGradualRollout(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid test ID"})
	}

	test, err := h.service.StartGradualRollout(c.Context(), testID)
	if err != nil {
		h.logger.Error("failed to start gradual rollout", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	return c.JSON(test)
}
