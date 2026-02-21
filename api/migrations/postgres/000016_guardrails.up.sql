-- Guard rules table
CREATE TABLE IF NOT EXISTS guard_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(30) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    action VARCHAR(10) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_guard_rule_type CHECK (type IN ('cost_limit', 'latency_limit', 'file_restriction', 'pattern_block', 'custom')),
    CONSTRAINT valid_guard_rule_action CHECK (action IN ('block', 'alert', 'log'))
);

-- Indexes for guard_rules
CREATE INDEX idx_guard_rules_project ON guard_rules(project_id);

-- Trigger for updated_at
CREATE TRIGGER update_guard_rules_updated_at
    BEFORE UPDATE ON guard_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Guard violations table
CREATE TABLE IF NOT EXISTS guard_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES guard_rules(id) ON DELETE CASCADE,
    trace_id VARCHAR(255) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    details TEXT,
    action_taken VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_violation_severity CHECK (severity IN ('critical', 'warning', 'info'))
);

-- Indexes for guard_violations
CREATE INDEX idx_guard_violations_project ON guard_violations(project_id);
CREATE INDEX idx_guard_violations_project_time ON guard_violations(project_id, created_at DESC);
CREATE INDEX idx_guard_violations_rule ON guard_violations(rule_id);
