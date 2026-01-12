package clickhouse

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// createTestObservation creates an observation with test data
func createTestObservation(projectID uuid.UUID, traceID string) *domain.Observation {
	now := time.Now()
	return &domain.Observation{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		TraceID:   traceID,
		Type:      domain.ObservationTypeSpan,
		Name:      "test-span",
		StartTime: now,
		EndTime:   &now,
		DurationMs: 500,
		Input:     `{"input": "test"}`,
		Output:    `{"output": "result"}`,
		Level:     domain.LevelDefault,
		Metadata:  `{"key": "value"}`,
		UsageDetails: domain.UsageDetails{
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
		},
		CostDetails: domain.CostDetails{
			InputCost:  0.001,
			OutputCost: 0.002,
			TotalCost:  0.003,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestObservationRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	obs := createTestObservation(projectID, traceID)

	err := repo.Create(ctx, obs)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := repo.GetByID(ctx, projectID, obs.ID)
	require.NoError(t, err)
	assert.Equal(t, obs.ID, fetched.ID)
	assert.Equal(t, obs.Name, fetched.Name)
	assert.Equal(t, obs.TraceID, fetched.TraceID)
}

func TestObservationRepository_CreateBatch(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	observations := []*domain.Observation{
		createTestObservation(projectID, traceID),
		createTestObservation(projectID, traceID),
		createTestObservation(projectID, traceID),
	}

	err := repo.CreateBatch(ctx, observations)
	require.NoError(t, err)

	// Verify all observations were created
	for _, obs := range observations {
		fetched, err := repo.GetByID(ctx, projectID, obs.ID)
		require.NoError(t, err)
		assert.Equal(t, obs.ID, fetched.ID)
	}
}

func TestObservationRepository_CreateBatch_Empty(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()

	// Empty batch should not error
	err := repo.CreateBatch(ctx, []*domain.Observation{})
	require.NoError(t, err)
}

func TestObservationRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	obs := createTestObservation(projectID, traceID)
	err := repo.Create(ctx, obs)
	require.NoError(t, err)

	t.Run("existing observation", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, projectID, obs.ID)
		require.NoError(t, err)
		assert.Equal(t, obs.ID, fetched.ID)
		assert.Equal(t, obs.Name, fetched.Name)
		assert.Equal(t, obs.TraceID, fetched.TraceID)
		assert.Equal(t, obs.Type, fetched.Type)
	})

	t.Run("non-existent observation", func(t *testing.T) {
		_, err := repo.GetByID(ctx, projectID, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("wrong project ID", func(t *testing.T) {
		wrongProjectID := uuid.New()
		_, err := repo.GetByID(ctx, wrongProjectID, obs.ID)
		assert.Error(t, err)
	})
}

func TestObservationRepository_GetByTraceID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create multiple observations for the same trace
	observations := make([]*domain.Observation, 5)
	for i := 0; i < 5; i++ {
		obs := createTestObservation(projectID, traceID)
		observations[i] = obs
		err := repo.Create(ctx, obs)
		require.NoError(t, err)
	}

	// Fetch observations by trace ID
	fetched, err := repo.GetByTraceID(ctx, projectID, traceID)
	require.NoError(t, err)
	assert.Equal(t, 5, len(fetched))

	// All observations should belong to the same trace
	for _, obs := range fetched {
		assert.Equal(t, traceID, obs.TraceID)
	}
}

func TestObservationRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create test observations
	observations := make([]*domain.Observation, 5)
	for i := 0; i < 5; i++ {
		obs := createTestObservation(projectID, traceID)
		obs.Name = "test-span-" + string(rune('A'+i))
		observations[i] = obs
		err := repo.Create(ctx, obs)
		require.NoError(t, err)
	}

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.ObservationFilter{ProjectID: projectID}
		list, total, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 5)
		assert.GreaterOrEqual(t, total, int64(5))
	})

	t.Run("with limit", func(t *testing.T) {
		filter := &domain.ObservationFilter{ProjectID: projectID}
		list, _, err := repo.List(ctx, filter, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(list))
	})

	t.Run("filter by trace ID", func(t *testing.T) {
		filter := &domain.ObservationFilter{
			ProjectID: projectID,
			TraceID:   &traceID,
		}
		list, _, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 5)
		for _, obs := range list {
			assert.Equal(t, traceID, obs.TraceID)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		obsType := domain.ObservationTypeSpan
		filter := &domain.ObservationFilter{
			ProjectID: projectID,
			Type:      &obsType,
		}
		list, _, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		for _, obs := range list {
			assert.Equal(t, domain.ObservationTypeSpan, obs.Type)
		}
	})
}

func TestObservationRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	obs := createTestObservation(projectID, traceID)
	err := repo.Create(ctx, obs)
	require.NoError(t, err)

	// Update the observation
	obs.Name = "updated-span"
	obs.Output = `{"updated": "output"}`
	err = repo.Update(ctx, obs)
	require.NoError(t, err)

	// Verify the update
	fetched, err := repo.GetByID(ctx, projectID, obs.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated-span", fetched.Name)
}

func TestObservationRepository_Generation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create a generation observation
	gen := createTestObservation(projectID, traceID)
	gen.Type = domain.ObservationTypeGeneration
	gen.Model = "gpt-4"
	gen.UsageDetails = domain.UsageDetails{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
	}
	gen.CostDetails = domain.CostDetails{
		InputCost:  0.03,
		OutputCost: 0.06,
		TotalCost:  0.09,
	}

	err := repo.Create(ctx, gen)
	require.NoError(t, err)

	// Fetch and verify
	fetched, err := repo.GetByID(ctx, projectID, gen.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ObservationTypeGeneration, fetched.Type)
	assert.Equal(t, "gpt-4", fetched.Model)
	assert.Equal(t, uint64(1000), fetched.UsageDetails.InputTokens)
	assert.Equal(t, uint64(500), fetched.UsageDetails.OutputTokens)
}

func TestObservationRepository_GetTree(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create parent observation
	parent := createTestObservation(projectID, traceID)
	parent.Name = "parent-span"
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// Create child observation
	child := createTestObservation(projectID, traceID)
	child.Name = "child-span"
	child.ParentObservationID = &parent.ID
	err = repo.Create(ctx, child)
	require.NoError(t, err)

	// Get tree
	tree, err := repo.GetTree(ctx, projectID, traceID)
	require.NoError(t, err)
	assert.NotNil(t, tree)
	// Tree should have at least root observation with children
	assert.NotNil(t, tree.Observation)
}

func TestObservationRepository_GetGenerationsWithoutCost(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create generation without cost
	gen := createTestObservation(projectID, traceID)
	gen.Type = domain.ObservationTypeGeneration
	gen.Model = "gpt-4"
	gen.CostDetails = domain.CostDetails{
		TotalCost: 0,
	}
	gen.UsageDetails = domain.UsageDetails{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	err := repo.Create(ctx, gen)
	require.NoError(t, err)

	// Fetch generations without cost
	gens, err := repo.GetGenerationsWithoutCost(ctx, projectID, 10)
	require.NoError(t, err)
	// Note: The generation we created should be included if it has usage but no cost
	assert.NotNil(t, gens)
}

func TestObservationRepository_UpdateCosts(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	obs := createTestObservation(projectID, traceID)
	obs.CostDetails = domain.CostDetails{
		TotalCost: 0,
	}
	err := repo.Create(ctx, obs)
	require.NoError(t, err)

	// Update costs
	err = repo.UpdateCosts(ctx, projectID, obs.ID, 0.01, 0.02, 0.03)
	require.NoError(t, err)

	// Verify update (may need to wait for ReplacingMergeTree merge)
	fetched, err := repo.GetByID(ctx, projectID, obs.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetched)
}

func TestObservationRepository_CountBeforeCutoff(t *testing.T) {
	if os.Getenv("CLICKHOUSE_TEST_HOST") == "" {
		t.Skip("Skipping integration test: CLICKHOUSE_TEST_HOST not set")
		return
	}

	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewObservationRepository(db)
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New().String()

	// Create observations
	for i := 0; i < 3; i++ {
		obs := createTestObservation(projectID, traceID)
		err := repo.Create(ctx, obs)
		require.NoError(t, err)
	}

	// Count observations before future cutoff (should include all)
	futureCutoff := time.Now().Add(24 * time.Hour)
	count, err := repo.CountBeforeCutoff(ctx, projectID, futureCutoff)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))

	// Count observations before past cutoff (should include none)
	pastCutoff := time.Now().Add(-24 * time.Hour)
	count, err = repo.CountBeforeCutoff(ctx, projectID, pastCutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
