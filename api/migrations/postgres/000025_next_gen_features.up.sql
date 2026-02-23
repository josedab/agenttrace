-- Migration: Next-Gen Features (streaming, diff intelligence, federation, playbooks)
-- Depends on: 000024_scorecards_tickets

-- ============================================================
-- Diff Analyses (Feature 2: Agent Diff Intelligence)
-- ============================================================
CREATE TABLE IF NOT EXISTS diff_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- File change counts
    files_added INT NOT NULL DEFAULT 0,
    files_modified INT NOT NULL DEFAULT 0,
    files_deleted INT NOT NULL DEFAULT 0,
    lines_added INT NOT NULL DEFAULT 0,
    lines_removed INT NOT NULL DEFAULT 0,

    -- Quality scores
    overall_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    dimension_scores JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Analysis results
    findings JSONB NOT NULL DEFAULT '[]'::jsonb,
    file_analyses JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Metadata
    agent_name VARCHAR(255) DEFAULT '',
    git_commit_sha VARCHAR(64) DEFAULT '',
    git_branch VARCHAR(255) DEFAULT '',

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_diff_status CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

CREATE INDEX idx_diff_analyses_project_id ON diff_analyses(project_id);
CREATE INDEX idx_diff_analyses_trace_id ON diff_analyses(trace_id);
CREATE INDEX idx_diff_analyses_status ON diff_analyses(status);
CREATE INDEX idx_diff_analyses_created_at ON diff_analyses(project_id, created_at DESC);
CREATE INDEX idx_diff_analyses_score ON diff_analyses(project_id, overall_score);

-- ============================================================
-- Federation Peers (Feature 10: OTLP Federation)
-- ============================================================
CREATE TABLE IF NOT EXISTS federation_peers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(1024) NOT NULL,
    api_key VARCHAR(512) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'connected',
    last_seen TIMESTAMP WITH TIME ZONE,

    -- Health metrics
    traces_exported BIGINT NOT NULL DEFAULT 0,
    spans_exported BIGINT NOT NULL DEFAULT 0,
    error_count BIGINT NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_export_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_peer_status CHECK (status IN ('connected', 'disconnected', 'error'))
);

CREATE INDEX idx_federation_peers_project_id ON federation_peers(project_id);

CREATE TRIGGER update_federation_peers_updated_at
    BEFORE UPDATE ON federation_peers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Export Destinations (Feature 10: OTLP Federation)
-- ============================================================
CREATE TABLE IF NOT EXISTS export_destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    endpoint VARCHAR(1024) NOT NULL,
    protocol VARCHAR(10) NOT NULL DEFAULT 'grpc',
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    sampling DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    batch_size INT NOT NULL DEFAULT 100,
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Export stats
    total_exported BIGINT NOT NULL DEFAULT 0,
    total_failed BIGINT NOT NULL DEFAULT 0,
    last_export_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT DEFAULT '',
    avg_batch_ms DOUBLE PRECISION NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_dest_type CHECK (type IN ('datadog', 'grafana', 'honeycomb', 'jaeger', 'newrelic', 'custom')),
    CONSTRAINT valid_dest_protocol CHECK (protocol IN ('grpc', 'http')),
    CONSTRAINT valid_dest_status CHECK (status IN ('active', 'paused', 'error')),
    CONSTRAINT valid_sampling CHECK (sampling >= 0.0 AND sampling <= 1.0)
);

CREATE INDEX idx_export_destinations_project_id ON export_destinations(project_id);

CREATE TRIGGER update_export_destinations_updated_at
    BEFORE UPDATE ON export_destinations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Intervention Requests (Feature 1: Real-Time Streaming)
-- ============================================================
CREATE TABLE IF NOT EXISTS intervention_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id VARCHAR(64) NOT NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL,
    message TEXT DEFAULT '',
    user_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_intervention_action CHECK (action IN ('pause', 'resume', 'cancel', 'message')),
    CONSTRAINT valid_intervention_status CHECK (status IN ('pending', 'delivered', 'acknowledged'))
);

