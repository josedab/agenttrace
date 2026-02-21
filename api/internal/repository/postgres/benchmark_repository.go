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

// BenchmarkRepository handles benchmark data operations in PostgreSQL
type BenchmarkRepository struct {
	db *database.PostgresDB
}

// NewBenchmarkRepository creates a new benchmark repository
func NewBenchmarkRepository(db *database.PostgresDB) *BenchmarkRepository {
	return &BenchmarkRepository{db: db}
}

// Save creates a new benchmark
func (r *BenchmarkRepository) Save(ctx context.Context, benchmark *domain.Benchmark) error {
	metricsJSON, err := json.Marshal(benchmark.Metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	query := `
		INSERT INTO benchmarks (id, name, description, category, dataset_id, evaluator_ids, metrics, is_public, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		benchmark.ID,
		benchmark.Name,
		benchmark.Description,
		benchmark.Category,
		benchmark.DatasetID,
		benchmark.EvaluatorIDs,
		metricsJSON,
		benchmark.IsPublic,
		benchmark.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save benchmark: %w", err)
	}

	return nil
}

// GetByID retrieves a benchmark by ID
func (r *BenchmarkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Benchmark, error) {
	query := `
		SELECT id, name, description, category, dataset_id, evaluator_ids, metrics, is_public, created_at
		FROM benchmarks
		WHERE id = $1
	`

	var benchmark domain.Benchmark
	var metricsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&benchmark.ID,
		&benchmark.Name,
		&benchmark.Description,
		&benchmark.Category,
		&benchmark.DatasetID,
		&benchmark.EvaluatorIDs,
		&metricsJSON,
		&benchmark.IsPublic,
		&benchmark.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("benchmark")
		}
		return nil, fmt.Errorf("failed to get benchmark: %w", err)
	}

	if len(metricsJSON) > 0 {
		if err := json.Unmarshal(metricsJSON, &benchmark.Metrics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
		}
	}

	return &benchmark, nil
}

// List retrieves benchmarks, optionally filtered by category
func (r *BenchmarkRepository) List(ctx context.Context, category *domain.BenchmarkCategory) ([]domain.Benchmark, error) {
	var query string
	var args []interface{}

	if category != nil {
		query = `
			SELECT id, name, description, category, dataset_id, evaluator_ids, metrics, is_public, created_at
			FROM benchmarks
			WHERE category = $1
			ORDER BY created_at DESC
		`
		args = append(args, *category)
	} else {
		query = `
			SELECT id, name, description, category, dataset_id, evaluator_ids, metrics, is_public, created_at
			FROM benchmarks
			ORDER BY created_at DESC
		`
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmarks: %w", err)
	}
	defer rows.Close()

	var benchmarks []domain.Benchmark
	for rows.Next() {
		var b domain.Benchmark
		var metricsJSON []byte

		if err := rows.Scan(
			&b.ID,
			&b.Name,
			&b.Description,
			&b.Category,
			&b.DatasetID,
			&b.EvaluatorIDs,
			&metricsJSON,
			&b.IsPublic,
			&b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan benchmark: %w", err)
		}

		if len(metricsJSON) > 0 {
			if err := json.Unmarshal(metricsJSON, &b.Metrics); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
			}
		}

		benchmarks = append(benchmarks, b)
	}

	return benchmarks, nil
}

// SaveSubmission creates a new benchmark submission
func (r *BenchmarkRepository) SaveSubmission(ctx context.Context, submission *domain.BenchmarkSubmission) error {
	scoresJSON, err := json.Marshal(submission.Scores)
	if err != nil {
		return fmt.Errorf("failed to marshal scores: %w", err)
	}

	query := `
		INSERT INTO benchmark_submissions (id, benchmark_id, project_id, agent_name, agent_version, scores, overall_score, rank, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		submission.ID,
		submission.BenchmarkID,
		submission.ProjectID,
		submission.AgentName,
		submission.AgentVersion,
		scoresJSON,
		submission.OverallScore,
		submission.Rank,
		submission.Metadata,
		submission.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save benchmark submission: %w", err)
	}

	return nil
}

// ListSubmissions retrieves benchmark submissions ordered by overall score
func (r *BenchmarkRepository) ListSubmissions(ctx context.Context, benchmarkID uuid.UUID, limit int) ([]domain.BenchmarkSubmission, error) {
	query := `
		SELECT id, benchmark_id, project_id, agent_name, agent_version, scores, overall_score, rank, metadata, created_at
		FROM benchmark_submissions
		WHERE benchmark_id = $1
		ORDER BY overall_score DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, benchmarkID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark submissions: %w", err)
	}
	defer rows.Close()

	var submissions []domain.BenchmarkSubmission
	for rows.Next() {
		var s domain.BenchmarkSubmission
		var scoresJSON []byte

		if err := rows.Scan(
			&s.ID,
			&s.BenchmarkID,
			&s.ProjectID,
			&s.AgentName,
			&s.AgentVersion,
			&scoresJSON,
			&s.OverallScore,
			&s.Rank,
			&s.Metadata,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan benchmark submission: %w", err)
		}

		if len(scoresJSON) > 0 {
			if err := json.Unmarshal(scoresJSON, &s.Scores); err != nil {
				return nil, fmt.Errorf("failed to unmarshal scores: %w", err)
			}
		}

		submissions = append(submissions, s)
	}

	return submissions, nil
}
