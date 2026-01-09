-- Create database if not exists
CREATE DATABASE IF NOT EXISTS agenttrace;

-- Traces table with ReplacingMergeTree for deduplication
CREATE TABLE IF NOT EXISTS agenttrace.traces
(
    -- Identity
    id FixedString(32),
    project_id UUID,

    -- Metadata
    name String DEFAULT '',
    user_id String DEFAULT '',
    session_id String DEFAULT '',
    release String DEFAULT '',
    version String DEFAULT '',
    tags Array(String) DEFAULT [],
    metadata String DEFAULT '{}',

    -- Public flag
    public Bool DEFAULT false,

    -- Bookmarking
    bookmarked Bool DEFAULT false,

    -- Timing
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),
    duration_ms Float64 MATERIALIZED if(isNotNull(end_time), dateDiff('millisecond', start_time, end_time), 0),

    -- Input/Output
    input String DEFAULT '',
    output String DEFAULT '',

    -- Status
    level Enum8('DEBUG' = 0, 'DEFAULT' = 1, 'WARNING' = 2, 'ERROR' = 3) DEFAULT 'DEFAULT',
    status_message String DEFAULT '',

    -- Aggregated costs (updated by worker)
    total_cost Decimal64(10) DEFAULT 0,
    input_cost Decimal64(10) DEFAULT 0,
    output_cost Decimal64(10) DEFAULT 0,

    -- Aggregated tokens
    total_tokens UInt64 DEFAULT 0,
    input_tokens UInt64 DEFAULT 0,
    output_tokens UInt64 DEFAULT 0,

    -- Git integration
    git_commit_sha String DEFAULT '',
    git_branch String DEFAULT '',
    git_repo_url String DEFAULT '',

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3),

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, start_time, id)
SETTINGS index_granularity = 8192;

-- Indexes for common query patterns
ALTER TABLE agenttrace.traces ADD INDEX idx_user_id user_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.traces ADD INDEX idx_session_id session_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.traces ADD INDEX idx_name name TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.traces ADD INDEX idx_tags tags TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.traces ADD INDEX idx_release release TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.traces ADD INDEX idx_level level TYPE set(4) GRANULARITY 1;

-- Sessions materialized view (aggregates traces by session)
CREATE TABLE IF NOT EXISTS agenttrace.sessions
(
    id String,
    project_id UUID,
    user_id String,
    bookmarked Bool DEFAULT false,
    public Bool DEFAULT false,
    created_at DateTime64(3),
    updated_at DateTime64(3),
    trace_count UInt64,
    total_cost Decimal64(10),
    total_tokens UInt64,
    first_trace_time DateTime64(3),
    last_trace_time DateTime64(3),
    _version UInt64
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(first_trace_time)
ORDER BY (project_id, id)
SETTINGS index_granularity = 8192;
