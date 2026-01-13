-- Webhooks table
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL DEFAULT 'generic',
    name VARCHAR(100) NOT NULL,
    url VARCHAR(2000) NOT NULL,
    secret VARCHAR(255),
    events JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    headers JSONB DEFAULT '{}'::jsonb,
    cost_threshold DECIMAL(10, 4),
    latency_threshold BIGINT,
    score_threshold DECIMAL(5, 4),
    rate_limit_per_hour INTEGER,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_webhook_type CHECK (type IN ('slack', 'discord', 'msteams', 'pagerduty', 'generic')),
    CONSTRAINT unique_webhook_name_per_project UNIQUE (project_id, name)
);

-- Indexes for webhooks
CREATE INDEX idx_webhooks_project ON webhooks(project_id);
CREATE INDEX idx_webhooks_project_enabled ON webhooks(project_id, is_enabled) WHERE is_enabled = true;
CREATE INDEX idx_webhooks_type ON webhooks(type);

-- Trigger for updated_at
CREATE TRIGGER update_webhooks_updated_at
    BEFORE UPDATE ON webhooks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Webhook Deliveries table (partitioned by time for performance)
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload TEXT NOT NULL,
    status_code INTEGER,
    response TEXT,
    success BOOLEAN NOT NULL DEFAULT false,
    error TEXT,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions for webhook_deliveries
CREATE TABLE webhook_deliveries_y2026m01 PARTITION OF webhook_deliveries
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE webhook_deliveries_y2026m02 PARTITION OF webhook_deliveries
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE webhook_deliveries_y2026m03 PARTITION OF webhook_deliveries
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE webhook_deliveries_y2026m04 PARTITION OF webhook_deliveries
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE webhook_deliveries_y2026m05 PARTITION OF webhook_deliveries
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE webhook_deliveries_y2026m06 PARTITION OF webhook_deliveries
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

-- Indexes for webhook_deliveries
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, created_at DESC);
CREATE INDEX idx_webhook_deliveries_success ON webhook_deliveries(webhook_id, success, created_at DESC);
CREATE INDEX idx_webhook_deliveries_event ON webhook_deliveries(event_type, created_at DESC);

-- Webhook Rate Limiting table (for tracking delivery counts)
CREATE TABLE IF NOT EXISTS webhook_rate_limits (
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    hour_bucket TIMESTAMP WITH TIME ZONE NOT NULL,
    delivery_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (webhook_id, hour_bucket)
);

-- Index for cleanup of old rate limit records
CREATE INDEX idx_webhook_rate_limits_bucket ON webhook_rate_limits(hour_bucket);

-- Trigger for updated_at
CREATE TRIGGER update_webhook_rate_limits_updated_at
    BEFORE UPDATE ON webhook_rate_limits
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to check if webhook can send (rate limit check)
CREATE OR REPLACE FUNCTION check_webhook_rate_limit(
    p_webhook_id UUID,
    p_limit_per_hour INTEGER
) RETURNS BOOLEAN AS $$
DECLARE
    v_current_bucket TIMESTAMP WITH TIME ZONE;
    v_current_count INTEGER;
BEGIN
    -- Calculate current hour bucket
    v_current_bucket := date_trunc('hour', NOW());

    -- Get or create rate limit record
    INSERT INTO webhook_rate_limits (webhook_id, hour_bucket, delivery_count)
    VALUES (p_webhook_id, v_current_bucket, 0)
    ON CONFLICT (webhook_id, hour_bucket) DO NOTHING;

    -- Get current count
    SELECT delivery_count INTO v_current_count
    FROM webhook_rate_limits
    WHERE webhook_id = p_webhook_id AND hour_bucket = v_current_bucket;

    -- Check against limit (NULL limit means unlimited)
    IF p_limit_per_hour IS NULL THEN
        RETURN true;
    END IF;

    RETURN v_current_count < p_limit_per_hour;
END;
$$ LANGUAGE plpgsql;

-- Function to increment webhook delivery count
CREATE OR REPLACE FUNCTION increment_webhook_delivery_count(
    p_webhook_id UUID
) RETURNS VOID AS $$
DECLARE
    v_current_bucket TIMESTAMP WITH TIME ZONE;
BEGIN
    v_current_bucket := date_trunc('hour', NOW());

    INSERT INTO webhook_rate_limits (webhook_id, hour_bucket, delivery_count)
    VALUES (p_webhook_id, v_current_bucket, 1)
    ON CONFLICT (webhook_id, hour_bucket)
    DO UPDATE SET
        delivery_count = webhook_rate_limits.delivery_count + 1,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;
