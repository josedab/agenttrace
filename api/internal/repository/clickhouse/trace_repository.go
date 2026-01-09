package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// TraceRepository handles trace data operations in ClickHouse
type TraceRepository struct {
	db *database.ClickHouseDB
}

// NewTraceRepository creates a new trace repository
func NewTraceRepository(db *database.ClickHouseDB) *TraceRepository {
	return &TraceRepository{db: db}
}

// Create inserts a new trace
func (r *TraceRepository) Create(ctx context.Context, trace *domain.Trace) error {
	query := `
		INSERT INTO traces (
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time,
			input, output, level, status_message, total_cost, input_cost,
			output_cost, total_tokens, input_tokens, output_tokens,
			git_commit_sha, git_branch, git_repo_url, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		trace.ID,
		trace.ProjectID,
		trace.Name,
		trace.UserID,
		trace.SessionID,
		trace.Release,
		trace.Version,
		trace.Tags,
		trace.Metadata,
		trace.Public,
		trace.Bookmarked,
		trace.StartTime,
		trace.EndTime,
		trace.Input,
		trace.Output,
		string(trace.Level),
		trace.StatusMessage,
		trace.TotalCost,
		trace.InputCost,
		trace.OutputCost,
		trace.TotalTokens,
		trace.InputTokens,
		trace.OutputTokens,
		trace.GitCommitSha,
		trace.GitBranch,
		trace.GitRepoURL,
		trace.CreatedAt,
		trace.UpdatedAt,
	)
}

// CreateBatch inserts multiple traces
func (r *TraceRepository) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO traces (
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time,
			input, output, level, status_message, total_cost, input_cost,
			output_cost, total_tokens, input_tokens, output_tokens,
			git_commit_sha, git_branch, git_repo_url, created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, trace := range traces {
		if err := batch.Append(
			trace.ID,
			trace.ProjectID,
			trace.Name,
			trace.UserID,
			trace.SessionID,
			trace.Release,
			trace.Version,
			trace.Tags,
			trace.Metadata,
			trace.Public,
			trace.Bookmarked,
			trace.StartTime,
			trace.EndTime,
			trace.Input,
			trace.Output,
			string(trace.Level),
			trace.StatusMessage,
			trace.TotalCost,
			trace.InputCost,
			trace.OutputCost,
			trace.TotalTokens,
			trace.InputTokens,
			trace.OutputTokens,
			trace.GitCommitSha,
			trace.GitBranch,
			trace.GitRepoURL,
			trace.CreatedAt,
			trace.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	return batch.Send()
}

// GetByID retrieves a trace by ID
func (r *TraceRepository) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	var trace domain.Trace

	query := `
		SELECT
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time, duration_ms,
			input, output, level, status_message, total_cost, input_cost,
			output_cost, total_tokens, input_tokens, output_tokens,
			git_commit_sha, git_branch, git_repo_url, created_at, updated_at
		FROM traces FINAL
		WHERE project_id = ? AND id = ?
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, projectID, traceID)
	err := row.Scan(
		&trace.ID,
		&trace.ProjectID,
		&trace.Name,
		&trace.UserID,
		&trace.SessionID,
		&trace.Release,
		&trace.Version,
		&trace.Tags,
		&trace.Metadata,
		&trace.Public,
		&trace.Bookmarked,
		&trace.StartTime,
		&trace.EndTime,
		&trace.DurationMs,
		&trace.Input,
		&trace.Output,
		&trace.Level,
		&trace.StatusMessage,
		&trace.TotalCost,
		&trace.InputCost,
		&trace.OutputCost,
		&trace.TotalTokens,
		&trace.InputTokens,
		&trace.OutputTokens,
		&trace.GitCommitSha,
		&trace.GitBranch,
		&trace.GitRepoURL,
		&trace.CreatedAt,
		&trace.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &trace, nil
}

