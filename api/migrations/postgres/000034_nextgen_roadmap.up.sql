-- Safe replay planning. Additional roadmap tables are appended in this migration
-- so a single release boundary owns the next-generation domain additions.
CREATE TABLE IF NOT EXISTS replay_plans (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    checkpoint_id UUID,
    status VARCHAR(20) NOT NULL
        CHECK (status IN ('planned', 'ready', 'running', 'completed', 'failed', 'unsupported')),
    request JSONB NOT NULL DEFAULT '{}'::jsonb,
    capabilities JSONB NOT NULL DEFAULT '{}'::jsonb,
    result JSONB,
    failure_reason TEXT NOT NULL DEFAULT '',
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_replay_plans_project
    ON replay_plans(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_replay_plans_trace
    ON replay_plans(project_id, trace_id, created_at DESC);

ALTER TABLE replay_events
    DROP CONSTRAINT IF EXISTS replay_events_event_type_check;
ALTER TABLE replay_events
    ADD CONSTRAINT replay_events_event_type_check
    CHECK (event_type IN (
        'llm_call',
        'tool_call',
        'file_operation',
        'terminal_command',
        'checkpoint',
        'git_operation',
        'user_input',
        'agent_thought',
        'error',
        'state_change',
        'agent_message'
    ));

CREATE TABLE IF NOT EXISTS eval_hub_packages (
    id UUID PRIMARY KEY,
    owner_project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    kind VARCHAR(20) NOT NULL
        CHECK (kind IN ('dataset', 'evaluator', 'prompt', 'experiment', 'benchmark')),
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility VARCHAR(20) NOT NULL DEFAULT 'private'
        CHECK (visibility IN ('private', 'organization', 'public')),
    latest_version INTEGER NOT NULL DEFAULT 1 CHECK (latest_version > 0),
    forked_from_package_id UUID REFERENCES eval_hub_packages(id) ON DELETE SET NULL,
    forked_from_version INTEGER,
    published_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_eval_hub_packages_access
    ON eval_hub_packages(visibility, organization_id, owner_project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_eval_hub_packages_kind
    ON eval_hub_packages(kind, updated_at DESC);

CREATE TABLE IF NOT EXISTS eval_hub_versions (
    id UUID PRIMARY KEY,
    package_id UUID NOT NULL REFERENCES eval_hub_packages(id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    source_resource_id UUID NOT NULL,
    manifest JSONB NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    version_note TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(package_id, version)
);

CREATE TABLE IF NOT EXISTS eval_hub_runs (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    package_id UUID NOT NULL REFERENCES eval_hub_packages(id) ON DELETE CASCADE,
    package_version INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL
        CHECK (status IN ('ready', 'running', 'completed', 'unsupported', 'failed')),
    dataset_run_id UUID,
    experiment_id UUID,
    result JSONB,
    capability_message TEXT NOT NULL DEFAULT '',
    idempotency_key VARCHAR(200),
    created_by UUID NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE(project_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_eval_hub_runs_project
    ON eval_hub_runs(project_id, started_at DESC);

CREATE OR REPLACE VIEW eval_hub_package_latest AS
SELECT
    p.id AS package_id,
    p.owner_project_id,
    p.organization_id,
    p.kind,
    p.name,
    p.description,
    p.visibility,
    p.latest_version,
    p.forked_from_package_id,
    p.forked_from_version,
    p.published_by,
    p.created_at AS package_created_at,
    p.updated_at AS package_updated_at,
    v.id AS version_id,
    v.source_resource_id,
    v.manifest,
    v.checksum,
    v.version_note,
    v.created_by AS version_created_by,
    v.created_at AS version_created_at
FROM eval_hub_packages p
JOIN eval_hub_versions v
    ON v.package_id = p.id AND v.version = p.latest_version;

CREATE TABLE IF NOT EXISTS share_links (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    resource_type VARCHAR(20) NOT NULL
        CHECK (resource_type IN ('trace', 'replay_plan')),
    resource_id VARCHAR(128) NOT NULL,
    token_hash BYTEA NOT NULL UNIQUE,
    redaction_version INTEGER NOT NULL DEFAULT 1,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_share_links_project
    ON share_links(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_share_links_expiry
    ON share_links(expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS migration_import_items (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES migration_jobs(id) ON DELETE CASCADE,
    source_type VARCHAR(30) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    imported_id VARCHAR(255),
    status VARCHAR(20) NOT NULL CHECK (status IN ('imported', 'failed')),
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (job_id, source_type, source_id)
);

CREATE INDEX IF NOT EXISTS idx_migration_import_items_project
    ON migration_import_items(project_id, job_id, status);
