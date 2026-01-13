-- Drop experiment_metrics table
DROP TABLE IF EXISTS experiment_metrics;

-- Drop experiment_assignments table
DROP TABLE IF EXISTS experiment_assignments;

-- Drop foreign key constraint from experiments
ALTER TABLE experiments DROP CONSTRAINT IF EXISTS fk_experiments_winning_variant;

-- Drop experiment_variants table
DROP TABLE IF EXISTS experiment_variants;

-- Drop experiments table
DROP TABLE IF EXISTS experiments;
