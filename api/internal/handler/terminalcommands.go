package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TerminalCommandsHandler handles terminal command endpoints
type TerminalCommandsHandler struct {
	termCmdService *service.TerminalCommandService
	logger         *zap.Logger
}

// NewTerminalCommandsHandler creates a new terminal commands handler
func NewTerminalCommandsHandler(termCmdService *service.TerminalCommandService, logger *zap.Logger) *TerminalCommandsHandler {
	return &TerminalCommandsHandler{
		termCmdService: termCmdService,
		logger:         logger,
	}
}

// ListTerminalCommands handles GET /v1/terminal-commands
func (h *TerminalCommandsHandler) ListTerminalCommands(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseTerminalCommandFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.termCmdService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list terminal commands", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list terminal commands",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.TerminalCommands,
		"totalCount": list.TotalCount,
		"hasMore":    list.HasMore,
	})
}

// CreateTerminalCommand handles POST /v1/terminal-commands
func (h *TerminalCommandsHandler) CreateTerminalCommand(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.TerminalCommandInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	if input.Command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "command is required",
		})
	}

	cmd, err := h.termCmdService.Log(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create terminal command", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create terminal command",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(cmd)
}

// BatchCreateTerminalCommands handles POST /v1/terminal-commands/batch
func (h *TerminalCommandsHandler) BatchCreateTerminalCommands(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request struct {
		Commands []*domain.TerminalCommandInput `json:"commands"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if len(request.Commands) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "At least one command required",
		})
	}

	if len(request.Commands) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Maximum 100 commands per batch",
		})
	}

	results, err := h.termCmdService.LogBatch(c.Context(), projectID, request.Commands)
	if err != nil {
		h.logger.Error("failed to batch create terminal commands", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create terminal commands",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": results,
	})
}

// GetTraceTerminalCommands handles GET /v1/traces/:traceId/terminal-commands
func (h *TerminalCommandsHandler) GetTraceTerminalCommands(c *fiber.Ctx) error {
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
			"message": "Trace ID required",
		})
	}

	cmds, err := h.termCmdService.GetByTraceID(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get trace terminal commands", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace terminal commands",
		})
	}

	return c.JSON(fiber.Map{
		"data": cmds,
	})
}

// GetTerminalCommandStats handles GET /v1/terminal-commands/stats
func (h *TerminalCommandsHandler) GetTerminalCommandStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var traceID *string
	if tid := c.Query("traceId"); tid != "" {
		traceID = &tid
	}

	stats, err := h.termCmdService.GetStats(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get terminal command stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get terminal command stats",
		})
	}

	return c.JSON(stats)
}

// parseTerminalCommandFilter parses terminal command filter from query params
func (h *TerminalCommandsHandler) parseTerminalCommandFilter(c *fiber.Ctx) *domain.TerminalCommandFilter {
	filter := &domain.TerminalCommandFilter{}

	if traceID := c.Query("traceId"); traceID != "" {
		filter.TraceID = &traceID
	}

	if obsID := c.Query("observationId"); obsID != "" {
		filter.ObservationID = &obsID
	}

	if command := c.Query("command"); command != "" {
		filter.Command = &command
	}

	if exitCode := c.Query("exitCode"); exitCode != "" {
		ec := parseIntParam(c, "exitCode", 0)
		ec32 := int32(ec)
		filter.ExitCode = &ec32
	}

	if success := c.Query("success"); success != "" {
		s := success == "true"
		filter.Success = &s
	}

	if from := c.Query("fromTimestamp"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.FromTime = &t
		}
	}

	if to := c.Query("toTimestamp"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.ToTime = &t
		}
	}

	return filter
}

// RegisterRoutes registers terminal command routes
func (h *TerminalCommandsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/terminal-commands", h.ListTerminalCommands)
	v1.Get("/terminal-commands/stats", h.GetTerminalCommandStats)
	v1.Post("/terminal-commands", h.CreateTerminalCommand)
	v1.Post("/terminal-commands/batch", h.BatchCreateTerminalCommands)

	v1.Get("/traces/:traceId/terminal-commands", h.GetTraceTerminalCommands)
}
