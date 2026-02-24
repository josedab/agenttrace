-- Rollback Migration 000029: V5 Features

DROP TABLE IF EXISTS copilot_insights;
DROP TABLE IF EXISTS copilot_queries;
DROP TABLE IF EXISTS federation_configs;
DROP TABLE IF EXISTS federated_insights;
DROP TABLE IF EXISTS federation_rings;
DROP TABLE IF EXISTS pattern_deployments;
DROP TABLE IF EXISTS collab_patterns;
DROP TABLE IF EXISTS trace_attachments;
DROP TABLE IF EXISTS compliance_scores;
DROP TABLE IF EXISTS compliance_rules;
DROP TABLE IF EXISTS compliance_policies;
DROP TABLE IF EXISTS kg_edges;
DROP TABLE IF EXISTS kg_nodes;
DROP TABLE IF EXISTS cost_attributions;
DROP TABLE IF EXISTS intent_declarations;
DROP TABLE IF EXISTS cross_org_submissions;
DROP TABLE IF EXISTS trust_history;
DROP TABLE IF EXISTS autonomy_configs;
