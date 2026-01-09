-- GitHub App installations table
CREATE TABLE IF NOT EXISTS github_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    installation_id BIGINT NOT NULL UNIQUE,
    account_id BIGINT NOT NULL,
    account_login VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) NOT NULL, -- 'User' or 'Organization'
    target_type VARCHAR(50) NOT NULL,
    app_id BIGINT NOT NULL,
    app_slug VARCHAR(255) NOT NULL,
    repository_selection VARCHAR(50) NOT NULL DEFAULT 'selected', -- 'all' or 'selected'
    access_tokens_url TEXT,
    repositories_url TEXT,
    html_url TEXT,
    permissions JSONB DEFAULT '{}',
    events TEXT[] DEFAULT '{}',
    suspended_at TIMESTAMP WITH TIME ZONE,
    suspended_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- GitHub repositories linked to projects
CREATE TABLE IF NOT EXISTS github_repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id UUID NOT NULL REFERENCES github_installations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    repo_id BIGINT NOT NULL,
    repo_full_name VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    owner VARCHAR(255) NOT NULL,
    private BOOLEAN DEFAULT false,
    default_branch VARCHAR(255) DEFAULT 'main',
    html_url TEXT,
    clone_url TEXT,
    sync_enabled BOOLEAN DEFAULT true,
    auto_link BOOLEAN DEFAULT true, -- Auto-link commits to traces
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(installation_id, repo_id)
);

-- GitHub webhook events for processing
CREATE TABLE IF NOT EXISTS github_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id BIGINT NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    action VARCHAR(100),
    delivery_id VARCHAR(100) NOT NULL UNIQUE,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_github_installations_org ON github_installations(organization_id);
CREATE INDEX idx_github_installations_account ON github_installations(account_login);
CREATE INDEX idx_github_installations_installation_id ON github_installations(installation_id);

CREATE INDEX idx_github_repositories_installation ON github_repositories(installation_id);
CREATE INDEX idx_github_repositories_project ON github_repositories(project_id);
CREATE INDEX idx_github_repositories_full_name ON github_repositories(repo_full_name);

CREATE INDEX idx_github_webhook_events_installation ON github_webhook_events(installation_id);
CREATE INDEX idx_github_webhook_events_type ON github_webhook_events(event_type);
CREATE INDEX idx_github_webhook_events_processed ON github_webhook_events(processed) WHERE NOT processed;
CREATE INDEX idx_github_webhook_events_created ON github_webhook_events(created_at);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_github_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER github_installations_updated_at
    BEFORE UPDATE ON github_installations
    FOR EACH ROW
    EXECUTE FUNCTION update_github_updated_at();

CREATE TRIGGER github_repositories_updated_at
    BEFORE UPDATE ON github_repositories
    FOR EACH ROW
    EXECUTE FUNCTION update_github_updated_at();
