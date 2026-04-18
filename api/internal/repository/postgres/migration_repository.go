package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// MigrationRepository handles migration job data operations in PostgreSQL
type MigrationRepository struct {
	db *database.PostgresDB
}

// NewMigrationRepository creates a new migration repository
func NewMigrationRepository(db *database.PostgresDB) *MigrationRepository {
	return &MigrationRepository{db: db}
}

// Save creates a new migration job
func (r *MigrationRepository) Save(ctx context.Context, job *domain.MigrationJob) error {
	configJSON, err := json.Marshal(job.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	progressJSON, err := json.Marshal(job.Progress)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	query := `
		INSERT INTO migration_jobs (id, project_id, source, status, config, progress, errors, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		job.ID,
		job.ProjectID,
		job.Source,
		job.Status,
		configJSON,
		progressJSON,
		job.Errors,
		job.CreatedAt,
		job.CompletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.Conflict("migration job already exists")
		}
		return fmt.Errorf("failed to save migration job: %w", err)
	}

	return nil
}

// GetByID retrieves a migration job by ID
func (r *MigrationRepository) GetByID(ctx context.Context, projectID, id uuid.UUID) (*domain.MigrationJob, error) {
	query := `
		SELECT id, project_id, source, status, config, progress, errors, created_at, completed_at
		FROM migration_jobs
		WHERE project_id = $1 AND id = $2
	`

	var job domain.MigrationJob
	var configJSON, progressJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, projectID, id).Scan(
		&job.ID,
		&job.ProjectID,
		&job.Source,
		&job.Status,
		&configJSON,
		&progressJSON,
		&job.Errors,
		&job.CreatedAt,
		&job.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("migration job")
		}
		return nil, fmt.Errorf("failed to get migration job: %w", err)
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &job.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	if len(progressJSON) > 0 {
		if err := json.Unmarshal(progressJSON, &job.Progress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
		}
	}

	return &job, nil
}

// ListByProject retrieves all migration jobs for a project
func (r *MigrationRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.MigrationJob, error) {
	query := `
		SELECT id, project_id, source, status, config, progress, errors, created_at, completed_at
		FROM migration_jobs
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list migration jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.MigrationJob
	for rows.Next() {
		var job domain.MigrationJob
		var configJSON, progressJSON []byte

		if err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.Source,
			&job.Status,
			&configJSON,
			&progressJSON,
			&job.Errors,
			&job.CreatedAt,
			&job.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan migration job: %w", err)
		}

		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &job.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}

		if len(progressJSON) > 0 {
			if err := json.Unmarshal(progressJSON, &job.Progress); err != nil {
				return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// UpdateProgress updates the progress of a migration job
func (r *MigrationRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress domain.MigrationProgress) error {
	progressJSON, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	query := `
		UPDATE migration_jobs
		SET progress = $2
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, id, progressJSON)
	if err != nil {
		return fmt.Errorf("failed to update migration progress: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("migration job")
	}

	return nil
}

// UpdateJob persists status, progress, redacted errors, and completion state.
func (r *MigrationRepository) UpdateJob(
	ctx context.Context,
	job *domain.MigrationJob,
) error {
	progressJSON, err := json.Marshal(job.Progress)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}
	result, err := r.db.Pool.Exec(ctx, `
		UPDATE migration_jobs
		SET status = $3, progress = $4, errors = $5, completed_at = $6
		WHERE project_id = $1 AND id = $2
	`,
		job.ProjectID,
		job.ID,
		job.Status,
		progressJSON,
		job.Errors,
		job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("update migration job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("migration job")
	}
	return nil
}

// FindImportedItem reads the idempotency ledger and returns the identifier
// recorded for a previously imported source item.
func (r *MigrationRepository) FindImportedItem(
	ctx context.Context,
	projectID, jobID uuid.UUID,
	sourceType, sourceID string,
) (recordedID string, imported bool, err error) {
	var importedID string
	if scanErr := r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(imported_id, '')
		FROM migration_import_items
		WHERE project_id = $1 AND job_id = $2
			AND source_type = $3 AND source_id = $4
			AND status = 'imported'
	`, projectID, jobID, sourceType, sourceID).Scan(&importedID); scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("check imported migration item: %w", scanErr)
	}
	return importedID, true, nil
}

// RecordItem records an imported or failed source item without storing source payloads.
func (r *MigrationRepository) RecordItem(
	ctx context.Context,
	projectID, jobID uuid.UUID,
	sourceType, sourceID, checksum, importedID, status, errorMessage string,
) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO migration_import_items (
			project_id, job_id, source_type, source_id, checksum,
			imported_id, status, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, NOW())
		ON CONFLICT (job_id, source_type, source_id)
		DO UPDATE SET
			checksum = CASE
				WHEN migration_import_items.status = 'imported'
					AND EXCLUDED.status = 'failed'
				THEN migration_import_items.checksum
				ELSE EXCLUDED.checksum
			END,
			imported_id = CASE
				WHEN migration_import_items.status = 'imported'
					AND EXCLUDED.status = 'failed'
				THEN migration_import_items.imported_id
				ELSE EXCLUDED.imported_id
			END,
			status = CASE
				WHEN migration_import_items.status = 'imported'
				THEN migration_import_items.status
				ELSE EXCLUDED.status
			END,
			error_message = CASE
				WHEN migration_import_items.status = 'imported'
				THEN ''
				ELSE EXCLUDED.error_message
			END
	`,
		projectID,
		jobID,
		sourceType,
		sourceID,
		checksum,
		importedID,
		status,
		errorMessage,
	)
	if err != nil {
		return fmt.Errorf("record migration item: %w", err)
	}
	return nil
}

// batchLockReleaseTimeout bounds the detached commit/rollback that releases the
// transaction-scoped advisory lock so the lock is always freed even when the
// request context has already been canceled.
const batchLockReleaseTimeout = 5 * time.Second

// WithBatchLock runs fn while holding a PostgreSQL transaction-scoped advisory
// lock keyed by (projectID, jobID). The lock serializes Langfuse import batches
// for the same job across processes; competing callers block on acquisition
// rather than racing a read-modify-write of the job's progress. The lock is
// held on a dedicated connection for the transaction's lifetime and released
// when it commits (or rolls back on error). Both acquisition and release
// failures are propagated so the caller can surface them instead of silently
// proceeding without mutual exclusion.
func (r *MigrationRepository) WithBatchLock(
	ctx context.Context,
	projectID, jobID uuid.UUID,
	fn func(ctx context.Context) error,
) error {
	return withMigrationBatchLock(ctx, r.db.Pool.Begin, projectID, jobID, fn)
}

func withMigrationBatchLock(
	ctx context.Context,
	begin func(context.Context) (pgx.Tx, error),
	projectID, jobID uuid.UUID,
	fn func(ctx context.Context) error,
) (resultErr error) {
	tx, err := begin(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration batch lock connection: %w", err)
	}

	released := false
	defer func() {
		if released {
			return
		}
		// Release on a short detached timeout so a canceled request context
		// cannot leave the advisory lock held on the pooled connection.
		releaseCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			batchLockReleaseTimeout,
		)
		defer cancel()
		if rollbackErr := tx.Rollback(releaseCtx); rollbackErr != nil &&
			!errors.Is(rollbackErr, pgx.ErrTxClosed) {
			resultErr = errors.Join(
				resultErr,
				fmt.Errorf("release migration batch lock: %w", rollbackErr),
			)
		}
	}()

	// hashtext maps the identifiers onto the two int4 keys accepted by
	// pg_advisory_xact_lock, keeping the lock scoped to this exact job.
	lockNamespace := "langfuse-import:" + projectID.String()
	if _, err := tx.Exec(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtext($1), hashtext($2))`,
		lockNamespace,
		jobID.String(),
	); err != nil {
		return fmt.Errorf("acquire migration batch lock: %w", err)
	}

	if err := fn(ctx); err != nil {
		return err
	}

	// Commit on a short detached timeout so the lock is released reliably even
	// if the request context was canceled during processing.
	commitCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		batchLockReleaseTimeout,
	)
	defer cancel()
	if err := tx.Commit(commitCtx); err != nil {
		return fmt.Errorf("release migration batch lock: %w", err)
	}
	released = true
	return nil
}
