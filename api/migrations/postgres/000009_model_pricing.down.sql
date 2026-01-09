DROP TABLE IF EXISTS export_jobs;
DROP TRIGGER IF EXISTS update_export_configs_updated_at ON export_configs;
DROP TABLE IF EXISTS export_configs;
DROP TRIGGER IF EXISTS update_project_model_pricing_updated_at ON project_model_pricing;
DROP TABLE IF EXISTS project_model_pricing;
DROP TRIGGER IF EXISTS update_model_pricing_updated_at ON model_pricing;
DROP TABLE IF EXISTS model_pricing;
