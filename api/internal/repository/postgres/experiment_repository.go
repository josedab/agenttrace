package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ExperimentRepository handles experiment data operations in PostgreSQL
type ExperimentRepository struct {
	db *database.PostgresDB
}

// NewExperimentRepository creates a new experiment repository
func NewExperimentRepository(db *database.PostgresDB) *ExperimentRepository {
	return &ExperimentRepository{db: db}
}

// Create creates a new experiment with its variants
func (r *ExperimentRepository) Create(ctx context.Context, experiment *domain.Experiment) error {
	return database.Transaction(ctx, r.db, func(tx pgx.Tx) error {
		// Marshal JSON fields
		userIDFilter, err := json.Marshal(experiment.UserIDFilter)
		if err != nil {
			return fmt.Errorf("failed to marshal user_id_filter: %w", err)
		}

		metadataFilters, err := json.Marshal(experiment.MetadataFilters)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata_filters: %w", err)
		}

		// Insert experiment
		query := `
			INSERT INTO experiments (
				id, project_id, name, description, status, target_metric, target_goal,
				traffic_percent, trace_name_filter, user_id_filter, metadata_filters,
				min_duration_hours, min_samples_per_variant, started_at, ended_at,
				statistical_power, created_at, updated_at, created_by
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		`

		_, err = tx.Exec(ctx, query,
			experiment.ID,
			experiment.ProjectID,
			experiment.Name,
			experiment.Description,
			string(experiment.Status),
			experiment.TargetMetric,
			experiment.TargetGoal,
			experiment.TrafficPercent,
			experiment.TraceNameFilter,
			userIDFilter,
			metadataFilters,
			experiment.MinDuration,
			experiment.MinSamples,
			experiment.StartedAt,
			experiment.EndedAt,
			experiment.StatisticalPower,
			experiment.CreatedAt,
			experiment.UpdatedAt,
			experiment.CreatedBy,
		)
		if err != nil {
			if strings.Contains(err.Error(), "unique_experiment_name_per_project") {
				return apperrors.Conflict("experiment with this name already exists in the project")
			}
			return fmt.Errorf("failed to create experiment: %w", err)
		}

		// Insert variants
		for _, variant := range experiment.Variants {
			if err := r.createVariant(ctx, tx, &variant); err != nil {
				return err
			}
		}

		return nil
	})
}

// createVariant inserts a single variant within a transaction
func (r *ExperimentRepository) createVariant(ctx context.Context, tx pgx.Tx, variant *domain.ExperimentVariant) error {
	configJSON, err := json.Marshal(variant.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal variant config: %w", err)
	}

	query := `
		INSERT INTO experiment_variants (
			id, experiment_id, name, description, weight, is_control, config, sample_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = tx.Exec(ctx, query,
		variant.ID,
		variant.ExperimentID,
		variant.Name,
		variant.Description,
		variant.Weight,
		variant.IsControl,
		configJSON,
		variant.SampleCount,
	)
	if err != nil {
		return fmt.Errorf("failed to create variant: %w", err)
	}

	return nil
}

// GetByID retrieves an experiment by ID with its variants
func (r *ExperimentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Experiment, error) {
	query := `
		SELECT id, project_id, name, description, status, target_metric, target_goal,
			traffic_percent, trace_name_filter, user_id_filter, metadata_filters,
			min_duration_hours, min_samples_per_variant, started_at, ended_at,
			winning_variant_id, results, statistical_power, created_at, updated_at, created_by
		FROM experiments
		WHERE id = $1
	`

	var experiment domain.Experiment
	var status string
	var userIDFilter, metadataFilters, resultsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&experiment.ID,
		&experiment.ProjectID,
		&experiment.Name,
		&experiment.Description,
		&status,
		&experiment.TargetMetric,
		&experiment.TargetGoal,
		&experiment.TrafficPercent,
		&experiment.TraceNameFilter,
		&userIDFilter,
		&metadataFilters,
		&experiment.MinDuration,
		&experiment.MinSamples,
		&experiment.StartedAt,
		&experiment.EndedAt,
		&experiment.WinningVariant,
		&resultsJSON,
		&experiment.StatisticalPower,
		&experiment.CreatedAt,
		&experiment.UpdatedAt,
		&experiment.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("experiment not found")
		}
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	experiment.Status = domain.ExperimentStatus(status)

	// Unmarshal JSON fields
	if len(userIDFilter) > 0 {
		if err := json.Unmarshal(userIDFilter, &experiment.UserIDFilter); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user_id_filter: %w", err)
		}
	}

	if len(metadataFilters) > 0 {
		if err := json.Unmarshal(metadataFilters, &experiment.MetadataFilters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata_filters: %w", err)
		}
	}

	if len(resultsJSON) > 0 {
		experiment.Results = &domain.ExperimentResults{}
		if err := json.Unmarshal(resultsJSON, experiment.Results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}
	}

	// Fetch variants
	variants, err := r.getVariantsByExperimentID(ctx, experiment.ID)
	if err != nil {
		return nil, err
	}
	experiment.Variants = variants

	return &experiment, nil
}

// getVariantsByExperimentID retrieves all variants for an experiment
func (r *ExperimentRepository) getVariantsByExperimentID(ctx context.Context, experimentID uuid.UUID) ([]domain.ExperimentVariant, error) {
	query := `
		SELECT id, experiment_id, name, description, weight, is_control, config,
			sample_count, metric_mean, metric_std_dev, metric_min, metric_max
		FROM experiment_variants
		WHERE experiment_id = $1
		ORDER BY is_control DESC, name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query variants: %w", err)
	}
	defer rows.Close()

	var variants []domain.ExperimentVariant
	for rows.Next() {
		var variant domain.ExperimentVariant
		var configJSON []byte

		err := rows.Scan(
			&variant.ID,
			&variant.ExperimentID,
			&variant.Name,
			&variant.Description,
			&variant.Weight,
			&variant.IsControl,
			&configJSON,
			&variant.SampleCount,
			&variant.MetricMean,
			&variant.MetricStdDev,
			&variant.MetricMin,
			&variant.MetricMax,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}

		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &variant.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variant config: %w", err)
			}
		}

		variants = append(variants, variant)
	}

	return variants, nil
}

