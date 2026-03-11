package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RunbookEngineHandler handles runbook engine HTTP requests
type RunbookEngineHandler struct {
	service *service.RunbookEngineService
	logger  *zap.Logger
}

// NewRunbookEngineHandler creates a new runbook engine handler
func NewRunbookEngineHandler(svc *service.RunbookEngineService, logger *zap.Logger) *RunbookEngineHandler {
	return &RunbookEngineHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateRunbook handles POST /api/public/runbook-engine/runbooks
func (h *RunbookEngineHandler) CreateRunbook(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.RunbookInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.YAMLContent == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "YAML content is required"})
	}

	userID := uuid.New()
	runbook, err := h.service.CreateRunbook(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to create runbook", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(runbook)
}

// GetRunbook handles GET /api/public/runbook-engine/runbooks/:runbookId
func (h *RunbookEngineHandler) GetRunbook(c *fiber.Ctx) error {
	runbookID, err := uuid.Parse(c.Params("runbookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid runbook ID"})
	}

	runbook, err := h.service.GetRunbook(c.Context(), runbookID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Runbook not found"})
	}

	return c.JSON(runbook)
}

// ListRunbooks handles GET /api/public/runbook-engine/runbooks
func (h *RunbookEngineHandler) ListRunbooks(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	runbooks, err := h.service.ListRunbooks(c.Context(), projectID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if runbooks == nil {
		runbooks = []domain.Runbook{}
	}

	return c.JSON(fiber.Map{"runbooks": runbooks, "count": len(runbooks)})
}

// UpdateRunbook handles PUT /api/public/runbook-engine/runbooks/:runbookId
func (h *RunbookEngineHandler) UpdateRunbook(c *fiber.Ctx) error {
	runbookID, err := uuid.Parse(c.Params("runbookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid runbook ID"})
	}

	var input domain.RunbookInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	runbook, err := h.service.UpdateRunbook(c.Context(), runbookID, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(runbook)
}

// Activate handles POST /api/public/runbook-engine/runbooks/:runbookId/activate
func (h *RunbookEngineHandler) Activate(c *fiber.Ctx) error {
	runbookID, err := uuid.Parse(c.Params("runbookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid runbook ID"})
	}

	runbook, err := h.service.Activate(c.Context(), runbookID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(runbook)
}

// TestRunbook handles POST /api/public/runbook-engine/runbooks/:runbookId/test
func (h *RunbookEngineHandler) TestRunbook(c *fiber.Ctx) error {
	runbookID, err := uuid.Parse(c.Params("runbookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid runbook ID"})
	}

	var input struct {
		TraceID string `json:"traceId"`
		DryRun  bool   `json:"dryRun"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	testInput := &domain.RunbookTestInput{
		RunbookID: runbookID,
		TraceID:   input.TraceID,
		DryRun:    input.DryRun,
	}

	result, err := h.service.TestRunbook(c.Context(), testInput)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ListExecutions handles GET /api/public/runbook-engine/executions
func (h *RunbookEngineHandler) ListExecutions(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	execs, err := h.service.ListExecutions(c.Context(), projectID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if execs == nil {
		execs = []domain.RunbookExecution{}
	}

	return c.JSON(fiber.Map{"executions": execs, "count": len(execs)})
}

// GetStats handles GET /api/public/runbook-engine/stats
func (h *RunbookEngineHandler) GetStats(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	stats := h.service.GetStats(c.Context(), projectID)
	return c.JSON(stats)
}
