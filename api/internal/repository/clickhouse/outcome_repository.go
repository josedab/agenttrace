package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

const outcomeTraceAggregateQuery = `
	SELECT
		count() AS total_runs,
		countIf(isNotNull(end_time) AND level != 'ERROR') AS successful_runs,
		countIf(level = 'ERROR') AS failed_runs,
		countIf(isNull(end_time) AND level != 'ERROR') AS in_progress_runs,
		toFloat64(sum(total_cost)) AS total_cost
	FROM traces FINAL
	WHERE project_id = ? AND start_time >= ? AND start_time < ?
`

const outcomeGitAggregateQuery = `
	SELECT
		countDistinctIf(commit_sha, commit_sha != '') AS linked_commits,
		countDistinct(trace_id) AS linked_traces,
		countIf(startsWith(lowerUTF8(trim(commit_message)), 'revert')) AS revert_signals
	FROM git_links FINAL
	WHERE project_id = ? AND created_at >= ? AND created_at < ?
`

const outcomeCIAggregateQuery = `
	SELECT
		count() AS total_runs,
		countIf(status = 'success') AS passed_runs,
		countIf(status = 'failure') AS failed_runs,
		countIf(status = 'cancelled') AS canceled_runs,
		countIf(status IN ('pending', 'running')) AS in_progress_runs,
		countDistinctIf(pr_number, pr_number > 0) AS linked_prs
	FROM ci_runs FINAL
	WHERE project_id = ? AND started_at >= ? AND started_at < ?
`

const outcomeLinkedCIAggregateQuery = `
	SELECT
		count() AS linked_runs,
		countIf(ci.status = 'failure') AS linked_failures
	FROM (
		SELECT git_commit_sha, status
		FROM ci_runs FINAL
		WHERE project_id = ? AND started_at >= ? AND started_at < ? AND git_commit_sha != ''
	) AS ci
	INNER JOIN (
		SELECT DISTINCT commit_sha
		FROM git_links FINAL
		WHERE project_id = ? AND created_at >= ? AND created_at < ? AND commit_sha != ''
	) AS links ON ci.git_commit_sha = links.commit_sha
`

const outcomeAgentBreakdownQuery = `
	SELECT
		JSONExtractString(metadata, 'agent_name') AS name,
		count() AS runs,
		countIf(isNotNull(end_time) AND level != 'ERROR') AS successful_runs,
		toFloat64(sum(total_cost)) AS total_cost
	FROM traces FINAL
	WHERE project_id = ? AND start_time >= ? AND start_time < ? AND name != ''
	GROUP BY name
	ORDER BY runs DESC, name ASC
	LIMIT 50
`

const outcomeModelBreakdownQuery = `
	SELECT
		observations.model AS name,
		countDistinct(traces.id) AS runs,
		countDistinctIf(traces.id, isNotNull(traces.end_time) AND traces.level != 'ERROR') AS successful_runs,
		toFloat64(sum(observations.total_cost)) AS total_cost
	FROM (
		SELECT id, project_id, level, end_time
		FROM traces FINAL
		WHERE project_id = ? AND start_time >= ? AND start_time < ?
	) AS traces
	INNER JOIN (
		SELECT trace_id, project_id, model, total_cost
		FROM observations FINAL
		WHERE project_id = ? AND start_time >= ? AND start_time < ?
			AND type = 'GENERATION' AND model != ''
	) AS observations
		ON traces.project_id = observations.project_id AND traces.id = observations.trace_id
	GROUP BY name
	ORDER BY runs DESC, name ASC
	LIMIT 50
`

