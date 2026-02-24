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

type PromptBaselineRepository struct {
	db *database.PostgresDB
}

func NewPromptBaselineRepository(db *database.PostgresDB) *PromptBaselineRepository {
	return &PromptBaselineRepository{db: db}
}

func (r *PromptBaselineRepository) SaveBaseline(ctx context.Context, baseline *domain.PromptBaseline) error {
	scoresJSON, err := json.Marshal(baseline.Scores)
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	query := `
		INSERT INTO prompt_baselines (id, project_id, dataset_id, prompt_id, prompt_version,
			name, branch, scores, sample_size, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		baseline.ID,
		baseline.ProjectID,
		baseline.DatasetID,
		baseline.PromptID,
		baseline.PromptVersion,
		baseline.Name,
		baseline.Branch,
		scoresJSON,
		baseline.SampleSize,
		baseline.CreatedAt,
		baseline.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save prompt baseline: %w", err)
	}

	return nil
}

func (r *PromptBaselineRepository) GetBaselineByID(ctx context.Context, id uuid.UUID) (*domain.PromptBaseline, error) {
	query := `
		SELECT id, project_id, dataset_id, prompt_id, prompt_version,
			name, branch, scores, sample_size, created_at, created_by
		FROM prompt_baselines
		WHERE id = $1
	`

	var b domain.PromptBaseline
	var scoresJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.ProjectID, &b.DatasetID, &b.PromptID, &b.PromptVersion,
		&b.Name, &b.Branch, &scoresJSON, &b.SampleSize,
		&b.CreatedAt, &b.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt baseline")
		}
		return nil, fmt.Errorf("failed to get prompt baseline: %w", err)
	}

	if len(scoresJSON) > 0 {
		if err := json.Unmarshal(scoresJSON, &b.Scores); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scores: %w", err)
		}
	}

	return &b, nil
}

func (r *PromptBaselineRepository) ListBaselines(ctx context.Context, projectID uuid.UUID) ([]domain.PromptBaseline, error) {
	query := `
		SELECT id, project_id, dataset_id, prompt_id, prompt_version,
			name, branch, scores, sample_size, created_at, created_by
		FROM prompt_baselines
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompt baselines: %w", err)
	}
	defer rows.Close()

	var baselines []domain.PromptBaseline
	for rows.Next() {
		var b domain.PromptBaseline
		var scoresJSON []byte

		if err := rows.Scan(
			&b.ID, &b.ProjectID, &b.DatasetID, &b.PromptID, &b.PromptVersion,
			&b.Name, &b.Branch, &scoresJSON, &b.SampleSize,
			&b.CreatedAt, &b.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prompt baseline: %w", err)
		}

		if len(scoresJSON) > 0 {
			if err := json.Unmarshal(scoresJSON, &b.Scores); err != nil {
				return nil, fmt.Errorf("failed to unmarshal scores: %w", err)
			}
		}

		baselines = append(baselines, b)
	}

	return baselines, nil
}

func (r *PromptBaselineRepository) SaveCIRun(ctx context.Context, run *domain.PromptCIRun) error {
	scoreComparisonJSON, err := json.Marshal(run.ScoreComparison)
	if err != nil {
		return fmt.Errorf("failed to marshal score comparison: %w", err)
	}

	query := `
		INSERT INTO prompt_ci_runs (id, project_id, baseline_id, branch, commit_sha,
			pr_number, status, score_comparison, overall_severity, summary,
			started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		run.ID,
		run.ProjectID,
		run.BaselineID,
		run.Branch,
		run.CommitSHA,
		run.PRNumber,
		run.Status,
		scoreComparisonJSON,
		run.OverallSeverity,
		run.Summary,
		run.StartedAt,
		run.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save prompt CI run: %w", err)
	}

	return nil
}

func (r *PromptBaselineRepository) ListCIRuns(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.PromptCIRun, error) {
	query := `
		SELECT id, project_id, baseline_id, branch, commit_sha,
			pr_number, status, score_comparison, overall_severity, summary,
			started_at, completed_at
		FROM prompt_ci_runs
		WHERE project_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompt CI runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.PromptCIRun
	for rows.Next() {
		var run domain.PromptCIRun
		var scoreComparisonJSON []byte

		if err := rows.Scan(
			&run.ID, &run.ProjectID, &run.BaselineID, &run.Branch, &run.CommitSHA,
			&run.PRNumber, &run.Status, &scoreComparisonJSON, &run.OverallSeverity,
			&run.Summary, &run.StartedAt, &run.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prompt CI run: %w", err)
		}

		if len(scoreComparisonJSON) > 0 {
			if err := json.Unmarshal(scoreComparisonJSON, &run.ScoreComparison); err != nil {
				return nil, fmt.Errorf("failed to unmarshal score comparison: %w", err)
			}
		}

		runs = append(runs, run)
	}

	return runs, nil
}

func (r *PromptBaselineRepository) GetCIRunByID(ctx context.Context, id uuid.UUID) (*domain.PromptCIRun, error) {
	query := `
		SELECT id, project_id, baseline_id, branch, commit_sha,
			pr_number, status, score_comparison, overall_severity, summary,
			started_at, completed_at
		FROM prompt_ci_runs
		WHERE id = $1
	`

	var run domain.PromptCIRun
	var scoreComparisonJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&run.ID, &run.ProjectID, &run.BaselineID, &run.Branch, &run.CommitSHA,
		&run.PRNumber, &run.Status, &scoreComparisonJSON, &run.OverallSeverity,
		&run.Summary, &run.StartedAt, &run.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt CI run")
		}
		return nil, fmt.Errorf("failed to get prompt CI run: %w", err)
	}

	if len(scoreComparisonJSON) > 0 {
		if err := json.Unmarshal(scoreComparisonJSON, &run.ScoreComparison); err != nil {
			return nil, fmt.Errorf("failed to unmarshal score comparison: %w", err)
		}
	}

	return &run, nil
}
