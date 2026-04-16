DROP INDEX IF EXISTS idx_webhook_deliveries_key;
DROP INDEX IF EXISTS idx_team_digest_delivery_claims_expiry;
DROP TABLE IF EXISTS team_digest_delivery_claims;

DROP TABLE IF EXISTS webhook_deliveries_default;

ALTER TABLE webhook_deliveries
    DROP COLUMN IF EXISTS delivery_key;
