package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TenantRepository defines repository operations needed for tenant management
type TenantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	Save(ctx context.Context, tenant *domain.Tenant) error
	RecordUsageEvent(ctx context.Context, event *domain.UsageEvent) error
	GetUsage(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error)
}

// TenantService manages multi-tenant isolation and usage metering
type TenantService struct {
	logger     *zap.Logger
	tenantRepo TenantRepository
}

// NewTenantService creates a new tenant service
func NewTenantService(
	logger *zap.Logger,
	tenantRepo TenantRepository,
) *TenantService {
	return &TenantService{
		logger:     logger,
		tenantRepo: tenantRepo,
	}
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// CheckUsageLimit checks whether a tenant is allowed to perform an action
// based on their plan limits. Returns whether the action is allowed and the remaining quota.
func (s *TenantService) CheckUsageLimit(ctx context.Context, tenantID uuid.UUID, eventType domain.UsageEventType) (bool, int64, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get tenant: %w", err)
	}

	switch eventType {
	case domain.UsageEventTraceIngested:
		limit := int64(tenant.UsageLimits.MaxTracesPerMonth)
		current := int64(tenant.CurrentUsage.TracesThisMonth)
		remaining := limit - current
		if remaining <= 0 {
			s.logger.Warn("tenant trace limit reached",
				zap.String("tenantId", tenantID.String()),
				zap.Int("limit", tenant.UsageLimits.MaxTracesPerMonth),
			)
			return false, 0, nil
		}
		return true, remaining, nil

	case domain.UsageEventStorageUsed:
		limitGB := tenant.UsageLimits.MaxStorageGB
		currentGB := int64(tenant.CurrentUsage.StorageUsedGB)
		remaining := limitGB - currentGB
		if remaining <= 0 {
			s.logger.Warn("tenant storage limit reached",
				zap.String("tenantId", tenantID.String()),
				zap.Int64("limitGB", limitGB),
			)
			return false, 0, nil
		}
		return true, remaining, nil

	default:
		return false, 0, fmt.Errorf("unknown usage event type: %s", eventType)
	}
}

// RecordUsage records a usage event for a tenant
func (s *TenantService) RecordUsage(ctx context.Context, tenantID uuid.UUID, eventType domain.UsageEventType, value int64) error {
	event := &domain.UsageEvent{
		TenantID:  tenantID,
		EventType: eventType,
		Value:     value,
		Timestamp: time.Now(),
	}

	if err := s.tenantRepo.RecordUsageEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to record usage event: %w", err)
	}

	s.logger.Debug("recorded usage event",
		zap.String("tenantId", tenantID.String()),
		zap.String("eventType", string(eventType)),
		zap.Int64("value", value),
	)

	return nil
}

// GetUsageSummary retrieves the current usage summary for a tenant
func (s *TenantService) GetUsageSummary(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error) {
	usage, err := s.tenantRepo.GetUsage(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}
	return usage, nil
}

// EnforceLimits returns the current plan limits for a tenant
func (s *TenantService) EnforceLimits(ctx context.Context, tenantID uuid.UUID) (*domain.TenantLimits, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	limits := planLimits(tenant.Plan)
	return &limits, nil
}

// planLimits returns the default limits for a given plan
func planLimits(plan domain.TenantPlan) domain.TenantLimits {
	switch plan {
	case domain.TenantPlanPro:
		return domain.TenantLimits{
			MaxTracesPerMonth: 1_000_000,
			MaxProjects:       50,
			MaxUsers:          25,
			MaxRetentionDays:  90,
			MaxStorageGB:      100,
		}
	case domain.TenantPlanEnterprise:
		return domain.TenantLimits{
			MaxTracesPerMonth: 10_000_000,
			MaxProjects:       500,
			MaxUsers:          500,
			MaxRetentionDays:  365,
			MaxStorageGB:      1000,
		}
	default: // FREE
		return domain.TenantLimits{
			MaxTracesPerMonth: 50_000,
			MaxProjects:       5,
			MaxUsers:          3,
			MaxRetentionDays:  30,
			MaxStorageGB:      5,
		}
	}
}
