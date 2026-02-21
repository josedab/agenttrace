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

// RegressionRepository handles regression test data operations in PostgreSQL
type RegressionRepository struct {
	db *database.PostgresDB
}

// NewRegressionRepository creates a new regression repository
func NewRegressionRepository(db *database.PostgresDB) *RegressionRepository {
	return &RegressionRepository{db: db}
}

// Save creates a new regression test
func (r *RegressionRepository) Save(ctx context.Context, test *domain.RegressionTest) error {
	thresholdsJSON, err := json.Marshal(test.Thresholds)
	if err != nil {
		return fmt.Errorf("failed to marshal thresholds: %w", err)
	}

	query := `
		INSERT INTO regression_tests (id, project_id, name, baseline_dataset_id, evaluator_ids, thresholds, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		test.ID,
		test.ProjectID,
		test.Name,
		test.BaselineDatasetID,
		test.EvaluatorIDs,
		thresholdsJSON,
		test.Status,
		test.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save regression test: %w", err)
	}

	return nil
}

// GetByID retrieves a regression test by ID
func (r *RegressionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RegressionTest, error) {
	query := `
		SELECT id, project_id, name, baseline_dataset_id, evaluator_ids, thresholds, status, created_at
		FROM regression_tests
		WHERE id = $1
	`

	var test domain.RegressionTest
	var thresholdsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&test.ID,
		&test.ProjectID,
		&test.Name,
		&test.BaselineDatasetID,
		&test.EvaluatorIDs,
		&thresholdsJSON,
		&test.Status,
		&test.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("regression test")
		}
		return nil, fmt.Errorf("failed to get regression test: %w", err)
	}

	if len(thresholdsJSON) > 0 {
		if err := json.Unmarshal(thresholdsJSON, &test.Thresholds); err != nil {
			return nil, fmt.Errorf("failed to unmarshal thresholds: %w", err)
		}
	}

	return &test, nil
}

// List retrieves all regression tests for a project
func (r *RegressionRepository) List(ctx context.Context, projectID uuid.UUID) ([]domain.RegressionTest, error) {
	query := `
		SELECT id, project_id, name, baseline_dataset_id, evaluator_ids, thresholds, status, created_at
		FROM regression_tests
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list regression tests: %w", err)
	}
	defer rows.Close()

	var tests []domain.RegressionTest
	for rows.Next() {
		var test domain.RegressionTest
		var thresholdsJSON []byte

		if err := rows.Scan(
			&test.ID,
			&test.ProjectID,
			&test.Name,
			&test.BaselineDatasetID,
			&test.EvaluatorIDs,
			&thresholdsJSON,
			&test.Status,
			&test.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan regression test: %w", err)
		}

		if len(thresholdsJSON) > 0 {
			if err := json.Unmarshal(thresholdsJSON, &test.Thresholds); err != nil {
				return nil, fmt.Errorf("failed to unmarshal thresholds: %w", err)
			}
		}

		tests = append(tests, test)
	}

	return tests, nil
}

// SaveResult creates a new regression result
func (r *RegressionRepository) SaveResult(ctx context.Context, result *domain.RegressionResult) error {
	scoresJSON, err := json.Marshal(result.Scores)
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	baselineScoresJSON, err := json.Marshal(result.BaselineScores)
	if err != nil {
		return fmt.Errorf("failed to marshal baseline scores: %w", err)
	}

	deltasJSON, err := json.Marshal(result.Deltas)
	if err != nil {
		return fmt.Errorf("failed to marshal deltas: %w", err)
	}

	query := `
		INSERT INTO regression_results (id, test_id, run_id, scores, baseline_scores, passed, deltas, failed_metrics, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		result.ID,
		result.TestID,
		result.RunID,
		scoresJSON,
		baselineScoresJSON,
		result.Passed,
		deltasJSON,
		result.FailedMetrics,
		result.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save regression result: %w", err)
	}

	return nil
}

// GetResultByID retrieves a regression result by ID
func (r *RegressionRepository) GetResultByID(ctx context.Context, id uuid.UUID) (*domain.RegressionResult, error) {
	query := `
		SELECT id, test_id, run_id, scores, baseline_scores, passed, deltas, failed_metrics, created_at
		FROM regression_results
		WHERE id = $1
	`

	var result domain.RegressionResult
	var scoresJSON, baselineScoresJSON, deltasJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.TestID,
		&result.RunID,
		&scoresJSON,
		&baselineScoresJSON,
		&result.Passed,
		&deltasJSON,
		&result.FailedMetrics,
		&result.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("regression result")
		}
		return nil, fmt.Errorf("failed to get regression result: %w", err)
	}

	if len(scoresJSON) > 0 {
		if err := json.Unmarshal(scoresJSON, &result.Scores); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scores: %w", err)
		}
	}

	if len(baselineScoresJSON) > 0 {
		if err := json.Unmarshal(baselineScoresJSON, &result.BaselineScores); err != nil {
			return nil, fmt.Errorf("failed to unmarshal baseline scores: %w", err)
		}
	}

	if len(deltasJSON) > 0 {
		if err := json.Unmarshal(deltasJSON, &result.Deltas); err != nil {
			return nil, fmt.Errorf("failed to unmarshal deltas: %w", err)
		}
	}

	return &result, nil
}
