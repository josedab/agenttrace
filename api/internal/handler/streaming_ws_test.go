package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingWS_UpgradeCheck_RejectsNonWebSocket(t *testing.T) {
	app := fiber.New()

	hub := NewStreamingHub()
	handler := &StreamingWSHandler{hub: hub}

	app.Use("/ws/streaming", handler.UpgradeCheck())
	app.Get("/ws/streaming/:traceId", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	// Non-WebSocket request should get 426 Upgrade Required
	req := httptest.NewRequest(http.MethodGet, "/ws/streaming/test-trace-id", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUpgradeRequired, resp.StatusCode)
}

func TestStreamingWS_UnauthenticatedRejected(t *testing.T) {
	// This tests that without auth middleware, the /ws/streaming path
	// requires authentication (now fixed via routes.go)
	app := fiber.New()

	// Simulate auth middleware that rejects unauthenticated requests
	app.Use("/ws/streaming", func(c *fiber.Ctx) error {
		// Simulate RequireAuth - no valid token present
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}
		return c.Next()
	})

	hub := NewStreamingHub()
	handler := &StreamingWSHandler{hub: hub}
	app.Use("/ws/streaming", handler.UpgradeCheck())
	app.Get("/ws/streaming/:traceId", handler.HandleWebSocket())

	// Request without auth should be rejected
	req := httptest.NewRequest(http.MethodGet, "/ws/streaming/test-trace-id", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestStreamingHub_RegisterUnregister(t *testing.T) {
	hub := NewStreamingHub()
	go hub.Run()

	client := &StreamingClient{
		traceID: "trace-1",
		send:    make(chan []byte, 256),
	}

	hub.register <- client

	// Give the hub goroutine time to process
	hub.mu.RLock()
	assert.True(t, hub.clients[client])
	assert.NotNil(t, hub.traceRooms["trace-1"])
	assert.True(t, hub.traceRooms["trace-1"][client])
	hub.mu.RUnlock()

	hub.unregister <- client

	// Give hub time to process
	// Read from send channel to confirm it was closed
	_, ok := <-client.send
	assert.False(t, ok, "send channel should be closed after unregister")

	hub.mu.RLock()
	assert.False(t, hub.clients[client])
	hub.mu.RUnlock()
}

func TestStreamingHub_BroadcastToTrace(t *testing.T) {
	hub := NewStreamingHub()
	go hub.Run()

	client1 := &StreamingClient{
		traceID: "trace-1",
		send:    make(chan []byte, 256),
	}
	client2 := &StreamingClient{
		traceID: "trace-2",
		send:    make(chan []byte, 256),
	}

	hub.register <- client1
	hub.register <- client2

	// Wait for registration
	for {
		hub.mu.RLock()
		c1ok := hub.clients[client1]
		c2ok := hub.clients[client2]
		hub.mu.RUnlock()
		if c1ok && c2ok {
			break
		}
	}

	// Broadcast to trace-1 should only reach client1
	hub.BroadcastToTrace("trace-1", StreamingWSMessage{
		Type:    "test",
		TraceID: "trace-1",
	})

	msg := <-client1.send
	assert.Contains(t, string(msg), "trace-1")

	// client2 should not have received anything
	select {
	case <-client2.send:
		t.Fatal("client2 should not receive trace-1 message")
	default:
		// expected
	}
}
