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

// ScoreRepository handles score data operations in ClickHouse
type ScoreRepository struct {
	db *database.ClickHouseDB
}

// NewScoreRepository creates a new score repository
func NewScoreRepository(db *database.ClickHouseDB) *ScoreRepository {
	return &ScoreRepository{db: db}
}

// Create inserts a new score
func (r *ScoreRepository) Create(ctx context.Context, score *domain.Score) error {
	query := `
		INSERT INTO scores (
			id, project_id, trace_id, observation_id, name, source,
			data_type, value, string_value, comment, config_id,
			author_user_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		score.ID,
		score.ProjectID,
		score.TraceID,
		score.ObservationID,
		score.Name,
		string(score.Source),
		string(score.DataType),
		score.Value,
		score.StringValue,
		score.Comment,
		score.ConfigID,
		score.AuthorUserID,
		score.CreatedAt,
		score.UpdatedAt,
	)
}

// CreateBatch inserts multiple scores
func (r *ScoreRepository) CreateBatch(ctx context.Context, scores []*domain.Score) error {
	if len(scores) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO scores (
			id, project_id, trace_id, observation_id, name, source,
			data_type, value, string_value, comment, config_id,
			author_user_id, created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, score := range scores {
		if err := batch.Append(
			score.ID,
			score.ProjectID,
			score.TraceID,
			score.ObservationID,
			score.Name,
			string(score.Source),
			string(score.DataType),
			score.Value,
			score.StringValue,
			score.Comment,
			score.ConfigID,
			score.AuthorUserID,
			score.CreatedAt,
			score.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	return batch.Send()
}

// GetByID retrieves a score by ID
func (r *ScoreRepository) GetByID(ctx context.Context, projectID uuid.UUID, scoreID string) (*domain.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, name, source,
			data_type, value, string_value, comment, config_id,
			author_user_id, created_at, updated_at
		FROM scores FINAL
		WHERE project_id = ? AND id = ?
		LIMIT 1
	`

	var score domain.Score
	row := r.db.QueryRow(ctx, query, projectID, scoreID)
	err := row.Scan(
		&score.ID,
		&score.ProjectID,
		&score.TraceID,
		&score.ObservationID,
		&score.Name,
		&score.Source,
		&score.DataType,
		&score.Value,
		&score.StringValue,
		&score.Comment,
		&score.ConfigID,
		&score.AuthorUserID,
		&score.CreatedAt,
		&score.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &score, nil
}

// GetByTraceID retrieves all scores for a trace
func (r *ScoreRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, name, source,
			data_type, value, string_value, comment, config_id,
			author_user_id, created_at, updated_at
		FROM scores FINAL
		WHERE project_id = ? AND trace_id = ?
		ORDER BY created_at DESC
	`

	var scores []domain.Score
	if err := r.db.Select(ctx, &scores, query, projectID, traceID); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetByObservationID retrieves all scores for an observation
func (r *ScoreRepository) GetByObservationID(ctx context.Context, projectID uuid.UUID, observationID string) ([]domain.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, name, source,
			data_type, value, string_value, comment, config_id,
			author_user_id, created_at, updated_at
		FROM scores FINAL
		WHERE project_id = ? AND observation_id = ?
		ORDER BY created_at DESC
	`

	var scores []domain.Score
	if err := r.db.Select(ctx, &scores, query, projectID, observationID); err != nil {
		return nil, err
	}

	return scores, nil
}

// List retrieves scores with filtering
func (r *ScoreRepository) List(ctx context.Context, filter *domain.ScoreFilter, limit, offset int) (*domain.ScoreList, error) {
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

	if filter.Name != nil {
		conditions = append(conditions, "name = ?")
		args = append(args, *filter.Name)
	}

	if filter.Source != nil {
		conditions = append(conditions, "source = ?")
		args = append(args, string(*filter.Source))
	}

	if filter.DataType != nil {
		conditions = append(conditions, "data_type = ?")
		args = append(args, string(*filter.DataType))
	}

	if filter.ConfigID != nil {
		conditions = append(conditions, "config_id = ?")
		args = append(args, *filter.ConfigID)
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
	countQuery := fmt.Sprintf("SELECT count() FROM scores FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get scores
	query := fmt.Sprintf(`
		SELECT
			id, project_id, trace_id, observation_id, name, source,
			data_type, value, string_value, comment, config_id,
			author_user_id, created_at, updated_at
		FROM scores FINAL
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var scores []domain.Score
	if err := r.db.Select(ctx, &scores, query, args...); err != nil {
		return nil, err
	}

	return &domain.ScoreList{
		Scores:     scores,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(scores)) < totalCount,
	}, nil
}

// Update updates a score
func (r *ScoreRepository) Update(ctx context.Context, score *domain.Score) error {
	score.UpdatedAt = time.Now()
	return r.Create(ctx, score) // ReplacingMergeTree handles updates
}

// Delete deletes a score
func (r *ScoreRepository) Delete(ctx context.Context, projectID, scoreID uuid.UUID) error {
	// In ClickHouse, we rely on retention; this is a no-op
	return nil
}

// GetStats retrieves score statistics
func (r *ScoreRepository) GetStats(ctx context.Context, projectID uuid.UUID, name string) (*domain.ScoreStats, error) {
	query := `
		SELECT
			name,
			count() AS count,
			avg(value) AS avg_value,
			min(value) AS min_value,
			max(value) AS max_value,
			quantile(0.5)(value) AS median_value
		FROM scores FINAL
		WHERE project_id = ? AND name = ? AND data_type = 'NUMERIC' AND value IS NOT NULL
		GROUP BY name
	`

	var stats domain.ScoreStats
	row := r.db.QueryRow(ctx, query, projectID, name)
	err := row.Scan(
		&stats.Name,
		&stats.Count,
		&stats.AvgValue,
		&stats.MinValue,
		&stats.MaxValue,
		&stats.MedianValue,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetDistinctNames retrieves distinct score names for a project
func (r *ScoreRepository) GetDistinctNames(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT name
		FROM scores FINAL
		WHERE project_id = ?
		ORDER BY name
	`

	var names []string
	if err := r.db.Select(ctx, &names, query, projectID); err != nil {
		return nil, err
	}

	return names, nil
}

// CountBeforeCutoff counts scores created before the cutoff date for a project
func (r *ScoreRepository) CountBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	query := `
		SELECT count()
		FROM scores FINAL
		WHERE project_id = ? AND created_at < ?
	`

	var count int64
	row := r.db.QueryRow(ctx, query, projectID, cutoff)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count scores: %w", err)
	}

	return count, nil
}

// DeleteBeforeCutoff deletes scores created before the cutoff date for a project
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *ScoreRepository) DeleteBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	// First count how many we'll delete
	count, err := r.CountBeforeCutoff(ctx, projectID, cutoff)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	// ClickHouse uses ALTER TABLE DELETE for mutations
	query := `ALTER TABLE scores DELETE WHERE project_id = ? AND created_at < ?`
	if err := r.db.Exec(ctx, query, projectID, cutoff); err != nil {
		return 0, fmt.Errorf("failed to delete scores: %w", err)
	}

	return count, nil
}

// DeleteByProjectID deletes all scores for a project
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *ScoreRepository) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	query := `ALTER TABLE scores DELETE WHERE project_id = ?`
	return r.db.Exec(ctx, query, projectID)
}

// CountOrphans counts scores that don't have a corresponding trace
func (r *ScoreRepository) CountOrphans(ctx context.Context) (int64, error) {
	query := `
		SELECT count()
		FROM scores s FINAL
		WHERE NOT EXISTS (
			SELECT 1 FROM traces t FINAL
			WHERE t.id = s.trace_id AND t.project_id = s.project_id
		)
	`

	var count int64
	row := r.db.QueryRow(ctx, query)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count orphan scores: %w", err)
	}

	return count, nil
}

// DeleteOrphans deletes scores that don't have a corresponding trace
func (r *ScoreRepository) DeleteOrphans(ctx context.Context) (int64, error) {
	count, err := r.CountOrphans(ctx)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	query := `
		ALTER TABLE scores DELETE
		WHERE (project_id, trace_id) NOT IN (
			SELECT project_id, id FROM traces FINAL
		)
	`
	if err := r.db.Exec(ctx, query); err != nil {
		return 0, fmt.Errorf("failed to delete orphan scores: %w", err)
	}

	return count, nil
}
