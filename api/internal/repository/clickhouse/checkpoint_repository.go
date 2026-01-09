package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// CheckpointRepository handles checkpoint data operations in ClickHouse
type CheckpointRepository struct {
	db *database.ClickHouseDB
}

// NewCheckpointRepository creates a new checkpoint repository
func NewCheckpointRepository(db *database.ClickHouseDB) *CheckpointRepository {
	return &CheckpointRepository{db: db}
}

// Create inserts a new checkpoint
func (r *CheckpointRepository) Create(ctx context.Context, checkpoint *domain.Checkpoint) error {
	query := `
		INSERT INTO checkpoints (
			id, project_id, trace_id, observation_id, name, description,
			checkpoint_type, git_commit_sha, git_branch, git_repo_url,
			files_snapshot, files_changed, storage_path, total_files,
			total_size_bytes, restored_from, restored_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		checkpoint.ID,
		checkpoint.ProjectID,
		checkpoint.TraceID,
		checkpoint.ObservationID,
		checkpoint.Name,
		checkpoint.Description,
		string(checkpoint.Type),
		checkpoint.GitCommitSha,
		checkpoint.GitBranch,
		checkpoint.GitRepoURL,
		checkpoint.FilesSnapshot,
		checkpoint.FilesChanged,
		checkpoint.StoragePath,
		checkpoint.TotalFiles,
		checkpoint.TotalSizeBytes,
		checkpoint.RestoredFrom,
		checkpoint.RestoredAt,
		checkpoint.CreatedAt,
	)
}

// GetByID retrieves a checkpoint by ID
func (r *CheckpointRepository) GetByID(ctx context.Context, projectID, checkpointID uuid.UUID) (*domain.Checkpoint, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, name, description,
			checkpoint_type, git_commit_sha, git_branch, git_repo_url,
			files_snapshot, files_changed, storage_path, total_files,
			total_size_bytes, restored_from, restored_at, created_at
		FROM checkpoints FINAL
		WHERE project_id = ? AND id = ?
		LIMIT 1
	`

	var checkpoint domain.Checkpoint
	row := r.db.QueryRow(ctx, query, projectID, checkpointID)
	err := row.Scan(
		&checkpoint.ID,
		&checkpoint.ProjectID,
		&checkpoint.TraceID,
		&checkpoint.ObservationID,
		&checkpoint.Name,
		&checkpoint.Description,
		&checkpoint.Type,
		&checkpoint.GitCommitSha,
		&checkpoint.GitBranch,
		&checkpoint.GitRepoURL,
		&checkpoint.FilesSnapshot,
		&checkpoint.FilesChanged,
		&checkpoint.StoragePath,
		&checkpoint.TotalFiles,
		&checkpoint.TotalSizeBytes,
		&checkpoint.RestoredFrom,
		&checkpoint.RestoredAt,
		&checkpoint.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &checkpoint, nil
}

// GetByTraceID retrieves all checkpoints for a trace
func (r *CheckpointRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Checkpoint, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, name, description,
			checkpoint_type, git_commit_sha, git_branch, git_repo_url,
			files_snapshot, files_changed, storage_path, total_files,
			total_size_bytes, restored_from, restored_at, created_at
		FROM checkpoints FINAL
		WHERE project_id = ? AND trace_id = ?
		ORDER BY created_at DESC
	`

	var checkpoints []domain.Checkpoint
	if err := r.db.Select(ctx, &checkpoints, query, projectID, traceID); err != nil {
		return nil, err
	}

	return checkpoints, nil
}

// List retrieves checkpoints with filtering
func (r *CheckpointRepository) List(ctx context.Context, filter *domain.CheckpointFilter, limit, offset int) (*domain.CheckpointList, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{filter.ProjectID}

	if filter.TraceID != nil {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, *filter.TraceID)
	}

	if filter.ObservationID != nil {
		conditions = append(conditions, "observation_id = ?")
		args = append(args, *filter.ObservationID)
	}

	if filter.Type != nil {
		conditions = append(conditions, "checkpoint_type = ?")
		args = append(args, string(*filter.Type))
	}

	if filter.GitCommitSha != nil {
		conditions = append(conditions, "git_commit_sha = ?")
		args = append(args, *filter.GitCommitSha)
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
	countQuery := fmt.Sprintf("SELECT count() FROM checkpoints FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get checkpoints
	query := fmt.Sprintf(`
		SELECT
			id, project_id, trace_id, observation_id, name, description,
			checkpoint_type, git_commit_sha, git_branch, git_repo_url,
			files_snapshot, files_changed, storage_path, total_files,
			total_size_bytes, restored_from, restored_at, created_at
		FROM checkpoints FINAL
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var checkpoints []domain.Checkpoint
	if err := r.db.Select(ctx, &checkpoints, query, args...); err != nil {
		return nil, err
	}

	return &domain.CheckpointList{
		Checkpoints: checkpoints,
		TotalCount:  totalCount,
		HasMore:     int64(offset+len(checkpoints)) < totalCount,
	}, nil
}
