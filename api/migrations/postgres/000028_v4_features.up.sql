-- Migration 000028: V4 Features
-- Memory, Distributed Tracing, Prompt Cache, Chaos Testing, Custom Metrics,
-- Handoffs, Annotations, Carbon Tracking, Synthetic Data, SLOs

-- Memory Snapshots (Feature 1: Agent Memory)
CREATE TABLE IF NOT EXISTS memory_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    step_index INT NOT NULL DEFAULT 0,
    context_window_tokens INT NOT NULL DEFAULT 0,
    used_tokens INT NOT NULL DEFAULT 0,
    retention_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_type VARCHAR(30) NOT NULL DEFAULT 'context',
    content_summary TEXT DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_memory_type CHECK (memory_type IN ('context', 'long_term', 'working', 'episodic'))
);
CREATE INDEX idx_memory_snapshots_project ON memory_snapshots(project_id);
CREATE INDEX idx_memory_snapshots_trace ON memory_snapshots(trace_id);
CREATE INDEX idx_memory_snapshots_trace_step ON memory_snapshots(trace_id, step_index);

-- Distributed Spans (Feature 2: Distributed Tracing)
CREATE TABLE IF NOT EXISTS distributed_spans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(64) NOT NULL,
    parent_span_id VARCHAR(64),
    service_name VARCHAR(255) NOT NULL,
    operation_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ok',
    duration_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    events JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_span_status CHECK (status IN ('ok', 'error', 'unset'))
);
CREATE INDEX idx_distributed_spans_project ON distributed_spans(project_id);
CREATE INDEX idx_distributed_spans_trace ON distributed_spans(trace_id);
CREATE INDEX idx_distributed_spans_parent ON distributed_spans(parent_span_id);
CREATE INDEX idx_distributed_spans_service ON distributed_spans(service_name);

-- Service Map Nodes (Feature 2: Distributed Tracing)
CREATE TABLE IF NOT EXISTS service_map_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    service_type VARCHAR(30) NOT NULL DEFAULT 'service',
    dependencies TEXT[] NOT NULL DEFAULT '{}',
    avg_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    error_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_service_type CHECK (service_type IN ('service', 'database', 'queue', 'cache', 'external', 'agent')),
    CONSTRAINT unique_service_project UNIQUE (project_id, service_name)
);
CREATE INDEX idx_service_map_nodes_project ON service_map_nodes(project_id);
CREATE TRIGGER update_service_map_nodes_updated_at BEFORE UPDATE ON service_map_nodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Prompt Cache Configs (Feature 3: Prompt Cache)
CREATE TABLE IF NOT EXISTS prompt_cache_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    strategy VARCHAR(30) NOT NULL DEFAULT 'lru',
    max_entries INT NOT NULL DEFAULT 1000,
    ttl_seconds INT NOT NULL DEFAULT 3600,
    similarity_threshold DOUBLE PRECISION NOT NULL DEFAULT 0.95,
    excluded_models TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_cache_strategy CHECK (strategy IN ('lru', 'lfu', 'semantic', 'hybrid')),
    CONSTRAINT unique_cache_config_project UNIQUE (project_id)
);
CREATE INDEX idx_prompt_cache_configs_project ON prompt_cache_configs(project_id);
CREATE TRIGGER update_prompt_cache_configs_updated_at BEFORE UPDATE ON prompt_cache_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Prompt Cache Entries (Feature 3: Prompt Cache)
CREATE TABLE IF NOT EXISTS prompt_cache_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    prompt_hash VARCHAR(128) NOT NULL,
    prompt_prefix TEXT NOT NULL DEFAULT '',
    model VARCHAR(255) NOT NULL DEFAULT '',
    hit_count BIGINT NOT NULL DEFAULT 0,
    tokens_saved BIGINT NOT NULL DEFAULT 0,
    cost_saved DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_hit_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_prompt_cache_entries_project ON prompt_cache_entries(project_id);
CREATE INDEX idx_prompt_cache_entries_hash ON prompt_cache_entries(prompt_hash);
CREATE INDEX idx_prompt_cache_entries_expires ON prompt_cache_entries(expires_at);

