-- Git links table for associating traces with git commits
CREATE TABLE IF NOT EXISTS agenttrace.git_links
(
    -- Identity
    id UUID,
    project_id UUID,
    trace_id FixedString(32),

    -- Git info
    commit_sha String,
    parent_sha String DEFAULT '',
    branch String DEFAULT '',
    tag String DEFAULT '',
    repo_url String DEFAULT '',

    -- Commit metadata
    commit_message String DEFAULT '',
    commit_author String DEFAULT '',
    commit_author_email String DEFAULT '',
    commit_timestamp DateTime64(3),

    -- File changes in this commit
    files_added Array(String) DEFAULT [],
    files_modified Array(String) DEFAULT [],
    files_deleted Array(String) DEFAULT [],
    files_changed_count UInt32 DEFAULT 0,
    additions UInt32 DEFAULT 0,
    deletions UInt32 DEFAULT 0,

    -- Link type
    link_type Enum8('current' = 0, 'start' = 1, 'end' = 2, 'referenced' = 3) DEFAULT 'current',

    -- CI context (if from CI)
    ci_run_id Nullable(UUID),

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(3),

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, commit_sha, trace_id, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.git_links ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.git_links ADD INDEX idx_commit_sha commit_sha TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.git_links ADD INDEX idx_branch branch TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.git_links ADD INDEX idx_repo_url repo_url TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.git_links ADD INDEX idx_link_type link_type TYPE set(4) GRANULARITY 1;
