package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStripeWebhookService_VerifySignature(t *testing.T) {
	logger := zap.NewNop()
	secret := "whsec_test_secret"
	svc := NewStripeWebhookService(logger, secret, nil, nil)

	payload := []byte(`{"type":"invoice.paid","tenant_id":"abc"}`)
	ts := time.Now().Unix()

	// Compute valid signature
	signedPayload := fmt.Sprintf("%d.%s", ts, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))

	header := fmt.Sprintf("t=%d,v1=%s", ts, sig)

	timestamp, err := svc.VerifySignature(payload, header)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), timestamp, 10*time.Second)
}

func TestStripeWebhookService_VerifySignatureInvalid(t *testing.T) {
	logger := zap.NewNop()
	secret := "whsec_test_secret"
	svc := NewStripeWebhookService(logger, secret, nil, nil)

	payload := []byte(`{"type":"test"}`)
	header := fmt.Sprintf("t=%d,v1=invalidsig", time.Now().Unix())

	_, err := svc.VerifySignature(payload, header)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Stripe webhook signature")
}

func TestStripeWebhookService_VerifySignatureExpired(t *testing.T) {
	logger := zap.NewNop()
	secret := "whsec_test_secret"
	svc := NewStripeWebhookService(logger, secret, nil, nil)

	payload := []byte(`{"type":"test"}`)
	oldTs := time.Now().Add(-10 * time.Minute).Unix()

	signedPayload := fmt.Sprintf("%d.%s", oldTs, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))

	header := fmt.Sprintf("t=%d,v1=%s", oldTs, sig)

	_, err := svc.VerifySignature(payload, header)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too old")
}

func TestStripeWebhookService_VerifySignatureDevMode(t *testing.T) {
	logger := zap.NewNop()
	svc := NewStripeWebhookService(logger, "", nil, nil) // empty secret = dev mode

	_, err := svc.VerifySignature([]byte("anything"), "")
	assert.NoError(t, err)
}

func TestStripeWebhookService_ProcessEventUnknown(t *testing.T) {
	logger := zap.NewNop()
	svc := NewStripeWebhookService(logger, "", nil, nil)

	err := svc.ProcessEvent(context.Background(), "unknown.event", map[string]interface{}{})
	assert.NoError(t, err) // Unknown events are silently ignored
}

func TestStripeWebhookService_HandleTrialEnding(t *testing.T) {
	logger := zap.NewNop()
	svc := NewStripeWebhookService(logger, "", nil, nil)

	err := svc.ProcessEvent(context.Background(), "customer.subscription.trial_will_end", map[string]interface{}{
		"tenant_id": uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestStripeWebhookService_HandleCustomerCreated(t *testing.T) {
	logger := zap.NewNop()
	svc := NewStripeWebhookService(logger, "", nil, nil)

	err := svc.ProcessEvent(context.Background(), "customer.created", map[string]interface{}{
		"email":              "test@example.com",
		"stripe_customer_id": "cus_test123",
	})
	assert.NoError(t, err)
}

func TestStripeWebhookService_HandlePaymentFailed(t *testing.T) {
	logger := zap.NewNop()

	t.Run("valid payload", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "invoice.payment_failed", map[string]interface{}{
			"tenant_id":    uuid.New().String(),
			"amount_cents": float64(9900),
		})
		assert.NoError(t, err)
	})

	t.Run("missing tenant_id returns error", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "invoice.payment_failed", map[string]interface{}{
			"amount_cents": float64(9900),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing or invalid tenant_id")
	})

	t.Run("invalid tenant_id returns error", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "invoice.payment_failed", map[string]interface{}{
			"tenant_id": "not-a-uuid",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tenant_id")
	})

	t.Run("wrong type for amount_cents uses zero", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "invoice.payment_failed", map[string]interface{}{
			"tenant_id":    uuid.New().String(),
			"amount_cents": "not-a-number",
		})
		assert.NoError(t, err)
	})
}

