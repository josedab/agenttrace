package handler

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// EventsHandler handles Server-Sent Events endpoints
type EventsHandler struct {
	realtimeService *service.RealtimeService
	logger          *zap.Logger
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(realtimeService *service.RealtimeService, logger *zap.Logger) *EventsHandler {
	return &EventsHandler{
		realtimeService: realtimeService,
		logger:          logger,
	}
}

// StreamEvents handles GET /v1/events/stream
func (h *EventsHandler) StreamEvents(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create subscriber
	sub := h.realtimeService.Subscribe(c.Context(), projectID)

	h.logger.Info("SSE client connected",
		zap.String("project_id", projectID.String()),
		zap.String("subscriber_id", sub.ID),
	)

	// Use Fiber's streaming
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Send initial connection event
		fmt.Fprintf(w, "event: connected\n")
		fmt.Fprintf(w, "data: {\"subscriberId\":\"%s\"}\n\n", sub.ID)
		w.Flush()

		// Send heartbeat every 30 seconds
		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case event, ok := <-sub.Channel:
				if !ok {
					// Channel closed
					return
				}

				// Format SSE event
				data, err := service.FormatSSE(event)
				if err != nil {
					h.logger.Error("failed to format SSE event", zap.Error(err))
					continue
				}

				fmt.Fprintf(w, "event: %s\n", event.Type)
				w.Write(data)
				w.Flush()

			case <-heartbeat.C:
				// Send heartbeat to keep connection alive
				fmt.Fprintf(w, ": heartbeat\n\n")
				w.Flush()

			case <-sub.Done:
				return

			case <-c.Context().Done():
				h.realtimeService.Unsubscribe(sub.ID)
				return
			}
		}
	}))

	return nil
}

// GetSubscribers handles GET /v1/events/subscribers
func (h *EventsHandler) GetSubscribers(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	count := h.realtimeService.GetSubscriberCount(projectID)

	return c.JSON(fiber.Map{
		"count": count,
	})
}

// RegisterRoutes registers event routes
func (h *EventsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/events/stream", h.StreamEvents)
	v1.Get("/events/subscribers", h.GetSubscribers)
}