const outcomeRecentQuery = `
	SELECT
		links.commit_sha,
		links.commit_message,
		links.branch,
		links.committed_at,
		links.trace_count,
		ifNull(ci.status, '') AS ci_status,
		ifNull(ci.conclusion, '') AS ci_conclusion,
		ifNull(ci.provider_run_url, '') AS provider_run_url,
		ifNull(ci.pr_number, 0) AS pr_number,
		ifNull(ci.pr_title, '') AS pr_title
	FROM (
		SELECT
			commit_sha,
			any(commit_message) AS commit_message,
			any(branch) AS branch,
			max(commit_timestamp) AS committed_at,
			countDistinct(trace_id) AS trace_count
		FROM git_links FINAL
		WHERE project_id = ? AND created_at >= ? AND created_at < ? AND commit_sha != ''
		GROUP BY commit_sha
	) AS links
	LEFT JOIN (
		SELECT
			git_commit_sha,
			argMax(toString(status), started_at) AS status,
			argMax(conclusion, started_at) AS conclusion,
			argMax(provider_run_url, started_at) AS provider_run_url,
			argMax(pr_number, started_at) AS pr_number,
			argMax(pr_title, started_at) AS pr_title
		FROM ci_runs FINAL
		WHERE project_id = ? AND started_at >= ? AND started_at < ? AND git_commit_sha != ''
		GROUP BY git_commit_sha
	) AS ci ON links.commit_sha = ci.git_commit_sha
	ORDER BY links.committed_at DESC
	LIMIT 20
`

// OutcomeRepository aggregates project-scoped trace, git, and CI outcomes.
type OutcomeRepository struct {
	db *database.ClickHouseDB
}

// NewOutcomeRepository creates an outcome analytics repository.
func NewOutcomeRepository(db *database.ClickHouseDB) *OutcomeRepository {
	return &OutcomeRepository{db: db}
}

// GetSnapshot returns real outcome data for a single project and period.
func (r *OutcomeRepository) GetSnapshot(
	ctx context.Context,
	projectID uuid.UUID,
	from, to time.Time,
) (*domain.OutcomeSnapshot, error) {
	snapshot := &domain.OutcomeSnapshot{
		AgentBreakdown: []domain.OutcomeBreakdownAggregate{},
		ModelBreakdown: []domain.OutcomeBreakdownAggregate{},
		RecentOutcomes: []domain.LinkedOutcomeAggregate{},
	}

	if err := r.db.QueryRow(ctx, outcomeTraceAggregateQuery, projectID, from, to).Scan(
		&snapshot.Traces.TotalRuns,
		&snapshot.Traces.SuccessfulRuns,
		&snapshot.Traces.FailedRuns,
		&snapshot.Traces.InProgressRuns,
		&snapshot.Traces.TotalCost,
	); err != nil {
		return nil, fmt.Errorf("aggregate trace outcomes: %w", err)
	}

	if err := r.db.QueryRow(ctx, outcomeGitAggregateQuery, projectID, from, to).Scan(
		&snapshot.Git.LinkedCommits,
		&snapshot.Git.LinkedTraces,
		&snapshot.Git.RevertSignals,
	); err != nil {
		return nil, fmt.Errorf("aggregate git outcomes: %w", err)
	}

	if err := r.db.QueryRow(ctx, outcomeCIAggregateQuery, projectID, from, to).Scan(
		&snapshot.CI.TotalRuns,
		&snapshot.CI.PassedRuns,
		&snapshot.CI.FailedRuns,
		&snapshot.CI.CanceledRuns,
		&snapshot.CI.InProgressRuns,
		&snapshot.CI.LinkedPRs,
	); err != nil {
		return nil, fmt.Errorf("aggregate CI outcomes: %w", err)
	}

	if err := r.db.QueryRow(
		ctx,
		outcomeLinkedCIAggregateQuery,
		projectID,
		from,
		to,
		projectID,
		from,
		to,
	).Scan(&snapshot.CI.LinkedRuns, &snapshot.CI.LinkedFailures); err != nil {
		return nil, fmt.Errorf("aggregate linked CI outcomes: %w", err)
	}

	if err := r.db.Select(
		ctx,
		&snapshot.AgentBreakdown,
		outcomeAgentBreakdownQuery,
		projectID,
		from,
		to,
	); err != nil {
		return nil, fmt.Errorf("aggregate agent outcomes: %w", err)
	}

	if err := r.db.Select(
		ctx,
		&snapshot.ModelBreakdown,
		outcomeModelBreakdownQuery,
		projectID,
		from,
		to,
		projectID,
		from,
		to,
	); err != nil {
		return nil, fmt.Errorf("aggregate model outcomes: %w", err)
	}

	if err := r.db.Select(
		ctx,
		&snapshot.RecentOutcomes,
		outcomeRecentQuery,
		projectID,
		from,
		to,
		projectID,
		from,
		to,
	); err != nil {
		return nil, fmt.Errorf("list recent linked outcomes: %w", err)
	}

	return snapshot, nil
}
