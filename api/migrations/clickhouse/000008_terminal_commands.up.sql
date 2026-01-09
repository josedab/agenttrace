-- Terminal commands table for logging shell commands executed by agents
CREATE TABLE IF NOT EXISTS agenttrace.terminal_commands
(
    -- Identity
    id UUID,
    project_id UUID,
    trace_id FixedString(32),
    observation_id Nullable(FixedString(16)),

    -- Command details
    command String,
    args Array(String) DEFAULT [],
    working_directory String DEFAULT '',
    shell String DEFAULT 'bash',

    -- Environment (filtered for safety)
    env_vars String DEFAULT '{}', -- JSON of safe env vars

    -- Execution
    started_at DateTime64(3) DEFAULT now64(3),
    completed_at Nullable(DateTime64(3)),
    duration_ms UInt32 DEFAULT 0,

    -- Result
    exit_code Int32 DEFAULT 0,
    stdout String DEFAULT '',
    stderr String DEFAULT '',
    stdout_truncated Bool DEFAULT false,
    stderr_truncated Bool DEFAULT false,

    -- Status
    success Bool DEFAULT true,
    timed_out Bool DEFAULT false,
    killed Bool DEFAULT false,

    -- Resource usage
    max_memory_bytes UInt64 DEFAULT 0,
    cpu_time_ms UInt32 DEFAULT 0,

    -- Context
    tool_name String DEFAULT '', -- e.g., "bash_executor"
    reason String DEFAULT '', -- Why this command was run

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(started_at)
ORDER BY (project_id, trace_id, started_at, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.terminal_commands ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.terminal_commands ADD INDEX idx_observation_id observation_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.terminal_commands ADD INDEX idx_command command TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.terminal_commands ADD INDEX idx_exit_code exit_code TYPE set(256) GRANULARITY 1;
