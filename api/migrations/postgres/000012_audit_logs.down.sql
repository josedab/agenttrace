-- Drop audit log tables
DROP FUNCTION IF EXISTS apply_audit_retention_policy(UUID);
DROP FUNCTION IF EXISTS create_audit_log_partition(DATE);
DROP TABLE IF EXISTS audit_action_definitions;
DROP TABLE IF EXISTS audit_export_jobs;
DROP TRIGGER IF EXISTS update_audit_retention_policies_updated_at ON audit_retention_policies;
DROP TABLE IF EXISTS audit_retention_policies;

-- Drop partitioned audit_logs table (also drops all partitions)
DROP TABLE IF EXISTS audit_logs;
