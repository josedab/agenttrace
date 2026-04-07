package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollaborationHub_RegisterAndUnregister(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	traceID := "trace-123"
	userID := uuid.New()

	client := &CollabClient{
		traceID:  traceID,
		userID:   userID,
		userName: "Alice",
		send:     make(chan []byte, 256),
	}

	// Register
	hub.Register <- client
	time.Sleep(50 * time.Millisecond) // allow goroutine to process

	presence := hub.GetPresence(traceID)
	require.Len(t, presence, 1)
	assert.Equal(t, userID, presence[0].UserID)
	assert.Equal(t, "Alice", presence[0].UserName)

	// Unregister
	hub.Unregister <- client
	time.Sleep(50 * time.Millisecond)

	presence = hub.GetPresence(traceID)
	assert.Len(t, presence, 0)
}

func TestCollaborationHub_MultipleClients(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	traceID := "trace-456"
	user1 := uuid.New()
	user2 := uuid.New()

	client1 := &CollabClient{
		traceID:  traceID,
		userID:   user1,
		userName: "Alice",
		send:     make(chan []byte, 256),
	}
	client2 := &CollabClient{
		traceID:  traceID,
		userID:   user2,
		userName: "Bob",
		send:     make(chan []byte, 256),
	}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(50 * time.Millisecond)

	presence := hub.GetPresence(traceID)
	assert.Len(t, presence, 2)

	// Unregister one
	hub.Unregister <- client1
	time.Sleep(50 * time.Millisecond)

	presence = hub.GetPresence(traceID)
	assert.Len(t, presence, 1)
	assert.Equal(t, user2, presence[0].UserID)
}

func TestCollaborationHub_BroadcastToTrace(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	traceID := "trace-789"
	user1 := uuid.New()
	user2 := uuid.New()

	client1 := &CollabClient{
		traceID:  traceID,
		userID:   user1,
		userName: "Alice",
		send:     make(chan []byte, 256),
	}
	client2 := &CollabClient{
		traceID:  traceID,
		userID:   user2,
		userName: "Bob",
		send:     make(chan []byte, 256),
	}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(50 * time.Millisecond)

	t.Run("broadcasts to all when no exclusion", func(t *testing.T) {
		msg := CollabMessage{
			Type:      "test_msg",
			Payload:   map[string]any{"data": "hello"},
			Timestamp: time.Now(),
		}
		hub.BroadcastToTrace(traceID, msg, nil)

		// Both clients should receive
		select {
		case data := <-client1.send:
			var received CollabMessage
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, "test_msg", received.Type)
		case <-time.After(time.Second):
			t.Fatal("client1 did not receive message")
		}

		select {
		case data := <-client2.send:
			var received CollabMessage
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, "test_msg", received.Type)
		case <-time.After(time.Second):
			t.Fatal("client2 did not receive message")
		}
	})

	t.Run("excludes specified user", func(t *testing.T) {
		msg := CollabMessage{
			Type:      "cursor_move",
			Timestamp: time.Now(),
		}
		hub.BroadcastToTrace(traceID, msg, &user1)

		// client1 should NOT receive (excluded)
		select {
		case <-client1.send:
			t.Fatal("client1 should not have received excluded message")
		case <-time.After(100 * time.Millisecond):
			// expected
		}

		// client2 should receive
		select {
		case data := <-client2.send:
			var received CollabMessage
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, "cursor_move", received.Type)
		case <-time.After(time.Second):
			t.Fatal("client2 did not receive message")
		}
	})

	// Cleanup
	hub.Unregister <- client1
	hub.Unregister <- client2
}

func TestCollaborationHub_BroadcastToNonexistentTrace(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	// Should not panic
	msg := CollabMessage{Type: "test", Timestamp: time.Now()}
	hub.BroadcastToTrace("nonexistent-trace", msg, nil)
}

func TestCollaborationHub_GetPresenceEmptyTrace(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	presence := hub.GetPresence("nonexistent")
	assert.Nil(t, presence)
}

func TestCollaborationHub_DuplicateUserPresence(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	traceID := "trace-dup"
	userID := uuid.New()

	// Same user, two connections (e.g., multiple tabs)
	client1 := &CollabClient{
		traceID:  traceID,
		userID:   userID,
		userName: "Alice",
		send:     make(chan []byte, 256),
	}
	client2 := &CollabClient{
		traceID:  traceID,
		userID:   userID,
		userName: "Alice",
		send:     make(chan []byte, 256),
	}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(50 * time.Millisecond)

	// Presence should deduplicate by userID
	presence := hub.GetPresence(traceID)
	assert.Len(t, presence, 1)

	// Both connections should receive broadcasts
	msg := CollabMessage{Type: "test", Timestamp: time.Now()}
	hub.BroadcastToTrace(traceID, msg, nil)

	select {
	case <-client1.send:
	case <-time.After(time.Second):
		t.Fatal("client1 did not receive")
	}
	select {
	case <-client2.send:
	case <-time.After(time.Second):
		t.Fatal("client2 did not receive")
	}

	// Unregister one: trace should still exist
	hub.Unregister <- client1
	time.Sleep(50 * time.Millisecond)

	presence = hub.GetPresence(traceID)
	assert.Len(t, presence, 1)

	hub.Unregister <- client2
	time.Sleep(50 * time.Millisecond)

	presence = hub.GetPresence(traceID)
	assert.Len(t, presence, 0)
}

func TestCollaborationHub_SendChannelClosedOnUnregister(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	client := &CollabClient{
		traceID:  "trace-close",
		userID:   uuid.New(),
		userName: "Alice",
		send:     make(chan []byte, 256),
	}

	hub.Register <- client
	time.Sleep(50 * time.Millisecond)

	hub.Unregister <- client
	time.Sleep(50 * time.Millisecond)

	// Verify the send channel is closed (write pump goroutine should exit)
	_, ok := <-client.send
	assert.False(t, ok, "send channel should be closed after unregister")
}

func TestCollabMessage_MarshalJSON(t *testing.T) {
	msg := CollabMessage{
		Type: "user_joined",
		Payload: map[string]any{
			"userId":   "abc-123",
			"userName": "Alice",
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded CollabMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "user_joined", decoded.Type)
	assert.Equal(t, "abc-123", decoded.Payload["userId"])
}

func TestCollaborationHub_IsolatedTraces(t *testing.T) {
	hub := NewCollaborationHub()
	go hub.Run()

	client1 := &CollabClient{
		traceID:  "trace-A",
		userID:   uuid.New(),
		userName: "Alice",
		send:     make(chan []byte, 256),
	}
	client2 := &CollabClient{
		traceID:  "trace-B",
		userID:   uuid.New(),
		userName: "Bob",
		send:     make(chan []byte, 256),
	}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(50 * time.Millisecond)

	// Broadcast to trace-A should not reach trace-B
	hub.BroadcastToTrace("trace-A", CollabMessage{Type: "test", Timestamp: time.Now()}, nil)

	select {
	case <-client1.send:
		// expected
	case <-time.After(time.Second):
		t.Fatal("client1 should receive")
	}

	select {
	case <-client2.send:
		t.Fatal("client2 should NOT receive message for different trace")
	case <-time.After(100 * time.Millisecond):
		// expected
	}

	hub.Unregister <- client1
	hub.Unregister <- client2
}
