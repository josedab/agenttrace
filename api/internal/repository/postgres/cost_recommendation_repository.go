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

// CostRecommendationRepository handles cost recommendation data operations in PostgreSQL
type CostRecommendationRepository struct {
	db *database.PostgresDB
}

// NewCostRecommendationRepository creates a new cost recommendation repository
func NewCostRecommendationRepository(db *database.PostgresDB) *CostRecommendationRepository {
	return &CostRecommendationRepository{db: db}
}

// Save creates a new cost recommendation
func (r *CostRecommendationRepository) Save(ctx context.Context, rec *domain.CostRecommendation) error {
	query := `
		INSERT INTO cost_recommendations (id, project_id, current_model, recommended_model, trace_count, estimated_savings_per_month, quality_impact_estimate, confidence, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		rec.ID,
		rec.ProjectID,
		rec.CurrentModel,
		rec.RecommendedModel,
		rec.TraceCount,
		rec.EstimatedSavingsPerMonth,
		rec.QualityImpactEstimate,
		rec.Confidence,
		rec.Status,
		rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save cost recommendation: %w", err)
	}

	return nil
}

// GetByID retrieves a cost recommendation by ID
func (r *CostRecommendationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CostRecommendation, error) {
	query := `
		SELECT id, project_id, current_model, recommended_model, trace_count, estimated_savings_per_month, quality_impact_estimate, confidence, status, created_at
		FROM cost_recommendations
		WHERE id = $1
	`

	var rec domain.CostRecommendation
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rec.ID,
		&rec.ProjectID,
		&rec.CurrentModel,
		&rec.RecommendedModel,
		&rec.TraceCount,
		&rec.EstimatedSavingsPerMonth,
		&rec.QualityImpactEstimate,
		&rec.Confidence,
		&rec.Status,
		&rec.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("cost recommendation")
		}
		return nil, fmt.Errorf("failed to get cost recommendation: %w", err)
	}

	return &rec, nil
}

// ListByProject retrieves all cost recommendations for a project
func (r *CostRecommendationRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.CostRecommendation, error) {
	query := `
		SELECT id, project_id, current_model, recommended_model, trace_count, estimated_savings_per_month, quality_impact_estimate, confidence, status, created_at
		FROM cost_recommendations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost recommendations: %w", err)
	}
	defer rows.Close()

	var recs []domain.CostRecommendation
	for rows.Next() {
		var rec domain.CostRecommendation
		if err := rows.Scan(
			&rec.ID,
			&rec.ProjectID,
			&rec.CurrentModel,
			&rec.RecommendedModel,
			&rec.TraceCount,
			&rec.EstimatedSavingsPerMonth,
			&rec.QualityImpactEstimate,
			&rec.Confidence,
			&rec.Status,
			&rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost recommendation: %w", err)
		}
		recs = append(recs, rec)
	}

	return recs, nil
}

// Update updates a cost recommendation
func (r *CostRecommendationRepository) Update(ctx context.Context, rec *domain.CostRecommendation) error {
	query := `
		UPDATE cost_recommendations
		SET current_model = $2, recommended_model = $3, trace_count = $4, estimated_savings_per_month = $5, quality_impact_estimate = $6, confidence = $7, status = $8
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		rec.ID,
		rec.CurrentModel,
		rec.RecommendedModel,
		rec.TraceCount,
		rec.EstimatedSavingsPerMonth,
		rec.QualityImpactEstimate,
		rec.Confidence,
		rec.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update cost recommendation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("cost recommendation")
	}

	return nil
}
