package clickhouse

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestSessionRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	traceRepo := NewTraceRepository(db, zap.NewNop())
	sessionRepo := NewSessionRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	sessionID := "test-session-" + uuid.New().String()

	// Create traces with the session ID
	traces := make([]*domain.Trace, 3)
	for i := 0; i < 3; i++ {
		trace := createTestTrace(projectID)
		trace.SessionID = sessionID
		trace.UserID = "test-user"
		trace.TotalCost = 0.01
		trace.TotalTokens = 100
		traces[i] = trace
		err := traceRepo.Create(ctx, trace)
		require.NoError(t, err)
	}

	// Cleanup
	defer func() {
		for _, trace := range traces {
			_ = traceRepo.Delete(ctx, projectID, trace.ID)
		}
	}()

	t.Run("existing session", func(t *testing.T) {
		session, err := sessionRepo.GetByID(ctx, projectID, sessionID)
		require.NoError(t, err)
		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, projectID, session.ProjectID)
		assert.Equal(t, "test-user", session.UserID)
		assert.Equal(t, int64(3), session.TraceCount)
		assert.InDelta(t, 0.03, session.TotalCost, 0.001)
		assert.Equal(t, uint64(300), session.TotalTokens)
	})

	t.Run("non-existent session", func(t *testing.T) {
		_, err := sessionRepo.GetByID(ctx, projectID, "non-existent-session")
		assert.Error(t, err)
	})
}

func TestSessionRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	traceRepo := NewTraceRepository(db, zap.NewNop())
	sessionRepo := NewSessionRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	// Create traces for multiple sessions
	var allTraces []*domain.Trace
	sessionIDs := []string{
		"session-a-" + uuid.New().String(),
		"session-b-" + uuid.New().String(),
		"session-c-" + uuid.New().String(),
	}

	for _, sessionID := range sessionIDs {
		for i := 0; i < 2; i++ {
			trace := createTestTrace(projectID)
			trace.SessionID = sessionID
			trace.UserID = "test-user"
			allTraces = append(allTraces, trace)
			err := traceRepo.Create(ctx, trace)
			require.NoError(t, err)
		}
	}

	// Cleanup
	defer func() {
		for _, trace := range allTraces {
			_ = traceRepo.Delete(ctx, projectID, trace.ID)
		}
	}()

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.SessionFilter{ProjectID: projectID}
		list, err := sessionRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Sessions), 3)
	})

	t.Run("with limit", func(t *testing.T) {
		filter := &domain.SessionFilter{ProjectID: projectID}
		list, err := sessionRepo.List(ctx, filter, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(list.Sessions))
		assert.True(t, list.HasMore)
	})

	t.Run("filter by user ID", func(t *testing.T) {
		userID := "test-user"
		filter := &domain.SessionFilter{
			ProjectID: projectID,
			UserID:    &userID,
		}
		list, err := sessionRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		for _, session := range list.Sessions {
			assert.Equal(t, "test-user", session.UserID)
		}
	})

	t.Run("filter by time range", func(t *testing.T) {
		now := time.Now()
		past := now.Add(-24 * time.Hour)
		filter := &domain.SessionFilter{
			ProjectID: projectID,
			FromTime:  &past,
			ToTime:    &now,
		}
		list, err := sessionRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Sessions), 3)
	})
}

func TestSessionRepository_GetDistinctUserIDs(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	traceRepo := NewTraceRepository(db, zap.NewNop())
	sessionRepo := NewSessionRepository(db)
	ctx := context.Background()
	projectID := uuid.New()

	// Create traces with different user IDs
	var allTraces []*domain.Trace
	userIDs := []string{"user-a", "user-b", "user-c"}

	for i, userID := range userIDs {
		trace := createTestTrace(projectID)
		trace.SessionID = "session-" + uuid.New().String()
		trace.UserID = userID
		allTraces = append(allTraces, trace)
		err := traceRepo.Create(ctx, trace)
		require.NoError(t, err)
		_ = i
	}

	// Cleanup
	defer func() {
		for _, trace := range allTraces {
			_ = traceRepo.Delete(ctx, projectID, trace.ID)
		}
	}()

	distinctUserIDs, err := sessionRepo.GetDistinctUserIDs(ctx, projectID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(distinctUserIDs), 3)

	// Verify all test user IDs are present
	for _, expectedUserID := range userIDs {
		found := false
		for _, actualUserID := range distinctUserIDs {
			if actualUserID == expectedUserID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected user ID %s not found", expectedUserID)
	}
}

func TestSessionRepository_Upsert(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	sessionRepo := NewSessionRepository(db)
	ctx := context.Background()

	// Upsert is a no-op for sessions (they're aggregated from traces)
	session := &domain.Session{
		ID:        "test-session",
		ProjectID: uuid.New(),
	}

	err := sessionRepo.Upsert(ctx, session)
	require.NoError(t, err)
}
