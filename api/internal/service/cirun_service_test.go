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

// MockCIRunRepository is a mock implementation of CIRunRepository
type MockCIRunRepository struct {
	mock.Mock
}

func (m *MockCIRunRepository) Create(ctx context.Context, ciRun *domain.CIRun) error {
	args := m.Called(ctx, ciRun)
	return args.Error(0)
}

func (m *MockCIRunRepository) GetByID(ctx context.Context, projectID, ciRunID uuid.UUID) (*domain.CIRun, error) {
	args := m.Called(ctx, projectID, ciRunID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CIRun), args.Error(1)
}

func (m *MockCIRunRepository) GetByProviderRunID(ctx context.Context, projectID uuid.UUID, providerRunID string) (*domain.CIRun, error) {
	args := m.Called(ctx, projectID, providerRunID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CIRun), args.Error(1)
}

func (m *MockCIRunRepository) List(ctx context.Context, filter *domain.CIRunFilter, limit, offset int) (*domain.CIRunList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CIRunList), args.Error(1)
}

func (m *MockCIRunRepository) Update(ctx context.Context, ciRun *domain.CIRun) error {
	args := m.Called(ctx, ciRun)
	return args.Error(0)
}

func (m *MockCIRunRepository) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.CIRunStats, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CIRunStats), args.Error(1)
}

func TestNewCIRunService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		assert.NotNil(t, svc)
		assert.Equal(t, repo, svc.ciRunRepo)
	})
}

