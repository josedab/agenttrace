-- Migration 000027: V3 Features
-- Orchestration, RCA, Agent Versions, Predictions, Embed, Agent Builder, Fleet, Privacy, Mobile, Plugins

-- Orchestration Sessions (Feature 1: Multi-Agent Debugger)
CREATE TABLE IF NOT EXISTS orchestration_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    agent_ids TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    state JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_orchestration_status CHECK (status IN ('created', 'running', 'paused', 'completed', 'failed'))
);
CREATE INDEX idx_orchestration_sessions_project ON orchestration_sessions(project_id);
CREATE INDEX idx_orchestration_sessions_status ON orchestration_sessions(status);
CREATE TRIGGER update_orchestration_sessions_updated_at BEFORE UPDATE ON orchestration_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS orchestration_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES orchestration_sessions(id) ON DELETE CASCADE,
    from_agent VARCHAR(255) NOT NULL DEFAULT '',
    to_agent VARCHAR(255) NOT NULL DEFAULT '',
    message_type VARCHAR(30) NOT NULL DEFAULT 'text',
    content TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    step_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_message_type CHECK (message_type IN ('text', 'tool_call', 'tool_result', 'system', 'error'))
);
CREATE INDEX idx_orchestration_messages_session ON orchestration_messages(session_id);
CREATE INDEX idx_orchestration_messages_step ON orchestration_messages(session_id, step_index);

-- RCA Reports (Feature 2: Root Cause Analysis)
CREATE TABLE IF NOT EXISTS rca_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'analyzing',
    root_cause TEXT DEFAULT '',
    causal_chain JSONB NOT NULL DEFAULT '[]'::jsonb,
    remediation JSONB NOT NULL DEFAULT '[]'::jsonb,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_rca_status CHECK (status IN ('analyzing', 'completed', 'failed')),
    CONSTRAINT valid_rca_severity CHECK (severity IN ('low', 'medium', 'high', 'critical'))
);
CREATE INDEX idx_rca_reports_project ON rca_reports(project_id);
CREATE INDEX idx_rca_reports_trace ON rca_reports(trace_id);
CREATE INDEX idx_rca_reports_status ON rca_reports(status);

-- Agent Versions (Feature 3: Agent Versioning)
CREATE TABLE IF NOT EXISTS agent_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_name VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    change_log TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT unique_agent_version UNIQUE (project_id, agent_name, version)
);
CREATE INDEX idx_agent_versions_project ON agent_versions(project_id);
CREATE INDEX idx_agent_versions_agent ON agent_versions(project_id, agent_name);
CREATE INDEX idx_agent_versions_active ON agent_versions(project_id, agent_name, is_active) WHERE is_active = true;

-- Cost Predictions (Feature 4: Pre-Execution Predictions)
CREATE TABLE IF NOT EXISTS cost_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_description TEXT NOT NULL DEFAULT '',
    model VARCHAR(255) NOT NULL DEFAULT '',
    predicted_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    predicted_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    predicted_quality DOUBLE PRECISION NOT NULL DEFAULT 0,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    constraints JSONB NOT NULL DEFAULT '{}'::jsonb,
    actual_cost DOUBLE PRECISION,
    actual_latency_ms DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'predicted',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_prediction_status CHECK (status IN ('predicted', 'approved', 'rejected', 'executed'))
);
CREATE INDEX idx_cost_predictions_project ON cost_predictions(project_id);
CREATE INDEX idx_cost_predictions_status ON cost_predictions(status);

-- Budget Approvals (Feature 4: Pre-Execution Predictions)
CREATE TABLE IF NOT EXISTS budget_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prediction_id UUID NOT NULL REFERENCES cost_predictions(id) ON DELETE CASCADE,
    requested_by UUID,
    decided_by UUID,
    decision VARCHAR(20) NOT NULL DEFAULT 'pending',
    reason TEXT DEFAULT '',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT valid_approval_decision CHECK (decision IN ('pending', 'approved', 'rejected'))
);
CREATE INDEX idx_budget_approvals_prediction ON budget_approvals(prediction_id);
CREATE INDEX idx_budget_approvals_decision ON budget_approvals(decision);

