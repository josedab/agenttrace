package domain

import (
	"time"

	"github.com/google/uuid"
)

// OutcomePeriod bounds an analytics query.
type OutcomePeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// OutcomeTraceAggregate contains trace-level outcome counts.
type OutcomeTraceAggregate struct {
	TotalRuns      uint64
	SuccessfulRuns uint64
	FailedRuns     uint64
	InProgressRuns uint64
	TotalCost      float64
}

// OutcomeGitAggregate contains git-link outcome counts.
type OutcomeGitAggregate struct {
	LinkedCommits uint64
	LinkedTraces  uint64
	RevertSignals uint64
}

// OutcomeCIAggregate contains CI outcome counts.
type OutcomeCIAggregate struct {
	TotalRuns      uint64
	PassedRuns     uint64
	FailedRuns     uint64
	CanceledRuns   uint64
	InProgressRuns uint64
	LinkedRuns     uint64
	LinkedFailures uint64
	LinkedPRs      uint64
}

// OutcomeBreakdownAggregate is a repository-level agent or model aggregation.
type OutcomeBreakdownAggregate struct {
	Name           string  `ch:"name"`
	Runs           uint64  `ch:"runs"`
	SuccessfulRuns uint64  `ch:"successful_runs"`
	TotalCost      float64 `ch:"total_cost"`
}

// LinkedOutcomeAggregate describes a recent commit and its latest known CI result.
type LinkedOutcomeAggregate struct {
	CommitSHA     string    `ch:"commit_sha"`
	CommitMessage string    `ch:"commit_message"`
	Branch        string    `ch:"branch"`
	CommittedAt   time.Time `ch:"committed_at"`
	TraceCount    uint64    `ch:"trace_count"`
	CIStatus      string    `ch:"ci_status"`
	CIConclusion  string    `ch:"ci_conclusion"`
	CIProviderURL string    `ch:"provider_run_url"`
	PRNumber      uint32    `ch:"pr_number"`
	PRTitle       string    `ch:"pr_title"`
}

// OutcomeSnapshot is the raw project-scoped repository result.
type OutcomeSnapshot struct {
	Traces         OutcomeTraceAggregate
	Git            OutcomeGitAggregate
	CI             OutcomeCIAggregate
	AgentBreakdown []OutcomeBreakdownAggregate
	ModelBreakdown []OutcomeBreakdownAggregate
	RecentOutcomes []LinkedOutcomeAggregate
}

// OutcomeRateMetric explicitly represents whether a ratio can be calculated.
type OutcomeRateMetric struct {
	Value     *float64 `json:"value"`
	Available bool     `json:"available"`
}

// OutcomeRunMetrics contains agent-run outcome metrics.
type OutcomeRunMetrics struct {
	Total       uint64            `json:"total"`
	Successful  uint64            `json:"successful"`
	Failed      uint64            `json:"failed"`
	InProgress  uint64            `json:"inProgress"`
	SuccessRate OutcomeRateMetric `json:"successRate"`
}

// OutcomeCIMetrics contains CI result metrics.
type OutcomeCIMetrics struct {
	Total      uint64            `json:"total"`
	Passed     uint64            `json:"passed"`
	Failed     uint64            `json:"failed"`
	Canceled   uint64            `json:"cancelled"` //nolint:misspell // Preserve the existing API spelling.
	InProgress uint64            `json:"inProgress"`
	LinkedRuns uint64            `json:"linkedRuns"`
	PassRate   OutcomeRateMetric `json:"passRate"`
}

// OutcomeSCMMetrics contains linked source-control outcomes.
type OutcomeSCMMetrics struct {
	LinkedCommits      uint64 `json:"linkedCommits"`
	LinkedTraces       uint64 `json:"linkedTraces"`
	LinkedPullRequests uint64 `json:"linkedPullRequests"`
	RegressionSignals  uint64 `json:"regressionSignals"`
	RevertSignals      uint64 `json:"revertSignals"`
}

// OutcomeCostMetrics contains cost metrics without fabricating unavailable ratios.
type OutcomeCostMetrics struct {
	TotalCost                float64           `json:"totalCost"`
	CostPerSuccessfulOutcome OutcomeRateMetric `json:"costPerSuccessfulOutcome"`
}

// OutcomeBreakdown is an API-facing model or agent aggregation.
type OutcomeBreakdown struct {
	Name                     string            `json:"name"`
	Runs                     uint64            `json:"runs"`
	SuccessfulRuns           uint64            `json:"successfulRuns"`
	SuccessRate              OutcomeRateMetric `json:"successRate"`
	TotalCost                float64           `json:"totalCost"`
	CostPerSuccessfulOutcome OutcomeRateMetric `json:"costPerSuccessfulOutcome"`
}

// LinkedOutcome is a recent trace-to-git-to-CI correlation.
type LinkedOutcome struct {
	CommitSHA     string    `json:"commitSha"`
	CommitMessage string    `json:"commitMessage,omitempty"`
	Branch        string    `json:"branch,omitempty"`
	CommittedAt   time.Time `json:"committedAt"`
	TraceCount    uint64    `json:"traceCount"`
	CIStatus      *string   `json:"ciStatus"`
	CIConclusion  *string   `json:"ciConclusion"`
	CIProviderURL *string   `json:"ciProviderUrl"`
	PRNumber      *uint32   `json:"prNumber"`
	PRTitle       *string   `json:"prTitle"`
}

// OutcomeAvailability makes missing data explicit to API and UI consumers.
type OutcomeAvailability struct {
	TraceData        bool     `json:"traceData"`
	GitData          bool     `json:"gitData"`
	CIData           bool     `json:"ciData"`
	PullRequestData  bool     `json:"pullRequestData"`
	AgentAttribution bool     `json:"agentAttribution"`
	ModelAttribution bool     `json:"modelAttribution"`
	Unavailable      []string `json:"unavailable"`
}

// OutcomeOverview is the canonical project outcome analytics response.
type OutcomeOverview struct {
	ProjectID      uuid.UUID           `json:"projectId"`
	Period         OutcomePeriod       `json:"period"`
	Runs           OutcomeRunMetrics   `json:"runs"`
	CI             OutcomeCIMetrics    `json:"ci"`
	SourceControl  OutcomeSCMMetrics   `json:"sourceControl"`
	Cost           OutcomeCostMetrics  `json:"cost"`
	ByAgent        []OutcomeBreakdown  `json:"byAgent"`
	ByModel        []OutcomeBreakdown  `json:"byModel"`
	RecentOutcomes []LinkedOutcome     `json:"recentOutcomes"`
	Availability   OutcomeAvailability `json:"availability"`
	GeneratedAt    time.Time           `json:"generatedAt"`
}

// OutcomeDigest is a channel-neutral team report.
type OutcomeDigest struct {
	ProjectID   uuid.UUID     `json:"projectId"`
	Period      OutcomePeriod `json:"period"`
	Title       string        `json:"title"`
	Summary     string        `json:"summary"`
	Highlights  []string      `json:"highlights"`
	Attention   []string      `json:"attention"`
	GeneratedAt time.Time     `json:"generatedAt"`
}
