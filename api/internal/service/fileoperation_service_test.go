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

// MockFileOperationRepository is a mock implementation of FileOperationRepository
type MockFileOperationRepository struct {
	mock.Mock
}

func (m *MockFileOperationRepository) Create(ctx context.Context, fileOp *domain.FileOperation) error {
	args := m.Called(ctx, fileOp)
	return args.Error(0)
}

func (m *MockFileOperationRepository) CreateBatch(ctx context.Context, fileOps []*domain.FileOperation) error {
	args := m.Called(ctx, fileOps)
	return args.Error(0)
}

func (m *MockFileOperationRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.FileOperation, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.FileOperation), args.Error(1)
}

func (m *MockFileOperationRepository) List(ctx context.Context, filter *domain.FileOperationFilter, limit, offset int) (*domain.FileOperationList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileOperationList), args.Error(1)
}

func (m *MockFileOperationRepository) GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.FileOperationStats, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileOperationStats), args.Error(1)
}

// MockTraceRepoForFileOp is a mock for trace repository
type MockTraceRepoForFileOp struct {
	mock.Mock
}

func (m *MockTraceRepoForFileOp) Create(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForFileOp) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepoForFileOp) Update(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForFileOp) UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error {
	args := m.Called(ctx, projectID, traceID, inputCost, outputCost, totalCost)
	return args.Error(0)
}

func (m *MockTraceRepoForFileOp) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForFileOp) GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	args := m.Called(ctx, projectID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForFileOp) List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TraceList), args.Error(1)
}

func (m *MockTraceRepoForFileOp) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	args := m.Called(ctx, projectID, traceID, bookmarked)
	return args.Error(0)
}

func (m *MockTraceRepoForFileOp) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	args := m.Called(ctx, projectID, traceID)
	return args.Error(0)
}

func TestNewFileOperationService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)

		svc := NewFileOperationService(fileOpRepo, traceRepo)

		assert.NotNil(t, svc)
		assert.Equal(t, fileOpRepo, svc.fileOpRepo)
		assert.Equal(t, traceRepo, svc.traceRepo)
	})
}

func TestFileOperationService_Track(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("tracks file operation successfully", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		fileOpRepo.On("Create", ctx, mock.AnythingOfType("*domain.FileOperation")).Return(nil)

		toolName := "edit_file"
		reason := "Adding new feature"
		linesAdded := uint32(10)
		linesRemoved := uint32(5)
		input := &domain.FileOperationInput{
			TraceID:      traceID,
			Operation:    domain.FileOperationUpdate,
			FilePath:     "/path/to/file.go",
			ToolName:     &toolName,
			Reason:       &reason,
			LinesAdded:   &linesAdded,
			LinesRemoved: &linesRemoved,
		}

		fileOp, err := svc.Track(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, fileOp)
		assert.Equal(t, projectID, fileOp.ProjectID)
		assert.Equal(t, traceID, fileOp.TraceID)
		assert.Equal(t, domain.FileOperationUpdate, fileOp.Operation)
		assert.Equal(t, "/path/to/file.go", fileOp.FilePath)
		assert.Equal(t, toolName, fileOp.ToolName)
		assert.Equal(t, reason, fileOp.Reason)
		assert.Equal(t, linesAdded, fileOp.LinesAdded)
		assert.Equal(t, linesRemoved, fileOp.LinesRemoved)
		assert.True(t, fileOp.Success)
		traceRepo.AssertExpectations(t)
		fileOpRepo.AssertExpectations(t)
	})

	t.Run("uses provided timestamps", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		fileOpRepo.On("Create", ctx, mock.AnythingOfType("*domain.FileOperation")).Return(nil)

		startedAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		completedAt := time.Date(2025, 1, 1, 12, 0, 1, 0, time.UTC)
		input := &domain.FileOperationInput{
			TraceID:     traceID,
			Operation:   domain.FileOperationCreate,
			FilePath:    "/path/to/file.go",
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		}

		fileOp, err := svc.Track(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, fileOp)
		assert.Equal(t, startedAt, fileOp.StartedAt)
		assert.Equal(t, completedAt, *fileOp.CompletedAt)
		assert.Equal(t, uint32(1000), fileOp.DurationMs)
	})

	t.Run("tracks failed operation", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		fileOpRepo.On("Create", ctx, mock.AnythingOfType("*domain.FileOperation")).Return(nil)

		success := false
		errorMsg := "Permission denied"
		input := &domain.FileOperationInput{
			TraceID:      traceID,
			Operation:    domain.FileOperationDelete,
			FilePath:     "/path/to/file.go",
			Success:      &success,
			ErrorMessage: &errorMsg,
		}

		fileOp, err := svc.Track(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, fileOp)
		assert.False(t, fileOp.Success)
		assert.Equal(t, errorMsg, fileOp.ErrorMessage)
	})

	t.Run("returns error when trace not found", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		traceRepo.On("GetByID", ctx, projectID, traceID).Return(nil, errors.New("trace not found"))

		input := &domain.FileOperationInput{
			TraceID:   traceID,
			Operation: domain.FileOperationCreate,
			FilePath:  "/path/to/file.go",
		}

		fileOp, err := svc.Track(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, fileOp)
		assert.Contains(t, err.Error(), "failed to get trace")
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		fileOpRepo.On("Create", ctx, mock.AnythingOfType("*domain.FileOperation")).Return(errors.New("db error"))

		input := &domain.FileOperationInput{
			TraceID:   traceID,
			Operation: domain.FileOperationCreate,
			FilePath:  "/path/to/file.go",
		}

		fileOp, err := svc.Track(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, fileOp)
		assert.Contains(t, err.Error(), "failed to create file operation")
	})
}

