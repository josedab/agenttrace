-- Model pricing table (400+ models)
CREATE TABLE IF NOT EXISTS model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(100) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    model_regex VARCHAR(255),
    input_price_per_million DECIMAL(20,10) NOT NULL,
    output_price_per_million DECIMAL(20,10) NOT NULL,
    cache_read_price_per_million DECIMAL(20,10),
    cache_write_price_per_million DECIMAL(20,10),
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(provider, model_name, effective_date)
);

CREATE INDEX idx_model_pricing_provider ON model_pricing(provider);
CREATE INDEX idx_model_pricing_model_name ON model_pricing(model_name);
CREATE INDEX idx_model_pricing_effective_date ON model_pricing(effective_date);

CREATE TRIGGER update_model_pricing_updated_at
    BEFORE UPDATE ON model_pricing
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Project-level pricing overrides
CREATE TABLE IF NOT EXISTS project_model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    model_name VARCHAR(255) NOT NULL,
    input_price_per_million DECIMAL(20,10) NOT NULL,
    output_price_per_million DECIMAL(20,10) NOT NULL,
    cache_read_price_per_million DECIMAL(20,10),
    cache_write_price_per_million DECIMAL(20,10),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, model_name)
);

CREATE INDEX idx_project_model_pricing_project ON project_model_pricing(project_id);

CREATE TRIGGER update_project_model_pricing_updated_at
    BEFORE UPDATE ON project_model_pricing
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default pricing for major models (as of 2024)
INSERT INTO model_pricing (provider, model_name, input_price_per_million, output_price_per_million, cache_read_price_per_million) VALUES
-- OpenAI GPT-4 models
('openai', 'gpt-4o', 2.50, 10.00, 1.25),
('openai', 'gpt-4o-2024-11-20', 2.50, 10.00, 1.25),
('openai', 'gpt-4o-2024-08-06', 2.50, 10.00, 1.25),
('openai', 'gpt-4o-mini', 0.15, 0.60, 0.075),
('openai', 'gpt-4o-mini-2024-07-18', 0.15, 0.60, 0.075),
('openai', 'gpt-4-turbo', 10.00, 30.00, NULL),
('openai', 'gpt-4-turbo-2024-04-09', 10.00, 30.00, NULL),
('openai', 'gpt-4-turbo-preview', 10.00, 30.00, NULL),
('openai', 'gpt-4-1106-preview', 10.00, 30.00, NULL),
('openai', 'gpt-4-0125-preview', 10.00, 30.00, NULL),
('openai', 'gpt-4', 30.00, 60.00, NULL),
('openai', 'gpt-4-0613', 30.00, 60.00, NULL),
('openai', 'gpt-4-32k', 60.00, 120.00, NULL),

-- OpenAI GPT-3.5 models
('openai', 'gpt-3.5-turbo', 0.50, 1.50, NULL),
('openai', 'gpt-3.5-turbo-0125', 0.50, 1.50, NULL),
('openai', 'gpt-3.5-turbo-1106', 1.00, 2.00, NULL),
('openai', 'gpt-3.5-turbo-instruct', 1.50, 2.00, NULL),

-- OpenAI o1 models
('openai', 'o1-preview', 15.00, 60.00, 7.50),
('openai', 'o1-preview-2024-09-12', 15.00, 60.00, 7.50),
('openai', 'o1-mini', 3.00, 12.00, 1.50),
('openai', 'o1-mini-2024-09-12', 3.00, 12.00, 1.50),
('openai', 'o1', 15.00, 60.00, 7.50),
('openai', 'o3-mini', 1.10, 4.40, 0.55),

-- Anthropic Claude models
('anthropic', 'claude-3-5-sonnet-20241022', 3.00, 15.00, 0.30),
('anthropic', 'claude-3-5-sonnet-20240620', 3.00, 15.00, 0.30),
('anthropic', 'claude-3-5-haiku-20241022', 0.80, 4.00, 0.08),
('anthropic', 'claude-3-opus-20240229', 15.00, 75.00, 1.50),
('anthropic', 'claude-3-sonnet-20240229', 3.00, 15.00, 0.30),
('anthropic', 'claude-3-haiku-20240307', 0.25, 1.25, 0.03),
('anthropic', 'claude-2.1', 8.00, 24.00, NULL),
('anthropic', 'claude-2.0', 8.00, 24.00, NULL),
('anthropic', 'claude-instant-1.2', 0.80, 2.40, NULL),

