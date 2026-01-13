package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	"github.com/agenttrace/agenttrace/api/internal/pkg/logger"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	_ = logger.Init(logger.Config{
		Level:  "error", // Only show errors in tests to reduce noise
		Format: "console",
	})
	os.Exit(m.Run())
}

func TestTruncateSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		maxLen   int
		expected string
	}{
		{
			name:     "short SQL unchanged",
			sql:      "SELECT * FROM users",
			maxLen:   100,
			expected: "SELECT * FROM users",
		},
		{
			name:     "exactly at max length",
			sql:      "SELECT * FROM users",
			maxLen:   19,
			expected: "SELECT * FROM users",
		},
		{
			name:     "truncated with ellipsis",
			sql:      "SELECT * FROM users WHERE id = 1",
			maxLen:   20,
			expected: "SELECT * FROM users ...",
		},
		{
			name:     "empty string",
			sql:      "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "max length of 0",
			sql:      "SELECT",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "very long query",
			sql:      "SELECT id, name, email, created_at, updated_at, deleted_at FROM users WHERE organization_id = $1 AND status = 'active' ORDER BY created_at DESC LIMIT 100",
			maxLen:   50,
			expected: "SELECT id, name, email, created_at, updated_at, de...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateSQL(tt.sql, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewQueryTracer(t *testing.T) {
	t.Run("creates tracer with debug disabled", func(t *testing.T) {
		tracer := newQueryTracer(false)
		assert.NotNil(t, tracer)
		assert.False(t, tracer.enableDebug)
		assert.NotNil(t, tracer.metrics)
	})

	t.Run("creates tracer with debug enabled", func(t *testing.T) {
		tracer := newQueryTracer(true)
		assert.NotNil(t, tracer)
		assert.True(t, tracer.enableDebug)
		assert.NotNil(t, tracer.metrics)
	})
}

func TestQueryTracerGetMetrics(t *testing.T) {
	t.Run("returns copy of metrics", func(t *testing.T) {
		tracer := newQueryTracer(false)
		tracer.metrics.TotalQueries = 10
		tracer.metrics.SlowQueries = 2
		tracer.metrics.FailedQueries = 1
		tracer.metrics.TotalDurationMs = 500

		metrics := tracer.GetMetrics()

		assert.Equal(t, int64(10), metrics.TotalQueries)
		assert.Equal(t, int64(2), metrics.SlowQueries)
		assert.Equal(t, int64(1), metrics.FailedQueries)
		assert.Equal(t, int64(500), metrics.TotalDurationMs)
	})

	t.Run("initial metrics are zero", func(t *testing.T) {
		tracer := newQueryTracer(false)
		metrics := tracer.GetMetrics()

		assert.Equal(t, int64(0), metrics.TotalQueries)
		assert.Equal(t, int64(0), metrics.SlowQueries)
		assert.Equal(t, int64(0), metrics.FailedQueries)
		assert.Equal(t, int64(0), metrics.TotalDurationMs)
	})
}

func TestQueryTracerTraceQueryStart(t *testing.T) {
	t.Run("adds start time to context", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		data := pgx.TraceQueryStartData{
			SQL:  "SELECT 1",
			Args: []interface{}{},
		}

		newCtx := tracer.TraceQueryStart(ctx, nil, data)

		// Verify start time is in context
		start, ok := newCtx.Value(queryStartKey{}).(time.Time)
		assert.True(t, ok)
		assert.False(t, start.IsZero())
	})

	t.Run("adds SQL to context", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		data := pgx.TraceQueryStartData{
			SQL:  "SELECT * FROM users WHERE id = $1",
			Args: []interface{}{1},
		}

		newCtx := tracer.TraceQueryStart(ctx, nil, data)

		sql, ok := newCtx.Value(querySQLKey{}).(string)
		assert.True(t, ok)
		assert.Equal(t, "SELECT * FROM users WHERE id = $1", sql)
	})

	t.Run("adds arg count to context", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		data := pgx.TraceQueryStartData{
			SQL:  "INSERT INTO users (name, email) VALUES ($1, $2)",
			Args: []interface{}{"test", "test@example.com"},
		}

		newCtx := tracer.TraceQueryStart(ctx, nil, data)

		argCount, ok := newCtx.Value(queryArgsKey{}).(int)
		assert.True(t, ok)
		assert.Equal(t, 2, argCount)
	})
}

