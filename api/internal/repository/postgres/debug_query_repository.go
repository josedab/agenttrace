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

// DebugQueryRepository handles debug query data operations in PostgreSQL
type DebugQueryRepository struct {
	db *database.PostgresDB
}

// NewDebugQueryRepository creates a new debug query repository
func NewDebugQueryRepository(db *database.PostgresDB) *DebugQueryRepository {
	return &DebugQueryRepository{db: db}
}

// Save creates a new debug query
func (r *DebugQueryRepository) Save(ctx context.Context, dq *domain.DebugQuery) error {
	contextJSON, err := json.Marshal(dq.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	responseJSON, err := json.Marshal(dq.Response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	query := `
		INSERT INTO debug_queries (id, project_id, trace_id, query, query_type, context, response, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		dq.ID,
		dq.ProjectID,
		dq.TraceID,
		dq.Query,
		dq.QueryType,
		contextJSON,
		responseJSON,
		dq.CreatedAt,
		dq.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save debug query: %w", err)
	}

	return nil
}

// GetByID retrieves a debug query by ID
func (r *DebugQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DebugQuery, error) {
	query := `
		SELECT id, project_id, trace_id, query, query_type, context, response, created_at, created_by
		FROM debug_queries
		WHERE id = $1
	`

	var dq domain.DebugQuery
	var contextJSON, responseJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&dq.ID,
		&dq.ProjectID,
		&dq.TraceID,
		&dq.Query,
		&dq.QueryType,
		&contextJSON,
		&responseJSON,
		&dq.CreatedAt,
		&dq.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("debug query")
		}
		return nil, fmt.Errorf("failed to get debug query: %w", err)
	}

	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &dq.Context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}
	}
	if len(responseJSON) > 0 {
		if err := json.Unmarshal(responseJSON, &dq.Response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return &dq, nil
}

// ListByTrace retrieves debug queries for a specific trace in a project
func (r *DebugQueryRepository) ListByTrace(ctx context.Context, projectID uuid.UUID, traceID uuid.UUID, limit int) ([]domain.DebugQuery, error) {
	query := `
		SELECT id, project_id, trace_id, query, query_type, context, response, created_at, created_by
		FROM debug_queries
		WHERE project_id = $1 AND trace_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, traceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list debug queries: %w", err)
	}
	defer rows.Close()

	var queries []domain.DebugQuery
	for rows.Next() {
		var dq domain.DebugQuery
		var contextJSON, responseJSON []byte

		if err := rows.Scan(
			&dq.ID,
			&dq.ProjectID,
			&dq.TraceID,
			&dq.Query,
			&dq.QueryType,
			&contextJSON,
			&responseJSON,
			&dq.CreatedAt,
			&dq.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan debug query: %w", err)
		}

		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &dq.Context); err != nil {
				return nil, fmt.Errorf("failed to unmarshal context: %w", err)
			}
		}
		if len(responseJSON) > 0 {
			if err := json.Unmarshal(responseJSON, &dq.Response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		queries = append(queries, dq)
	}

	return queries, nil
}
