package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// GitHubAppRepository defines the interface for GitHub App data operations
type GitHubAppRepository interface {
	CreateInstallation(ctx context.Context, installation *domain.GitHubInstallation) error
	GetInstallationByID(ctx context.Context, id uuid.UUID) (*domain.GitHubInstallation, error)
	GetInstallationByInstallationID(ctx context.Context, installationID int64) (*domain.GitHubInstallation, error)
	ListInstallations(ctx context.Context, organizationID uuid.UUID) ([]domain.GitHubInstallation, error)
	UpdateInstallation(ctx context.Context, installation *domain.GitHubInstallation) error
	DeleteInstallation(ctx context.Context, id uuid.UUID) error

	CreateRepository(ctx context.Context, repo *domain.GitHubRepository) error
	GetRepositoryByID(ctx context.Context, id uuid.UUID) (*domain.GitHubRepository, error)
	GetRepositoryByRepoID(ctx context.Context, installationID uuid.UUID, repoID int64) (*domain.GitHubRepository, error)
	GetRepositoryByFullName(ctx context.Context, fullName string) (*domain.GitHubRepository, error)
	ListRepositoriesByInstallation(ctx context.Context, installationID uuid.UUID) ([]domain.GitHubRepository, error)
	ListRepositoriesByProject(ctx context.Context, projectID uuid.UUID) ([]domain.GitHubRepository, error)
	UpdateRepository(ctx context.Context, repo *domain.GitHubRepository) error
	LinkRepositoryToProject(ctx context.Context, repoID, projectID uuid.UUID, autoLink bool) error
	DeleteRepository(ctx context.Context, id uuid.UUID) error
	GetAutoLinkRepositories(ctx context.Context, filter *domain.GitHubRepositoryFilter) ([]domain.GitHubRepository, error)

	CreateWebhookEvent(ctx context.Context, event *domain.GitHubWebhookEvent) error
	GetUnprocessedWebhookEvents(ctx context.Context, limit int) ([]domain.GitHubWebhookEvent, error)
	MarkWebhookEventProcessed(ctx context.Context, id uuid.UUID, errMsg *string) error
}

// GitHubAppService handles GitHub App operations
type GitHubAppService struct {
	repo          GitHubAppRepository
	gitLinkRepo   GitLinkRepository
	logger        *zap.Logger
	webhookSecret string
	appID         int64
}

// NewGitHubAppService creates a new GitHub App service
func NewGitHubAppService(
	repo GitHubAppRepository,
	gitLinkRepo GitLinkRepository,
	logger *zap.Logger,
	webhookSecret string,
	appID int64,
) *GitHubAppService {
	return &GitHubAppService{
		repo:          repo,
		gitLinkRepo:   gitLinkRepo,
		logger:        logger,
		webhookSecret: webhookSecret,
		appID:         appID,
	}
}

