-- Migration jobs table
CREATE TABLE IF NOT EXISTS migration_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    config JSONB,
    progress JSONB NOT NULL DEFAULT '{}'::jsonb,
    errors JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_migration_job_status CHECK (status IN ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED'))
);

-- Indexes for migration_jobs
CREATE INDEX idx_migration_jobs_project ON migration_jobs(project_id);
CREATE INDEX idx_migration_jobs_project_status ON migration_jobs(project_id, status);

-- Debug sessions table
CREATE TABLE IF NOT EXISTS debug_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    current_step INTEGER NOT NULL DEFAULT 0,
    total_steps INTEGER NOT NULL DEFAULT 0,
    breakpoints JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_debug_session_status CHECK (status IN ('ACTIVE', 'PAUSED', 'COMPLETE'))
);

-- Indexes for debug_sessions
CREATE INDEX idx_debug_sessions_project ON debug_sessions(project_id);
CREATE INDEX idx_debug_sessions_trace ON debug_sessions(trace_id);

-- Trigger for updated_at
CREATE TRIGGER update_debug_sessions_updated_at
    BEFORE UPDATE ON debug_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Debug annotations table
CREATE TABLE IF NOT EXISTS debug_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES debug_sessions(id) ON DELETE CASCADE,
    event_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for debug_annotations
CREATE INDEX idx_debug_annotations_session ON debug_annotations(session_id);

-- Cost recommendations table
CREATE TABLE IF NOT EXISTS cost_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    current_model VARCHAR(100) NOT NULL,
    recommended_model VARCHAR(100) NOT NULL,
    trace_count INTEGER NOT NULL,
    estimated_savings_per_month DECIMAL(10, 2),
    quality_impact_estimate DECIMAL(5, 4),
    confidence DECIMAL(5, 4),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_cost_recommendation_status CHECK (status IN ('PENDING', 'APPLIED', 'DISMISSED'))
);

-- Indexes for cost_recommendations
CREATE INDEX idx_cost_recommendations_project ON cost_recommendations(project_id);
CREATE INDEX idx_cost_recommendations_project_status ON cost_recommendations(project_id, status);
