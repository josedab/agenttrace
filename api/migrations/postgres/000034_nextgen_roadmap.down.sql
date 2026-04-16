DROP TABLE IF EXISTS migration_import_items;
DROP TABLE IF EXISTS share_links;
DROP VIEW IF EXISTS eval_hub_package_latest;
DROP TABLE IF EXISTS eval_hub_runs;
DROP TABLE IF EXISTS eval_hub_versions;
DROP TABLE IF EXISTS eval_hub_packages;

ALTER TABLE replay_events
    DROP CONSTRAINT IF EXISTS replay_events_event_type_check;
ALTER TABLE replay_events
    ADD CONSTRAINT replay_events_event_type_check
    CHECK (event_type IN (
        'llm_call',
        'file_operation',
        'terminal_command',
        'checkpoint',
        'state_change',
        'agent_message',
        'tool_call',
        'error'
    ));

DROP TABLE IF EXISTS replay_plans;
