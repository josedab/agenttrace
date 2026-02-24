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

// CloudOnboardingRepository handles cloud onboarding data operations in PostgreSQL
type CloudOnboardingRepository struct {
	db *database.PostgresDB
}

// NewCloudOnboardingRepository creates a new cloud onboarding repository
func NewCloudOnboardingRepository(db *database.PostgresDB) *CloudOnboardingRepository {
	return &CloudOnboardingRepository{db: db}
}

// SaveOnboarding creates a new cloud onboarding record
func (r *CloudOnboardingRepository) SaveOnboarding(ctx context.Context, onboarding *domain.CloudOnboarding) error {
	stepsJSON, err := json.Marshal(onboarding.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	query := `
		INSERT INTO cloud_onboarding (id, tenant_id, steps, current_step, sdk_language, framework, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		onboarding.ID,
		onboarding.TenantID,
		stepsJSON,
		onboarding.CurrentStep,
		onboarding.SDKLanguage,
		onboarding.Framework,
		onboarding.CompletedAt,
		onboarding.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save cloud onboarding: %w", err)
	}

	return nil
}

// GetByTenantID retrieves a cloud onboarding record by tenant ID
func (r *CloudOnboardingRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*domain.CloudOnboarding, error) {
	query := `
		SELECT id, tenant_id, steps, current_step, sdk_language, framework, completed_at, created_at
		FROM cloud_onboarding
		WHERE tenant_id = $1
	`

	var o domain.CloudOnboarding
	var stepsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, tenantID).Scan(
		&o.ID,
		&o.TenantID,
		&stepsJSON,
		&o.CurrentStep,
		&o.SDKLanguage,
		&o.Framework,
		&o.CompletedAt,
		&o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("cloud onboarding")
		}
		return nil, fmt.Errorf("failed to get cloud onboarding: %w", err)
	}

	if len(stepsJSON) > 0 {
		if err := json.Unmarshal(stepsJSON, &o.Steps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
		}
	}

	return &o, nil
}

// UpdateStep updates the current onboarding step
func (r *CloudOnboardingRepository) UpdateStep(ctx context.Context, tenantID uuid.UUID, step domain.OnboardingStep, stepsJSON []byte) error {
	query := `
		UPDATE cloud_onboarding
		SET current_step = $2, steps = $3
		WHERE tenant_id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, tenantID, step, stepsJSON)
	if err != nil {
		return fmt.Errorf("failed to update onboarding step: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("cloud onboarding")
	}

	return nil
}

// SaveUsageMeter creates a new usage meter record
func (r *CloudOnboardingRepository) SaveUsageMeter(ctx context.Context, meter *domain.UsageMeter) error {
	query := `
		INSERT INTO usage_meters (tenant_id, period, traces_used, traces_limit, storage_used_bytes, storage_limit_bytes, api_calls_used, api_calls_limit, percent_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		meter.TenantID,
		meter.Period,
		meter.TracesUsed,
		meter.TracesLimit,
		meter.StorageUsedBytes,
		meter.StorageLimitBytes,
		meter.APICallsUsed,
		meter.APICallsLimit,
		meter.PercentUsed,
	)
	if err != nil {
		return fmt.Errorf("failed to save usage meter: %w", err)
	}

	return nil
}

// GetUsageMeter retrieves a usage meter by tenant ID and period
func (r *CloudOnboardingRepository) GetUsageMeter(ctx context.Context, tenantID uuid.UUID, period string) (*domain.UsageMeter, error) {
	query := `
		SELECT tenant_id, period, traces_used, traces_limit, storage_used_bytes, storage_limit_bytes, api_calls_used, api_calls_limit, percent_used
		FROM usage_meters
		WHERE tenant_id = $1 AND period = $2
	`

	var meter domain.UsageMeter

	err := r.db.Pool.QueryRow(ctx, query, tenantID, period).Scan(
		&meter.TenantID,
		&meter.Period,
		&meter.TracesUsed,
		&meter.TracesLimit,
		&meter.StorageUsedBytes,
		&meter.StorageLimitBytes,
		&meter.APICallsUsed,
		&meter.APICallsLimit,
		&meter.PercentUsed,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("usage meter")
		}
		return nil, fmt.Errorf("failed to get usage meter: %w", err)
	}

	return &meter, nil
}

// UpdateUsage updates usage counters for a tenant's meter
func (r *CloudOnboardingRepository) UpdateUsage(ctx context.Context, meter *domain.UsageMeter) error {
	query := `
		UPDATE usage_meters
		SET traces_used = $3, storage_used_bytes = $4, api_calls_used = $5, percent_used = $6
		WHERE tenant_id = $1 AND period = $2
	`

	result, err := r.db.Pool.Exec(ctx, query,
		meter.TenantID,
		meter.Period,
		meter.TracesUsed,
		meter.StorageUsedBytes,
		meter.APICallsUsed,
		meter.PercentUsed,
	)
	if err != nil {
		return fmt.Errorf("failed to update usage meter: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("usage meter")
	}

	return nil
}
