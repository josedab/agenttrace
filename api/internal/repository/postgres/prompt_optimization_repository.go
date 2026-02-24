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

// PromptOptimizationRepository handles prompt optimization data operations in PostgreSQL
type PromptOptimizationRepository struct {
	db *database.PostgresDB
}

// NewPromptOptimizationRepository creates a new prompt optimization repository
func NewPromptOptimizationRepository(db *database.PostgresDB) *PromptOptimizationRepository {
	return &PromptOptimizationRepository{db: db}
}

// SaveOptimization creates a new prompt optimization
func (r *PromptOptimizationRepository) SaveOptimization(ctx context.Context, opt *domain.ContinuousPromptOptimization) error {
	failurePatternsJSON, err := json.Marshal(opt.FailurePatterns)
	if err != nil {
		return fmt.Errorf("failed to marshal failure_patterns: %w", err)
	}

	variantsJSON, err := json.Marshal(opt.Variants)
	if err != nil {
		return fmt.Errorf("failed to marshal variants: %w", err)
	}

	query := `
		INSERT INTO prompt_optimizations (id, project_id, prompt_id, prompt_version, status, failure_patterns, variants, best_variant_id, improvement_pct, started_at, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		opt.ID,
		opt.ProjectID,
		opt.PromptID,
		opt.PromptVersion,
		opt.Status,
		failurePatternsJSON,
		variantsJSON,
		opt.BestVariantID,
		opt.ImprovementPct,
		opt.StartedAt,
		opt.CompletedAt,
		opt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save prompt optimization: %w", err)
	}

	return nil
}

// GetOptimizationByID retrieves a prompt optimization by ID
func (r *PromptOptimizationRepository) GetOptimizationByID(ctx context.Context, id uuid.UUID) (*domain.ContinuousPromptOptimization, error) {
	query := `
		SELECT id, project_id, prompt_id, prompt_version, status, failure_patterns, variants, best_variant_id, improvement_pct, started_at, completed_at, created_at
		FROM prompt_optimizations
		WHERE id = $1
	`

	var opt domain.ContinuousPromptOptimization
	var failurePatternsJSON, variantsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&opt.ID,
		&opt.ProjectID,
		&opt.PromptID,
		&opt.PromptVersion,
		&opt.Status,
		&failurePatternsJSON,
		&variantsJSON,
		&opt.BestVariantID,
		&opt.ImprovementPct,
		&opt.StartedAt,
		&opt.CompletedAt,
		&opt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt optimization")
		}
		return nil, fmt.Errorf("failed to get prompt optimization: %w", err)
	}

	if len(failurePatternsJSON) > 0 {
		if err := json.Unmarshal(failurePatternsJSON, &opt.FailurePatterns); err != nil {
			return nil, fmt.Errorf("failed to unmarshal failure_patterns: %w", err)
		}
	}
	if len(variantsJSON) > 0 {
		if err := json.Unmarshal(variantsJSON, &opt.Variants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
		}
	}

	return &opt, nil
}

// ListOptimizations retrieves prompt optimizations for a project
func (r *PromptOptimizationRepository) ListOptimizations(ctx context.Context, projectID uuid.UUID) ([]domain.ContinuousPromptOptimization, error) {
	query := `
		SELECT id, project_id, prompt_id, prompt_version, status, failure_patterns, variants, best_variant_id, improvement_pct, started_at, completed_at, created_at
		FROM prompt_optimizations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompt optimizations: %w", err)
	}
	defer rows.Close()

	var optimizations []domain.ContinuousPromptOptimization
	for rows.Next() {
		var opt domain.ContinuousPromptOptimization
		var failurePatternsJSON, variantsJSON []byte

		if err := rows.Scan(
			&opt.ID,
			&opt.ProjectID,
			&opt.PromptID,
			&opt.PromptVersion,
			&opt.Status,
			&failurePatternsJSON,
			&variantsJSON,
			&opt.BestVariantID,
			&opt.ImprovementPct,
			&opt.StartedAt,
			&opt.CompletedAt,
			&opt.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prompt optimization: %w", err)
		}

		if len(failurePatternsJSON) > 0 {
			if err := json.Unmarshal(failurePatternsJSON, &opt.FailurePatterns); err != nil {
				return nil, fmt.Errorf("failed to unmarshal failure_patterns: %w", err)
			}
		}
		if len(variantsJSON) > 0 {
			if err := json.Unmarshal(variantsJSON, &opt.Variants); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
			}
		}

		optimizations = append(optimizations, opt)
	}

	return optimizations, nil
}

