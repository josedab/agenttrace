-- Drop SSO tables in reverse order
DROP FUNCTION IF EXISTS cleanup_expired_sso_states();
DROP TRIGGER IF EXISTS update_sso_configurations_updated_at ON sso_configurations;
DROP TABLE IF EXISTS sso_identity_mappings;
DROP TABLE IF EXISTS sso_states;
DROP TABLE IF EXISTS sso_sessions;
DROP TABLE IF EXISTS sso_configurations;
