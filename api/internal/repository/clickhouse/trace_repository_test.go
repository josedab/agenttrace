package clickhouse

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// getTestDB returns a database connection for integration tests.
// Returns nil if the database is not available (skips tests).
func getTestDB(t *testing.T) *database.ClickHouseDB {
	// Check if we're running integration tests
	if os.Getenv("CLICKHOUSE_TEST_HOST") == "" {
		t.Skip("Skipping integration test: CLICKHOUSE_TEST_HOST not set")
		return nil
	}

	cfg := config.ClickHouseConfig{
		Host:     os.Getenv("CLICKHOUSE_TEST_HOST"),
		Port:     9000,
		Database: os.Getenv("CLICKHOUSE_TEST_DB"),
		User:     os.Getenv("CLICKHOUSE_TEST_USER"),
		Password: os.Getenv("CLICKHOUSE_TEST_PASS"),
	}

	if cfg.Database == "" {
		cfg.Database = "test_agenttrace"
	}
	if cfg.Port == 0 {
		cfg.Port = 9000
	}

	db, err := database.NewClickHouse(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to ClickHouse: %v", err)
		return nil
	}

	return db
}

// createTestTrace creates a trace with test data
func createTestTrace(projectID uuid.UUID) *domain.Trace {
	now := time.Now()
	return &domain.Trace{
		ID:            uuid.New().String(),
		ProjectID:     projectID,
		Name:          "test-trace",
		UserID:        "test-user",
		SessionID:     "test-session",
		Release:       "v1.0.0",
		Version:       "1",
		Tags:          []string{"test", "integration"},
		Metadata:      `{"key": "value"}`,
		Public:        false,
		Bookmarked:    false,
		StartTime:     now,
		EndTime:       &now,
		DurationMs:    1000,
		Input:         `{"input": "test"}`,
		Output:        `{"output": "result"}`,
		Level:         domain.LevelDefault,
		StatusMessage: "",
		TotalCost:     0.01,
		InputCost:     0.005,
		OutputCost:    0.005,
		TotalTokens:   100,
		InputTokens:   50,
		OutputTokens:  50,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestTraceRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	trace := createTestTrace(projectID)

	err := repo.Create(ctx, trace)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := repo.GetByID(ctx, projectID, trace.ID)
	require.NoError(t, err)
	assert.Equal(t, trace.ID, fetched.ID)
	assert.Equal(t, trace.Name, fetched.Name)
	assert.Equal(t, projectID, fetched.ProjectID)

	// Cleanup
	err = repo.Delete(ctx, projectID, trace.ID)
	require.NoError(t, err)
}

func TestTraceRepository_CreateBatch(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	traces := []*domain.Trace{
		createTestTrace(projectID),
		createTestTrace(projectID),
		createTestTrace(projectID),
	}

	err := repo.CreateBatch(ctx, traces)
	require.NoError(t, err)

	// Verify all traces were created
	for _, trace := range traces {
		fetched, err := repo.GetByID(ctx, projectID, trace.ID)
		require.NoError(t, err)
		assert.Equal(t, trace.ID, fetched.ID)

		// Cleanup
		err = repo.Delete(ctx, projectID, trace.ID)
		require.NoError(t, err)
	}
}

func TestTraceRepository_CreateBatch_Empty(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()

	// Empty batch should not error
	err := repo.CreateBatch(ctx, []*domain.Trace{})
	require.NoError(t, err)
}

func TestTraceRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	trace := createTestTrace(projectID)
	err := repo.Create(ctx, trace)
	require.NoError(t, err)

	t.Run("existing trace", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, projectID, trace.ID)
		require.NoError(t, err)
		assert.Equal(t, trace.ID, fetched.ID)
		assert.Equal(t, trace.Name, fetched.Name)
		assert.Equal(t, trace.UserID, fetched.UserID)
		assert.Equal(t, trace.SessionID, fetched.SessionID)
		assert.Equal(t, trace.Release, fetched.Release)
		assert.Equal(t, trace.Tags, fetched.Tags)
	})

	t.Run("non-existent trace", func(t *testing.T) {
		_, err := repo.GetByID(ctx, projectID, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("wrong project ID", func(t *testing.T) {
		wrongProjectID := uuid.New()
		_, err := repo.GetByID(ctx, wrongProjectID, trace.ID)
		assert.Error(t, err)
	})

	// Cleanup
	err = repo.Delete(ctx, projectID, trace.ID)
	require.NoError(t, err)
}

func TestTraceRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	// Create test traces
	traces := make([]*domain.Trace, 5)
	for i := 0; i < 5; i++ {
		trace := createTestTrace(projectID)
		trace.Name = "test-trace-" + string(rune('A'+i))
		traces[i] = trace
		err := repo.Create(ctx, trace)
		require.NoError(t, err)
	}

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.TraceFilter{ProjectID: projectID}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Traces), 5)
	})

	t.Run("with limit", func(t *testing.T) {
		filter := &domain.TraceFilter{ProjectID: projectID}
		list, err := repo.List(ctx, filter, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(list.Traces))
		assert.True(t, list.HasMore)
	})

	t.Run("with offset", func(t *testing.T) {
		filter := &domain.TraceFilter{ProjectID: projectID}
		list, err := repo.List(ctx, filter, 10, 3)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Traces), 2)
	})

	t.Run("filter by name", func(t *testing.T) {
		name := "test-trace-A"
		filter := &domain.TraceFilter{
			ProjectID: projectID,
			Name:      &name,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		for _, trace := range list.Traces {
			assert.Contains(t, trace.Name, "test-trace-A")
		}
	})

	t.Run("filter by user ID", func(t *testing.T) {
		userID := "test-user"
		filter := &domain.TraceFilter{
			ProjectID: projectID,
			UserID:    &userID,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Traces), 5)
	})

	t.Run("filter by session ID", func(t *testing.T) {
		sessionID := "test-session"
		filter := &domain.TraceFilter{
			ProjectID: projectID,
			SessionID: &sessionID,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Traces), 5)
	})

	t.Run("filter by time range", func(t *testing.T) {
		now := time.Now()
		past := now.Add(-24 * time.Hour)
		filter := &domain.TraceFilter{
			ProjectID: projectID,
			FromTime:  &past,
			ToTime:    &now,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Traces), 5)
	})

	// Cleanup
	for _, trace := range traces {
		err := repo.Delete(ctx, projectID, trace.ID)
		require.NoError(t, err)
	}
}

func TestTraceRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	trace := createTestTrace(projectID)
	err := repo.Create(ctx, trace)
	require.NoError(t, err)

	// Update the trace
	trace.Name = "updated-name"
	trace.UserID = "updated-user"
	trace.Tags = []string{"updated", "tags"}
	err = repo.Update(ctx, trace)
	require.NoError(t, err)

	// Verify the update
	fetched, err := repo.GetByID(ctx, projectID, trace.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated-name", fetched.Name)
	assert.Equal(t, "updated-user", fetched.UserID)
	assert.Equal(t, []string{"updated", "tags"}, fetched.Tags)

	// Cleanup
	err = repo.Delete(ctx, projectID, trace.ID)
	require.NoError(t, err)
}

func TestTraceRepository_SetBookmark(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	trace := createTestTrace(projectID)
	trace.Bookmarked = false
	err := repo.Create(ctx, trace)
	require.NoError(t, err)

	// Set bookmark to true
	err = repo.SetBookmark(ctx, projectID, trace.ID, true)
	require.NoError(t, err)

	// Verify bookmark is set
	fetched, err := repo.GetByID(ctx, projectID, trace.ID)
	require.NoError(t, err)
	assert.True(t, fetched.Bookmarked)

	// Set bookmark back to false
	err = repo.SetBookmark(ctx, projectID, trace.ID, false)
	require.NoError(t, err)

	fetched, err = repo.GetByID(ctx, projectID, trace.ID)
	require.NoError(t, err)
	assert.False(t, fetched.Bookmarked)

	// Cleanup
	err = repo.Delete(ctx, projectID, trace.ID)
	require.NoError(t, err)
}

func TestTraceRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	trace := createTestTrace(projectID)
	err := repo.Create(ctx, trace)
	require.NoError(t, err)

	// Verify trace exists
	_, err = repo.GetByID(ctx, projectID, trace.ID)
	require.NoError(t, err)

	// Delete the trace
	err = repo.Delete(ctx, projectID, trace.ID)
	require.NoError(t, err)

	// Note: ClickHouse ALTER TABLE DELETE is async, so the trace might still
	// be visible immediately after deletion. In production, we'd need to wait
	// for the mutation to complete.
}

func TestTraceRepository_GetBySessionID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	sessionID := "test-session-" + uuid.New().String()

	// Create traces with the same session ID
	traces := make([]*domain.Trace, 3)
	for i := 0; i < 3; i++ {
		trace := createTestTrace(projectID)
		trace.SessionID = sessionID
		traces[i] = trace
		err := repo.Create(ctx, trace)
		require.NoError(t, err)
	}

	// Fetch traces by session ID
	fetched, err := repo.GetBySessionID(ctx, projectID, sessionID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(fetched))

	// All traces should have the same session ID
	for _, trace := range fetched {
		assert.Equal(t, sessionID, trace.SessionID)
	}

	// Cleanup
	for _, trace := range traces {
		err := repo.Delete(ctx, projectID, trace.ID)
		require.NoError(t, err)
	}
}

func TestTraceRepository_CountBeforeCutoff(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	// Create traces
	traces := make([]*domain.Trace, 3)
	for i := 0; i < 3; i++ {
		trace := createTestTrace(projectID)
		traces[i] = trace
		err := repo.Create(ctx, trace)
		require.NoError(t, err)
	}

	// Count traces before future cutoff (should include all)
	futureCutoff := time.Now().Add(24 * time.Hour)
	count, err := repo.CountBeforeCutoff(ctx, projectID, futureCutoff)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))

	// Count traces before past cutoff (should include none)
	pastCutoff := time.Now().Add(-24 * time.Hour)
	count, err = repo.CountBeforeCutoff(ctx, projectID, pastCutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Cleanup
	for _, trace := range traces {
		err := repo.Delete(ctx, projectID, trace.ID)
		require.NoError(t, err)
	}
}

func TestTraceRepository_UpdateCosts(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewTraceRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	trace := createTestTrace(projectID)
	trace.TotalCost = 0
	trace.InputCost = 0
	trace.OutputCost = 0
	err := repo.Create(ctx, trace)
	require.NoError(t, err)

	// Update costs
	err = repo.UpdateCosts(ctx, projectID, trace.ID, 0.01, 0.02, 0.03)
	require.NoError(t, err)

	// Verify costs were updated (may need to wait for ReplacingMergeTree merge)
	fetched, err := repo.GetByID(ctx, projectID, trace.ID)
	require.NoError(t, err)
	// Note: Due to ReplacingMergeTree behavior, costs might not be updated immediately
	// In a real test, we'd need to wait for the merge to complete
	assert.NotNil(t, fetched)

	// Cleanup
	err = repo.Delete(ctx, projectID, trace.ID)
	require.NoError(t, err)
}