-- Chaos Experiments (Feature 4: Chaos Testing)
CREATE TABLE IF NOT EXISTS chaos_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    agent_name VARCHAR(255) NOT NULL,
    fault_type VARCHAR(30) NOT NULL DEFAULT 'latency',
    fault_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    schedule VARCHAR(100) DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_fault_type CHECK (fault_type IN ('latency', 'error', 'timeout', 'corruption', 'token_limit', 'rate_limit'))
);
CREATE INDEX idx_chaos_experiments_project ON chaos_experiments(project_id);
CREATE INDEX idx_chaos_experiments_agent ON chaos_experiments(agent_name);
CREATE TRIGGER update_chaos_experiments_updated_at BEFORE UPDATE ON chaos_experiments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Chaos Results (Feature 4: Chaos Testing)
CREATE TABLE IF NOT EXISTS chaos_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES chaos_experiments(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    resilience_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    recovery_time_ms DOUBLE PRECISION,
    degradation_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    observations JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_chaos_result_status CHECK (status IN ('running', 'passed', 'failed', 'aborted'))
);
CREATE INDEX idx_chaos_results_experiment ON chaos_results(experiment_id);
CREATE INDEX idx_chaos_results_status ON chaos_results(status);

-- Custom Metrics (Feature 5: Custom Metrics)
CREATE TABLE IF NOT EXISTS custom_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    unit VARCHAR(50) NOT NULL DEFAULT '',
    metric_type VARCHAR(20) NOT NULL DEFAULT 'gauge',
    aggregation VARCHAR(20) NOT NULL DEFAULT 'avg',
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_metric_type CHECK (metric_type IN ('counter', 'gauge', 'histogram', 'summary')),
    CONSTRAINT valid_aggregation CHECK (aggregation IN ('avg', 'sum', 'min', 'max', 'p50', 'p95', 'p99', 'count')),
    CONSTRAINT unique_metric_project_name UNIQUE (project_id, name)
);
CREATE INDEX idx_custom_metrics_project ON custom_metrics(project_id);
CREATE TRIGGER update_custom_metrics_updated_at BEFORE UPDATE ON custom_metrics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Metric Values (Feature 5: Custom Metrics)
CREATE TABLE IF NOT EXISTS metric_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES custom_metrics(id) ON DELETE CASCADE,
    value DOUBLE PRECISION NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    trace_id VARCHAR(64),
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_metric_values_metric ON metric_values(metric_id);
CREATE INDEX idx_metric_values_recorded ON metric_values(metric_id, recorded_at);

-- Metric Dashboards (Feature 5: Custom Metrics)
CREATE TABLE IF NOT EXISTS metric_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    layout JSONB NOT NULL DEFAULT '[]'::jsonb,
    filters JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_metric_dashboards_project ON metric_dashboards(project_id);
CREATE TRIGGER update_metric_dashboards_updated_at BEFORE UPDATE ON metric_dashboards FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Metric Alerts (Feature 5: Custom Metrics)
CREATE TABLE IF NOT EXISTS metric_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES custom_metrics(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    condition VARCHAR(20) NOT NULL DEFAULT 'gt',
    threshold DOUBLE PRECISION NOT NULL,
    window_seconds INT NOT NULL DEFAULT 300,
    notification_channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_alert_condition CHECK (condition IN ('gt', 'gte', 'lt', 'lte', 'eq', 'neq'))
);
CREATE INDEX idx_metric_alerts_metric ON metric_alerts(metric_id);
CREATE TRIGGER update_metric_alerts_updated_at BEFORE UPDATE ON metric_alerts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Handoffs (Feature 6: Agent Handoffs)
CREATE TABLE IF NOT EXISTS handoffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    from_agent VARCHAR(255) NOT NULL,
    to_agent VARCHAR(255) NOT NULL,
    reason TEXT DEFAULT '',
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    initiated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    context_preserved_pct DOUBLE PRECISION,
    CONSTRAINT valid_handoff_status CHECK (status IN ('pending', 'accepted', 'completed', 'rejected', 'failed'))
);
CREATE INDEX idx_handoffs_project ON handoffs(project_id);
CREATE INDEX idx_handoffs_trace ON handoffs(trace_id);
CREATE INDEX idx_handoffs_status ON handoffs(status);
CREATE INDEX idx_handoffs_agents ON handoffs(from_agent, to_agent);

