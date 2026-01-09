-- Scores table
CREATE TABLE IF NOT EXISTS agenttrace.scores
(
    -- Identity
    id UUID,
    project_id UUID,
    trace_id FixedString(32),
    observation_id Nullable(FixedString(16)),

    -- Score definition
    name String,
    source Enum8('API' = 0, 'EVAL' = 1, 'ANNOTATION' = 2) DEFAULT 'API',
    data_type Enum8('NUMERIC' = 0, 'BOOLEAN' = 1, 'CATEGORICAL' = 2) DEFAULT 'NUMERIC',

    -- Score value (use appropriate field based on data_type)
    value Nullable(Float64),
    string_value Nullable(String),

    -- Metadata
    comment String DEFAULT '',
    config_id Nullable(UUID),

    -- User info
    author_user_id Nullable(UUID),

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3),

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, trace_id, name, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.scores ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.scores ADD INDEX idx_observation_id observation_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.scores ADD INDEX idx_name name TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.scores ADD INDEX idx_source source TYPE set(3) GRANULARITY 1;
ALTER TABLE agenttrace.scores ADD INDEX idx_config_id config_id TYPE bloom_filter GRANULARITY 1;
