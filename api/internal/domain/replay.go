package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReplayEventType represents the type of event in a replay timeline
type ReplayEventType string

const (
	ReplayEventLLMCall       ReplayEventType = "llm_call"
	ReplayEventToolCall      ReplayEventType = "tool_call"
	ReplayEventFileOperation ReplayEventType = "file_operation"
	ReplayEventTerminalCmd   ReplayEventType = "terminal_command"
	ReplayEventCheckpoint    ReplayEventType = "checkpoint"
	ReplayEventGitOperation  ReplayEventType = "git_operation"
	ReplayEventUserInput     ReplayEventType = "user_input"
	ReplayEventAgentThought  ReplayEventType = "agent_thought"
	ReplayEventError         ReplayEventType = "error"
)

// ReplayTimeline represents the complete replay data for a trace
type ReplayTimeline struct {
	TraceID     uuid.UUID       `json:"traceId"`
	TraceName   string          `json:"traceName"`
	StartTime   time.Time       `json:"startTime"`
	EndTime     *time.Time      `json:"endTime,omitempty"`
	Duration    int64           `json:"durationMs"`
	Events      []ReplayEvent   `json:"events"`
	Summary     ReplaySummary   `json:"summary"`
	GitContext  *ReplayGitContext `json:"gitContext,omitempty"`
}

// ReplayEvent represents a single event in the replay timeline
type ReplayEvent struct {
	ID          string          `json:"id"`
	Type        ReplayEventType `json:"type"`
	Timestamp   time.Time       `json:"timestamp"`
	Duration    int64           `json:"durationMs,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Status      string          `json:"status"` // success, error, pending, running
	Data        ReplayEventData `json:"data"`
	Children    []ReplayEvent   `json:"children,omitempty"`
}

// ReplayEventData contains type-specific data for a replay event
type ReplayEventData struct {
	// LLM Call data
	Model         string         `json:"model,omitempty"`
	Input         any            `json:"input,omitempty"`
	Output        any            `json:"output,omitempty"`
	TokensInput   int            `json:"tokensInput,omitempty"`
	TokensOutput  int            `json:"tokensOutput,omitempty"`
	Cost          float64        `json:"cost,omitempty"`

	// Tool Call data
	ToolName      string         `json:"toolName,omitempty"`
	Arguments     any            `json:"arguments,omitempty"`
	Result        any            `json:"result,omitempty"`

	// File Operation data
	FilePath      string         `json:"filePath,omitempty"`
	Operation     string         `json:"operation,omitempty"` // read, write, create, delete
	Diff          string         `json:"diff,omitempty"`
	ContentBefore string         `json:"contentBefore,omitempty"`
	ContentAfter  string         `json:"contentAfter,omitempty"`

	// Terminal Command data
	Command       string         `json:"command,omitempty"`
	WorkingDir    string         `json:"workingDir,omitempty"`
	ExitCode      *int           `json:"exitCode,omitempty"`
	Stdout        string         `json:"stdout,omitempty"`
	Stderr        string         `json:"stderr,omitempty"`

	// Checkpoint data
	CheckpointID  string         `json:"checkpointId,omitempty"`
	FileManifest  []string       `json:"fileManifest,omitempty"`

	// Git Operation data
	GitCommit     string         `json:"gitCommit,omitempty"`
	GitBranch     string         `json:"gitBranch,omitempty"`
	GitMessage    string         `json:"gitMessage,omitempty"`
	ChangedFiles  []string       `json:"changedFiles,omitempty"`

	// Error data
	Error         string         `json:"error,omitempty"`
	ErrorType     string         `json:"errorType,omitempty"`
	StackTrace    string         `json:"stackTrace,omitempty"`
}

// ReplaySummary contains summary statistics for the replay
type ReplaySummary struct {
	TotalEvents      int     `json:"totalEvents"`
	LLMCalls         int     `json:"llmCalls"`
	ToolCalls        int     `json:"toolCalls"`
	FileOperations   int     `json:"fileOperations"`
	TerminalCommands int     `json:"terminalCommands"`
	Checkpoints      int     `json:"checkpoints"`
	Errors           int     `json:"errors"`
	TotalTokens      int     `json:"totalTokens"`
	TotalCost        float64 `json:"totalCost"`
	AverageLatency   float64 `json:"averageLatencyMs"`
}

// ReplayGitContext contains git context for the replay
type ReplayGitContext struct {
	Repository    string   `json:"repository"`
	Branch        string   `json:"branch"`
	StartCommit   string   `json:"startCommit"`
	EndCommit     string   `json:"endCommit,omitempty"`
	CommitsInRange int     `json:"commitsInRange"`
	FilesChanged  []string `json:"filesChanged"`
}

// ReplayState represents the playback state for the UI
type ReplayState struct {
	CurrentEventIndex int     `json:"currentEventIndex"`
	IsPlaying         bool    `json:"isPlaying"`
	PlaybackSpeed     float64 `json:"playbackSpeed"` // 1.0 = real-time, 2.0 = 2x speed
	ElapsedTime       int64   `json:"elapsedTimeMs"`
}

// ReplayFilter represents filter options for replay events
type ReplayFilter struct {
	EventTypes     []ReplayEventType `json:"eventTypes,omitempty"`
	ShowErrors     bool              `json:"showErrors"`
	ShowSuccessful bool              `json:"showSuccessful"`
	MinDuration    *int64            `json:"minDurationMs,omitempty"`
	SearchQuery    string            `json:"searchQuery,omitempty"`
}

// ReplayExport represents an exportable replay format
type ReplayExport struct {
	Version   string          `json:"version"`
	ExportedAt time.Time      `json:"exportedAt"`
	Timeline  ReplayTimeline  `json:"timeline"`
}