// UpdateOptimization updates an existing prompt optimization
func (r *PromptOptimizationRepository) UpdateOptimization(ctx context.Context, opt *domain.ContinuousPromptOptimization) error {
	failurePatternsJSON, err := json.Marshal(opt.FailurePatterns)
	if err != nil {
		return fmt.Errorf("failed to marshal failure_patterns: %w", err)
	}

	variantsJSON, err := json.Marshal(opt.Variants)
	if err != nil {
		return fmt.Errorf("failed to marshal variants: %w", err)
	}

	query := `
		UPDATE prompt_optimizations
		SET status = $2, failure_patterns = $3, variants = $4, best_variant_id = $5, improvement_pct = $6, started_at = $7, completed_at = $8
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		opt.ID,
		opt.Status,
		failurePatternsJSON,
		variantsJSON,
		opt.BestVariantID,
		opt.ImprovementPct,
		opt.StartedAt,
		opt.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update prompt optimization: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("prompt optimization")
	}

	return nil
}

// SaveConfig creates a new optimization config
func (r *PromptOptimizationRepository) SaveConfig(ctx context.Context, config *domain.OptimizationConfig) error {
	query := `
		INSERT INTO optimization_configs (id, project_id, enabled, min_samples_for_analysis, min_samples_for_promotion, p_value_threshold, require_approval, max_variants_per_round, schedule_cron)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		config.ID,
		config.ProjectID,
		config.Enabled,
		config.MinSamplesForAnalysis,
		config.MinSamplesForPromotion,
		config.PValueThreshold,
		config.RequireApproval,
		config.MaxVariantsPerRound,
		config.ScheduleCron,
	)
	if err != nil {
		return fmt.Errorf("failed to save optimization config: %w", err)
	}

	return nil
}

// GetConfig retrieves the optimization config for a project
func (r *PromptOptimizationRepository) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.OptimizationConfig, error) {
	query := `
		SELECT id, project_id, enabled, min_samples_for_analysis, min_samples_for_promotion, p_value_threshold, require_approval, max_variants_per_round, schedule_cron
		FROM optimization_configs
		WHERE project_id = $1
	`

	var config domain.OptimizationConfig

	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(
		&config.ID,
		&config.ProjectID,
		&config.Enabled,
		&config.MinSamplesForAnalysis,
		&config.MinSamplesForPromotion,
		&config.PValueThreshold,
		&config.RequireApproval,
		&config.MaxVariantsPerRound,
		&config.ScheduleCron,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("optimization config")
		}
		return nil, fmt.Errorf("failed to get optimization config: %w", err)
	}

	return &config, nil
}

// UpdateConfig updates an existing optimization config
func (r *PromptOptimizationRepository) UpdateConfig(ctx context.Context, config *domain.OptimizationConfig) error {
	query := `
		UPDATE optimization_configs
		SET enabled = $2, min_samples_for_analysis = $3, min_samples_for_promotion = $4, p_value_threshold = $5, require_approval = $6, max_variants_per_round = $7, schedule_cron = $8
		WHERE project_id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		config.ProjectID,
		config.Enabled,
		config.MinSamplesForAnalysis,
		config.MinSamplesForPromotion,
		config.PValueThreshold,
		config.RequireApproval,
		config.MaxVariantsPerRound,
		config.ScheduleCron,
	)
	if err != nil {
		return fmt.Errorf("failed to update optimization config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("optimization config")
	}

	return nil
}
