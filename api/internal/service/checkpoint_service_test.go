package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockCheckpointRepository is a mock implementation of CheckpointRepository
type MockCheckpointRepository struct {
	mock.Mock
}

func (m *MockCheckpointRepository) Create(ctx context.Context, checkpoint *domain.Checkpoint) error {
	args := m.Called(ctx, checkpoint)
	return args.Error(0)
}

func (m *MockCheckpointRepository) GetByID(ctx context.Context, projectID, checkpointID uuid.UUID) (*domain.Checkpoint, error) {
	args := m.Called(ctx, projectID, checkpointID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Checkpoint), args.Error(1)
}

func (m *MockCheckpointRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Checkpoint, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Checkpoint), args.Error(1)
}

func (m *MockCheckpointRepository) List(ctx context.Context, filter *domain.CheckpointFilter, limit, offset int) (*domain.CheckpointList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CheckpointList), args.Error(1)
}

// MockTraceRepoForCheckpoint is a mock for trace repository
type MockTraceRepoForCheckpoint struct {
	mock.Mock
}

func (m *MockTraceRepoForCheckpoint) Create(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForCheckpoint) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepoForCheckpoint) Update(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForCheckpoint) UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error {
	args := m.Called(ctx, projectID, traceID, inputCost, outputCost, totalCost)
	return args.Error(0)
}

func (m *MockTraceRepoForCheckpoint) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForCheckpoint) GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	args := m.Called(ctx, projectID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForCheckpoint) List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TraceList), args.Error(1)
}

func (m *MockTraceRepoForCheckpoint) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	args := m.Called(ctx, projectID, traceID, bookmarked)
	return args.Error(0)
}

func (m *MockTraceRepoForCheckpoint) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	args := m.Called(ctx, projectID, traceID)
	return args.Error(0)
}

func TestNewCheckpointService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)

		svc := NewCheckpointService(checkpointRepo, traceRepo)

		assert.NotNil(t, svc)
		assert.Equal(t, checkpointRepo, svc.checkpointRepo)
		assert.Equal(t, traceRepo, svc.traceRepo)
	})
}

func TestCheckpointService_Create(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("creates checkpoint successfully", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		checkpointRepo.On("Create", ctx, mock.AnythingOfType("*domain.Checkpoint")).Return(nil)

		description := "Test checkpoint"
		input := &domain.CheckpointInput{
			TraceID:     traceID,
			Name:        "test-checkpoint",
			Description: &description,
			Type:        domain.CheckpointTypeManual,
		}

		checkpoint, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, checkpoint)
		assert.Equal(t, projectID, checkpoint.ProjectID)
		assert.Equal(t, traceID, checkpoint.TraceID)
		assert.Equal(t, "test-checkpoint", checkpoint.Name)
		assert.Equal(t, description, checkpoint.Description)
		assert.Equal(t, domain.CheckpointTypeManual, checkpoint.Type)
		traceRepo.AssertExpectations(t)
		checkpointRepo.AssertExpectations(t)
	})

	t.Run("uses default type when not specified", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		checkpointRepo.On("Create", ctx, mock.AnythingOfType("*domain.Checkpoint")).Return(nil)

		input := &domain.CheckpointInput{
			TraceID: traceID,
			Name:    "test-checkpoint",
		}

		checkpoint, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, checkpoint)
		assert.Equal(t, domain.CheckpointTypeManual, checkpoint.Type)
	})

	t.Run("returns error when trace not found", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		traceRepo.On("GetByID", ctx, projectID, traceID).Return(nil, errors.New("trace not found"))

		input := &domain.CheckpointInput{
			TraceID: traceID,
			Name:    "test-checkpoint",
		}

		checkpoint, err := svc.Create(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, checkpoint)
		assert.Contains(t, err.Error(), "failed to get trace")
		traceRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		checkpointRepo.On("Create", ctx, mock.AnythingOfType("*domain.Checkpoint")).Return(errors.New("db error"))

		input := &domain.CheckpointInput{
			TraceID: traceID,
			Name:    "test-checkpoint",
		}

		checkpoint, err := svc.Create(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, checkpoint)
		assert.Contains(t, err.Error(), "failed to create checkpoint")
	})
}

func TestCheckpointService_Get(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	checkpointID := uuid.New()

	t.Run("gets checkpoint successfully", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		expected := &domain.Checkpoint{
			ID:        checkpointID,
			ProjectID: projectID,
			Name:      "test-checkpoint",
		}
		checkpointRepo.On("GetByID", ctx, projectID, checkpointID).Return(expected, nil)

		checkpoint, err := svc.Get(ctx, projectID, checkpointID)

		require.NoError(t, err)
		require.NotNil(t, checkpoint)
		assert.Equal(t, expected, checkpoint)
		checkpointRepo.AssertExpectations(t)
	})

	t.Run("returns error when checkpoint not found", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		checkpointRepo.On("GetByID", ctx, projectID, checkpointID).Return(nil, errors.New("not found"))

		checkpoint, err := svc.Get(ctx, projectID, checkpointID)

		require.Error(t, err)
		assert.Nil(t, checkpoint)
		checkpointRepo.AssertExpectations(t)
	})
}

