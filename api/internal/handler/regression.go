package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RegressionHandler handles regression test HTTP requests
type RegressionHandler struct {
	regressionService *service.RegressionService
	logger            *zap.Logger
}

// NewRegressionHandler creates a new regression handler
func NewRegressionHandler(regressionService *service.RegressionService, logger *zap.Logger) *RegressionHandler {
	return &RegressionHandler{
		regressionService: regressionService,
		logger:            logger,
	}
}

// CreateTest handles POST /regression
func (h *RegressionHandler) CreateTest(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.RegressionTestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	test, err := h.regressionService.CreateTest(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create regression test", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create regression test",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(test)
}

// ListTests handles GET /regression
func (h *RegressionHandler) ListTests(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	tests, err := h.regressionService.ListTests(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list regression tests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list regression tests",
		})
	}

	return c.JSON(tests)
}

// RunTest handles POST /regression/:testId/run
func (h *RegressionHandler) RunTest(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	testID, err := uuid.Parse(c.Params("testId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid test ID",
		})
	}

	result, err := h.regressionService.RunTest(c.Context(), testID)
	if err != nil {
		h.logger.Error("failed to run regression test",
			zap.String("testId", testID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to run regression test",
		})
	}

	return c.JSON(result)
}

// GetResult handles GET /regression/:testId/results/:resultId
func (h *RegressionHandler) GetResult(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	resultID, err := uuid.Parse(c.Params("resultId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid result ID",
		})
	}

	result, err := h.regressionService.GetResult(c.Context(), resultID)
	if err != nil {
		h.logger.Error("failed to get regression result",
			zap.String("resultId", resultID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Regression result not found",
		})
	}

	return c.JSON(result)
}

// CheckGate handles POST /regression/gate
func (h *RegressionHandler) CheckGate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var body struct {
		TestIDs []uuid.UUID `json:"testIds"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if len(body.TestIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "testIds is required",
		})
	}

	gate, err := h.regressionService.CheckGate(c.Context(), projectID, body.TestIDs)
	if err != nil {
		h.logger.Error("failed to check regression gate", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to check regression gate",
		})
	}

	return c.JSON(gate)
}
