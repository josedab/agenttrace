package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// InterventionRepository handles persistence for intervention requests
type InterventionRepository struct {
	db *database.PostgresDB
}

// NewInterventionRepository creates a new repository
func NewInterventionRepository(db *database.PostgresDB) *InterventionRepository {
	return &InterventionRepository{db: db}
}

// Save saves an intervention request
func (r *InterventionRepository) Save(ctx context.Context, req *domain.InterventionRequest) error {
	query := `
		INSERT INTO intervention_requests (id, trace_id, project_id, action, message, user_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		req.ID, req.TraceID.String(), req.ProjectID, req.Action, req.Message,
		req.UserID, req.Status, req.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save intervention request: %w", err)
	}
	return nil
}

// GetPending returns pending interventions for a trace
func (r *InterventionRepository) GetPending(ctx context.Context, traceID uuid.UUID) ([]domain.InterventionRequest, error) {
	query := `
		SELECT id, project_id, action, message, user_id, status, created_at
		FROM intervention_requests WHERE trace_id = $1 AND status = 'pending'
		ORDER BY created_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, traceID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get pending interventions: %w", err)
	}
	defer rows.Close()

	var results []domain.InterventionRequest
	for rows.Next() {
		var req domain.InterventionRequest
		req.TraceID = traceID
		if err := rows.Scan(
			&req.ID, &req.ProjectID, &req.Action, &req.Message,
			&req.UserID, &req.Status, &req.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, req)
	}
	return results, nil
}

// Acknowledge marks an intervention as acknowledged
func (r *InterventionRepository) Acknowledge(ctx context.Context, interventionID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE intervention_requests SET status = 'acknowledged', acknowledged_at = $1 WHERE id = $2`,
		time.Now(), interventionID,
	)
	if err != nil {
		return fmt.Errorf("failed to acknowledge intervention: %w", err)
	}
	return nil
}
