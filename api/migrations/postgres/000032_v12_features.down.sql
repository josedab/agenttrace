DROP TABLE IF EXISTS privacy_budgets;
DROP TABLE IF EXISTS federated_trace_insights;
DROP TABLE IF EXISTS guardrail_audit_trail;
DROP TABLE IF EXISTS self_healing_policies;
DROP TABLE IF EXISTS ab_test_results;
DROP TABLE IF EXISTS ab_tests;
DROP TABLE IF EXISTS notification_integrations;
DROP TABLE IF EXISTS review_comments;
DROP TABLE IF EXISTS trace_reviews;
DROP TABLE IF EXISTS delegation_chains;
DROP TABLE IF EXISTS cost_autopilot_rules;
DROP TABLE IF EXISTS cost_hotspots;
DROP TABLE IF EXISTS adapter_events;
DROP TABLE IF EXISTS agent_adapters;
DROP TABLE IF EXISTS prompt_ci_gate_configs;
DROP TABLE IF EXISTS prompt_ci_baselines;
DROP TABLE IF EXISTS prompt_regression_history;
DROP TABLE IF EXISTS rca_investigations;
DROP TABLE IF EXISTS correlation_rules;
DROP TABLE IF EXISTS alert_delivery_channels;
DROP TABLE IF EXISTS correlated_anomalies;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM prompt_ci_runs
        WHERE length(branch) > 200 OR length(commit_sha) > 40
    ) THEN
        RAISE EXCEPTION
            'cannot roll back v12: prompt_ci_runs contains branch or commit values exceeding v11 limits';
    END IF;

    ALTER TABLE prompt_ci_runs
        ALTER COLUMN branch TYPE VARCHAR(200),
        ALTER COLUMN commit_sha TYPE VARCHAR(40);
END
$$;
