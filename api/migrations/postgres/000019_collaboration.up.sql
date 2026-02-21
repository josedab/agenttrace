-- Trace annotations table
CREATE TABLE IF NOT EXISTS trace_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(255) NOT NULL,
    event_id VARCHAR(255),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name VARCHAR(100),
    content TEXT NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for trace_annotations
CREATE INDEX idx_trace_annotations_project ON trace_annotations(project_id);
CREATE INDEX idx_trace_annotations_trace ON trace_annotations(trace_id);

-- Shared sessions table
CREATE TABLE IF NOT EXISTS shared_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trace_id VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participants JSONB NOT NULL DEFAULT '[]'::jsonb,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for shared_sessions
CREATE INDEX idx_shared_sessions_project ON shared_sessions(project_id);
CREATE INDEX idx_shared_sessions_trace ON shared_sessions(trace_id);
