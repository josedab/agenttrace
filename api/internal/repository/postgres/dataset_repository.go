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

// DatasetRepository handles dataset data operations in PostgreSQL
type DatasetRepository struct {
	db *database.PostgresDB
}

// NewDatasetRepository creates a new dataset repository
func NewDatasetRepository(db *database.PostgresDB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

// Create creates a new dataset
func (r *DatasetRepository) Create(ctx context.Context, dataset *domain.Dataset) error {
	query := `
		INSERT INTO datasets (id, project_id, name, description, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		dataset.ID,
		dataset.ProjectID,
		dataset.Name,
		dataset.Description,
		dataset.Metadata,
		dataset.CreatedAt,
		dataset.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	return nil
}

// GetByID retrieves a dataset by ID
func (r *DatasetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dataset, error) {
	query := `
		SELECT id, project_id, name, description, metadata, created_at, updated_at
		FROM datasets
		WHERE id = $1
	`

	var dataset domain.Dataset
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&dataset.ID,
		&dataset.ProjectID,
		&dataset.Name,
		&dataset.Description,
		&dataset.Metadata,
		&dataset.CreatedAt,
		&dataset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("dataset")
		}
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	return &dataset, nil
}

// GetByName retrieves a dataset by project and name
func (r *DatasetRepository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Dataset, error) {
	query := `
		SELECT id, project_id, name, description, metadata, created_at, updated_at
		FROM datasets
		WHERE project_id = $1 AND name = $2
	`

	var dataset domain.Dataset
	err := r.db.Pool.QueryRow(ctx, query, projectID, name).Scan(
		&dataset.ID,
		&dataset.ProjectID,
		&dataset.Name,
		&dataset.Description,
		&dataset.Metadata,
		&dataset.CreatedAt,
		&dataset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("dataset")
		}
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	return &dataset, nil
}

// Update updates a dataset
func (r *DatasetRepository) Update(ctx context.Context, dataset *domain.Dataset) error {
	query := `
		UPDATE datasets
		SET name = $2, description = $3, metadata = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		dataset.ID,
		dataset.Name,
		dataset.Description,
		dataset.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to update dataset: %w", err)
	}

	return nil
}

// Delete deletes a dataset
func (r *DatasetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM datasets WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}

	return nil
}

// List retrieves datasets with filtering
func (r *DatasetRepository) List(ctx context.Context, filter *domain.DatasetFilter, limit, offset int) (*domain.DatasetList, error) {
	baseQuery := `FROM datasets WHERE project_id = $1`
	args := []interface{}{filter.ProjectID}
	argIndex := 2

	if filter.Name != nil {
		baseQuery += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	// Get count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count datasets: %w", err)
	}

	// Get datasets
	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, metadata, created_at, updated_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}
	defer rows.Close()

	var datasets []domain.Dataset
	for rows.Next() {
		var dataset domain.Dataset
		if err := rows.Scan(
			&dataset.ID,
			&dataset.ProjectID,
			&dataset.Name,
			&dataset.Description,
			&dataset.Metadata,
			&dataset.CreatedAt,
			&dataset.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dataset: %w", err)
		}
		datasets = append(datasets, dataset)
	}

	return &domain.DatasetList{
		Datasets:   datasets,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(datasets)) < totalCount,
	}, nil
}

// GetItemCount returns the number of items in a dataset
func (r *DatasetRepository) GetItemCount(ctx context.Context, datasetID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM dataset_items WHERE dataset_id = $1`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, datasetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}

	return count, nil
}

// GetRunCount returns the number of runs for a dataset
func (r *DatasetRepository) GetRunCount(ctx context.Context, datasetID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM dataset_runs WHERE dataset_id = $1`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, datasetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count runs: %w", err)
	}

	return count, nil
}

// NameExists checks if a dataset name already exists
func (r *DatasetRepository) NameExists(ctx context.Context, projectID uuid.UUID, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM datasets WHERE project_id = $1 AND name = $2)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, projectID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check name: %w", err)
	}

	return exists, nil
}

