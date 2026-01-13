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

// ObservationRepository handles observation data operations in ClickHouse
type ObservationRepository struct {
	db     *database.ClickHouseDB
	logger *zap.Logger
}

// NewObservationRepository creates a new observation repository
func NewObservationRepository(db *database.ClickHouseDB, logger *zap.Logger) *ObservationRepository {
	return &ObservationRepository{
		db:     db,
		logger: logger.Named("observation_repository"),
	}
}

// Create inserts a new observation
func (r *ObservationRepository) Create(ctx context.Context, obs *domain.Observation) error {
	r.logger.Debug("creating observation",
		zap.String("observation_id", obs.ID),
		zap.String("trace_id", obs.TraceID),
		zap.String("project_id", obs.ProjectID.String()),
		zap.String("type", string(obs.Type)),
	)

	query := `
		INSERT INTO observations (
			id, trace_id, project_id, parent_observation_id, type, name,
			level, status_message, metadata, start_time, end_time,
			completion_start_time, input, output, model, model_parameters,
			usage_input_tokens, usage_output_tokens, usage_total_tokens,
			usage_cache_read_tokens, usage_cache_creation_tokens,
			input_cost, output_cost, total_cost,
			prompt_id, prompt_version, prompt_name, version,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.db.Exec(ctx, query,
		obs.ID,
		obs.TraceID,
		obs.ProjectID,
		obs.ParentObservationID,
		string(obs.Type),
		obs.Name,
		string(obs.Level),
		obs.StatusMessage,
		obs.Metadata,
		obs.StartTime,
		obs.EndTime,
		obs.CompletionStartTime,
		obs.Input,
		obs.Output,
		obs.Model,
		obs.ModelParameters,
		obs.UsageDetails.InputTokens,
		obs.UsageDetails.OutputTokens,
		obs.UsageDetails.TotalTokens,
		obs.UsageDetails.CacheReadTokens,
		obs.UsageDetails.CacheCreationTokens,
		obs.CostDetails.InputCost,
		obs.CostDetails.OutputCost,
		obs.CostDetails.TotalCost,
		obs.PromptID,
		obs.PromptVersion,
		obs.PromptName,
		obs.Version,
		obs.CreatedAt,
		obs.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create observation",
			zap.String("observation_id", obs.ID),
			zap.String("trace_id", obs.TraceID),
			zap.Error(err),
		)
	}
	return err
}

// CreateBatch inserts multiple observations
func (r *ObservationRepository) CreateBatch(ctx context.Context, observations []*domain.Observation) error {
	if len(observations) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO observations (
			id, trace_id, project_id, parent_observation_id, type, name,
			level, status_message, metadata, start_time, end_time,
			completion_start_time, input, output, model, model_parameters,
			usage_input_tokens, usage_output_tokens, usage_total_tokens,
			usage_cache_read_tokens, usage_cache_creation_tokens,
			input_cost, output_cost, total_cost,
			prompt_id, prompt_version, prompt_name, version,
			created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, obs := range observations {
		if err := batch.Append(
			obs.ID,
			obs.TraceID,
			obs.ProjectID,
			obs.ParentObservationID,
			string(obs.Type),
			obs.Name,
			string(obs.Level),
			obs.StatusMessage,
			obs.Metadata,
			obs.StartTime,
			obs.EndTime,
			obs.CompletionStartTime,
			obs.Input,
			obs.Output,
			obs.Model,
			obs.ModelParameters,
			obs.UsageDetails.InputTokens,
			obs.UsageDetails.OutputTokens,
			obs.UsageDetails.TotalTokens,
			obs.UsageDetails.CacheReadTokens,
			obs.UsageDetails.CacheCreationTokens,
			obs.CostDetails.InputCost,
			obs.CostDetails.OutputCost,
			obs.CostDetails.TotalCost,
			obs.PromptID,
			obs.PromptVersion,
			obs.PromptName,
			obs.Version,
			obs.CreatedAt,
			obs.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	return batch.Send()
}

// GetByID retrieves an observation by ID
func (r *ObservationRepository) GetByID(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error) {
	query := `
		SELECT
			id, trace_id, project_id, parent_observation_id, type, name,
			level, status_message, metadata, start_time, end_time,
			completion_start_time, duration_ms, time_to_first_token_ms,
			input, output, model, model_parameters,
			usage_input_tokens, usage_output_tokens, usage_total_tokens,
			usage_cache_read_tokens, usage_cache_creation_tokens,
			input_cost, output_cost, total_cost,
			prompt_id, prompt_version, prompt_name, version,
			created_at, updated_at
		FROM observations FINAL
		WHERE project_id = ? AND id = ?
		LIMIT 1
	`

	var obs domain.Observation
	row := r.db.QueryRow(ctx, query, projectID, observationID)
	err := row.Scan(
		&obs.ID,
		&obs.TraceID,
		&obs.ProjectID,
		&obs.ParentObservationID,
		&obs.Type,
		&obs.Name,
		&obs.Level,
		&obs.StatusMessage,
		&obs.Metadata,
		&obs.StartTime,
		&obs.EndTime,
		&obs.CompletionStartTime,
		&obs.DurationMs,
		&obs.TimeToFirstTokenMs,
		&obs.Input,
		&obs.Output,
		&obs.Model,
		&obs.ModelParameters,
		&obs.UsageDetails.InputTokens,
		&obs.UsageDetails.OutputTokens,
		&obs.UsageDetails.TotalTokens,
		&obs.UsageDetails.CacheReadTokens,
		&obs.UsageDetails.CacheCreationTokens,
		&obs.CostDetails.InputCost,
		&obs.CostDetails.OutputCost,
		&obs.CostDetails.TotalCost,
		&obs.PromptID,
		&obs.PromptVersion,
		&obs.PromptName,
		&obs.Version,
		&obs.CreatedAt,
		&obs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &obs, nil
}

// GetByTraceID retrieves all observations for a trace
func (r *ObservationRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error) {
	query := `
		SELECT
			id, trace_id, project_id, parent_observation_id, type, name,
			level, status_message, metadata, start_time, end_time,
			completion_start_time, duration_ms, time_to_first_token_ms,
			input, output, model, model_parameters,
			usage_input_tokens, usage_output_tokens, usage_total_tokens,
			usage_cache_read_tokens, usage_cache_creation_tokens,
			input_cost, output_cost, total_cost,
			prompt_id, prompt_version, prompt_name, version,
			created_at, updated_at
		FROM observations FINAL
		WHERE project_id = ? AND trace_id = ?
		ORDER BY start_time ASC
	`

	var observations []domain.Observation
	if err := r.db.Select(ctx, &observations, query, projectID, traceID); err != nil {
		return nil, err
	}

	return observations, nil
}

// List retrieves observations with filtering
func (r *ObservationRepository) List(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{filter.ProjectID}

	if filter.TraceID != nil {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, *filter.TraceID)
	}

	if filter.ParentObservationID != nil {
		conditions = append(conditions, "parent_observation_id = ?")
		args = append(args, *filter.ParentObservationID)
	}

	if filter.Type != nil {
		conditions = append(conditions, "type = ?")
		args = append(args, string(*filter.Type))
	}

	if filter.Name != nil {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Model != nil {
		conditions = append(conditions, "model = ?")
		args = append(args, *filter.Model)
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

	whereClause := strings.Join(conditions, " AND ")

	// Get count
	countQuery := fmt.Sprintf("SELECT count() FROM observations FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Get observations
	query := fmt.Sprintf(`
		SELECT
			id, trace_id, project_id, parent_observation_id, type, name,
			level, status_message, metadata, start_time, end_time,
			completion_start_time, duration_ms, time_to_first_token_ms,
			input, output, model, model_parameters,
			usage_input_tokens, usage_output_tokens, usage_total_tokens,
			usage_cache_read_tokens, usage_cache_creation_tokens,
			input_cost, output_cost, total_cost,
			prompt_id, prompt_version, prompt_name, version,
			created_at, updated_at
		FROM observations FINAL
		WHERE %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var observations []domain.Observation
	if err := r.db.Select(ctx, &observations, query, args...); err != nil {
		return nil, 0, err
	}

	return observations, totalCount, nil
}

