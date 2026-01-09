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

// EvaluatorRepository handles evaluator data operations in PostgreSQL
type EvaluatorRepository struct {
	db *database.PostgresDB
}

// NewEvaluatorRepository creates a new evaluator repository
func NewEvaluatorRepository(db *database.PostgresDB) *EvaluatorRepository {
	return &EvaluatorRepository{db: db}
}

// Create creates a new evaluator
func (r *EvaluatorRepository) Create(ctx context.Context, eval *domain.Evaluator) error {
	query := `
		INSERT INTO evaluators (id, project_id, name, description, type, config, prompt_template, variables, target_filter, sampling_rate, score_name, score_data_type, score_categories, enabled, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		eval.ID,
		eval.ProjectID,
		eval.Name,
		eval.Description,
		eval.Type,
		eval.Config,
		eval.PromptTemplate,
		eval.Variables,
		eval.TargetFilter,
		eval.SamplingRate,
		eval.ScoreName,
		eval.ScoreDataType,
		eval.ScoreCategories,
		eval.Enabled,
		eval.CreatedBy,
		eval.CreatedAt,
		eval.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create evaluator: %w", err)
	}

	return nil
}

// GetByID retrieves an evaluator by ID
func (r *EvaluatorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Evaluator, error) {
	query := `
		SELECT id, project_id, name, description, type, config, prompt_template, variables, target_filter, sampling_rate, score_name, score_data_type, score_categories, enabled, created_by, created_at, updated_at
		FROM evaluators
		WHERE id = $1
	`

	var eval domain.Evaluator
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&eval.ID,
		&eval.ProjectID,
		&eval.Name,
		&eval.Description,
		&eval.Type,
		&eval.Config,
		&eval.PromptTemplate,
		&eval.Variables,
		&eval.TargetFilter,
		&eval.SamplingRate,
		&eval.ScoreName,
		&eval.ScoreDataType,
		&eval.ScoreCategories,
		&eval.Enabled,
		&eval.CreatedBy,
		&eval.CreatedAt,
		&eval.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("evaluator")
		}
		return nil, fmt.Errorf("failed to get evaluator: %w", err)
	}

	return &eval, nil
}

// GetByName retrieves an evaluator by project and name
func (r *EvaluatorRepository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Evaluator, error) {
	query := `
		SELECT id, project_id, name, description, type, config, prompt_template, variables, target_filter, sampling_rate, score_name, score_data_type, score_categories, enabled, created_by, created_at, updated_at
		FROM evaluators
		WHERE project_id = $1 AND name = $2
	`

	var eval domain.Evaluator
	err := r.db.Pool.QueryRow(ctx, query, projectID, name).Scan(
		&eval.ID,
		&eval.ProjectID,
		&eval.Name,
		&eval.Description,
		&eval.Type,
		&eval.Config,
		&eval.PromptTemplate,
		&eval.Variables,
		&eval.TargetFilter,
		&eval.SamplingRate,
		&eval.ScoreName,
		&eval.ScoreDataType,
		&eval.ScoreCategories,
		&eval.Enabled,
		&eval.CreatedBy,
		&eval.CreatedAt,
		&eval.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("evaluator")
		}
		return nil, fmt.Errorf("failed to get evaluator: %w", err)
	}

	return &eval, nil
}

// Update updates an evaluator
func (r *EvaluatorRepository) Update(ctx context.Context, eval *domain.Evaluator) error {
	query := `
		UPDATE evaluators
		SET name = $2, description = $3, config = $4, prompt_template = $5, variables = $6, target_filter = $7, sampling_rate = $8, score_categories = $9, enabled = $10, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		eval.ID,
		eval.Name,
		eval.Description,
		eval.Config,
		eval.PromptTemplate,
		eval.Variables,
		eval.TargetFilter,
		eval.SamplingRate,
		eval.ScoreCategories,
		eval.Enabled,
	)
	if err != nil {
		return fmt.Errorf("failed to update evaluator: %w", err)
	}

	return nil
}

// Delete deletes an evaluator
func (r *EvaluatorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM evaluators WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete evaluator: %w", err)
	}

	return nil
}

