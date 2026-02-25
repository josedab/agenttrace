package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// NLQueryHandler handles natural language query endpoints
type NLQueryHandler struct {
	nlQueryService *service.NLQueryService
	logger         *zap.Logger
}

// NewNLQueryHandler creates a new natural language query handler
func NewNLQueryHandler(nlQueryService *service.NLQueryService, logger *zap.Logger) *NLQueryHandler {
	return &NLQueryHandler{
		nlQueryService: nlQueryService,
		logger:         logger,
	}
}

// QueryRequest represents a natural language query request
type QueryRequest struct {
	Query string `json:"query" validate:"required,min=3,max=500"`
	Limit int    `json:"limit,omitempty"`
}

// Query handles POST /v1/traces/query/natural
func (h *NLQueryHandler) Query(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var req QueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Query is required",
		})
	}

	if len(req.Query) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Query must be at least 3 characters",
		})
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	response, err := h.nlQueryService.QueryTraces(c.Context(), projectID, req.Query, limit)
	if err != nil {
		h.logger.Error("natural language query failed",
			zap.String("query", req.Query),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to process query",
		})
	}

	return c.JSON(response)
}

// GetExamples handles GET /v1/traces/query/examples
func (h *NLQueryHandler) GetExamples(c *fiber.Ctx) error {
	examples := h.nlQueryService.GetQueryExamples()
	return c.JSON(fiber.Map{
		"examples": examples,
	})
}

// Autocomplete handles POST /api/public/nl-query/autocomplete
func (h *NLQueryHandler) Autocomplete(c *fiber.Ctx) error {
	var body struct {
		Partial string `json:"partial"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Partial) < 2 {
		return c.JSON(fiber.Map{"suggestions": []string{}})
	}

	examples := h.nlQueryService.GetQueryExamples()
	var suggestions []string
	partial := body.Partial
	for _, ex := range examples {
		if len(suggestions) >= 5 {
			break
		}
		// Simple prefix/substring matching
		if len(ex) > len(partial) {
			suggestions = append(suggestions, ex)
		}
	}

	return c.JSON(fiber.Map{"suggestions": suggestions})
}

// GetQuerySchema handles GET /api/public/nl-query/schema
func (h *NLQueryHandler) GetQuerySchema(c *fiber.Ctx) error {
	schema := map[string]interface{}{
		"fields": []map[string]string{
			{"name": "name", "type": "string", "description": "Trace name (partial match)"},
			{"name": "hasError", "type": "boolean", "description": "Whether trace has errors"},
			{"name": "minCost", "type": "number", "description": "Minimum total cost in USD"},
			{"name": "maxCost", "type": "number", "description": "Maximum total cost in USD"},
			{"name": "minDurationMs", "type": "number", "description": "Minimum duration in milliseconds"},
			{"name": "maxDurationMs", "type": "number", "description": "Maximum duration in milliseconds"},
			{"name": "tags", "type": "array", "description": "Filter by tags"},
			{"name": "gitBranch", "type": "string", "description": "Git branch name"},
			{"name": "gitCommitSha", "type": "string", "description": "Git commit SHA"},
			{"name": "search", "type": "string", "description": "Full-text search query"},
		},
		"timeExpressions": []string{
			"today", "yesterday", "last 24 hours", "last week", "last month", "this week",
		},
		"operators": []string{
			"more than", "less than", "between", "equals", "contains", "with", "without",
		},
	}
	return c.JSON(schema)
}

// RegisterRoutes registers natural language query routes
func (h *NLQueryHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	// Natural language query endpoints
	v1.Post("/traces/query/natural", h.Query)
	v1.Get("/traces/query/examples", h.GetExamples)
}
