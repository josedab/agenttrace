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

// GoldenDatasetRepository handles golden dataset data operations in PostgreSQL
type GoldenDatasetRepository struct {
	db *database.PostgresDB
}

// NewGoldenDatasetRepository creates a new golden dataset repository
func NewGoldenDatasetRepository(db *database.PostgresDB) *GoldenDatasetRepository {
	return &GoldenDatasetRepository{db: db}
}

// SaveDataset creates a new golden dataset
func (r *GoldenDatasetRepository) SaveDataset(ctx context.Context, dataset *domain.GoldenDataset) error {
	itemsJSON, err := json.Marshal(dataset.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		INSERT INTO golden_datasets (id, project_id, name, description, category, language, items, item_count, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		dataset.ID,
		dataset.ProjectID,
		dataset.Name,
		dataset.Description,
		dataset.Category,
		dataset.Language,
		itemsJSON,
		dataset.ItemCount,
		dataset.CreatedAt,
		dataset.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save golden dataset: %w", err)
	}

	return nil
}

// GetDatasetByID retrieves a golden dataset by ID
func (r *GoldenDatasetRepository) GetDatasetByID(ctx context.Context, id uuid.UUID) (*domain.GoldenDataset, error) {
	query := `
		SELECT id, project_id, name, description, category, language, items, item_count, created_at, created_by
		FROM golden_datasets
		WHERE id = $1
	`

	var dataset domain.GoldenDataset
	var itemsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&dataset.ID,
		&dataset.ProjectID,
		&dataset.Name,
		&dataset.Description,
		&dataset.Category,
		&dataset.Language,
		&itemsJSON,
		&dataset.ItemCount,
		&dataset.CreatedAt,
		&dataset.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("golden dataset")
		}
		return nil, fmt.Errorf("failed to get golden dataset: %w", err)
	}

	if len(itemsJSON) > 0 {
		if err := json.Unmarshal(itemsJSON, &dataset.Items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}
	}

	return &dataset, nil
}

// ListDatasets retrieves golden datasets for a project
func (r *GoldenDatasetRepository) ListDatasets(ctx context.Context, projectID uuid.UUID) ([]domain.GoldenDataset, error) {
	query := `
		SELECT id, project_id, name, description, category, language, items, item_count, created_at, created_by
		FROM golden_datasets
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list golden datasets: %w", err)
	}
	defer rows.Close()

	var datasets []domain.GoldenDataset
	for rows.Next() {
		var dataset domain.GoldenDataset
		var itemsJSON []byte

		if err := rows.Scan(
			&dataset.ID,
			&dataset.ProjectID,
			&dataset.Name,
			&dataset.Description,
			&dataset.Category,
			&dataset.Language,
			&itemsJSON,
			&dataset.ItemCount,
			&dataset.CreatedAt,
			&dataset.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan golden dataset: %w", err)
		}

		if len(itemsJSON) > 0 {
			if err := json.Unmarshal(itemsJSON, &dataset.Items); err != nil {
				return nil, fmt.Errorf("failed to unmarshal items: %w", err)
			}
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

// SaveRun creates a new regression run
func (r *GoldenDatasetRepository) SaveRun(ctx context.Context, run *domain.RegressionRun) error {
	resultsJSON, err := json.Marshal(run.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	baselineJSON, err := json.Marshal(run.BaselineComparison)
	if err != nil {
		return fmt.Errorf("failed to marshal baseline_comparison: %w", err)
	}

	query := `
		INSERT INTO regression_runs (id, project_id, suite_id, agent_config, status, results, pass_rate, total_tests, passed, failed, baseline_comparison, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		run.ID,
		run.ProjectID,
		run.SuiteID,
		run.AgentConfig,
		run.Status,
		resultsJSON,
		run.PassRate,
		run.TotalTests,
		run.Passed,
		run.Failed,
		baselineJSON,
		run.StartedAt,
		run.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save regression run: %w", err)
	}

	return nil
}

// GetRunByID retrieves a regression run by ID
func (r *GoldenDatasetRepository) GetRunByID(ctx context.Context, id uuid.UUID) (*domain.RegressionRun, error) {
	query := `
		SELECT id, project_id, suite_id, agent_config, status, results, pass_rate, total_tests, passed, failed, baseline_comparison, started_at, completed_at
		FROM regression_runs
		WHERE id = $1
	`

	var run domain.RegressionRun
	var resultsJSON, baselineJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&run.ID,
		&run.ProjectID,
		&run.SuiteID,
		&run.AgentConfig,
		&run.Status,
		&resultsJSON,
		&run.PassRate,
		&run.TotalTests,
		&run.Passed,
		&run.Failed,
		&baselineJSON,
		&run.StartedAt,
		&run.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("regression run")
		}
		return nil, fmt.Errorf("failed to get regression run: %w", err)
	}

	if len(resultsJSON) > 0 {
		if err := json.Unmarshal(resultsJSON, &run.Results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}
	}
	if len(baselineJSON) > 0 {
		if err := json.Unmarshal(baselineJSON, &run.BaselineComparison); err != nil {
			return nil, fmt.Errorf("failed to unmarshal baseline_comparison: %w", err)
		}
	}

	return &run, nil
}

// ListRuns retrieves regression runs for a project
func (r *GoldenDatasetRepository) ListRuns(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.RegressionRun, error) {
	query := `
		SELECT id, project_id, suite_id, agent_config, status, results, pass_rate, total_tests, passed, failed, baseline_comparison, started_at, completed_at
		FROM regression_runs
		WHERE project_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list regression runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.RegressionRun
	for rows.Next() {
		var run domain.RegressionRun
		var resultsJSON, baselineJSON []byte

		if err := rows.Scan(
			&run.ID,
			&run.ProjectID,
			&run.SuiteID,
			&run.AgentConfig,
			&run.Status,
			&resultsJSON,
			&run.PassRate,
			&run.TotalTests,
			&run.Passed,
			&run.Failed,
			&baselineJSON,
			&run.StartedAt,
			&run.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan regression run: %w", err)
		}

		if len(resultsJSON) > 0 {
			if err := json.Unmarshal(resultsJSON, &run.Results); err != nil {
				return nil, fmt.Errorf("failed to unmarshal results: %w", err)
			}
		}
		if len(baselineJSON) > 0 {
			if err := json.Unmarshal(baselineJSON, &run.BaselineComparison); err != nil {
				return nil, fmt.Errorf("failed to unmarshal baseline_comparison: %w", err)
			}
		}

		runs = append(runs, run)
	}

	return runs, nil
}
