package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CIRunsHandler handles CI run endpoints
type CIRunsHandler struct {
	ciRunService *service.CIRunService
	logger       *zap.Logger
}

// NewCIRunsHandler creates a new CI runs handler
func NewCIRunsHandler(ciRunService *service.CIRunService, logger *zap.Logger) *CIRunsHandler {
	return &CIRunsHandler{
		ciRunService: ciRunService,
		logger:       logger,
	}
}

// ListCIRuns handles GET /v1/ci-runs
func (h *CIRunsHandler) ListCIRuns(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseCIRunFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.ciRunService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list CI runs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list CI runs",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.CIRuns,
		"totalCount": list.TotalCount,
		"hasMore":    list.HasMore,
	})
}

// GetCIRun handles GET /v1/ci-runs/:ciRunId
func (h *CIRunsHandler) GetCIRun(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	ciRunIDStr := c.Params("ciRunId")
	if ciRunIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "CI run ID required",
		})
	}

	ciRunID, err := uuid.Parse(ciRunIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid CI run ID format",
		})
	}

	ciRun, err := h.ciRunService.Get(c.Context(), projectID, ciRunID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "CI run not found",
			})
		}
		h.logger.Error("failed to get CI run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get CI run",
		})
	}

	return c.JSON(ciRun)
}

// GetCIRunByProviderID handles GET /v1/ci-runs/provider/:providerRunId
func (h *CIRunsHandler) GetCIRunByProviderID(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	providerRunID := c.Params("providerRunId")
	if providerRunID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Provider run ID required",
		})
	}

	ciRun, err := h.ciRunService.GetByProviderRunID(c.Context(), projectID, providerRunID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "CI run not found",
			})
		}
		h.logger.Error("failed to get CI run by provider ID", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get CI run",
		})
	}

	return c.JSON(ciRun)
}

// CreateCIRun handles POST /v1/ci-runs
func (h *CIRunsHandler) CreateCIRun(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.CIRunInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if input.Provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "provider is required",
		})
	}

	if input.ProviderRunID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "providerRunId is required",
		})
	}

	ciRun, err := h.ciRunService.Create(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create CI run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create CI run",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ciRun)
}

// UpdateCIRun handles PATCH /v1/ci-runs/:ciRunId
func (h *CIRunsHandler) UpdateCIRun(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	ciRunIDStr := c.Params("ciRunId")
	if ciRunIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "CI run ID required",
		})
	}

	ciRunID, err := uuid.Parse(ciRunIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid CI run ID format",
		})
	}

	var input domain.CIRunUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	ciRun, err := h.ciRunService.Update(c.Context(), projectID, ciRunID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "CI run not found",
			})
		}
		h.logger.Error("failed to update CI run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update CI run",
		})
	}

	return c.JSON(ciRun)
}

// AddTraceToCIRun handles POST /v1/ci-runs/:ciRunId/traces
func (h *CIRunsHandler) AddTraceToCIRun(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	ciRunIDStr := c.Params("ciRunId")
	if ciRunIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "CI run ID required",
		})
	}

	ciRunID, err := uuid.Parse(ciRunIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid CI run ID format",
		})
	}

	var request struct {
		TraceID string `json:"traceId"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if request.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	err = h.ciRunService.AddTrace(c.Context(), projectID, ciRunID, request.TraceID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "CI run not found",
			})
		}
		h.logger.Error("failed to add trace to CI run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to add trace to CI run",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// CompleteCIRun handles POST /v1/ci-runs/:ciRunId/complete
func (h *CIRunsHandler) CompleteCIRun(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	ciRunIDStr := c.Params("ciRunId")
	if ciRunIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "CI run ID required",
		})
	}

	ciRunID, err := uuid.Parse(ciRunIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid CI run ID format",
		})
	}

	var request struct {
		Status     domain.CIRunStatus `json:"status"`
		Conclusion string             `json:"conclusion"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if request.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "status is required",
		})
	}

	ciRun, err := h.ciRunService.Complete(c.Context(), projectID, ciRunID, request.Status, request.Conclusion)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "CI run not found",
			})
		}
		h.logger.Error("failed to complete CI run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to complete CI run",
		})
	}

	return c.JSON(ciRun)
}

// GetCIRunStats handles GET /v1/ci-runs/stats
func (h *CIRunsHandler) GetCIRunStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	stats, err := h.ciRunService.GetStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get CI run stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get CI run stats",
		})
	}

	return c.JSON(stats)
}

// parseCIRunFilter parses CI run filter from query params
func (h *CIRunsHandler) parseCIRunFilter(c *fiber.Ctx) *domain.CIRunFilter {
	filter := &domain.CIRunFilter{}

	if provider := c.Query("provider"); provider != "" {
		p := domain.CIProvider(provider)
		filter.Provider = &p
	}

	if providerRunID := c.Query("providerRunId"); providerRunID != "" {
		filter.ProviderRunID = &providerRunID
	}

	if commitSha := c.Query("gitCommitSha"); commitSha != "" {
		filter.GitCommitSha = &commitSha
	}

	if branch := c.Query("gitBranch"); branch != "" {
		filter.GitBranch = &branch
	}

	if status := c.Query("status"); status != "" {
		s := domain.CIRunStatus(status)
		filter.Status = &s
	}

	if from := c.Query("fromTimestamp"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.FromTime = &t
		}
	}

	if to := c.Query("toTimestamp"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.ToTime = &t
		}
	}

	return filter
}

// RegisterRoutes registers CI run routes
func (h *CIRunsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/ci-runs", h.ListCIRuns)
	v1.Get("/ci-runs/stats", h.GetCIRunStats)
	v1.Get("/ci-runs/provider/:providerRunId", h.GetCIRunByProviderID)
	v1.Get("/ci-runs/:ciRunId", h.GetCIRun)
	v1.Post("/ci-runs", h.CreateCIRun)
	v1.Patch("/ci-runs/:ciRunId", h.UpdateCIRun)
	v1.Post("/ci-runs/:ciRunId/traces", h.AddTraceToCIRun)
	v1.Post("/ci-runs/:ciRunId/complete", h.CompleteCIRun)
}
