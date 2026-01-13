package clickhouse

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// createTestScore creates a score with test data
func createTestScore(projectID uuid.UUID, traceID string) *domain.Score {
	now := time.Now()
	value := 0.85
	return &domain.Score{
		ID:        uuid.New(),
		ProjectID: projectID,
		TraceID:   traceID,
		Name:      "test-score",
		Value:     &value,
		DataType:  domain.ScoreDataTypeNumeric,
		Source:    domain.ScoreSourceAPI,
		Comment:   "Test comment",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func strPtr(s string) *string {
	return &s
}

func TestScoreRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	score := createTestScore(projectID, traceID)

	err := repo.Create(ctx, score)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
	require.NoError(t, err)
	assert.Equal(t, score.ID, fetched.ID)
	assert.Equal(t, score.Name, fetched.Name)
	assert.Equal(t, *score.Value, *fetched.Value)
}

func TestScoreRepository_CreateBatch(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	scores := []*domain.Score{
		createTestScore(projectID, traceID),
		createTestScore(projectID, traceID),
		createTestScore(projectID, traceID),
	}

	err := repo.CreateBatch(ctx, scores)
	require.NoError(t, err)

	// Verify all scores were created
	for _, score := range scores {
		fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
		require.NoError(t, err)
		assert.Equal(t, score.ID, fetched.ID)
	}
}

func TestScoreRepository_CreateBatch_Empty(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()

	// Empty batch should not error
	err := repo.CreateBatch(ctx, []*domain.Score{})
	require.NoError(t, err)
}

func TestScoreRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	score := createTestScore(projectID, traceID)
	err := repo.Create(ctx, score)
	require.NoError(t, err)

	t.Run("existing score", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
		require.NoError(t, err)
		assert.Equal(t, score.ID, fetched.ID)
		assert.Equal(t, score.Name, fetched.Name)
		assert.Equal(t, score.TraceID, fetched.TraceID)
	})

	t.Run("non-existent score", func(t *testing.T) {
		_, err := repo.GetByID(ctx, projectID, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("wrong project ID", func(t *testing.T) {
		wrongProjectID := uuid.New()
		_, err := repo.GetByID(ctx, wrongProjectID, score.ID.String())
		assert.Error(t, err)
	})
}

func TestScoreRepository_GetByTraceID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create multiple scores for the same trace
	scores := make([]*domain.Score, 5)
	for i := 0; i < 5; i++ {
		score := createTestScore(projectID, traceID)
		score.Name = "score-" + string(rune('A'+i))
		scores[i] = score
		err := repo.Create(ctx, score)
		require.NoError(t, err)
	}

	// Fetch scores by trace ID
	fetched, err := repo.GetByTraceID(ctx, projectID, traceID)
	require.NoError(t, err)
	assert.Equal(t, 5, len(fetched))

	// All scores should belong to the same trace
	for _, score := range fetched {
		assert.Equal(t, traceID, score.TraceID)
	}
}

func TestScoreRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create test scores
	scores := make([]*domain.Score, 5)
	for i := 0; i < 5; i++ {
		score := createTestScore(projectID, traceID)
		score.Name = "score-" + string(rune('A'+i))
		scores[i] = score
		err := repo.Create(ctx, score)
		require.NoError(t, err)
	}

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.ScoreFilter{ProjectID: projectID}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Scores), 5)
	})

	t.Run("with limit", func(t *testing.T) {
		filter := &domain.ScoreFilter{ProjectID: projectID}
		list, err := repo.List(ctx, filter, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(list.Scores))
		assert.True(t, list.HasMore)
	})

	t.Run("filter by trace ID", func(t *testing.T) {
		filter := &domain.ScoreFilter{
			ProjectID: projectID,
			TraceID:   &traceID,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list.Scores), 5)
		for _, score := range list.Scores {
			assert.Equal(t, traceID, score.TraceID)
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		name := "score-A"
		filter := &domain.ScoreFilter{
			ProjectID: projectID,
			Name:      &name,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		for _, score := range list.Scores {
			assert.Equal(t, "score-A", score.Name)
		}
	})

	t.Run("filter by source", func(t *testing.T) {
		source := domain.ScoreSourceAPI
		filter := &domain.ScoreFilter{
			ProjectID: projectID,
			Source:    &source,
		}
		list, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		for _, score := range list.Scores {
			assert.Equal(t, domain.ScoreSourceAPI, score.Source)
		}
	})
}

func TestScoreRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	score := createTestScore(projectID, traceID)
	err := repo.Create(ctx, score)
	require.NoError(t, err)

	// Update the score
	newValue := 0.95
	score.Value = &newValue
	score.Comment = "Updated comment"
	err = repo.Update(ctx, score)
	require.NoError(t, err)

	// Verify the update
	fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 0.95, *fetched.Value)
	assert.Equal(t, "Updated comment", fetched.Comment)
}

func TestScoreRepository_CategoricalScore(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create a categorical score
	score := &domain.Score{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     traceID,
		Name:        "quality",
		StringValue: strPtr("excellent"),
		DataType:    domain.ScoreDataTypeCategorical,
		Source:      domain.ScoreSourceAnnotation,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, score)
	require.NoError(t, err)

	// Fetch and verify
	fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
	require.NoError(t, err)
	assert.Equal(t, domain.ScoreDataTypeCategorical, fetched.DataType)
	assert.Equal(t, "excellent", *fetched.StringValue)
	assert.Nil(t, fetched.Value)
}

func TestScoreRepository_BooleanScore(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create a boolean score (using numeric 1.0)
	value := 1.0
	score := &domain.Score{
		ID:        uuid.New(),
		ProjectID: projectID,
		TraceID:   traceID,
		Name:      "passed",
		Value:     &value,
		DataType:  domain.ScoreDataTypeBoolean,
		Source:    domain.ScoreSourceEval,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, score)
	require.NoError(t, err)

	// Fetch and verify
	fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
	require.NoError(t, err)
	assert.Equal(t, domain.ScoreDataTypeBoolean, fetched.DataType)
	assert.Equal(t, 1.0, *fetched.Value)
}

func TestScoreRepository_ScoreWithObservation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()
	obsID := uuid.New().String()

	// Create a score linked to an observation
	score := createTestScore(projectID, traceID)
	score.ObservationID = &obsID

	err := repo.Create(ctx, score)
	require.NoError(t, err)

	// Fetch and verify
	fetched, err := repo.GetByID(ctx, projectID, score.ID.String())
	require.NoError(t, err)
	assert.Equal(t, &obsID, fetched.ObservationID)

	// Fetch by observation ID
	obsScores, err := repo.GetByObservationID(ctx, projectID, obsID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(obsScores))
	assert.Equal(t, score.ID, obsScores[0].ID)
}

func TestScoreRepository_CountBeforeCutoff(t *testing.T) {
	if os.Getenv("CLICKHOUSE_TEST_HOST") == "" {
		t.Skip("Skipping integration test: CLICKHOUSE_TEST_HOST not set")
		return
	}

	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewScoreRepository(db, zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create scores
	for i := 0; i < 3; i++ {
		score := createTestScore(projectID, traceID)
		err := repo.Create(ctx, score)
		require.NoError(t, err)
	}

	// Count scores before future cutoff (should include all)
	futureCutoff := time.Now().Add(24 * time.Hour)
	count, err := repo.CountBeforeCutoff(ctx, projectID, futureCutoff)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))

	// Count scores before past cutoff (should include none)
	pastCutoff := time.Now().Add(-24 * time.Hour)
	count, err = repo.CountBeforeCutoff(ctx, projectID, pastCutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

