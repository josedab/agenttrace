package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// SemanticTraceSearchHandler handles semantic trace search HTTP requests
type SemanticTraceSearchHandler struct {
	logger                    *zap.Logger
	semanticTraceSearchService *service.SemanticTraceSearchService
}

// NewSemanticTraceSearchHandler creates a new semantic trace search handler
func NewSemanticTraceSearchHandler(
	semanticTraceSearchService *service.SemanticTraceSearchService,
	logger *zap.Logger,
) *SemanticTraceSearchHandler {
	return &SemanticTraceSearchHandler{
		logger:                    logger,
		semanticTraceSearchService: semanticTraceSearchService,
	}
}

// SearchRequest represents the request to perform a semantic trace search
type SearchRequest struct {
	Query   string            `json:"query"`
	Filters map[string]string `json:"filters,omitempty"`
}

// Search performs a semantic search across traces
// @Summary Semantic trace search
// @Description Perform a semantic search across traces using natural language queries
// @Tags semantic-trace-search
// @Accept json
// @Produce json
// @Param body body SearchRequest true "Search request"
// @Success 200 {object} domain.SemanticSearchResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/semantic-trace-search/search [post]
func (h *SemanticTraceSearchHandler) Search(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}

	query := domain.SemanticTraceSearchQuery{
		Query:     req.Query,
		ProjectID: projectID,
	}

	result, err := h.semanticTraceSearchService.Search(c.Context(), query)
	if err != nil {
		h.logger.Error("failed to perform semantic trace search", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to perform search"})
	}

	return c.JSON(result)
}

// GetClusters returns trace clusters for a project
// @Summary Get trace clusters
// @Description Get automatically detected trace clusters for a project
// @Tags semantic-trace-search
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.TraceCluster
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/semantic-trace-search/clusters [get]
func (h *SemanticTraceSearchHandler) GetClusters(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	clusters, err := h.semanticTraceSearchService.GetClusters(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get trace clusters", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get clusters"})
	}

	return c.JSON(clusters)
}

// GetAnomalyPatterns returns anomaly patterns detected across traces
// @Summary Get anomaly patterns
// @Description Get anomaly patterns detected through semantic analysis of traces
// @Tags semantic-trace-search
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.AnomalyPattern
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/semantic-trace-search/anomaly-patterns [get]
func (h *SemanticTraceSearchHandler) GetAnomalyPatterns(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	patterns, err := h.semanticTraceSearchService.GetAnomalyPatterns(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get anomaly patterns", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get anomaly patterns"})
	}

	return c.JSON(patterns)
}

// GetDashboard returns the semantic trace search dashboard for a project
// @Summary Get semantic search dashboard
// @Description Get the semantic trace search dashboard with search insights and statistics
// @Tags semantic-trace-search
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {object} domain.SemanticSearchDashboard
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/semantic-trace-search/dashboard [get]
func (h *SemanticTraceSearchHandler) GetDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.semanticTraceSearchService.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get semantic search dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard"})
	}

	return c.JSON(dashboard)
}
