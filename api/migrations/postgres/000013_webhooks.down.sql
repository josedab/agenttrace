-- Drop functions
DROP FUNCTION IF EXISTS increment_webhook_delivery_count(UUID);
DROP FUNCTION IF EXISTS check_webhook_rate_limit(UUID, INTEGER);

-- Drop rate limits table
DROP TABLE IF EXISTS webhook_rate_limits;

-- Drop webhook deliveries partitions and table
DROP TABLE IF EXISTS webhook_deliveries_y2026m06;
DROP TABLE IF EXISTS webhook_deliveries_y2026m05;
DROP TABLE IF EXISTS webhook_deliveries_y2026m04;
DROP TABLE IF EXISTS webhook_deliveries_y2026m03;
DROP TABLE IF EXISTS webhook_deliveries_y2026m02;
DROP TABLE IF EXISTS webhook_deliveries_y2026m01;
DROP TABLE IF EXISTS webhook_deliveries;

-- Drop webhooks table
DROP TABLE IF EXISTS webhooks;
