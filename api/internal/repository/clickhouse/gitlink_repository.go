package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// GitLinkRepository handles git link data operations in ClickHouse
type GitLinkRepository struct {
	db *database.ClickHouseDB
}

// NewGitLinkRepository creates a new git link repository
func NewGitLinkRepository(db *database.ClickHouseDB) *GitLinkRepository {
	return &GitLinkRepository{db: db}
}

// Create inserts a new git link
func (r *GitLinkRepository) Create(ctx context.Context, gitLink *domain.GitLink) error {
	query := `
		INSERT INTO git_links (
			id, project_id, trace_id, commit_sha, parent_sha, branch, tag,
			repo_url, commit_message, commit_author, commit_author_email,
			commit_timestamp, files_added, files_modified, files_deleted,
			files_changed_count, additions, deletions, link_type, ci_run_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		gitLink.ID,
		gitLink.ProjectID,
		gitLink.TraceID,
		gitLink.CommitSha,
		gitLink.ParentSha,
		gitLink.Branch,
		gitLink.Tag,
		gitLink.RepoURL,
		gitLink.CommitMessage,
		gitLink.CommitAuthor,
		gitLink.CommitAuthorEmail,
		gitLink.CommitTimestamp,
		gitLink.FilesAdded,
		gitLink.FilesModified,
		gitLink.FilesDeleted,
		gitLink.FilesChangedCount,
		gitLink.Additions,
		gitLink.Deletions,
		string(gitLink.LinkType),
		gitLink.CIRunID,
		gitLink.CreatedAt,
	)
}

// GetByID retrieves a git link by ID
func (r *GitLinkRepository) GetByID(ctx context.Context, projectID, gitLinkID uuid.UUID) (*domain.GitLink, error) {
	query := `
		SELECT
			id, project_id, trace_id, commit_sha, parent_sha, branch, tag,
			repo_url, commit_message, commit_author, commit_author_email,
			commit_timestamp, files_added, files_modified, files_deleted,
			files_changed_count, additions, deletions, link_type, ci_run_id, created_at
		FROM git_links FINAL
		WHERE project_id = ? AND id = ?
		LIMIT 1
	`

	var gitLink domain.GitLink
	row := r.db.QueryRow(ctx, query, projectID, gitLinkID)
	err := row.Scan(
		&gitLink.ID,
		&gitLink.ProjectID,
		&gitLink.TraceID,
		&gitLink.CommitSha,
		&gitLink.ParentSha,
		&gitLink.Branch,
		&gitLink.Tag,
		&gitLink.RepoURL,
		&gitLink.CommitMessage,
		&gitLink.CommitAuthor,
		&gitLink.CommitAuthorEmail,
		&gitLink.CommitTimestamp,
		&gitLink.FilesAdded,
		&gitLink.FilesModified,
		&gitLink.FilesDeleted,
		&gitLink.FilesChangedCount,
		&gitLink.Additions,
		&gitLink.Deletions,
		&gitLink.LinkType,
		&gitLink.CIRunID,
		&gitLink.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &gitLink, nil
}

// GetByCommitSha retrieves all git links for a commit
func (r *GitLinkRepository) GetByCommitSha(ctx context.Context, projectID uuid.UUID, commitSha string) ([]domain.GitLink, error) {
	query := `
		SELECT
			id, project_id, trace_id, commit_sha, parent_sha, branch, tag,
			repo_url, commit_message, commit_author, commit_author_email,
			commit_timestamp, files_added, files_modified, files_deleted,
			files_changed_count, additions, deletions, link_type, ci_run_id, created_at
		FROM git_links FINAL
		WHERE project_id = ? AND commit_sha = ?
		ORDER BY created_at DESC
	`

	var gitLinks []domain.GitLink
	if err := r.db.Select(ctx, &gitLinks, query, projectID, commitSha); err != nil {
		return nil, err
	}

	return gitLinks, nil
}

// GetByTraceID retrieves all git links for a trace
func (r *GitLinkRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.GitLink, error) {
	query := `
		SELECT
			id, project_id, trace_id, commit_sha, parent_sha, branch, tag,
			repo_url, commit_message, commit_author, commit_author_email,
			commit_timestamp, files_added, files_modified, files_deleted,
			files_changed_count, additions, deletions, link_type, ci_run_id, created_at
		FROM git_links FINAL
		WHERE project_id = ? AND trace_id = ?
		ORDER BY commit_timestamp DESC
	`

	var gitLinks []domain.GitLink
	if err := r.db.Select(ctx, &gitLinks, query, projectID, traceID); err != nil {
		return nil, err
	}

	return gitLinks, nil
}

// List retrieves git links with filtering
func (r *GitLinkRepository) List(ctx context.Context, filter *domain.GitLinkFilter, limit, offset int) (*domain.GitLinkList, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{filter.ProjectID}

	if filter.TraceID != nil {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, *filter.TraceID)
	}

	if filter.CommitSha != nil {
		conditions = append(conditions, "commit_sha = ?")
		args = append(args, *filter.CommitSha)
	}

	if filter.Branch != nil {
		conditions = append(conditions, "branch = ?")
		args = append(args, *filter.Branch)
	}

	if filter.RepoURL != nil {
		conditions = append(conditions, "repo_url = ?")
		args = append(args, *filter.RepoURL)
	}

	if filter.LinkType != nil {
		conditions = append(conditions, "link_type = ?")
		args = append(args, string(*filter.LinkType))
	}

	if filter.CIRunID != nil {
		conditions = append(conditions, "ci_run_id = ?")
		args = append(args, *filter.CIRunID)
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *filter.ToTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get count
	countQuery := fmt.Sprintf("SELECT count() FROM git_links FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get git links
	query := fmt.Sprintf(`
		SELECT
			id, project_id, trace_id, commit_sha, parent_sha, branch, tag,
			repo_url, commit_message, commit_author, commit_author_email,
			commit_timestamp, files_added, files_modified, files_deleted,
			files_changed_count, additions, deletions, link_type, ci_run_id, created_at
		FROM git_links FINAL
		WHERE %s
		ORDER BY commit_timestamp DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var gitLinks []domain.GitLink
	if err := r.db.Select(ctx, &gitLinks, query, args...); err != nil {
		return nil, err
	}

	return &domain.GitLinkList{
		GitLinks:   gitLinks,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(gitLinks)) < totalCount,
	}, nil
}

// GetTimeline retrieves a git timeline for a project
func (r *GitLinkRepository) GetTimeline(ctx context.Context, projectID uuid.UUID, branch string, limit int) (*domain.GitTimeline, error) {
	query := `
		SELECT
			commit_sha,
			any(commit_message) AS commit_message,
			any(commit_author) AS commit_author,
			any(commit_timestamp) AS commit_time,
			any(branch) AS branch,
			count(DISTINCT trace_id) AS trace_count,
			groupArray(DISTINCT trace_id) AS trace_ids
		FROM git_links FINAL
		WHERE project_id = ? AND branch = ?
		GROUP BY commit_sha
		ORDER BY commit_time DESC
		LIMIT ?
	`

	var entries []domain.GitTimelineEntry
	if err := r.db.Select(ctx, &entries, query, projectID, branch, limit); err != nil {
		return nil, err
	}

	return &domain.GitTimeline{Commits: entries}, nil
}
