package handler

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CollaborationWSHandler manages WebSocket connections for real-time collaboration.
type CollaborationWSHandler struct {
	logger      *zap.Logger
	collabSvc   *service.CollaborationService
	hub         *CollaborationHub
}

// NewCollaborationWSHandler creates a new WebSocket handler for collaboration.
func NewCollaborationWSHandler(
	logger *zap.Logger,
	collabSvc *service.CollaborationService,
) *CollaborationWSHandler {
	h := &CollaborationWSHandler{
		logger:    logger.Named("collab-ws"),
		collabSvc: collabSvc,
		hub:       NewCollaborationHub(),
	}
	go h.hub.Run()
	return h
}

// UpgradeCheck returns a middleware that checks if the request can be upgraded to WebSocket.
func (h *CollaborationWSHandler) UpgradeCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			// Pass authenticated userID to WebSocket handler via locals
			if userID, ok := middleware.GetUserID(c); ok {
				c.Locals("wsUserID", userID)
			}
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocket handles WebSocket connections for trace collaboration.
func (h *CollaborationWSHandler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		traceID := c.Params("traceId")
		userName := c.Query("userName", "Anonymous")

		if traceID == "" {
			c.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "traceId required"))
			return
		}

		// Extract userID from authenticated context (set by auth middleware)
		uidVal := c.Locals("wsUserID")
		uid, ok := uidVal.(uuid.UUID)
		if !ok || uid == uuid.Nil {
			c.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "authentication required"))
			return
		}

		client := &CollabClient{
			conn:     c,
			traceID:  traceID,
			userID:   uid,
			userName: userName,
			send:     make(chan []byte, 256),
		}

		h.hub.Register <- client

		// Notify others about new user
		h.hub.BroadcastToTrace(traceID, CollabMessage{
			Type: "user_joined",
			Payload: map[string]any{
				"userId":   uid.String(),
				"userName": userName,
			},
			Timestamp: time.Now(),
		}, &uid)

		defer func() {
			h.hub.Unregister <- client
			h.hub.BroadcastToTrace(traceID, CollabMessage{
				Type: "user_left",
				Payload: map[string]any{
					"userId":   uid.String(),
					"userName": userName,
				},
				Timestamp: time.Now(),
			}, &uid)
		}()

		// Write pump
		go func() {
			for msg := range client.send {
				if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
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

			var inbound CollabMessage
			if err := json.Unmarshal(msg, &inbound); err != nil {
				continue
			}
			inbound.Timestamp = time.Now()

			switch inbound.Type {
			case "cursor_move":
				h.hub.BroadcastToTrace(traceID, inbound, &uid)

			case "annotation_add":
				content, _ := inbound.Payload["content"].(string)
				eventID, _ := inbound.Payload["eventId"].(string)
				annotation := &domain.TraceAnnotation{
					ID:        uuid.New(),
					TraceID:   traceID,
					EventID:   eventID,
					UserID:    uid,
					UserName:  userName,
					Content:   content,
					CreatedAt: time.Now(),
				}
				h.hub.BroadcastToTrace(traceID, CollabMessage{
					Type:      "annotation_added",
					Payload:   map[string]any{"annotation": annotation},
					Timestamp: time.Now(),
				}, nil)

			case "annotation_resolve":
				h.hub.BroadcastToTrace(traceID, inbound, nil)
			}
		}
	})
}

// CollabMessage represents a real-time collaboration message.
type CollabMessage struct {
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// CollabClient represents a connected WebSocket client.
type CollabClient struct {
	conn     *websocket.Conn
	traceID  string
	userID   uuid.UUID
	userName string
	send     chan []byte
}

// CollaborationHub manages all active WebSocket clients grouped by trace.
type CollaborationHub struct {
	clients    map[string]map[*CollabClient]bool // traceID -> clients
	Register   chan *CollabClient
	Unregister chan *CollabClient
	mu         sync.RWMutex
}

// NewCollaborationHub creates a new collaboration hub.
func NewCollaborationHub() *CollaborationHub {
	return &CollaborationHub{
		clients:    make(map[string]map[*CollabClient]bool),
		Register:   make(chan *CollabClient),
		Unregister: make(chan *CollabClient),
	}
}

// Run starts the hub's main loop for managing client registrations.
func (h *CollaborationHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.clients[client.traceID] == nil {
				h.clients[client.traceID] = make(map[*CollabClient]bool)
			}
			h.clients[client.traceID][client] = true
			h.mu.Unlock()

			// Send current presence to new client
			h.sendPresence(client)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.traceID]; ok {
				delete(clients, client)
				close(client.send)
				if len(clients) == 0 {
					delete(h.clients, client.traceID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastToTrace sends a message to all clients viewing a trace.
// If excludeUserID is non-nil, that user's clients are skipped.
func (h *CollaborationHub) BroadcastToTrace(traceID string, msg CollabMessage, excludeUserID *uuid.UUID) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[traceID]
	if !ok {
		return
	}

	for client := range clients {
		if excludeUserID != nil && client.userID == *excludeUserID {
			continue
		}
		select {
		case client.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

// GetPresence returns the list of users currently viewing a trace.
func (h *CollaborationHub) GetPresence(traceID string) []domain.UserPresence {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[traceID]
	if !ok {
		return nil
	}

	seen := make(map[uuid.UUID]bool)
	var presence []domain.UserPresence
	for client := range clients {
		if seen[client.userID] {
			continue
		}
		seen[client.userID] = true
		presence = append(presence, domain.UserPresence{
			UserID:   client.userID,
			UserName: client.userName,
			TraceID:  traceID,
			LastSeen: time.Now(),
		})
	}

	return presence
}

func (h *CollaborationHub) sendPresence(target *CollabClient) {
	presence := h.GetPresence(target.traceID)
	msg := CollabMessage{
		Type:      "presence",
		Payload:   map[string]any{"users": presence},
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case target.send <- data:
	default:
	}
}
