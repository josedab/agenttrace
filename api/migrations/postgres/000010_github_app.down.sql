-- Drop triggers
DROP TRIGGER IF EXISTS github_repositories_updated_at ON github_repositories;
DROP TRIGGER IF EXISTS github_installations_updated_at ON github_installations;
DROP FUNCTION IF EXISTS update_github_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_github_webhook_events_created;
DROP INDEX IF EXISTS idx_github_webhook_events_processed;
DROP INDEX IF EXISTS idx_github_webhook_events_type;
DROP INDEX IF EXISTS idx_github_webhook_events_installation;
DROP INDEX IF EXISTS idx_github_repositories_full_name;
DROP INDEX IF EXISTS idx_github_repositories_project;
DROP INDEX IF EXISTS idx_github_repositories_installation;
DROP INDEX IF EXISTS idx_github_installations_installation_id;
DROP INDEX IF EXISTS idx_github_installations_account;
DROP INDEX IF EXISTS idx_github_installations_org;

-- Drop tables
DROP TABLE IF EXISTS github_webhook_events;
DROP TABLE IF EXISTS github_repositories;
DROP TABLE IF EXISTS github_installations;