// Update updates an observation
func (r *ObservationRepository) Update(ctx context.Context, obs *domain.Observation) error {
	obs.UpdatedAt = time.Now()
	return r.Create(ctx, obs) // ReplacingMergeTree handles updates
}

// UpdateCosts updates observation costs
func (r *ObservationRepository) UpdateCosts(ctx context.Context, projectID uuid.UUID, observationID string, inputCost, outputCost, totalCost float64) error {
	query := `
		INSERT INTO observations (
			id, project_id, input_cost, output_cost, total_cost, updated_at
		)
		SELECT
			id, project_id, ?, ?, ?, now64(3)
		FROM observations FINAL
		WHERE id = ? AND project_id = ?
	`

	return r.db.Exec(ctx, query,
		inputCost, outputCost, totalCost,
		observationID, projectID,
	)
}

// GetGenerationsWithoutCost retrieves generations that need cost calculation
func (r *ObservationRepository) GetGenerationsWithoutCost(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.Observation, error) {
	query := `
		SELECT
			id, trace_id, project_id, parent_observation_id, type, name,
			level, status_message, metadata, start_time, end_time,
			completion_start_time, duration_ms, time_to_first_token_ms,
			input, output, model, model_parameters,
			usage_input_tokens, usage_output_tokens, usage_total_tokens,
			usage_cache_read_tokens, usage_cache_creation_tokens,
			input_cost, output_cost, total_cost,
			prompt_id, prompt_version, prompt_name, version,
			created_at, updated_at
		FROM observations FINAL
		WHERE project_id = ?
			AND type = 'GENERATION'
			AND model != ''
			AND (usage_input_tokens > 0 OR usage_output_tokens > 0)
			AND total_cost = 0
		ORDER BY start_time DESC
		LIMIT ?
	`

	var observations []domain.Observation
	if err := r.db.Select(ctx, &observations, query, projectID, limit); err != nil {
		return nil, err
	}

	return observations, nil
}

