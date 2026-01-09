DROP TABLE IF EXISTS evaluation_jobs;
DROP TABLE IF EXISTS evaluator_templates;
DROP TRIGGER IF EXISTS update_evaluators_updated_at ON evaluators;
DROP TABLE IF EXISTS evaluators;
DROP TYPE IF EXISTS score_data_type;
DROP TYPE IF EXISTS evaluator_type;