func TestCIRunService_Create(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("creates CI run successfully with minimal input", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("Create", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		input := &domain.CIRunInput{
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
		}

		ciRun, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, projectID, ciRun.ProjectID)
		assert.Equal(t, domain.CIProviderGitHubActions, ciRun.Provider)
		assert.Equal(t, "12345678", ciRun.ProviderRunID)
		assert.Equal(t, domain.CIRunStatusPending, ciRun.Status)
		assert.NotEmpty(t, ciRun.ID)
		repo.AssertExpectations(t)
	})

	t.Run("creates CI run with full input", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("Create", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		startedAt := time.Now().Add(-10 * time.Minute)
		status := domain.CIRunStatusRunning
		pipelineName := "build-and-test"
		jobName := "test"
		workflowName := "CI"
		gitCommitSha := "abc123def456"
		gitBranch := "feature/new-feature"
		gitRepoURL := "https://github.com/org/repo"
		gitRef := "refs/heads/feature/new-feature"
		runnerName := "ubuntu-latest"
		runnerOS := "linux"
		runnerArch := "amd64"
		triggeredBy := "push"
		triggerEvent := "push"
		providerRunURL := "https://github.com/org/repo/actions/runs/12345678"

		input := &domain.CIRunInput{
			Provider:       domain.CIProviderGitHubActions,
			ProviderRunID:  "12345678",
			ProviderRunURL: &providerRunURL,
			PipelineName:   &pipelineName,
			JobName:        &jobName,
			WorkflowName:   &workflowName,
			GitCommitSha:   &gitCommitSha,
			GitBranch:      &gitBranch,
			GitRepoURL:     &gitRepoURL,
			GitRef:         &gitRef,
			StartedAt:      &startedAt,
			Status:         &status,
			RunnerName:     &runnerName,
			RunnerOS:       &runnerOS,
			RunnerArch:     &runnerArch,
			TriggeredBy:    &triggeredBy,
			TriggerEvent:   &triggerEvent,
		}

		ciRun, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, pipelineName, ciRun.PipelineName)
		assert.Equal(t, jobName, ciRun.JobName)
		assert.Equal(t, workflowName, ciRun.WorkflowName)
		assert.Equal(t, gitCommitSha, ciRun.GitCommitSha)
		assert.Equal(t, gitBranch, ciRun.GitBranch)
		assert.Equal(t, gitRepoURL, ciRun.GitRepoURL)
		assert.Equal(t, gitRef, ciRun.GitRef)
		assert.Equal(t, startedAt, ciRun.StartedAt)
		assert.Equal(t, status, ciRun.Status)
		assert.Equal(t, runnerName, ciRun.RunnerName)
		assert.Equal(t, runnerOS, ciRun.RunnerOS)
		assert.Equal(t, runnerArch, ciRun.RunnerArch)
		assert.Equal(t, triggeredBy, ciRun.TriggeredBy)
		assert.Equal(t, triggerEvent, ciRun.TriggerEvent)
		assert.Equal(t, providerRunURL, ciRun.ProviderRunURL)
		repo.AssertExpectations(t)
	})

	t.Run("creates CI run with PR context", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("Create", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		prNumber := uint32(123)
		prTitle := "Add new feature"
		prSourceBranch := "feature/new-feature"
		prTargetBranch := "main"

		input := &domain.CIRunInput{
			Provider:       domain.CIProviderGitHubActions,
			ProviderRunID:  "12345678",
			PRNumber:       &prNumber,
			PRTitle:        &prTitle,
			PRSourceBranch: &prSourceBranch,
			PRTargetBranch: &prTargetBranch,
		}

		ciRun, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, prNumber, ciRun.PRNumber)
		assert.Equal(t, prTitle, ciRun.PRTitle)
		assert.Equal(t, prSourceBranch, ciRun.PRSourceBranch)
		assert.Equal(t, prTargetBranch, ciRun.PRTargetBranch)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("Create", ctx, mock.AnythingOfType("*domain.CIRun")).Return(errors.New("db error"))

		input := &domain.CIRunInput{
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
		}

		ciRun, err := svc.Create(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, ciRun)
		assert.Contains(t, err.Error(), "failed to create CI run")
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_Get(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	ciRunID := uuid.New()

	t.Run("gets CI run successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		expected := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(expected, nil)

		ciRun, err := svc.Get(ctx, projectID, ciRunID)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, expected, ciRun)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when CI run not found", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("GetByID", ctx, projectID, ciRunID).Return(nil, errors.New("not found"))

		ciRun, err := svc.Get(ctx, projectID, ciRunID)

		require.Error(t, err)
		assert.Nil(t, ciRun)
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_GetByProviderRunID(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	providerRunID := "12345678"

	t.Run("gets CI run by provider run ID successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		expected := &domain.CIRun{
			ID:            uuid.New(),
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: providerRunID,
			Status:        domain.CIRunStatusRunning,
		}
		repo.On("GetByProviderRunID", ctx, projectID, providerRunID).Return(expected, nil)

		ciRun, err := svc.GetByProviderRunID(ctx, projectID, providerRunID)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, expected, ciRun)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when CI run not found", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("GetByProviderRunID", ctx, projectID, providerRunID).Return(nil, errors.New("not found"))

		ciRun, err := svc.GetByProviderRunID(ctx, projectID, providerRunID)

		require.Error(t, err)
		assert.Nil(t, ciRun)
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_List(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("lists CI runs successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		filter := &domain.CIRunFilter{ProjectID: projectID}
		expected := &domain.CIRunList{
			CIRuns: []domain.CIRun{
				{ID: uuid.New(), ProviderRunID: "123", Status: domain.CIRunStatusSuccess},
				{ID: uuid.New(), ProviderRunID: "456", Status: domain.CIRunStatusRunning},
			},
			TotalCount: 2,
		}
		repo.On("List", ctx, filter, 10, 0).Return(expected, nil)

		result, err := svc.List(ctx, filter, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.CIRuns, 2)
		assert.Equal(t, int64(2), result.TotalCount)
		repo.AssertExpectations(t)
	})

	t.Run("lists with filters", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		status := domain.CIRunStatusSuccess
		branch := "main"
		filter := &domain.CIRunFilter{
			ProjectID: projectID,
			Status:    &status,
			GitBranch: &branch,
		}
		expected := &domain.CIRunList{
			CIRuns: []domain.CIRun{
				{ID: uuid.New(), ProviderRunID: "123", Status: domain.CIRunStatusSuccess, GitBranch: "main"},
			},
			TotalCount: 1,
		}
		repo.On("List", ctx, filter, 10, 0).Return(expected, nil)

		result, err := svc.List(ctx, filter, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.CIRuns, 1)
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_Update(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	ciRunID := uuid.New()

	t.Run("updates CI run status successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     time.Now().Add(-5 * time.Minute),
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		newStatus := domain.CIRunStatusSuccess
		conclusion := "success"
		input := &domain.CIRunUpdateInput{
			Status:     &newStatus,
			Conclusion: &conclusion,
		}

		ciRun, err := svc.Update(ctx, projectID, ciRunID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, domain.CIRunStatusSuccess, ciRun.Status)
		assert.Equal(t, "success", ciRun.Conclusion)
		repo.AssertExpectations(t)
	})

	t.Run("updates CI run with completion time", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		startTime := time.Now().Add(-10 * time.Minute)
		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     startTime,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		completedAt := time.Now()
		input := &domain.CIRunUpdateInput{
			CompletedAt: &completedAt,
		}

		ciRun, err := svc.Update(ctx, projectID, ciRunID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.NotNil(t, ciRun.CompletedAt)
		assert.True(t, ciRun.DurationMs > 0)
		repo.AssertExpectations(t)
	})

	t.Run("updates CI run with traces", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     time.Now().Add(-5 * time.Minute),
			TraceIDs:      []string{},
			TraceCount:    0,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		traceIDs := []string{"trace-1", "trace-2", "trace-3"}
		input := &domain.CIRunUpdateInput{
			TraceIDs: traceIDs,
		}

		ciRun, err := svc.Update(ctx, projectID, ciRunID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, traceIDs, ciRun.TraceIDs)
		assert.Equal(t, uint32(3), ciRun.TraceCount)
		repo.AssertExpectations(t)
	})

	t.Run("updates CI run with metrics", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     time.Now().Add(-5 * time.Minute),
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		totalCost := 0.05
		totalTokens := uint64(5000)
		totalObservations := uint64(10)
		input := &domain.CIRunUpdateInput{
			TotalCost:         &totalCost,
			TotalTokens:       &totalTokens,
			TotalObservations: &totalObservations,
		}

		ciRun, err := svc.Update(ctx, projectID, ciRunID, input)

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, totalCost, ciRun.TotalCost)
		assert.Equal(t, totalTokens, ciRun.TotalTokens)
		assert.Equal(t, totalObservations, ciRun.TotalObservations)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when CI run not found", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("GetByID", ctx, projectID, ciRunID).Return(nil, errors.New("not found"))

		newStatus := domain.CIRunStatusSuccess
		input := &domain.CIRunUpdateInput{
			Status: &newStatus,
		}

		ciRun, err := svc.Update(ctx, projectID, ciRunID, input)

		require.Error(t, err)
		assert.Nil(t, ciRun)
		assert.Contains(t, err.Error(), "failed to get CI run")
		repo.AssertExpectations(t)
	})

	t.Run("returns error when repository update fails", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     time.Now().Add(-5 * time.Minute),
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(errors.New("db error"))

		newStatus := domain.CIRunStatusSuccess
		input := &domain.CIRunUpdateInput{
			Status: &newStatus,
		}

		ciRun, err := svc.Update(ctx, projectID, ciRunID, input)

		require.Error(t, err)
		assert.Nil(t, ciRun)
		assert.Contains(t, err.Error(), "failed to update CI run")
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_AddTrace(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	ciRunID := uuid.New()

	t.Run("adds trace successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			TraceIDs:      []string{"trace-1"},
			TraceCount:    1,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.MatchedBy(func(run *domain.CIRun) bool {
			return len(run.TraceIDs) == 2 && run.TraceCount == 2
		})).Return(nil)

		err := svc.AddTrace(ctx, projectID, ciRunID, "trace-2")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("does not add duplicate trace", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			TraceIDs:      []string{"trace-1", "trace-2"},
			TraceCount:    2,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		// Update should NOT be called since trace already exists

		err := svc.AddTrace(ctx, projectID, ciRunID, "trace-1")

		require.NoError(t, err)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when CI run not found", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("GetByID", ctx, projectID, ciRunID).Return(nil, errors.New("not found"))

		err := svc.AddTrace(ctx, projectID, ciRunID, "trace-1")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get CI run")
		repo.AssertExpectations(t)
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			TraceIDs:      []string{},
			TraceCount:    0,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(errors.New("db error"))

		err := svc.AddTrace(ctx, projectID, ciRunID, "trace-1")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update CI run")
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_Complete(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	ciRunID := uuid.New()

	t.Run("completes CI run successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		startTime := time.Now().Add(-10 * time.Minute)
		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     startTime,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.MatchedBy(func(run *domain.CIRun) bool {
			return run.Status == domain.CIRunStatusSuccess &&
				run.Conclusion == "success" &&
				run.CompletedAt != nil &&
				run.DurationMs > 0
		})).Return(nil)

		ciRun, err := svc.Complete(ctx, projectID, ciRunID, domain.CIRunStatusSuccess, "success")

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, domain.CIRunStatusSuccess, ciRun.Status)
		assert.Equal(t, "success", ciRun.Conclusion)
		assert.NotNil(t, ciRun.CompletedAt)
		assert.True(t, ciRun.DurationMs > 0)
		repo.AssertExpectations(t)
	})

	t.Run("completes CI run with failure status", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		startTime := time.Now().Add(-5 * time.Minute)
		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     startTime,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(nil)

		ciRun, err := svc.Complete(ctx, projectID, ciRunID, domain.CIRunStatusFailure, "Tests failed")

		require.NoError(t, err)
		require.NotNil(t, ciRun)
		assert.Equal(t, domain.CIRunStatusFailure, ciRun.Status)
		assert.Equal(t, "Tests failed", ciRun.Conclusion)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when CI run not found", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("GetByID", ctx, projectID, ciRunID).Return(nil, errors.New("not found"))

		ciRun, err := svc.Complete(ctx, projectID, ciRunID, domain.CIRunStatusSuccess, "success")

		require.Error(t, err)
		assert.Nil(t, ciRun)
		assert.Contains(t, err.Error(), "failed to get CI run")
		repo.AssertExpectations(t)
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		startTime := time.Now().Add(-5 * time.Minute)
		existingRun := &domain.CIRun{
			ID:            ciRunID,
			ProjectID:     projectID,
			Provider:      domain.CIProviderGitHubActions,
			ProviderRunID: "12345678",
			Status:        domain.CIRunStatusRunning,
			StartedAt:     startTime,
		}
		repo.On("GetByID", ctx, projectID, ciRunID).Return(existingRun, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.CIRun")).Return(errors.New("db error"))

		ciRun, err := svc.Complete(ctx, projectID, ciRunID, domain.CIRunStatusSuccess, "success")

		require.Error(t, err)
		assert.Nil(t, ciRun)
		assert.Contains(t, err.Error(), "failed to update CI run")
		repo.AssertExpectations(t)
	})
}

func TestCIRunService_GetStats(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("gets stats successfully", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		expected := &domain.CIRunStats{
			TotalRuns:      100,
			SuccessCount:   80,
			FailureCount:   15,
			CancelledCount: 5,
			AvgDurationMs:  300000, // 5 minutes average
			TotalCost:      25.50,
			TotalTokens:    500000,
		}
		repo.On("GetStats", ctx, projectID).Return(expected, nil)

		stats, err := svc.GetStats(ctx, projectID)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, expected, stats)
		assert.Equal(t, uint64(100), stats.TotalRuns)
		assert.Equal(t, uint64(80), stats.SuccessCount)
		assert.Equal(t, uint64(15), stats.FailureCount)
		assert.Equal(t, uint64(5), stats.CancelledCount)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		repo := new(MockCIRunRepository)
		svc := NewCIRunService(repo)

		repo.On("GetStats", ctx, projectID).Return(nil, errors.New("db error"))

		stats, err := svc.GetStats(ctx, projectID)

		require.Error(t, err)
		assert.Nil(t, stats)
		repo.AssertExpectations(t)
	})
}
