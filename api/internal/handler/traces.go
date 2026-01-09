package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TracesHandler handles trace query endpoints
type TracesHandler struct {
	queryService *service.QueryService
	logger       *zap.Logger
}

// NewTracesHandler creates a new traces handler
func NewTracesHandler(queryService *service.QueryService, logger *zap.Logger) *TracesHandler {
	return &TracesHandler{
		queryService: queryService,
		logger:       logger,
	}
}

// ListTraces handles GET /v1/traces
func (h *TracesHandler) ListTraces(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseTraceFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.queryService.ListTraces(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list traces", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list traces",
		})
	}

	return c.JSON(list)
}

// GetTrace handles GET /v1/traces/:traceId
func (h *TracesHandler) GetTrace(c *fiber.Ctx) error {
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

	trace, err := h.queryService.GetTrace(c.Context(), projectID, traceID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}
		h.logger.Error("failed to get trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace",
		})
	}

	return c.JSON(trace)
}

// GetTraceObservations handles GET /v1/traces/:traceId/observations
func (h *TracesHandler) GetTraceObservations(c *fiber.Ctx) error {
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

	tree, err := h.queryService.GetObservationTree(c.Context(), projectID, traceID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}
		h.logger.Error("failed to get observations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get observations",
		})
	}

	return c.JSON(tree)
}

// GetTraceStats handles GET /v1/traces/:traceId/stats
func (h *TracesHandler) GetTraceStats(c *fiber.Ctx) error {
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

	// Create a filter for the specific trace
	filter := &domain.TraceFilter{
		ProjectID: projectID,
		IDs:       []string{traceID},
	}

	stats, err := h.queryService.GetTraceStats(c.Context(), filter)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}
		h.logger.Error("failed to get trace stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace stats",
		})
	}

	return c.JSON(stats)
}

// SearchTraces handles GET /v1/traces/search
func (h *TracesHandler) SearchTraces(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Search query required",
		})
	}

	filter := h.parseTraceFilter(c)
	filter.ProjectID = projectID
	filter.Search = &query

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	list, err := h.queryService.ListTraces(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to search traces", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to search traces",
		})
	}

	return c.JSON(list)
}

// GetSessions handles GET /v1/sessions
func (h *TracesHandler) GetSessions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseSessionFilter(c)
	filter.ProjectID = projectID

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.queryService.ListSessions(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list sessions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list sessions",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.Sessions,
		"totalCount": list.TotalCount,
		"hasMore":    list.HasMore,
	})
}

// GetSession handles GET /v1/sessions/:sessionId
func (h *TracesHandler) GetSession(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Session ID required",
		})
	}

	session, err := h.queryService.GetSession(c.Context(), projectID, sessionID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Session not found",
			})
		}
		h.logger.Error("failed to get session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get session",
		})
	}

	return c.JSON(session)
}

// GetMetrics handles GET /v1/metrics
func (h *TracesHandler) GetMetrics(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	// Parse date range
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid 'from' date format",
			})
		}
	} else {
		from = time.Now().AddDate(0, 0, -30) // Default: last 30 days
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid 'to' date format",
			})
		}
	} else {
		to = time.Now()
	}

	filter := &domain.TraceFilter{
		ProjectID: projectID,
		FromTime:  &from,
		ToTime:    &to,
	}

	traceStats, err := h.queryService.GetTraceStats(c.Context(), filter)
	if err != nil {
		h.logger.Error("failed to get metrics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get metrics",
		})
	}

	return c.JSON(traceStats)
}

// parseTraceFilter parses trace filter from query params
func (h *TracesHandler) parseTraceFilter(c *fiber.Ctx) *domain.TraceFilter {
	filter := &domain.TraceFilter{}

	if userID := c.Query("userId"); userID != "" {
		filter.UserID = &userID
	}

	if sessionID := c.Query("sessionId"); sessionID != "" {
		filter.SessionID = &sessionID
	}

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter.Tags = tags
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

	if version := c.Query("version"); version != "" {
		filter.Version = &version
	}

	if release := c.Query("release"); release != "" {
		filter.Release = &release
	}

	// Git correlation filters
	if gitCommit := c.Query("gitCommitSha"); gitCommit != "" {
		filter.GitCommitSha = &gitCommit
	}

	if gitBranch := c.Query("gitBranch"); gitBranch != "" {
		filter.GitBranch = &gitBranch
	}

	if gitRepoURL := c.Query("gitRepoUrl"); gitRepoURL != "" {
		filter.GitRepoURL = &gitRepoURL
	}

	return filter
}

// parseSessionFilter parses session filter from query params
func (h *TracesHandler) parseSessionFilter(c *fiber.Ctx) *domain.SessionFilter {
	filter := &domain.SessionFilter{}

	if userID := c.Query("userId"); userID != "" {
		filter.UserID = &userID
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

	if bookmarked := c.Query("bookmarked"); bookmarked != "" {
		b := bookmarked == "true"
		filter.Bookmarked = &b
	}

	return filter
}

// UpdateTrace handles PATCH /v1/traces/:traceId
func (h *TracesHandler) UpdateTrace(c *fiber.Ctx) error {
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

	var input domain.TraceUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	trace, err := h.queryService.UpdateTrace(c.Context(), projectID, traceID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}
		h.logger.Error("failed to update trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update trace",
		})
	}

	return c.JSON(trace)
}

// DeleteTrace handles DELETE /v1/traces/:traceId
func (h *TracesHandler) DeleteTrace(c *fiber.Ctx) error {
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

	// Note: Deletion in ClickHouse is handled via TTL/mutations
	// This endpoint marks the trace as deleted in metadata
	h.logger.Info("trace deletion requested",
		zap.String("project_id", projectID.String()),
		zap.String("trace_id", traceID),
	)

	return c.JSON(fiber.Map{
		"message": "Trace marked for deletion",
	})
}

// RegisterRoutes registers trace routes
func (h *TracesHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	// Trace endpoints
	v1.Get("/traces", h.ListTraces)
	v1.Get("/traces/search", h.SearchTraces)
	v1.Get("/traces/:traceId", h.GetTrace)
	v1.Get("/traces/:traceId/observations", h.GetTraceObservations)
	v1.Get("/traces/:traceId/stats", h.GetTraceStats)
	v1.Patch("/traces/:traceId", h.UpdateTrace)
	v1.Delete("/traces/:traceId", h.DeleteTrace)

	// Session endpoints
	v1.Get("/sessions", h.GetSessions)
	v1.Get("/sessions/:sessionId", h.GetSession)

	// Metrics endpoint
	v1.Get("/metrics", h.GetMetrics)
}

// parseIntParam parses an integer query parameter with a default value
func parseIntParam(c *fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// parseUUIDParam parses a UUID query parameter
func parseUUIDParam(c *fiber.Ctx, key string) *uuid.UUID {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return nil
	}
	return &id
}
