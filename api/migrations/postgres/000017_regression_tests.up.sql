-- Regression tests table
CREATE TABLE IF NOT EXISTS regression_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    baseline_dataset_id UUID REFERENCES datasets(id) ON DELETE SET NULL,
    evaluator_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    thresholds JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_regression_test_status CHECK (status IN ('PENDING', 'RUNNING', 'PASSED', 'FAILED'))
);

-- Indexes for regression_tests
CREATE INDEX idx_regression_tests_project ON regression_tests(project_id);

-- Trigger for updated_at
CREATE TRIGGER update_regression_tests_updated_at
    BEFORE UPDATE ON regression_tests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Regression results table
CREATE TABLE IF NOT EXISTS regression_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES regression_tests(id) ON DELETE CASCADE,
    run_id UUID NOT NULL,
    scores JSONB,
    baseline_scores JSONB,
    passed BOOLEAN,
    deltas JSONB,
    failed_metrics JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for regression_results
CREATE INDEX idx_regression_results_test ON regression_results(test_id);
CREATE INDEX idx_regression_results_test_time ON regression_results(test_id, created_at DESC);