// List retrieves experiments for a project with filtering
func (r *ExperimentRepository) List(ctx context.Context, filter domain.ExperimentFilter, limit, offset int) (*domain.ExperimentList, error) {
	// Build query conditions
	conditions := []string{"project_id = $1"}
	args := []any{filter.ProjectID}
	argCount := 1

	if filter.Status != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, string(*filter.Status))
	}

	if filter.Search != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+filter.Search+"%")
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM experiments WHERE %s", whereClause)
	var totalCount int64
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count experiments: %w", err)
	}

	// Fetch experiments
	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, status, target_metric, target_goal,
			traffic_percent, trace_name_filter, started_at, ended_at,
			winning_variant_id, statistical_power, created_at, updated_at, created_by
		FROM experiments
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount+1, argCount+2)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query experiments: %w", err)
	}
	defer rows.Close()

	var experiments []domain.Experiment
	for rows.Next() {
		var experiment domain.Experiment
		var status string

		err := rows.Scan(
			&experiment.ID,
			&experiment.ProjectID,
			&experiment.Name,
			&experiment.Description,
			&status,
			&experiment.TargetMetric,
			&experiment.TargetGoal,
			&experiment.TrafficPercent,
			&experiment.TraceNameFilter,
			&experiment.StartedAt,
			&experiment.EndedAt,
			&experiment.WinningVariant,
			&experiment.StatisticalPower,
			&experiment.CreatedAt,
			&experiment.UpdatedAt,
			&experiment.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experiment: %w", err)
		}

		experiment.Status = domain.ExperimentStatus(status)

		// Fetch variants for each experiment
		variants, err := r.getVariantsByExperimentID(ctx, experiment.ID)
		if err != nil {
			return nil, err
		}
		experiment.Variants = variants

		experiments = append(experiments, experiment)
	}

	return &domain.ExperimentList{
		Experiments: experiments,
		TotalCount:  totalCount,
		HasMore:     int64(offset+len(experiments)) < totalCount,
	}, nil
}

// Update updates an experiment
func (r *ExperimentRepository) Update(ctx context.Context, experiment *domain.Experiment) error {
	// Marshal JSON fields
	userIDFilter, err := json.Marshal(experiment.UserIDFilter)
	if err != nil {
		return fmt.Errorf("failed to marshal user_id_filter: %w", err)
	}

	metadataFilters, err := json.Marshal(experiment.MetadataFilters)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata_filters: %w", err)
	}

	var resultsJSON []byte
	if experiment.Results != nil {
		resultsJSON, err = json.Marshal(experiment.Results)
		if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
		}
	}

	query := `
		UPDATE experiments SET
			name = $1,
			description = $2,
			status = $3,
			target_metric = $4,
			target_goal = $5,
			traffic_percent = $6,
			trace_name_filter = $7,
			user_id_filter = $8,
			metadata_filters = $9,
			min_duration_hours = $10,
			min_samples_per_variant = $11,
			started_at = $12,
			ended_at = $13,
			winning_variant_id = $14,
			results = $15,
			statistical_power = $16,
			updated_at = NOW()
		WHERE id = $17
	`

	result, err := r.db.Pool.Exec(ctx, query,
		experiment.Name,
		experiment.Description,
		string(experiment.Status),
		experiment.TargetMetric,
		experiment.TargetGoal,
		experiment.TrafficPercent,
		experiment.TraceNameFilter,
		userIDFilter,
		metadataFilters,
		experiment.MinDuration,
		experiment.MinSamples,
		experiment.StartedAt,
		experiment.EndedAt,
		experiment.WinningVariant,
		resultsJSON,
		experiment.StatisticalPower,
		experiment.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique_experiment_name_per_project") {
			return apperrors.Conflict("experiment with this name already exists in the project")
		}
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("experiment not found")
	}

	return nil
}

