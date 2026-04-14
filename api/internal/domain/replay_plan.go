package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReplayPlanStatus represents the lifecycle of a safe replay plan.
type ReplayPlanStatus string

// Replay plan states.
const (
	ReplayPlanPlanned     ReplayPlanStatus = "planned"
	ReplayPlanReady       ReplayPlanStatus = "ready"
	ReplayPlanRunning     ReplayPlanStatus = "running"
	ReplayPlanCompleted   ReplayPlanStatus = "completed"
	ReplayPlanFailed      ReplayPlanStatus = "failed"
	ReplayPlanUnsupported ReplayPlanStatus = "unsupported"
)

// ReplayExecutionMode controls what a replay is allowed to do.
type ReplayExecutionMode string

// Supported replay execution modes.
const (
	ReplayModeRecordedGeneration ReplayExecutionMode = "recorded_generation"
	ReplayModeSandbox            ReplayExecutionMode = "sandbox"
)

// ReplayPlanInput constructs a replay request without accepting commands or executable code.
type ReplayPlanInput struct {
	CheckpointID   *uuid.UUID          `json:"checkpointId,omitempty"`
	Mode           ReplayExecutionMode `json:"mode,omitempty"`
	ModelOverride  string              `json:"modelOverride,omitempty"`
	PromptOverride string              `json:"promptOverride,omitempty"`
	Temperature    *float64            `json:"temperature,omitempty"`
}

// ReplayCapabilityReport explicitly describes supported and unsupported replay behavior.
type ReplayCapabilityReport struct {
	CanInspectTimeline          bool     `json:"canInspectTimeline"`
	CanReplayRecordedGeneration bool     `json:"canReplayRecordedGeneration"`
	CanExecuteInSandbox         bool     `json:"canExecuteInSandbox"`
	HasCheckpoint               bool     `json:"hasCheckpoint"`
	HasFileOperations           bool     `json:"hasFileOperations"`
	HasTerminalCommands         bool     `json:"hasTerminalCommands"`
	GenerationCount             int      `json:"generationCount"`
	UnsupportedReasons          []string `json:"unsupportedReasons"`
	SafetyNotice                string   `json:"safetyNotice"`
}

// ReplayGenerationResult records a deterministic replay result without duplicating source content.
type ReplayGenerationResult struct {
	EventID       string  `json:"eventId"`
	Model         string  `json:"model,omitempty"`
	OutputSHA256  string  `json:"outputSha256"`
	Tokens        int     `json:"tokens"`
	ReferenceCost float64 `json:"referenceCost"`
}

// ReplayPlanComparison compares the original generation branch to its safe replay branch.
type ReplayPlanComparison struct {
	OriginalGenerationCount int      `json:"originalGenerationCount"`
	ReplayGenerationCount   int      `json:"replayGenerationCount"`
	OriginalTokens          int      `json:"originalTokens"`
	ReplayTokens            int      `json:"replayTokens"`
	OriginalCost            float64  `json:"originalCost"`
	ReplayProviderCost      float64  `json:"replayProviderCost"`
	Equivalent              bool     `json:"equivalent"`
	Verdict                 string   `json:"verdict"`
	Notes                   []string `json:"notes"`
}

// ReplayPlanResult contains safe, non-executable replay output.
type ReplayPlanResult struct {
	StartedAt                  time.Time                `json:"startedAt"`
	CompletedAt                time.Time                `json:"completedAt"`
	Generations                []ReplayGenerationResult `json:"generations"`
	NonGenerationEventsSkipped int                      `json:"nonGenerationEventsSkipped"`
	Comparison                 ReplayPlanComparison     `json:"comparison"`
}

// ReplayPlanTransition describes an atomic, project-scoped status change.
type ReplayPlanTransition struct {
	From ReplayPlanStatus
	To   ReplayPlanStatus
	At   time.Time
}

// ReplayPlan is a persisted, project-scoped safe replay request.
type ReplayPlan struct {
	ID            uuid.UUID              `json:"id"`
	ProjectID     uuid.UUID              `json:"projectId"`
	TraceID       string                 `json:"traceId"`
	CheckpointID  *uuid.UUID             `json:"checkpointId,omitempty"`
	Status        ReplayPlanStatus       `json:"status"`
	Request       ReplayPlanInput        `json:"request"`
	Capabilities  ReplayCapabilityReport `json:"capabilities"`
	Result        *ReplayPlanResult      `json:"result,omitempty"`
	FailureReason string                 `json:"failureReason,omitempty"`
	CreatedBy     *uuid.UUID             `json:"createdBy,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}
