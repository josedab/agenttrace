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

// TenantRepository handles tenant data operations in PostgreSQL
type TenantRepository struct {
	db *database.PostgresDB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *database.PostgresDB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, usage_limits, current_usage, created_at
		FROM tenants
		WHERE id = $1
	`

	var tenant domain.Tenant
	var usageLimitsJSON, currentUsageJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Plan,
		&usageLimitsJSON,
		&currentUsageJSON,
		&tenant.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("tenant")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if len(usageLimitsJSON) > 0 {
		if err := json.Unmarshal(usageLimitsJSON, &tenant.UsageLimits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal usage limits: %w", err)
		}
	}

	if len(currentUsageJSON) > 0 {
		if err := json.Unmarshal(currentUsageJSON, &tenant.CurrentUsage); err != nil {
			return nil, fmt.Errorf("failed to unmarshal current usage: %w", err)
		}
	}

	return &tenant, nil
}

// Save creates or updates a tenant
func (r *TenantRepository) Save(ctx context.Context, tenant *domain.Tenant) error {
	usageLimitsJSON, err := json.Marshal(tenant.UsageLimits)
	if err != nil {
		return fmt.Errorf("failed to marshal usage limits: %w", err)
	}

	currentUsageJSON, err := json.Marshal(tenant.CurrentUsage)
	if err != nil {
		return fmt.Errorf("failed to marshal current usage: %w", err)
	}

	query := `
		INSERT INTO tenants (id, name, slug, plan, usage_limits, current_usage, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			plan = EXCLUDED.plan,
			usage_limits = EXCLUDED.usage_limits,
			current_usage = EXCLUDED.current_usage
	`

	_, err = r.db.Pool.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Plan,
		usageLimitsJSON,
		currentUsageJSON,
		tenant.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}

	return nil
}

// RecordUsageEvent records a tenant usage event
func (r *TenantRepository) RecordUsageEvent(ctx context.Context, event *domain.UsageEvent) error {
	query := `
		INSERT INTO usage_events (tenant_id, event_type, value, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		event.TenantID,
		event.EventType,
		event.Value,
		event.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to record usage event: %w", err)
	}

	return nil
}

// GetUsage retrieves current usage for a tenant
func (r *TenantRepository) GetUsage(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error) {
	query := `
		SELECT current_usage
		FROM tenants
		WHERE id = $1
	`

	var currentUsageJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, tenantID).Scan(&currentUsageJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("tenant")
		}
		return nil, fmt.Errorf("failed to get tenant usage: %w", err)
	}

	var usage domain.TenantUsage
	if len(currentUsageJSON) > 0 {
		if err := json.Unmarshal(currentUsageJSON, &usage); err != nil {
			return nil, fmt.Errorf("failed to unmarshal current usage: %w", err)
		}
	}

	return &usage, nil
}
