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

// MockGitLinkRepository is a mock implementation of GitLinkRepository
type MockGitLinkRepository struct {
	mock.Mock
}

func (m *MockGitLinkRepository) Create(ctx context.Context, gitLink *domain.GitLink) error {
	args := m.Called(ctx, gitLink)
	return args.Error(0)
}

func (m *MockGitLinkRepository) GetByID(ctx context.Context, projectID, gitLinkID uuid.UUID) (*domain.GitLink, error) {
	args := m.Called(ctx, projectID, gitLinkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitLink), args.Error(1)
}

func (m *MockGitLinkRepository) GetByCommitSha(ctx context.Context, projectID uuid.UUID, commitSha string) ([]domain.GitLink, error) {
	args := m.Called(ctx, projectID, commitSha)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.GitLink), args.Error(1)
}

func (m *MockGitLinkRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.GitLink, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.GitLink), args.Error(1)
}

func (m *MockGitLinkRepository) List(ctx context.Context, filter *domain.GitLinkFilter, limit, offset int) (*domain.GitLinkList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitLinkList), args.Error(1)
}

func (m *MockGitLinkRepository) GetTimeline(ctx context.Context, projectID uuid.UUID, branch string, limit int) (*domain.GitTimeline, error) {
	args := m.Called(ctx, projectID, branch, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitTimeline), args.Error(1)
}

// MockTraceRepoForGitLink is a mock for trace repository
type MockTraceRepoForGitLink struct {
	mock.Mock
}

func (m *MockTraceRepoForGitLink) Create(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForGitLink) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepoForGitLink) Update(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepoForGitLink) UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error {
	args := m.Called(ctx, projectID, traceID, inputCost, outputCost, totalCost)
	return args.Error(0)
}

func (m *MockTraceRepoForGitLink) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForGitLink) GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	args := m.Called(ctx, projectID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trace), args.Error(1)
}

func (m *MockTraceRepoForGitLink) List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TraceList), args.Error(1)
}

func (m *MockTraceRepoForGitLink) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	args := m.Called(ctx, projectID, traceID, bookmarked)
	return args.Error(0)
}

func (m *MockTraceRepoForGitLink) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	args := m.Called(ctx, projectID, traceID)
	return args.Error(0)
}

func TestNewGitLinkService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)

		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		assert.NotNil(t, svc)
		assert.Equal(t, gitLinkRepo, svc.gitLinkRepo)
		assert.Equal(t, traceRepo, svc.traceRepo)
	})
}

func TestGitLinkService_Create(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("creates git link successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		gitLinkRepo.On("Create", ctx, mock.AnythingOfType("*domain.GitLink")).Return(nil)

		commitMsg := "Initial commit"
		author := "John Doe"
		authorEmail := "john@example.com"
		input := &domain.GitLinkInput{
			TraceID:           traceID,
			CommitSha:         "abc123def456",
			Branch:            "main",
			RepoURL:           "https://github.com/test/repo",
			CommitMessage:     &commitMsg,
			CommitAuthor:      &author,
			CommitAuthorEmail: &authorEmail,
			FilesAdded:        []string{"file1.go"},
			FilesModified:     []string{"file2.go"},
			FilesDeleted:      []string{},
		}

		gitLink, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, gitLink)
		assert.Equal(t, projectID, gitLink.ProjectID)
		assert.Equal(t, traceID, gitLink.TraceID)
		assert.Equal(t, "abc123def456", gitLink.CommitSha)
		assert.Equal(t, "main", gitLink.Branch)
		assert.Equal(t, commitMsg, gitLink.CommitMessage)
		assert.Equal(t, domain.GitLinkTypeCurrent, gitLink.LinkType)
		assert.Equal(t, uint32(2), gitLink.FilesChangedCount)
		traceRepo.AssertExpectations(t)
		gitLinkRepo.AssertExpectations(t)
	})

	t.Run("uses provided link type", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		gitLinkRepo.On("Create", ctx, mock.AnythingOfType("*domain.GitLink")).Return(nil)

		input := &domain.GitLinkInput{
			TraceID:   traceID,
			CommitSha: "abc123",
			Branch:    "main",
			LinkType:  domain.GitLinkTypeStart,
		}

		gitLink, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, gitLink)
		assert.Equal(t, domain.GitLinkTypeStart, gitLink.LinkType)
	})

	t.Run("uses provided commit timestamp", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		gitLinkRepo.On("Create", ctx, mock.AnythingOfType("*domain.GitLink")).Return(nil)

		timestamp := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		input := &domain.GitLinkInput{
			TraceID:         traceID,
			CommitSha:       "abc123",
			Branch:          "main",
			CommitTimestamp: &timestamp,
		}

		gitLink, err := svc.Create(ctx, projectID, input)

		require.NoError(t, err)
		require.NotNil(t, gitLink)
		assert.Equal(t, timestamp, gitLink.CommitTimestamp)
	})

	t.Run("returns error when trace not found", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		traceRepo.On("GetByID", ctx, projectID, traceID).Return(nil, errors.New("trace not found"))

		input := &domain.GitLinkInput{
			TraceID:   traceID,
			CommitSha: "abc123",
			Branch:    "main",
		}

		gitLink, err := svc.Create(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, gitLink)
		assert.Contains(t, err.Error(), "failed to get trace")
		traceRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		trace := &domain.Trace{ID: traceID, ProjectID: projectID}
		traceRepo.On("GetByID", ctx, projectID, traceID).Return(trace, nil)
		gitLinkRepo.On("Create", ctx, mock.AnythingOfType("*domain.GitLink")).Return(errors.New("db error"))

		input := &domain.GitLinkInput{
			TraceID:   traceID,
			CommitSha: "abc123",
			Branch:    "main",
		}

		gitLink, err := svc.Create(ctx, projectID, input)

		require.Error(t, err)
		assert.Nil(t, gitLink)
		assert.Contains(t, err.Error(), "failed to create git link")
	})
}

