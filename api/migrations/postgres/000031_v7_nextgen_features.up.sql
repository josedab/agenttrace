-- Next-Gen v7 Features Migration
-- Features: Workflow Simulator, Auto-Discovery, Cloud Onboarding, AI Debugger,
--           Prompt Optimization, Cost Alerting, Regression Suite, Collab Hub,
--           OTel Compatibility, Security Scanner

-- Feature 1: Agent Workflow Simulator
CREATE TABLE IF NOT EXISTS workflow_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'simulating', 'completed', 'archived')),
    nodes JSONB NOT NULL DEFAULT '[]'::jsonb,
    edges JSONB NOT NULL DEFAULT '[]'::jsonb,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);
CREATE INDEX idx_workflow_definitions_project ON workflow_definitions(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS workflow_simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflow_definitions(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    predicted_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    predicted_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    predicted_quality_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    trace_data_used INTEGER NOT NULL DEFAULT 0,
    results JSONB NOT NULL DEFAULT '[]'::jsonb,
    scenario_overrides JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_workflow_simulations_workflow ON workflow_simulations(workflow_id, created_at DESC);

-- Feature 2: Zero-Config Auto-Discovery
CREATE TABLE IF NOT EXISTS discovered_frameworks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    framework VARCHAR(30) NOT NULL
        CHECK (framework IN ('langchain', 'crewai', 'autogen', 'llamaindex', 'openai', 'anthropic', 'custom')),
    version VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'detected'
        CHECK (status IN ('detected', 'confirmed', 'instrumented', 'disabled')),
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    components JSONB NOT NULL DEFAULT '[]'::jsonb,
    auto_instrumented BOOLEAN NOT NULL DEFAULT false,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE(project_id, framework)
);
CREATE INDEX idx_discovered_frameworks_project ON discovered_frameworks(project_id);

-- Feature 3: Cloud Onboarding
CREATE TABLE IF NOT EXISTS cloud_onboarding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    current_step VARCHAR(30) NOT NULL DEFAULT 'account_created',
    sdk_language VARCHAR(30),
    framework VARCHAR(50),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS usage_meters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    period VARCHAR(20) NOT NULL,
    traces_used BIGINT NOT NULL DEFAULT 0,
    traces_limit BIGINT NOT NULL DEFAULT 10000,
    storage_used_bytes BIGINT NOT NULL DEFAULT 0,
    storage_limit_bytes BIGINT NOT NULL DEFAULT 1073741824,
    api_calls_used BIGINT NOT NULL DEFAULT 0,
    api_calls_limit BIGINT NOT NULL DEFAULT 100000,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, period)
);

-- Feature 4: AI-Powered Trace Debugger
CREATE TABLE IF NOT EXISTS debug_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    query TEXT NOT NULL,
    query_type VARCHAR(20) NOT NULL
        CHECK (query_type IN ('root_cause', 'explain', 'suggest_fix', 'compare', 'optimize')),
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    response JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);
CREATE INDEX idx_debug_queries_trace ON debug_queries(project_id, trace_id, created_at DESC);

