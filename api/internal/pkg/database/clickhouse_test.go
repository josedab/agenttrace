package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClickHouseDBClose(t *testing.T) {
	t.Run("handles nil connection", func(t *testing.T) {
		db := &ClickHouseDB{Conn: nil}
		err := db.Close()
		assert.NoError(t, err)
	})
}

func TestTruncateSQLClickHouse(t *testing.T) {
	// truncateSQL is shared between postgres and clickhouse
	// Additional test cases for ClickHouse-style queries

	tests := []struct {
		name     string
		sql      string
		maxLen   int
		expected string
	}{
		{
			name:     "ClickHouse insert query truncated",
			sql:      "INSERT INTO traces (id, project_id, name) VALUES",
			maxLen:   30,
			expected: "INSERT INTO traces (id, projec...",
		},
		{
			name:     "ClickHouse select with array functions",
			sql:      "SELECT arrayJoin(groupArray(name)) FROM traces WHERE project_id = ?",
			maxLen:   40,
			expected: "SELECT arrayJoin(groupArray(name)) FROM ...",
		},
		{
			name:     "ClickHouse batch insert",
			sql:      "INSERT INTO observations (id, trace_id, type, name, input, output, start_time, end_time) VALUES",
			maxLen:   50,
			expected: "INSERT INTO observations (id, trace_id, type, name...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateSQL(tt.sql, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClickHouseDBNilOperations(t *testing.T) {
	// Test that operations on nil connection are handled properly
	t.Run("Close with nil returns no error", func(t *testing.T) {
		db := &ClickHouseDB{Conn: nil}
		err := db.Close()
		assert.NoError(t, err)
	})
}

func TestBatchInsertTracesEmpty(t *testing.T) {
	t.Run("empty traces slice returns nil", func(t *testing.T) {
		// Using nil conn since we're testing the empty case which returns early
		db := &ClickHouseDB{Conn: nil}
		err := db.BatchInsertTraces(nil, []map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("nil traces slice returns nil", func(t *testing.T) {
		db := &ClickHouseDB{Conn: nil}
		err := db.BatchInsertTraces(nil, nil)
		assert.NoError(t, err)
	})
}
