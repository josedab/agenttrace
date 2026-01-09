package domain

import (
	"time"

	"github.com/google/uuid"
)

// TerminalCommand represents a terminal command executed during agent execution
type TerminalCommand struct {
	ID            uuid.UUID `json:"id" ch:"id"`
	ProjectID     uuid.UUID `json:"projectId" ch:"project_id"`
	TraceID       string    `json:"traceId" ch:"trace_id"`
	ObservationID *string   `json:"observationId,omitempty" ch:"observation_id"`

	// Command details
	Command          string   `json:"command" ch:"command"`
	Args             []string `json:"args,omitempty" ch:"args"`
	WorkingDirectory string   `json:"workingDirectory,omitempty" ch:"working_directory"`
	Shell            string   `json:"shell,omitempty" ch:"shell"`
	EnvVars          string   `json:"envVars,omitempty" ch:"env_vars"`

	// Execution timing
	StartedAt   time.Time  `json:"startedAt" ch:"started_at"`
	CompletedAt *time.Time `json:"completedAt,omitempty" ch:"completed_at"`
	DurationMs  uint32     `json:"durationMs" ch:"duration_ms"`

	// Result
	ExitCode        int32  `json:"exitCode" ch:"exit_code"`
	Stdout          string `json:"stdout,omitempty" ch:"stdout"`
	Stderr          string `json:"stderr,omitempty" ch:"stderr"`
	StdoutTruncated bool   `json:"stdoutTruncated" ch:"stdout_truncated"`
	StderrTruncated bool   `json:"stderrTruncated" ch:"stderr_truncated"`

	// Status
	Success  bool `json:"success" ch:"success"`
	TimedOut bool `json:"timedOut" ch:"timed_out"`
	Killed   bool `json:"killed" ch:"killed"`

	// Resource usage
	MaxMemoryBytes uint64 `json:"maxMemoryBytes" ch:"max_memory_bytes"`
	CPUTimeMs      uint32 `json:"cpuTimeMs" ch:"cpu_time_ms"`

	// Context
	ToolName string `json:"toolName,omitempty" ch:"tool_name"`
	Reason   string `json:"reason,omitempty" ch:"reason"`
}

// TerminalCommandInput represents input for creating a terminal command
type TerminalCommandInput struct {
	TraceID          string   `json:"traceId" validate:"required"`
	ObservationID    *string  `json:"observationId,omitempty"`
	Command          string   `json:"command" validate:"required"`
	Args             []string `json:"args,omitempty"`
	WorkingDirectory *string  `json:"workingDirectory,omitempty"`
	Shell            *string  `json:"shell,omitempty"`
	EnvVars          *string  `json:"envVars,omitempty"`

	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	DurationMs  *uint32    `json:"durationMs,omitempty"`

	ExitCode        *int32  `json:"exitCode,omitempty"`
	Stdout          *string `json:"stdout,omitempty"`
	Stderr          *string `json:"stderr,omitempty"`
	StdoutTruncated *bool   `json:"stdoutTruncated,omitempty"`
	StderrTruncated *bool   `json:"stderrTruncated,omitempty"`

	Success  *bool `json:"success,omitempty"`
	TimedOut *bool `json:"timedOut,omitempty"`
	Killed   *bool `json:"killed,omitempty"`

	MaxMemoryBytes *uint64 `json:"maxMemoryBytes,omitempty"`
	CPUTimeMs      *uint32 `json:"cpuTimeMs,omitempty"`

	ToolName *string `json:"toolName,omitempty"`
	Reason   *string `json:"reason,omitempty"`
}

// TerminalCommandFilter represents filter options for querying terminal commands
type TerminalCommandFilter struct {
	ProjectID     uuid.UUID
	TraceID       *string
	ObservationID *string
	Command       *string
	ExitCode      *int32
	Success       *bool
	FromTime      *time.Time
	ToTime        *time.Time
}

// TerminalCommandList represents a paginated list of terminal commands
type TerminalCommandList struct {
	TerminalCommands []TerminalCommand `json:"terminalCommands"`
	TotalCount       int64             `json:"totalCount"`
	HasMore          bool              `json:"hasMore"`
}

// TerminalCommandStats represents statistics for terminal commands
type TerminalCommandStats struct {
	TotalCommands  uint64  `json:"totalCommands"`
	SuccessCount   uint64  `json:"successCount"`
	FailureCount   uint64  `json:"failureCount"`
	TimeoutCount   uint64  `json:"timeoutCount"`
	AvgDurationMs  float64 `json:"avgDurationMs"`
	TotalDurationMs uint64  `json:"totalDurationMs"`
}
