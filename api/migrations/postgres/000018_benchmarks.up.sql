-- Benchmarks table
CREATE TABLE IF NOT EXISTS benchmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50),
    dataset_id UUID REFERENCES datasets(id) ON DELETE SET NULL,
    evaluator_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    metrics JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for benchmarks
CREATE INDEX idx_benchmarks_category ON benchmarks(category);

-- Trigger for updated_at
CREATE TRIGGER update_benchmarks_updated_at
    BEFORE UPDATE ON benchmarks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Benchmark submissions table
CREATE TABLE IF NOT EXISTS benchmark_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    benchmark_id UUID NOT NULL REFERENCES benchmarks(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    agent_version VARCHAR(50),
    scores JSONB,
    overall_score DECIMAL(5, 4),
    rank INTEGER,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for benchmark_submissions
CREATE INDEX idx_benchmark_submissions_benchmark ON benchmark_submissions(benchmark_id);
CREATE INDEX idx_benchmark_submissions_score ON benchmark_submissions(benchmark_id, overall_score DESC);
CREATE INDEX idx_benchmark_submissions_agent ON benchmark_submissions(agent_name);
