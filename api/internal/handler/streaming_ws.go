package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// StreamingWSMessage represents a WebSocket message for trace streaming.
type StreamingWSMessage struct {
	Type      string    `json:"type"`
	TraceID   string    `json:"traceId"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// StreamingClient represents a connected WebSocket client for trace streaming.
type StreamingClient struct {
	conn         *websocket.Conn
	traceID      string
	send         chan []byte
	subscription domain.StreamSubscription
}

// StreamingHub manages WebSocket connections for trace streaming, supporting trace-specific rooms.
type StreamingHub struct {
	clients    map[*StreamingClient]bool
	traceRooms map[string]map[*StreamingClient]bool
	register   chan *StreamingClient
	unregister chan *StreamingClient
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewStreamingHub creates a new streaming hub.
func NewStreamingHub() *StreamingHub {
	return &StreamingHub{
		clients:    make(map[*StreamingClient]bool),
		traceRooms: make(map[string]map[*StreamingClient]bool),
		register:   make(chan *StreamingClient),
		unregister: make(chan *StreamingClient),
		broadcast:  make(chan []byte, 256),
	}
}

// Run starts the hub's main loop for managing client registrations.
func (h *StreamingHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if h.traceRooms[client.traceID] == nil {
				h.traceRooms[client.traceID] = make(map[*StreamingClient]bool)
			}
			h.traceRooms[client.traceID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if room, ok := h.traceRooms[client.traceID]; ok {
					delete(room, client)
					if len(room) == 0 {
						delete(h.traceRooms, client.traceID)
					}
				}
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToTrace sends a message to all clients subscribed to a specific trace.
func (h *StreamingHub) BroadcastToTrace(traceID string, msg StreamingWSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.traceRooms[traceID]
	if !ok {
		return
	}

	for client := range room {
		select {
		case client.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

// StreamingWSHandler manages WebSocket connections for real-time trace streaming.
type StreamingWSHandler struct {
	logger           *zap.Logger
	streamingService *service.StreamingService
	hub              *StreamingHub
}

// NewStreamingWSHandler creates a new WebSocket handler for trace streaming.
func NewStreamingWSHandler(
	logger *zap.Logger,
	streamingSvc *service.StreamingService,
) *StreamingWSHandler {
	h := &StreamingWSHandler{
		logger:           logger.Named("streaming-ws"),
		streamingService: streamingSvc,
		hub:              NewStreamingHub(),
	}
	go h.hub.Run()
	return h
}

// UpgradeCheck returns a middleware that checks if the request can be upgraded to WebSocket.
func (h *StreamingWSHandler) UpgradeCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocket handles WebSocket connections for real-time trace streaming.
func (h *StreamingWSHandler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		traceID := c.Params("traceId")
		followMode := c.Query("follow", "true") == "true"

		if traceID == "" {
			c.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "traceId required"))
			return
		}

		client := &StreamingClient{
			conn:    c,
			traceID: traceID,
			send:    make(chan []byte, 256),
			subscription: domain.StreamSubscription{
				FollowMode: followMode,
			},
		}

		h.hub.register <- client

		// Send connection confirmation
		h.sendToClient(client, StreamingWSMessage{
			Type:      "connected",
			TraceID:   traceID,
			Data:      map[string]any{"followMode": followMode},
			Timestamp: time.Now(),
		})

		// Send initial metrics snapshot
		traceUUID, err := uuid.Parse(traceID)
		if err == nil {
			if metrics := h.streamingService.GetLiveMetrics(traceUUID); metrics != nil {
				h.sendToClient(client, StreamingWSMessage{
					Type:      "metrics",
					TraceID:   traceID,
					Data:      metrics,
					Timestamp: time.Now(),
				})
			}
		}

		done := make(chan struct{})

		defer func() {
			close(done)
			h.hub.unregister <- client
		}()

		// Write pump
		go func() {
			for msg := range client.send {
				if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}()

		// Metrics ticker goroutine
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					h.hub.BroadcastToTrace(traceID, StreamingWSMessage{
						Type:    "metrics",
						TraceID: traceID,
						Data: domain.LiveMetrics{
							TraceID:         traceUUID,
							ActiveSpans:     rand.Intn(10) + 1,
							CompletedSpans:  rand.Intn(50) + 10,
							TotalTokens:     rand.Intn(10000) + 500,
							TotalCost:       float64(rand.Intn(100)) / 100.0,
							ErrorCount:      rand.Intn(3),
							ElapsedMs:       time.Since(time.Now().Add(-1 * time.Minute)).Milliseconds(),
							TokensPerSecond: float64(rand.Intn(100)) + 10.0,
							CostPerMinute:   float64(rand.Intn(10)) / 100.0,
							FilesModified:   rand.Intn(5),
							LastUpdated:     time.Now(),
						},
						Timestamp: time.Now(),
					})
				case <-done:
					return
				}
			}
		}()

		// Activity ticker goroutine
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			activityTypes := []domain.StreamEventType{
				domain.StreamEventTraceActivity,
				domain.StreamEventObservationStart,
				domain.StreamEventObservationEnd,
				domain.StreamEventFileChange,
				domain.StreamEventTerminalOutput,
			}
			for {
				select {
				case <-ticker.C:
					evtType := activityTypes[rand.Intn(len(activityTypes))]
					h.hub.BroadcastToTrace(traceID, StreamingWSMessage{
						Type:    "activity",
						TraceID: traceID,
						Data: domain.StreamActivity{
							ID:        uuid.New().String(),
							TraceID:   traceUUID,
							Type:      evtType,
							Title:     fmt.Sprintf("Mock %s event", evtType),
							Timestamp: time.Now(),
							Status:    "running",
						},
						Timestamp: time.Now(),
					})
				case <-done:
					return
				}
			}
		}()

		// Read pump
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}

			var inbound struct {
				Type    string         `json:"type"`
				Payload map[string]any `json:"payload,omitempty"`
			}
			if err := json.Unmarshal(msg, &inbound); err != nil {
				h.sendToClient(client, StreamingWSMessage{
					Type:      "error",
					TraceID:   traceID,
					Data:      map[string]any{"message": "invalid message format"},
					Timestamp: time.Now(),
				})
				continue
			}

			switch inbound.Type {
			case "subscribe":
				if tid, ok := inbound.Payload["traceId"].(string); ok {
					client.subscription.FollowMode = true
					h.sendToClient(client, StreamingWSMessage{
						Type:      "subscribed",
						TraceID:   tid,
						Data:      map[string]any{"traceId": tid, "followMode": true},
						Timestamp: time.Now(),
					})
				}

			case "unsubscribe":
				client.subscription.FollowMode = false
				h.sendToClient(client, StreamingWSMessage{
					Type:      "subscribed",
					TraceID:   traceID,
					Data:      map[string]any{"followMode": false},
					Timestamp: time.Now(),
				})

			case "ping":
				h.sendToClient(client, StreamingWSMessage{
					Type:      "pong",
					TraceID:   traceID,
					Timestamp: time.Now(),
				})
			}
		}
	})
}

// sendToClient sends a message to a specific client.
func (h *StreamingWSHandler) sendToClient(client *StreamingClient, msg StreamingWSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.send <- data:
	default:
	}
}
