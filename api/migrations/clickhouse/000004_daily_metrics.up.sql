-- Daily metrics materialized view for dashboard
CREATE MATERIALIZED VIEW IF NOT EXISTS agenttrace.daily_trace_metrics
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date)
AS SELECT
    project_id,
    toDate(start_time) AS date,
    count() AS trace_count,
    countIf(level = 'ERROR') AS error_count,
    sum(total_cost) AS total_cost,
    sum(input_cost) AS input_cost,
    sum(output_cost) AS output_cost,
    sum(total_tokens) AS total_tokens,
    sum(input_tokens) AS input_tokens,
    sum(output_tokens) AS output_tokens,
    avg(duration_ms) AS avg_duration_ms,
    quantile(0.5)(duration_ms) AS p50_duration_ms,
    quantile(0.95)(duration_ms) AS p95_duration_ms,
    quantile(0.99)(duration_ms) AS p99_duration_ms,
    uniqExact(user_id) AS unique_users,
    uniqExact(session_id) AS unique_sessions
FROM agenttrace.traces
GROUP BY project_id, date;

-- Daily model usage
CREATE MATERIALIZED VIEW IF NOT EXISTS agenttrace.daily_model_usage
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date, model)
AS SELECT
    project_id,
    toDate(start_time) AS date,
    model,
    count() AS generation_count,
    sum(total_cost) AS total_cost,
    sum(input_cost) AS input_cost,
    sum(output_cost) AS output_cost,
    sum(usage_total_tokens) AS total_tokens,
    sum(usage_input_tokens) AS input_tokens,
    sum(usage_output_tokens) AS output_tokens,
    sum(usage_cache_read_tokens) AS cache_read_tokens,
    avg(duration_ms) AS avg_duration_ms,
    quantile(0.5)(duration_ms) AS p50_duration_ms,
    quantile(0.95)(duration_ms) AS p95_duration_ms
FROM agenttrace.observations
WHERE type = 'GENERATION' AND model != ''
GROUP BY project_id, date, model;

-- Daily user metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS agenttrace.daily_user_metrics
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date, user_id)
AS SELECT
    project_id,
    toDate(start_time) AS date,
    user_id,
    count() AS trace_count,
    sum(total_cost) AS total_cost,
    sum(total_tokens) AS total_tokens,
    avg(duration_ms) AS avg_duration_ms
FROM agenttrace.traces
WHERE user_id != ''
GROUP BY project_id, date, user_id;

-- Hourly metrics for recent data
CREATE MATERIALIZED VIEW IF NOT EXISTS agenttrace.hourly_trace_metrics
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (project_id, hour)
TTL hour + INTERVAL 7 DAY
AS SELECT
    project_id,
    toStartOfHour(start_time) AS hour,
    count() AS trace_count,
    countIf(level = 'ERROR') AS error_count,
    sum(total_cost) AS total_cost,
    sum(total_tokens) AS total_tokens,
    avg(duration_ms) AS avg_duration_ms
FROM agenttrace.traces
GROUP BY project_id, hour;

-- Score aggregations
CREATE MATERIALIZED VIEW IF NOT EXISTS agenttrace.daily_score_metrics
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date, name)
AS SELECT
    project_id,
    toDate(created_at) AS date,
    name,
    source,
    count() AS score_count,
    avg(value) AS avg_value,
    min(value) AS min_value,
    max(value) AS max_value,
    quantile(0.5)(value) AS median_value
FROM agenttrace.scores
WHERE data_type = 'NUMERIC' AND value IS NOT NULL
GROUP BY project_id, date, name, source;
