-- Migration 000026: Next-Gen Features V2
-- Prompt Lab, Security Sandbox, Orchestration, Marketplace, Training, Compliance, RBAC

-- Prompt Experiments (Feature 2: Prompt Lab)
CREATE TABLE IF NOT EXISTS prompt_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    prompt_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    winner_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_experiment_status CHECK (status IN ('draft', 'running', 'completed', 'cancelled'))
);
CREATE INDEX idx_prompt_experiments_project ON prompt_experiments(project_id);
CREATE TRIGGER update_prompt_experiments_updated_at BEFORE UPDATE ON prompt_experiments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS prompt_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES prompt_experiments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    prompt_content TEXT NOT NULL,
    traffic_weight DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    is_control BOOLEAN NOT NULL DEFAULT false,
    traces INT NOT NULL DEFAULT 0,
    avg_quality DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_tokens DOUBLE PRECISION NOT NULL DEFAULT 0,
    error_rate DOUBLE PRECISION NOT NULL DEFAULT 0
);
CREATE INDEX idx_prompt_variants_experiment ON prompt_variants(experiment_id);

-- Sandbox Reviews (Feature 3: Security Sandbox)
CREATE TABLE IF NOT EXISTS sandbox_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    proposed_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
    risk_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    reviewer_id UUID,
    review_note TEXT DEFAULT '',
    policy_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),
    CONSTRAINT valid_sandbox_status CHECK (status IN ('pending', 'reviewing', 'approved', 'rejected', 'expired')),
    CONSTRAINT valid_risk_level CHECK (risk_level IN ('low', 'medium', 'high', 'critical'))
);
CREATE INDEX idx_sandbox_reviews_project ON sandbox_reviews(project_id);
CREATE INDEX idx_sandbox_reviews_status ON sandbox_reviews(status);

CREATE TABLE IF NOT EXISTS sandbox_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    allowed_paths TEXT[] DEFAULT '{}',
    blocked_paths TEXT[] DEFAULT '{}',
    allowed_commands TEXT[] DEFAULT '{}',
    blocked_commands TEXT[] DEFAULT '{}',
    allow_network BOOLEAN NOT NULL DEFAULT false,
    allow_env_access BOOLEAN NOT NULL DEFAULT false,
    require_review VARCHAR(20) NOT NULL DEFAULT 'high_risk',
    max_file_size BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_review_level CHECK (require_review IN ('always', 'high_risk', 'never'))
);
CREATE INDEX idx_sandbox_policies_project ON sandbox_policies(project_id);
CREATE TRIGGER update_sandbox_policies_updated_at BEFORE UPDATE ON sandbox_policies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Webhook Rules (Feature 8: Orchestration)
CREATE TABLE IF NOT EXISTS webhook_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    trigger VARCHAR(50) NOT NULL,
    condition JSONB NOT NULL DEFAULT '{}'::jsonb,
    action VARCHAR(50) NOT NULL,
    action_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    cooldown_minutes INT NOT NULL DEFAULT 5,
    last_fired_at TIMESTAMP WITH TIME ZONE,
    fire_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_wh_trigger CHECK (trigger IN ('cost_exceeded', 'error_detected', 'guardrail_violation', 'anomaly_detected', 'eval_score_low', 'trace_completed')),
    CONSTRAINT valid_wh_action CHECK (action IN ('slack', 'pagerduty', 'jira', 'github_issue', 'email', 'custom_webhook'))
);
CREATE INDEX idx_webhook_rules_project ON webhook_rules(project_id);

CREATE TABLE IF NOT EXISTS webhook_rule_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES webhook_rules(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payload TEXT DEFAULT '',
    response TEXT DEFAULT '',
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_delivery_status CHECK (status IN ('pending', 'success', 'failed', 'retrying'))
);
CREATE INDEX idx_webhook_rule_deliveries_rule ON webhook_rule_deliveries(rule_id);
CREATE INDEX idx_webhook_rule_deliveries_status ON webhook_rule_deliveries(status);

-- Marketplace (Feature 9)
CREATE TABLE IF NOT EXISTS marketplace_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    type VARCHAR(20) NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    author VARCHAR(255) NOT NULL DEFAULT '',
    tags TEXT[] DEFAULT '{}',
    downloads INT NOT NULL DEFAULT 0,
    rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating_count INT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT true,
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_package_type CHECK (type IN ('prompt', 'guardrail', 'evaluator', 'benchmark', 'bundle'))
);
CREATE INDEX idx_marketplace_packages_type ON marketplace_packages(type);
CREATE INDEX idx_marketplace_packages_public ON marketplace_packages(is_public) WHERE is_public = true;
CREATE TRIGGER update_marketplace_packages_updated_at BEFORE UPDATE ON marketplace_packages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS package_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES marketplace_packages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    score INT NOT NULL CHECK (score BETWEEN 1 AND 5),
    review TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_rating UNIQUE (package_id, user_id)
);

-- Training Datasets (Feature 6)
CREATE TABLE IF NOT EXISTS training_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    format VARCHAR(30) NOT NULL DEFAULT 'jsonl',
    source_filter JSONB NOT NULL DEFAULT '{}'::jsonb,
    item_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'building',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_training_format CHECK (format IN ('openai_finetune', 'anthropic_rlhf', 'jsonl')),
    CONSTRAINT valid_training_status CHECK (status IN ('building', 'ready', 'exported'))
);
CREATE INDEX idx_training_datasets_project ON training_datasets(project_id);

CREATE TABLE IF NOT EXISTS training_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES training_datasets(id) ON DELETE CASCADE,
    format VARCHAR(30) NOT NULL,
    url TEXT DEFAULT '',
    line_count INT NOT NULL DEFAULT 0,
    token_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Compliance Reports (Feature 10)
CREATE TABLE IF NOT EXISTS compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    template VARCHAR(20) NOT NULL,
    title VARCHAR(500) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'generating',
    sections JSONB NOT NULL DEFAULT '[]'::jsonb,
    summary TEXT DEFAULT '',
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_report_template CHECK (template IN ('eu_ai_act', 'soc2', 'custom')),
    CONSTRAINT valid_report_status CHECK (status IN ('generating', 'ready', 'error'))
);
CREATE INDEX idx_compliance_reports_project ON compliance_reports(project_id);

-- Role Assignments (Feature 7: RBAC)
CREATE TABLE IF NOT EXISTS role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    granted_by UUID,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_role CHECK (role IN ('admin', 'developer', 'viewer', 'auditor')),
    CONSTRAINT unique_user_project_role UNIQUE (user_id, project_id)
);
CREATE INDEX idx_role_assignments_user ON role_assignments(user_id);
CREATE INDEX idx_role_assignments_project ON role_assignments(project_id);

-- SSO Configurations (Feature 7: RBAC)
CREATE TABLE IF NOT EXISTS sso_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider VARCHAR(10) NOT NULL,
    issuer_url VARCHAR(1024) NOT NULL DEFAULT '',
    client_id VARCHAR(255) NOT NULL DEFAULT '',
    client_secret VARCHAR(512) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT false,
    auto_provision BOOLEAN NOT NULL DEFAULT false,
    default_role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_sso_provider CHECK (provider IN ('saml', 'oidc')),
    CONSTRAINT unique_org_sso UNIQUE (org_id)
);
CREATE TRIGGER update_sso_configs_updated_at BEFORE UPDATE ON sso_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- API Key Scopes (Feature 7: RBAC)
CREATE TABLE IF NOT EXISTS api_key_scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    permissions TEXT[] NOT NULL DEFAULT '{}',
    resource_types TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_key_scope UNIQUE (api_key_id)
);
