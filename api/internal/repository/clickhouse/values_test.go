package clickhouse

import (
	"math"
	"testing"
)

func TestUint64ToInt64(t *testing.T) {
	t.Parallel()

	value, err := uint64ToInt64(42)
	if err != nil {
		t.Fatalf("unexpected conversion error: %v", err)
	}
	if value != 42 {
		t.Fatalf("expected 42, got %d", value)
	}

	if _, err := uint64ToInt64(uint64(math.MaxInt64) + 1); err == nil {
		t.Fatal("expected overflow error")
	}
}
