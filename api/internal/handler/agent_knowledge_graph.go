package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AgentKnowledgeGraphHandler handles agent knowledge graph HTTP requests
type AgentKnowledgeGraphHandler struct {
	logger                    *zap.Logger
	agentKnowledgeGraphService *service.AgentKnowledgeGraphService
}

// NewAgentKnowledgeGraphHandler creates a new agent knowledge graph handler
func NewAgentKnowledgeGraphHandler(
	agentKnowledgeGraphService *service.AgentKnowledgeGraphService,
	logger *zap.Logger,
) *AgentKnowledgeGraphHandler {
	return &AgentKnowledgeGraphHandler{
		logger:                    logger,
		agentKnowledgeGraphService: agentKnowledgeGraphService,
	}
}

// BuildGraph builds or retrieves the agent knowledge graph
// @Summary Build agent knowledge graph
// @Description Build or retrieve the agent knowledge graph with optional focus node and depth
// @Tags agent-knowledge-graph
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param focusNode query string false "Focus node ID to center the graph on"
// @Param depth query int false "Depth of graph traversal" default(3)
// @Success 200 {object} domain.KnowledgeGraph
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-knowledge-graph/graph [get]
func (h *AgentKnowledgeGraphHandler) BuildGraph(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	focusNode := c.Query("focusNode")

	graph, err := h.agentKnowledgeGraphService.BuildGraph(c.Context(), projectID, focusNode)
	if err != nil {
		h.logger.Error("failed to build agent knowledge graph", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build graph"})
	}

	return c.JSON(graph)
}

// QueryGraphRequest represents the request to query the knowledge graph
type QueryGraphRequest struct {
	Query string `json:"query"`
}

// QueryGraph queries the agent knowledge graph
// @Summary Query agent knowledge graph
// @Description Query the agent knowledge graph using natural language or structured queries
// @Tags agent-knowledge-graph
// @Accept json
// @Produce json
// @Param body body QueryGraphRequest true "Graph query request"
// @Success 200 {object} domain.KnowledgeGraphQueryResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-knowledge-graph/query [post]
func (h *AgentKnowledgeGraphHandler) QueryGraph(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req QueryGraphRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}

	result, err := h.agentKnowledgeGraphService.QueryGraph(c.Context(), projectID, req.Query)
	if err != nil {
		h.logger.Error("failed to query agent knowledge graph", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query graph"})
	}

	return c.JSON(result)
}

// GetEvolution returns the evolution of the agent knowledge graph over time
// @Summary Get knowledge graph evolution
// @Description Get the evolution of the agent knowledge graph over time
// @Tags agent-knowledge-graph
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.KnowledgeGraphEvolution
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-knowledge-graph/evolution [get]
func (h *AgentKnowledgeGraphHandler) GetEvolution(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	evolution, err := h.agentKnowledgeGraphService.GetEvolution(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get knowledge graph evolution", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get evolution"})
	}

	return c.JSON(evolution)
}

// GetStats returns statistics about the agent knowledge graph
// @Summary Get knowledge graph stats
// @Description Get statistics about the agent knowledge graph
// @Tags agent-knowledge-graph
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.KnowledgeGraphStats
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/agent-knowledge-graph/stats [get]
func (h *AgentKnowledgeGraphHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.agentKnowledgeGraphService.GetStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get knowledge graph stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(stats)
}
