-- Reverse migration 000026: Drop Next-Gen Features V2 tables
-- Drop in reverse dependency order

DROP TABLE IF EXISTS api_key_scopes;
DROP TABLE IF EXISTS sso_configs;
DROP TABLE IF EXISTS role_assignments;
DROP TABLE IF EXISTS compliance_reports;
DROP TABLE IF EXISTS training_exports;
DROP TABLE IF EXISTS training_datasets;
DROP TABLE IF EXISTS package_ratings;
DROP TABLE IF EXISTS marketplace_packages;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhook_rules;
DROP TABLE IF EXISTS sandbox_policies;
DROP TABLE IF EXISTS sandbox_reviews;
DROP TABLE IF EXISTS prompt_variants;
DROP TABLE IF EXISTS prompt_experiments;