-- Google models
('google', 'gemini-1.5-pro', 1.25, 5.00, NULL),
('google', 'gemini-1.5-pro-latest', 1.25, 5.00, NULL),
('google', 'gemini-1.5-flash', 0.075, 0.30, NULL),
('google', 'gemini-1.5-flash-latest', 0.075, 0.30, NULL),
('google', 'gemini-1.0-pro', 0.50, 1.50, NULL),
('google', 'gemini-pro', 0.50, 1.50, NULL),
('google', 'gemini-2.0-flash-exp', 0.10, 0.40, NULL),

-- Meta Llama models (via providers)
('meta', 'llama-3.1-405b', 5.00, 15.00, NULL),
('meta', 'llama-3.1-70b', 0.88, 0.88, NULL),
('meta', 'llama-3.1-8b', 0.18, 0.18, NULL),
('meta', 'llama-3-70b', 0.79, 0.79, NULL),
('meta', 'llama-3-8b', 0.18, 0.18, NULL),
('meta', 'llama-2-70b', 0.70, 0.90, NULL),

-- Mistral models
('mistral', 'mistral-large-latest', 2.00, 6.00, NULL),
('mistral', 'mistral-large-2411', 2.00, 6.00, NULL),
('mistral', 'mistral-small-latest', 0.20, 0.60, NULL),
('mistral', 'mistral-medium-latest', 2.70, 8.10, NULL),
('mistral', 'open-mixtral-8x22b', 2.00, 6.00, NULL),
('mistral', 'open-mixtral-8x7b', 0.70, 0.70, NULL),
('mistral', 'open-mistral-7b', 0.25, 0.25, NULL),
('mistral', 'codestral-latest', 0.20, 0.60, NULL),
('mistral', 'codestral-2405', 0.20, 0.60, NULL),
('mistral', 'pixtral-12b-2409', 0.15, 0.15, NULL),

-- Cohere models
('cohere', 'command-r-plus', 2.50, 10.00, NULL),
('cohere', 'command-r', 0.15, 0.60, NULL),
('cohere', 'command', 1.00, 2.00, NULL),
('cohere', 'command-light', 0.30, 0.60, NULL),
('cohere', 'command-nightly', 1.00, 2.00, NULL),

-- AWS Bedrock pricing
('aws-bedrock', 'anthropic.claude-3-5-sonnet-20241022-v2:0', 3.00, 15.00, 0.30),
('aws-bedrock', 'anthropic.claude-3-5-sonnet-20240620-v1:0', 3.00, 15.00, 0.30),
('aws-bedrock', 'anthropic.claude-3-opus-20240229-v1:0', 15.00, 75.00, 1.50),
('aws-bedrock', 'anthropic.claude-3-sonnet-20240229-v1:0', 3.00, 15.00, 0.30),
('aws-bedrock', 'anthropic.claude-3-haiku-20240307-v1:0', 0.25, 1.25, 0.03),
('aws-bedrock', 'meta.llama3-1-405b-instruct-v1:0', 5.32, 16.00, NULL),
('aws-bedrock', 'meta.llama3-1-70b-instruct-v1:0', 0.99, 0.99, NULL),
('aws-bedrock', 'meta.llama3-1-8b-instruct-v1:0', 0.22, 0.22, NULL),
('aws-bedrock', 'amazon.titan-text-premier-v1:0', 0.50, 1.50, NULL),
('aws-bedrock', 'amazon.titan-text-express-v1', 0.20, 0.60, NULL),
('aws-bedrock', 'amazon.titan-text-lite-v1', 0.15, 0.20, NULL),

