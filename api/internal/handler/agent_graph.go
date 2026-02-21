package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentGraphHandler handles agent graph HTTP requests
type AgentGraphHandler struct {
	agentGraphService *service.AgentGraphService
	logger            *zap.Logger
}

// NewAgentGraphHandler creates a new agent graph handler
func NewAgentGraphHandler(agentGraphService *service.AgentGraphService, logger *zap.Logger) *AgentGraphHandler {
	return &AgentGraphHandler{
		agentGraphService: agentGraphService,
		logger:            logger,
	}
}

// BuildGraph handles GET /agent-graph/traces/:traceId/graph
func (h *AgentGraphHandler) BuildGraph(c *fiber.Ctx) error {
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

	graph, err := h.agentGraphService.BuildGraph(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to build agent graph",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to build agent graph",
		})
	}

	return c.JSON(graph)
}

// CompareGraphs handles POST /agent-graph/compare
func (h *AgentGraphHandler) CompareGraphs(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var body struct {
		TraceIDA string `json:"traceIdA"`
		TraceIDB string `json:"traceIdB"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if body.TraceIDA == "" || body.TraceIDB == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Both traceIdA and traceIdB are required",
		})
	}

	comparison, err := h.agentGraphService.CompareGraphs(c.Context(), projectID, body.TraceIDA, body.TraceIDB)
	if err != nil {
		h.logger.Error("failed to compare agent graphs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to compare agent graphs",
		})
	}

	return c.JSON(comparison)
}