// Delete deletes an experiment
func (r *ExperimentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Pool.Exec(ctx, "DELETE FROM experiments WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete experiment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("experiment not found")
	}

	return nil
}

// UpdateVariantStats updates the statistics for a variant
func (r *ExperimentRepository) UpdateVariantStats(ctx context.Context, variantID uuid.UUID, sampleCount int, mean, stdDev, min, max float64) error {
	query := `
		UPDATE experiment_variants SET
			sample_count = $1,
			metric_mean = $2,
			metric_std_dev = $3,
			metric_min = $4,
			metric_max = $5,
			updated_at = NOW()
		WHERE id = $6
	`

	result, err := r.db.Pool.Exec(ctx, query, sampleCount, mean, stdDev, min, max, variantID)
	if err != nil {
		return fmt.Errorf("failed to update variant stats: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("variant not found")
	}

	return nil
}

// CreateAssignment creates an experiment assignment for a trace
func (r *ExperimentRepository) CreateAssignment(ctx context.Context, assignment *domain.ExperimentAssignment) error {
	configJSON, err := json.Marshal(assignment.VariantConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal variant config: %w", err)
	}

	query := `
		INSERT INTO experiment_assignments (
			experiment_id, variant_id, trace_id, assigned_at, variant_config
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (experiment_id, trace_id) DO NOTHING
	`

	_, err = r.db.Pool.Exec(ctx, query,
		assignment.ExperimentID,
		assignment.VariantID,
		assignment.TraceID,
		assignment.AssignedAt,
		configJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

// GetAssignment retrieves an assignment for a trace in an experiment
func (r *ExperimentRepository) GetAssignment(ctx context.Context, experimentID, traceID uuid.UUID) (*domain.ExperimentAssignment, error) {
	query := `
		SELECT experiment_id, variant_id, trace_id, assigned_at, variant_config
		FROM experiment_assignments
		WHERE experiment_id = $1 AND trace_id = $2
	`

	var assignment domain.ExperimentAssignment
	var configJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, experimentID, traceID).Scan(
		&assignment.ExperimentID,
		&assignment.VariantID,
		&assignment.TraceID,
		&assignment.AssignedAt,
		&configJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is ok for assignments
		}
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &assignment.VariantConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variant config: %w", err)
		}
	}

	return &assignment, nil
}

// CreateMetric records a metric value for an experiment
func (r *ExperimentRepository) CreateMetric(ctx context.Context, metric *domain.ExperimentMetric) error {
	query := `
		INSERT INTO experiment_metrics (
			experiment_id, variant_id, trace_id, metric_name, metric_value, recorded_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (experiment_id, trace_id, metric_name)
		DO UPDATE SET metric_value = EXCLUDED.metric_value, recorded_at = EXCLUDED.recorded_at
	`

	_, err := r.db.Pool.Exec(ctx, query,
		metric.ExperimentID,
		metric.VariantID,
		metric.TraceID,
		metric.MetricName,
		metric.MetricValue,
		metric.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}

	return nil
}

// GetMetrics retrieves metrics for an experiment
func (r *ExperimentRepository) GetMetrics(ctx context.Context, experimentID uuid.UUID, metricName string) ([]domain.ExperimentMetric, error) {
	query := `
		SELECT experiment_id, variant_id, trace_id, metric_name, metric_value, recorded_at
		FROM experiment_metrics
		WHERE experiment_id = $1 AND metric_name = $2
		ORDER BY recorded_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, experimentID, metricName)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.ExperimentMetric
	for rows.Next() {
		var metric domain.ExperimentMetric
		err := rows.Scan(
			&metric.ExperimentID,
			&metric.VariantID,
			&metric.TraceID,
			&metric.MetricName,
			&metric.MetricValue,
			&metric.RecordedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAllMetrics retrieves all metrics for an experiment (all metric names)
func (r *ExperimentRepository) GetAllMetrics(ctx context.Context, experimentID uuid.UUID) ([]domain.ExperimentMetric, error) {
	query := `
		SELECT experiment_id, variant_id, trace_id, metric_name, metric_value, recorded_at
		FROM experiment_metrics
		WHERE experiment_id = $1
		ORDER BY recorded_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query all metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.ExperimentMetric
	for rows.Next() {
		var metric domain.ExperimentMetric
		err := rows.Scan(
			&metric.ExperimentID,
			&metric.VariantID,
			&metric.TraceID,
			&metric.MetricName,
			&metric.MetricValue,
			&metric.RecordedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// IncrementVariantSampleCount increments the sample count for a variant
func (r *ExperimentRepository) IncrementVariantSampleCount(ctx context.Context, variantID uuid.UUID) error {
	query := `
		UPDATE experiment_variants SET
			sample_count = sample_count + 1,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, variantID)
	if err != nil {
		return fmt.Errorf("failed to increment variant sample count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("variant not found")
	}

	return nil
}
