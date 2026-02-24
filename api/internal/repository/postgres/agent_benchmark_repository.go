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

type AgentBenchmarkRepository struct {
	db *database.PostgresDB
}

func NewAgentBenchmarkRepository(db *database.PostgresDB) *AgentBenchmarkRepository {
	return &AgentBenchmarkRepository{db: db}
}

func (r *AgentBenchmarkRepository) SaveSuite(ctx context.Context, suite *domain.AgentBenchmarkSuite) error {
	tasksJSON, err := json.Marshal(suite.Tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	query := `
		INSERT INTO agent_benchmark_suites (id, project_id, name, description, category,
			tasks, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		suite.ID,
		suite.ProjectID,
		suite.Name,
		suite.Description,
		suite.Category,
		tasksJSON,
		suite.CreatedAt,
		suite.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save agent benchmark suite: %w", err)
	}

	return nil
}

func (r *AgentBenchmarkRepository) GetSuiteByID(ctx context.Context, id uuid.UUID) (*domain.AgentBenchmarkSuite, error) {
	query := `
		SELECT id, project_id, name, description, category, tasks, created_at, created_by
		FROM agent_benchmark_suites
		WHERE id = $1
	`

	var s domain.AgentBenchmarkSuite
	var tasksJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.ProjectID, &s.Name, &s.Description, &s.Category,
		&tasksJSON, &s.CreatedAt, &s.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("agent benchmark suite")
		}
		return nil, fmt.Errorf("failed to get agent benchmark suite: %w", err)
	}

	if len(tasksJSON) > 0 {
		if err := json.Unmarshal(tasksJSON, &s.Tasks); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
		}
	}

	return &s, nil
}

func (r *AgentBenchmarkRepository) ListSuites(ctx context.Context, projectID uuid.UUID) ([]domain.AgentBenchmarkSuite, error) {
	query := `
		SELECT id, project_id, name, description, category, tasks, created_at, created_by
		FROM agent_benchmark_suites
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent benchmark suites: %w", err)
	}
	defer rows.Close()

	var suites []domain.AgentBenchmarkSuite
	for rows.Next() {
		var s domain.AgentBenchmarkSuite
		var tasksJSON []byte

		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.Name, &s.Description, &s.Category,
			&tasksJSON, &s.CreatedAt, &s.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan agent benchmark suite: %w", err)
		}

		if len(tasksJSON) > 0 {
			if err := json.Unmarshal(tasksJSON, &s.Tasks); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
			}
		}

		suites = append(suites, s)
	}

	return suites, nil
}

func (r *AgentBenchmarkRepository) SaveRun(ctx context.Context, run *domain.BenchmarkRun) error {
	resultsJSON, err := json.Marshal(run.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	query := `
		INSERT INTO agent_benchmark_runs (id, suite_id, project_id, agent_name, model_name,
			results, overall_score, avg_latency_ms, total_cost_usd, started_at, completed_at)
		VALUES ($1, $2, (SELECT project_id FROM agent_benchmark_suites WHERE id = $2),
			$3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		run.ID,
		run.SuiteID,
		run.AgentName,
		run.ModelName,
		resultsJSON,
		run.OverallScore,
		run.AvgLatencyMs,
		run.TotalCostUsd,
		run.StartedAt,
		run.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save benchmark run: %w", err)
	}

	return nil
}

func (r *AgentBenchmarkRepository) ListRuns(ctx context.Context, suiteID uuid.UUID, limit int) ([]domain.BenchmarkRun, error) {
	query := `
		SELECT id, suite_id, agent_name, model_name, results,
			overall_score, avg_latency_ms, total_cost_usd, started_at, completed_at
		FROM agent_benchmark_runs
		WHERE suite_id = $1
		ORDER BY overall_score DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, suiteID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.BenchmarkRun
	for rows.Next() {
		var run domain.BenchmarkRun
		var resultsJSON []byte

		if err := rows.Scan(
			&run.ID, &run.SuiteID, &run.AgentName, &run.ModelName,
			&resultsJSON, &run.OverallScore, &run.AvgLatencyMs,
			&run.TotalCostUsd, &run.StartedAt, &run.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan benchmark run: %w", err)
		}

		if len(resultsJSON) > 0 {
			if err := json.Unmarshal(resultsJSON, &run.Results); err != nil {
				return nil, fmt.Errorf("failed to unmarshal results: %w", err)
			}
		}

		runs = append(runs, run)
	}

	return runs, nil
}
