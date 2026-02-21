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