// GetTree retrieves observations as a tree structure
func (r *ObservationRepository) GetTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error) {
	observations, err := r.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, err
	}

	trees := domain.BuildObservationTree(observations)
	if len(trees) == 0 {
		return nil, nil
	}
	return trees[0], nil
}

// CountBeforeCutoff counts observations created before the cutoff date for a project
func (r *ObservationRepository) CountBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	query := `
		SELECT count()
		FROM observations FINAL
		WHERE project_id = ? AND created_at < ?
	`

	var count int64
	row := r.db.QueryRow(ctx, query, projectID, cutoff)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count observations: %w", err)
	}

	return count, nil
}

// DeleteBeforeCutoff deletes observations created before the cutoff date for a project
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *ObservationRepository) DeleteBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	// First count how many we'll delete
	count, err := r.CountBeforeCutoff(ctx, projectID, cutoff)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	// ClickHouse uses ALTER TABLE DELETE for mutations
	query := `ALTER TABLE observations DELETE WHERE project_id = ? AND created_at < ?`
	if err := r.db.Exec(ctx, query, projectID, cutoff); err != nil {
		return 0, fmt.Errorf("failed to delete observations: %w", err)
	}

	return count, nil
}

// DeleteByProjectID deletes all observations for a project
// Note: ClickHouse ALTER TABLE DELETE is a heavy operation, use with caution
func (r *ObservationRepository) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	query := `ALTER TABLE observations DELETE WHERE project_id = ?`
	return r.db.Exec(ctx, query, projectID)
}

// CountOrphans counts observations that don't have a corresponding trace
func (r *ObservationRepository) CountOrphans(ctx context.Context) (int64, error) {
	query := `
		SELECT count()
		FROM observations o FINAL
		WHERE NOT EXISTS (
			SELECT 1 FROM traces t FINAL
			WHERE t.id = o.trace_id AND t.project_id = o.project_id
		)
	`

	var count int64
	row := r.db.QueryRow(ctx, query)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count orphan observations: %w", err)
	}

	return count, nil
}

// DeleteOrphans deletes observations that don't have a corresponding trace
func (r *ObservationRepository) DeleteOrphans(ctx context.Context) (int64, error) {
	count, err := r.CountOrphans(ctx)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	query := `
		ALTER TABLE observations DELETE
		WHERE (project_id, trace_id) NOT IN (
			SELECT project_id, id FROM traces FINAL
		)
	`
	if err := r.db.Exec(ctx, query); err != nil {
		return 0, fmt.Errorf("failed to delete orphan observations: %w", err)
	}

	return count, nil
}
