-- Rollback Migration 000027: V3 Features

DROP TABLE IF EXISTS plugin_executions;
DROP TABLE IF EXISTS plugins;
DROP TABLE IF EXISTS push_notifications;
DROP TABLE IF EXISTS mobile_devices;
DROP TABLE IF EXISTS data_deletion_requests;
DROP TABLE IF EXISTS pii_configs;
DROP TABLE IF EXISTS fleet_policies;
DROP TABLE IF EXISTS agent_blueprints;
DROP TABLE IF EXISTS embed_configs;
DROP TABLE IF EXISTS budget_approvals;
DROP TABLE IF EXISTS cost_predictions;
DROP TABLE IF EXISTS agent_versions;
DROP TABLE IF EXISTS rca_reports;
DROP TABLE IF EXISTS orchestration_messages;
DROP TABLE IF EXISTS orchestration_sessions;
