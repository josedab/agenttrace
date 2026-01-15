package id

import (
	"sync"
	"testing"
)

// BenchmarkNewTraceID benchmarks trace ID generation
func BenchmarkNewTraceID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewTraceID()
	}
}

// BenchmarkNewTraceIDParallel benchmarks trace ID generation concurrently
func BenchmarkNewTraceIDParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewTraceID()
		}
	})
}

// BenchmarkNewSpanID benchmarks span ID generation
func BenchmarkNewSpanID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewSpanID()
	}
}

// BenchmarkNewSpanIDParallel benchmarks span ID generation concurrently
func BenchmarkNewSpanIDParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewSpanID()
		}
	})
}

// BenchmarkNewUUID benchmarks UUID generation
func BenchmarkNewUUID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewUUID()
	}
}

// BenchmarkNewUUIDParallel benchmarks UUID generation concurrently
func BenchmarkNewUUIDParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewUUID()
		}
	})
}

// BenchmarkValidateTraceID benchmarks trace ID validation
func BenchmarkValidateTraceID(b *testing.B) {
	id := NewTraceID()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateTraceID(id)
	}
}

// BenchmarkValidateSpanID benchmarks span ID validation
func BenchmarkValidateSpanID(b *testing.B) {
	id := NewSpanID()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSpanID(id)
	}
}

// BenchmarkValidateUUID benchmarks UUID validation
func BenchmarkValidateUUID(b *testing.B) {
	id := NewUUID()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateUUID(id)
	}
}

// BenchmarkNewAPIKeyPublic benchmarks public API key generation
func BenchmarkNewAPIKeyPublic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewAPIKeyPublic()
	}
}

// BenchmarkNewAPIKeySecret benchmarks secret API key generation
func BenchmarkNewAPIKeySecret(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewAPIKeySecret()
	}
}

// BenchmarkIDGenerationMixed simulates real-world ID generation patterns
func BenchmarkIDGenerationMixed(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewTraceID()
		// Typically spans outnumber traces
		_ = NewSpanID()
		_ = NewSpanID()
		_ = NewSpanID()
	}
}

// BenchmarkConcurrentIDGeneration benchmarks concurrent ID generation under load
func BenchmarkConcurrentIDGeneration(b *testing.B) {
	b.ReportAllocs()
	var wg sync.WaitGroup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			_ = NewTraceID()
		}()
		go func() {
			defer wg.Done()
			_ = NewSpanID()
		}()
		go func() {
			defer wg.Done()
			_ = NewSpanID()
		}()
		go func() {
			defer wg.Done()
			_ = NewUUID()
		}()
		wg.Wait()
	}
}
