package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// StreamingHandler handles real-time streaming endpoints
type StreamingHandler struct {
	streamingService *service.StreamingService
	logger           *zap.Logger
}

// NewStreamingHandler creates a new streaming handler
func NewStreamingHandler(streamingService *service.StreamingService, logger *zap.Logger) *StreamingHandler {
	return &StreamingHandler{
		streamingService: streamingService,
		logger:           logger,
	}
}

// StreamTrace handles GET /api/public/traces/:traceId/stream - SSE per-trace stream
func (h *StreamingHandler) StreamTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Project ID not found",
		})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid trace ID",
		})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no")

	filter := domain.StreamSubscription{
		TraceID:    &traceID,
		FollowMode: c.Query("follow", "true") == "true",
	}

	sub := h.streamingService.SubscribeToTrace(c.Context(), traceID, projectID, filter)

	h.logger.Info("trace stream client connected",
		zap.String("traceId", traceID.String()),
		zap.String("subscriberId", sub.ID),
	)

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Send initial metrics snapshot
		metrics := h.streamingService.GetLiveMetrics(traceID)
		if metrics != nil {
			data, _ := json.Marshal(metrics)
			fmt.Fprintf(w, "event: metrics\ndata: %s\n\n", data)
			w.Flush()
		}

		// Send recent activities as backfill
		activities := h.streamingService.GetRecentActivities(traceID, 50)
		for _, act := range activities {
			data, _ := json.Marshal(act)
			fmt.Fprintf(w, "event: activity\ndata: %s\n\n", data)
		}
		w.Flush()

		heartbeat := time.NewTicker(15 * time.Second)
		metricsTicker := time.NewTicker(5 * time.Second)
		defer heartbeat.Stop()
		defer metricsTicker.Stop()

		for {
			select {
			case activity, ok := <-sub.Channel:
				if !ok {
					return
				}
				data, err := json.Marshal(activity)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: activity\ndata: %s\n\n", data)
				w.Flush()

			case <-metricsTicker.C:
				metrics := h.streamingService.GetLiveMetrics(traceID)
				if metrics != nil {
					data, _ := json.Marshal(metrics)
					fmt.Fprintf(w, "event: metrics\ndata: %s\n\n", data)
					w.Flush()
				}

			case <-heartbeat.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				w.Flush()

			case <-sub.Done:
				return

			case <-c.Context().Done():
				h.streamingService.UnsubscribeFromTrace(traceID, sub.ID)
				return
			}
		}
	}))

	return nil
}

// GetLiveMetrics handles GET /api/public/traces/:traceId/live-metrics
func (h *StreamingHandler) GetLiveMetrics(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	metrics := h.streamingService.GetLiveMetrics(traceID)
	if metrics == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No active stream for this trace"})
	}

	return c.JSON(metrics)
}

// GetRecentActivities handles GET /api/public/traces/:traceId/activities
func (h *StreamingHandler) GetRecentActivities(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	activities := h.streamingService.GetRecentActivities(traceID, limit)
	return c.JSON(fiber.Map{
		"activities": activities,
		"count":      len(activities),
	})
}

// RequestIntervention handles POST /api/public/traces/:traceId/intervene
func (h *StreamingHandler) RequestIntervention(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	var input struct {
		Action  domain.InterventionAction `json:"action"`
		Message string                    `json:"message"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	req := domain.InterventionRequest{
		ID:        uuid.New(),
		TraceID:   traceID,
		ProjectID: projectID,
		Action:    input.Action,
		Message:   input.Message,
		CreatedAt: time.Now(),
		Status:    "pending",
	}

	if err := h.streamingService.RequestIntervention(c.Context(), req); err != nil {
		h.logger.Error("failed to request intervention", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to request intervention"})
	}

	return c.Status(fiber.StatusCreated).JSON(req)
}

// GetActiveStreams handles GET /api/public/streams
func (h *StreamingHandler) GetActiveStreams(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	streams := h.streamingService.GetActiveStreams(projectID)
	return c.JSON(fiber.Map{
		"streams": streams,
		"count":   len(streams),
	})
}

// GetPendingInterventions handles GET /api/public/traces/:traceId/interventions
func (h *StreamingHandler) GetPendingInterventions(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	interventions := h.streamingService.GetPendingInterventions(traceID)
	return c.JSON(fiber.Map{
		"interventions": interventions,
		"count":         len(interventions),
	})
}

// AcknowledgeIntervention handles POST /api/public/traces/:traceId/interventions/:interventionId/ack
func (h *StreamingHandler) AcknowledgeIntervention(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	interventionID, err := uuid.Parse(c.Params("interventionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid intervention ID"})
	}

	if err := h.streamingService.AcknowledgeIntervention(traceID, interventionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to acknowledge"})
	}

	return c.JSON(fiber.Map{"status": "acknowledged"})
}
