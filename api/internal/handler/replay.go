package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ReplayHandler handles replay-related HTTP requests
type ReplayHandler struct {
	logger        *zap.Logger
	replayService *service.ReplayService
}

// NewReplayHandler creates a new replay handler
func NewReplayHandler(
	logger *zap.Logger,
	replayService *service.ReplayService,
) *ReplayHandler {
	return &ReplayHandler{
		logger:        logger,
		replayService: replayService,
	}
}

// GetTimeline returns the replay timeline for a trace
// @Summary Get replay timeline
// @Description Get the complete replay timeline for a trace
// @Tags replay
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Success 200 {object} domain.ReplayTimeline
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/traces/{traceId}/replay [get]
func (h *ReplayHandler) GetTimeline(c *fiber.Ctx) error {
	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid trace ID", err)
	}

	h.logger.Debug("Getting replay timeline", zap.String("traceId", traceID.String()))

	// In a real implementation, this would:
	// 1. Fetch the trace from the trace repository
	// 2. Fetch all observations for the trace
	// 3. Fetch file operations
	// 4. Fetch terminal commands
	// 5. Fetch checkpoints
	// 6. Fetch git links
	// 7. Build the timeline using the replay service

	// For now, return a mock timeline demonstrating the structure
	timeline := &domain.ReplayTimeline{
		TraceID:   traceID,
		TraceName: "sample-agent-run",
		Events:    []domain.ReplayEvent{},
		Summary: domain.ReplaySummary{
			TotalEvents: 0,
		},
	}

	return c.JSON(timeline)
}

// ExportTimeline exports the replay timeline in a portable format
// @Summary Export replay timeline
// @Description Export the replay timeline for sharing or offline viewing
// @Tags replay
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Success 200 {object} domain.ReplayExport
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/traces/{traceId}/replay/export [get]
func (h *ReplayHandler) ExportTimeline(c *fiber.Ctx) error {
	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid trace ID", err)
	}

	h.logger.Debug("Exporting replay timeline", zap.String("traceId", traceID.String()))

	// Get the timeline first
	timeline, err := h.replayService.GetTimelineForTrace(c.Context(), traceID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to build timeline", err)
	}

	// Export it
	export := h.replayService.ExportTimeline(timeline)

	// Set download headers
	c.Set("Content-Disposition", "attachment; filename=trace-replay-"+traceID.String()+".json")

	return c.JSON(export)
}

// GetTimelineEvents returns a paginated list of replay events
// @Summary Get replay events
// @Description Get paginated replay events with optional filtering
// @Tags replay
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Param types query []string false "Filter by event types"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} ReplayEventsResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/public/traces/{traceId}/replay/events [get]
func (h *ReplayHandler) GetTimelineEvents(c *fiber.Ctx) error {
	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid trace ID", err)
	}

	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	h.logger.Debug("Getting replay events",
		zap.String("traceId", traceID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Return empty events for now
	response := ReplayEventsResponse{
		Events:     []domain.ReplayEvent{},
		TotalCount: 0,
		HasMore:    false,
		Limit:      limit,
		Offset:     offset,
	}

	return c.JSON(response)
}

// ReplayEventsResponse represents the response for paginated replay events
type ReplayEventsResponse struct {
	Events     []domain.ReplayEvent `json:"events"`
	TotalCount int                  `json:"totalCount"`
	HasMore    bool                 `json:"hasMore"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
}

// GetEventDetails returns detailed information about a specific replay event
// @Summary Get event details
// @Description Get detailed information about a specific replay event
// @Tags replay
// @Accept json
// @Produce json
// @Param traceId path string true "Trace ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} domain.ReplayEvent
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/traces/{traceId}/replay/events/{eventId} [get]
func (h *ReplayHandler) GetEventDetails(c *fiber.Ctx) error {
	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid trace ID", err)
	}

	eventID := c.Params("eventId")
	if eventID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Event ID required", nil)
	}

	h.logger.Debug("Getting event details",
		zap.String("traceId", traceID.String()),
		zap.String("eventId", eventID),
	)

	// In a real implementation, this would fetch the specific event
	return errorResponse(c, fiber.StatusNotFound, "Event not found", nil)
}

// CompareTimelines compares two replay timelines
// @Summary Compare timelines
// @Description Compare two replay timelines to see differences
// @Tags replay
// @Accept json
// @Produce json
// @Param body body CompareTimelinesRequest true "Trace IDs to compare"
// @Success 200 {object} TimelineComparison
// @Failure 400 {object} ErrorResponse
// @Router /api/public/replay/compare [post]
func (h *ReplayHandler) CompareTimelines(c *fiber.Ctx) error {
	var req CompareTimelinesRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.TraceID1 == uuid.Nil || req.TraceID2 == uuid.Nil {
		return errorResponse(c, fiber.StatusBadRequest, "Both trace IDs are required", nil)
	}

	h.logger.Debug("Comparing timelines",
		zap.String("traceId1", req.TraceID1.String()),
		zap.String("traceId2", req.TraceID2.String()),
	)

	// Build comparison
	comparison := TimelineComparison{
		TraceID1:    req.TraceID1,
		TraceID2:    req.TraceID2,
		Summary1:    domain.ReplaySummary{},
		Summary2:    domain.ReplaySummary{},
		Differences: []TimelineDifference{},
	}

	return c.JSON(comparison)
}

// CompareTimelinesRequest represents the request body for comparing timelines
type CompareTimelinesRequest struct {
	TraceID1 uuid.UUID `json:"traceId1"`
	TraceID2 uuid.UUID `json:"traceId2"`
}

// TimelineComparison represents the result of comparing two timelines
type TimelineComparison struct {
	TraceID1    uuid.UUID             `json:"traceId1"`
	TraceID2    uuid.UUID             `json:"traceId2"`
	Summary1    domain.ReplaySummary  `json:"summary1"`
	Summary2    domain.ReplaySummary  `json:"summary2"`
	Differences []TimelineDifference  `json:"differences"`
}

// TimelineDifference represents a difference between two timelines
type TimelineDifference struct {
	Type        string `json:"type"` // added, removed, changed
	Description string `json:"description"`
	Event1      *domain.ReplayEvent `json:"event1,omitempty"`
	Event2      *domain.ReplayEvent `json:"event2,omitempty"`
}
