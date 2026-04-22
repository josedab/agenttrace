package clickhouse

import (
	"fmt"
	"math"

	"github.com/google/uuid"
)

func nullableUUID(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return *value
}

func uint64ToInt64(value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("value %d exceeds int64", value)
	}
	return int64(value), nil
}
