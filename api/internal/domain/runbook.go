package domain

import (
	"time"

	"github.com/google/uuid"
)

// RunbookStatus represents the status of a runbook
type RunbookStatus string

const (
	RunbookStatusDraft    RunbookStatus = "draft"
	RunbookStatusActive   RunbookStatus = "active"
	RunbookStatusDisabled RunbookStatus = "disabled"
)

// RunbookActionType represents the type of action a runbook can take
type RunbookActionType string

const (
	RunbookActionRetryWithModel   RunbookActionType = "retry_with_model"
	RunbookActionEscalateToHuman  RunbookActionType = "escalate_to_human"
	RunbookActionRollbackPrompt   RunbookActionType = "rollback_prompt"
	RunbookActionAdjustTemperature RunbookActionType = "adjust_temperature"
	RunbookActionSendNotification RunbookActionType = "send_notification"
	RunbookActionWebhook          RunbookActionType = "webhook"
	RunbookActionCustomScript     RunbookActionType = "custom_script"
)

// RunbookExecutionStatus represents the status of a runbook execution
type RunbookExecutionStatus string

const (
	RunbookExecPending   RunbookExecutionStatus = "pending"
	RunbookExecRunning   RunbookExecutionStatus = "running"
	RunbookExecCompleted RunbookExecutionStatus = "completed"
	RunbookExecFailed    RunbookExecutionStatus = "failed"
	RunbookExecSkipped   RunbookExecutionStatus = "skipped"
)

// Runbook represents a YAML-defined runbook for automated trace remediation
type Runbook struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Status      RunbookStatus  `json:"status"`
	Version     int            `json:"version"`
	YAMLContent string         `json:"yamlContent"` // Full YAML definition

	// Parsed from YAML
	Triggers []RunbookTrigger `json:"triggers"`
	Actions  []RunbookAction  `json:"actions"`

	// Configuration
	MaxExecutionsPerHour int  `json:"maxExecutionsPerHour"`
	RequireApproval      bool `json:"requireApproval"`

	// Audit
	CreatedBy uuid.UUID `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RunbookTrigger defines when a runbook should fire
type RunbookTrigger struct {
	Type        string                 `json:"type" yaml:"type"` // "pattern", "threshold", "error_match", "anomaly"
	Conditions  map[string]interface{} `json:"conditions" yaml:"conditions"`
	Description string                 `json:"description,omitempty" yaml:"description"`
}

// RunbookAction defines an action to take when triggered
type RunbookAction struct {
	Name       string            `json:"name" yaml:"name"`
	Type       RunbookActionType `json:"type" yaml:"type"`
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters"`
	Timeout    string            `json:"timeout,omitempty" yaml:"timeout"` // e.g., "30s", "5m"
	OnFailure  string            `json:"onFailure,omitempty" yaml:"on_failure"` // "continue", "abort", "retry"
	RetryCount int               `json:"retryCount,omitempty" yaml:"retry_count"`
}

// RunbookExecution represents a single execution of a runbook
type RunbookExecution struct {
	ID           uuid.UUID              `json:"id"`
	RunbookID    uuid.UUID              `json:"runbookId"`
	ProjectID    uuid.UUID              `json:"projectId"`
	TraceID      string                 `json:"traceId"`
	Status       RunbookExecutionStatus `json:"status"`
	TriggerMatch string                 `json:"triggerMatch"` // Which trigger fired
	ActionResults []ActionResult        `json:"actionResults"`
	StartedAt    time.Time              `json:"startedAt"`
	CompletedAt  *time.Time             `json:"completedAt,omitempty"`
	Error        string                 `json:"error,omitempty"`

	// Approval tracking
	ApprovalRequired bool       `json:"approvalRequired"`
	ApprovedBy       *uuid.UUID `json:"approvedBy,omitempty"`
	ApprovedAt       *time.Time `json:"approvedAt,omitempty"`
}

// ActionResult tracks the result of a single runbook action
type ActionResult struct {
	ActionName string                 `json:"actionName"`
	ActionType RunbookActionType      `json:"actionType"`
	Status     RunbookExecutionStatus `json:"status"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Duration   int64                  `json:"durationMs"`
	StartedAt  time.Time              `json:"startedAt"`
}

// RunbookInput represents input for creating/updating a runbook
type RunbookInput struct {
	Name                 string `json:"name" validate:"required"`
	Description          string `json:"description,omitempty"`
	YAMLContent          string `json:"yamlContent" validate:"required"`
	MaxExecutionsPerHour *int   `json:"maxExecutionsPerHour,omitempty"`
	RequireApproval      *bool  `json:"requireApproval,omitempty"`
}

// RunbookFilter for querying runbooks
type RunbookFilter struct {
	ProjectID uuid.UUID
	Status    *RunbookStatus
}

// RunbookList represents a paginated list of runbooks
type RunbookList struct {
	Runbooks   []Runbook `json:"runbooks"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}

// RunbookExecutionList represents a paginated list of executions
type RunbookExecutionList struct {
	Executions []RunbookExecution `json:"executions"`
	TotalCount int64              `json:"totalCount"`
	HasMore    bool               `json:"hasMore"`
}

// RunbookTestInput for testing a runbook against a trace
type RunbookTestInput struct {
	RunbookID uuid.UUID `json:"runbookId" validate:"required"`
	TraceID   string    `json:"traceId" validate:"required"`
	DryRun    bool      `json:"dryRun"` // If true, don't execute actions
}

// RunbookTestResult represents the result of testing a runbook
type RunbookTestResult struct {
	Matched      bool              `json:"matched"`
	MatchedTriggers []string      `json:"matchedTriggers"`
	PlannedActions  []RunbookAction `json:"plannedActions"`
	Execution    *RunbookExecution `json:"execution,omitempty"` // Only if not dry run
}

// RunbookYAMLExample provides the YAML schema reference
const RunbookYAMLExample = `
name: high-cost-retry
description: Retry with cheaper model when cost exceeds threshold

triggers:
  - type: threshold
    conditions:
      metric: cost
      operator: gt
      value: 1.0
    description: Cost exceeds $1.00

  - type: error_match
    conditions:
      pattern: "rate_limit_exceeded"
    description: Rate limit hit

actions:
  - name: retry-with-cheaper-model
    type: retry_with_model
    parameters:
      model: gpt-4o-mini
      max_retries: "2"
    timeout: 60s
    on_failure: continue

  - name: notify-team
    type: send_notification
    parameters:
      channel: slack
      message: "High-cost trace detected and retried"
    on_failure: continue
`