func TestFileOperationService_TrackBatch(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("tracks batch successfully", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		fileOpRepo.On("CreateBatch", ctx, mock.AnythingOfType("[]*domain.FileOperation")).Return(nil)

		inputs := []*domain.FileOperationInput{
			{TraceID: traceID, Operation: domain.FileOperationCreate, FilePath: "/file1.go"},
			{TraceID: traceID, Operation: domain.FileOperationUpdate, FilePath: "/file2.go"},
		}

		fileOps, err := svc.TrackBatch(ctx, projectID, inputs)

		require.NoError(t, err)
		assert.Len(t, fileOps, 2)
		assert.Equal(t, "/file1.go", fileOps[0].FilePath)
		assert.Equal(t, "/file2.go", fileOps[1].FilePath)
		traceRepo.AssertExpectations(t)
		fileOpRepo.AssertExpectations(t)
	})

	t.Run("returns empty slice for empty inputs", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		fileOps, err := svc.TrackBatch(ctx, projectID, []*domain.FileOperationInput{})

		require.NoError(t, err)
		assert.Empty(t, fileOps)
	})

	t.Run("returns error when trace not found", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		traceRepo.On("GetByID", ctx, projectID, traceID).Return(nil, errors.New("trace not found"))

		inputs := []*domain.FileOperationInput{
			{TraceID: traceID, Operation: domain.FileOperationCreate, FilePath: "/file1.go"},
		}

		fileOps, err := svc.TrackBatch(ctx, projectID, inputs)

		require.Error(t, err)
		assert.Nil(t, fileOps)
		assert.Contains(t, err.Error(), "failed to get trace")
	})
}

func TestFileOperationService_GetByTraceID(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("gets file operations by trace ID successfully", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		expected := []domain.FileOperation{
			{ID: uuid.New(), TraceID: traceID, FilePath: "/file1.go"},
			{ID: uuid.New(), TraceID: traceID, FilePath: "/file2.go"},
		}
		fileOpRepo.On("GetByTraceID", ctx, projectID, traceID).Return(expected, nil)

		fileOps, err := svc.GetByTraceID(ctx, projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, fileOps, 2)
		fileOpRepo.AssertExpectations(t)
	})
}

func TestFileOperationService_List(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("lists file operations successfully", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		filter := &domain.FileOperationFilter{ProjectID: projectID}
		expected := &domain.FileOperationList{
			FileOperations: []domain.FileOperation{
				{ID: uuid.New(), FilePath: "/file1.go"},
				{ID: uuid.New(), FilePath: "/file2.go"},
			},
			TotalCount: 2,
		}
		fileOpRepo.On("List", ctx, filter, 10, 0).Return(expected, nil)

		result, err := svc.List(ctx, filter, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.FileOperations, 2)
		assert.Equal(t, int64(2), result.TotalCount)
		fileOpRepo.AssertExpectations(t)
	})
}

func TestFileOperationService_GetStats(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("gets stats successfully", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		expected := &domain.FileOperationStats{
			TotalOperations: 100,
			CreateCount:     30,
			ReadCount:       40,
			UpdateCount:     20,
			DeleteCount:     10,
			SuccessCount:    95,
			FailureCount:    5,
		}
		fileOpRepo.On("GetStats", ctx, projectID, (*string)(nil)).Return(expected, nil)

		result, err := svc.GetStats(ctx, projectID, nil)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(100), result.TotalOperations)
		assert.Equal(t, uint64(95), result.SuccessCount)
		fileOpRepo.AssertExpectations(t)
	})

	t.Run("gets stats for specific trace", func(t *testing.T) {
		fileOpRepo := new(MockFileOperationRepository)
		traceRepo := new(MockTraceRepoForFileOp)
		svc := NewFileOperationService(fileOpRepo, traceRepo)

		traceID := "trace-123"
		expected := &domain.FileOperationStats{
			TotalOperations: 10,
			CreateCount:     3,
		}
		fileOpRepo.On("GetStats", ctx, projectID, &traceID).Return(expected, nil)

		result, err := svc.GetStats(ctx, projectID, &traceID)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(10), result.TotalOperations)
		fileOpRepo.AssertExpectations(t)
	})
}
