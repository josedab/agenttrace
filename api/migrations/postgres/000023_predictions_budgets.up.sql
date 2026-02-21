-- Health predictions table
CREATE TABLE IF NOT EXISTS health_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    metric_name VARCHAR(50) NOT NULL,
    current_value DECIMAL(20,6),
    predicted_value DECIMAL(20,6),
    trend_direction VARCHAR(20) NOT NULL,
    confidence_level DECIMAL(5,4),
    time_horizon_hours INTEGER NOT NULL,
    alert_level VARCHAR(20) NOT NULL,
    root_cause JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_prediction_trend CHECK (trend_direction IN ('improving', 'stable', 'degrading')),
    CONSTRAINT valid_prediction_alert CHECK (alert_level IN ('none', 'warning', 'critical'))
);

-- Indexes for health_predictions
CREATE INDEX idx_health_predictions_project ON health_predictions(project_id);
CREATE INDEX idx_health_predictions_project_metric ON health_predictions(project_id, metric_name);
CREATE INDEX idx_health_predictions_project_time ON health_predictions(project_id, created_at DESC);

-- Cost budgets table
CREATE TABLE IF NOT EXISTS cost_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    monthly_limit_cents BIGINT NOT NULL,
    alert_thresholds JSONB NOT NULL DEFAULT '[]'::jsonb,
    auto_action VARCHAR(20) NOT NULL DEFAULT 'NONE',
    current_spend_cents BIGINT NOT NULL DEFAULT 0,
    forecasted_spend_cents BIGINT NOT NULL DEFAULT 0,
    forecast_exhaustion_date TIMESTAMP WITH TIME ZONE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_budget_auto_action CHECK (auto_action IN ('NONE', 'ALERT_ONLY', 'THROTTLE', 'SWITCH_MODEL', 'BLOCK'))
);

-- Indexes for cost_budgets
CREATE INDEX idx_cost_budgets_project ON cost_budgets(project_id);
CREATE INDEX idx_cost_budgets_project_enabled ON cost_budgets(project_id, enabled) WHERE enabled = true;

-- Trigger for updated_at
CREATE TRIGGER update_cost_budgets_updated_at
    BEFORE UPDATE ON cost_budgets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
