-- Prompt Regression CI Gate: baseline score storage and history
CREATE TABLE IF NOT EXISTS prompt_regression_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    gate_config_id UUID NOT NULL,
    run_id UUID NOT NULL,
    branch VARCHAR(255) NOT NULL,
    commit_sha VARCHAR(64) NOT NULL,
    pr_number INTEGER,
    passed BOOLEAN NOT NULL DEFAULT true,
    severity VARCHAR(20) NOT NULL DEFAULT 'none',
    metric_deltas JSONB NOT NULL DEFAULT '{}'::jsonb,
    blocked_pr BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_regression_severity CHECK (severity IN ('none', 'minor', 'major', 'critical'))
);

CREATE INDEX idx_prompt_regression_history_project ON prompt_regression_history(project_id);
CREATE INDEX idx_prompt_regression_history_branch ON prompt_regression_history(project_id, branch);
CREATE INDEX idx_prompt_regression_history_created ON prompt_regression_history(created_at DESC);

-- Prompt CI baseline scores persistence
CREATE TABLE IF NOT EXISTS prompt_ci_baselines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    dataset_id UUID NOT NULL,
    prompt_id UUID NOT NULL,
    prompt_version INTEGER NOT NULL DEFAULT 1,
    name VARCHAR(200) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    scores JSONB NOT NULL DEFAULT '{}'::jsonb,
    sample_size INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,

    CONSTRAINT valid_baseline_status CHECK (status IN ('active', 'superseded', 'archived'))
);

CREATE INDEX idx_prompt_ci_baselines_project ON prompt_ci_baselines(project_id);
CREATE INDEX idx_prompt_ci_baselines_prompt ON prompt_ci_baselines(project_id, prompt_id);

-- Prompt CI gate configurations persistence
CREATE TABLE IF NOT EXISTS prompt_ci_gate_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    baseline_id UUID NOT NULL,
    thresholds JSONB NOT NULL DEFAULT '{}'::jsonb,
    block_on_severity VARCHAR(20) NOT NULL DEFAULT 'major',
    confidence_level DOUBLE PRECISION NOT NULL DEFAULT 0.95,
    required_metrics JSONB DEFAULT '[]'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prompt_ci_gate_configs_project ON prompt_ci_gate_configs(project_id);

-- Prompt CI runs persistence
CREATE TABLE IF NOT EXISTS prompt_ci_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    baseline_id UUID NOT NULL,
    branch VARCHAR(255) NOT NULL,
    commit_sha VARCHAR(64) NOT NULL,
    pr_number INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    score_comparison JSONB NOT NULL DEFAULT '[]'::jsonb,
    overall_severity VARCHAR(20) NOT NULL DEFAULT 'none',
    summary TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_prompt_ci_runs_project ON prompt_ci_runs(project_id);
CREATE INDEX idx_prompt_ci_runs_baseline ON prompt_ci_runs(baseline_id);

-- Universal Agent Protocol Adapter
CREATE TABLE IF NOT EXISTS agent_adapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    framework VARCHAR(50) NOT NULL,
    version VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'registered',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    capabilities JSONB DEFAULT '[]'::jsonb,
    lifecycle_hooks JSONB DEFAULT '[]'::jsonb,
    total_traces BIGINT DEFAULT 0,
    total_spans BIGINT DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION DEFAULT 0,
    error_rate DOUBLE PRECISION DEFAULT 0,
    last_active_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_adapter_status CHECK (status IN ('registered', 'active', 'inactive', 'deprecated')),
    CONSTRAINT valid_adapter_framework CHECK (framework IN ('langchain', 'crewai', 'autogen', 'langgraph', 'openhands', 'semantic_kernel', 'custom'))
);

CREATE INDEX idx_agent_adapters_project ON agent_adapters(project_id);
CREATE INDEX idx_agent_adapters_framework ON agent_adapters(framework);

