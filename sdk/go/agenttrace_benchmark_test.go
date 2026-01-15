package agenttrace

import (
	"testing"
)

// BenchmarkClient_AddEvent benchmarks the addEvent function
func BenchmarkClient_AddEvent(b *testing.B) {
	client := New(Config{
		APIKey:       "benchmark-key",
		Host:         "http://localhost:8080",
		MaxQueueSize: 100000,
		FlushAt:      100000, // Prevent auto-flush
	})
	defer client.Shutdown()

	event := map[string]any{
		"type": "benchmark-event",
		"data": map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.addEvent(event)
	}
}

// BenchmarkClient_AddEvent_Parallel benchmarks concurrent addEvent calls
func BenchmarkClient_AddEvent_Parallel(b *testing.B) {
	client := New(Config{
		APIKey:       "benchmark-key",
		Host:         "http://localhost:8080",
		MaxQueueSize: 1000000,
		FlushAt:      1000000, // Prevent auto-flush
	})
	defer client.Shutdown()

	event := map[string]any{
		"type": "benchmark-event",
		"data": map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.addEvent(event)
		}
	})
}

// BenchmarkClient_New benchmarks client creation
func BenchmarkClient_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client := New(Config{
			APIKey: "benchmark-key",
			Host:   "http://localhost:8080",
		})
		client.Shutdown()
	}
}
