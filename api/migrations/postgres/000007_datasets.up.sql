-- Dataset item status
CREATE TYPE dataset_item_status AS ENUM ('active', 'archived');

-- Datasets table
CREATE TABLE IF NOT EXISTS datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_datasets_project_id ON datasets(project_id);
CREATE INDEX idx_datasets_name ON datasets(name);

CREATE TRIGGER update_datasets_updated_at
    BEFORE UPDATE ON datasets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Dataset items table
CREATE TABLE IF NOT EXISTS dataset_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    input JSONB NOT NULL,
    expected_output JSONB,
    metadata JSONB NOT NULL DEFAULT '{}',
    source_trace_id VARCHAR(32),
    source_observation_id VARCHAR(16),
    status dataset_item_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dataset_items_dataset_id ON dataset_items(dataset_id);
CREATE INDEX idx_dataset_items_status ON dataset_items(status);
CREATE INDEX idx_dataset_items_source_trace ON dataset_items(source_trace_id);

CREATE TRIGGER update_dataset_items_updated_at
    BEFORE UPDATE ON dataset_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Dataset runs table
CREATE TABLE IF NOT EXISTS dataset_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dataset_runs_dataset_id ON dataset_runs(dataset_id);

CREATE TRIGGER update_dataset_runs_updated_at
    BEFORE UPDATE ON dataset_runs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Dataset run items (links items to traces generated during run)
CREATE TABLE IF NOT EXISTS dataset_run_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_run_id UUID NOT NULL REFERENCES dataset_runs(id) ON DELETE CASCADE,
    dataset_item_id UUID NOT NULL REFERENCES dataset_items(id) ON DELETE CASCADE,
    trace_id VARCHAR(32) NOT NULL,
    observation_id VARCHAR(16),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(dataset_run_id, dataset_item_id)
);

CREATE INDEX idx_dataset_run_items_run_id ON dataset_run_items(dataset_run_id);
CREATE INDEX idx_dataset_run_items_item_id ON dataset_run_items(dataset_item_id);
CREATE INDEX idx_dataset_run_items_trace_id ON dataset_run_items(trace_id);