// List retrieves evaluators with filtering
func (r *EvaluatorRepository) List(ctx context.Context, filter *domain.EvaluatorFilter, limit, offset int) (*domain.EvaluatorList, error) {
	baseQuery := `FROM evaluators WHERE project_id = $1`
	args := []interface{}{filter.ProjectID}
	argIndex := 2

	if filter.Name != nil {
		baseQuery += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.Type != nil {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Enabled != nil {
		baseQuery += fmt.Sprintf(" AND enabled = $%d", argIndex)
		args = append(args, *filter.Enabled)
		argIndex++
	}

	// Get count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count evaluators: %w", err)
	}

	// Get evaluators
	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, type, config, prompt_template, variables, target_filter, sampling_rate, score_name, score_data_type, score_categories, enabled, created_by, created_at, updated_at
		%s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluators: %w", err)
	}
	defer rows.Close()

	var evaluators []domain.Evaluator
	for rows.Next() {
		var eval domain.Evaluator
		if err := rows.Scan(
			&eval.ID,
			&eval.ProjectID,
			&eval.Name,
			&eval.Description,
			&eval.Type,
			&eval.Config,
			&eval.PromptTemplate,
			&eval.Variables,
			&eval.TargetFilter,
			&eval.SamplingRate,
			&eval.ScoreName,
			&eval.ScoreDataType,
			&eval.ScoreCategories,
			&eval.Enabled,
			&eval.CreatedBy,
			&eval.CreatedAt,
			&eval.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan evaluator: %w", err)
		}
		evaluators = append(evaluators, eval)
	}

	return &domain.EvaluatorList{
		Evaluators: evaluators,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(evaluators)) < totalCount,
	}, nil
}

// ListEnabled retrieves all enabled evaluators for a project
func (r *EvaluatorRepository) ListEnabled(ctx context.Context, projectID uuid.UUID) ([]domain.Evaluator, error) {
	query := `
		SELECT id, project_id, name, description, type, config, prompt_template, variables, target_filter, sampling_rate, score_name, score_data_type, score_categories, enabled, created_by, created_at, updated_at
		FROM evaluators
		WHERE project_id = $1 AND enabled = true
		ORDER BY name
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled evaluators: %w", err)
	}
	defer rows.Close()

	var evaluators []domain.Evaluator
	for rows.Next() {
		var eval domain.Evaluator
		if err := rows.Scan(
			&eval.ID,
			&eval.ProjectID,
			&eval.Name,
			&eval.Description,
			&eval.Type,
			&eval.Config,
			&eval.PromptTemplate,
			&eval.Variables,
			&eval.TargetFilter,
			&eval.SamplingRate,
			&eval.ScoreName,
			&eval.ScoreDataType,
			&eval.ScoreCategories,
			&eval.Enabled,
			&eval.CreatedBy,
			&eval.CreatedAt,
			&eval.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan evaluator: %w", err)
		}
		evaluators = append(evaluators, eval)
	}

	return evaluators, nil
}

// NameExists checks if an evaluator name already exists
func (r *EvaluatorRepository) NameExists(ctx context.Context, projectID uuid.UUID, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM evaluators WHERE project_id = $1 AND name = $2)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, projectID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check name: %w", err)
	}

	return exists, nil
}

// GetTemplate retrieves an evaluator template by ID
func (r *EvaluatorRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*domain.EvaluatorTemplate, error) {
	query := `
		SELECT id, name, description, prompt_template, variables, score_data_type, score_categories, config, created_at
		FROM evaluator_templates
		WHERE id = $1
	`

	var tmpl domain.EvaluatorTemplate
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&tmpl.ID,
		&tmpl.Name,
		&tmpl.Description,
		&tmpl.PromptTemplate,
		&tmpl.Variables,
		&tmpl.ScoreDataType,
		&tmpl.ScoreCategories,
		&tmpl.Config,
		&tmpl.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("evaluator template")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &tmpl, nil
}

// GetTemplateByName retrieves an evaluator template by name
func (r *EvaluatorRepository) GetTemplateByName(ctx context.Context, name string) (*domain.EvaluatorTemplate, error) {
	query := `
		SELECT id, name, description, prompt_template, variables, score_data_type, score_categories, config, created_at
		FROM evaluator_templates
		WHERE name = $1
	`

	var tmpl domain.EvaluatorTemplate
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&tmpl.ID,
		&tmpl.Name,
		&tmpl.Description,
		&tmpl.PromptTemplate,
		&tmpl.Variables,
		&tmpl.ScoreDataType,
		&tmpl.ScoreCategories,
		&tmpl.Config,
		&tmpl.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("evaluator template")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &tmpl, nil
}

// ListTemplates retrieves all evaluator templates
func (r *EvaluatorRepository) ListTemplates(ctx context.Context) ([]domain.EvaluatorTemplate, error) {
	query := `
		SELECT id, name, description, prompt_template, variables, score_data_type, score_categories, config, created_at
		FROM evaluator_templates
		ORDER BY name
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []domain.EvaluatorTemplate
	for rows.Next() {
		var tmpl domain.EvaluatorTemplate
		if err := rows.Scan(
			&tmpl.ID,
			&tmpl.Name,
			&tmpl.Description,
			&tmpl.PromptTemplate,
			&tmpl.Variables,
			&tmpl.ScoreDataType,
			&tmpl.ScoreCategories,
			&tmpl.Config,
			&tmpl.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, tmpl)
	}

	return templates, nil
}

