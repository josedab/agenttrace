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

// CollaborationRepository handles collaboration data operations in PostgreSQL
type CollaborationRepository struct {
	db *database.PostgresDB
}

// NewCollaborationRepository creates a new collaboration repository
func NewCollaborationRepository(db *database.PostgresDB) *CollaborationRepository {
	return &CollaborationRepository{db: db}
}

// SaveAnnotation creates a new trace annotation
func (r *CollaborationRepository) SaveAnnotation(ctx context.Context, annotation *domain.TraceAnnotation) error {
	query := `
		INSERT INTO trace_annotations (id, project_id, trace_id, event_id, user_id, user_name, content, resolved_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		annotation.ID,
		annotation.ProjectID,
		annotation.TraceID,
		annotation.EventID,
		annotation.UserID,
		annotation.UserName,
		annotation.Content,
		annotation.ResolvedAt,
		annotation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save trace annotation: %w", err)
	}

	return nil
}

// GetAnnotationByID retrieves a trace annotation by ID
func (r *CollaborationRepository) GetAnnotationByID(ctx context.Context, id uuid.UUID) (*domain.TraceAnnotation, error) {
	query := `
		SELECT id, project_id, trace_id, event_id, user_id, user_name, content, resolved_at, created_at
		FROM trace_annotations
		WHERE id = $1
	`

	var annotation domain.TraceAnnotation
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&annotation.ID,
		&annotation.ProjectID,
		&annotation.TraceID,
		&annotation.EventID,
		&annotation.UserID,
		&annotation.UserName,
		&annotation.Content,
		&annotation.ResolvedAt,
		&annotation.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("trace annotation")
		}
		return nil, fmt.Errorf("failed to get trace annotation: %w", err)
	}

	return &annotation, nil
}

// ListAnnotations retrieves trace annotations for a project and trace
func (r *CollaborationRepository) ListAnnotations(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TraceAnnotation, error) {
	query := `
		SELECT id, project_id, trace_id, event_id, user_id, user_name, content, resolved_at, created_at
		FROM trace_annotations
		WHERE project_id = $1 AND trace_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list trace annotations: %w", err)
	}
	defer rows.Close()

	var annotations []domain.TraceAnnotation
	for rows.Next() {
		var a domain.TraceAnnotation
		if err := rows.Scan(
			&a.ID,
			&a.ProjectID,
			&a.TraceID,
			&a.EventID,
			&a.UserID,
			&a.UserName,
			&a.Content,
			&a.ResolvedAt,
			&a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trace annotation: %w", err)
		}
		annotations = append(annotations, a)
	}

	return annotations, nil
}

// UpdateAnnotation updates a trace annotation
func (r *CollaborationRepository) UpdateAnnotation(ctx context.Context, annotation *domain.TraceAnnotation) error {
	query := `
		UPDATE trace_annotations
		SET content = $2, resolved_at = $3
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		annotation.ID,
		annotation.Content,
		annotation.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update trace annotation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("trace annotation")
	}

	return nil
}

// SaveSharedSession creates a new shared session
func (r *CollaborationRepository) SaveSharedSession(ctx context.Context, session *domain.SharedSession) error {
	query := `
		INSERT INTO shared_sessions (id, project_id, trace_id, created_by, participants, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		session.ID,
		session.ProjectID,
		session.TraceID,
		session.CreatedBy,
		session.Participants,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save shared session: %w", err)
	}

	return nil
}

// GetSharedSessionByID retrieves a shared session by ID
func (r *CollaborationRepository) GetSharedSessionByID(ctx context.Context, id uuid.UUID) (*domain.SharedSession, error) {
	query := `
		SELECT id, project_id, trace_id, created_by, participants, expires_at, created_at
		FROM shared_sessions
		WHERE id = $1
	`

	var session domain.SharedSession
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.ProjectID,
		&session.TraceID,
		&session.CreatedBy,
		&session.Participants,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("shared session")
		}
		return nil, fmt.Errorf("failed to get shared session: %w", err)
	}

	return &session, nil
}
