package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// GitHubAppRepository handles GitHub App related database operations
type GitHubAppRepository struct {
	db *sqlx.DB
}

// NewGitHubAppRepository creates a new GitHub App repository
func NewGitHubAppRepository(db *sqlx.DB) *GitHubAppRepository {
	return &GitHubAppRepository{db: db}
}

// Installation operations

// CreateInstallation creates a new GitHub installation
func (r *GitHubAppRepository) CreateInstallation(ctx context.Context, installation *domain.GitHubInstallation) error {
	permissions, _ := json.Marshal(installation.Permissions)

	query := `
		INSERT INTO github_installations (
			id, organization_id, installation_id, account_id, account_login,
			account_type, target_type, app_id, app_slug, repository_selection,
			access_tokens_url, repositories_url, html_url, permissions, events,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	now := time.Now()
	installation.CreatedAt = now
	installation.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		installation.ID, installation.OrganizationID, installation.InstallationID,
		installation.AccountID, installation.AccountLogin, installation.AccountType,
		installation.TargetType, installation.AppID, installation.AppSlug,
		installation.RepositorySelection, installation.AccessTokensURL,
		installation.RepositoriesURL, installation.HTMLURL, permissions,
		pq.Array(installation.Events), installation.CreatedAt, installation.UpdatedAt,
	)

	return err
}

// GetInstallationByID retrieves an installation by ID
func (r *GitHubAppRepository) GetInstallationByID(ctx context.Context, id uuid.UUID) (*domain.GitHubInstallation, error) {
	var installation domain.GitHubInstallation
	var permissions []byte
	var events pq.StringArray

	query := `
		SELECT id, organization_id, installation_id, account_id, account_login,
			account_type, target_type, app_id, app_slug, repository_selection,
			access_tokens_url, repositories_url, html_url, permissions, events,
			suspended_at, suspended_by, created_at, updated_at
		FROM github_installations
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&installation.ID, &installation.OrganizationID, &installation.InstallationID,
		&installation.AccountID, &installation.AccountLogin, &installation.AccountType,
		&installation.TargetType, &installation.AppID, &installation.AppSlug,
		&installation.RepositorySelection, &installation.AccessTokensURL,
		&installation.RepositoriesURL, &installation.HTMLURL, &permissions, &events,
		&installation.SuspendedAt, &installation.SuspendedBy, &installation.CreatedAt, &installation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("installation not found")
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(permissions, &installation.Permissions)
	installation.Events = events

	return &installation, nil
}

