package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// ReplayPlanUseCase defines the transport boundary for safe replay planning.
type ReplayPlanUseCase interface {
	AssessCapabilities(
		ctx context.Context,
		projectID uuid.UUID,
		traceID string,
		input domain.ReplayPlanInput,
	) (*domain.ReplayCapabilityReport, error)
	CreatePlan(
		ctx context.Context,
		projectID uuid.UUID,
		traceID string,
		createdBy *uuid.UUID,
		input domain.ReplayPlanInput,
	) (*domain.ReplayPlan, error)
	GetPlan(ctx context.Context, projectID, planID uuid.UUID) (*domain.ReplayPlan, error)
	ExecutePlan(ctx context.Context, projectID, planID uuid.UUID) (*domain.ReplayPlan, error)
	RetryPlan(ctx context.Context, projectID, planID uuid.UUID) (*domain.ReplayPlan, error)
	GetComparison(
		ctx context.Context,
		projectID, planID uuid.UUID,
	) (*domain.ReplayPlanComparison, error)
}

// ReplayPlanHandler transports safe replay planning requests.
type ReplayPlanHandler struct {
	service ReplayPlanUseCase
	logger  *zap.Logger
}

// NewReplayPlanHandler creates a replay plan handler.
func NewReplayPlanHandler(
	replayPlanService ReplayPlanUseCase,
	logger *zap.Logger,
) *ReplayPlanHandler {
	return &ReplayPlanHandler{service: replayPlanService, logger: logger}
}

// AssessCapabilities handles GET /traces/:traceId/replay-capabilities.
func (h *ReplayPlanHandler) AssessCapabilities(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return replayPlanError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	traceID := c.Params("traceId")
	if traceID == "" {
		return replayPlanError(c, fiber.StatusBadRequest, "Trace ID is required")
	}

	input := domain.ReplayPlanInput{Mode: domain.ReplayModeRecordedGeneration}
	if checkpoint := c.Query("checkpointId"); checkpoint != "" {
		checkpointID, err := uuid.Parse(checkpoint)
		if err != nil {
			return replayPlanError(c, fiber.StatusBadRequest, "Invalid checkpoint ID")
		}
		input.CheckpointID = &checkpointID
	}
	if mode := c.Query("mode"); mode != "" {
		input.Mode = domain.ReplayExecutionMode(mode)
	}

	report, err := h.service.AssessCapabilities(c.Context(), projectID, traceID, input)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(report)
}

// CreatePlan handles POST /traces/:traceId/replay-plans.
func (h *ReplayPlanHandler) CreatePlan(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return replayPlanError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	traceID := c.Params("traceId")
	if traceID == "" {
		return replayPlanError(c, fiber.StatusBadRequest, "Trace ID is required")
	}

	var input domain.ReplayPlanInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return replayPlanError(c, fiber.StatusBadRequest, "Invalid request body")
		}
	}

	var createdBy *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		createdBy = &userID
	}
	plan, err := h.service.CreatePlan(c.Context(), projectID, traceID, createdBy, input)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(plan)
}

// GetPlan handles GET /replay-plans/:planId.
func (h *ReplayPlanHandler) GetPlan(c *fiber.Ctx) error {
	projectID, planID, err := replayPlanIDs(c)
	if err != nil {
		return err
	}
	plan, err := h.service.GetPlan(c.Context(), projectID, planID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(plan)
}

// ExecutePlan handles POST /replay-plans/:planId/execute.
func (h *ReplayPlanHandler) ExecutePlan(c *fiber.Ctx) error {
	projectID, planID, err := replayPlanIDs(c)
	if err != nil {
		return err
	}
	plan, err := h.service.ExecutePlan(c.Context(), projectID, planID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(plan)
}

// RetryPlan handles POST /replay-plans/:planId/retry.
func (h *ReplayPlanHandler) RetryPlan(c *fiber.Ctx) error {
	projectID, planID, err := replayPlanIDs(c)
	if err != nil {
		return err
	}
	plan, err := h.service.RetryPlan(c.Context(), projectID, planID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(plan)
}

// GetComparison handles GET /replay-plans/:planId/comparison.
func (h *ReplayPlanHandler) GetComparison(c *fiber.Ctx) error {
	projectID, planID, err := replayPlanIDs(c)
	if err != nil {
		return err
	}
	comparison, err := h.service.GetComparison(c.Context(), projectID, planID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(comparison)
}

func replayPlanIDs(
	c *fiber.Ctx,
) (projectID, planID uuid.UUID, resultErr error) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, replayPlanError(
			c,
			fiber.StatusUnauthorized,
			"Project ID not found",
		)
	}
	planID, err := uuid.Parse(c.Params("planId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, replayPlanError(
			c,
			fiber.StatusBadRequest,
			"Invalid replay plan ID",
		)
	}
	return projectID, planID, nil
}

func (h *ReplayPlanHandler) handleError(c *fiber.Ctx, err error) error {
	return roadmapAppError(
		c,
		h.logger,
		err,
		fiber.StatusInternalServerError,
		"Replay plan request failed",
	)
}

func replayPlanError(c *fiber.Ctx, status int, message string) error {
	return roadmapError(c, status, message)
}
