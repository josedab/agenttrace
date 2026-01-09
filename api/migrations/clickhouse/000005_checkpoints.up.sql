-- Checkpoints table for code state snapshots
CREATE TABLE IF NOT EXISTS agenttrace.checkpoints
(
    -- Identity
    id UUID,
    project_id UUID,
    trace_id FixedString(32),
    observation_id Nullable(FixedString(16)),

    -- Checkpoint info
    name String DEFAULT '',
    description String DEFAULT '',
    checkpoint_type Enum8('manual' = 0, 'auto' = 1, 'pre_edit' = 2, 'post_edit' = 3, 'rollback' = 4) DEFAULT 'manual',

    -- Git context at checkpoint time
    git_commit_sha String DEFAULT '',
    git_branch String DEFAULT '',
    git_repo_url String DEFAULT '',

    -- File state
    files_snapshot String DEFAULT '{}', -- JSON: {path: {content_hash, size, mode}}
    files_changed Array(String) DEFAULT [],
    storage_path String DEFAULT '', -- Path in object storage for full snapshot

    -- Metrics at checkpoint
    total_files UInt32 DEFAULT 0,
    total_size_bytes UInt64 DEFAULT 0,

    -- Restoration info
    restored_from Nullable(UUID), -- If this checkpoint was restored from another
    restored_at Nullable(DateTime64(3)),

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(3),

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, trace_id, created_at, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.checkpoints ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.checkpoints ADD INDEX idx_observation_id observation_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.checkpoints ADD INDEX idx_git_commit git_commit_sha TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.checkpoints ADD INDEX idx_type checkpoint_type TYPE set(5) GRANULARITY 1;
