-- Migration 000029: V5 Features
-- Autonomy Gradient, Cross-Org Benchmarks, Intent Verification, Cost Attribution,
-- Knowledge Graph, Compliance Monitor, Multi-Modal Traces, Collaboration Patterns,
-- Federated Learning, Observability Copilot

-- Autonomy Configs (Feature 1: Autonomy Gradient)
CREATE TABLE IF NOT EXISTS autonomy_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    level VARCHAR(30) NOT NULL DEFAULT 'supervised',
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    constraints JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_autonomy_level CHECK (level IN ('supervised', 'assisted', 'semi_autonomous', 'autonomous', 'fully_autonomous')),
    CONSTRAINT uq_autonomy_project_agent UNIQUE (project_id, agent_id)
);
CREATE INDEX idx_autonomy_configs_project ON autonomy_configs(project_id);
CREATE INDEX idx_autonomy_configs_agent ON autonomy_configs(agent_id);

-- Trust History (Feature 1: Autonomy Gradient)
CREATE TABLE IF NOT EXISTS trust_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    trust_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    reason TEXT DEFAULT '',
    event_type VARCHAR(50) NOT NULL DEFAULT 'manual',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_trust_event CHECK (event_type IN ('manual', 'auto_promote', 'auto_demote', 'violation', 'success'))
);
CREATE INDEX idx_trust_history_project ON trust_history(project_id);
CREATE INDEX idx_trust_history_agent ON trust_history(agent_id);
CREATE INDEX idx_trust_history_created ON trust_history(created_at);

-- Cross-Org Submissions (Feature 2: Cross-Org Benchmarks)
CREATE TABLE IF NOT EXISTS cross_org_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'general',
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    anonymized_hash VARCHAR(64) NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_cross_org_hash UNIQUE (anonymized_hash)
);
CREATE INDEX idx_cross_org_project ON cross_org_submissions(project_id);
CREATE INDEX idx_cross_org_category ON cross_org_submissions(category);
CREATE INDEX idx_cross_org_submitted ON cross_org_submissions(submitted_at);

-- Intent Declarations (Feature 3: Intent Verification)
CREATE TABLE IF NOT EXISTS intent_declarations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    trace_id VARCHAR(64),
    description TEXT NOT NULL,
    expected_outcome TEXT NOT NULL,
    actual_outcome TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'declared',
    alignment_score DOUBLE PRECISION,
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    declared_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_intent_status CHECK (status IN ('declared', 'in_progress', 'verified', 'violated', 'expired'))
);
CREATE INDEX idx_intent_declarations_project ON intent_declarations(project_id);
CREATE INDEX idx_intent_declarations_agent ON intent_declarations(agent_id);
CREATE INDEX idx_intent_declarations_trace ON intent_declarations(trace_id);
CREATE INDEX idx_intent_declarations_status ON intent_declarations(status);

-- Cost Attributions (Feature 4: Cost Attribution)
CREATE TABLE IF NOT EXISTS cost_attributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    feature VARCHAR(255) NOT NULL,
    team VARCHAR(255) NOT NULL DEFAULT 'unassigned',
    cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    business_outcome TEXT DEFAULT '',
    roi DOUBLE PRECISION,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_cost_attributions_project ON cost_attributions(project_id);
CREATE INDEX idx_cost_attributions_trace ON cost_attributions(trace_id);
CREATE INDEX idx_cost_attributions_feature ON cost_attributions(feature);
CREATE INDEX idx_cost_attributions_team ON cost_attributions(team);

-- Knowledge Graph Nodes (Feature 5: Knowledge Graph)
CREATE TABLE IF NOT EXISTS kg_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    node_type VARCHAR(50) NOT NULL,
    label VARCHAR(255) NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_kg_node_type CHECK (node_type IN ('file', 'tool', 'agent', 'trace', 'model', 'prompt', 'dataset'))
);
CREATE INDEX idx_kg_nodes_project ON kg_nodes(project_id);
CREATE INDEX idx_kg_nodes_type ON kg_nodes(node_type);
CREATE INDEX idx_kg_nodes_label ON kg_nodes(label);

-- Knowledge Graph Edges (Feature 5: Knowledge Graph)
CREATE TABLE IF NOT EXISTS kg_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES kg_nodes(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES kg_nodes(id) ON DELETE CASCADE,
    relationship VARCHAR(100) NOT NULL,
    weight DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_kg_edge UNIQUE (project_id, source_id, target_id, relationship)
);
CREATE INDEX idx_kg_edges_project ON kg_edges(project_id);
CREATE INDEX idx_kg_edges_source ON kg_edges(source_id);
CREATE INDEX idx_kg_edges_target ON kg_edges(target_id);
CREATE INDEX idx_kg_edges_relationship ON kg_edges(relationship);

-- Compliance Policies (Feature 6: Compliance Monitor)
CREATE TABLE IF NOT EXISTS compliance_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    framework VARCHAR(100) NOT NULL DEFAULT 'custom',
    description TEXT DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_compliance_policies_project ON compliance_policies(project_id);
CREATE INDEX idx_compliance_policies_framework ON compliance_policies(framework);

