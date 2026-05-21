package agenttrace

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestClient_ConcurrentAddEvent tests for race conditions in concurrent event additions
func TestClient_ConcurrentAddEvent(t *testing.T) {
	t.Parallel()

	numGoroutines := 100
	eventsPerGoroutine := 10
	expectedEvents := numGoroutines * eventsPerGoroutine

	client := New(Config{
		APIKey:        "test-api-key",
		Host:          newRaceTestServer(t),
		MaxQueueSize:  1000,
		FlushAt:       expectedEvents + 1,
		FlushInterval: time.Hour,
	})
	defer client.Shutdown()

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				client.addEvent(map[string]any{
					"type": "test-event",
					"id":   fmt.Sprintf("event-%d-%d", id, j),
				})
			}
		}(i)
	}
	wg.Wait()

	// Verify no panic occurred and queue has expected events
	client.queueMu.Lock()
	queueLen := len(client.queue)
	client.queueMu.Unlock()

	if queueLen != expectedEvents {
		t.Errorf("expected %d events in queue, got %d", expectedEvents, queueLen)
	}
}

// TestClient_ConcurrentFlush tests for race conditions in concurrent flush operations
func TestClient_ConcurrentFlush(t *testing.T) {
	t.Parallel()

	client := New(Config{
		APIKey:        "test-api-key",
		Host:          newRaceTestServer(t),
		MaxQueueSize:  1000,
		FlushAt:       1000, // Prevent auto-flush
		FlushInterval: time.Hour,
	})
	defer client.Shutdown()

	// Add some events
	for i := 0; i < 50; i++ {
		client.addEvent(map[string]any{"id": i})
	}

	// Concurrent flushes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.Flush()
		}()
	}
	wg.Wait()
}

// TestClient_ConcurrentQueueOverflow tests queue overflow under concurrent load
func TestClient_ConcurrentQueueOverflow(t *testing.T) {
	t.Parallel()

	maxSize := 100
	errCount := 0
	var errMu sync.Mutex

	client := New(Config{
		APIKey:        "test-api-key",
		Host:          newRaceTestServer(t),
		MaxQueueSize:  maxSize,
		FlushAt:       maxSize + 100, // Prevent auto-flush
		FlushInterval: time.Hour,
		OnError: func(err error) {
			errMu.Lock()
			errCount++
			errMu.Unlock()
		},
	})
	defer client.Shutdown()

	var wg sync.WaitGroup
	numGoroutines := 50
	eventsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				client.addEvent(map[string]any{"id": fmt.Sprintf("%d-%d", id, j)})
			}
		}(i)
	}
	wg.Wait()

	// Queue should be at max size
	client.queueMu.Lock()
	queueLen := len(client.queue)
	client.queueMu.Unlock()

	if queueLen > maxSize {
		t.Errorf("queue exceeded max size: got %d, max %d", queueLen, maxSize)
	}
}

func newRaceTestServer(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server.URL
}
