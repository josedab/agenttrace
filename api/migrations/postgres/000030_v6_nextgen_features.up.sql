-- Next-Gen Features Migration (v6)
-- Features: Agent Replay, Cost Guardrails, Multi-Agent Graph, Prompt CI,
--           Agent Benchmarks, Semantic Trace Search, Knowledge Graph,
--           IDE Trace View, Federated Aggregation

-- Feature 1: Real-Time Agent Replay
CREATE TABLE IF NOT EXISTS replay_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'recording'
        CHECK (status IN ('recording', 'completed', 'playing', 'paused', 'failed')),
    recording_fidelity VARCHAR(20) NOT NULL DEFAULT 'standard'
        CHECK (recording_fidelity IN ('full', 'standard', 'minimal')),
    total_events INTEGER NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    files_tracked INTEGER NOT NULL DEFAULT 0,
    checkpoint_count INTEGER NOT NULL DEFAULT 0,
    parent_session_id UUID REFERENCES replay_sessions(id) ON DELETE SET NULL,
    branch_point INTEGER DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT false,
    share_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_replay_sessions_project ON replay_sessions(project_id, created_at DESC);
CREATE INDEX idx_replay_sessions_trace ON replay_sessions(trace_id);

CREATE TABLE IF NOT EXISTS replay_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES replay_sessions(id) ON DELETE CASCADE,
    event_index INTEGER NOT NULL,
    event_type VARCHAR(30) NOT NULL
        CHECK (event_type IN ('llm_call', 'file_operation', 'terminal_command', 'checkpoint', 'state_change', 'agent_message', 'tool_call', 'error')),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    input JSONB,
    output JSONB,
    duration_ms BIGINT DEFAULT 0,
    state_snapshot_id UUID,
    observation_id UUID,
    file_delta JSONB,
    UNIQUE(session_id, event_index)
);

CREATE INDEX idx_replay_events_session ON replay_events(session_id, event_index);

-- Feature 2: Intelligent Cost Guardrails
CREATE TABLE IF NOT EXISTS cost_guardrail_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    policy_type VARCHAR(20) NOT NULL
        CHECK (policy_type IN ('per_project', 'per_user', 'per_model', 'per_session')),
    action VARCHAR(20) NOT NULL DEFAULT 'warn'
        CHECK (action IN ('warn', 'throttle', 'pause', 'notify', 'downgrade_model')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    budget_limit DOUBLE PRECISION NOT NULL,
    budget_period VARCHAR(20) NOT NULL DEFAULT 'monthly'
        CHECK (budget_period IN ('daily', 'weekly', 'monthly', 'quarterly')),
    current_spend DOUBLE PRECISION NOT NULL DEFAULT 0,
    threshold_percent DOUBLE PRECISION NOT NULL DEFAULT 80,
    model_downgrade_map JSONB NOT NULL DEFAULT '{}'::jsonb,
    notify_channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    cooldown_minutes INTEGER NOT NULL DEFAULT 60,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'triggered', 'paused', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cost_guardrail_policies_project ON cost_guardrail_policies(project_id);

CREATE TABLE IF NOT EXISTS cost_guardrail_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES cost_guardrail_policies(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id UUID,
    user_id VARCHAR(255),
    action VARCHAR(20) NOT NULL,
    amount_at_violation DOUBLE PRECISION NOT NULL,
    budget_limit DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cost_guardrail_violations_project ON cost_guardrail_violations(project_id, timestamp DESC);

-- Feature 3: Multi-Agent Collaboration Graph
CREATE TABLE IF NOT EXISTS multi_agent_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    topology VARCHAR(20) NOT NULL DEFAULT 'pipeline'
        CHECK (topology IN ('pipeline', 'hub_spoke', 'mesh', 'debate', 'hierarchical')),
    agents JSONB NOT NULL DEFAULT '[]'::jsonb,
    messages JSONB NOT NULL DEFAULT '[]'::jsonb,
    bottlenecks JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'completed', 'failed')),
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_multi_agent_sessions_project ON multi_agent_sessions(project_id, created_at DESC);
CREATE INDEX idx_multi_agent_sessions_trace ON multi_agent_sessions(trace_id);

-- Feature 5: Prompt Regression Testing in CI
CREATE TABLE IF NOT EXISTS prompt_baselines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    dataset_id UUID,
    prompt_id UUID,
    prompt_version INTEGER,
    name VARCHAR(200) NOT NULL,
    branch VARCHAR(200) NOT NULL DEFAULT 'main',
    scores JSONB NOT NULL DEFAULT '{}'::jsonb,
    sample_size INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'superseded', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);

