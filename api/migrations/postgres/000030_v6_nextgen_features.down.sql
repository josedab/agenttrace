-- Rollback Next-Gen Features Migration (v6)

DROP TABLE IF EXISTS federated_metrics;
DROP TABLE IF EXISTS federated_instances;
DROP TABLE IF EXISTS file_trace_mappings;
DROP TABLE IF EXISTS agent_knowledge_snapshots;
DROP TABLE IF EXISTS trace_clusters;
DROP TABLE IF EXISTS trace_embeddings;
DROP TABLE IF EXISTS agent_benchmark_runs;
DROP TABLE IF EXISTS agent_benchmark_suites;
DROP TABLE IF EXISTS prompt_ci_runs;
DROP TABLE IF EXISTS prompt_baselines;
DROP TABLE IF EXISTS multi_agent_sessions;
DROP TABLE IF EXISTS cost_guardrail_violations;
DROP TABLE IF EXISTS cost_guardrail_policies;
DROP TABLE IF EXISTS replay_events;
DROP TABLE IF EXISTS replay_sessions;
