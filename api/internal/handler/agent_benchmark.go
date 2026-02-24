package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentBenchmarkHandler handles agent performance benchmark HTTP requests
type AgentBenchmarkHandler struct {
	logger                *zap.Logger
	agentBenchmarkService *service.AgentBenchmarkService
}

// NewAgentBenchmarkHandler creates a new agent benchmark handler
func NewAgentBenchmarkHandler(
	agentBenchmarkService *service.AgentBenchmarkService,
	logger *zap.Logger,
) *AgentBenchmarkHandler {
	return &AgentBenchmarkHandler{
		logger:                logger,
		agentBenchmarkService: agentBenchmarkService,
	}
}

// CreateSuite creates a new benchmark suite
// @Summary Create benchmark suite
// @Description Create a new agent performance benchmark suite
// @Tags agent-benchmark
// @Accept json
// @Produce json
// @Param suite body domain.BenchmarkSuiteInput true "Suite configuration"
// @Success 201 {object} domain.BenchmarkSuite
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-benchmark/suites [post]
func (h *AgentBenchmarkHandler) CreateSuite(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.AgentBenchmarkSuiteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	suite, err := h.agentBenchmarkService.CreateSuite(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create benchmark suite", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create suite"})
	}

	return c.Status(fiber.StatusCreated).JSON(suite)
}

// ListSuites returns all benchmark suites for a project
// @Summary List benchmark suites
// @Description Get all agent performance benchmark suites for a project
// @Tags agent-benchmark
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.BenchmarkSuite
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-benchmark/suites [get]
func (h *AgentBenchmarkHandler) ListSuites(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	suites, err := h.agentBenchmarkService.ListSuites(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list benchmark suites", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list suites"})
	}

	return c.JSON(suites)
}

// GetSuite returns a specific benchmark suite
// @Summary Get benchmark suite
// @Description Get a specific agent performance benchmark suite by ID
// @Tags agent-benchmark
// @Accept json
// @Produce json
// @Param suiteId path string true "Suite ID"
// @Success 200 {object} domain.BenchmarkSuite
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-benchmark/suites/{suiteId} [get]
func (h *AgentBenchmarkHandler) GetSuite(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	suite, err := h.agentBenchmarkService.GetSuite(c.Context(), suiteID)
	if err != nil {
		h.logger.Error("failed to get benchmark suite", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get suite"})
	}

	return c.JSON(suite)
}

// RunBenchmarkRequest represents the request to run a benchmark
type RunBenchmarkRequest struct {
	SuiteID   string `json:"suiteId"`
	AgentName string `json:"agentName"`
	ModelName string `json:"modelName"`
}

// RunBenchmark runs a benchmark for a specific agent and model
// @Summary Run benchmark
// @Description Run an agent performance benchmark for a specific agent and model
// @Tags agent-benchmark
// @Accept json
// @Produce json
// @Param body body RunBenchmarkRequest true "Benchmark run request"
// @Success 201 {object} domain.BenchmarkRun
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-benchmark/runs [post]
func (h *AgentBenchmarkHandler) RunBenchmark(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req RunBenchmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	suiteID, err := uuid.Parse(req.SuiteID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	run, err := h.agentBenchmarkService.RunBenchmark(c.Context(), suiteID, req.AgentName, req.ModelName)
	if err != nil {
		h.logger.Error("failed to run benchmark", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run benchmark"})
	}

	return c.Status(fiber.StatusCreated).JSON(run)
}

// GetLeaderboard returns the benchmark leaderboard for a suite
// @Summary Get benchmark leaderboard
// @Description Get the agent performance leaderboard for a specific benchmark suite
// @Tags agent-benchmark
// @Accept json
// @Produce json
// @Param suiteId path string true "Suite ID"
// @Success 200 {object} domain.BenchmarkLeaderboard
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-benchmark/suites/{suiteId}/leaderboard [get]
func (h *AgentBenchmarkHandler) GetLeaderboard(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	suiteID, err := uuid.Parse(c.Params("suiteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid suite ID"})
	}

	leaderboard, err := h.agentBenchmarkService.GetLeaderboard(c.Context(), suiteID)
	if err != nil {
		h.logger.Error("failed to get benchmark leaderboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get leaderboard"})
	}

	return c.JSON(leaderboard)
}
