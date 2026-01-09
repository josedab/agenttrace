-- Audit Logs table (partitioned by time for performance)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    actor_id UUID,
    actor_email VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL DEFAULT 'user',
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    resource_name VARCHAR(255),
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    session_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, created_at),

    CONSTRAINT valid_actor_type CHECK (actor_type IN ('user', 'api_key', 'system')),
    CONSTRAINT valid_resource_type CHECK (resource_type IN (
        'user', 'organization', 'project', 'api_key', 'sso',
        'prompt', 'dataset', 'evaluator', 'trace', 'settings'
    ))
) PARTITION BY RANGE (created_at);

-- Create partitions for the current and next months
CREATE TABLE audit_logs_y2026m01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE audit_logs_y2026m02 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE audit_logs_y2026m03 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE audit_logs_y2026m04 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE audit_logs_y2026m05 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE audit_logs_y2026m06 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

-- Create indexes
CREATE INDEX idx_audit_logs_org_created ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id, created_at DESC) WHERE actor_id IS NOT NULL;
CREATE INDEX idx_audit_logs_action ON audit_logs(action, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);
CREATE INDEX idx_audit_logs_ip ON audit_logs(ip_address, created_at DESC) WHERE ip_address IS NOT NULL;

-- Full-text search index on description and resource_name
CREATE INDEX idx_audit_logs_search ON audit_logs USING gin(
    to_tsvector('english', coalesce(description, '') || ' ' || coalesce(resource_name, ''))
);

-- Audit Retention Policies table
CREATE TABLE IF NOT EXISTS audit_retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    retention_days INTEGER NOT NULL DEFAULT 365,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_retention_policy UNIQUE (organization_id),
    CONSTRAINT valid_retention_days CHECK (retention_days >= 0)
);

CREATE TRIGGER update_audit_retention_policies_updated_at
    BEFORE UPDATE ON audit_retention_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Audit Export Jobs table (for async export)
CREATE TABLE IF NOT EXISTS audit_export_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    filter JSONB NOT NULL,
    format VARCHAR(10) NOT NULL DEFAULT 'csv',
    compress BOOLEAN NOT NULL DEFAULT true,
    file_path VARCHAR(500),
    file_size BIGINT,
    record_count INTEGER,
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    CONSTRAINT valid_format CHECK (format IN ('csv', 'json'))
);

CREATE INDEX idx_audit_export_jobs_org ON audit_export_jobs(organization_id, created_at DESC);
CREATE INDEX idx_audit_export_jobs_status ON audit_export_jobs(status) WHERE status IN ('pending', 'processing');

-- Function to create new audit log partitions
CREATE OR REPLACE FUNCTION create_audit_log_partition(partition_date DATE) RETURNS void AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    partition_name := 'audit_logs_y' || to_char(partition_date, 'YYYY') || 'm' || to_char(partition_date, 'MM');
    start_date := date_trunc('month', partition_date)::DATE;
    end_date := (date_trunc('month', partition_date) + interval '1 month')::DATE;

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_logs FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        start_date,
        end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Function to apply retention policy
CREATE OR REPLACE FUNCTION apply_audit_retention_policy(org_id UUID) RETURNS INTEGER AS $$
DECLARE
    retention_days INTEGER;
    deleted_count INTEGER;
BEGIN
    SELECT arp.retention_days INTO retention_days
    FROM audit_retention_policies arp
    WHERE arp.organization_id = org_id AND arp.enabled = true;

    IF retention_days IS NULL OR retention_days = 0 THEN
        RETURN 0;
    END IF;

    DELETE FROM audit_logs
    WHERE organization_id = org_id
    AND created_at < NOW() - (retention_days || ' days')::INTERVAL;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Insert default actions for reference (optional, for UI dropdowns)
CREATE TABLE IF NOT EXISTS audit_action_definitions (
    action VARCHAR(100) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info'
);

INSERT INTO audit_action_definitions (action, category, description, severity) VALUES
    ('login', 'authentication', 'User logged in', 'info'),
    ('logout', 'authentication', 'User logged out', 'info'),
    ('login_failed', 'authentication', 'Failed login attempt', 'warning'),
    ('sso_login', 'authentication', 'User logged in via SSO', 'info'),
    ('api_key_used', 'authentication', 'API key was used', 'info'),
    ('user_created', 'user_management', 'New user was created', 'info'),
    ('user_updated', 'user_management', 'User details were updated', 'info'),
    ('user_deleted', 'user_management', 'User was deleted', 'warning'),
    ('user_invited', 'user_management', 'User was invited', 'info'),
    ('user_role_changed', 'user_management', 'User role was changed', 'info'),
    ('org_created', 'organization', 'Organization was created', 'info'),
    ('org_updated', 'organization', 'Organization was updated', 'info'),
    ('org_deleted', 'organization', 'Organization was deleted', 'critical'),
    ('member_added', 'organization', 'Member was added to organization', 'info'),
    ('member_removed', 'organization', 'Member was removed from organization', 'info'),
    ('project_created', 'project', 'Project was created', 'info'),
    ('project_updated', 'project', 'Project was updated', 'info'),
    ('project_deleted', 'project', 'Project was deleted', 'warning'),
    ('api_key_created', 'api_key', 'API key was created', 'info'),
    ('api_key_revoked', 'api_key', 'API key was revoked', 'info'),
    ('sso_configured', 'sso', 'SSO was configured', 'info'),
    ('sso_enabled', 'sso', 'SSO was enabled', 'info'),
    ('sso_disabled', 'sso', 'SSO was disabled', 'warning'),
    ('data_exported', 'data', 'Data was exported', 'info'),
    ('data_deleted', 'data', 'Data was deleted', 'warning'),
    ('settings_changed', 'settings', 'Settings were changed', 'info'),
    ('prompt_created', 'prompt', 'Prompt was created', 'info'),
    ('prompt_updated', 'prompt', 'Prompt was updated', 'info'),
    ('prompt_deleted', 'prompt', 'Prompt was deleted', 'warning'),
    ('prompt_published', 'prompt', 'Prompt was published', 'info'),
    ('dataset_created', 'dataset', 'Dataset was created', 'info'),
    ('dataset_updated', 'dataset', 'Dataset was updated', 'info'),
    ('dataset_deleted', 'dataset', 'Dataset was deleted', 'warning'),
    ('evaluator_created', 'evaluator', 'Evaluator was created', 'info'),
    ('evaluator_updated', 'evaluator', 'Evaluator was updated', 'info'),
    ('evaluator_deleted', 'evaluator', 'Evaluator was deleted', 'warning')
ON CONFLICT (action) DO NOTHING;
