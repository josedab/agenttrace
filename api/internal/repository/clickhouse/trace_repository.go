package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// TraceRepository handles trace data operations in ClickHouse
type TraceRepository struct {
	db     *database.ClickHouseDB
	logger *zap.Logger
}

// NewTraceRepository creates a new trace repository
func NewTraceRepository(db *database.ClickHouseDB, logger *zap.Logger) *TraceRepository {
	return &TraceRepository{
		db:     db,
		logger: logger.Named("trace_repository"),
	}
}

// Create inserts a new trace
func (r *TraceRepository) Create(ctx context.Context, trace *domain.Trace) error {
	r.logger.Debug("creating trace",
		zap.String("trace_id", trace.ID),
		zap.String("project_id", trace.ProjectID.String()),
	)

	query := `
		INSERT INTO traces (
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time,
			input, output, level, status_message, total_cost, input_cost,
			output_cost, total_tokens, input_tokens, output_tokens,
			git_commit_sha, git_branch, git_repo_url, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.db.Exec(ctx, query,
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
	if err != nil {
		r.logger.Error("failed to create trace",
			zap.String("trace_id", trace.ID),
			zap.String("project_id", trace.ProjectID.String()),
			zap.Error(err),
		)
	}
	return err
}

// CreateBatch inserts multiple traces
func (r *TraceRepository) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	if len(traces) == 0 {
		r.logger.Debug("skipping empty batch insert")
		return nil
	}

	r.logger.Debug("creating traces batch", zap.Int("count", len(traces)))

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
		r.logger.Error("failed to prepare batch", zap.Error(err))
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
			r.logger.Error("failed to append to batch",
				zap.String("trace_id", trace.ID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		r.logger.Error("failed to send batch", zap.Int("count", len(traces)), zap.Error(err))
		return err
	}
	return nil
}

// GetByID retrieves a trace by ID
func (r *TraceRepository) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	r.logger.Debug("getting trace by ID",
		zap.String("trace_id", traceID),
		zap.String("project_id", projectID.String()),
	)

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
		r.logger.Warn("trace not found or error",
			zap.String("trace_id", traceID),
			zap.String("project_id", projectID.String()),
			zap.Error(err),
		)
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

// Delete deletes a trace by ID
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *TraceRepository) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	r.logger.Info("deleting trace",
		zap.String("trace_id", traceID),
		zap.String("project_id", projectID.String()),
	)

	query := `ALTER TABLE traces DELETE WHERE project_id = ? AND id = ?`
	if err := r.db.Exec(ctx, query, projectID, traceID); err != nil {
		r.logger.Error("failed to delete trace",
			zap.String("trace_id", traceID),
			zap.String("project_id", projectID.String()),
			zap.Error(err),
		)
		return err
	}
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

// CountBeforeCutoff counts traces created before the cutoff date for a project
func (r *TraceRepository) CountBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	query := `
		SELECT count()
		FROM traces FINAL
		WHERE project_id = ? AND created_at < ?
	`

	var count int64
	row := r.db.QueryRow(ctx, query, projectID, cutoff)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count traces: %w", err)
	}

	return count, nil
}

// DeleteBeforeCutoff deletes traces created before the cutoff date for a project
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *TraceRepository) DeleteBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	r.logger.Info("deleting traces before cutoff",
		zap.String("project_id", projectID.String()),
		zap.Time("cutoff", cutoff),
	)

	// First count how many we'll delete
	count, err := r.CountBeforeCutoff(ctx, projectID, cutoff)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		r.logger.Debug("no traces to delete before cutoff",
			zap.String("project_id", projectID.String()),
			zap.Time("cutoff", cutoff),
		)
		return 0, nil
	}

	// ClickHouse uses ALTER TABLE DELETE for mutations
	query := `ALTER TABLE traces DELETE WHERE project_id = ? AND created_at < ?`
	if err := r.db.Exec(ctx, query, projectID, cutoff); err != nil {
		r.logger.Error("failed to delete traces before cutoff",
			zap.String("project_id", projectID.String()),
			zap.Time("cutoff", cutoff),
			zap.Int64("count", count),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to delete traces: %w", err)
	}

	r.logger.Info("deleted traces before cutoff",
		zap.String("project_id", projectID.String()),
		zap.Time("cutoff", cutoff),
		zap.Int64("count", count),
	)
	return count, nil
}

// DeleteByProjectID deletes all traces for a project
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *TraceRepository) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	r.logger.Info("deleting all traces for project",
		zap.String("project_id", projectID.String()),
	)

	query := `ALTER TABLE traces DELETE WHERE project_id = ?`
	if err := r.db.Exec(ctx, query, projectID); err != nil {
		r.logger.Error("failed to delete all traces for project",
			zap.String("project_id", projectID.String()),
			zap.Error(err),
		)
		return err
	}
	return nil
}