-- Compliance Rules (Feature 6: Compliance Monitor)
CREATE TABLE IF NOT EXISTS compliance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES compliance_policies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    condition TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    action VARCHAR(50) NOT NULL DEFAULT 'alert',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_rule_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_rule_action CHECK (action IN ('alert', 'block', 'log', 'quarantine'))
);
CREATE INDEX idx_compliance_rules_policy ON compliance_rules(policy_id);

-- Compliance Scores (Feature 6: Compliance Monitor)
CREATE TABLE IF NOT EXISTS compliance_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    framework VARCHAR(100) NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_score DOUBLE PRECISION NOT NULL DEFAULT 100,
    violations INT NOT NULL DEFAULT 0,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_compliance_scores_project ON compliance_scores(project_id);
CREATE INDEX idx_compliance_scores_framework ON compliance_scores(framework);
CREATE INDEX idx_compliance_scores_evaluated ON compliance_scores(evaluated_at);

-- Trace Attachments (Feature 7: Multi-Modal Traces)
CREATE TABLE IF NOT EXISTS trace_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    attachment_type VARCHAR(30) NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_attachment_type CHECK (attachment_type IN ('image', 'audio', 'video', 'document', 'other'))
);
CREATE INDEX idx_trace_attachments_project ON trace_attachments(project_id);
CREATE INDEX idx_trace_attachments_trace ON trace_attachments(trace_id);
CREATE INDEX idx_trace_attachments_type ON trace_attachments(attachment_type);

-- Collaboration Patterns (Feature 8: Collaboration Patterns)
CREATE TABLE IF NOT EXISTS collab_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL,
    description TEXT DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_pattern_type CHECK (pattern_type IN ('pipeline', 'consensus', 'debate', 'delegation', 'swarm', 'custom'))
);
CREATE INDEX idx_collab_patterns_project ON collab_patterns(project_id);
CREATE INDEX idx_collab_patterns_type ON collab_patterns(pattern_type);

-- Pattern Deployments (Feature 8: Collaboration Patterns)
CREATE TABLE IF NOT EXISTS pattern_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_id UUID NOT NULL REFERENCES collab_patterns(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_deployment_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'))
);
CREATE INDEX idx_pattern_deployments_pattern ON pattern_deployments(pattern_id);
CREATE INDEX idx_pattern_deployments_project ON pattern_deployments(project_id);
CREATE INDEX idx_pattern_deployments_status ON pattern_deployments(status);

-- Federation Rings (Feature 9: Federated Learning)
CREATE TABLE IF NOT EXISTS federation_rings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    privacy_level VARCHAR(30) NOT NULL DEFAULT 'standard',
    min_participants INT NOT NULL DEFAULT 3,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_privacy_level CHECK (privacy_level IN ('basic', 'standard', 'strict', 'maximum')),
    CONSTRAINT valid_ring_status CHECK (status IN ('active', 'paused', 'closed'))
);
CREATE INDEX idx_federation_rings_status ON federation_rings(status);

-- Federated Insights (Feature 9: Federated Learning)
CREATE TABLE IF NOT EXISTS federated_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ring_id UUID NOT NULL REFERENCES federation_rings(id) ON DELETE CASCADE,
    metric VARCHAR(255) NOT NULL,
    aggregated_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    participant_count INT NOT NULL DEFAULT 0,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    noise_added BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_federated_insights_ring ON federated_insights(ring_id);
CREATE INDEX idx_federated_insights_metric ON federated_insights(metric);
CREATE INDEX idx_federated_insights_generated ON federated_insights(generated_at);

-- Federation Configs (Feature 9: Federated Learning)
CREATE TABLE IF NOT EXISTS federation_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ring_id UUID NOT NULL REFERENCES federation_rings(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    privacy_budget DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    noise_multiplier DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    contribution JSONB NOT NULL DEFAULT '{}'::jsonb,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_federation_project_ring UNIQUE (project_id, ring_id)
);
CREATE INDEX idx_federation_configs_project ON federation_configs(project_id);
CREATE INDEX idx_federation_configs_ring ON federation_configs(ring_id);

-- Copilot Queries (Feature 10: Observability Copilot)
CREATE TABLE IF NOT EXISTS copilot_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID,
    question TEXT NOT NULL,
    answer TEXT NOT NULL DEFAULT '',
    sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_time_ms INT NOT NULL DEFAULT 0,
    helpful BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_copilot_queries_project ON copilot_queries(project_id);
CREATE INDEX idx_copilot_queries_user ON copilot_queries(user_id);
CREATE INDEX idx_copilot_queries_created ON copilot_queries(created_at);

-- Copilot Insights (Feature 10: Observability Copilot)
CREATE TABLE IF NOT EXISTS copilot_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    affected_traces INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    acknowledged BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_insight_severity CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    CONSTRAINT valid_insight_category CHECK (category IN ('performance', 'cost', 'reliability', 'security', 'quality'))
);
CREATE INDEX idx_copilot_insights_project ON copilot_insights(project_id);
CREATE INDEX idx_copilot_insights_category ON copilot_insights(category);
CREATE INDEX idx_copilot_insights_severity ON copilot_insights(severity);
CREATE INDEX idx_copilot_insights_created ON copilot_insights(created_at);