-- Adapter events log
CREATE TABLE IF NOT EXISTS adapter_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    adapter_id UUID NOT NULL REFERENCES agent_adapters(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    trace_id VARCHAR(100),
    span_id VARCHAR(100),
    name VARCHAR(200),
    status_code VARCHAR(20),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_adapter_events_adapter ON adapter_events(adapter_id);
CREATE INDEX idx_adapter_events_trace ON adapter_events(trace_id);

-- Agent Cost Autopilot
CREATE TABLE IF NOT EXISTS cost_hotspots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    total_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    trace_count INTEGER NOT NULL DEFAULT 0,
    avg_cost_per_trace DOUBLE PRECISION NOT NULL DEFAULT 0,
    trend VARCHAR(20) NOT NULL DEFAULT 'stable',
    trend_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    model_breakdown JSONB DEFAULT '[]'::jsonb,
    analyzed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cost_hotspots_project ON cost_hotspots(project_id);

CREATE TABLE IF NOT EXISTS cost_autopilot_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    rule_type VARCHAR(50) NOT NULL,
    condition JSONB NOT NULL DEFAULT '{}'::jsonb,
    action JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    execution_count INTEGER NOT NULL DEFAULT 0,
    last_executed TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_autopilot_rule_type CHECK (rule_type IN ('model_downgrade', 'cache_enable', 'rate_limit', 'budget_alert'))
);

CREATE INDEX idx_cost_autopilot_rules_project ON cost_autopilot_rules(project_id);

-- Multi-Agent Topology
CREATE TABLE IF NOT EXISTS delegation_chains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    initiator_id VARCHAR(100) NOT NULL,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    total_time_ms BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_chain_status CHECK (status IN ('active', 'completed', 'failed'))
);

CREATE INDEX idx_delegation_chains_session ON delegation_chains(session_id);

-- Collaborative Trace Review
CREATE TABLE IF NOT EXISTS trace_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    requested_by UUID NOT NULL,
    assigned_to JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    labels JSONB DEFAULT '[]'::jsonb,
    due_at TIMESTAMP WITH TIME ZONE,
    approval_count INTEGER NOT NULL DEFAULT 0,
    required_approvals INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_review_status CHECK (status IN ('pending', 'in_review', 'approved', 'rejected', 'closed'))
);

CREATE INDEX idx_trace_reviews_project ON trace_reviews(project_id);
CREATE INDEX idx_trace_reviews_status ON trace_reviews(project_id, status);

CREATE TABLE IF NOT EXISTS review_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES trace_reviews(id) ON DELETE CASCADE,
    parent_id UUID,
    author_id UUID NOT NULL,
    author_name VARCHAR(200),
    content TEXT NOT NULL,
    mentions JSONB DEFAULT '[]'::jsonb,
    span_id VARCHAR(100),
    resolved BOOLEAN NOT NULL DEFAULT false,
    reactions JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_comments_review ON review_comments(review_id);

CREATE TABLE IF NOT EXISTS notification_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    webhook_url TEXT,
    channel_id VARCHAR(100),
    enabled BOOLEAN NOT NULL DEFAULT true,
    events JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_integrations_project ON notification_integrations(project_id);

-- Intelligent Alerting with RCA
CREATE TABLE IF NOT EXISTS correlated_anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    anomaly_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    title VARCHAR(300) NOT NULL,
    description TEXT,
    affected_traces JSONB DEFAULT '[]'::jsonb,
    affected_models JSONB DEFAULT '[]'::jsonb,
    correlation DOUBLE PRECISION DEFAULT 0,
    root_causes JSONB DEFAULT '[]'::jsonb,
    remediations JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    assigned_to UUID,
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_correlated_anomalies_project ON correlated_anomalies(project_id);
CREATE INDEX idx_correlated_anomalies_status ON correlated_anomalies(project_id, status);

CREATE TABLE IF NOT EXISTS alert_delivery_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    test_status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_delivery_channels_project ON alert_delivery_channels(project_id);

