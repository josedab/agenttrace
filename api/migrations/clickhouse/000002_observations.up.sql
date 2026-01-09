-- Observations table (spans, generations, events)
CREATE TABLE IF NOT EXISTS agenttrace.observations
(
    -- Identity
    id FixedString(16),
    trace_id FixedString(32),
    project_id UUID,
    parent_observation_id Nullable(FixedString(16)),

    -- Type
    type Enum8('SPAN' = 0, 'GENERATION' = 1, 'EVENT' = 2) DEFAULT 'SPAN',

    -- Metadata
    name String DEFAULT '',
    level Enum8('DEBUG' = 0, 'DEFAULT' = 1, 'WARNING' = 2, 'ERROR' = 3) DEFAULT 'DEFAULT',
    status_message String DEFAULT '',
    metadata String DEFAULT '{}',

    -- Timing
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),
    completion_start_time Nullable(DateTime64(3)),
    duration_ms Float64 MATERIALIZED if(isNotNull(end_time), dateDiff('millisecond', start_time, end_time), 0),
    time_to_first_token_ms Float64 MATERIALIZED if(isNotNull(completion_start_time), dateDiff('millisecond', start_time, completion_start_time), 0),

    -- Input/Output
    input String DEFAULT '',
    output String DEFAULT '',

    -- Generation-specific fields
    model String DEFAULT '',
    model_parameters String DEFAULT '{}',

    -- Token usage
    usage_input_tokens UInt64 DEFAULT 0,
    usage_output_tokens UInt64 DEFAULT 0,
    usage_total_tokens UInt64 DEFAULT 0,
    usage_cache_read_tokens UInt64 DEFAULT 0,
    usage_cache_creation_tokens UInt64 DEFAULT 0,

    -- Token usage as materialized columns for convenience
    input_tokens UInt64 MATERIALIZED usage_input_tokens,
    output_tokens UInt64 MATERIALIZED usage_output_tokens,
    total_tokens UInt64 MATERIALIZED usage_total_tokens,

    -- Costs (calculated by worker)
    input_cost Decimal64(10) DEFAULT 0,
    output_cost Decimal64(10) DEFAULT 0,
    total_cost Decimal64(10) DEFAULT 0,

    -- Prompt tracking
    prompt_id Nullable(UUID),
    prompt_version Nullable(UInt32),
    prompt_name Nullable(String),

    -- Version for versioning
    version String DEFAULT '',

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3),

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, trace_id, start_time, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.observations ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.observations ADD INDEX idx_parent_id parent_observation_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.observations ADD INDEX idx_type type TYPE set(3) GRANULARITY 1;
ALTER TABLE agenttrace.observations ADD INDEX idx_name name TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.observations ADD INDEX idx_model model TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.observations ADD INDEX idx_level level TYPE set(4) GRANULARITY 1;
ALTER TABLE agenttrace.observations ADD INDEX idx_prompt_id prompt_id TYPE bloom_filter GRANULARITY 1;
