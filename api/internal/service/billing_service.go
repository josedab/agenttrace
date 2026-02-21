package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// BillingRepository defines repository operations for billing management
type BillingRepository interface {
	SaveSubscription(ctx context.Context, sub *domain.BillingSubscription) error
	GetSubscription(ctx context.Context, id uuid.UUID) (*domain.BillingSubscription, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*domain.BillingSubscription, error)
	SaveInvoice(ctx context.Context, invoice *domain.BillingInvoice) error
	ListInvoices(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.BillingInvoice, error)
}

// BillingService manages subscription billing for the managed cloud offering,
// including plan management, invoicing, and Stripe webhook processing.
type BillingService struct {
	logger        *zap.Logger
	billingRepo   BillingRepository
	tenantService *TenantService
}

// NewBillingService creates a new billing service
func NewBillingService(
	logger *zap.Logger,
	billingRepo BillingRepository,
	tenantService *TenantService,
) *BillingService {
	return &BillingService{
		logger:        logger,
		billingRepo:   billingRepo,
		tenantService: tenantService,
	}
}

// CreateSubscription creates a new billing subscription for a tenant on the
// specified plan.
func (s *BillingService) CreateSubscription(ctx context.Context, tenantID uuid.UUID, planSlug string) (*domain.BillingSubscription, error) {
	// Verify the plan exists
	plans := domain.DefaultPlans()
	var found bool
	for _, p := range plans {
		if p.Slug == planSlug {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown plan slug: %s", planSlug)
	}

	now := time.Now()
	sub := &domain.BillingSubscription{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		PlanSlug:           planSlug,
		Status:             domain.BillingSubscriptionStatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.billingRepo.SaveSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	s.logger.Info("created subscription",
		zap.String("tenantId", tenantID.String()),
		zap.String("plan", planSlug),
	)

	return sub, nil
}

// GetSubscription retrieves the billing subscription for a tenant.
func (s *BillingService) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*domain.BillingSubscription, error) {
	sub, err := s.billingRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// UpgradePlan upgrades a tenant's subscription to a new plan.
func (s *BillingService) UpgradePlan(ctx context.Context, tenantID uuid.UUID, input domain.PlanUpgradeInput) (*domain.BillingSubscription, error) {
	sub, err := s.billingRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription for upgrade: %w", err)
	}

	// Verify the new plan exists
	plans := domain.DefaultPlans()
	var found bool
	for _, p := range plans {
		if p.Slug == input.PlanSlug {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown plan slug: %s", input.PlanSlug)
	}

	sub.PlanSlug = input.PlanSlug
	sub.UpdatedAt = time.Now()

	if err := s.billingRepo.SaveSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to save upgraded subscription: %w", err)
	}

	s.logger.Info("upgraded subscription plan",
		zap.String("tenantId", tenantID.String()),
		zap.String("newPlan", input.PlanSlug),
	)

	return sub, nil
}

// CancelSubscription marks a tenant's subscription for cancellation at the
// end of the current billing period.
func (s *BillingService) CancelSubscription(ctx context.Context, tenantID uuid.UUID) error {
	sub, err := s.billingRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get subscription for cancellation: %w", err)
	}

	sub.CancelAtPeriodEnd = true
	sub.UpdatedAt = time.Now()

	if err := s.billingRepo.SaveSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to save cancelled subscription: %w", err)
	}

	s.logger.Info("cancelled subscription",
		zap.String("tenantId", tenantID.String()),
	)

	return nil
}

// GetInvoices retrieves the most recent invoices for a tenant.
func (s *BillingService) GetInvoices(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.BillingInvoice, error) {
	invoices, err := s.billingRepo.ListInvoices(ctx, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

// GetPlans returns all available billing plans.
func (s *BillingService) GetPlans() []domain.BillingPlan {
	return domain.DefaultPlans()
}

// HandleWebhook processes incoming Stripe webhook events such as payment
// completions and subscription state changes.
func (s *BillingService) HandleWebhook(ctx context.Context, eventType string, payload map[string]any) error {
	s.logger.Info("processing billing webhook",
		zap.String("eventType", eventType),
	)

	switch eventType {
	case "invoice.paid":
		return s.handleInvoicePaid(ctx, payload)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, payload)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, payload)
	default:
		s.logger.Debug("ignoring unhandled webhook event",
			zap.String("eventType", eventType),
		)
		return nil
	}
}

func (s *BillingService) handleInvoicePaid(ctx context.Context, payload map[string]any) error {
	tenantIDStr, _ := payload["tenant_id"].(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID from webhook: %w", err)
	}

	amountCents, _ := payload["amount_cents"].(float64)
	invoiceID, _ := payload["stripe_invoice_id"].(string)

	now := time.Now()
	invoice := &domain.BillingInvoice{
		ID:              uuid.New(),
		TenantID:        tenantID,
		StripeInvoiceID: invoiceID,
		AmountCents:     int(amountCents),
		Status:          "paid",
		PeriodStart:     now.AddDate(0, -1, 0),
		PeriodEnd:       now,
		PaidAt:          &now,
		CreatedAt:       now,
	}

	if err := s.billingRepo.SaveInvoice(ctx, invoice); err != nil {
		return fmt.Errorf("failed to save invoice from webhook: %w", err)
	}

	return nil
}

func (s *BillingService) handleSubscriptionUpdated(ctx context.Context, payload map[string]any) error {
	tenantIDStr, _ := payload["tenant_id"].(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID from webhook: %w", err)
	}

	sub, err := s.billingRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get subscription from webhook: %w", err)
	}

	if statusStr, ok := payload["status"].(string); ok {
		sub.Status = domain.BillingSubscriptionStatus(statusStr)
	}
	sub.UpdatedAt = time.Now()

	if err := s.billingRepo.SaveSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription from webhook: %w", err)
	}

	return nil
}

func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, payload map[string]any) error {
	tenantIDStr, _ := payload["tenant_id"].(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID from webhook: %w", err)
	}

	sub, err := s.billingRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get subscription from webhook: %w", err)
	}

	sub.Status = domain.BillingSubscriptionStatusCanceled
	sub.UpdatedAt = time.Now()

	if err := s.billingRepo.SaveSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to cancel subscription from webhook: %w", err)
	}

	s.logger.Info("subscription deleted via webhook",
		zap.String("tenantId", tenantID.String()),
	)

	return nil
}
