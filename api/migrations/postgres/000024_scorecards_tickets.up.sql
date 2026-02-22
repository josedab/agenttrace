-- Agent scorecards table
CREATE TABLE IF NOT EXISTS agent_scorecards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    period VARCHAR(20) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    metrics JSONB,
    trends JSONB,
    grade VARCHAR(2),
    summary TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_scorecard_period CHECK (period IN ('weekly', 'monthly'))
);

-- Indexes for agent_scorecards
CREATE INDEX idx_agent_scorecards_project ON agent_scorecards(project_id);
CREATE INDEX idx_agent_scorecards_project_agent ON agent_scorecards(project_id, agent_name);
CREATE INDEX idx_agent_scorecards_project_time ON agent_scorecards(project_id, created_at DESC);

-- Scorecard configs table
CREATE TABLE IF NOT EXISTS scorecard_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    period VARCHAR(20) NOT NULL,
    recipients JSONB NOT NULL DEFAULT '[]'::jsonb,
    slack_webhook VARCHAR(500),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_scorecard_config_period CHECK (period IN ('weekly', 'monthly')),
    CONSTRAINT unique_scorecard_config UNIQUE (project_id, agent_name)
);

-- Indexes for scorecard_configs
CREATE INDEX idx_scorecard_configs_project ON scorecard_configs(project_id);

-- Trigger for updated_at
CREATE TRIGGER update_scorecard_configs_updated_at
    BEFORE UPDATE ON scorecard_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Ticket integrations table
CREATE TABLE IF NOT EXISTS ticket_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    config JSONB,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_ticket_provider CHECK (provider IN ('GITHUB', 'JIRA', 'LINEAR'))
);

-- Indexes for ticket_integrations
CREATE INDEX idx_ticket_integrations_project ON ticket_integrations(project_id);

-- Trigger for updated_at
CREATE TRIGGER update_ticket_integrations_updated_at
    BEFORE UPDATE ON ticket_integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Ticket results table
CREATE TABLE IF NOT EXISTS ticket_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(255) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    external_id VARCHAR(255),
    url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_ticket_result_provider CHECK (provider IN ('GITHUB', 'JIRA', 'LINEAR'))
);

-- Indexes for ticket_results
CREATE INDEX idx_ticket_results_project ON ticket_results(project_id);
CREATE INDEX idx_ticket_results_trace ON ticket_results(project_id, trace_id);

-- Compliance export jobs table
CREATE TABLE IF NOT EXISTS compliance_export_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    format VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    date_range_start TIMESTAMP WITH TIME ZONE,
    date_range_end TIMESTAMP WITH TIME ZONE,
    file_url VARCHAR(500),
    file_size_bytes BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_export_format CHECK (format IN ('SOC2', 'ISO_42001', 'EU_AI_ACT', 'PDF_SUMMARY', 'JSON_FULL')),
    CONSTRAINT valid_export_status CHECK (status IN ('pending', 'generating', 'completed', 'failed'))
);

-- Indexes for compliance_export_jobs
CREATE INDEX idx_compliance_export_jobs_project ON compliance_export_jobs(project_id);
CREATE INDEX idx_compliance_export_jobs_project_status ON compliance_export_jobs(project_id, status);
