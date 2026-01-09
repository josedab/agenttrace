package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// FileOperationRepository handles file operation data in ClickHouse
type FileOperationRepository struct {
	db *database.ClickHouseDB
}

// NewFileOperationRepository creates a new file operation repository
func NewFileOperationRepository(db *database.ClickHouseDB) *FileOperationRepository {
	return &FileOperationRepository{db: db}
}

// Create inserts a new file operation
func (r *FileOperationRepository) Create(ctx context.Context, fileOp *domain.FileOperation) error {
	query := `
		INSERT INTO file_operations (
			id, project_id, trace_id, observation_id, operation, file_path,
			new_path, file_size, file_mode, content_hash, mime_type,
			lines_added, lines_removed, diff_preview, content_before_hash,
			content_after_hash, tool_name, reason, started_at, completed_at,
			duration_ms, success, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		fileOp.ID,
		fileOp.ProjectID,
		fileOp.TraceID,
		fileOp.ObservationID,
		string(fileOp.Operation),
		fileOp.FilePath,
		fileOp.NewPath,
		fileOp.FileSize,
		fileOp.FileMode,
		fileOp.ContentHash,
		fileOp.MimeType,
		fileOp.LinesAdded,
		fileOp.LinesRemoved,
		fileOp.DiffPreview,
		fileOp.ContentBeforeHash,
		fileOp.ContentAfterHash,
		fileOp.ToolName,
		fileOp.Reason,
		fileOp.StartedAt,
		fileOp.CompletedAt,
		fileOp.DurationMs,
		fileOp.Success,
		fileOp.ErrorMessage,
	)
}

// CreateBatch inserts multiple file operations
func (r *FileOperationRepository) CreateBatch(ctx context.Context, fileOps []*domain.FileOperation) error {
	if len(fileOps) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO file_operations (
			id, project_id, trace_id, observation_id, operation, file_path,
			new_path, file_size, file_mode, content_hash, mime_type,
			lines_added, lines_removed, diff_preview, content_before_hash,
			content_after_hash, tool_name, reason, started_at, completed_at,
			duration_ms, success, error_message
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, fileOp := range fileOps {
		if err := batch.Append(
			fileOp.ID,
			fileOp.ProjectID,
			fileOp.TraceID,
			fileOp.ObservationID,
			string(fileOp.Operation),
			fileOp.FilePath,
			fileOp.NewPath,
			fileOp.FileSize,
			fileOp.FileMode,
			fileOp.ContentHash,
			fileOp.MimeType,
			fileOp.LinesAdded,
			fileOp.LinesRemoved,
			fileOp.DiffPreview,
			fileOp.ContentBeforeHash,
			fileOp.ContentAfterHash,
			fileOp.ToolName,
			fileOp.Reason,
			fileOp.StartedAt,
			fileOp.CompletedAt,
			fileOp.DurationMs,
			fileOp.Success,
			fileOp.ErrorMessage,
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	return batch.Send()
}

// GetByTraceID retrieves all file operations for a trace
func (r *FileOperationRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.FileOperation, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, operation, file_path,
			new_path, file_size, file_mode, content_hash, mime_type,
			lines_added, lines_removed, diff_preview, content_before_hash,
			content_after_hash, tool_name, reason, started_at, completed_at,
			duration_ms, success, error_message
		FROM file_operations FINAL
		WHERE project_id = ? AND trace_id = ?
		ORDER BY started_at ASC
	`

	var fileOps []domain.FileOperation
	if err := r.db.Select(ctx, &fileOps, query, projectID, traceID); err != nil {
		return nil, err
	}

	return fileOps, nil
}

// List retrieves file operations with filtering
func (r *FileOperationRepository) List(ctx context.Context, filter *domain.FileOperationFilter, limit, offset int) (*domain.FileOperationList, error) {
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

	if filter.Operation != nil {
		conditions = append(conditions, "operation = ?")
		args = append(args, string(*filter.Operation))
	}

	if filter.FilePath != nil {
		conditions = append(conditions, "file_path LIKE ?")
		args = append(args, "%"+*filter.FilePath+"%")
	}

	if filter.Success != nil {
		conditions = append(conditions, "success = ?")
		args = append(args, *filter.Success)
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "started_at >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "started_at <= ?")
		args = append(args, *filter.ToTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get count
	countQuery := fmt.Sprintf("SELECT count() FROM file_operations FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get file operations
	query := fmt.Sprintf(`
		SELECT
			id, project_id, trace_id, observation_id, operation, file_path,
			new_path, file_size, file_mode, content_hash, mime_type,
			lines_added, lines_removed, diff_preview, content_before_hash,
			content_after_hash, tool_name, reason, started_at, completed_at,
			duration_ms, success, error_message
		FROM file_operations FINAL
		WHERE %s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var fileOps []domain.FileOperation
	if err := r.db.Select(ctx, &fileOps, query, args...); err != nil {
		return nil, err
	}

	return &domain.FileOperationList{
		FileOperations: fileOps,
		TotalCount:     totalCount,
		HasMore:        int64(offset+len(fileOps)) < totalCount,
	}, nil
}

// GetStats retrieves file operation statistics for a project
func (r *FileOperationRepository) GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.FileOperationStats, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{projectID}

	if traceID != nil {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, *traceID)
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT
			count() AS total_operations,
			countIf(operation = 'create') AS create_count,
			countIf(operation = 'read') AS read_count,
			countIf(operation = 'update') AS update_count,
			countIf(operation = 'delete') AS delete_count,
			countIf(success = true) AS success_count,
			countIf(success = false) AS failure_count,
			sum(lines_added) AS total_lines_added,
			sum(lines_removed) AS total_lines_removed
		FROM file_operations FINAL
		WHERE %s
	`, whereClause)

	var stats domain.FileOperationStats
	row := r.db.QueryRow(ctx, query, args...)
	err := row.Scan(
		&stats.TotalOperations,
		&stats.CreateCount,
		&stats.ReadCount,
		&stats.UpdateCount,
		&stats.DeleteCount,
		&stats.SuccessCount,
		&stats.FailureCount,
		&stats.TotalLinesAdded,
		&stats.TotalLinesRemoved,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
