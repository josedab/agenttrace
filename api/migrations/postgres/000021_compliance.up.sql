-- Compliance records table
CREATE TABLE IF NOT EXISTS compliance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    risk_level VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    assessment_date TIMESTAMP WITH TIME ZONE NOT NULL,
    next_review_date TIMESTAMP WITH TIME ZONE NOT NULL,
    findings JSONB NOT NULL DEFAULT '[]'::jsonb,
    auditor_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_compliance_risk_level CHECK (risk_level IN ('HIGH_RISK', 'LIMITED_RISK', 'MINIMAL_RISK')),
    CONSTRAINT valid_compliance_status CHECK (status IN ('COMPLIANT', 'NON_COMPLIANT', 'UNDER_REVIEW', 'NOT_ASSESSED'))
);

-- Indexes for compliance_records
CREATE INDEX idx_compliance_records_project ON compliance_records(project_id);

-- Trigger for updated_at
CREATE TRIGGER update_compliance_records_updated_at
    BEFORE UPDATE ON compliance_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Immutable audit entries table (hash-chained tamper-evident log)
CREATE TABLE IF NOT EXISTS immutable_audit_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(255),
    entry_type VARCHAR(50) NOT NULL,
    actor VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    details TEXT,
    previous_hash VARCHAR(64),
    hash VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Indexes for immutable_audit_entries
CREATE INDEX idx_immutable_audit_entries_project ON immutable_audit_entries(project_id);
CREATE INDEX idx_immutable_audit_entries_timestamp ON immutable_audit_entries(project_id, timestamp DESC);
CREATE UNIQUE INDEX idx_immutable_audit_entries_hash_chain ON immutable_audit_entries(project_id, previous_hash) WHERE previous_hash IS NOT NULL;

-- Conformity assessments table (EU AI Act)
CREATE TABLE IF NOT EXISTS conformity_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    system_name VARCHAR(200) NOT NULL,
    system_description TEXT,
    risk_level VARCHAR(20) NOT NULL,
    provider VARCHAR(200),
    deployment_date TIMESTAMP WITH TIME ZONE,
    transparency_score DECIMAL(5,4),
    fairness_score DECIMAL(5,4),
    robustness_score DECIMAL(5,4),
    human_oversight_level VARCHAR(50),
    data_governance_notes TEXT,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_conformity_risk_level CHECK (risk_level IN ('HIGH_RISK', 'LIMITED_RISK', 'MINIMAL_RISK')),
    CONSTRAINT valid_conformity_status CHECK (status IN ('COMPLIANT', 'NON_COMPLIANT', 'UNDER_REVIEW', 'NOT_ASSESSED'))
);

-- Indexes for conformity_assessments
CREATE INDEX idx_conformity_assessments_project ON conformity_assessments(project_id);

-- Trigger for updated_at
CREATE TRIGGER update_conformity_assessments_updated_at
    BEFORE UPDATE ON conformity_assessments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