// CreateJob creates a new evaluation job
func (r *EvaluatorRepository) CreateJob(ctx context.Context, job *domain.EvaluationJob) error {
	query := `
		INSERT INTO evaluation_jobs (id, evaluator_id, trace_id, observation_id, status, scheduled_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		job.ID,
		job.EvaluatorID,
		job.TraceID,
		job.ObservationID,
		job.Status,
		job.ScheduledAt,
		job.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create evaluation job: %w", err)
	}

	return nil
}

// GetJobByID retrieves an evaluation job by ID
func (r *EvaluatorRepository) GetJobByID(ctx context.Context, id uuid.UUID) (*domain.EvaluationJob, error) {
	query := `
		SELECT id, evaluator_id, trace_id, observation_id, status, result, error, attempts, scheduled_at, started_at, completed_at, created_at
		FROM evaluation_jobs
		WHERE id = $1
	`

	var job domain.EvaluationJob
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.EvaluatorID,
		&job.TraceID,
		&job.ObservationID,
		&job.Status,
		&job.Result,
		&job.Error,
		&job.Attempts,
		&job.ScheduledAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("evaluation job")
		}
		return nil, fmt.Errorf("failed to get evaluation job: %w", err)
	}

	return &job, nil
}

// UpdateJob updates an evaluation job
func (r *EvaluatorRepository) UpdateJob(ctx context.Context, job *domain.EvaluationJob) error {
	query := `
		UPDATE evaluation_jobs
		SET status = $2, result = $3, error = $4, attempts = $5, started_at = $6, completed_at = $7
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		job.ID,
		job.Status,
		job.Result,
		job.Error,
		job.Attempts,
		job.StartedAt,
		job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update evaluation job: %w", err)
	}

	return nil
}

// ListPendingJobs retrieves pending evaluation jobs
func (r *EvaluatorRepository) ListPendingJobs(ctx context.Context, limit int) ([]domain.EvaluationJob, error) {
	query := `
		SELECT id, evaluator_id, trace_id, observation_id, status, result, error, attempts, scheduled_at, started_at, completed_at, created_at
		FROM evaluation_jobs
		WHERE status = 'pending' AND scheduled_at <= NOW()
		ORDER BY scheduled_at
		LIMIT $1
	`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.EvaluationJob
	for rows.Next() {
		var job domain.EvaluationJob
		if err := rows.Scan(
			&job.ID,
			&job.EvaluatorID,
			&job.TraceID,
			&job.ObservationID,
			&job.Status,
			&job.Result,
			&job.Error,
			&job.Attempts,
			&job.ScheduledAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// JobExists checks if a job already exists for the evaluator and trace
func (r *EvaluatorRepository) JobExists(ctx context.Context, evaluatorID uuid.UUID, traceID string, observationID *string) (bool, error) {
	var query string
	var args []interface{}

	if observationID == nil {
		query = `SELECT EXISTS(SELECT 1 FROM evaluation_jobs WHERE evaluator_id = $1 AND trace_id = $2 AND observation_id IS NULL)`
		args = []interface{}{evaluatorID, traceID}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM evaluation_jobs WHERE evaluator_id = $1 AND trace_id = $2 AND observation_id = $3)`
		args = []interface{}{evaluatorID, traceID, *observationID}
	}

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check job: %w", err)
	}

	return exists, nil
}

// GetEvalCount returns the number of evaluations for an evaluator
func (r *EvaluatorRepository) GetEvalCount(ctx context.Context, evaluatorID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM evaluation_jobs WHERE evaluator_id = $1 AND status = 'completed'`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, evaluatorID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count evaluations: %w", err)
	}

	return count, nil
}

// CreateAnnotationQueue creates a new annotation queue
func (r *EvaluatorRepository) CreateAnnotationQueue(ctx context.Context, queue *domain.AnnotationQueue) error {
	query := `
		INSERT INTO annotation_queues (id, project_id, name, description, score_name, score_config, filters, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		queue.ID,
		queue.ProjectID,
		queue.Name,
		queue.Description,
		queue.ScoreName,
		queue.ScoreConfig,
		queue.Filters,
		queue.CreatedAt,
		queue.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create annotation queue: %w", err)
	}

	return nil
}

// GetAnnotationQueueByID retrieves an annotation queue by ID
func (r *EvaluatorRepository) GetAnnotationQueueByID(ctx context.Context, id uuid.UUID) (*domain.AnnotationQueue, error) {
	query := `
		SELECT id, project_id, name, description, score_name, score_config, filters, created_at, updated_at
		FROM annotation_queues
		WHERE id = $1
	`

	var queue domain.AnnotationQueue
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&queue.ID,
		&queue.ProjectID,
		&queue.Name,
		&queue.Description,
		&queue.ScoreName,
		&queue.ScoreConfig,
		&queue.Filters,
		&queue.CreatedAt,
		&queue.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("annotation queue")
		}
		return nil, fmt.Errorf("failed to get annotation queue: %w", err)
	}

	return &queue, nil
}

// UpdateAnnotationQueue updates an annotation queue
func (r *EvaluatorRepository) UpdateAnnotationQueue(ctx context.Context, queue *domain.AnnotationQueue) error {
	query := `
		UPDATE annotation_queues
		SET name = $2, description = $3, score_config = $4, filters = $5, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		queue.ID,
		queue.Name,
		queue.Description,
		queue.ScoreConfig,
		queue.Filters,
	)
	if err != nil {
		return fmt.Errorf("failed to update annotation queue: %w", err)
	}

	return nil
}

// DeleteAnnotationQueue deletes an annotation queue
func (r *EvaluatorRepository) DeleteAnnotationQueue(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM annotation_queues WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete annotation queue: %w", err)
	}

	return nil
}

