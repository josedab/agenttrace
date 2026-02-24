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

// SecurityScanRepository handles security scan data operations in PostgreSQL
type SecurityScanRepository struct {
	db *database.PostgresDB
}

// NewSecurityScanRepository creates a new security scan repository
func NewSecurityScanRepository(db *database.PostgresDB) *SecurityScanRepository {
	return &SecurityScanRepository{db: db}
}

// SaveScanResult creates a new security scan result
func (r *SecurityScanRepository) SaveScanResult(ctx context.Context, result *domain.SecurityScanResult) error {
	findingsJSON, err := json.Marshal(result.Findings)
	if err != nil {
		return fmt.Errorf("failed to marshal findings: %w", err)
	}

	query := `
		INSERT INTO security_scan_results (id, project_id, trace_id, observation_id, findings, overall_risk, scanned_at, scan_duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		result.ID,
		result.ProjectID,
		result.TraceID,
		result.ObservationID,
		findingsJSON,
		result.OverallRisk,
		result.ScannedAt,
		result.ScanDurationMs,
	)
	if err != nil {
		return fmt.Errorf("failed to save security scan result: %w", err)
	}

	return nil
}

// GetScanResult retrieves a security scan result by ID
func (r *SecurityScanRepository) GetScanResult(ctx context.Context, id uuid.UUID) (*domain.SecurityScanResult, error) {
	query := `
		SELECT id, project_id, trace_id, observation_id, findings, overall_risk, scanned_at, scan_duration_ms
		FROM security_scan_results
		WHERE id = $1
	`

	var sr domain.SecurityScanResult
	var findingsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&sr.ID,
		&sr.ProjectID,
		&sr.TraceID,
		&sr.ObservationID,
		&findingsJSON,
		&sr.OverallRisk,
		&sr.ScannedAt,
		&sr.ScanDurationMs,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("security scan result")
		}
		return nil, fmt.Errorf("failed to get security scan result: %w", err)
	}

	if len(findingsJSON) > 0 {
		if err := json.Unmarshal(findingsJSON, &sr.Findings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal findings: %w", err)
		}
	}

	return &sr, nil
}

// ListScanResults retrieves security scan results for a project
func (r *SecurityScanRepository) ListScanResults(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.SecurityScanResult, error) {
	query := `
		SELECT id, project_id, trace_id, observation_id, findings, overall_risk, scanned_at, scan_duration_ms
		FROM security_scan_results
		WHERE project_id = $1
		ORDER BY scanned_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list security scan results: %w", err)
	}
	defer rows.Close()

	var results []domain.SecurityScanResult
	for rows.Next() {
		var sr domain.SecurityScanResult
		var findingsJSON []byte

		if err := rows.Scan(
			&sr.ID,
			&sr.ProjectID,
			&sr.TraceID,
			&sr.ObservationID,
			&findingsJSON,
			&sr.OverallRisk,
			&sr.ScannedAt,
			&sr.ScanDurationMs,
		); err != nil {
			return nil, fmt.Errorf("failed to scan security scan result: %w", err)
		}

		if len(findingsJSON) > 0 {
			if err := json.Unmarshal(findingsJSON, &sr.Findings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal findings: %w", err)
			}
		}

		results = append(results, sr)
	}

	return results, nil
}

// SavePolicy creates a new security policy
func (r *SecurityScanRepository) SavePolicy(ctx context.Context, policy *domain.SecurityPolicy) error {
	rulesJSON, err := json.Marshal(policy.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	excludePatternsJSON, err := json.Marshal(policy.ExcludePatterns)
	if err != nil {
		return fmt.Errorf("failed to marshal exclude_patterns: %w", err)
	}

	query := `
		INSERT INTO security_policies (id, project_id, name, enabled, rules, action, exclude_patterns, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		policy.ID,
		policy.ProjectID,
		policy.Name,
		policy.Enabled,
		rulesJSON,
		policy.Action,
		excludePatternsJSON,
		policy.CreatedAt,
		policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save security policy: %w", err)
	}

	return nil
}

// GetPolicyByID retrieves a security policy by ID
func (r *SecurityScanRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*domain.SecurityPolicy, error) {
	query := `
		SELECT id, project_id, name, enabled, rules, action, exclude_patterns, created_at, updated_at
		FROM security_policies
		WHERE id = $1
	`

	var policy domain.SecurityPolicy
	var rulesJSON, excludePatternsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&policy.ID,
		&policy.ProjectID,
		&policy.Name,
		&policy.Enabled,
		&rulesJSON,
		&policy.Action,
		&excludePatternsJSON,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("security policy")
		}
		return nil, fmt.Errorf("failed to get security policy: %w", err)
	}

	if len(rulesJSON) > 0 {
		if err := json.Unmarshal(rulesJSON, &policy.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
	}
	if len(excludePatternsJSON) > 0 {
		if err := json.Unmarshal(excludePatternsJSON, &policy.ExcludePatterns); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exclude_patterns: %w", err)
		}
	}

	return &policy, nil
}

// ListPolicies retrieves security policies for a project
func (r *SecurityScanRepository) ListPolicies(ctx context.Context, projectID uuid.UUID) ([]domain.SecurityPolicy, error) {
	query := `
		SELECT id, project_id, name, enabled, rules, action, exclude_patterns, created_at, updated_at
		FROM security_policies
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list security policies: %w", err)
	}
	defer rows.Close()

	var policies []domain.SecurityPolicy
	for rows.Next() {
		var policy domain.SecurityPolicy
		var rulesJSON, excludePatternsJSON []byte

		if err := rows.Scan(
			&policy.ID,
			&policy.ProjectID,
			&policy.Name,
			&policy.Enabled,
			&rulesJSON,
			&policy.Action,
			&excludePatternsJSON,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan security policy: %w", err)
		}

		if len(rulesJSON) > 0 {
			if err := json.Unmarshal(rulesJSON, &policy.Rules); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
			}
		}
		if len(excludePatternsJSON) > 0 {
			if err := json.Unmarshal(excludePatternsJSON, &policy.ExcludePatterns); err != nil {
				return nil, fmt.Errorf("failed to unmarshal exclude_patterns: %w", err)
			}
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

// UpdatePolicy updates an existing security policy
func (r *SecurityScanRepository) UpdatePolicy(ctx context.Context, policy *domain.SecurityPolicy) error {
	rulesJSON, err := json.Marshal(policy.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	excludePatternsJSON, err := json.Marshal(policy.ExcludePatterns)
	if err != nil {
		return fmt.Errorf("failed to marshal exclude_patterns: %w", err)
	}

	query := `
		UPDATE security_policies
		SET name = $2, enabled = $3, rules = $4, action = $5, exclude_patterns = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		policy.ID,
		policy.Name,
		policy.Enabled,
		rulesJSON,
		policy.Action,
		excludePatternsJSON,
		policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update security policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("security policy")
	}

	return nil
}
