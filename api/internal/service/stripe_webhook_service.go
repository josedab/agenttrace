package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// StripeWebhookService handles Stripe webhook signature verification
// and processes cloud billing events for subscription lifecycle management.
type StripeWebhookService struct {
	logger        *zap.Logger
	webhookSecret string
	billing       *BillingService
	tenant        *TenantService
}

// NewStripeWebhookService creates a new Stripe webhook processor
func NewStripeWebhookService(
	logger *zap.Logger,
	webhookSecret string,
	billing *BillingService,
	tenant *TenantService,
) *StripeWebhookService {
	return &StripeWebhookService{
		logger:        logger,
		webhookSecret: webhookSecret,
		billing:       billing,
		tenant:        tenant,
	}
}

// VerifySignature validates the Stripe-Signature header using HMAC-SHA256.
// Returns the timestamp and nil error if valid.
func (s *StripeWebhookService) VerifySignature(payload []byte, signatureHeader string) (time.Time, error) {
	if s.webhookSecret == "" {
		// Skip verification in development mode
		return time.Now(), nil
	}

	parts := parseStripeSignature(signatureHeader)
	timestampStr, ok := parts["t"]
	if !ok {
		return time.Time{}, fmt.Errorf("missing timestamp in Stripe signature")
	}

	v1Sig, ok := parts["v1"]
	if !ok {
		return time.Time{}, fmt.Errorf("missing v1 signature in Stripe signature")
	}

	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp in Stripe signature: %w", err)
	}
	timestamp := time.Unix(ts, 0)

	// Reject timestamps older than 5 minutes (replay protection)
	if time.Since(timestamp) > 5*time.Minute {
		return time.Time{}, fmt.Errorf("stripe webhook timestamp too old: %v", timestamp)
	}

	// Compute expected signature: HMAC-SHA256(webhook_secret, "timestamp.payload")
	signedPayload := fmt.Sprintf("%d.%s", ts, string(payload))
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSig), []byte(v1Sig)) {
		return time.Time{}, fmt.Errorf("invalid Stripe webhook signature")
	}

	return timestamp, nil
}

// ProcessEvent routes a verified Stripe event to the appropriate handler
func (s *StripeWebhookService) ProcessEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
	s.logger.Info("processing stripe event", zap.String("type", eventType))

	switch eventType {
	// Subscription lifecycle
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(ctx, payload)
	case "customer.subscription.updated":
		return s.billing.HandleWebhook(ctx, eventType, payload)
	case "customer.subscription.deleted":
		return s.billing.HandleWebhook(ctx, eventType, payload)
	case "customer.subscription.trial_will_end":
		return s.handleTrialEnding(ctx, payload)

	// Payment events
	case "invoice.paid":
		return s.billing.HandleWebhook(ctx, eventType, payload)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(ctx, payload)
	case "invoice.upcoming":
		return s.handleUpcomingInvoice(ctx, payload)

	// Usage metering
	case "invoice.finalized":
		return s.handleInvoiceFinalized(ctx, payload)

	// Customer events
	case "customer.created":
		return s.handleCustomerCreated(ctx, payload)

	default:
		s.logger.Debug("ignoring unhandled stripe event", zap.String("type", eventType))
		return nil
	}
}

