package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentMemoryHandler handles agent memory HTTP requests
type AgentMemoryHandler struct {
	service *service.AgentMemoryService
	logger  *zap.Logger
}

// NewAgentMemoryHandler creates a new agent memory handler
func NewAgentMemoryHandler(svc *service.AgentMemoryService, logger *zap.Logger) *AgentMemoryHandler {
	return &AgentMemoryHandler{
		service: svc,
		logger:  logger,
	}
}

// AnalyzeMemory handles POST /api/public/memory/analyze
func (h *AgentMemoryHandler) AnalyzeMemory(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.MemoryAnalysisInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	traceID, err := uuid.Parse(input.TraceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	timeline, err := h.service.AnalyzeMemory(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to analyze memory", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to analyze memory"})
	}

	return c.JSON(timeline)
}

// GetSnapshot handles GET /api/public/memory/traces/:traceId/snapshots/:stepIndex
func (h *AgentMemoryHandler) GetSnapshot(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	stepIndex, err := strconv.Atoi(c.Params("stepIndex"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid step index"})
	}

	snapshot, err := h.service.GetSnapshot(c.Context(), traceID, stepIndex)
	if err != nil {
		h.logger.Error("failed to get snapshot", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get snapshot"})
	}

	return c.JSON(snapshot)
}

// GetOptimizations handles GET /api/public/memory/optimizations
func (h *AgentMemoryHandler) GetOptimizations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	optimizations, err := h.service.GetOptimizations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get optimizations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get optimizations"})
	}

	return c.JSON(optimizations)
}