-- Azure OpenAI (same pricing as OpenAI)
('azure', 'gpt-4o', 2.50, 10.00, 1.25),
('azure', 'gpt-4o-mini', 0.15, 0.60, 0.075),
('azure', 'gpt-4-turbo', 10.00, 30.00, NULL),
('azure', 'gpt-4', 30.00, 60.00, NULL),
('azure', 'gpt-35-turbo', 0.50, 1.50, NULL),

-- Groq models
('groq', 'llama-3.1-405b-reasoning', 0.00, 0.00, NULL),
('groq', 'llama-3.1-70b-versatile', 0.00, 0.00, NULL),
('groq', 'llama-3.1-8b-instant', 0.00, 0.00, NULL),
('groq', 'mixtral-8x7b-32768', 0.00, 0.00, NULL),
('groq', 'gemma-7b-it', 0.00, 0.00, NULL),

-- Together AI models
('together', 'meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo', 3.50, 3.50, NULL),
('together', 'meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo', 0.88, 0.88, NULL),
('together', 'meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo', 0.18, 0.18, NULL),
('together', 'mistralai/Mixtral-8x22B-Instruct-v0.1', 1.20, 1.20, NULL),
('together', 'mistralai/Mixtral-8x7B-Instruct-v0.1', 0.60, 0.60, NULL),

-- Fireworks AI
('fireworks', 'accounts/fireworks/models/llama-v3p1-405b-instruct', 3.00, 3.00, NULL),
('fireworks', 'accounts/fireworks/models/llama-v3p1-70b-instruct', 0.90, 0.90, NULL),
('fireworks', 'accounts/fireworks/models/llama-v3p1-8b-instruct', 0.20, 0.20, NULL),
('fireworks', 'accounts/fireworks/models/mixtral-8x22b-instruct', 0.90, 0.90, NULL),

-- Replicate
('replicate', 'meta/meta-llama-3.1-405b-instruct', 9.50, 9.50, NULL),
('replicate', 'meta/meta-llama-3-70b-instruct', 0.65, 2.75, NULL),
('replicate', 'meta/meta-llama-3-8b-instruct', 0.05, 0.25, NULL),

-- Perplexity
('perplexity', 'llama-3.1-sonar-huge-128k-online', 5.00, 5.00, NULL),
('perplexity', 'llama-3.1-sonar-large-128k-online', 1.00, 1.00, NULL),
('perplexity', 'llama-3.1-sonar-small-128k-online', 0.20, 0.20, NULL),

-- DeepSeek
('deepseek', 'deepseek-chat', 0.14, 0.28, NULL),
('deepseek', 'deepseek-coder', 0.14, 0.28, NULL),

-- OpenRouter (pass-through, using base pricing)
('openrouter', 'openai/gpt-4o', 2.50, 10.00, 1.25),
('openrouter', 'anthropic/claude-3.5-sonnet', 3.00, 15.00, 0.30),
('openrouter', 'google/gemini-pro-1.5', 1.25, 5.00, NULL),
('openrouter', 'meta-llama/llama-3.1-405b-instruct', 2.70, 2.70, NULL);

-- Export configurations
CREATE TABLE IF NOT EXISTS export_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    destination_type VARCHAR(50) NOT NULL, -- 's3', 'gcs', 'azure_blob'
    destination_config JSONB NOT NULL,
    format VARCHAR(50) NOT NULL DEFAULT 'json', -- 'json', 'csv', 'openai_finetune'
    filter JSONB NOT NULL DEFAULT '{}',
    schedule VARCHAR(100), -- cron expression
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_export_configs_project ON export_configs(project_id);

CREATE TRIGGER update_export_configs_updated_at
    BEFORE UPDATE ON export_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Export jobs
CREATE TABLE IF NOT EXISTS export_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    export_config_id UUID REFERENCES export_configs(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    format VARCHAR(50) NOT NULL,
    filter JSONB NOT NULL DEFAULT '{}',
    result_url TEXT,
    error TEXT,
    row_count INTEGER,
    file_size_bytes BIGINT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_export_jobs_project ON export_jobs(project_id);
CREATE INDEX idx_export_jobs_status ON export_jobs(status);
