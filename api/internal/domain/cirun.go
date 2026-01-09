package domain

import (
	"time"

	"github.com/google/uuid"
)

// CIRun represents a CI/CD pipeline execution
type CIRun struct {
	ID        uuid.UUID `json:"id" ch:"id"`
	ProjectID uuid.UUID `json:"projectId" ch:"project_id"`

	// CI provider info
	Provider       CIProvider `json:"provider" ch:"provider"`
	ProviderRunID  string     `json:"providerRunId" ch:"provider_run_id"`
	ProviderRunURL string     `json:"providerRunUrl,omitempty" ch:"provider_run_url"`

	// Pipeline info
	PipelineName string `json:"pipelineName,omitempty" ch:"pipeline_name"`
	JobName      string `json:"jobName,omitempty" ch:"job_name"`
	WorkflowName string `json:"workflowName,omitempty" ch:"workflow_name"`

	// Git context
	GitCommitSha string `json:"gitCommitSha,omitempty" ch:"git_commit_sha"`
	GitBranch    string `json:"gitBranch,omitempty" ch:"git_branch"`
	GitTag       string `json:"gitTag,omitempty" ch:"git_tag"`
	GitRepoURL   string `json:"gitRepoUrl,omitempty" ch:"git_repo_url"`
	GitRef       string `json:"gitRef,omitempty" ch:"git_ref"`

	// Pull request context
	PRNumber       uint32 `json:"prNumber,omitempty" ch:"pr_number"`
	PRTitle        string `json:"prTitle,omitempty" ch:"pr_title"`
	PRSourceBranch string `json:"prSourceBranch,omitempty" ch:"pr_source_branch"`
	PRTargetBranch string `json:"prTargetBranch,omitempty" ch:"pr_target_branch"`

	// Execution
	StartedAt   time.Time  `json:"startedAt" ch:"started_at"`
	CompletedAt *time.Time `json:"completedAt,omitempty" ch:"completed_at"`
	DurationMs  uint32     `json:"durationMs" ch:"duration_ms"`

	// Status
	Status       CIRunStatus `json:"status" ch:"status"`
	Conclusion   string      `json:"conclusion,omitempty" ch:"conclusion"`
	ErrorMessage string      `json:"errorMessage,omitempty" ch:"error_message"`

	// Associated traces
	TraceIDs   []string `json:"traceIds,omitempty" ch:"trace_ids"`
	TraceCount uint32   `json:"traceCount" ch:"trace_count"`

	// Aggregated metrics
	TotalCost         float64 `json:"totalCost" ch:"total_cost"`
	TotalTokens       uint64  `json:"totalTokens" ch:"total_tokens"`
	TotalObservations uint64  `json:"totalObservations" ch:"total_observations"`

	// Runner info
	RunnerName string `json:"runnerName,omitempty" ch:"runner_name"`
	RunnerOS   string `json:"runnerOs,omitempty" ch:"runner_os"`
	RunnerArch string `json:"runnerArch,omitempty" ch:"runner_arch"`

	// Trigger info
	TriggeredBy  string `json:"triggeredBy,omitempty" ch:"triggered_by"`
	TriggerEvent string `json:"triggerEvent,omitempty" ch:"trigger_event"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" ch:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" ch:"updated_at"`
}

// CIRunInput represents input for creating a CI run
type CIRunInput struct {
	Provider       CIProvider `json:"provider" validate:"required"`
	ProviderRunID  string     `json:"providerRunId" validate:"required"`
	ProviderRunURL *string    `json:"providerRunUrl,omitempty"`

	PipelineName *string `json:"pipelineName,omitempty"`
	JobName      *string `json:"jobName,omitempty"`
	WorkflowName *string `json:"workflowName,omitempty"`

	GitCommitSha *string `json:"gitCommitSha,omitempty"`
	GitBranch    *string `json:"gitBranch,omitempty"`
	GitTag       *string `json:"gitTag,omitempty"`
	GitRepoURL   *string `json:"gitRepoUrl,omitempty"`
	GitRef       *string `json:"gitRef,omitempty"`

	PRNumber       *uint32 `json:"prNumber,omitempty"`
	PRTitle        *string `json:"prTitle,omitempty"`
	PRSourceBranch *string `json:"prSourceBranch,omitempty"`
	PRTargetBranch *string `json:"prTargetBranch,omitempty"`

	StartedAt *time.Time `json:"startedAt,omitempty"`

	Status       *CIRunStatus `json:"status,omitempty"`
	Conclusion   *string      `json:"conclusion,omitempty"`
	ErrorMessage *string      `json:"errorMessage,omitempty"`

	RunnerName *string `json:"runnerName,omitempty"`
	RunnerOS   *string `json:"runnerOs,omitempty"`
	RunnerArch *string `json:"runnerArch,omitempty"`

	TriggeredBy  *string `json:"triggeredBy,omitempty"`
	TriggerEvent *string `json:"triggerEvent,omitempty"`
}

// CIRunUpdateInput represents input for updating a CI run
type CIRunUpdateInput struct {
	Status            *CIRunStatus `json:"status,omitempty"`
	Conclusion        *string      `json:"conclusion,omitempty"`
	ErrorMessage      *string      `json:"errorMessage,omitempty"`
	CompletedAt       *time.Time   `json:"completedAt,omitempty"`
	TraceIDs          []string     `json:"traceIds,omitempty"`
	TotalCost         *float64     `json:"totalCost,omitempty"`
	TotalTokens       *uint64      `json:"totalTokens,omitempty"`
	TotalObservations *uint64      `json:"totalObservations,omitempty"`
}

// CIRunFilter represents filter options for querying CI runs
type CIRunFilter struct {
	ProjectID     uuid.UUID
	Provider      *CIProvider
	ProviderRunID *string
	GitCommitSha  *string
	GitBranch     *string
	Status        *CIRunStatus
	FromTime      *time.Time
	ToTime        *time.Time
}

// CIRunList represents a paginated list of CI runs
type CIRunList struct {
	CIRuns     []CIRun `json:"ciRuns"`
	TotalCount int64   `json:"totalCount"`
	HasMore    bool    `json:"hasMore"`
}

// CIRunStats represents statistics for CI runs
type CIRunStats struct {
	TotalRuns      uint64  `json:"totalRuns"`
	SuccessCount   uint64  `json:"successCount"`
	FailureCount   uint64  `json:"failureCount"`
	CancelledCount uint64  `json:"cancelledCount"`
	AvgDurationMs  float64 `json:"avgDurationMs"`
	TotalCost      float64 `json:"totalCost"`
	TotalTokens    uint64  `json:"totalTokens"`
}