// ListAnnotationQueues retrieves annotation queues for a project
func (r *EvaluatorRepository) ListAnnotationQueues(ctx context.Context, projectID uuid.UUID) ([]domain.AnnotationQueue, error) {
	query := `
		SELECT id, project_id, name, description, score_name, score_config, filters, created_at, updated_at
		FROM annotation_queues
		WHERE project_id = $1
		ORDER BY name
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotation queues: %w", err)
	}
	defer rows.Close()

	var queues []domain.AnnotationQueue
	for rows.Next() {
		var queue domain.AnnotationQueue
		if err := rows.Scan(
			&queue.ID,
			&queue.ProjectID,
			&queue.Name,
			&queue.Description,
			&queue.ScoreName,
			&queue.ScoreConfig,
			&queue.Filters,
			&queue.CreatedAt,
			&queue.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan annotation queue: %w", err)
		}
		queues = append(queues, queue)
	}

	return queues, nil
}

// CreateAnnotationQueueItem creates a new annotation queue item
func (r *EvaluatorRepository) CreateAnnotationQueueItem(ctx context.Context, item *domain.AnnotationQueueItem) error {
	query := `
		INSERT INTO annotation_queue_items (id, queue_id, trace_id, observation_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		item.ID,
		item.QueueID,
		item.TraceID,
		item.ObservationID,
		item.Status,
		item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create annotation queue item: %w", err)
	}

	return nil
}

// GetNextAnnotationItem retrieves the next pending annotation item
func (r *EvaluatorRepository) GetNextAnnotationItem(ctx context.Context, queueID uuid.UUID) (*domain.AnnotationQueueItem, error) {
	query := `
		SELECT id, queue_id, trace_id, observation_id, status, completed_by, completed_at, created_at
		FROM annotation_queue_items
		WHERE queue_id = $1 AND status = 'pending'
		ORDER BY created_at
		LIMIT 1
	`

	var item domain.AnnotationQueueItem
	err := r.db.Pool.QueryRow(ctx, query, queueID).Scan(
		&item.ID,
		&item.QueueID,
		&item.TraceID,
		&item.ObservationID,
		&item.Status,
		&item.CompletedBy,
		&item.CompletedAt,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("annotation item")
		}
		return nil, fmt.Errorf("failed to get next annotation item: %w", err)
	}

	return &item, nil
}

// CompleteAnnotationItem marks an annotation item as completed
func (r *EvaluatorRepository) CompleteAnnotationItem(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE annotation_queue_items
		SET status = 'completed', completed_by = $2, completed_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to complete annotation item: %w", err)
	}

	return nil
}

// GetAnnotationQueueStats retrieves stats for an annotation queue
func (r *EvaluatorRepository) GetAnnotationQueueStats(ctx context.Context, queueID uuid.UUID) (pending, completed int64, err error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'completed') as completed
		FROM annotation_queue_items
		WHERE queue_id = $1
	`

	err = r.db.Pool.QueryRow(ctx, query, queueID).Scan(&pending, &completed)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get annotation queue stats: %w", err)
	}

	return pending, completed, nil
}
