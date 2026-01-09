DROP TABLE IF EXISTS dataset_run_items;
DROP TRIGGER IF EXISTS update_dataset_runs_updated_at ON dataset_runs;
DROP TABLE IF EXISTS dataset_runs;
DROP TRIGGER IF EXISTS update_dataset_items_updated_at ON dataset_items;
DROP TABLE IF EXISTS dataset_items;
DROP TRIGGER IF EXISTS update_datasets_updated_at ON datasets;
DROP TABLE IF EXISTS datasets;
DROP TYPE IF EXISTS dataset_item_status;
