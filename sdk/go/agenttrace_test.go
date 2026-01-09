package agenttrace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("creates client with defaults", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		if client.config.APIKey != "test-api-key" {
			t.Errorf("expected APIKey to be 'test-api-key', got '%s'", client.config.APIKey)
		}
		if client.config.Host != "https://api.agenttrace.io" {
			t.Errorf("expected Host to be default, got '%s'", client.config.Host)
		}
		if !client.Enabled() {
			t.Error("expected client to be enabled by default")
		}
		if client.config.FlushAt != 20 {
			t.Errorf("expected FlushAt to be 20, got %d", client.config.FlushAt)
		}
		if client.config.FlushInterval != 5*time.Second {
			t.Errorf("expected FlushInterval to be 5s, got %v", client.config.FlushInterval)
		}
	})

	t.Run("creates client with custom config", func(t *testing.T) {
		enabled := false
		client := New(Config{
			APIKey:        "test-api-key",
			Host:          "https://custom.example.com",
			Enabled:       &enabled,
			FlushAt:       10,
			FlushInterval: time.Second,
		})
		defer client.Shutdown()

		if client.config.Host != "https://custom.example.com" {
			t.Errorf("expected Host to be custom, got '%s'", client.config.Host)
		}
		if client.Enabled() {
			t.Error("expected client to be disabled")
		}
		if client.config.FlushAt != 10 {
			t.Errorf("expected FlushAt to be 10, got %d", client.config.FlushAt)
		}
	})

	t.Run("sets global client", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		globalClient := GetGlobalClient()
		if globalClient != client {
			t.Error("expected global client to be set")
		}
	})
}

func TestClient_Trace(t *testing.T) {
	t.Run("creates trace with name", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		ctx := context.Background()
		trace := client.Trace(ctx, TraceOptions{
			Name: "test-trace",
		})

		if trace.name != "test-trace" {
			t.Errorf("expected name to be 'test-trace', got '%s'", trace.name)
		}
		if trace.id == "" {
			t.Error("expected trace ID to be generated")
		}
	})

	t.Run("creates trace with custom ID", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		ctx := context.Background()
		trace := client.Trace(ctx, TraceOptions{
			Name: "test-trace",
			ID:   "custom-id",
		})

		if trace.id != "custom-id" {
			t.Errorf("expected ID to be 'custom-id', got '%s'", trace.id)
		}
	})

	t.Run("creates trace with metadata", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		ctx := context.Background()
		trace := client.Trace(ctx, TraceOptions{
			Name:      "test-trace",
			UserID:    "user-123",
			SessionID: "session-456",
			Metadata:  map[string]any{"key": "value"},
			Tags:      []string{"tag1", "tag2"},
		})

		if trace.userID != "user-123" {
			t.Errorf("expected userID to be 'user-123', got '%s'", trace.userID)
		}
		if trace.sessionID != "session-456" {
			t.Errorf("expected sessionID to be 'session-456', got '%s'", trace.sessionID)
		}
		if trace.metadata["key"] != "value" {
			t.Error("expected metadata to be set")
		}
		if len(trace.tags) != 2 || trace.tags[0] != "tag1" {
			t.Error("expected tags to be set")
		}
	})
}

func TestTrace_Span(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("creates span", func(t *testing.T) {
		span := trace.Span(SpanOptions{
			Name: "test-span",
		})

		if span.name != "test-span" {
			t.Errorf("expected name to be 'test-span', got '%s'", span.name)
		}
		if span.traceID != trace.id {
			t.Error("expected span to have trace's ID")
		}
	})

	t.Run("creates span with parent", func(t *testing.T) {
		span1 := trace.Span(SpanOptions{Name: "parent-span"})
		span2 := trace.Span(SpanOptions{
			Name:                "child-span",
			ParentObservationID: span1.id,
		})

		if span2.parentObservationID != span1.id {
			t.Error("expected child span to have parent ID")
		}
	})
}

func TestTrace_Generation(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("creates generation", func(t *testing.T) {
		gen := trace.Generation(GenerationOptions{
			Name:  "test-generation",
			Model: "gpt-4",
		})

		if gen.name != "test-generation" {
			t.Errorf("expected name to be 'test-generation', got '%s'", gen.name)
		}
		if gen.model != "gpt-4" {
			t.Errorf("expected model to be 'gpt-4', got '%s'", gen.model)
		}
	})

	t.Run("creates generation with parameters", func(t *testing.T) {
		gen := trace.Generation(GenerationOptions{
			Name:            "test-generation",
			Model:           "gpt-4",
			ModelParameters: map[string]any{"temperature": 0.7},
		})

		if gen.modelParameters["temperature"] != 0.7 {
			t.Error("expected model parameters to be set")
		}
	})
}

func TestClient_Score(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/ingestion" {
			t.Errorf("expected /api/public/ingestion, got %s", r.URL.Path)
		}

		var payload map[string]any
		json.NewDecoder(r.Body).Decode(&payload)

		batch := payload["batch"].([]any)
		if len(batch) == 0 {
			t.Error("expected batch to have events")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{
		APIKey:  "test-api-key",
		Host:    server.URL,
		FlushAt: 1, // Flush immediately
	})
	defer client.Shutdown()

	client.Score(ScoreOptions{
		TraceID: "trace-123",
		Name:    "accuracy",
		Value:   0.95,
	})

	// Wait for flush
	time.Sleep(100 * time.Millisecond)
}

func TestClient_Flush(t *testing.T) {
	eventsReceived := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		json.NewDecoder(r.Body).Decode(&payload)
		batch := payload["batch"].([]any)
		eventsReceived += len(batch)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{
		APIKey:        "test-api-key",
		Host:          server.URL,
		FlushAt:       100, // High threshold to prevent auto-flush
		FlushInterval: time.Hour,
	})
	defer client.Shutdown()

	// Add events
	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{Name: "test"})
	trace.Span(SpanOptions{Name: "span1"})
	trace.Span(SpanOptions{Name: "span2"})

	// Manually flush
	client.Flush()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	if eventsReceived == 0 {
		t.Error("expected events to be sent after flush")
	}
}
