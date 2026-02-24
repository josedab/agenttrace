package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// PrivacyRepository handles persistence for privacy configuration and deletion requests
type PrivacyRepository struct {
	db *database.PostgresDB
}

// NewPrivacyRepository creates a new privacy repository
func NewPrivacyRepository(db *database.PostgresDB) *PrivacyRepository {
	return &PrivacyRepository{db: db}
}

// SaveConfig persists a PII configuration
func (r *PrivacyRepository) SaveConfig(ctx context.Context, config *domain.PIIConfig) error {
	query := `INSERT INTO pii_configs (id, project_id, enabled, sensitivity_level, auto_redact, data_residency, retention_days, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (project_id) DO UPDATE SET
			enabled = EXCLUDED.enabled, sensitivity_level = EXCLUDED.sensitivity_level,
			auto_redact = EXCLUDED.auto_redact, data_residency = EXCLUDED.data_residency,
			retention_days = EXCLUDED.retention_days`
	_, err := r.db.Pool.Exec(ctx, query, config.ID, config.ProjectID, config.Enabled,
		config.SensitivityLevel, config.AutoRedact, config.DataResidency, config.RetentionDays, config.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save PII config: %w", err)
	}
	return nil
}

// GetConfig retrieves the PII configuration for a project
func (r *PrivacyRepository) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.PIIConfig, error) {
	query := `SELECT id, project_id, enabled, sensitivity_level, auto_redact, data_residency, retention_days, created_at
		FROM pii_configs WHERE project_id = $1`
	var c domain.PIIConfig
	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(&c.ID, &c.ProjectID, &c.Enabled,
		&c.SensitivityLevel, &c.AutoRedact, &c.DataResidency, &c.RetentionDays, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get PII config: %w", err)
	}
	return &c, nil
}

// SaveDeletionRequest persists a data deletion request
func (r *PrivacyRepository) SaveDeletionRequest(ctx context.Context, req *domain.DataDeletionRequest) error {
	query := `INSERT INTO data_deletion_requests (id, project_id, request_type, target_id, status, requested_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Pool.Exec(ctx, query, req.ID, req.ProjectID, req.RequestType, req.TargetID,
		req.Status, req.RequestedAt, req.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to save deletion request: %w", err)
	}
	return nil
}

// ListDeletionRequests returns all deletion requests for a project
func (r *PrivacyRepository) ListDeletionRequests(ctx context.Context, projectID uuid.UUID) ([]domain.DataDeletionRequest, error) {
	query := `SELECT id, project_id, request_type, target_id, status, requested_at, completed_at
		FROM data_deletion_requests WHERE project_id = $1 ORDER BY requested_at DESC LIMIT 100`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deletion requests: %w", err)
	}
	defer rows.Close()
	var requests []domain.DataDeletionRequest
	for rows.Next() {
		var req domain.DataDeletionRequest
		if err := rows.Scan(&req.ID, &req.ProjectID, &req.RequestType, &req.TargetID,
			&req.Status, &req.RequestedAt, &req.CompletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan deletion request: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, nil
}
