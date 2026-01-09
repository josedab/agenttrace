-- CI runs table for tracking CI/CD pipeline executions
CREATE TABLE IF NOT EXISTS agenttrace.ci_runs
(
    -- Identity
    id UUID,
    project_id UUID,

    -- CI provider info
    provider Enum8('github_actions' = 0, 'gitlab_ci' = 1, 'jenkins' = 2, 'circleci' = 3, 'azure_devops' = 4, 'bitbucket' = 5, 'other' = 6) DEFAULT 'other',
    provider_run_id String, -- Provider's native run ID
    provider_run_url String DEFAULT '',

    -- Pipeline info
    pipeline_name String DEFAULT '',
    job_name String DEFAULT '',
    workflow_name String DEFAULT '',

    -- Git context
    git_commit_sha String DEFAULT '',
    git_branch String DEFAULT '',
    git_tag String DEFAULT '',
    git_repo_url String DEFAULT '',
    git_ref String DEFAULT '',

    -- Pull request context (if applicable)
    pr_number UInt32 DEFAULT 0,
    pr_title String DEFAULT '',
    pr_source_branch String DEFAULT '',
    pr_target_branch String DEFAULT '',

    -- Execution
    started_at DateTime64(3) DEFAULT now64(3),
    completed_at Nullable(DateTime64(3)),
    duration_ms UInt32 DEFAULT 0,

    -- Status
    status Enum8('pending' = 0, 'running' = 1, 'success' = 2, 'failure' = 3, 'cancelled' = 4, 'skipped' = 5) DEFAULT 'pending',
    conclusion String DEFAULT '',
    error_message String DEFAULT '',

    -- Associated traces
    trace_ids Array(FixedString(32)) DEFAULT [],
    trace_count UInt32 DEFAULT 0,

    -- Aggregated metrics from traces
    total_cost Decimal64(10) DEFAULT 0,
    total_tokens UInt64 DEFAULT 0,
    total_observations UInt64 DEFAULT 0,

    -- Runner info
    runner_name String DEFAULT '',
    runner_os String DEFAULT '',
    runner_arch String DEFAULT '',

    -- Trigger info
    triggered_by String DEFAULT '', -- User or event that triggered
    trigger_event String DEFAULT '', -- push, pull_request, schedule, etc.

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3),

    -- Version for ReplacingMergeTree
    _version UInt64 DEFAULT toUInt64(now64(3))
)
ENGINE = ReplacingMergeTree(_version)
PARTITION BY toYYYYMM(started_at)
ORDER BY (project_id, started_at, id)
SETTINGS index_granularity = 8192;

-- Indexes
ALTER TABLE agenttrace.ci_runs ADD INDEX idx_provider_run_id provider_run_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.ci_runs ADD INDEX idx_git_commit git_commit_sha TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.ci_runs ADD INDEX idx_git_branch git_branch TYPE bloom_filter GRANULARITY 1;
ALTER TABLE agenttrace.ci_runs ADD INDEX idx_status status TYPE set(6) GRANULARITY 1;
ALTER TABLE agenttrace.ci_runs ADD INDEX idx_provider provider TYPE set(7) GRANULARITY 1;