// GetInstallationByInstallationID retrieves by GitHub installation ID
func (r *GitHubAppRepository) GetInstallationByInstallationID(ctx context.Context, installationID int64) (*domain.GitHubInstallation, error) {
	var installation domain.GitHubInstallation
	var permissions []byte
	var events pq.StringArray

	query := `
		SELECT id, organization_id, installation_id, account_id, account_login,
			account_type, target_type, app_id, app_slug, repository_selection,
			access_tokens_url, repositories_url, html_url, permissions, events,
			suspended_at, suspended_by, created_at, updated_at
		FROM github_installations
		WHERE installation_id = $1`

	err := r.db.QueryRowContext(ctx, query, installationID).Scan(
		&installation.ID, &installation.OrganizationID, &installation.InstallationID,
		&installation.AccountID, &installation.AccountLogin, &installation.AccountType,
		&installation.TargetType, &installation.AppID, &installation.AppSlug,
		&installation.RepositorySelection, &installation.AccessTokensURL,
		&installation.RepositoriesURL, &installation.HTMLURL, &permissions, &events,
		&installation.SuspendedAt, &installation.SuspendedBy, &installation.CreatedAt, &installation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("installation not found")
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(permissions, &installation.Permissions)
	installation.Events = events

	return &installation, nil
}

// ListInstallations retrieves installations by organization
func (r *GitHubAppRepository) ListInstallations(ctx context.Context, organizationID uuid.UUID) ([]domain.GitHubInstallation, error) {
	query := `
		SELECT id, organization_id, installation_id, account_id, account_login,
			account_type, target_type, app_id, app_slug, repository_selection,
			access_tokens_url, repositories_url, html_url, permissions, events,
			suspended_at, suspended_by, created_at, updated_at
		FROM github_installations
		WHERE organization_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var installations []domain.GitHubInstallation
	for rows.Next() {
		var installation domain.GitHubInstallation
		var permissions []byte
		var events pq.StringArray

		err := rows.Scan(
			&installation.ID, &installation.OrganizationID, &installation.InstallationID,
			&installation.AccountID, &installation.AccountLogin, &installation.AccountType,
			&installation.TargetType, &installation.AppID, &installation.AppSlug,
			&installation.RepositorySelection, &installation.AccessTokensURL,
			&installation.RepositoriesURL, &installation.HTMLURL, &permissions, &events,
			&installation.SuspendedAt, &installation.SuspendedBy, &installation.CreatedAt, &installation.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(permissions, &installation.Permissions)
		installation.Events = events
		installations = append(installations, installation)
	}

	return installations, nil
}

// UpdateInstallation updates an installation
func (r *GitHubAppRepository) UpdateInstallation(ctx context.Context, installation *domain.GitHubInstallation) error {
	permissions, _ := json.Marshal(installation.Permissions)
	installation.UpdatedAt = time.Now()

	query := `
		UPDATE github_installations SET
			organization_id = $2,
			repository_selection = $3,
			permissions = $4,
			events = $5,
			suspended_at = $6,
			suspended_by = $7,
			updated_at = $8
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		installation.ID, installation.OrganizationID, installation.RepositorySelection,
		permissions, pq.Array(installation.Events),
		installation.SuspendedAt, installation.SuspendedBy, installation.UpdatedAt,
	)

	return err
}