CREATE TABLE IF NOT EXISTS correlation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    anomaly_types JSONB DEFAULT '[]'::jsonb,
    window_minutes INTEGER NOT NULL DEFAULT 30,
    min_correlation DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    auto_remediate BOOLEAN NOT NULL DEFAULT false,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    channels JSONB DEFAULT '[]'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_correlation_rules_project ON correlation_rules(project_id);

CREATE TABLE IF NOT EXISTS rca_investigations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    anomaly_id UUID NOT NULL,
    title VARCHAR(300) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'open',
    findings JSONB DEFAULT '[]'::jsonb,
    timeline JSONB DEFAULT '[]'::jsonb,
    root_cause TEXT,
    resolution TEXT,
    investigator_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rca_investigations_project ON rca_investigations(project_id);

-- Prompt A/B Testing
CREATE TABLE IF NOT EXISTS ab_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    prompt_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    variants JSONB NOT NULL DEFAULT '[]'::jsonb,
    traffic_split JSONB NOT NULL DEFAULT '{}'::jsonb,
    target_metric VARCHAR(50) NOT NULL,
    secondary_metrics JSONB DEFAULT '[]'::jsonb,
    min_sample_size INTEGER NOT NULL DEFAULT 100,
    confidence_level DOUBLE PRECISION NOT NULL DEFAULT 0.95,
    winner_id UUID,
    auto_select_winner BOOLEAN NOT NULL DEFAULT false,
    gradual_rollout JSONB,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_ab_test_status CHECK (status IN ('draft', 'running', 'paused', 'completed', 'cancelled'))
);

CREATE INDEX idx_ab_tests_project ON ab_tests(project_id);
CREATE INDEX idx_ab_tests_status ON ab_tests(project_id, status);

CREATE TABLE IF NOT EXISTS ab_test_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL,
    score DOUBLE PRECISION,
    latency_ms DOUBLE PRECISION,
    cost_usd DOUBLE PRECISION,
    tokens INTEGER,
    is_error BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ab_test_results_test ON ab_test_results(test_id);
CREATE INDEX idx_ab_test_results_variant ON ab_test_results(test_id, variant_id);

-- Federated Trace Analytics
CREATE TABLE IF NOT EXISTS federated_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT,
    impact VARCHAR(20) NOT NULL DEFAULT 'medium',
    recommendation TEXT,
    benchmark_value DOUBLE PRECISION,
    your_value DOUBLE PRECISION,
    percentile DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_federated_insights_instance ON federated_insights(instance_id);

CREATE TABLE IF NOT EXISTS privacy_budgets (
    instance_id UUID PRIMARY KEY,
    total_epsilon DOUBLE PRECISION NOT NULL DEFAULT 10.0,
    used_epsilon DOUBLE PRECISION NOT NULL DEFAULT 0,
    queries_count INTEGER NOT NULL DEFAULT 0,
    reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + interval '30 days')
);

-- Self-Healing Agent Guardrails
CREATE TABLE IF NOT EXISTS self_healing_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    rule_id UUID NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    remediation_action JSONB NOT NULL DEFAULT '{}'::jsonb,
    circuit_breaker JSONB,
    retry_policy JSONB,
    fallback_chain JSONB DEFAULT '[]'::jsonb,
    trigger_count INTEGER NOT NULL DEFAULT 0,
    last_triggered TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_self_healing_policies_project ON self_healing_policies(project_id);
CREATE INDEX idx_self_healing_policies_rule ON self_healing_policies(rule_id);

CREATE TABLE IF NOT EXISTS guardrail_audit_trail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES self_healing_policies(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    trace_id VARCHAR(100),
    span_id VARCHAR(100),
    details JSONB DEFAULT '{}'::jsonb,
    original_input TEXT,
    remediated_output TEXT,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT true,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guardrail_audit_trail_policy ON guardrail_audit_trail(policy_id);
CREATE INDEX idx_guardrail_audit_trail_timestamp ON guardrail_audit_trail(timestamp DESC);
