package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// BenchmarksHandler handles benchmark HTTP requests
type BenchmarksHandler struct {
	benchmarkService *service.BenchmarkService
	logger           *zap.Logger
}

// NewBenchmarksHandler creates a new benchmarks handler
func NewBenchmarksHandler(benchmarkService *service.BenchmarkService, logger *zap.Logger) *BenchmarksHandler {
	return &BenchmarksHandler{
		benchmarkService: benchmarkService,
		logger:           logger,
	}
}

// ListBenchmarks handles GET /benchmarks
func (h *BenchmarksHandler) ListBenchmarks(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var category *domain.BenchmarkCategory
	if cat := c.Query("category"); cat != "" {
		bc := domain.BenchmarkCategory(cat)
		category = &bc
	}

	benchmarks, err := h.benchmarkService.ListBenchmarks(c.Context(), category)
	if err != nil {
		h.logger.Error("failed to list benchmarks", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list benchmarks",
		})
	}

	return c.JSON(benchmarks)
}

// GetBenchmark handles GET /benchmarks/:benchmarkId
func (h *BenchmarksHandler) GetBenchmark(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	benchmarkID, err := uuid.Parse(c.Params("benchmarkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid benchmark ID",
		})
	}

	benchmark, err := h.benchmarkService.GetBenchmark(c.Context(), benchmarkID)
	if err != nil {
		h.logger.Error("failed to get benchmark",
			zap.String("benchmarkId", benchmarkID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Benchmark not found",
		})
	}

	return c.JSON(benchmark)
}

// Submit handles POST /benchmarks/:benchmarkId/submit
func (h *BenchmarksHandler) Submit(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	benchmarkID, err := uuid.Parse(c.Params("benchmarkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid benchmark ID",
		})
	}

	var input domain.SubmitBenchmarkInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}
	input.BenchmarkID = benchmarkID

	submission, err := h.benchmarkService.Submit(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to submit benchmark",
			zap.String("benchmarkId", benchmarkID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to submit benchmark",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(submission)
}

// CreateBenchmark handles POST /api/public/benchmarks
func (h *BenchmarksHandler) CreateBenchmark(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CreateBenchmarkInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	benchmark := &domain.Benchmark{
		Name:         input.Name,
		Description:  input.Description,
		Category:     input.Category,
		DatasetID:    input.DatasetID,
		EvaluatorIDs: input.EvaluatorIDs,
		Metrics:      input.Metrics,
		IsPublic:     input.IsPublic,
	}
	_ = projectID // projectID available for future use

	result, err := h.benchmarkService.CreateBenchmark(c.Context(), benchmark)
	if err != nil {
		h.logger.Error("failed to create benchmark", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create benchmark"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// CompareSubmissions handles POST /api/public/benchmarks/:benchmarkId/compare
func (h *BenchmarksHandler) CompareSubmissions(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	benchmarkID, err := uuid.Parse(c.Params("benchmarkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid benchmark ID"})
	}

	var input domain.CompareInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	comparison, err := h.benchmarkService.CompareSubmissions(c.Context(), benchmarkID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to compare"})
	}

	return c.JSON(comparison)
}

// GetStats handles GET /api/public/benchmarks/:benchmarkId/stats
func (h *BenchmarksHandler) GetStats(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	benchmarkID, err := uuid.Parse(c.Params("benchmarkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid benchmark ID"})
	}

	stats, err := h.benchmarkService.GetBenchmarkStats(c.Context(), benchmarkID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(stats)
}

// GetLeaderboard handles GET /benchmarks/:benchmarkId/leaderboard
func (h *BenchmarksHandler) GetLeaderboard(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	benchmarkID, err := uuid.Parse(c.Params("benchmarkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid benchmark ID",
		})
	}

	limit := parseIntParam(c, "limit", 50)

	leaderboard, err := h.benchmarkService.GetLeaderboard(c.Context(), benchmarkID, limit)
	if err != nil {
		h.logger.Error("failed to get leaderboard",
			zap.String("benchmarkId", benchmarkID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get leaderboard",
		})
	}

	return c.JSON(leaderboard)
}