// ListUnlinkedInstallations retrieves installations not linked to any organization
func (r *GitHubAppRepository) ListUnlinkedInstallations(ctx context.Context) ([]domain.GitHubInstallation, error) {
	query := `
		SELECT id, organization_id, installation_id, account_id, account_login,
			account_type, target_type, app_id, app_slug, repository_selection,
			access_tokens_url, repositories_url, html_url, permissions, events,
			suspended_at, suspended_by, created_at, updated_at
		FROM github_installations
		WHERE organization_id = '00000000-0000-0000-0000-000000000000'
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var installations []domain.GitHubInstallation
	for rows.Next() {
		var installation domain.GitHubInstallation
		var permissions []byte
		var events pq.StringArray

		err := rows.Scan(
			&installation.ID, &installation.OrganizationID, &installation.InstallationID,
			&installation.AccountID, &installation.AccountLogin, &installation.AccountType,
			&installation.TargetType, &installation.AppID, &installation.AppSlug,
			&installation.RepositorySelection, &installation.AccessTokensURL,
			&installation.RepositoriesURL, &installation.HTMLURL, &permissions, &events,
			&installation.SuspendedAt, &installation.SuspendedBy, &installation.CreatedAt, &installation.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(permissions, &installation.Permissions)
		installation.Events = events
		installations = append(installations, installation)
	}

	return installations, nil
}

// DeleteInstallation removes an installation
func (r *GitHubAppRepository) DeleteInstallation(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM github_installations WHERE id = $1", id)
	return err
}

// Repository operations

// CreateRepository creates a new GitHub repository
func (r *GitHubAppRepository) CreateRepository(ctx context.Context, repo *domain.GitHubRepository) error {
	query := `
		INSERT INTO github_repositories (
			id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (installation_id, repo_id) DO UPDATE SET
			repo_full_name = EXCLUDED.repo_full_name,
			repo_name = EXCLUDED.repo_name,
			private = EXCLUDED.private,
			default_branch = EXCLUDED.default_branch,
			html_url = EXCLUDED.html_url,
			clone_url = EXCLUDED.clone_url,
			updated_at = EXCLUDED.updated_at`

	now := time.Now()
	repo.CreatedAt = now
	repo.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		repo.ID, repo.InstallationID, repo.ProjectID, repo.RepoID, repo.RepoFullName,
		repo.RepoName, repo.Owner, repo.Private, repo.DefaultBranch, repo.HTMLURL,
		repo.CloneURL, repo.SyncEnabled, repo.AutoLink, repo.CreatedAt, repo.UpdatedAt,
	)

	return err
}

// GetRepositoryByID retrieves a repository by ID
func (r *GitHubAppRepository) GetRepositoryByID(ctx context.Context, id uuid.UUID) (*domain.GitHubRepository, error) {
	var repo domain.GitHubRepository

	query := `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&repo.ID, &repo.InstallationID, &repo.ProjectID, &repo.RepoID, &repo.RepoFullName,
		&repo.RepoName, &repo.Owner, &repo.Private, &repo.DefaultBranch, &repo.HTMLURL,
		&repo.CloneURL, &repo.SyncEnabled, &repo.AutoLink, &repo.CreatedAt, &repo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("repository not found")
	}
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepositoryByRepoID retrieves by GitHub repo ID
func (r *GitHubAppRepository) GetRepositoryByRepoID(ctx context.Context, installationID uuid.UUID, repoID int64) (*domain.GitHubRepository, error) {
	var repo domain.GitHubRepository

	query := `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE installation_id = $1 AND repo_id = $2`

	err := r.db.QueryRowContext(ctx, query, installationID, repoID).Scan(
		&repo.ID, &repo.InstallationID, &repo.ProjectID, &repo.RepoID, &repo.RepoFullName,
		&repo.RepoName, &repo.Owner, &repo.Private, &repo.DefaultBranch, &repo.HTMLURL,
		&repo.CloneURL, &repo.SyncEnabled, &repo.AutoLink, &repo.CreatedAt, &repo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("repository not found")
	}
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepositoryByFullName retrieves by repo full name
func (r *GitHubAppRepository) GetRepositoryByFullName(ctx context.Context, fullName string) (*domain.GitHubRepository, error) {
	var repo domain.GitHubRepository

	query := `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE repo_full_name = $1`

	err := r.db.QueryRowContext(ctx, query, fullName).Scan(
		&repo.ID, &repo.InstallationID, &repo.ProjectID, &repo.RepoID, &repo.RepoFullName,
		&repo.RepoName, &repo.Owner, &repo.Private, &repo.DefaultBranch, &repo.HTMLURL,
		&repo.CloneURL, &repo.SyncEnabled, &repo.AutoLink, &repo.CreatedAt, &repo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("repository not found")
	}
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// ListRepositoriesByInstallation retrieves repositories by installation
func (r *GitHubAppRepository) ListRepositoriesByInstallation(ctx context.Context, installationID uuid.UUID) ([]domain.GitHubRepository, error) {
	query := `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE installation_id = $1
		ORDER BY repo_full_name`

	rows, err := r.db.QueryContext(ctx, query, installationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []domain.GitHubRepository
	for rows.Next() {
		var repo domain.GitHubRepository
		err := rows.Scan(
			&repo.ID, &repo.InstallationID, &repo.ProjectID, &repo.RepoID, &repo.RepoFullName,
			&repo.RepoName, &repo.Owner, &repo.Private, &repo.DefaultBranch, &repo.HTMLURL,
			&repo.CloneURL, &repo.SyncEnabled, &repo.AutoLink, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// ListRepositoriesByProject retrieves repositories linked to a project
func (r *GitHubAppRepository) ListRepositoriesByProject(ctx context.Context, projectID uuid.UUID) ([]domain.GitHubRepository, error) {
	query := `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE project_id = $1 AND sync_enabled = true
		ORDER BY repo_full_name`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []domain.GitHubRepository
	for rows.Next() {
		var repo domain.GitHubRepository
		err := rows.Scan(
			&repo.ID, &repo.InstallationID, &repo.ProjectID, &repo.RepoID, &repo.RepoFullName,
			&repo.RepoName, &repo.Owner, &repo.Private, &repo.DefaultBranch, &repo.HTMLURL,
			&repo.CloneURL, &repo.SyncEnabled, &repo.AutoLink, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// UpdateRepository updates a repository
func (r *GitHubAppRepository) UpdateRepository(ctx context.Context, repo *domain.GitHubRepository) error {
	repo.UpdatedAt = time.Now()

	query := `
		UPDATE github_repositories SET
			project_id = $2,
			sync_enabled = $3,
			auto_link = $4,
			default_branch = $5,
			updated_at = $6
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		repo.ID, repo.ProjectID, repo.SyncEnabled, repo.AutoLink,
		repo.DefaultBranch, repo.UpdatedAt,
	)

	return err
}

// LinkRepositoryToProject links a repository to a project
func (r *GitHubAppRepository) LinkRepositoryToProject(ctx context.Context, repoID, projectID uuid.UUID, autoLink bool) error {
	query := `
		UPDATE github_repositories SET
			project_id = $2,
			auto_link = $3,
			updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, repoID, projectID, autoLink)
	return err
}

// DeleteRepository removes a repository
func (r *GitHubAppRepository) DeleteRepository(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM github_repositories WHERE id = $1", id)
	return err
}

// Webhook event operations

// CreateWebhookEvent stores a webhook event for processing
func (r *GitHubAppRepository) CreateWebhookEvent(ctx context.Context, event *domain.GitHubWebhookEvent) error {
	payload, _ := json.Marshal(event.Payload)

	query := `
		INSERT INTO github_webhook_events (
			id, installation_id, event_type, action, delivery_id, payload, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	event.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.InstallationID, event.EventType, event.Action,
		event.DeliveryID, payload, event.CreatedAt,
	)

	return err
}

// GetUnprocessedWebhookEvents retrieves unprocessed events
func (r *GitHubAppRepository) GetUnprocessedWebhookEvents(ctx context.Context, limit int) ([]domain.GitHubWebhookEvent, error) {
	query := `
		SELECT id, installation_id, event_type, action, delivery_id, payload,
			processed, processed_at, error, created_at
		FROM github_webhook_events
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.GitHubWebhookEvent
	for rows.Next() {
		var event domain.GitHubWebhookEvent
		var payload []byte

		err := rows.Scan(
			&event.ID, &event.InstallationID, &event.EventType, &event.Action,
			&event.DeliveryID, &payload, &event.Processed, &event.ProcessedAt,
			&event.Error, &event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(payload, &event.Payload)
		events = append(events, event)
	}

	return events, nil
}

// MarkWebhookEventProcessed marks an event as processed
func (r *GitHubAppRepository) MarkWebhookEventProcessed(ctx context.Context, id uuid.UUID, errMsg *string) error {
	now := time.Now()

	query := `
		UPDATE github_webhook_events SET
			processed = true,
			processed_at = $2,
			error = $3
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, now, errMsg)
	return err
}

// GetAutoLinkRepositories retrieves repositories with auto-link enabled
func (r *GitHubAppRepository) GetAutoLinkRepositories(ctx context.Context, filter *domain.GitHubRepositoryFilter) ([]domain.GitHubRepository, error) {
	var args []any
	var conditions []string

	query := `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE auto_link = true AND project_id IS NOT NULL`

	argIdx := 1

	if filter != nil {
		if filter.InstallationID != nil {
			conditions = append(conditions, fmt.Sprintf("installation_id = $%d", argIdx))
			args = append(args, *filter.InstallationID)
			argIdx++
		}
		if filter.RepoFullName != nil {
			conditions = append(conditions, fmt.Sprintf("repo_full_name = $%d", argIdx))
			args = append(args, *filter.RepoFullName)
			argIdx++
		}
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY repo_full_name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []domain.GitHubRepository
	for rows.Next() {
		var repo domain.GitHubRepository
		err := rows.Scan(
			&repo.ID, &repo.InstallationID, &repo.ProjectID, &repo.RepoID, &repo.RepoFullName,
			&repo.RepoName, &repo.Owner, &repo.Private, &repo.DefaultBranch, &repo.HTMLURL,
			&repo.CloneURL, &repo.SyncEnabled, &repo.AutoLink, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, nil
}
