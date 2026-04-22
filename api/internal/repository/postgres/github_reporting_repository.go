package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// GitHubReportingRepository reads project-linked GitHub repository configuration.
type GitHubReportingRepository struct {
	db *database.PostgresDB
}

// NewGitHubReportingRepository creates a focused reporting repository.
func NewGitHubReportingRepository(db *database.PostgresDB) *GitHubReportingRepository {
	return &GitHubReportingRepository{db: db}
}

// GetProjectRepository prevents cross-project report delivery.
func (r *GitHubReportingRepository) GetProjectRepository(
	ctx context.Context,
	projectID, repositoryID uuid.UUID,
) (*domain.GitHubRepository, error) {
	var repository domain.GitHubRepository
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, installation_id, project_id, repo_id, repo_full_name, repo_name,
			owner, private, default_branch, html_url, clone_url, sync_enabled,
			auto_link, created_at, updated_at
		FROM github_repositories
		WHERE project_id = $1 AND id = $2
	`, projectID, repositoryID).Scan(
		&repository.ID,
		&repository.InstallationID,
		&repository.ProjectID,
		&repository.RepoID,
		&repository.RepoFullName,
		&repository.RepoName,
		&repository.Owner,
		&repository.Private,
		&repository.DefaultBranch,
		&repository.HTMLURL,
		&repository.CloneURL,
		&repository.SyncEnabled,
		&repository.AutoLink,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("GitHub repository")
		}
		return nil, fmt.Errorf("get project GitHub repository: %w", err)
	}
	return &repository, nil
}
