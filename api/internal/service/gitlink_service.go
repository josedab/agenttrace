package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// GitLinkRepository defines git link repository operations
type GitLinkRepository interface {
	Create(ctx context.Context, gitLink *domain.GitLink) error
	GetByID(ctx context.Context, projectID, gitLinkID uuid.UUID) (*domain.GitLink, error)
	GetByCommitSha(ctx context.Context, projectID uuid.UUID, commitSha string) ([]domain.GitLink, error)
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.GitLink, error)
	List(ctx context.Context, filter *domain.GitLinkFilter, limit, offset int) (*domain.GitLinkList, error)
	GetTimeline(ctx context.Context, projectID uuid.UUID, branch string, limit int) (*domain.GitTimeline, error)
}

// GitLinkService handles git link operations
type GitLinkService struct {
	gitLinkRepo GitLinkRepository
	traceRepo   TraceRepository
}

// NewGitLinkService creates a new git link service
func NewGitLinkService(
	gitLinkRepo GitLinkRepository,
	traceRepo TraceRepository,
) *GitLinkService {
	return &GitLinkService{
		gitLinkRepo: gitLinkRepo,
		traceRepo:   traceRepo,
	}
}

// Create creates a new git link
func (s *GitLinkService) Create(ctx context.Context, projectID uuid.UUID, input *domain.GitLinkInput) (*domain.GitLink, error) {
	// Verify trace exists
	_, err := s.traceRepo.GetByID(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	linkType := input.LinkType
	if linkType == "" {
		linkType = domain.GitLinkTypeCurrent
	}

	commitTimestamp := time.Now()
	if input.CommitTimestamp != nil {
		commitTimestamp = *input.CommitTimestamp
	}

	var commitMessage, commitAuthor, commitAuthorEmail string
	if input.CommitMessage != nil {
		commitMessage = *input.CommitMessage
	}
	if input.CommitAuthor != nil {
		commitAuthor = *input.CommitAuthor
	}
	if input.CommitAuthorEmail != nil {
		commitAuthorEmail = *input.CommitAuthorEmail
	}

	var additions, deletions uint32
	if input.Additions != nil {
		additions = *input.Additions
	}
	if input.Deletions != nil {
		deletions = *input.Deletions
	}

	filesChangedCount := uint32(len(input.FilesAdded) + len(input.FilesModified) + len(input.FilesDeleted))

	gitLink := &domain.GitLink{
		ID:                uuid.New(),
		ProjectID:         projectID,
		TraceID:           input.TraceID,
		CommitSha:         input.CommitSha,
		ParentSha:         input.ParentSha,
		Branch:            input.Branch,
		Tag:               input.Tag,
		RepoURL:           input.RepoURL,
		CommitMessage:     commitMessage,
		CommitAuthor:      commitAuthor,
		CommitAuthorEmail: commitAuthorEmail,
		CommitTimestamp:   commitTimestamp,
		FilesAdded:        input.FilesAdded,
		FilesModified:     input.FilesModified,
		FilesDeleted:      input.FilesDeleted,
		FilesChangedCount: filesChangedCount,
		Additions:         additions,
		Deletions:         deletions,
		LinkType:          linkType,
		CreatedAt:         time.Now(),
	}

	if input.CIRunID != nil {
		ciRunID, err := uuid.Parse(*input.CIRunID)
		if err == nil {
			gitLink.CIRunID = &ciRunID
		}
	}

	if err := s.gitLinkRepo.Create(ctx, gitLink); err != nil {
		return nil, fmt.Errorf("failed to create git link: %w", err)
	}

	return gitLink, nil
}

// Get retrieves a git link by ID
func (s *GitLinkService) Get(ctx context.Context, projectID, gitLinkID uuid.UUID) (*domain.GitLink, error) {
	return s.gitLinkRepo.GetByID(ctx, projectID, gitLinkID)
}

// GetByCommitSha retrieves git links for a commit
func (s *GitLinkService) GetByCommitSha(ctx context.Context, projectID uuid.UUID, commitSha string) ([]domain.GitLink, error) {
	return s.gitLinkRepo.GetByCommitSha(ctx, projectID, commitSha)
}

// GetByTraceID retrieves git links for a trace
func (s *GitLinkService) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.GitLink, error) {
	return s.gitLinkRepo.GetByTraceID(ctx, projectID, traceID)
}

// List retrieves git links with filtering
func (s *GitLinkService) List(ctx context.Context, filter *domain.GitLinkFilter, limit, offset int) (*domain.GitLinkList, error) {
	return s.gitLinkRepo.List(ctx, filter, limit, offset)
}

// GetTimeline retrieves a git timeline for a project
func (s *GitLinkService) GetTimeline(ctx context.Context, projectID uuid.UUID, branch string, limit int) (*domain.GitTimeline, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.gitLinkRepo.GetTimeline(ctx, projectID, branch, limit)
}