func TestQueryTracerTraceQueryEnd(t *testing.T) {
	t.Run("increments total queries on success", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		// Simulate query start
		ctx = context.WithValue(ctx, queryStartKey{}, time.Now())
		ctx = context.WithValue(ctx, querySQLKey{}, "SELECT 1")
		ctx = context.WithValue(ctx, queryArgsKey{}, 0)

		data := pgx.TraceQueryEndData{
			Err:        nil,
			CommandTag: pgconn.CommandTag{},
		}

		tracer.TraceQueryEnd(ctx, nil, data)

		assert.Equal(t, int64(1), tracer.metrics.TotalQueries)
		assert.Equal(t, int64(0), tracer.metrics.FailedQueries)
	})

	t.Run("increments failed queries on error", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		// Simulate query start
		ctx = context.WithValue(ctx, queryStartKey{}, time.Now())
		ctx = context.WithValue(ctx, querySQLKey{}, "SELECT 1")
		ctx = context.WithValue(ctx, queryArgsKey{}, 0)

		data := pgx.TraceQueryEndData{
			Err:        errors.New("connection refused"),
			CommandTag: pgconn.CommandTag{},
		}

		tracer.TraceQueryEnd(ctx, nil, data)

		assert.Equal(t, int64(1), tracer.metrics.TotalQueries)
		assert.Equal(t, int64(1), tracer.metrics.FailedQueries)
	})

	t.Run("handles missing start time in context", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		data := pgx.TraceQueryEndData{
			Err:        nil,
			CommandTag: pgconn.CommandTag{},
		}

		// Should not panic
		tracer.TraceQueryEnd(ctx, nil, data)

		// Metrics should not be updated
		assert.Equal(t, int64(0), tracer.metrics.TotalQueries)
	})

	t.Run("accumulates duration", func(t *testing.T) {
		tracer := newQueryTracer(false)
		ctx := context.Background()

		// Simulate query start with a past time
		past := time.Now().Add(-50 * time.Millisecond)
		ctx = context.WithValue(ctx, queryStartKey{}, past)
		ctx = context.WithValue(ctx, querySQLKey{}, "SELECT 1")
		ctx = context.WithValue(ctx, queryArgsKey{}, 0)

		data := pgx.TraceQueryEndData{
			Err:        nil,
			CommandTag: pgconn.CommandTag{},
		}

		tracer.TraceQueryEnd(ctx, nil, data)

		// Duration should be at least 50ms
		assert.GreaterOrEqual(t, tracer.metrics.TotalDurationMs, int64(50))
	})
}

func TestQueryMetrics(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		metrics := QueryMetrics{
			TotalQueries:    100,
			SlowQueries:     5,
			FailedQueries:   2,
			TotalDurationMs: 5000,
		}

		assert.Equal(t, int64(100), metrics.TotalQueries)
		assert.Equal(t, int64(5), metrics.SlowQueries)
		assert.Equal(t, int64(2), metrics.FailedQueries)
		assert.Equal(t, int64(5000), metrics.TotalDurationMs)
	})
}

func TestPostgresDBClose(t *testing.T) {
	t.Run("handles nil pool", func(t *testing.T) {
		db := &PostgresDB{Pool: nil}
		// Should not panic
		db.Close()
	})
}

func TestContextKeys(t *testing.T) {
	t.Run("context keys are distinct", func(t *testing.T) {
		ctx := context.Background()

		// Set different values with different keys
		ctx = context.WithValue(ctx, queryStartKey{}, time.Now())
		ctx = context.WithValue(ctx, querySQLKey{}, "SELECT 1")
		ctx = context.WithValue(ctx, queryArgsKey{}, 5)

		// Verify each key retrieves correct type
		_, startOk := ctx.Value(queryStartKey{}).(time.Time)
		_, sqlOk := ctx.Value(querySQLKey{}).(string)
		_, argsOk := ctx.Value(queryArgsKey{}).(int)

		assert.True(t, startOk)
		assert.True(t, sqlOk)
		assert.True(t, argsOk)
	})
}
