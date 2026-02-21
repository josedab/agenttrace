-- Billing subscriptions table
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_slug VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'trialing',
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_billing_subscription_status CHECK (status IN ('active', 'past_due', 'canceled', 'trialing'))
);

-- Indexes for billing_subscriptions
CREATE INDEX idx_billing_subscriptions_tenant ON billing_subscriptions(tenant_id);
CREATE UNIQUE INDEX idx_billing_subscriptions_stripe_customer ON billing_subscriptions(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE UNIQUE INDEX idx_billing_subscriptions_stripe_subscription ON billing_subscriptions(stripe_subscription_id) WHERE stripe_subscription_id IS NOT NULL;

-- Trigger for updated_at
CREATE TRIGGER update_billing_subscriptions_updated_at
    BEFORE UPDATE ON billing_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Billing invoices table
CREATE TABLE IF NOT EXISTS billing_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(255),
    amount_cents INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    invoice_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_billing_invoice_status CHECK (status IN ('draft', 'open', 'paid', 'void', 'uncollectible'))
);

-- Indexes for billing_invoices
CREATE INDEX idx_billing_invoices_tenant ON billing_invoices(tenant_id);
CREATE UNIQUE INDEX idx_billing_invoices_stripe_invoice ON billing_invoices(stripe_invoice_id) WHERE stripe_invoice_id IS NOT NULL;
