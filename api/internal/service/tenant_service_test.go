package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type mockTenantRepo struct {
	tenant *domain.Tenant
	events []domain.UsageEvent
}

func (m *mockTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Tenant, error) {
	return m.tenant, nil
}

func (m *mockTenantRepo) Save(_ context.Context, _ *domain.Tenant) error {
	return nil
}

func (m *mockTenantRepo) RecordUsageEvent(_ context.Context, event *domain.UsageEvent) error {
	m.events = append(m.events, *event)
	return nil
}

func (m *mockTenantRepo) GetUsage(_ context.Context, _ uuid.UUID) (*domain.TenantUsage, error) {
	return &m.tenant.CurrentUsage, nil
}

func TestTenantService_CheckUsageLimit(t *testing.T) {
	tenantID := uuid.New()

	t.Run("allows when under limit", func(t *testing.T) {
		repo := &mockTenantRepo{
			tenant: &domain.Tenant{
				ID:   tenantID,
				Plan: domain.TenantPlanFree,
				UsageLimits: domain.TenantLimits{
					MaxTracesPerMonth: 50000,
				},
				CurrentUsage: domain.TenantUsage{
					TracesThisMonth: 100,
				},
			},
		}
		svc := NewTenantService(zap.NewNop(), repo)

		allowed, remaining, err := svc.CheckUsageLimit(context.Background(), tenantID, domain.UsageEventTraceIngested)
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, int64(49900), remaining)
	})

	t.Run("blocks when at limit", func(t *testing.T) {
		repo := &mockTenantRepo{
			tenant: &domain.Tenant{
				ID:   tenantID,
				Plan: domain.TenantPlanFree,
				UsageLimits: domain.TenantLimits{
					MaxTracesPerMonth: 50000,
				},
				CurrentUsage: domain.TenantUsage{
					TracesThisMonth: 50000,
				},
			},
		}
		svc := NewTenantService(zap.NewNop(), repo)

		allowed, remaining, err := svc.CheckUsageLimit(context.Background(), tenantID, domain.UsageEventTraceIngested)
		require.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, int64(0), remaining)
	})

	t.Run("rejects unknown event type", func(t *testing.T) {
		repo := &mockTenantRepo{
			tenant: &domain.Tenant{ID: tenantID, Plan: domain.TenantPlanFree},
		}
		svc := NewTenantService(zap.NewNop(), repo)

		_, _, err := svc.CheckUsageLimit(context.Background(), tenantID, "unknown_type")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown usage event type")
	})
}

func TestTenantService_RecordUsage(t *testing.T) {
	tenantID := uuid.New()
	repo := &mockTenantRepo{
		tenant: &domain.Tenant{ID: tenantID},
	}
	svc := NewTenantService(zap.NewNop(), repo)

	err := svc.RecordUsage(context.Background(), tenantID, domain.UsageEventTraceIngested, 1)
	require.NoError(t, err)
	assert.Len(t, repo.events, 1)
	assert.Equal(t, domain.UsageEventTraceIngested, repo.events[0].EventType)
}

func TestPlanLimits(t *testing.T) {
	free := planLimits(domain.TenantPlanFree)
	assert.Equal(t, 50_000, free.MaxTracesPerMonth)
	assert.Equal(t, 5, free.MaxProjects)

	pro := planLimits(domain.TenantPlanPro)
	assert.Equal(t, 1_000_000, pro.MaxTracesPerMonth)
	assert.Equal(t, 50, pro.MaxProjects)

	enterprise := planLimits(domain.TenantPlanEnterprise)
	assert.Equal(t, 10_000_000, enterprise.MaxTracesPerMonth)
	assert.Equal(t, 500, enterprise.MaxProjects)
}
