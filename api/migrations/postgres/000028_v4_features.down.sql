-- Rollback Migration 000028: V4 Features

DROP TABLE IF EXISTS slo_history;
DROP TABLE IF EXISTS slos;
DROP TABLE IF EXISTS synthetic_items;
DROP TABLE IF EXISTS synthetic_datasets;
DROP TABLE IF EXISTS carbon_configs;
DROP TABLE IF EXISTS annotation_replies;
DROP TABLE IF EXISTS trace_annotations;
DROP TABLE IF EXISTS handoffs;
DROP TABLE IF EXISTS metric_alerts;
DROP TABLE IF EXISTS metric_dashboards;
DROP TABLE IF EXISTS metric_values;
DROP TABLE IF EXISTS custom_metrics;
DROP TABLE IF EXISTS chaos_results;
DROP TABLE IF EXISTS chaos_experiments;
DROP TABLE IF EXISTS prompt_cache_entries;
DROP TABLE IF EXISTS prompt_cache_configs;
DROP TABLE IF EXISTS service_map_nodes;
DROP TABLE IF EXISTS distributed_spans;
DROP TABLE IF EXISTS memory_snapshots;
