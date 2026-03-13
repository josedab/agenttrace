package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockBillingRepository is a mock for BillingRepository
type MockBillingRepository struct {
	mock.Mock
}

func (m *MockBillingRepository) SaveSubscription(ctx context.Context, sub *domain.BillingSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockBillingRepository) GetSubscription(ctx context.Context, id uuid.UUID) (*domain.BillingSubscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BillingSubscription), args.Error(1)
}

func (m *MockBillingRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*domain.BillingSubscription, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BillingSubscription), args.Error(1)
}

func (m *MockBillingRepository) SaveInvoice(ctx context.Context, invoice *domain.BillingInvoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockBillingRepository) ListInvoices(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.BillingInvoice, error) {
	args := m.Called(ctx, tenantID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BillingInvoice), args.Error(1)
}

func newTestBillingService(repo *MockBillingRepository) *BillingService {
	logger, _ := zap.NewDevelopment()
	return NewBillingService(logger, repo, nil)
}

func TestBillingService_HandleInvoicePaid(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	tenantID := uuid.New()
	repo.On("SaveInvoice", ctx, mock.AnythingOfType("*domain.BillingInvoice")).Return(nil)

	payload := map[string]any{
		"tenant_id":         tenantID.String(),
		"amount_cents":      float64(4900),
		"stripe_invoice_id": "inv_test123",
	}

	err := svc.HandleWebhook(ctx, "invoice.paid", payload)
	require.NoError(t, err)

	// Verify invoice was saved with correct fields
	repo.AssertCalled(t, "SaveInvoice", ctx, mock.MatchedBy(func(inv *domain.BillingInvoice) bool {
		return inv.TenantID == tenantID &&
			inv.AmountCents == 4900 &&
			inv.StripeInvoiceID == "inv_test123" &&
			inv.Status == "paid" &&
			inv.PaidAt != nil
	}))
}

func TestBillingService_HandleSubscriptionUpdated(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	tenantID := uuid.New()
	existingSub := &domain.BillingSubscription{
		ID:       uuid.New(),
		TenantID: tenantID,
		PlanSlug: "pro",
		Status:   domain.BillingSubscriptionStatusActive,
	}

	repo.On("GetByTenantID", ctx, tenantID).Return(existingSub, nil)
	repo.On("SaveSubscription", ctx, mock.AnythingOfType("*domain.BillingSubscription")).Return(nil)

	payload := map[string]any{
		"tenant_id": tenantID.String(),
		"status":    "past_due",
	}

	err := svc.HandleWebhook(ctx, "customer.subscription.updated", payload)
	require.NoError(t, err)

	repo.AssertCalled(t, "SaveSubscription", ctx, mock.MatchedBy(func(sub *domain.BillingSubscription) bool {
		return sub.Status == domain.BillingSubscriptionStatus("past_due")
	}))
}

func TestBillingService_HandleSubscriptionDeleted(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	tenantID := uuid.New()
	existingSub := &domain.BillingSubscription{
		ID:       uuid.New(),
		TenantID: tenantID,
		PlanSlug: "pro",
		Status:   domain.BillingSubscriptionStatusActive,
	}

	repo.On("GetByTenantID", ctx, tenantID).Return(existingSub, nil)
	repo.On("SaveSubscription", ctx, mock.AnythingOfType("*domain.BillingSubscription")).Return(nil)

	payload := map[string]any{
		"tenant_id": tenantID.String(),
	}

	err := svc.HandleWebhook(ctx, "customer.subscription.deleted", payload)
	require.NoError(t, err)

	repo.AssertCalled(t, "SaveSubscription", ctx, mock.MatchedBy(func(sub *domain.BillingSubscription) bool {
		return sub.Status == domain.BillingSubscriptionStatusCanceled
	}))
}

func TestBillingService_InvalidPlanSlug(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	_, err := svc.CreateSubscription(ctx, uuid.New(), "nonexistent-plan")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown plan slug")
}

func TestBillingService_CancelSubscription(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	tenantID := uuid.New()
	existingSub := &domain.BillingSubscription{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		PlanSlug:           "pro",
		Status:             domain.BillingSubscriptionStatusActive,
		CancelAtPeriodEnd:  false,
	}

	repo.On("GetByTenantID", ctx, tenantID).Return(existingSub, nil)
	repo.On("SaveSubscription", ctx, mock.AnythingOfType("*domain.BillingSubscription")).Return(nil)

	err := svc.CancelSubscription(ctx, tenantID)
	require.NoError(t, err)

	// Should set CancelAtPeriodEnd without immediately canceling
	repo.AssertCalled(t, "SaveSubscription", ctx, mock.MatchedBy(func(sub *domain.BillingSubscription) bool {
		return sub.CancelAtPeriodEnd == true &&
			sub.Status == domain.BillingSubscriptionStatusActive // still active
	}))
}

func TestBillingService_UpgradePlan_SamePlan(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	tenantID := uuid.New()
	existingSub := &domain.BillingSubscription{
		ID:       uuid.New(),
		TenantID: tenantID,
		PlanSlug: "pro",
		Status:   domain.BillingSubscriptionStatusActive,
	}

	repo.On("GetByTenantID", ctx, tenantID).Return(existingSub, nil)
	repo.On("SaveSubscription", ctx, mock.AnythingOfType("*domain.BillingSubscription")).Return(nil)

	// Upgrade to same plan - currently saves (no-op check in code)
	result, err := svc.UpgradePlan(ctx, tenantID, domain.PlanUpgradeInput{PlanSlug: "pro"})
	require.NoError(t, err)
	assert.Equal(t, "pro", result.PlanSlug)
}

func TestBillingService_UnknownWebhookEventIgnored(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	err := svc.HandleWebhook(ctx, "unknown.event.type", map[string]any{})
	assert.NoError(t, err, "unknown webhook event should be ignored gracefully")
}

func TestBillingService_UpgradePlan_InvalidPlan(t *testing.T) {
	repo := new(MockBillingRepository)
	svc := newTestBillingService(repo)
	ctx := context.Background()

	tenantID := uuid.New()
	existingSub := &domain.BillingSubscription{
		ID:       uuid.New(),
		TenantID: tenantID,
		PlanSlug: "pro",
	}

	repo.On("GetByTenantID", ctx, tenantID).Return(existingSub, nil)

	_, err := svc.UpgradePlan(ctx, tenantID, domain.PlanUpgradeInput{PlanSlug: "invalid-plan"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown plan slug")
}
