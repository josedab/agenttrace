# ADR-002: Dual Database Strategy - PostgreSQL + ClickHouse

## Status

Accepted

## Context

AgentTrace stores two fundamentally different types of data:

1. **Transactional data**: Users, organizations, projects, API keys, prompts, evaluator configurations
   - Requires ACID guarantees
   - Complex relationships (foreign keys, constraints)
   - Moderate volume, frequent updates
   - Query patterns: point lookups, joins, transactions

2. **Analytical data**: Traces, observations, scores, metrics
   - Append-heavy workload (rarely updated)
   - High volume (millions to billions of rows)
   - Query patterns: aggregations, time-series analysis, filtering
   - Requires fast analytical queries across large datasets

A single database cannot optimally serve both workloads. Traditional RDBMS struggle with analytical queries at scale, while analytical databases lack transactional guarantees.

### Alternatives Considered

1. **PostgreSQL only**
   - Pros: Single database, simpler operations, ACID everywhere
   - Cons: Poor performance for analytical queries at scale, expensive partitioning

2. **ClickHouse only**
   - Pros: Excellent analytical performance
   - Cons: No ACID transactions, poor for frequent updates, no foreign keys

3. **PostgreSQL + TimescaleDB**
   - Pros: Single PostgreSQL ecosystem
   - Cons: Less performant than ClickHouse for pure analytics, licensing concerns

4. **PostgreSQL + ClickHouse** (chosen)
   - Pros: Best of both worlds, proven at scale (Langfuse, PostHog, Plausible)
   - Cons: Operational complexity, data synchronization challenges

## Decision

We use a **dual database architecture**:

- **PostgreSQL**: Transactional data (users, projects, prompts, API keys, evaluators)
- **ClickHouse**: Analytical data (traces, observations, scores, file operations, terminal commands)

Data flow:
1. Traces ingested directly to ClickHouse for immediate availability
2. Metadata references (project_id, user_id) validated against PostgreSQL
3. Materialized views in ClickHouse pre-aggregate common metrics
4. PostgreSQL stores configuration that ClickHouse queries reference

### Schema Design Principles

**PostgreSQL tables:**
- `users`, `organizations`, `projects`, `api_keys`
- `prompts`, `prompt_versions`, `evaluators`
- `datasets`, `dataset_items`, `experiments`

**ClickHouse tables:**
- `traces` (ReplacingMergeTree for updates)
- `observations` (spans, generations, events)
- `scores` (evaluation results)
- `file_operations`, `terminal_commands`, `git_links`
- `daily_costs_mv` (materialized view for cost aggregation)

## Consequences

### Positive

- **Query performance**: ClickHouse handles analytical queries 10-100x faster than PostgreSQL at scale
- **Cost efficiency**: ClickHouse columnar compression reduces storage costs by 5-10x
- **Scalability**: Each database can scale independently based on workload
- **Proven pattern**: Used successfully by Langfuse, PostHog, Plausible Analytics
- **Real-time analytics**: ClickHouse materialized views enable instant dashboard updates

### Negative

- **Operational complexity**: Two databases to manage, backup, and monitor
- **Data consistency**: No cross-database transactions; eventual consistency between systems
- **Schema synchronization**: Changes may require coordinated migrations
- **Learning curve**: Team must understand both PostgreSQL and ClickHouse patterns
- **Join limitations**: Cannot join across databases; must denormalize or make multiple queries

### Neutral

- ClickHouse requires understanding of MergeTree engine variants
- Connection pooling strategies differ between databases
- Backup and disaster recovery procedures are database-specific

## Data Synchronization Strategy

1. **Reference data**: Project/user IDs stored in both databases (ClickHouse references PostgreSQL as source of truth)
2. **Validation**: Trace ingestion validates project_id exists in PostgreSQL before ClickHouse insert
3. **Denormalization**: Common lookups (project name, user email) denormalized into ClickHouse for query performance
4. **Retention**: ClickHouse retention policies independent of PostgreSQL

## References

- [ClickHouse MergeTree Engines](https://clickhouse.com/docs/en/engines/table-engines/mergetree-family)
- [Langfuse Database Architecture](https://langfuse.com/docs/deployment/self-host)
- [PostHog ClickHouse Usage](https://posthog.com/blog/how-we-turned-clickhouse-into-our-eventmansion)