// List retrieves traces with filtering and pagination
func (r *TraceRepository) List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	// Build WHERE clause
	conditions := []string{"project_id = ?"}
	args := []interface{}{filter.ProjectID}

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.SessionID != nil {
		conditions = append(conditions, "session_id = ?")
		args = append(args, *filter.SessionID)
	}

	if filter.Name != nil {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Release != nil {
		conditions = append(conditions, "release = ?")
		args = append(args, *filter.Release)
	}

	if filter.Level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, string(*filter.Level))
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "start_time >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "start_time <= ?")
		args = append(args, *filter.ToTime)
	}

	if filter.Bookmarked != nil {
		conditions = append(conditions, "bookmarked = ?")
		args = append(args, *filter.Bookmarked)
	}

	if filter.HasError != nil && *filter.HasError {
		conditions = append(conditions, "level = 'ERROR'")
	}

	if len(filter.Tags) > 0 {
		conditions = append(conditions, "hasAny(tags, ?)")
		args = append(args, filter.Tags)
	}

	if len(filter.IDs) > 0 {
		placeholders := make([]string, len(filter.IDs))
		for i := range filter.IDs {
			placeholders[i] = "?"
			args = append(args, filter.IDs[i])
		}
		conditions = append(conditions, fmt.Sprintf("id IN (%s)", strings.Join(placeholders, ",")))
	}

	// Git correlation filters
	if filter.GitCommitSha != nil {
		conditions = append(conditions, "git_commit_sha = ?")
		args = append(args, *filter.GitCommitSha)
	}

	if filter.GitBranch != nil {
		conditions = append(conditions, "git_branch = ?")
		args = append(args, *filter.GitBranch)
	}

	if filter.GitRepoURL != nil {
		conditions = append(conditions, "git_repo_url = ?")
		args = append(args, *filter.GitRepoURL)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT count() FROM traces FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count traces: %w", err)
	}

	// Get traces
	query := fmt.Sprintf(`
		SELECT
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time, duration_ms,
			input, output, level, status_message, total_cost, input_cost,
			output_cost, total_tokens, input_tokens, output_tokens,
			git_commit_sha, git_branch, git_repo_url, created_at, updated_at
		FROM traces FINAL
		WHERE %s
		ORDER BY start_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit+1, offset)

	var traces []domain.Trace
	if err := r.db.Select(ctx, &traces, query, args...); err != nil {
		return nil, fmt.Errorf("failed to select traces: %w", err)
	}

	hasMore := len(traces) > limit
	if hasMore {
		traces = traces[:limit]
	}

	return &domain.TraceList{
		Traces:     traces,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// Update updates a trace
func (r *TraceRepository) Update(ctx context.Context, trace *domain.Trace) error {
	trace.UpdatedAt = time.Now()
	return r.Create(ctx, trace) // ReplacingMergeTree handles updates
}

// UpdateCosts updates trace costs
func (r *TraceRepository) UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error {
	query := `
		INSERT INTO traces (
			id, project_id, total_cost, input_cost, output_cost, updated_at
		)
		SELECT
			id, project_id, ?, ?, ?, now64(3)
		FROM traces FINAL
		WHERE id = ? AND project_id = ?
	`

	return r.db.Exec(ctx, query,
		totalCost, inputCost, outputCost,
		traceID, projectID,
	)
}

// SetBookmark sets the bookmark status of a trace
func (r *TraceRepository) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	query := `
		INSERT INTO traces (id, project_id, bookmarked, updated_at)
		SELECT id, project_id, ?, now64(3)
		FROM traces FINAL
		WHERE id = ? AND project_id = ?
	`

	return r.db.Exec(ctx, query, bookmarked, traceID, projectID)
}

// Delete marks a trace for deletion (soft delete via update)
func (r *TraceRepository) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	// In ClickHouse with ReplacingMergeTree, we can't truly delete
	// Instead, we could add a deleted flag or move to a different table
	// For now, this is a no-op as we rely on retention policies
	return nil
}

// GetBySessionID retrieves all traces for a session
func (r *TraceRepository) GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	query := `
		SELECT
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time, duration_ms,
			input, output, level, status_message, total_cost, input_cost,
			output_cost, total_tokens, input_tokens, output_tokens,
			git_commit_sha, git_branch, git_repo_url, created_at, updated_at
		FROM traces FINAL
		WHERE project_id = ? AND session_id = ?
		ORDER BY start_time ASC
	`

	var traces []domain.Trace
	if err := r.db.Select(ctx, &traces, query, projectID, sessionID); err != nil {
		return nil, err
	}

	return traces, nil
}
