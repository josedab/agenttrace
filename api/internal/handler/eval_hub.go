package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// EvalHubUseCase defines the transport boundary for Eval Hub workflows.
type EvalHubUseCase interface {
	ListPackages(
		ctx context.Context,
		projectID uuid.UUID,
		filter domain.EvalHubPackageFilter,
	) (*domain.EvalHubPackageList, error)
	GetPackage(
		ctx context.Context,
		projectID, packageID uuid.UUID,
	) (*domain.EvalHubPackage, error)
	Publish(
		ctx context.Context,
		projectID, userID uuid.UUID,
		input domain.EvalHubPublishInput,
	) (*domain.EvalHubPackage, error)
	Fork(
		ctx context.Context,
		projectID, userID, packageID uuid.UUID,
		input domain.EvalHubForkInput,
	) (*domain.EvalHubPackage, error)
	Run(
		ctx context.Context,
		projectID, userID, packageID uuid.UUID,
		input domain.EvalHubRunInput,
	) (*domain.EvalHubRun, error)
	GetRun(ctx context.Context, projectID, runID uuid.UUID) (*domain.EvalHubRun, error)
	ListRuns(
		ctx context.Context,
		projectID uuid.UUID,
		limit, offset int,
	) (*domain.EvalHubRunList, error)
}

// EvalHubHandler transports canonical Eval Hub requests.
type EvalHubHandler struct {
	service EvalHubUseCase
	logger  *zap.Logger
}

// NewEvalHubHandler creates an Eval Hub handler.
func NewEvalHubHandler(evalHubService EvalHubUseCase, logger *zap.Logger) *EvalHubHandler {
	return &EvalHubHandler{service: evalHubService, logger: logger}
}

// ListPackages handles GET /eval-hub/packages.
func (h *EvalHubHandler) ListPackages(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Project ID not found")
	}

	filter := domain.EvalHubPackageFilter{
		Query:  c.Query("q"),
		Limit:  c.QueryInt("limit", 50),
		Offset: c.QueryInt("offset", 0),
	}
	if value := c.Query("kind"); value != "" {
		kind := domain.EvalHubAssetKind(value)
		if !kind.IsValid() {
			return evalHubError(c, fiber.StatusBadRequest, "Invalid asset kind")
		}
		filter.Kind = &kind
	}
	if value := c.Query("visibility"); value != "" {
		visibility := domain.EvalHubVisibility(value)
		if !visibility.IsValid() {
			return evalHubError(c, fiber.StatusBadRequest, "Invalid visibility")
		}
		filter.Visibility = &visibility
	}

	result, err := h.service.ListPackages(c.Context(), projectID, filter)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(result)
}

// GetPackage handles GET /eval-hub/packages/:packageId.
func (h *EvalHubHandler) GetPackage(c *fiber.Ctx) error {
	projectID, packageID, err := evalHubIDs(c, "packageId")
	if err != nil {
		return err
	}
	pkg, err := h.service.GetPackage(c.Context(), projectID, packageID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(pkg)
}

// Publish handles POST /eval-hub/packages.
func (h *EvalHubHandler) Publish(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	userID, ok := evalHubActorID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}

	var input domain.EvalHubPublishInput
	if err := c.BodyParser(&input); err != nil {
		return evalHubError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	pkg, err := h.service.Publish(c.Context(), projectID, userID, input)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(pkg)
}

// Fork handles POST /eval-hub/packages/:packageId/fork.
func (h *EvalHubHandler) Fork(c *fiber.Ctx) error {
	projectID, packageID, err := evalHubIDs(c, "packageId")
	if err != nil {
		return err
	}
	userID, ok := evalHubActorID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}

	var input domain.EvalHubForkInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return evalHubError(c, fiber.StatusBadRequest, "Invalid request body")
		}
	}
	pkg, err := h.service.Fork(c.Context(), projectID, userID, packageID, input)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(pkg)
}

// Run handles POST /eval-hub/packages/:packageId/runs.
func (h *EvalHubHandler) Run(c *fiber.Ctx) error {
	projectID, packageID, err := evalHubIDs(c, "packageId")
	if err != nil {
		return err
	}
	userID, ok := evalHubActorID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}

	var input domain.EvalHubRunInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return evalHubError(c, fiber.StatusBadRequest, "Invalid request body")
		}
	}
	run, err := h.service.Run(c.Context(), projectID, userID, packageID, input)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(run)
}

// GetRun handles GET /eval-hub/runs/:runId.
func (h *EvalHubHandler) GetRun(c *fiber.Ctx) error {
	projectID, runID, err := evalHubIDs(c, "runId")
	if err != nil {
		return err
	}
	run, err := h.service.GetRun(c.Context(), projectID, runID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(run)
}

// ListRuns handles GET /eval-hub/runs.
func (h *EvalHubHandler) ListRuns(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	runs, err := h.service.ListRuns(
		c.Context(),
		projectID,
		c.QueryInt("limit", 50),
		c.QueryInt("offset", 0),
	)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(runs)
}

// ListLegacyDatasets preserves the old marketplace list route as a dataset-package alias.
func (h *EvalHubHandler) ListLegacyDatasets(c *fiber.Ctx) error {
	c.Request().URI().QueryArgs().Set("kind", string(domain.EvalHubDataset))
	return h.ListPackages(c)
}

// GetLegacyDataset preserves the old marketplace detail route.
func (h *EvalHubHandler) GetLegacyDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	packageID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return evalHubError(c, fiber.StatusBadRequest, "Invalid datasetId")
	}
	pkg, err := h.service.GetPackage(c.Context(), projectID, packageID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(pkg)
}

// ImportLegacyDataset preserves the old import route as a fork operation.
func (h *EvalHubHandler) ImportLegacyDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	packageID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return evalHubError(c, fiber.StatusBadRequest, "Invalid datasetId")
	}
	userID, ok := evalHubActorID(c)
	if !ok {
		return evalHubError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}
	pkg, err := h.service.Fork(
		c.Context(),
		projectID,
		userID,
		packageID,
		domain.EvalHubForkInput{},
	)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(pkg)
}

func evalHubIDs(
	c *fiber.Ctx,
	parameter string,
) (projectID, resourceID uuid.UUID, resultErr error) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, evalHubError(
			c,
			fiber.StatusUnauthorized,
			"Project ID not found",
		)
	}
	rawID := c.Params(parameter)
	resourceID, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil, uuid.Nil, evalHubError(
			c,
			fiber.StatusBadRequest,
			"Invalid "+parameter,
		)
	}
	return projectID, resourceID, nil
}

func evalHubActorID(c *fiber.Ctx) (uuid.UUID, bool) {
	return roadmapActorID(c)
}

func (h *EvalHubHandler) handleError(c *fiber.Ctx, err error) error {
	return roadmapAppError(
		c,
		h.logger,
		err,
		fiber.StatusInternalServerError,
		"Eval Hub request failed",
	)
}

func evalHubError(c *fiber.Ctx, status int, message string) error {
	return roadmapError(c, status, message)
}