// CreateItem creates a new dataset item
func (r *DatasetRepository) CreateItem(ctx context.Context, item *domain.DatasetItem) error {
	query := `
		INSERT INTO dataset_items (id, dataset_id, input, expected_output, metadata, source_trace_id, source_observation_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		item.ID,
		item.DatasetID,
		item.Input,
		item.ExpectedOutput,
		item.Metadata,
		item.SourceTraceID,
		item.SourceObservationID,
		item.Status,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create dataset item: %w", err)
	}

	return nil
}

// GetItemByID retrieves a dataset item by ID
func (r *DatasetRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.DatasetItem, error) {
	query := `
		SELECT id, dataset_id, input, expected_output, metadata, source_trace_id, source_observation_id, status, created_at, updated_at
		FROM dataset_items
		WHERE id = $1
	`

	var item domain.DatasetItem
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.DatasetID,
		&item.Input,
		&item.ExpectedOutput,
		&item.Metadata,
		&item.SourceTraceID,
		&item.SourceObservationID,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("dataset item")
		}
		return nil, fmt.Errorf("failed to get dataset item: %w", err)
	}

	return &item, nil
}

// UpdateItem updates a dataset item
func (r *DatasetRepository) UpdateItem(ctx context.Context, item *domain.DatasetItem) error {
	query := `
		UPDATE dataset_items
		SET input = $2, expected_output = $3, metadata = $4, status = $5, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		item.ID,
		item.Input,
		item.ExpectedOutput,
		item.Metadata,
		item.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update dataset item: %w", err)
	}

	return nil
}

// DeleteItem deletes a dataset item
func (r *DatasetRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM dataset_items WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dataset item: %w", err)
	}

	return nil
}

// ListItems retrieves dataset items with filtering
func (r *DatasetRepository) ListItems(ctx context.Context, filter *domain.DatasetItemFilter, limit, offset int) ([]domain.DatasetItem, int64, error) {
	baseQuery := `FROM dataset_items WHERE dataset_id = $1`
	args := []interface{}{filter.DatasetID}
	argIndex := 2

	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	// Get count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	// Get items
	query := fmt.Sprintf(`
		SELECT id, dataset_id, input, expected_output, metadata, source_trace_id, source_observation_id, status, created_at, updated_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}
	defer rows.Close()

	var items []domain.DatasetItem
	for rows.Next() {
		var item domain.DatasetItem
		if err := rows.Scan(
			&item.ID,
			&item.DatasetID,
			&item.Input,
			&item.ExpectedOutput,
			&item.Metadata,
			&item.SourceTraceID,
			&item.SourceObservationID,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, nil
}

// CreateRun creates a new dataset run
func (r *DatasetRepository) CreateRun(ctx context.Context, run *domain.DatasetRun) error {
	query := `
		INSERT INTO dataset_runs (id, dataset_id, name, description, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		run.ID,
		run.DatasetID,
		run.Name,
		run.Description,
		run.Metadata,
		run.CreatedAt,
		run.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create dataset run: %w", err)
	}

	return nil
}

// GetRunByID retrieves a dataset run by ID
func (r *DatasetRepository) GetRunByID(ctx context.Context, id uuid.UUID) (*domain.DatasetRun, error) {
	query := `
		SELECT id, dataset_id, name, description, metadata, created_at, updated_at
		FROM dataset_runs
		WHERE id = $1
	`

	var run domain.DatasetRun
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&run.ID,
		&run.DatasetID,
		&run.Name,
		&run.Description,
		&run.Metadata,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("dataset run")
		}
		return nil, fmt.Errorf("failed to get dataset run: %w", err)
	}

	return &run, nil
}

// GetRunByName retrieves a dataset run by dataset and name
func (r *DatasetRepository) GetRunByName(ctx context.Context, datasetID uuid.UUID, name string) (*domain.DatasetRun, error) {
	query := `
		SELECT id, dataset_id, name, description, metadata, created_at, updated_at
		FROM dataset_runs
		WHERE dataset_id = $1 AND name = $2
	`

	var run domain.DatasetRun
	err := r.db.Pool.QueryRow(ctx, query, datasetID, name).Scan(
		&run.ID,
		&run.DatasetID,
		&run.Name,
		&run.Description,
		&run.Metadata,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("dataset run")
		}
		return nil, fmt.Errorf("failed to get dataset run: %w", err)
	}

	return &run, nil
}

