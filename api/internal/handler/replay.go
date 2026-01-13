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
	traceID := c.Params("traceId")
	if traceID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Trace ID required")
	}

	// Get project ID from context (set by API key middleware)
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	h.logger.Debug("Getting replay timeline",
		zap.String("traceId", traceID),
		zap.String("projectId", projectID.String()),
	)

	// Fetch all data and build timeline
	timeline, err := h.replayService.GetTimelineForTrace(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get replay timeline",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to build timeline")
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
	traceID := c.Params("traceId")
	if traceID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Trace ID required")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	h.logger.Debug("Exporting replay timeline",
		zap.String("traceId", traceID),
		zap.String("projectId", projectID.String()),
	)

	// Get the timeline first
	timeline, err := h.replayService.GetTimelineForTrace(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get timeline for export",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to build timeline")
	}

	// Export it
	export := h.replayService.ExportTimeline(timeline)

	// Set download headers
	c.Set("Content-Disposition", "attachment; filename=trace-replay-"+traceID+".json")

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
	traceID := c.Params("traceId")
	if traceID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Trace ID required")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	h.logger.Debug("Getting replay events",
		zap.String("traceId", traceID),
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Get full timeline first
	timeline, err := h.replayService.GetTimelineForTrace(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get timeline for events",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get events")
	}

	// Apply pagination
	totalCount := len(timeline.Events)
	start := offset
	if start > totalCount {
		start = totalCount
	}
	end := start + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedEvents := timeline.Events[start:end]

	response := ReplayEventsResponse{
		Events:     paginatedEvents,
		TotalCount: totalCount,
		HasMore:    end < totalCount,
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
	traceID := c.Params("traceId")
	if traceID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Trace ID required")
	}

	eventID := c.Params("eventId")
	if eventID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Event ID required")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	h.logger.Debug("Getting event details",
		zap.String("traceId", traceID),
		zap.String("eventId", eventID),
		zap.String("projectId", projectID.String()),
	)

	// Get the full timeline and find the event
	timeline, err := h.replayService.GetTimelineForTrace(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get timeline for event details",
			zap.String("traceId", traceID),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get event")
	}

	// Find the specific event
	for _, event := range timeline.Events {
		if event.ID == eventID {
			return c.JSON(event)
		}
	}

	return errorResponse(c, fiber.StatusNotFound, "Event not found")
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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.TraceID1 == uuid.Nil || req.TraceID2 == uuid.Nil {
		return errorResponse(c, fiber.StatusBadRequest, "Both trace IDs are required")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	h.logger.Debug("Comparing timelines",
		zap.String("traceId1", req.TraceID1.String()),
		zap.String("traceId2", req.TraceID2.String()),
		zap.String("projectId", projectID.String()),
	)

	// Fetch first timeline
	timeline1, err := h.replayService.GetTimelineForTrace(c.Context(), projectID, req.TraceID1.String())
	if err != nil {
		h.logger.Error("failed to get timeline 1 for comparison",
			zap.String("traceId", req.TraceID1.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get first timeline")
	}

	// Fetch second timeline
	timeline2, err := h.replayService.GetTimelineForTrace(c.Context(), projectID, req.TraceID2.String())
	if err != nil {
		h.logger.Error("failed to get timeline 2 for comparison",
			zap.String("traceId", req.TraceID2.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get second timeline")
	}

	// Build comparison
	comparison := TimelineComparison{
		TraceID1:    req.TraceID1,
		TraceID2:    req.TraceID2,
		Summary1:    timeline1.Summary,
		Summary2:    timeline2.Summary,
		Differences: h.compareTimelineEvents(timeline1.Events, timeline2.Events),
	}

	return c.JSON(comparison)
}

// compareTimelineEvents compares events from two timelines and returns differences
func (h *ReplayHandler) compareTimelineEvents(events1, events2 []domain.ReplayEvent) []TimelineDifference {
	differences := []TimelineDifference{}

	// Build maps for O(1) lookup by event type and title
	eventMap1 := make(map[string][]domain.ReplayEvent)
	eventMap2 := make(map[string][]domain.ReplayEvent)

	for _, e := range events1 {
		key := string(e.Type) + ":" + e.Title
		eventMap1[key] = append(eventMap1[key], e)
	}

	for _, e := range events2 {
		key := string(e.Type) + ":" + e.Title
		eventMap2[key] = append(eventMap2[key], e)
	}

	// Find events in timeline 1 that don't exist in timeline 2
	for key, eventsIn1 := range eventMap1 {
		eventsIn2, exists := eventMap2[key]
		if !exists {
			// All events of this type/title only exist in timeline 1
			for _, e := range eventsIn1 {
				eCopy := e
				differences = append(differences, TimelineDifference{
					Type:        "removed",
					Description: "Event present in first timeline but not in second: " + e.Title,
					Event1:      &eCopy,
					Event2:      nil,
				})
			}
		} else if len(eventsIn1) != len(eventsIn2) {
			// Different count of same event type
			differences = append(differences, TimelineDifference{
				Type:        "changed",
				Description: "Different number of '" + eventsIn1[0].Title + "' events: " +
					string(rune(len(eventsIn1))) + " vs " + string(rune(len(eventsIn2))),
				Event1:      nil,
				Event2:      nil,
			})
		} else {
			// Same count - compare individual events by index
			for i := range eventsIn1 {
				e1 := eventsIn1[i]
				e2 := eventsIn2[i]

				// Check for status differences
				if e1.Status != e2.Status {
					e1Copy, e2Copy := e1, e2
					differences = append(differences, TimelineDifference{
						Type:        "changed",
						Description: "Status changed for '" + e1.Title + "': " + e1.Status + " → " + e2.Status,
						Event1:      &e1Copy,
						Event2:      &e2Copy,
					})
				}

				// Check for significant duration differences (>20%)
				if e1.Duration > 0 && e2.Duration > 0 {
					diff := float64(e2.Duration - e1.Duration)
					percentChange := (diff / float64(e1.Duration)) * 100
					if percentChange > 20 || percentChange < -20 {
						e1Copy, e2Copy := e1, e2
						differences = append(differences, TimelineDifference{
							Type:        "changed",
							Description: "Duration significantly changed for '" + e1.Title + "'",
							Event1:      &e1Copy,
							Event2:      &e2Copy,
						})
					}
				}
			}
		}
	}

	// Find events in timeline 2 that don't exist in timeline 1
	for key, eventsIn2 := range eventMap2 {
		_, exists := eventMap1[key]
		if !exists {
			// These events only exist in timeline 2
			for _, e := range eventsIn2 {
				eCopy := e
				differences = append(differences, TimelineDifference{
					Type:        "added",
					Description: "Event present in second timeline but not in first: " + e.Title,
					Event1:      nil,
					Event2:      &eCopy,
				})
			}
		}
	}

	// Add summary-level differences
	// Compare total costs
	// (Summaries are already included in the response for the caller to analyze)

	return differences
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
