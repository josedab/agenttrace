DROP TABLE IF EXISTS compiled_prompts;
DROP TABLE IF EXISTS prompt_versions;
DROP TRIGGER IF EXISTS update_prompts_updated_at ON prompts;
DROP TABLE IF EXISTS prompts;
DROP TYPE IF EXISTS prompt_type;