func TestStripeWebhookService_HandleUpcomingInvoice(t *testing.T) {
	logger := zap.NewNop()
	svc := NewStripeWebhookService(logger, "", nil, nil)

	t.Run("valid payload", func(t *testing.T) {
		err := svc.ProcessEvent(context.Background(), "invoice.upcoming", map[string]interface{}{
			"tenant_id": uuid.New().String(),
		})
		assert.NoError(t, err)
	})

	t.Run("missing tenant_id does not panic", func(t *testing.T) {
		err := svc.ProcessEvent(context.Background(), "invoice.upcoming", map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("wrong type for tenant_id does not panic", func(t *testing.T) {
		err := svc.ProcessEvent(context.Background(), "invoice.upcoming", map[string]interface{}{
			"tenant_id": 12345,
		})
		assert.NoError(t, err)
	})
}

func TestStripeWebhookService_HandleInvoiceFinalized(t *testing.T) {
	logger := zap.NewNop()
	svc := NewStripeWebhookService(logger, "", nil, nil)

	t.Run("valid payload", func(t *testing.T) {
		err := svc.ProcessEvent(context.Background(), "invoice.finalized", map[string]interface{}{
			"tenant_id":    uuid.New().String(),
			"amount_cents": float64(4900),
		})
		assert.NoError(t, err)
	})

	t.Run("malformed payload does not panic", func(t *testing.T) {
		err := svc.ProcessEvent(context.Background(), "invoice.finalized", map[string]interface{}{
			"tenant_id":    12345,
			"amount_cents": "bad",
		})
		assert.NoError(t, err)
	})
}

func TestStripeWebhookService_HandleSubscriptionCreated(t *testing.T) {
	logger := zap.NewNop()

	t.Run("missing tenant_id returns error", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "customer.subscription.created", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing or invalid tenant_id")
	})

	t.Run("invalid tenant_id returns error", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "customer.subscription.created", map[string]interface{}{
			"tenant_id": "not-a-uuid",
		})
		assert.Error(t, err)
	})

	t.Run("valid payload with metadata", func(t *testing.T) {
		svc := NewStripeWebhookService(logger, "", &BillingService{}, nil)
		err := svc.ProcessEvent(context.Background(), "customer.subscription.created", map[string]interface{}{
			"metadata": map[string]interface{}{
				"tenant_id": uuid.New().String(),
				"plan_slug": "pro",
			},
			"stripe_subscription_id": "sub_test123",
		})
		assert.NoError(t, err)
	})
}

func TestUsageMeteringService_RecordUsage(t *testing.T) {
	logger := zap.NewNop()
	svc := NewUsageMeteringService(logger)

	err := svc.RecordUsage(context.Background(), UsageRecord{
		TenantID:   uuid.New(),
		MetricName: "traces_ingested",
		Quantity:   100,
		Timestamp:  time.Now(),
		ProjectID:  uuid.New(),
	})
	assert.NoError(t, err)
}

func TestUsageMeteringService_GetUsageSummary(t *testing.T) {
	logger := zap.NewNop()
	svc := NewUsageMeteringService(logger)

	summary, err := svc.GetUsageSummary(context.Background(), uuid.New(),
		time.Now().AddDate(0, -1, 0), time.Now())
	require.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Contains(t, summary.Metrics, "traces_ingested")
	assert.Contains(t, summary.Metrics, "tokens_used")
}

func TestUsageMeteringService_CheckQuota(t *testing.T) {
	logger := zap.NewNop()
	svc := NewUsageMeteringService(logger)

	result, err := svc.CheckQuota(context.Background(), uuid.New(), "traces_ingested", 100)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Greater(t, result.Remaining, int64(0))
}

func TestParseStripeSignature(t *testing.T) {
	header := "t=1234567890,v1=abc123def456"
	parts := parseStripeSignature(header)
	assert.Equal(t, "1234567890", parts["t"])
	assert.Equal(t, "abc123def456", parts["v1"])
}

func TestExtractNestedString(t *testing.T) {
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"tenant_id": "abc-123",
		},
	}

	val, ok := extractNestedString(payload, "metadata", "tenant_id")
	assert.True(t, ok)
	assert.Equal(t, "abc-123", val)

	_, ok = extractNestedString(payload, "metadata", "nonexistent")
	assert.False(t, ok)
}