// UpdateRun updates a dataset run
func (r *DatasetRepository) UpdateRun(ctx context.Context, run *domain.DatasetRun) error {
	query := `
		UPDATE dataset_runs
		SET name = $2, description = $3, metadata = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		run.ID,
		run.Name,
		run.Description,
		run.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to update dataset run: %w", err)
	}

	return nil
}

// DeleteRun deletes a dataset run
func (r *DatasetRepository) DeleteRun(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM dataset_runs WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dataset run: %w", err)
	}

	return nil
}

// ListRuns retrieves dataset runs
func (r *DatasetRepository) ListRuns(ctx context.Context, datasetID uuid.UUID, limit, offset int) ([]domain.DatasetRun, int64, error) {
	// Get count
	countQuery := `SELECT COUNT(*) FROM dataset_runs WHERE dataset_id = $1`
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, datasetID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count runs: %w", err)
	}

	// Get runs
	query := `
		SELECT id, dataset_id, name, description, metadata, created_at, updated_at
		FROM dataset_runs
		WHERE dataset_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, datasetID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.DatasetRun
	for rows.Next() {
		var run domain.DatasetRun
		if err := rows.Scan(
			&run.ID,
			&run.DatasetID,
			&run.Name,
			&run.Description,
			&run.Metadata,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, totalCount, nil
}

// GetRunItemCount returns the number of items in a run
func (r *DatasetRepository) GetRunItemCount(ctx context.Context, runID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM dataset_run_items WHERE dataset_run_id = $1`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, runID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count run items: %w", err)
	}

	return count, nil
}

// CreateRunItem creates a new dataset run item
func (r *DatasetRepository) CreateRunItem(ctx context.Context, item *domain.DatasetRunItem) error {
	query := `
		INSERT INTO dataset_run_items (id, dataset_run_id, dataset_item_id, trace_id, observation_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		item.ID,
		item.DatasetRunID,
		item.DatasetItemID,
		item.TraceID,
		item.ObservationID,
		item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create run item: %w", err)
	}

	return nil
}

// GetRunItemByID retrieves a dataset run item by ID
func (r *DatasetRepository) GetRunItemByID(ctx context.Context, id uuid.UUID) (*domain.DatasetRunItem, error) {
	query := `
		SELECT id, dataset_run_id, dataset_item_id, trace_id, observation_id, created_at
		FROM dataset_run_items
		WHERE id = $1
	`

	var item domain.DatasetRunItem
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.DatasetRunID,
		&item.DatasetItemID,
		&item.TraceID,
		&item.ObservationID,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("run item")
		}
		return nil, fmt.Errorf("failed to get run item: %w", err)
	}

	return &item, nil
}

// ListRunItems retrieves run items with dataset item info
func (r *DatasetRepository) ListRunItems(ctx context.Context, runID uuid.UUID, limit, offset int) ([]domain.DatasetRunItem, int64, error) {
	// Get count
	countQuery := `SELECT COUNT(*) FROM dataset_run_items WHERE dataset_run_id = $1`
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, runID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count run items: %w", err)
	}

	// Get items with dataset item info
	query := `
		SELECT dri.id, dri.dataset_run_id, dri.dataset_item_id, dri.trace_id, dri.observation_id, dri.created_at,
		       di.id, di.dataset_id, di.input, di.expected_output, di.metadata, di.source_trace_id, di.source_observation_id, di.status, di.created_at, di.updated_at
		FROM dataset_run_items dri
		JOIN dataset_items di ON dri.dataset_item_id = di.id
		WHERE dri.dataset_run_id = $1
		ORDER BY dri.created_at
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, runID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list run items: %w", err)
	}
	defer rows.Close()

	var items []domain.DatasetRunItem
	for rows.Next() {
		var item domain.DatasetRunItem
		var datasetItem domain.DatasetItem
		if err := rows.Scan(
			&item.ID,
			&item.DatasetRunID,
			&item.DatasetItemID,
			&item.TraceID,
			&item.ObservationID,
			&item.CreatedAt,
			&datasetItem.ID,
			&datasetItem.DatasetID,
			&datasetItem.Input,
			&datasetItem.ExpectedOutput,
			&datasetItem.Metadata,
			&datasetItem.SourceTraceID,
			&datasetItem.SourceObservationID,
			&datasetItem.Status,
			&datasetItem.CreatedAt,
			&datasetItem.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan run item: %w", err)
		}
		item.DatasetItem = &datasetItem
		items = append(items, item)
	}

	return items, totalCount, nil
}

// GetRunItemByDatasetItem retrieves a run item by run and dataset item
func (r *DatasetRepository) GetRunItemByDatasetItem(ctx context.Context, runID, itemID uuid.UUID) (*domain.DatasetRunItem, error) {
	query := `
		SELECT id, dataset_run_id, dataset_item_id, trace_id, observation_id, created_at
		FROM dataset_run_items
		WHERE dataset_run_id = $1 AND dataset_item_id = $2
	`

	var item domain.DatasetRunItem
	err := r.db.Pool.QueryRow(ctx, query, runID, itemID).Scan(
		&item.ID,
		&item.DatasetRunID,
		&item.DatasetItemID,
		&item.TraceID,
		&item.ObservationID,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("run item")
		}
		return nil, fmt.Errorf("failed to get run item: %w", err)
	}

	return &item, nil
}
