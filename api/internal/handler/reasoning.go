package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ReasoningHandler handles reasoning tree HTTP requests
type ReasoningHandler struct {
	reasoningService *service.ReasoningService
	logger           *zap.Logger
}

// NewReasoningHandler creates a new reasoning handler
func NewReasoningHandler(reasoningService *service.ReasoningService, logger *zap.Logger) *ReasoningHandler {
	return &ReasoningHandler{
		reasoningService: reasoningService,
		logger:           logger,
	}
}

// GetReasoningTree handles GET /traces/:traceId/reasoning
func (h *ReasoningHandler) GetReasoningTree(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Trace ID is required",
		})
	}

	tree, err := h.reasoningService.BuildReasoningTree(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to build reasoning tree",
			zap.String("projectId", projectID.String()),
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to build reasoning tree",
		})
	}

	return c.JSON(tree)
}

// GetNode handles GET /traces/:traceId/reasoning/:nodeId
func (h *ReasoningHandler) GetNode(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Trace ID is required",
		})
	}

	nodeID := c.Params("nodeId")
	if nodeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Node ID is required",
		})
	}

	node, err := h.reasoningService.GetNode(c.Context(), projectID, traceID, nodeID)
	if err != nil {
		h.logger.Error("failed to get reasoning node",
			zap.String("projectId", projectID.String()),
			zap.String("traceId", traceID),
			zap.String("nodeId", nodeID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Reasoning node not found",
		})
	}

	return c.JSON(node)
}