-- Trace Annotations (Feature 7: Annotations)
CREATE TABLE IF NOT EXISTS trace_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(64),
    user_id UUID NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    annotation_type VARCHAR(20) NOT NULL DEFAULT 'comment',
    is_resolved BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_annotation_type CHECK (annotation_type IN ('comment', 'issue', 'suggestion', 'question', 'highlight'))
);
CREATE INDEX idx_trace_annotations_project ON trace_annotations(project_id);
CREATE INDEX idx_trace_annotations_trace ON trace_annotations(trace_id);
CREATE INDEX idx_trace_annotations_user ON trace_annotations(user_id);
CREATE TRIGGER update_trace_annotations_updated_at BEFORE UPDATE ON trace_annotations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Annotation Replies (Feature 7: Annotations)
CREATE TABLE IF NOT EXISTS annotation_replies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    annotation_id UUID NOT NULL REFERENCES trace_annotations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_annotation_replies_annotation ON annotation_replies(annotation_id);

-- Carbon Configs (Feature 8: Carbon Tracking)
CREATE TABLE IF NOT EXISTS carbon_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    region VARCHAR(100) NOT NULL DEFAULT '',
    provider VARCHAR(100) NOT NULL DEFAULT '',
    energy_mix JSONB NOT NULL DEFAULT '{}'::jsonb,
    carbon_intensity_gco2_kwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    tracking_granularity VARCHAR(20) NOT NULL DEFAULT 'trace',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_tracking_granularity CHECK (tracking_granularity IN ('trace', 'span', 'session', 'daily')),
    CONSTRAINT unique_carbon_config_project UNIQUE (project_id)
);
CREATE INDEX idx_carbon_configs_project ON carbon_configs(project_id);
CREATE TRIGGER update_carbon_configs_updated_at BEFORE UPDATE ON carbon_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Synthetic Datasets (Feature 9: Synthetic Data)
CREATE TABLE IF NOT EXISTS synthetic_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    generation_type VARCHAR(30) NOT NULL DEFAULT 'standard',
    source_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    item_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_generation_type CHECK (generation_type IN ('standard', 'adversarial', 'edge_case', 'augmented', 'template')),
    CONSTRAINT valid_synthetic_status CHECK (status IN ('pending', 'generating', 'completed', 'failed'))
);
CREATE INDEX idx_synthetic_datasets_project ON synthetic_datasets(project_id);
CREATE INDEX idx_synthetic_datasets_status ON synthetic_datasets(status);

-- Synthetic Items (Feature 9: Synthetic Data)
CREATE TABLE IF NOT EXISTS synthetic_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES synthetic_datasets(id) ON DELETE CASCADE,
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    expected_output JSONB,
    difficulty VARCHAR(20) NOT NULL DEFAULT 'medium',
    tags TEXT[] NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_difficulty CHECK (difficulty IN ('easy', 'medium', 'hard', 'adversarial'))
);
CREATE INDEX idx_synthetic_items_dataset ON synthetic_items(dataset_id);
CREATE INDEX idx_synthetic_items_difficulty ON synthetic_items(difficulty);

-- SLOs (Feature 10: Service Level Objectives)
CREATE TABLE IF NOT EXISTS slos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    metric_name VARCHAR(255) NOT NULL,
    target_value DOUBLE PRECISION NOT NULL,
    comparison VARCHAR(10) NOT NULL DEFAULT 'gte',
    window_days INT NOT NULL DEFAULT 30,
    current_value DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'met',
    error_budget_remaining DOUBLE PRECISION NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_slo_comparison CHECK (comparison IN ('gt', 'gte', 'lt', 'lte', 'eq')),
    CONSTRAINT valid_slo_status CHECK (status IN ('met', 'at_risk', 'breached')),
    CONSTRAINT unique_slo_project_name UNIQUE (project_id, name)
);
CREATE INDEX idx_slos_project ON slos(project_id);
CREATE INDEX idx_slos_status ON slos(status);
CREATE TRIGGER update_slos_updated_at BEFORE UPDATE ON slos FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- SLO History (Feature 10: Service Level Objectives)
CREATE TABLE IF NOT EXISTS slo_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slo_id UUID NOT NULL REFERENCES slos(id) ON DELETE CASCADE,
    measured_value DOUBLE PRECISION NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'met',
    error_budget_consumed DOUBLE PRECISION NOT NULL DEFAULT 0,
    violations INT NOT NULL DEFAULT 0,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_slo_history_status CHECK (status IN ('met', 'at_risk', 'breached'))
);
CREATE INDEX idx_slo_history_slo ON slo_history(slo_id);
CREATE INDEX idx_slo_history_period ON slo_history(slo_id, period_start, period_end);
