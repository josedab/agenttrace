-- SSO Configurations table
CREATE TABLE IF NOT EXISTS sso_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    enforce_sso BOOLEAN NOT NULL DEFAULT false,
    allowed_domains TEXT[] DEFAULT '{}',

    -- SAML Configuration
    saml_entity_id VARCHAR(500),
    saml_sso_url VARCHAR(500),
    saml_slo_url VARCHAR(500),
    saml_certificate TEXT,
    saml_sign_requests BOOLEAN DEFAULT false,
    saml_name_id_format VARCHAR(100),

    -- OIDC Configuration
    oidc_client_id VARCHAR(500),
    oidc_client_secret VARCHAR(500),
    oidc_issuer_url VARCHAR(500),
    oidc_scopes TEXT[] DEFAULT '{openid,profile,email}',

    -- Attribute Mapping (stored as JSONB)
    attribute_mapping JSONB DEFAULT '{
        "email": "email",
        "firstName": "given_name",
        "lastName": "family_name",
        "displayName": "name",
        "groups": "groups",
        "department": "department"
    }'::jsonb,

    -- Auto-provisioning
    auto_provision_users BOOLEAN DEFAULT true,
    default_role VARCHAR(50) DEFAULT 'member',
    auto_assign_projects UUID[] DEFAULT '{}',

    -- Metadata
    metadata_url VARCHAR(500),
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_error_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_provider CHECK (provider IN ('saml', 'oidc', 'okta', 'azure_ad', 'google'))
);

-- Only one SSO config per organization
CREATE UNIQUE INDEX idx_sso_configurations_org ON sso_configurations(organization_id);
CREATE INDEX idx_sso_configurations_provider ON sso_configurations(provider) WHERE enabled = true;

-- SSO Sessions table
CREATE TABLE IF NOT EXISTS sso_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(500) NOT NULL,
    session_index VARCHAR(500),

    -- Tokens (encrypted in production)
    access_token TEXT,
    refresh_token TEXT,
    id_token TEXT,

    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sso_sessions_user ON sso_sessions(user_id);
CREATE INDEX idx_sso_sessions_org ON sso_sessions(organization_id);
CREATE INDEX idx_sso_sessions_external ON sso_sessions(external_id);
CREATE INDEX idx_sso_sessions_expires ON sso_sessions(expires_at) WHERE expires_at > NOW();

-- SSO State table (for OAuth/SAML state parameter)
CREATE TABLE IF NOT EXISTS sso_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    state VARCHAR(100) NOT NULL UNIQUE,
    return_url VARCHAR(500),
    nonce VARCHAR(100),
    code_verifier VARCHAR(200),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sso_states_state ON sso_states(state);
CREATE INDEX idx_sso_states_expires ON sso_states(expires_at);

-- Identity Provider Mappings (for user linking)
CREATE TABLE IF NOT EXISTS sso_identity_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(500) NOT NULL,
    external_email VARCHAR(255),
    external_name VARCHAR(255),
    attributes JSONB DEFAULT '{}'::jsonb,
    linked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT unique_identity_mapping UNIQUE (organization_id, provider, external_id)
);

CREATE INDEX idx_sso_identity_mappings_user ON sso_identity_mappings(user_id);
CREATE INDEX idx_sso_identity_mappings_org ON sso_identity_mappings(organization_id);

-- Trigger for updated_at
CREATE TRIGGER update_sso_configurations_updated_at
    BEFORE UPDATE ON sso_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Cleanup function for expired states
CREATE OR REPLACE FUNCTION cleanup_expired_sso_states() RETURNS void AS $$
BEGIN
    DELETE FROM sso_states WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