CREATE INDEX idx_intervention_requests_trace_id ON intervention_requests(trace_id);
CREATE INDEX idx_intervention_requests_project_id ON intervention_requests(project_id);
CREATE INDEX idx_intervention_requests_status ON intervention_requests(status);

-- ============================================================
-- Guardrail Playbooks (Feature 4: Guardrails Engine)
-- ============================================================
CREATE TABLE IF NOT EXISTS guard_playbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    template VARCHAR(50) DEFAULT 'custom',
    enforce_mode VARCHAR(20) NOT NULL DEFAULT 'warn',
    enabled BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_enforce_mode CHECK (enforce_mode IN ('audit', 'warn', 'enforce'))
);

CREATE INDEX idx_guard_playbooks_project_id ON guard_playbooks(project_id);

CREATE TRIGGER update_guard_playbooks_updated_at
    BEFORE UPDATE ON guard_playbooks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Link playbooks to their rules
CREATE TABLE IF NOT EXISTS guard_playbook_rules (
    playbook_id UUID NOT NULL REFERENCES guard_playbooks(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES guard_rules(id) ON DELETE CASCADE,
    PRIMARY KEY (playbook_id, rule_id)
);

-- ============================================================
-- Discussion Threads (Feature 3: Collaboration Workspace)
-- ============================================================
CREATE TABLE IF NOT EXISTS discussion_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    observation_id UUID,
    title VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_by UUID NOT NULL,
    created_by_name VARCHAR(255) NOT NULL DEFAULT '',
    tags TEXT[] DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_thread_status CHECK (status IN ('open', 'resolved', 'archived'))
);

CREATE INDEX idx_discussion_threads_project_id ON discussion_threads(project_id);
CREATE INDEX idx_discussion_threads_trace_id ON discussion_threads(trace_id);

CREATE TRIGGER update_discussion_threads_updated_at
    BEFORE UPDATE ON discussion_threads
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS thread_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES discussion_threads(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    user_name VARCHAR(255) NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    mentions UUID[] DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_thread_messages_thread_id ON thread_messages(thread_id);

-- ============================================================
-- Evaluation Queues (Feature 3: Collaboration Workspace)
-- ============================================================
CREATE TABLE IF NOT EXISTS eval_queues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    assignees UUID[] DEFAULT '{}',
    trace_ids UUID[] DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Progress tracking
    total INT NOT NULL DEFAULT 0,
    completed INT NOT NULL DEFAULT 0,
    in_progress INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_queue_status CHECK (status IN ('active', 'completed', 'paused'))
);

CREATE INDEX idx_eval_queues_project_id ON eval_queues(project_id);

CREATE TRIGGER update_eval_queues_updated_at
    BEFORE UPDATE ON eval_queues
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Alert Channels (Feature 5: Anomaly Detection)
-- ============================================================
CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_channel_type CHECK (type IN ('slack', 'pagerduty', 'email', 'webhook', 'teams'))
);

CREATE INDEX idx_alert_channels_project_id ON alert_channels(project_id);

CREATE TRIGGER update_alert_channels_updated_at
    BEFORE UPDATE ON alert_channels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Cost Autopilot Configuration (Feature 8: Cost Optimizer)
-- ============================================================
CREATE TABLE IF NOT EXISTS cost_autopilot_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    max_budget_daily DOUBLE PRECISION DEFAULT 0,
    max_budget_monthly DOUBLE PRECISION DEFAULT 0,
    optimization_level VARCHAR(20) NOT NULL DEFAULT 'balanced',
    auto_apply BOOLEAN NOT NULL DEFAULT false,
    notify_on_save BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_autopilot_per_project UNIQUE (project_id),
    CONSTRAINT valid_optimization_level CHECK (optimization_level IN ('conservative', 'balanced', 'aggressive'))
);

CREATE TRIGGER update_cost_autopilot_configs_updated_at
    BEFORE UPDATE ON cost_autopilot_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
