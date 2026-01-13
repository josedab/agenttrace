-- Experiments table (A/B testing)
CREATE TABLE IF NOT EXISTS experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    target_metric VARCHAR(100) NOT NULL,
    target_goal VARCHAR(20) NOT NULL DEFAULT 'minimize',
    traffic_percent DECIMAL(5, 2) NOT NULL DEFAULT 100.0,
    trace_name_filter VARCHAR(255),
    user_id_filter JSONB DEFAULT '[]'::jsonb,
    metadata_filters JSONB DEFAULT '{}'::jsonb,
    min_duration_hours INTEGER,
    min_samples_per_variant INTEGER,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    winning_variant_id UUID,
    results JSONB,
    statistical_power DECIMAL(5, 4),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),

    CONSTRAINT valid_experiment_status CHECK (status IN ('draft', 'running', 'paused', 'completed', 'archived')),
    CONSTRAINT valid_target_goal CHECK (target_goal IN ('minimize', 'maximize')),
    CONSTRAINT valid_traffic_percent CHECK (traffic_percent >= 0 AND traffic_percent <= 100),
    CONSTRAINT unique_experiment_name_per_project UNIQUE (project_id, name)
);

-- Indexes for experiments
CREATE INDEX idx_experiments_project ON experiments(project_id);
CREATE INDEX idx_experiments_status ON experiments(project_id, status);
CREATE INDEX idx_experiments_created ON experiments(project_id, created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_experiments_updated_at
    BEFORE UPDATE ON experiments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Experiment Variants table
CREATE TABLE IF NOT EXISTS experiment_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    weight DECIMAL(5, 2) NOT NULL DEFAULT 50.0,
    is_control BOOLEAN NOT NULL DEFAULT false,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    sample_count INTEGER NOT NULL DEFAULT 0,
    metric_mean DECIMAL(20, 6),
    metric_std_dev DECIMAL(20, 6),
    metric_min DECIMAL(20, 6),
    metric_max DECIMAL(20, 6),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_variant_weight CHECK (weight >= 0 AND weight <= 100),
    CONSTRAINT unique_variant_name_per_experiment UNIQUE (experiment_id, name)
);

-- Indexes for experiment_variants
CREATE INDEX idx_experiment_variants_experiment ON experiment_variants(experiment_id);
CREATE INDEX idx_experiment_variants_control ON experiment_variants(experiment_id, is_control) WHERE is_control = true;

-- Trigger for updated_at
CREATE TRIGGER update_experiment_variants_updated_at
    BEFORE UPDATE ON experiment_variants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key constraint for winning_variant_id
ALTER TABLE experiments
    ADD CONSTRAINT fk_experiments_winning_variant
    FOREIGN KEY (winning_variant_id) REFERENCES experiment_variants(id) ON DELETE SET NULL;

-- Experiment Assignments table (which variant a trace is assigned to)
CREATE TABLE IF NOT EXISTS experiment_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL REFERENCES experiment_variants(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    user_id VARCHAR(255),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    variant_config JSONB NOT NULL DEFAULT '{}'::jsonb,

    CONSTRAINT unique_assignment_per_trace UNIQUE (experiment_id, trace_id)
);

-- Indexes for experiment_assignments
CREATE INDEX idx_experiment_assignments_experiment ON experiment_assignments(experiment_id);
CREATE INDEX idx_experiment_assignments_variant ON experiment_assignments(variant_id);
CREATE INDEX idx_experiment_assignments_trace ON experiment_assignments(trace_id);
CREATE INDEX idx_experiment_assignments_user ON experiment_assignments(experiment_id, user_id) WHERE user_id IS NOT NULL;

-- Experiment Metrics table (recorded metric values for analysis)
CREATE TABLE IF NOT EXISTS experiment_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL REFERENCES experiment_variants(id) ON DELETE CASCADE,
    trace_id UUID NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(20, 6) NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_metric_per_trace UNIQUE (experiment_id, trace_id, metric_name)
);

-- Indexes for experiment_metrics
CREATE INDEX idx_experiment_metrics_experiment ON experiment_metrics(experiment_id);
CREATE INDEX idx_experiment_metrics_variant ON experiment_metrics(variant_id);
CREATE INDEX idx_experiment_metrics_analysis ON experiment_metrics(experiment_id, variant_id, metric_name);