CREATE INDEX idx_prompt_baselines_project ON prompt_baselines(project_id, branch);

CREATE TABLE IF NOT EXISTS prompt_ci_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    baseline_id UUID NOT NULL REFERENCES prompt_baselines(id) ON DELETE CASCADE,
    branch VARCHAR(200) NOT NULL,
    commit_sha VARCHAR(40) NOT NULL,
    pr_number INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    score_comparison JSONB NOT NULL DEFAULT '[]'::jsonb,
    overall_severity VARCHAR(20) NOT NULL DEFAULT 'none'
        CHECK (overall_severity IN ('none', 'minor', 'major', 'critical')),
    summary TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_prompt_ci_runs_project ON prompt_ci_runs(project_id, started_at DESC);
CREATE INDEX idx_prompt_ci_runs_baseline ON prompt_ci_runs(baseline_id);

-- Feature 6: Agent Performance Benchmarks
CREATE TABLE IF NOT EXISTS agent_benchmark_suites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(20) NOT NULL
        CHECK (category IN ('bug_fix', 'feature_impl', 'refactoring', 'code_review', 'test_writing')),
    tasks JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);

CREATE INDEX idx_agent_benchmark_suites_project ON agent_benchmark_suites(project_id);

CREATE TABLE IF NOT EXISTS agent_benchmark_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    suite_id UUID NOT NULL REFERENCES agent_benchmark_suites(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    results JSONB NOT NULL DEFAULT '[]'::jsonb,
    overall_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_agent_benchmark_runs_suite ON agent_benchmark_runs(suite_id, overall_score DESC);

-- Feature 7: Semantic Trace Search
CREATE TABLE IF NOT EXISTS trace_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    observation_id UUID,
    content_type VARCHAR(20) NOT NULL DEFAULT 'trace'
        CHECK (content_type IN ('trace', 'observation', 'generation', 'session')),
    content_hash VARCHAR(64) NOT NULL,
    embedding_model VARCHAR(100) NOT NULL DEFAULT 'text-embedding-3-small',
    indexed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trace_embeddings_project ON trace_embeddings(project_id);
CREATE INDEX idx_trace_embeddings_trace ON trace_embeddings(trace_id);

CREATE TABLE IF NOT EXISTS trace_clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    label VARCHAR(200) NOT NULL,
    description TEXT,
    trace_count INTEGER NOT NULL DEFAULT 0,
    common_patterns JSONB NOT NULL DEFAULT '[]'::jsonb,
    avg_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    representative_trace_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trace_clusters_project ON trace_clusters(project_id);

-- Feature 8: Agent Knowledge Graph (stored as JSONB for flexibility)
CREATE TABLE IF NOT EXISTS agent_knowledge_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    nodes JSONB NOT NULL DEFAULT '[]'::jsonb,
    edges JSONB NOT NULL DEFAULT '[]'::jsonb,
    stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_knowledge_snapshots_project ON agent_knowledge_snapshots(project_id, generated_at DESC);

-- Feature 9: IDE Trace View (mapping stored per file)
CREATE TABLE IF NOT EXISTS file_trace_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    file_path VARCHAR(1000) NOT NULL,
    annotations JSONB NOT NULL DEFAULT '[]'::jsonb,
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, file_path)
);

CREATE INDEX idx_file_trace_mappings_project ON file_trace_mappings(project_id);

-- Feature 10: Federated Trace Aggregation
CREATE TABLE IF NOT EXISTS federated_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    api_key_hash VARCHAR(255) NOT NULL,
    privacy_level VARCHAR(30) NOT NULL DEFAULT 'aggregated_only'
        CHECK (privacy_level IN ('full', 'aggregated_only', 'differential_privacy')),
    last_sync_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'syncing', 'error')),
    metrics_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_federated_instances_project ON federated_instances(project_id);

CREATE TABLE IF NOT EXISTS federated_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES federated_instances(id) ON DELETE CASCADE,
    metric_type VARCHAR(20) NOT NULL
        CHECK (metric_type IN ('latency', 'cost', 'error_rate', 'throughput', 'token_usage')),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    p50 DOUBLE PRECISION,
    p95 DOUBLE PRECISION,
    p99 DOUBLE PRECISION,
    std_dev DOUBLE PRECISION,
    sample_count BIGINT NOT NULL DEFAULT 0,
    model_distribution JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_federated_metrics_instance ON federated_metrics(instance_id, metric_type, period_start DESC);
