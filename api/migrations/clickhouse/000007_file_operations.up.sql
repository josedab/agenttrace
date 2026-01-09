-- File operations table for tracking file changes during agent execution
CREATE TABLE IF NOT EXISTS agenttrace.file_operations
(
    -- Identity
    id UUID,
    project_id UUID,
    trace_id FixedString(32),
    observation_id Nullable(FixedString(16)),

    -- Operation details
    operation Enum8('create' = 0, 'read' = 1, 'update' = 2, 'delete' = 3, 'rename' = 4, 'move' = 5, 'copy' = 6) DEFAULT 'read',
    file_path String,
    new_path String DEFAULT '', -- For rename/move/copy operations

    -- File info
    file_size UInt64 DEFAULT 0,
    file_mode String DEFAULT '',
    content_hash String DEFAULT '', -- SHA256 of content
    mime_type String DEFAULT '',

    -- Change details (for update operations)
    lines_added UInt32 DEFAULT 0,
    lines_removed UInt32 DEFAULT 0,
    diff_preview String DEFAULT '', -- First N lines of diff

    -- Content snapshots (optional)
    content_before_hash String DEFAULT '',
    content_after_hash String DEFAULT '',

    -- Context
    tool_name String DEFAULT '', -- e.g., "editor", "file_writer"
    reason String DEFAULT '', -- Why this operation was performed

    -- Timing
    started_at DateTime64(3) DEFAULT now64(3),
    completed_at Nullable(DateTime64(3)),
    duration_ms UInt32 DEFAULT 0,

    -- Status
    success Bool DEFAULT true,
    error_message String DEFAULT '',

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(started_at)
ORDER BY (project_id, trace_id, started_at, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.file_operations ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.file_operations ADD INDEX idx_observation_id observation_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.file_operations ADD INDEX idx_file_path file_path TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.file_operations ADD INDEX idx_operation operation TYPE set(7) GRANULARITY 1;
