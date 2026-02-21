package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
		return fmt.Errorf("failed to save migration job: %w", err)
	}

	return nil
}

// GetByID retrieves a migration job by ID
func (r *MigrationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MigrationJob, error) {
	query := `
		SELECT id, project_id, source, status, config, progress, errors, created_at, completed_at
		FROM migration_jobs
		WHERE id = $1
	`

	var job domain.MigrationJob
	var configJSON, progressJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
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
