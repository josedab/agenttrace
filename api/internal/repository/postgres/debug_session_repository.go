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

// DebugSessionRepository handles debug session data operations in PostgreSQL
type DebugSessionRepository struct {
	db *database.PostgresDB
}

// NewDebugSessionRepository creates a new debug session repository
func NewDebugSessionRepository(db *database.PostgresDB) *DebugSessionRepository {
	return &DebugSessionRepository{db: db}
}

// Save creates a new debug session
func (r *DebugSessionRepository) Save(ctx context.Context, session *domain.DebugSession) error {
	breakpointsJSON, err := json.Marshal(session.Breakpoints)
	if err != nil {
		return fmt.Errorf("failed to marshal breakpoints: %w", err)
	}

	annotationsJSON, err := json.Marshal(session.Annotations)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	query := `
		INSERT INTO debug_sessions (id, project_id, trace_id, user_id, status, current_step, total_steps, breakpoints, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		session.ID,
		session.ProjectID,
		session.TraceID,
		session.UserID,
		session.Status,
		session.CurrentStep,
		session.TotalSteps,
		breakpointsJSON,
		annotationsJSON,
		session.CreatedAt,
		session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save debug session: %w", err)
	}

	return nil
}

// Get retrieves a debug session by ID
func (r *DebugSessionRepository) Get(ctx context.Context, id uuid.UUID) (*domain.DebugSession, error) {
	query := `
		SELECT id, project_id, trace_id, user_id, status, current_step, total_steps, breakpoints, annotations, created_at, updated_at
		FROM debug_sessions
		WHERE id = $1
	`

	var session domain.DebugSession
	var breakpointsJSON, annotationsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.ProjectID,
		&session.TraceID,
		&session.UserID,
		&session.Status,
		&session.CurrentStep,
		&session.TotalSteps,
		&breakpointsJSON,
		&annotationsJSON,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("debug session")
		}
		return nil, fmt.Errorf("failed to get debug session: %w", err)
	}

	if len(breakpointsJSON) > 0 {
		if err := json.Unmarshal(breakpointsJSON, &session.Breakpoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal breakpoints: %w", err)
		}
	}

	if len(annotationsJSON) > 0 {
		if err := json.Unmarshal(annotationsJSON, &session.Annotations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	return &session, nil
}

// List retrieves all debug sessions for a project
func (r *DebugSessionRepository) List(ctx context.Context, projectID uuid.UUID) ([]domain.DebugSession, error) {
	query := `
		SELECT id, project_id, trace_id, user_id, status, current_step, total_steps, breakpoints, annotations, created_at, updated_at
		FROM debug_sessions
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list debug sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.DebugSession
	for rows.Next() {
		var s domain.DebugSession
		var breakpointsJSON, annotationsJSON []byte

		if err := rows.Scan(
			&s.ID,
			&s.ProjectID,
			&s.TraceID,
			&s.UserID,
			&s.Status,
			&s.CurrentStep,
			&s.TotalSteps,
			&breakpointsJSON,
			&annotationsJSON,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan debug session: %w", err)
		}

		if len(breakpointsJSON) > 0 {
			if err := json.Unmarshal(breakpointsJSON, &s.Breakpoints); err != nil {
				return nil, fmt.Errorf("failed to unmarshal breakpoints: %w", err)
			}
		}

		if len(annotationsJSON) > 0 {
			if err := json.Unmarshal(annotationsJSON, &s.Annotations); err != nil {
				return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
			}
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

// Delete deletes a debug session
func (r *DebugSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Pool.Exec(ctx, "DELETE FROM debug_sessions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete debug session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("debug session")
	}

	return nil
}