func TestCheckpointService_GetByTraceID(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("gets checkpoints by trace ID successfully", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		expected := []domain.Checkpoint{
			{ID: uuid.New(), TraceID: traceID, Name: "checkpoint-1"},
			{ID: uuid.New(), TraceID: traceID, Name: "checkpoint-2"},
		}
		checkpointRepo.On("GetByTraceID", ctx, projectID, traceID).Return(expected, nil)

		checkpoints, err := svc.GetByTraceID(ctx, projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, checkpoints, 2)
		checkpointRepo.AssertExpectations(t)
	})

	t.Run("returns empty slice when no checkpoints found", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		checkpointRepo.On("GetByTraceID", ctx, projectID, traceID).Return([]domain.Checkpoint{}, nil)

		checkpoints, err := svc.GetByTraceID(ctx, projectID, traceID)

		require.NoError(t, err)
		assert.Empty(t, checkpoints)
		checkpointRepo.AssertExpectations(t)
	})
}

func TestCheckpointService_List(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("lists checkpoints successfully", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		filter := &domain.CheckpointFilter{ProjectID: projectID}
		expected := &domain.CheckpointList{
			Checkpoints: []domain.Checkpoint{
				{ID: uuid.New(), Name: "checkpoint-1"},
				{ID: uuid.New(), Name: "checkpoint-2"},
			},
			TotalCount: 2,
		}
		checkpointRepo.On("List", ctx, filter, 10, 0).Return(expected, nil)

		result, err := svc.List(ctx, filter, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Checkpoints, 2)
		assert.Equal(t, int64(2), result.TotalCount)
		checkpointRepo.AssertExpectations(t)
	})
}

func TestCheckpointService_Restore(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	checkpointID := uuid.New()
	traceID := "trace-456"

	t.Run("restores checkpoint successfully", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		originalCheckpoint := &domain.Checkpoint{
			ID:            checkpointID,
			ProjectID:     projectID,
			TraceID:       "trace-123",
			Name:          "original-checkpoint",
			Type:          domain.CheckpointTypeAuto,
			GitCommitSha:  "abc123",
			GitBranch:     "main",
			FilesSnapshot: `{"file1.go": "content"}`,
			FilesChanged:  []string{"file1.go"},
			TotalFiles:    1,
			CreatedAt:     time.Now().Add(-time.Hour),
		}
		checkpointRepo.On("GetByID", ctx, projectID, checkpointID).Return(originalCheckpoint, nil)
		checkpointRepo.On("Create", ctx, mock.AnythingOfType("*domain.Checkpoint")).Return(nil)

		input := &domain.RestoreCheckpointInput{
			CheckpointID: checkpointID,
			TraceID:      traceID,
		}

		rollback, err := svc.Restore(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, rollback)
		assert.Equal(t, domain.CheckpointTypeRollback, rollback.Type)
		assert.Contains(t, rollback.Name, "Rollback from")
		assert.Equal(t, &originalCheckpoint.ID, rollback.RestoredFrom)
		assert.NotNil(t, rollback.RestoredAt)
		assert.Equal(t, originalCheckpoint.GitCommitSha, rollback.GitCommitSha)
		assert.Equal(t, originalCheckpoint.GitBranch, rollback.GitBranch)
		assert.Equal(t, originalCheckpoint.FilesSnapshot, rollback.FilesSnapshot)
		checkpointRepo.AssertExpectations(t)
	})

	t.Run("returns error when checkpoint not found", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		checkpointRepo.On("GetByID", ctx, projectID, checkpointID).Return(nil, errors.New("not found"))

		input := &domain.RestoreCheckpointInput{
			CheckpointID: checkpointID,
			TraceID:      traceID,
		}

		rollback, err := svc.Restore(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, rollback)
		assert.Contains(t, err.Error(), "failed to get checkpoint")
	})

	t.Run("returns error when create rollback fails", func(t *testing.T) {
		checkpointRepo := new(MockCheckpointRepository)
		traceRepo := new(MockTraceRepoForCheckpoint)
		svc := NewCheckpointService(checkpointRepo, traceRepo)

		originalCheckpoint := &domain.Checkpoint{
			ID:        checkpointID,
			ProjectID: projectID,
			Name:      "original-checkpoint",
		}
		checkpointRepo.On("GetByID", ctx, projectID, checkpointID).Return(originalCheckpoint, nil)
		checkpointRepo.On("Create", ctx, mock.AnythingOfType("*domain.Checkpoint")).Return(errors.New("db error"))

		input := &domain.RestoreCheckpointInput{
			CheckpointID: checkpointID,
			TraceID:      traceID,
		}

		rollback, err := svc.Restore(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, rollback)
		assert.Contains(t, err.Error(), "failed to create rollback checkpoint")
	})
}
