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

// RegisterRoutes registers natural language query routes
func (h *NLQueryHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	// Natural language query endpoints
	v1.Post("/traces/query/natural", h.Query)
	v1.Get("/traces/query/examples", h.GetExamples)
}