-- Embed Configs (Feature 5: Embedding & White-Label)
CREATE TABLE IF NOT EXISTS embed_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    allowed_domains TEXT[] NOT NULL DEFAULT '{}',
    branding JSONB NOT NULL DEFAULT '{}'::jsonb,
    dashboard_ids TEXT[] NOT NULL DEFAULT '{}',
    token_secret VARCHAR(512) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_embed_project UNIQUE (project_id)
);
CREATE INDEX idx_embed_configs_project ON embed_configs(project_id);
CREATE TRIGGER update_embed_configs_updated_at BEFORE UPDATE ON embed_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Agent Blueprints (Feature 6: Agent Builder)
CREATE TABLE IF NOT EXISTS agent_blueprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_description TEXT NOT NULL DEFAULT '',
    generated_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    recommended_model VARCHAR(255) NOT NULL DEFAULT '',
    recommended_tools JSONB NOT NULL DEFAULT '[]'::jsonb,
    estimated_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    deployed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_blueprint_status CHECK (status IN ('draft', 'ready', 'deployed', 'archived'))
);
CREATE INDEX idx_agent_blueprints_project ON agent_blueprints(project_id);
CREATE INDEX idx_agent_blueprints_status ON agent_blueprints(status);
CREATE TRIGGER update_agent_blueprints_updated_at BEFORE UPDATE ON agent_blueprints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Fleet Policies (Feature 7: Fleet Management)
CREATE TABLE IF NOT EXISTS fleet_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    scope VARCHAR(20) NOT NULL DEFAULT 'project',
    priority INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_fleet_scope CHECK (scope IN ('global', 'project', 'agent'))
);
CREATE INDEX idx_fleet_policies_project ON fleet_policies(project_id);
CREATE INDEX idx_fleet_policies_scope ON fleet_policies(scope);
CREATE TRIGGER update_fleet_policies_updated_at BEFORE UPDATE ON fleet_policies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- PII Configs (Feature 8: Privacy & PII)
CREATE TABLE IF NOT EXISTS pii_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    pii_types TEXT[] NOT NULL DEFAULT '{}',
    redaction_mode VARCHAR(20) NOT NULL DEFAULT 'mask',
    residency_region VARCHAR(50) NOT NULL DEFAULT '',
    auto_scan BOOLEAN NOT NULL DEFAULT false,
    retention_days INT NOT NULL DEFAULT 90,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_redaction_mode CHECK (redaction_mode IN ('mask', 'hash', 'remove', 'encrypt')),
    CONSTRAINT unique_pii_project UNIQUE (project_id)
);
CREATE INDEX idx_pii_configs_project ON pii_configs(project_id);
CREATE TRIGGER update_pii_configs_updated_at BEFORE UPDATE ON pii_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Data Deletion Requests (Feature 8: Privacy & PII)
CREATE TABLE IF NOT EXISTS data_deletion_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subject_id VARCHAR(255) NOT NULL,
    reason TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    requested_by UUID,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_deletion_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);
CREATE INDEX idx_data_deletion_requests_project ON data_deletion_requests(project_id);
CREATE INDEX idx_data_deletion_requests_status ON data_deletion_requests(status);

-- Mobile Devices (Feature 9: Mobile App)
CREATE TABLE IF NOT EXISTS mobile_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    device_token VARCHAR(512) NOT NULL,
    platform VARCHAR(20) NOT NULL,
    device_name VARCHAR(255) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_active_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_mobile_platform CHECK (platform IN ('ios', 'android', 'web'))
);
CREATE INDEX idx_mobile_devices_user ON mobile_devices(user_id);
CREATE INDEX idx_mobile_devices_token ON mobile_devices(device_token);

-- Push Notifications (Feature 9: Mobile App)
CREATE TABLE IF NOT EXISTS push_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES mobile_devices(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_notification_status CHECK (status IN ('pending', 'sent', 'delivered', 'read', 'failed'))
);
CREATE INDEX idx_push_notifications_device ON push_notifications(device_id);
CREATE INDEX idx_push_notifications_status ON push_notifications(status);

-- Plugins (Feature 10: Plugin System)
CREATE TABLE IF NOT EXISTS plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    source VARCHAR(512) NOT NULL DEFAULT '',
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    type VARCHAR(30) NOT NULL DEFAULT 'evaluator',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'installed',
    installed_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_plugin_type CHECK (type IN ('evaluator', 'processor', 'widget', 'exporter', 'custom')),
    CONSTRAINT valid_plugin_status CHECK (status IN ('installed', 'active', 'disabled', 'error')),
    CONSTRAINT unique_plugin_project_name UNIQUE (project_id, name)
);
CREATE INDEX idx_plugins_project ON plugins(project_id);
CREATE INDEX idx_plugins_status ON plugins(status);
CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Plugin Executions (Feature 10: Plugin System)
CREATE TABLE IF NOT EXISTS plugin_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    duration_ms INT,
    error TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_execution_status CHECK (status IN ('running', 'completed', 'failed'))
);
CREATE INDEX idx_plugin_executions_plugin ON plugin_executions(plugin_id);
CREATE INDEX idx_plugin_executions_status ON plugin_executions(status);