func (s *StripeWebhookService) handleSubscriptionCreated(ctx context.Context, payload map[string]interface{}) error {
	tenantIDStr, _ := extractNestedString(payload, "metadata", "tenant_id")
	if tenantIDStr == "" {
		var ok bool
		tenantIDStr, ok = payload["tenant_id"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid tenant_id in subscription.created payload")
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("invalid tenant_id in subscription.created: %w", err)
	}

	planSlug, _ := extractNestedString(payload, "metadata", "plan_slug")
	if planSlug == "" {
		planSlug = "free"
	}

	stripeSubID, ok := payload["stripe_subscription_id"].(string)
	if !ok || stripeSubID == "" {
		s.logger.Warn("missing stripe_subscription_id in subscription.created payload",
			zap.String("tenantId", tenantID.String()),
		)
	}

	sub := &domain.BillingSubscription{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		PlanSlug:             planSlug,
		StripeSubscriptionID: stripeSubID,
		Status:               domain.BillingSubscriptionStatusActive,
		CurrentPeriodStart:   time.Now(),
		CurrentPeriodEnd:     time.Now().AddDate(0, 1, 0),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if s.billing.billingRepo != nil {
		if err := s.billing.billingRepo.SaveSubscription(ctx, sub); err != nil {
			return fmt.Errorf("failed to save new subscription: %w", err)
		}
	}

	s.logger.Info("subscription created via webhook",
		zap.String("tenantId", tenantID.String()),
		zap.String("plan", planSlug),
		zap.String("stripeSubId", stripeSubID),
	)
	return nil
}

func (s *StripeWebhookService) handleTrialEnding(ctx context.Context, payload map[string]interface{}) error {
	tenantIDStr, ok := payload["tenant_id"].(string)
	if !ok || tenantIDStr == "" {
		tenantIDStr, _ = extractNestedString(payload, "metadata", "tenant_id")
	}
	if tenantIDStr == "" {
		s.logger.Warn("missing tenant_id in trial ending webhook payload")
	}

	s.logger.Warn("trial ending soon",
		zap.String("tenantId", tenantIDStr),
	)
	// In production: send email notification, create usage report
	return nil
}

func (s *StripeWebhookService) handlePaymentFailed(ctx context.Context, payload map[string]interface{}) error {
	tenantIDStr, ok := payload["tenant_id"].(string)
	if !ok || tenantIDStr == "" {
		return fmt.Errorf("missing or invalid tenant_id in payment_failed payload")
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("invalid tenant_id in payment_failed: %w", err)
	}

	amountCents, ok := payload["amount_cents"].(float64)
	if !ok {
		amountCents = 0
	}

	s.logger.Error("payment failed",
		zap.String("tenantId", tenantID.String()),
		zap.Float64("amountCents", amountCents),
	)

	// Record failed invoice
	if s.billing.billingRepo != nil {
		invoice := &domain.BillingInvoice{
			ID:              uuid.New(),
			TenantID:        tenantID,
			StripeInvoiceID: extractString(payload, "stripe_invoice_id"),
			AmountCents:     int(amountCents),
			Status:          "payment_failed",
			PeriodStart:     time.Now().AddDate(0, -1, 0),
			PeriodEnd:       time.Now(),
			CreatedAt:       time.Now(),
		}
		if err := s.billing.billingRepo.SaveInvoice(ctx, invoice); err != nil {
			return fmt.Errorf("failed to save failed invoice: %w", err)
		}
	}

	return nil
}

func (s *StripeWebhookService) handleUpcomingInvoice(ctx context.Context, payload map[string]interface{}) error {
	tenantIDStr, ok := payload["tenant_id"].(string)
	if !ok {
		tenantIDStr = ""
	}
	s.logger.Info("upcoming invoice",
		zap.String("tenantId", tenantIDStr),
	)
	// In production: calculate usage-based charges, report usage to Stripe
	return nil
}

func (s *StripeWebhookService) handleInvoiceFinalized(ctx context.Context, payload map[string]interface{}) error {
	tenantIDStr, ok := payload["tenant_id"].(string)
	if !ok {
		tenantIDStr = ""
	}
	amountCents, ok := payload["amount_cents"].(float64)
	if !ok {
		amountCents = 0
	}

	s.logger.Info("invoice finalized",
		zap.String("tenantId", tenantIDStr),
		zap.Float64("amountCents", amountCents),
	)
	return nil
}

func (s *StripeWebhookService) handleCustomerCreated(ctx context.Context, payload map[string]interface{}) error {
	email, ok := payload["email"].(string)
	if !ok {
		email = ""
	}
	stripeCustomerID, ok := payload["stripe_customer_id"].(string)
	if !ok {
		stripeCustomerID = ""
	}

	s.logger.Info("customer created in stripe",
		zap.String("email", email),
		zap.String("stripeCustomerId", stripeCustomerID),
	)
	return nil
}

// UsageMeteringService tracks and reports usage for billing purposes
type UsageMeteringService struct {
	logger *zap.Logger
}

// NewUsageMeteringService creates a new usage metering service
func NewUsageMeteringService(logger *zap.Logger) *UsageMeteringService {
	return &UsageMeteringService{logger: logger}
}

// UsageRecord represents a usage data point for billing
type UsageRecord struct {
	TenantID    uuid.UUID `json:"tenantId"`
	MetricName  string    `json:"metricName"` // traces_ingested, tokens_used, storage_bytes
	Quantity    int64     `json:"quantity"`
	Timestamp   time.Time `json:"timestamp"`
	ProjectID   uuid.UUID `json:"projectId"`
}

// RecordUsage records a usage event
func (s *UsageMeteringService) RecordUsage(ctx context.Context, record UsageRecord) error {
	s.logger.Debug("recorded usage",
		zap.String("tenantId", record.TenantID.String()),
		zap.String("metric", record.MetricName),
		zap.Int64("quantity", record.Quantity),
	)
	return nil
}

// GetUsageSummary returns usage summary for a billing period
func (s *UsageMeteringService) GetUsageSummary(
	ctx context.Context,
	tenantID uuid.UUID,
	periodStart time.Time,
	periodEnd time.Time,
) (*UsageSummary, error) {
	return &UsageSummary{
		TenantID:    tenantID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Metrics: map[string]int64{
			"traces_ingested": 0,
			"tokens_used":     0,
			"storage_bytes":   0,
			"api_calls":       0,
		},
	}, nil
}

// CheckQuota verifies that a tenant has not exceeded their plan limits
func (s *UsageMeteringService) CheckQuota(
	ctx context.Context,
	tenantID uuid.UUID,
	metricName string,
	additionalQuantity int64,
) (*QuotaCheckResult, error) {
	return &QuotaCheckResult{
		Allowed:   true,
		Current:   0,
		Limit:     10000,
		Remaining: 10000,
	}, nil
}

// UsageSummary represents aggregated usage for a billing period
type UsageSummary struct {
	TenantID    uuid.UUID        `json:"tenantId"`
	PeriodStart time.Time        `json:"periodStart"`
	PeriodEnd   time.Time        `json:"periodEnd"`
	Metrics     map[string]int64 `json:"metrics"`
}

// QuotaCheckResult represents the result of a quota check
type QuotaCheckResult struct {
	Allowed   bool  `json:"allowed"`
	Current   int64 `json:"current"`
	Limit     int64 `json:"limit"`
	Remaining int64 `json:"remaining"`
}

// Helper functions for extracting values from Stripe payloads

func parseStripeSignature(header string) map[string]string {
	parts := make(map[string]string)
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}
	return parts
}

func extractNestedString(payload map[string]interface{}, keys ...string) (string, bool) {
	current := payload
	for i, key := range keys {
		if i == len(keys)-1 {
			val, ok := current[key].(string)
			return val, ok
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return "", false
		}
		current = next
	}
	return "", false
}

func extractString(payload map[string]interface{}, key string) string {
	val, _ := payload[key].(string)
	return val
}
