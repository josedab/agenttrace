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

// GuardrailRepository handles guard rule and violation data operations in PostgreSQL
type GuardrailRepository struct {
	db *database.PostgresDB
}

// NewGuardrailRepository creates a new guardrail repository
func NewGuardrailRepository(db *database.PostgresDB) *GuardrailRepository {
	return &GuardrailRepository{db: db}
}

// SaveRule creates a new guard rule
func (r *GuardrailRepository) SaveRule(ctx context.Context, rule *domain.GuardRule) error {
	query := `
		INSERT INTO guard_rules (id, project_id, name, description, type, config, action, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		rule.ID,
		rule.ProjectID,
		rule.Name,
		rule.Description,
		rule.Type,
		rule.Config,
		rule.Action,
		rule.Enabled,
		rule.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save guard rule: %w", err)
	}

	return nil
}

// GetRuleByID retrieves a guard rule by ID
func (r *GuardrailRepository) GetRuleByID(ctx context.Context, id uuid.UUID) (*domain.GuardRule, error) {
	query := `
		SELECT id, project_id, name, description, type, config, action, enabled, created_at
		FROM guard_rules
		WHERE id = $1
	`

	var rule domain.GuardRule
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rule.ID,
		&rule.ProjectID,
		&rule.Name,
		&rule.Description,
		&rule.Type,
		&rule.Config,
		&rule.Action,
		&rule.Enabled,
		&rule.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("guard rule")
		}
		return nil, fmt.Errorf("failed to get guard rule: %w", err)
	}

	return &rule, nil
}

// UpdateRule updates a guard rule
func (r *GuardrailRepository) UpdateRule(ctx context.Context, rule *domain.GuardRule) error {
	query := `
		UPDATE guard_rules
		SET name = $2, description = $3, type = $4, config = $5, action = $6, enabled = $7
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		rule.ID,
		rule.Name,
		rule.Description,
		rule.Type,
		rule.Config,
		rule.Action,
		rule.Enabled,
	)
	if err != nil {
		return fmt.Errorf("failed to update guard rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("guard rule")
	}

	return nil
}

// DeleteRule deletes a guard rule
func (r *GuardrailRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Pool.Exec(ctx, "DELETE FROM guard_rules WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete guard rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("guard rule")
	}

	return nil
}

// ListRules retrieves all guard rules for a project
func (r *GuardrailRepository) ListRules(ctx context.Context, projectID uuid.UUID) ([]domain.GuardRule, error) {
	query := `
		SELECT id, project_id, name, description, type, config, action, enabled, created_at
		FROM guard_rules
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list guard rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.GuardRule
	for rows.Next() {
		var rule domain.GuardRule
		if err := rows.Scan(
			&rule.ID,
			&rule.ProjectID,
			&rule.Name,
			&rule.Description,
			&rule.Type,
			&rule.Config,
			&rule.Action,
			&rule.Enabled,
			&rule.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan guard rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// ListEnabledRules retrieves all enabled guard rules for a project
func (r *GuardrailRepository) ListEnabledRules(ctx context.Context, projectID uuid.UUID) ([]domain.GuardRule, error) {
	query := `
		SELECT id, project_id, name, description, type, config, action, enabled, created_at
		FROM guard_rules
		WHERE project_id = $1 AND enabled = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled guard rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.GuardRule
	for rows.Next() {
		var rule domain.GuardRule
		if err := rows.Scan(
			&rule.ID,
			&rule.ProjectID,
			&rule.Name,
			&rule.Description,
			&rule.Type,
			&rule.Config,
			&rule.Action,
			&rule.Enabled,
			&rule.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan guard rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveViolation creates a new guard violation
func (r *GuardrailRepository) SaveViolation(ctx context.Context, violation *domain.GuardViolation) error {
	query := `
		INSERT INTO guard_violations (id, project_id, rule_id, trace_id, severity, details, action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		violation.ID,
		violation.ProjectID,
		violation.RuleID,
		violation.TraceID,
		violation.Severity,
		violation.Details,
		violation.Action,
		violation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save guard violation: %w", err)
	}

	return nil
}

// ListViolations retrieves guard violations with filtering and pagination
func (r *GuardrailRepository) ListViolations(ctx context.Context, filter *domain.GuardViolationFilter, limit, offset int) ([]domain.GuardViolation, int64, error) {
	baseQuery := `FROM guard_violations WHERE project_id = $1`
	args := []interface{}{filter.ProjectID}
	argIndex := 2

	if filter.RuleID != nil {
		baseQuery += fmt.Sprintf(" AND rule_id = $%d", argIndex)
		args = append(args, *filter.RuleID)
		argIndex++
	}

	if filter.Severity != nil {
		baseQuery += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, *filter.Severity)
		argIndex++
	}

	if filter.TraceID != nil {
		baseQuery += fmt.Sprintf(" AND trace_id = $%d", argIndex)
		args = append(args, *filter.TraceID)
		argIndex++
	}

	if filter.StartTime != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	// Get count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int64
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count guard violations: %w", err)
	}

	// Get violations
	query := fmt.Sprintf(`
		SELECT id, project_id, rule_id, trace_id, severity, details, action, created_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list guard violations: %w", err)
	}
	defer rows.Close()

	var violations []domain.GuardViolation
	for rows.Next() {
		var v domain.GuardViolation
		if err := rows.Scan(
			&v.ID,
			&v.ProjectID,
			&v.RuleID,
			&v.TraceID,
			&v.Severity,
			&v.Details,
			&v.Action,
			&v.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan guard violation: %w", err)
		}
		violations = append(violations, v)
	}

	return violations, totalCount, nil
}
