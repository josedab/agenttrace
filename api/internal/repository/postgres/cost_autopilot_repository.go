package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// CostAutopilotRepository handles persistence for autopilot configurations
type CostAutopilotRepository struct {
	db *database.PostgresDB
}

// NewCostAutopilotRepository creates a new repository
func NewCostAutopilotRepository(db *database.PostgresDB) *CostAutopilotRepository {
	return &CostAutopilotRepository{db: db}
}

// Upsert creates or updates the autopilot config for a project
func (r *CostAutopilotRepository) Upsert(ctx context.Context, config *domain.CostAutopilotConfig) error {
	query := `
		INSERT INTO cost_autopilot_configs (
			id, project_id, enabled, max_budget_daily, max_budget_monthly,
			optimization_level, auto_apply, notify_on_save, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (project_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			max_budget_daily = EXCLUDED.max_budget_daily,
			max_budget_monthly = EXCLUDED.max_budget_monthly,
			optimization_level = EXCLUDED.optimization_level,
			auto_apply = EXCLUDED.auto_apply,
			notify_on_save = EXCLUDED.notify_on_save
	`
	_, err := r.db.Pool.Exec(ctx, query,
		config.ID, config.ProjectID, config.Enabled,
		config.MaxBudgetDaily, config.MaxBudgetMonthly,
		config.OptimizationLevel, config.AutoApply, config.NotifyOnSave,
		config.CreatedAt, config.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert cost autopilot config: %w", err)
	}
	return nil
}

// GetByProject returns the autopilot config for a project
func (r *CostAutopilotRepository) GetByProject(ctx context.Context, projectID uuid.UUID) (*domain.CostAutopilotConfig, error) {
	query := `
		SELECT id, project_id, enabled, max_budget_daily, max_budget_monthly,
			   optimization_level, auto_apply, notify_on_save, created_at, updated_at
		FROM cost_autopilot_configs WHERE project_id = $1
	`
	var config domain.CostAutopilotConfig
	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(
		&config.ID, &config.ProjectID, &config.Enabled,
		&config.MaxBudgetDaily, &config.MaxBudgetMonthly,
		&config.OptimizationLevel, &config.AutoApply, &config.NotifyOnSave,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost autopilot config: %w", err)
	}
	return &config, nil
}