-- Feature 5: Continuous Prompt Optimization
CREATE TABLE IF NOT EXISTS prompt_optimizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    prompt_id UUID NOT NULL,
    prompt_version INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'analyzing'
        CHECK (status IN ('analyzing', 'generating', 'testing', 'promoting', 'completed', 'failed')),
    failure_patterns JSONB NOT NULL DEFAULT '[]'::jsonb,
    variants JSONB NOT NULL DEFAULT '[]'::jsonb,
    best_variant_id UUID,
    improvement_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_prompt_optimizations_project ON prompt_optimizations(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS optimization_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    min_samples_analysis INTEGER NOT NULL DEFAULT 100,
    min_samples_promotion INTEGER NOT NULL DEFAULT 200,
    p_value_threshold DOUBLE PRECISION NOT NULL DEFAULT 0.05,
    require_approval BOOLEAN NOT NULL DEFAULT true,
    max_variants INTEGER NOT NULL DEFAULT 3,
    schedule_cron VARCHAR(50) DEFAULT '0 */6 * * *',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Feature 6: Real-Time Cost Anomaly Alerting
CREATE TABLE IF NOT EXISTS cost_alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning'
        CHECK (severity IN ('info', 'warning', 'critical', 'emergency')),
    actions JSONB NOT NULL DEFAULT '["notify"]'::jsonb,
    condition JSONB NOT NULL DEFAULT '{}'::jsonb,
    channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    cooldown_minutes INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_cost_alert_rules_project ON cost_alert_rules(project_id);

CREATE TABLE IF NOT EXISTS cost_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    rule_id UUID REFERENCES cost_alert_rules(id) ON DELETE SET NULL,
    severity VARCHAR(20) NOT NULL,
    action VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    current_cost DOUBLE PRECISION NOT NULL,
    threshold_cost DOUBLE PRECISION NOT NULL,
    affected_trace_id UUID,
    affected_model VARCHAR(100),
    channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    sent_at TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_cost_alerts_project ON cost_alerts(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS circuit_breaker_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    state VARCHAR(20) NOT NULL DEFAULT 'closed'
        CHECK (state IN ('closed', 'half_open', 'open')),
    max_cost_per_minute DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    max_cost_per_hour DOUBLE PRECISION NOT NULL DEFAULT 10.0,
    fallback_model_chain JSONB NOT NULL DEFAULT '[]'::jsonb,
    cooldown_seconds INTEGER NOT NULL DEFAULT 300,
    last_tripped_at TIMESTAMPTZ,
    reset_after_seconds INTEGER NOT NULL DEFAULT 600,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Feature 7: Agent Regression Test Suite
CREATE TABLE IF NOT EXISTS golden_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(20) NOT NULL
        CHECK (category IN ('bug_fix', 'refactoring', 'test_writing', 'feature_impl', 'code_review')),
    language VARCHAR(30) NOT NULL DEFAULT 'python',
    items JSONB NOT NULL DEFAULT '[]'::jsonb,
    item_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);
CREATE INDEX idx_golden_datasets_project ON golden_datasets(project_id);

CREATE TABLE IF NOT EXISTS regression_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    suite_id UUID NOT NULL REFERENCES golden_datasets(id) ON DELETE CASCADE,
    agent_config TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'passed', 'failed', 'error')),
    results JSONB NOT NULL DEFAULT '[]'::jsonb,
    pass_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_tests INTEGER NOT NULL DEFAULT 0,
    passed INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    baseline_comparison JSONB,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_regression_runs_project ON regression_runs(project_id, created_at DESC);

-- Feature 8: Collaboration Hub
CREATE TABLE IF NOT EXISTS review_queues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    filters JSONB NOT NULL DEFAULT '{}'::jsonb,
    assigned_to JSONB NOT NULL DEFAULT '[]'::jsonb,
    pending_count INTEGER NOT NULL DEFAULT 0,
    completed_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_review_queues_project ON review_queues(project_id);

CREATE TABLE IF NOT EXISTS review_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_id UUID NOT NULL REFERENCES review_queues(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    assigned_to UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_review', 'approved', 'rejected', 'needs_changes')),
    feedback TEXT,
    score DOUBLE PRECISION,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_review_assignments_queue ON review_assignments(queue_id, status);

CREATE TABLE IF NOT EXISTS quality_standards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    enforce_on_deploy BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS activity_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    activity_type VARCHAR(30) NOT NULL,
    user_id UUID NOT NULL,
    user_name VARCHAR(200),
    description TEXT NOT NULL,
    resource_id VARCHAR(100),
    resource_type VARCHAR(50),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_activity_feed_project ON activity_feed(project_id, timestamp DESC);

-- Feature 9: OTel Native Compatibility
CREATE TABLE IF NOT EXISTS otel_export_destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    format VARCHAR(20) NOT NULL
        CHECK (format IN ('otlp_grpc', 'otlp_http', 'jaeger', 'zipkin')),
    endpoint VARCHAR(500) NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    tls_enabled BOOLEAN NOT NULL DEFAULT false,
    sampling_rate DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    batch_size INTEGER NOT NULL DEFAULT 512,
    flush_interval_ms INTEGER NOT NULL DEFAULT 5000,
    last_export_at TIMESTAMPTZ,
    exported_count BIGINT NOT NULL DEFAULT 0,
    error_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_otel_export_destinations_project ON otel_export_destinations(project_id);

-- Feature 10: Agent Security Scanner
CREATE TABLE IF NOT EXISTS security_scan_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    observation_id UUID,
    findings JSONB NOT NULL DEFAULT '[]'::jsonb,
    overall_risk VARCHAR(20) NOT NULL DEFAULT 'low'
        CHECK (overall_risk IN ('low', 'medium', 'high', 'critical')),
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scan_duration_ms BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX idx_security_scan_results_project ON security_scan_results(project_id, scanned_at DESC);
CREATE INDEX idx_security_scan_results_trace ON security_scan_results(trace_id);

CREATE TABLE IF NOT EXISTS security_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    action VARCHAR(20) NOT NULL DEFAULT 'warn'
        CHECK (action IN ('log', 'warn', 'block', 'quarantine')),
    exclude_patterns JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_security_policies_project ON security_policies(project_id);