func TestGitLinkService_Get(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	gitLinkID := uuid.New()

	t.Run("gets git link successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		expected := &domain.GitLink{
			ID:        gitLinkID,
			ProjectID: projectID,
			CommitSha: "abc123",
		}
		gitLinkRepo.On("GetByID", ctx, projectID, gitLinkID).Return(expected, nil)

		gitLink, err := svc.Get(ctx, projectID, gitLinkID)

		require.NoError(t, err)
		require.NotNil(t, gitLink)
		assert.Equal(t, expected, gitLink)
		gitLinkRepo.AssertExpectations(t)
	})

	t.Run("returns error when git link not found", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		gitLinkRepo.On("GetByID", ctx, projectID, gitLinkID).Return(nil, errors.New("not found"))

		gitLink, err := svc.Get(ctx, projectID, gitLinkID)

		require.Error(t, err)
		assert.Nil(t, gitLink)
		gitLinkRepo.AssertExpectations(t)
	})
}

func TestGitLinkService_GetByCommitSha(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	commitSha := "abc123def456"

	t.Run("gets git links by commit sha successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		expected := []domain.GitLink{
			{ID: uuid.New(), CommitSha: commitSha, TraceID: "trace-1"},
			{ID: uuid.New(), CommitSha: commitSha, TraceID: "trace-2"},
		}
		gitLinkRepo.On("GetByCommitSha", ctx, projectID, commitSha).Return(expected, nil)

		gitLinks, err := svc.GetByCommitSha(ctx, projectID, commitSha)

		require.NoError(t, err)
		assert.Len(t, gitLinks, 2)
		gitLinkRepo.AssertExpectations(t)
	})
}

func TestGitLinkService_GetByTraceID(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	traceID := "trace-123"

	t.Run("gets git links by trace ID successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		expected := []domain.GitLink{
			{ID: uuid.New(), TraceID: traceID, CommitSha: "abc123"},
			{ID: uuid.New(), TraceID: traceID, CommitSha: "def456"},
		}
		gitLinkRepo.On("GetByTraceID", ctx, projectID, traceID).Return(expected, nil)

		gitLinks, err := svc.GetByTraceID(ctx, projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, gitLinks, 2)
		gitLinkRepo.AssertExpectations(t)
	})
}

func TestGitLinkService_List(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("lists git links successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		filter := &domain.GitLinkFilter{ProjectID: projectID}
		expected := &domain.GitLinkList{
			GitLinks: []domain.GitLink{
				{ID: uuid.New(), CommitSha: "abc123"},
				{ID: uuid.New(), CommitSha: "def456"},
			},
			TotalCount: 2,
		}
		gitLinkRepo.On("List", ctx, filter, 10, 0).Return(expected, nil)

		result, err := svc.List(ctx, filter, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.GitLinks, 2)
		assert.Equal(t, int64(2), result.TotalCount)
		gitLinkRepo.AssertExpectations(t)
	})
}

func TestGitLinkService_GetTimeline(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("gets timeline successfully", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		expected := &domain.GitTimeline{
			Commits: []domain.GitTimelineEntry{
				{CommitSha: "abc123", Branch: "main"},
				{CommitSha: "def456", Branch: "main"},
			},
		}
		gitLinkRepo.On("GetTimeline", ctx, projectID, "main", 50).Return(expected, nil)

		result, err := svc.GetTimeline(ctx, projectID, "main", 50)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Commits, 2)
		gitLinkRepo.AssertExpectations(t)
	})

	t.Run("uses default limit when zero", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		expected := &domain.GitTimeline{Commits: []domain.GitTimelineEntry{}}
		gitLinkRepo.On("GetTimeline", ctx, projectID, "main", 50).Return(expected, nil)

		result, err := svc.GetTimeline(ctx, projectID, "main", 0)

		require.NoError(t, err)
		require.NotNil(t, result)
		gitLinkRepo.AssertExpectations(t)
	})

	t.Run("uses default limit when negative", func(t *testing.T) {
		gitLinkRepo := new(MockGitLinkRepository)
		traceRepo := new(MockTraceRepoForGitLink)
		svc := NewGitLinkService(gitLinkRepo, traceRepo)

		expected := &domain.GitTimeline{Commits: []domain.GitTimelineEntry{}}
		gitLinkRepo.On("GetTimeline", ctx, projectID, "main", 50).Return(expected, nil)

		result, err := svc.GetTimeline(ctx, projectID, "main", -10)

		require.NoError(t, err)
		require.NotNil(t, result)
		gitLinkRepo.AssertExpectations(t)
	})
}
