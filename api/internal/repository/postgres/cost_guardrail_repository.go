package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type CostGuardrailRepository struct {
	db *database.PostgresDB
}

func NewCostGuardrailRepository(db *database.PostgresDB) *CostGuardrailRepository {
	return &CostGuardrailRepository{db: db}
}

func (r *CostGuardrailRepository) SavePolicy(ctx context.Context, policy *domain.CostGuardrailPolicy) error {
	modelDowngradeMapJSON, err := json.Marshal(policy.ModelDowngradeMap)
	if err != nil {
		return fmt.Errorf("failed to marshal model downgrade map: %w", err)
	}

	notifyChannelsJSON, err := json.Marshal(policy.NotifyChannels)
	if err != nil {
		return fmt.Errorf("failed to marshal notify channels: %w", err)
	}

	query := `
		INSERT INTO cost_guardrail_policies (id, project_id, name, policy_type, action, enabled,
			budget_limit, budget_period, current_spend, threshold_percent,
			model_downgrade_map, notify_channels, cooldown_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		policy.ID,
		policy.ProjectID,
		policy.Name,
		policy.Type,
		policy.Action,
		policy.Enabled,
		policy.BudgetLimit,
		policy.BudgetPeriod,
		policy.CurrentSpend,
		policy.ThresholdPercent,
		modelDowngradeMapJSON,
		notifyChannelsJSON,
		policy.CooldownMinutes,
		policy.CreatedAt,
		policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save cost guardrail policy: %w", err)
	}

	return nil
}

func (r *CostGuardrailRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*domain.CostGuardrailPolicy, error) {
	query := `
		SELECT id, project_id, name, policy_type, action, enabled,
			budget_limit, budget_period, current_spend, threshold_percent,
			model_downgrade_map, notify_channels, cooldown_minutes, created_at, updated_at
		FROM cost_guardrail_policies
		WHERE id = $1
	`

	var p domain.CostGuardrailPolicy
	var modelDowngradeMapJSON, notifyChannelsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.ProjectID, &p.Name, &p.Type, &p.Action, &p.Enabled,
		&p.BudgetLimit, &p.BudgetPeriod, &p.CurrentSpend, &p.ThresholdPercent,
		&modelDowngradeMapJSON, &notifyChannelsJSON,
		&p.CooldownMinutes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("cost guardrail policy")
		}
		return nil, fmt.Errorf("failed to get cost guardrail policy: %w", err)
	}

	if len(modelDowngradeMapJSON) > 0 {
		if err := json.Unmarshal(modelDowngradeMapJSON, &p.ModelDowngradeMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal model downgrade map: %w", err)
		}
	}
	if len(notifyChannelsJSON) > 0 {
		if err := json.Unmarshal(notifyChannelsJSON, &p.NotifyChannels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal notify channels: %w", err)
		}
	}

	return &p, nil
}

func (r *CostGuardrailRepository) ListPolicies(ctx context.Context, projectID uuid.UUID) ([]domain.CostGuardrailPolicy, error) {
	query := `
		SELECT id, project_id, name, policy_type, action, enabled,
			budget_limit, budget_period, current_spend, threshold_percent,
			model_downgrade_map, notify_channels, cooldown_minutes, created_at, updated_at
		FROM cost_guardrail_policies
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost guardrail policies: %w", err)
	}
	defer rows.Close()

	var policies []domain.CostGuardrailPolicy
	for rows.Next() {
		var p domain.CostGuardrailPolicy
		var modelDowngradeMapJSON, notifyChannelsJSON []byte

		if err := rows.Scan(
			&p.ID, &p.ProjectID, &p.Name, &p.Type, &p.Action, &p.Enabled,
			&p.BudgetLimit, &p.BudgetPeriod, &p.CurrentSpend, &p.ThresholdPercent,
			&modelDowngradeMapJSON, &notifyChannelsJSON,
			&p.CooldownMinutes, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost guardrail policy: %w", err)
		}

		if len(modelDowngradeMapJSON) > 0 {
			if err := json.Unmarshal(modelDowngradeMapJSON, &p.ModelDowngradeMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal model downgrade map: %w", err)
			}
		}
		if len(notifyChannelsJSON) > 0 {
			if err := json.Unmarshal(notifyChannelsJSON, &p.NotifyChannels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal notify channels: %w", err)
			}
		}

		policies = append(policies, p)
	}

	return policies, nil
}

func (r *CostGuardrailRepository) UpdatePolicy(ctx context.Context, policy *domain.CostGuardrailPolicy) error {
	modelDowngradeMapJSON, err := json.Marshal(policy.ModelDowngradeMap)
	if err != nil {
		return fmt.Errorf("failed to marshal model downgrade map: %w", err)
	}

	notifyChannelsJSON, err := json.Marshal(policy.NotifyChannels)
	if err != nil {
		return fmt.Errorf("failed to marshal notify channels: %w", err)
	}

	query := `
		UPDATE cost_guardrail_policies
		SET name = $2, policy_type = $3, action = $4, enabled = $5,
			budget_limit = $6, budget_period = $7, current_spend = $8,
			threshold_percent = $9, model_downgrade_map = $10,
			notify_channels = $11, cooldown_minutes = $12, updated_at = $13
		WHERE id = $1
	`

	_, err = r.db.Pool.Exec(ctx, query,
		policy.ID,
		policy.Name,
		policy.Type,
		policy.Action,
		policy.Enabled,
		policy.BudgetLimit,
		policy.BudgetPeriod,
		policy.CurrentSpend,
		policy.ThresholdPercent,
		modelDowngradeMapJSON,
		notifyChannelsJSON,
		policy.CooldownMinutes,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update cost guardrail policy: %w", err)
	}

	return nil
}

func (r *CostGuardrailRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cost_guardrail_policies WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cost guardrail policy: %w", err)
	}

	return nil
}

func (r *CostGuardrailRepository) SaveViolation(ctx context.Context, violation *domain.CostGuardrailViolation) error {
	query := `
		INSERT INTO cost_guardrail_violations (id, policy_id, project_id, trace_id, user_id,
			action, amount_at_violation, budget_limit, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		violation.ID,
		violation.PolicyID,
		violation.ProjectID,
		violation.TraceID,
		violation.UserID,
		violation.Action,
		violation.AmountAtViolation,
		violation.BudgetLimit,
		violation.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to save cost guardrail violation: %w", err)
	}

	return nil
}

func (r *CostGuardrailRepository) ListViolations(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.CostGuardrailViolation, error) {
	query := `
		SELECT id, policy_id, project_id, trace_id, user_id,
			action, amount_at_violation, budget_limit, timestamp
		FROM cost_guardrail_violations
		WHERE project_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost guardrail violations: %w", err)
	}
	defer rows.Close()

	var violations []domain.CostGuardrailViolation
	for rows.Next() {
		var v domain.CostGuardrailViolation
		if err := rows.Scan(
			&v.ID, &v.PolicyID, &v.ProjectID, &v.TraceID, &v.UserID,
			&v.Action, &v.AmountAtViolation, &v.BudgetLimit, &v.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost guardrail violation: %w", err)
		}
		violations = append(violations, v)
	}

	return violations, nil
}

func (r *CostGuardrailRepository) UpdatePolicySpend(ctx context.Context, id uuid.UUID, spend float64) error {
	query := `UPDATE cost_guardrail_policies SET current_spend = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, spend, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update policy spend: %w", err)
	}

	return nil
}
