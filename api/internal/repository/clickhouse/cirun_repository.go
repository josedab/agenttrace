package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// CIRunRepository handles CI run data operations in ClickHouse
type CIRunRepository struct {
	db *database.ClickHouseDB
}

// NewCIRunRepository creates a new CI run repository
func NewCIRunRepository(db *database.ClickHouseDB) *CIRunRepository {
	return &CIRunRepository{db: db}
}

// Create inserts a new CI run
func (r *CIRunRepository) Create(ctx context.Context, ciRun *domain.CIRun) error {
	query := `
		INSERT INTO ci_runs (
			id, project_id, provider, provider_run_id, provider_run_url,
			pipeline_name, job_name, workflow_name, git_commit_sha, git_branch,
			git_tag, git_repo_url, git_ref, pr_number, pr_title, pr_source_branch,
			pr_target_branch, started_at, completed_at, duration_ms, status,
			conclusion, error_message, trace_ids, trace_count, total_cost,
			total_tokens, total_observations, runner_name, runner_os, runner_arch,
			triggered_by, trigger_event, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		ciRun.ID,
		ciRun.ProjectID,
		string(ciRun.Provider),
		ciRun.ProviderRunID,
		ciRun.ProviderRunURL,
		ciRun.PipelineName,
		ciRun.JobName,
		ciRun.WorkflowName,
		ciRun.GitCommitSha,
		ciRun.GitBranch,
		ciRun.GitTag,
		ciRun.GitRepoURL,
		ciRun.GitRef,
		ciRun.PRNumber,
		ciRun.PRTitle,
		ciRun.PRSourceBranch,
		ciRun.PRTargetBranch,
		ciRun.StartedAt,
		ciRun.CompletedAt,
		ciRun.DurationMs,
		string(ciRun.Status),
		ciRun.Conclusion,
		ciRun.ErrorMessage,
		ciRun.TraceIDs,
		ciRun.TraceCount,
		ciRun.TotalCost,
		ciRun.TotalTokens,
		ciRun.TotalObservations,
		ciRun.RunnerName,
		ciRun.RunnerOS,
		ciRun.RunnerArch,
		ciRun.TriggeredBy,
		ciRun.TriggerEvent,
		ciRun.CreatedAt,
		ciRun.UpdatedAt,
	)
}

// GetByID retrieves a CI run by ID
func (r *CIRunRepository) GetByID(ctx context.Context, projectID, ciRunID uuid.UUID) (*domain.CIRun, error) {
	query := `
		SELECT
			id, project_id, provider, provider_run_id, provider_run_url,
			pipeline_name, job_name, workflow_name, git_commit_sha, git_branch,
			git_tag, git_repo_url, git_ref, pr_number, pr_title, pr_source_branch,
			pr_target_branch, started_at, completed_at, duration_ms, status,
			conclusion, error_message, trace_ids, trace_count, total_cost,
			total_tokens, total_observations, runner_name, runner_os, runner_arch,
			triggered_by, trigger_event, created_at, updated_at
		FROM ci_runs FINAL
		WHERE project_id = ? AND id = ?
		LIMIT 1
	`

	var ciRun domain.CIRun
	row := r.db.QueryRow(ctx, query, projectID, ciRunID)
	err := row.Scan(
		&ciRun.ID,
		&ciRun.ProjectID,
		&ciRun.Provider,
		&ciRun.ProviderRunID,
		&ciRun.ProviderRunURL,
		&ciRun.PipelineName,
		&ciRun.JobName,
		&ciRun.WorkflowName,
		&ciRun.GitCommitSha,
		&ciRun.GitBranch,
		&ciRun.GitTag,
		&ciRun.GitRepoURL,
		&ciRun.GitRef,
		&ciRun.PRNumber,
		&ciRun.PRTitle,
		&ciRun.PRSourceBranch,
		&ciRun.PRTargetBranch,
		&ciRun.StartedAt,
		&ciRun.CompletedAt,
		&ciRun.DurationMs,
		&ciRun.Status,
		&ciRun.Conclusion,
		&ciRun.ErrorMessage,
		&ciRun.TraceIDs,
		&ciRun.TraceCount,
		&ciRun.TotalCost,
		&ciRun.TotalTokens,
		&ciRun.TotalObservations,
		&ciRun.RunnerName,
		&ciRun.RunnerOS,
		&ciRun.RunnerArch,
		&ciRun.TriggeredBy,
		&ciRun.TriggerEvent,
		&ciRun.CreatedAt,
		&ciRun.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ciRun, nil
}

// GetByProviderRunID retrieves a CI run by provider's run ID
func (r *CIRunRepository) GetByProviderRunID(ctx context.Context, projectID uuid.UUID, providerRunID string) (*domain.CIRun, error) {
	query := `
		SELECT
			id, project_id, provider, provider_run_id, provider_run_url,
			pipeline_name, job_name, workflow_name, git_commit_sha, git_branch,
			git_tag, git_repo_url, git_ref, pr_number, pr_title, pr_source_branch,
			pr_target_branch, started_at, completed_at, duration_ms, status,
			conclusion, error_message, trace_ids, trace_count, total_cost,
			total_tokens, total_observations, runner_name, runner_os, runner_arch,
			triggered_by, trigger_event, created_at, updated_at
		FROM ci_runs FINAL
		WHERE project_id = ? AND provider_run_id = ?
		LIMIT 1
	`

	var ciRun domain.CIRun
	row := r.db.QueryRow(ctx, query, projectID, providerRunID)
	err := row.Scan(
		&ciRun.ID,
		&ciRun.ProjectID,
		&ciRun.Provider,
		&ciRun.ProviderRunID,
		&ciRun.ProviderRunURL,
		&ciRun.PipelineName,
		&ciRun.JobName,
		&ciRun.WorkflowName,
		&ciRun.GitCommitSha,
		&ciRun.GitBranch,
		&ciRun.GitTag,
		&ciRun.GitRepoURL,
		&ciRun.GitRef,
		&ciRun.PRNumber,
		&ciRun.PRTitle,
		&ciRun.PRSourceBranch,
		&ciRun.PRTargetBranch,
		&ciRun.StartedAt,
		&ciRun.CompletedAt,
		&ciRun.DurationMs,
		&ciRun.Status,
		&ciRun.Conclusion,
		&ciRun.ErrorMessage,
		&ciRun.TraceIDs,
		&ciRun.TraceCount,
		&ciRun.TotalCost,
		&ciRun.TotalTokens,
		&ciRun.TotalObservations,
		&ciRun.RunnerName,
		&ciRun.RunnerOS,
		&ciRun.RunnerArch,
		&ciRun.TriggeredBy,
		&ciRun.TriggerEvent,
		&ciRun.CreatedAt,
		&ciRun.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ciRun, nil
}

// List retrieves CI runs with filtering
func (r *CIRunRepository) List(ctx context.Context, filter *domain.CIRunFilter, limit, offset int) (*domain.CIRunList, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{filter.ProjectID}

	if filter.Provider != nil {
		conditions = append(conditions, "provider = ?")
		args = append(args, string(*filter.Provider))
	}

	if filter.ProviderRunID != nil {
		conditions = append(conditions, "provider_run_id = ?")
		args = append(args, *filter.ProviderRunID)
	}

	if filter.GitCommitSha != nil {
		conditions = append(conditions, "git_commit_sha = ?")
		args = append(args, *filter.GitCommitSha)
	}

	if filter.GitBranch != nil {
		conditions = append(conditions, "git_branch = ?")
		args = append(args, *filter.GitBranch)
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, string(*filter.Status))
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "started_at >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "started_at <= ?")
		args = append(args, *filter.ToTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get count
	countQuery := fmt.Sprintf("SELECT count() FROM ci_runs FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get CI runs
	query := fmt.Sprintf(`
		SELECT
			id, project_id, provider, provider_run_id, provider_run_url,
			pipeline_name, job_name, workflow_name, git_commit_sha, git_branch,
			git_tag, git_repo_url, git_ref, pr_number, pr_title, pr_source_branch,
			pr_target_branch, started_at, completed_at, duration_ms, status,
			conclusion, error_message, trace_ids, trace_count, total_cost,
			total_tokens, total_observations, runner_name, runner_os, runner_arch,
			triggered_by, trigger_event, created_at, updated_at
		FROM ci_runs FINAL
		WHERE %s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var ciRuns []domain.CIRun
	if err := r.db.Select(ctx, &ciRuns, query, args...); err != nil {
		return nil, err
	}

	return &domain.CIRunList{
		CIRuns:     ciRuns,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(ciRuns)) < totalCount,
	}, nil
}

// Update updates a CI run
func (r *CIRunRepository) Update(ctx context.Context, ciRun *domain.CIRun) error {
	ciRun.UpdatedAt = time.Now()
	return r.Create(ctx, ciRun) // ReplacingMergeTree handles updates
}

// GetStats retrieves CI run statistics
func (r *CIRunRepository) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.CIRunStats, error) {
	query := `
		SELECT
			count() AS total_runs,
			countIf(status = 'success') AS success_count,
			countIf(status = 'failure') AS failure_count,
			countIf(status = 'cancelled') AS cancelled_count,
			avg(duration_ms) AS avg_duration_ms,
			sum(total_cost) AS total_cost,
			sum(total_tokens) AS total_tokens
		FROM ci_runs FINAL
		WHERE project_id = ?
	`

	var stats domain.CIRunStats
	row := r.db.QueryRow(ctx, query, projectID)
	err := row.Scan(
		&stats.TotalRuns,
		&stats.SuccessCount,
		&stats.FailureCount,
		&stats.CancelledCount,
		&stats.AvgDurationMs,
		&stats.TotalCost,
		&stats.TotalTokens,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