// VerifyWebhookSignature verifies the GitHub webhook signature
func (s *GitHubAppService) VerifyWebhookSignature(payload []byte, signature string) bool {
	if s.webhookSecret == "" {
		return true // Skip verification if no secret configured
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	sig := strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

// HandleWebhook processes an incoming webhook event
func (s *GitHubAppService) HandleWebhook(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var genericPayload map[string]any
	if err := json.Unmarshal(payload, &genericPayload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract installation ID
	var installationID int64
	if inst, ok := genericPayload["installation"].(map[string]any); ok {
		if id, ok := inst["id"].(float64); ok {
			installationID = int64(id)
		}
	}

	// Extract action
	var action string
	if a, ok := genericPayload["action"].(string); ok {
		action = a
	}

	// Store webhook event for processing
	event := &domain.GitHubWebhookEvent{
		ID:             uuid.New(),
		InstallationID: installationID,
		EventType:      eventType,
		Action:         action,
		DeliveryID:     deliveryID,
		Payload:        genericPayload,
	}

	if err := s.repo.CreateWebhookEvent(ctx, event); err != nil {
		s.logger.Error("failed to store webhook event", zap.Error(err))
		return err
	}

	// Process synchronously for critical events
	switch eventType {
	case "installation":
		return s.handleInstallationEvent(ctx, payload, action)
	case "push":
		return s.handlePushEvent(ctx, payload)
	case "installation_repositories":
		return s.handleInstallationRepositoriesEvent(ctx, payload, action)
	}

	return nil
}

// handleInstallationEvent processes installation events
func (s *GitHubAppService) handleInstallationEvent(ctx context.Context, payload []byte, action string) error {
	var event domain.GitHubInstallationPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse installation payload: %w", err)
	}

	switch action {
	case "created":
		return s.createInstallation(ctx, &event)
	case "deleted":
		return s.deleteInstallation(ctx, event.Installation.ID)
	case "suspend":
		return s.suspendInstallation(ctx, event.Installation.ID, event.Sender.Login)
	case "unsuspend":
		return s.unsuspendInstallation(ctx, event.Installation.ID)
	}

	return nil
}

// createInstallation creates a new installation from webhook
func (s *GitHubAppService) createInstallation(ctx context.Context, event *domain.GitHubInstallationPayload) error {
	// Note: Organization ID needs to be resolved from account
	// For now, we'll store with a placeholder and require manual linking
	installation := &domain.GitHubInstallation{
		ID:                  uuid.New(),
		OrganizationID:      uuid.Nil, // Needs to be linked by user
		InstallationID:      event.Installation.ID,
		AccountID:           event.Installation.Account.ID,
		AccountLogin:        event.Installation.Account.Login,
		AccountType:         event.Installation.Account.Type,
		TargetType:          event.Installation.TargetType,
		AppID:               event.Installation.AppID,
		AppSlug:             event.Installation.AppSlug,
		RepositorySelection: event.Installation.RepositorySelection,
		AccessTokensURL:     event.Installation.AccessTokensURL,
		RepositoriesURL:     event.Installation.RepositoriesURL,
		HTMLURL:             event.Installation.HTMLURL,
		Permissions:         s.convertPermissions(event.Installation.Permissions),
		Events:              event.Installation.Events,
	}

	if err := s.repo.CreateInstallation(ctx, installation); err != nil {
		return fmt.Errorf("failed to create installation: %w", err)
	}

	// Create repository records
	for _, repo := range event.Repositories {
		repoRecord := &domain.GitHubRepository{
			ID:             uuid.New(),
			InstallationID: installation.ID,
			RepoID:         repo.ID,
			RepoFullName:   repo.FullName,
			RepoName:       repo.Name,
			Owner:          strings.Split(repo.FullName, "/")[0],
			Private:        repo.Private,
			SyncEnabled:    true,
			AutoLink:       false, // Disabled by default until project is linked
		}

		if err := s.repo.CreateRepository(ctx, repoRecord); err != nil {
			s.logger.Warn("failed to create repository",
				zap.String("repo", repo.FullName),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("GitHub App installation created",
		zap.Int64("installation_id", installation.InstallationID),
		zap.String("account", installation.AccountLogin),
	)

	return nil
}

// deleteInstallation removes an installation
func (s *GitHubAppService) deleteInstallation(ctx context.Context, installationID int64) error {
	installation, err := s.repo.GetInstallationByInstallationID(ctx, installationID)
	if err != nil {
		return fmt.Errorf("failed to find installation: %w", err)
	}

	if err := s.repo.DeleteInstallation(ctx, installation.ID); err != nil {
		return fmt.Errorf("failed to delete installation: %w", err)
	}

	s.logger.Info("GitHub App installation deleted",
		zap.Int64("installation_id", installationID),
	)

	return nil
}

// suspendInstallation marks an installation as suspended
func (s *GitHubAppService) suspendInstallation(ctx context.Context, installationID int64, suspendedBy string) error {
	installation, err := s.repo.GetInstallationByInstallationID(ctx, installationID)
	if err != nil {
		return fmt.Errorf("failed to find installation: %w", err)
	}

	now := time.Now()
	installation.SuspendedAt = &now
	installation.SuspendedBy = &suspendedBy

	if err := s.repo.UpdateInstallation(ctx, installation); err != nil {
		return fmt.Errorf("failed to update installation: %w", err)
	}

	s.logger.Info("GitHub App installation suspended",
		zap.Int64("installation_id", installationID),
		zap.String("suspended_by", suspendedBy),
	)

	return nil
}

// unsuspendInstallation removes suspension from an installation
func (s *GitHubAppService) unsuspendInstallation(ctx context.Context, installationID int64) error {
	installation, err := s.repo.GetInstallationByInstallationID(ctx, installationID)
	if err != nil {
		return fmt.Errorf("failed to find installation: %w", err)
	}

	installation.SuspendedAt = nil
	installation.SuspendedBy = nil

	if err := s.repo.UpdateInstallation(ctx, installation); err != nil {
		return fmt.Errorf("failed to update installation: %w", err)
	}

	s.logger.Info("GitHub App installation unsuspended",
		zap.Int64("installation_id", installationID),
	)

	return nil
}

// handleInstallationRepositoriesEvent processes repository add/remove events
func (s *GitHubAppService) handleInstallationRepositoriesEvent(ctx context.Context, payload []byte, action string) error {
	var event struct {
		Action       string `json:"action"`
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
		RepositoriesAdded   []domain.GitHubRepo `json:"repositories_added"`
		RepositoriesRemoved []domain.GitHubRepo `json:"repositories_removed"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse repositories event: %w", err)
	}

	installation, err := s.repo.GetInstallationByInstallationID(ctx, event.Installation.ID)
	if err != nil {
		return fmt.Errorf("failed to find installation: %w", err)
	}

	// Handle added repositories
	for _, repo := range event.RepositoriesAdded {
		repoRecord := &domain.GitHubRepository{
			ID:             uuid.New(),
			InstallationID: installation.ID,
			RepoID:         repo.ID,
			RepoFullName:   repo.FullName,
			RepoName:       repo.Name,
			Owner:          repo.Owner.Login,
			Private:        repo.Private,
			DefaultBranch:  repo.DefaultBranch,
			HTMLURL:        repo.HTMLURL,
			CloneURL:       repo.CloneURL,
			SyncEnabled:    true,
			AutoLink:       false,
		}

		if err := s.repo.CreateRepository(ctx, repoRecord); err != nil {
			s.logger.Warn("failed to create repository",
				zap.String("repo", repo.FullName),
				zap.Error(err),
			)
		}
	}

	// Handle removed repositories
	for _, repo := range event.RepositoriesRemoved {
		existing, err := s.repo.GetRepositoryByRepoID(ctx, installation.ID, repo.ID)
		if err != nil {
			continue
		}
		if err := s.repo.DeleteRepository(ctx, existing.ID); err != nil {
			s.logger.Warn("failed to delete repository",
				zap.String("repo", repo.FullName),
				zap.Error(err),
			)
		}
	}

	return nil
}

// handlePushEvent processes push events to auto-link commits to traces
func (s *GitHubAppService) handlePushEvent(ctx context.Context, payload []byte) error {
	var event domain.GitHubPushPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse push payload: %w", err)
	}

	// Find the repository with auto-link enabled
	repo, err := s.repo.GetRepositoryByFullName(ctx, event.Repository.FullName)
	if err != nil {
		s.logger.Debug("repository not found or not tracked", zap.String("repo", event.Repository.FullName))
		return nil
	}

	if !repo.AutoLink || repo.ProjectID == uuid.Nil {
		s.logger.Debug("auto-link disabled or no project linked", zap.String("repo", event.Repository.FullName))
		return nil
	}

	// Extract branch name from ref
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// Process each commit
	for _, commit := range event.Commits {
		// Parse commit timestamp
		commitTime, _ := time.Parse(time.RFC3339, commit.Timestamp)

		gitLink := &domain.GitLink{
			ID:                uuid.New(),
			ProjectID:         repo.ProjectID,
			TraceID:           "", // Will be linked later when trace references this commit
			CommitSha:         commit.ID,
			ParentSha:         event.Before,
			Branch:            branch,
			RepoURL:           event.Repository.HTMLURL,
			CommitMessage:     commit.Message,
			CommitAuthor:      commit.Author.Name,
			CommitAuthorEmail: commit.Author.Email,
			CommitTimestamp:   commitTime,
			FilesAdded:        commit.Added,
			FilesModified:     commit.Modified,
			FilesDeleted:      commit.Removed,
			FilesChangedCount: uint32(len(commit.Added) + len(commit.Modified) + len(commit.Removed)),
			LinkType:          domain.GitLinkTypePush,
			CreatedAt:         time.Now(),
		}

		// Store the git link for later correlation
		if err := s.gitLinkRepo.Create(ctx, gitLink); err != nil {
			s.logger.Warn("failed to create git link from push",
				zap.String("commit", commit.ID),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("processed push event",
		zap.String("repo", event.Repository.FullName),
		zap.Int("commits", len(event.Commits)),
	)

	return nil
}

// LinkInstallationToOrganization links an installation to an AgentTrace organization
func (s *GitHubAppService) LinkInstallationToOrganization(ctx context.Context, installationID int64, organizationID uuid.UUID) error {
	installation, err := s.repo.GetInstallationByInstallationID(ctx, installationID)
	if err != nil {
		return fmt.Errorf("failed to find installation: %w", err)
	}

	installation.OrganizationID = organizationID
	return s.repo.UpdateInstallation(ctx, installation)
}

// LinkRepositoryToProject links a GitHub repository to an AgentTrace project
func (s *GitHubAppService) LinkRepositoryToProject(ctx context.Context, input *domain.LinkRepoToProjectInput) error {
	return s.repo.LinkRepositoryToProject(ctx, input.RepositoryID, input.ProjectID, input.AutoLink)
}

// GetInstallations retrieves installations for an organization
func (s *GitHubAppService) GetInstallations(ctx context.Context, organizationID uuid.UUID) ([]domain.GitHubInstallation, error) {
	return s.repo.ListInstallations(ctx, organizationID)
}

// GetRepositories retrieves repositories for an installation
func (s *GitHubAppService) GetRepositories(ctx context.Context, installationID uuid.UUID) ([]domain.GitHubRepository, error) {
	return s.repo.ListRepositoriesByInstallation(ctx, installationID)
}

// GetProjectRepositories retrieves repositories linked to a project
func (s *GitHubAppService) GetProjectRepositories(ctx context.Context, projectID uuid.UUID) ([]domain.GitHubRepository, error) {
	return s.repo.ListRepositoriesByProject(ctx, projectID)
}

// convertPermissions converts map[string]string to JSONMap
func (s *GitHubAppService) convertPermissions(perms map[string]string) domain.JSONMap {
	result := make(domain.JSONMap)
	for k, v := range perms {
		result[k] = v
	}
	return result
}

// ProcessPendingWebhooks processes pending webhook events (for background worker)
func (s *GitHubAppService) ProcessPendingWebhooks(ctx context.Context, batchSize int) error {
	events, err := s.repo.GetUnprocessedWebhookEvents(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending webhooks: %w", err)
	}

	for _, event := range events {
		payload, _ := json.Marshal(event.Payload)

		var processErr error
		switch event.EventType {
		case "push":
			processErr = s.handlePushEvent(ctx, payload)
		}

		var errMsg *string
		if processErr != nil {
			msg := processErr.Error()
			errMsg = &msg
		}

		if err := s.repo.MarkWebhookEventProcessed(ctx, event.ID, errMsg); err != nil {
			s.logger.Error("failed to mark webhook processed", zap.Error(err))
		}
	}

	return nil
}
