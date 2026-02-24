package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// PredictiveCostRepository handles persistence for cost predictions and budget approvals
type PredictiveCostRepository struct {
	db *database.PostgresDB
}

// NewPredictiveCostRepository creates a new predictive cost repository
func NewPredictiveCostRepository(db *database.PostgresDB) *PredictiveCostRepository {
	return &PredictiveCostRepository{db: db}
}

// SavePrediction persists a cost prediction
func (r *PredictiveCostRepository) SavePrediction(ctx context.Context, pred *domain.CostPrediction) error {
	query := `INSERT INTO cost_predictions (id, project_id, task_description, predicted_cost, predicted_latency_ms,
		predicted_quality, predicted_tokens, confidence_level, recommended_model, budget_status, similar_traces, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Pool.Exec(ctx, query, pred.ID, pred.ProjectID, pred.TaskDescription,
		pred.PredictedCost, pred.PredictedLatencyMs, pred.PredictedQuality, pred.PredictedTokens,
		pred.ConfidenceLevel, pred.RecommendedModel, pred.BudgetStatus, pred.SimilarTraces, pred.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save cost prediction: %w", err)
	}
	return nil
}

// ListByProject returns cost predictions for a project
func (r *PredictiveCostRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.CostPrediction, error) {
	query := `SELECT id, project_id, task_description, predicted_cost, predicted_latency_ms,
		predicted_quality, predicted_tokens, confidence_level, recommended_model, budget_status, similar_traces, created_at
		FROM cost_predictions WHERE project_id = $1 ORDER BY created_at DESC LIMIT 100`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost predictions: %w", err)
	}
	defer rows.Close()
	var predictions []domain.CostPrediction
	for rows.Next() {
		var p domain.CostPrediction
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.TaskDescription, &p.PredictedCost,
			&p.PredictedLatencyMs, &p.PredictedQuality, &p.PredictedTokens, &p.ConfidenceLevel,
			&p.RecommendedModel, &p.BudgetStatus, &p.SimilarTraces, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cost prediction: %w", err)
		}
		predictions = append(predictions, p)
	}
	return predictions, nil
}

// SaveApproval persists a budget approval
func (r *PredictiveCostRepository) SaveApproval(ctx context.Context, approval *domain.BudgetApproval) error {
	query := `INSERT INTO budget_approvals (id, project_id, prediction_id, status, approver_id, note, created_at, decided_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Pool.Exec(ctx, query, approval.ID, approval.ProjectID, approval.PredictionID,
		approval.Status, approval.ApproverID, approval.Note, approval.CreatedAt, approval.DecidedAt)
	if err != nil {
		return fmt.Errorf("failed to save budget approval: %w", err)
	}
	return nil
}

// GetApproval retrieves a budget approval by ID
func (r *PredictiveCostRepository) GetApproval(ctx context.Context, id uuid.UUID) (*domain.BudgetApproval, error) {
	query := `SELECT id, project_id, prediction_id, status, approver_id, note, created_at, decided_at
		FROM budget_approvals WHERE id = $1`
	var a domain.BudgetApproval
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&a.ID, &a.ProjectID, &a.PredictionID,
		&a.Status, &a.ApproverID, &a.Note, &a.CreatedAt, &a.DecidedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget approval: %w", err)
	}
	return &a, nil
}

// UpdateApproval updates the status of a budget approval
func (r *PredictiveCostRepository) UpdateApproval(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE budget_approvals SET status = $1, decided_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update budget approval: %w", err)
	}
	return nil
}
