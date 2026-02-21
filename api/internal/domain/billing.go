package domain

import (
	"time"

	"github.com/google/uuid"
)

// BillingSubscriptionStatus represents the status of a billing subscription
type BillingSubscriptionStatus string

const (
	BillingSubscriptionStatusActive   BillingSubscriptionStatus = "active"
	BillingSubscriptionStatusPastDue  BillingSubscriptionStatus = "past_due"
	BillingSubscriptionStatusCanceled BillingSubscriptionStatus = "canceled"
	BillingSubscriptionStatusTrialing BillingSubscriptionStatus = "trialing"
)

// IsValid checks if the billing subscription status is valid
func (s BillingSubscriptionStatus) IsValid() bool {
	switch s {
	case BillingSubscriptionStatusActive, BillingSubscriptionStatusPastDue, BillingSubscriptionStatusCanceled, BillingSubscriptionStatusTrialing:
		return true
	}
	return false
}

// BillingPlan represents a subscription plan with its pricing and limits
type BillingPlan struct {
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	PriceMonthly     int      `json:"priceMonthly"`     // in cents
	TracesIncluded   int      `json:"tracesIncluded"`
	ProjectsIncluded int      `json:"projectsIncluded"`
	UsersIncluded    int      `json:"usersIncluded"`
	RetentionDays    int      `json:"retentionDays"`
	Features         []string `json:"features"`
}

// BillingSubscription represents a tenant's active subscription
type BillingSubscription struct {
	ID                   uuid.UUID                 `json:"id"`
	TenantID             uuid.UUID                 `json:"tenantId"`
	PlanSlug             string                    `json:"planSlug"`
	Status               BillingSubscriptionStatus `json:"status"`
	StripeCustomerID     string                    `json:"stripeCustomerId"`
	StripeSubscriptionID string                    `json:"stripeSubscriptionId"`
	CurrentPeriodStart   time.Time                 `json:"currentPeriodStart"`
	CurrentPeriodEnd     time.Time                 `json:"currentPeriodEnd"`
	CancelAtPeriodEnd    bool                      `json:"cancelAtPeriodEnd"`
	CreatedAt            time.Time                 `json:"createdAt"`
	UpdatedAt            time.Time                 `json:"updatedAt"`
}

// BillingInvoice represents an invoice for a billing period
type BillingInvoice struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenantId"`
	StripeInvoiceID string     `json:"stripeInvoiceId"`
	AmountCents     int        `json:"amountCents"`
	Status          string     `json:"status"`
	PeriodStart     time.Time  `json:"periodStart"`
	PeriodEnd       time.Time  `json:"periodEnd"`
	PaidAt          *time.Time `json:"paidAt,omitempty"`
	InvoiceURL      string     `json:"invoiceUrl"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// SignupInput represents input for new user registration
type SignupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"orgName"`
	Plan     string `json:"plan"`
}

// PlanUpgradeInput represents input for upgrading a billing plan
type PlanUpgradeInput struct {
	PlanSlug        string `json:"planSlug"`
	PaymentMethodID string `json:"paymentMethodId"`
}

// DefaultPlans returns the predefined billing plan configurations
func DefaultPlans() []BillingPlan {
	return []BillingPlan{
		{
			Slug:             "free",
			Name:             "Free",
			PriceMonthly:     0,
			TracesIncluded:   10000,
			ProjectsIncluded: 2,
			UsersIncluded:    1,
			RetentionDays:    7,
			Features:         []string{"basic_tracing", "dashboard"},
		},
		{
			Slug:             "pro",
			Name:             "Pro",
			PriceMonthly:     4900, // $49.00
			TracesIncluded:   500000,
			ProjectsIncluded: 20,
			UsersIncluded:    10,
			RetentionDays:    90,
			Features:         []string{"basic_tracing", "dashboard", "alerts", "evaluations", "integrations", "export"},
		},
		{
			Slug:             "enterprise",
			Name:             "Enterprise",
			PriceMonthly:     29900, // $299.00
			TracesIncluded:   5000000,
			ProjectsIncluded: 100,
			UsersIncluded:    50,
			RetentionDays:    365,
			Features:         []string{"basic_tracing", "dashboard", "alerts", "evaluations", "integrations", "export", "sso", "audit_logs", "sla", "dedicated_support"},
		},
	}
}
