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

// MockTerminalCommandRepository is a mock implementation of TerminalCommandRepository
type MockTerminalCommandRepository struct {
	mock.Mock
}

func (m *MockTerminalCommandRepository) Create(ctx context.Context, cmd *domain.TerminalCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockTerminalCommandRepository) CreateBatch(ctx context.Context, cmds []*domain.TerminalCommand) error {
	args := m.Called(ctx, cmds)
	return args.Error(0)
}

func (m *MockTerminalCommandRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TerminalCommand, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TerminalCommand), args.Error(1)
}

func (m *MockTerminalCommandRepository) List(ctx context.Context, filter *domain.TerminalCommandFilter, limit, offset int) (*domain.TerminalCommandList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TerminalCommandList), args.Error(1)
}

func (m *MockTerminalCommandRepository) GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.TerminalCommandStats, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TerminalCommandStats), args.Error(1)
}

// MockTraceRepoForTermCmd is a mock for trace repository
type MockTraceRepoForTermCmd struct {
	mock.Mock
}

func (m *MockTraceRepoForTermCmd) Create(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForTermCmd) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepoForTermCmd) Update(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForTermCmd) UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error {
	args := m.Called(ctx, projectID, traceID, inputCost, outputCost, totalCost)
	return args.Error(0)
}

func (m *MockTraceRepoForTermCmd) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForTermCmd) GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	args := m.Called(ctx, projectID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForTermCmd) List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TraceList), args.Error(1)
}

func (m *MockTraceRepoForTermCmd) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	args := m.Called(ctx, projectID, traceID, bookmarked)
	return args.Error(0)
}

func (m *MockTraceRepoForTermCmd) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	args := m.Called(ctx, projectID, traceID)
	return args.Error(0)
}

func TestNewTerminalCommandService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)

		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		assert.NotNil(t, svc)
		assert.Equal(t, termCmdRepo, svc.termCmdRepo)
		assert.Equal(t, traceRepo, svc.traceRepo)
	})
}

func TestTerminalCommandService_Log(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("logs terminal command successfully", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		termCmdRepo.On("Create", ctx, mock.AnythingOfType("*domain.TerminalCommand")).Return(nil)

		workDir := "/home/user/project"
		shell := "/bin/bash"
		stdout := "output text"
		exitCode := int32(0)
		input := &domain.TerminalCommandInput{
			TraceID:          traceID,
			Command:          "npm",
			Args:             []string{"test"},
			WorkingDirectory: &workDir,
			Shell:            &shell,
			ExitCode:         &exitCode,
			Stdout:           &stdout,
		}

		cmd, err := svc.Log(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, cmd)
		assert.Equal(t, projectID, cmd.ProjectID)
		assert.Equal(t, traceID, cmd.TraceID)
		assert.Equal(t, "npm", cmd.Command)
		assert.Equal(t, []string{"test"}, cmd.Args)
		assert.Equal(t, workDir, cmd.WorkingDirectory)
		assert.Equal(t, shell, cmd.Shell)
		assert.Equal(t, exitCode, cmd.ExitCode)
		assert.Equal(t, stdout, cmd.Stdout)
		assert.True(t, cmd.Success) // exit code 0 = success
		traceRepo.AssertExpectations(t)
		termCmdRepo.AssertExpectations(t)
	})

	t.Run("uses provided timestamps", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		termCmdRepo.On("Create", ctx, mock.AnythingOfType("*domain.TerminalCommand")).Return(nil)

		startedAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		completedAt := time.Date(2025, 1, 1, 12, 0, 5, 0, time.UTC)
		input := &domain.TerminalCommandInput{
			TraceID:     traceID,
			Command:     "sleep",
			Args:        []string{"5"},
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		}

		cmd, err := svc.Log(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, cmd)
		assert.Equal(t, startedAt, cmd.StartedAt)
		assert.Equal(t, completedAt, *cmd.CompletedAt)
		assert.Equal(t, uint32(5000), cmd.DurationMs)
	})

	t.Run("logs failed command", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		termCmdRepo.On("Create", ctx, mock.AnythingOfType("*domain.TerminalCommand")).Return(nil)

		exitCode := int32(1)
		stderr := "Error: command failed"
		input := &domain.TerminalCommandInput{
			TraceID:  traceID,
			Command:  "false",
			ExitCode: &exitCode,
			Stderr:   &stderr,
		}

		cmd, err := svc.Log(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, cmd)
		assert.False(t, cmd.Success) // exit code 1 = failure
		assert.Equal(t, exitCode, cmd.ExitCode)
		assert.Equal(t, stderr, cmd.Stderr)
	})

	t.Run("logs timed out command", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		termCmdRepo.On("Create", ctx, mock.AnythingOfType("*domain.TerminalCommand")).Return(nil)

		timedOut := true
		killed := true
		success := false
		input := &domain.TerminalCommandInput{
			TraceID:  traceID,
			Command:  "sleep",
			Args:     []string{"3600"},
			TimedOut: &timedOut,
			Killed:   &killed,
			Success:  &success,
		}

		cmd, err := svc.Log(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, cmd)
		assert.True(t, cmd.TimedOut)
		assert.True(t, cmd.Killed)
		assert.False(t, cmd.Success)
	})

	t.Run("returns error when trace not found", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		traceRepo.On("GetByID", ctx, projectID, traceID).Return(nil, errors.New("trace not found"))

		input := &domain.TerminalCommandInput{
			TraceID: traceID,
			Command: "ls",
		}

		cmd, err := svc.Log(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, cmd)
		assert.Contains(t, err.Error(), "failed to get trace")
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		termCmdRepo.On("Create", ctx, mock.AnythingOfType("*domain.TerminalCommand")).Return(errors.New("db error"))

		input := &domain.TerminalCommandInput{
			TraceID: traceID,
			Command: "ls",
		}

		cmd, err := svc.Log(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, cmd)
		assert.Contains(t, err.Error(), "failed to create terminal command")
	})
}

