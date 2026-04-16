INSERT INTO projects (
    organization_id,
    name,
    slug,
    settings,
    retention_days,
    rate_limit_per_minute
)
SELECT
    organizations.id,
    'Default Project',
    'default-project',
    '{"systemProvisioned": true}'::jsonb,
    30,
    1000
FROM organizations
WHERE NOT EXISTS (
    SELECT 1
    FROM projects
    WHERE projects.organization_id = organizations.id
);
