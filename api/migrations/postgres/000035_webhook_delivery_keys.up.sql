-- Delivery keys identify one logical webhook payload so an immediate retry of
-- the same content can be recognized instead of delivered twice.
ALTER TABLE webhook_deliveries
    ADD COLUMN IF NOT EXISTS delivery_key VARCHAR(64) NOT NULL DEFAULT '';

-- The original migration created only fixed monthly partitions through June
-- 2026. Keep audit persistence available for later dates without requiring a
-- release for every future partition.
CREATE TABLE IF NOT EXISTS webhook_deliveries_default
    PARTITION OF webhook_deliveries DEFAULT;

-- A separate unpartitioned claim table provides a real uniqueness boundary
-- before the external send. Claims expire so a crashed sender can be retried.
CREATE TABLE IF NOT EXISTS team_digest_delivery_claims (
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    delivery_key VARCHAR(64) NOT NULL,
    claimed_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (webhook_id, delivery_key)
);

CREATE INDEX IF NOT EXISTS idx_team_digest_delivery_claims_expiry
    ON team_digest_delivery_claims(expires_at);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_key
    ON webhook_deliveries(webhook_id, delivery_key, created_at DESC)
    WHERE delivery_key <> '';
