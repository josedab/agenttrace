package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TestGeneratorHandler handles test generation HTTP requests
type TestGeneratorHandler struct {
	service *service.TestGeneratorService
	logger  *zap.Logger
}

// NewTestGeneratorHandler creates a new test generator handler
func NewTestGeneratorHandler(svc *service.TestGeneratorService, logger *zap.Logger) *TestGeneratorHandler {
	return &TestGeneratorHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateSuite handles POST /api/public/test-suites
func (h *TestGeneratorHandler) CreateSuite(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.TestSuiteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	userID := uuid.New() // placeholder
	suite, err := h.service.CreateSuite(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to create test suite", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(suite)
}

// ListSuites handles GET /api/public/test-suites
func (h *TestGeneratorHandler) ListSuites(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	suites, err := h.service.ListSuites(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if suites == nil {
		suites = []domain.TestSuite{}
	}

	return c.JSON(fiber.Map{"suites": suites, "count": len(suites)})
}

// GetSuite handles GET /api/public/test-suites/:suiteId
func (h *TestGeneratorHandler) GetSuite(c *fiber.Ctx) error {
	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	suite, err := h.service.GetSuite(c.Context(), suiteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Suite not found"})
	}

	return c.JSON(suite)
}

// DeleteSuite handles DELETE /api/public/test-suites/:suiteId
func (h *TestGeneratorHandler) DeleteSuite(c *fiber.Ctx) error {
	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	if err := h.service.DeleteSuite(c.Context(), suiteID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// GenerateFromTraces handles POST /api/public/test-suites/generate
func (h *TestGeneratorHandler) GenerateFromTraces(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.TestGenerateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(input.TraceIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one trace ID is required"})
	}

	suite, err := h.service.GenerateFromTraces(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to generate tests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(suite)
}

// ListTestCases handles GET /api/public/test-suites/:suiteId/cases
func (h *TestGeneratorHandler) ListTestCases(c *fiber.Ctx) error {
	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	cases, err := h.service.ListTestCases(c.Context(), suiteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if cases == nil {
		cases = []domain.TestCase{}
	}

	return c.JSON(fiber.Map{"cases": cases, "count": len(cases)})
}

// RunSuite handles POST /api/public/test-suites/:suiteId/run
func (h *TestGeneratorHandler) RunSuite(c *fiber.Ctx) error {
	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	result, err := h.service.RunSuite(c.Context(), suiteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// GetResults handles GET /api/public/test-suites/:suiteId/results
func (h *TestGeneratorHandler) GetResults(c *fiber.Ctx) error {
	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	results, err := h.service.GetResults(c.Context(), suiteID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if results == nil {
		results = []domain.TestRunResult{}
	}

	return c.JSON(fiber.Map{"results": results, "count": len(results)})
}

// CreateSnapshot handles POST /api/public/test-suites/:suiteId/snapshot
func (h *TestGeneratorHandler) CreateSnapshot(c *fiber.Ctx) error {
	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		input.Name = "snapshot"
	}

	snapshot, err := h.service.CreateSnapshot(c.Context(), suiteID, input.Name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(snapshot)
}
