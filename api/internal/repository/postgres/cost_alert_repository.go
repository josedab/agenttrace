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

// CostAlertRepository handles cost alert data operations in PostgreSQL
type CostAlertRepository struct {
	db *database.PostgresDB
}

// NewCostAlertRepository creates a new cost alert repository
func NewCostAlertRepository(db *database.PostgresDB) *CostAlertRepository {
	return &CostAlertRepository{db: db}
}

// SaveRule creates a new cost alert rule
func (r *CostAlertRepository) SaveRule(ctx context.Context, rule *domain.CostAlertRule) error {
	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	conditionJSON, err := json.Marshal(rule.Condition)
	if err != nil {
		return fmt.Errorf("failed to marshal condition: %w", err)
	}

	channelsJSON, err := json.Marshal(rule.Channels)
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}

	query := `
		INSERT INTO cost_alert_rules (id, project_id, name, enabled, severity, actions, condition, channels, cooldown, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		rule.ID,
		rule.ProjectID,
		rule.Name,
		rule.Enabled,
		rule.Severity,
		actionsJSON,
		conditionJSON,
		channelsJSON,
		rule.Cooldown,
		rule.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save cost alert rule: %w", err)
	}

	return nil
}

// ListRules retrieves cost alert rules for a project
func (r *CostAlertRepository) ListRules(ctx context.Context, projectID uuid.UUID) ([]domain.CostAlertRule, error) {
	query := `
		SELECT id, project_id, name, enabled, severity, actions, condition, channels, cooldown, created_at
		FROM cost_alert_rules
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost alert rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.CostAlertRule
	for rows.Next() {
		var rule domain.CostAlertRule
		var actionsJSON, conditionJSON, channelsJSON []byte

		if err := rows.Scan(
			&rule.ID,
			&rule.ProjectID,
			&rule.Name,
			&rule.Enabled,
			&rule.Severity,
			&actionsJSON,
			&conditionJSON,
			&channelsJSON,
			&rule.Cooldown,
			&rule.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost alert rule: %w", err)
		}

		if len(actionsJSON) > 0 {
			if err := json.Unmarshal(actionsJSON, &rule.Actions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
			}
		}
		if len(conditionJSON) > 0 {
			if err := json.Unmarshal(conditionJSON, &rule.Condition); err != nil {
				return nil, fmt.Errorf("failed to unmarshal condition: %w", err)
			}
		}
		if len(channelsJSON) > 0 {
			if err := json.Unmarshal(channelsJSON, &rule.Channels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
			}
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// DeleteRule deletes a cost alert rule by ID
func (r *CostAlertRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cost_alert_rules WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cost alert rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("cost alert rule")
	}

	return nil
}

// SaveAlert creates a new cost alert
func (r *CostAlertRepository) SaveAlert(ctx context.Context, alert *domain.CostAlert) error {
	channelsJSON, err := json.Marshal(alert.Channels)
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}

	query := `
		INSERT INTO cost_alerts (id, project_id, severity, action, title, description, current_cost, threshold_cost, affected_trace_id, affected_model, channels, sent_at, acknowledged_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		alert.ID,
		alert.ProjectID,
		alert.Severity,
		alert.Action,
		alert.Title,
		alert.Description,
		alert.CurrentCost,
		alert.ThresholdCost,
		alert.AffectedTraceID,
		alert.AffectedModel,
		channelsJSON,
		alert.SentAt,
		alert.AcknowledgedAt,
		alert.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save cost alert: %w", err)
	}

	return nil
}

// ListAlerts retrieves cost alerts for a project
func (r *CostAlertRepository) ListAlerts(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.CostAlert, error) {
	query := `
		SELECT id, project_id, severity, action, title, description, current_cost, threshold_cost, affected_trace_id, affected_model, channels, sent_at, acknowledged_at, created_at
		FROM cost_alerts
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.CostAlert
	for rows.Next() {
		var alert domain.CostAlert
		var channelsJSON []byte

		if err := rows.Scan(
			&alert.ID,
			&alert.ProjectID,
			&alert.Severity,
			&alert.Action,
			&alert.Title,
			&alert.Description,
			&alert.CurrentCost,
			&alert.ThresholdCost,
			&alert.AffectedTraceID,
			&alert.AffectedModel,
			&channelsJSON,
			&alert.SentAt,
			&alert.AcknowledgedAt,
			&alert.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost alert: %w", err)
		}

		if len(channelsJSON) > 0 {
			if err := json.Unmarshal(channelsJSON, &alert.Channels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// AcknowledgeAlert marks a cost alert as acknowledged
func (r *CostAlertRepository) AcknowledgeAlert(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE cost_alerts SET acknowledged_at = NOW() WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to acknowledge cost alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("cost alert")
	}

	return nil
}

// SaveCircuitBreaker creates a new circuit breaker config
func (r *CostAlertRepository) SaveCircuitBreaker(ctx context.Context, cb *domain.CircuitBreakerConfig) error {
	fallbackJSON, err := json.Marshal(cb.FallbackModelChain)
	if err != nil {
		return fmt.Errorf("failed to marshal fallback_model_chain: %w", err)
	}

	query := `
		INSERT INTO circuit_breaker_configs (id, project_id, enabled, state, max_cost_per_minute, max_cost_per_hour, fallback_model_chain, cooldown_seconds, last_tripped_at, reset_after_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		cb.ID,
		cb.ProjectID,
		cb.Enabled,
		cb.State,
		cb.MaxCostPerMinute,
		cb.MaxCostPerHour,
		fallbackJSON,
		cb.CooldownSeconds,
		cb.LastTrippedAt,
		cb.ResetAfterSeconds,
	)
	if err != nil {
		return fmt.Errorf("failed to save circuit breaker config: %w", err)
	}

	return nil
}

// GetCircuitBreaker retrieves the circuit breaker config for a project
func (r *CostAlertRepository) GetCircuitBreaker(ctx context.Context, projectID uuid.UUID) (*domain.CircuitBreakerConfig, error) {
	query := `
		SELECT id, project_id, enabled, state, max_cost_per_minute, max_cost_per_hour, fallback_model_chain, cooldown_seconds, last_tripped_at, reset_after_seconds
		FROM circuit_breaker_configs
		WHERE project_id = $1
	`

	var cb domain.CircuitBreakerConfig
	var fallbackJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(
		&cb.ID,
		&cb.ProjectID,
		&cb.Enabled,
		&cb.State,
		&cb.MaxCostPerMinute,
		&cb.MaxCostPerHour,
		&fallbackJSON,
		&cb.CooldownSeconds,
		&cb.LastTrippedAt,
		&cb.ResetAfterSeconds,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("circuit breaker config")
		}
		return nil, fmt.Errorf("failed to get circuit breaker config: %w", err)
	}

	if len(fallbackJSON) > 0 {
		if err := json.Unmarshal(fallbackJSON, &cb.FallbackModelChain); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fallback_model_chain: %w", err)
		}
	}

	return &cb, nil
}

// UpdateCircuitBreaker updates an existing circuit breaker config
func (r *CostAlertRepository) UpdateCircuitBreaker(ctx context.Context, cb *domain.CircuitBreakerConfig) error {
	fallbackJSON, err := json.Marshal(cb.FallbackModelChain)
	if err != nil {
		return fmt.Errorf("failed to marshal fallback_model_chain: %w", err)
	}

	query := `
		UPDATE circuit_breaker_configs
		SET enabled = $2, state = $3, max_cost_per_minute = $4, max_cost_per_hour = $5, fallback_model_chain = $6, cooldown_seconds = $7, last_tripped_at = $8, reset_after_seconds = $9
		WHERE project_id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		cb.ProjectID,
		cb.Enabled,
		cb.State,
		cb.MaxCostPerMinute,
		cb.MaxCostPerHour,
		fallbackJSON,
		cb.CooldownSeconds,
		cb.LastTrippedAt,
		cb.ResetAfterSeconds,
	)
	if err != nil {
		return fmt.Errorf("failed to update circuit breaker config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("circuit breaker config")
	}

	return nil
}