func TestTerminalCommandService_LogBatch(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("logs batch successfully", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		termCmdRepo.On("CreateBatch", ctx, mock.AnythingOfType("[]*domain.TerminalCommand")).Return(nil)

		inputs := []*domain.TerminalCommandInput{
			{TraceID: traceID, Command: "npm", Args: []string{"install"}},
			{TraceID: traceID, Command: "npm", Args: []string{"test"}},
		}

		cmds, err := svc.LogBatch(ctx, projectID, inputs)

		require.NoError(t, err)
		assert.Len(t, cmds, 2)
		assert.Equal(t, "npm", cmds[0].Command)
		assert.Equal(t, []string{"install"}, cmds[0].Args)
		assert.Equal(t, "npm", cmds[1].Command)
		assert.Equal(t, []string{"test"}, cmds[1].Args)
		traceRepo.AssertExpectations(t)
		termCmdRepo.AssertExpectations(t)
	})

	t.Run("returns empty slice for empty inputs", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		cmds, err := svc.LogBatch(ctx, projectID, []*domain.TerminalCommandInput{})

		require.NoError(t, err)
		assert.Empty(t, cmds)
	})

	t.Run("returns error when trace not found", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		traceRepo.On("GetByID", ctx, projectID, traceID).Return(nil, errors.New("trace not found"))

		inputs := []*domain.TerminalCommandInput{
			{TraceID: traceID, Command: "npm", Args: []string{"test"}},
		}

		cmds, err := svc.LogBatch(ctx, projectID, inputs)

		require.Error(t, err)
		assert.Nil(t, cmds)
		assert.Contains(t, err.Error(), "failed to get trace")
	})
}

func TestTerminalCommandService_GetByTraceID(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("gets terminal commands by trace ID successfully", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		expected := []domain.TerminalCommand{
			{ID: uuid.New(), TraceID: traceID, Command: "npm"},
			{ID: uuid.New(), TraceID: traceID, Command: "git"},
		}
		termCmdRepo.On("GetByTraceID", ctx, projectID, traceID).Return(expected, nil)

		cmds, err := svc.GetByTraceID(ctx, projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, cmds, 2)
		termCmdRepo.AssertExpectations(t)
	})
}

func TestTerminalCommandService_List(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("lists terminal commands successfully", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		filter := &domain.TerminalCommandFilter{ProjectID: projectID}
		expected := &domain.TerminalCommandList{
			TerminalCommands: []domain.TerminalCommand{
				{ID: uuid.New(), Command: "npm"},
				{ID: uuid.New(), Command: "git"},
			},
			TotalCount: 2,
		}
		termCmdRepo.On("List", ctx, filter, 10, 0).Return(expected, nil)

		result, err := svc.List(ctx, filter, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.TerminalCommands, 2)
		assert.Equal(t, int64(2), result.TotalCount)
		termCmdRepo.AssertExpectations(t)
	})
}

func TestTerminalCommandService_GetStats(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("gets stats successfully", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		expected := &domain.TerminalCommandStats{
			TotalCommands:   100,
			SuccessCount:    90,
			FailureCount:    8,
			TimeoutCount:    2,
			AvgDurationMs:   500.5,
			TotalDurationMs: 50050,
		}
		termCmdRepo.On("GetStats", ctx, projectID, (*string)(nil)).Return(expected, nil)

		result, err := svc.GetStats(ctx, projectID, nil)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(100), result.TotalCommands)
		assert.Equal(t, uint64(90), result.SuccessCount)
		assert.Equal(t, uint64(8), result.FailureCount)
		assert.Equal(t, uint64(2), result.TimeoutCount)
		termCmdRepo.AssertExpectations(t)
	})

	t.Run("gets stats for specific trace", func(t *testing.T) {
		termCmdRepo := new(MockTerminalCommandRepository)
		traceRepo := new(MockTraceRepoForTermCmd)
		svc := NewTerminalCommandService(termCmdRepo, traceRepo)

		traceID := "trace-123"
		expected := &domain.TerminalCommandStats{
			TotalCommands: 10,
			SuccessCount:  9,
			FailureCount:  1,
		}
		termCmdRepo.On("GetStats", ctx, projectID, &traceID).Return(expected, nil)

		result, err := svc.GetStats(ctx, projectID, &traceID)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(10), result.TotalCommands)
		termCmdRepo.AssertExpectations(t)
	})
}
