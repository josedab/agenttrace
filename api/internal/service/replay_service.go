package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ReplayService handles building replay timelines from trace data
type ReplayService struct {
	logger *zap.Logger
}

// NewReplayService creates a new replay service
func NewReplayService(logger *zap.Logger) *ReplayService {
	return &ReplayService{
		logger: logger,
	}
}

// BuildTimeline constructs a replay timeline from trace data
func (s *ReplayService) BuildTimeline(
	ctx context.Context,
	trace *domain.Trace,
	observations []domain.Observation,
	fileOps []domain.FileOperation,
	terminalCmds []domain.TerminalCommand,
	checkpoints []domain.Checkpoint,
	gitLinks []domain.GitLink,
) (*domain.ReplayTimeline, error) {
	timeline := &domain.ReplayTimeline{
		TraceID:   trace.ID,
		TraceName: trace.Name,
		StartTime: trace.StartTime,
		EndTime:   trace.EndTime,
		Events:    make([]domain.ReplayEvent, 0),
	}

	if trace.EndTime != nil {
		timeline.Duration = trace.EndTime.Sub(trace.StartTime).Milliseconds()
	}

	// Convert observations to replay events
	for _, obs := range observations {
		event := s.observationToReplayEvent(obs)
		timeline.Events = append(timeline.Events, event)
	}

	// Convert file operations to replay events
	for _, fileOp := range fileOps {
		event := s.fileOpToReplayEvent(fileOp)
		timeline.Events = append(timeline.Events, event)
	}

	// Convert terminal commands to replay events
	for _, cmd := range terminalCmds {
		event := s.terminalCmdToReplayEvent(cmd)
		timeline.Events = append(timeline.Events, event)
	}

	// Convert checkpoints to replay events
	for _, cp := range checkpoints {
		event := s.checkpointToReplayEvent(cp)
		timeline.Events = append(timeline.Events, event)
	}

	// Add git context if available
	if len(gitLinks) > 0 {
		timeline.GitContext = s.buildGitContext(gitLinks)
	}

	// Sort events by timestamp
	sort.Slice(timeline.Events, func(i, j int) bool {
		return timeline.Events[i].Timestamp.Before(timeline.Events[j].Timestamp)
	})

	// Calculate summary
	timeline.Summary = s.calculateSummary(timeline.Events)

	return timeline, nil
}

// observationToReplayEvent converts an observation to a replay event
func (s *ReplayService) observationToReplayEvent(obs domain.Observation) domain.ReplayEvent {
	event := domain.ReplayEvent{
		ID:        obs.ID.String(),
		Timestamp: obs.StartTime,
		Title:     obs.Name,
		Status:    s.observationStatus(obs),
		Data:      domain.ReplayEventData{},
	}

	if obs.EndTime != nil {
		event.Duration = obs.EndTime.Sub(obs.StartTime).Milliseconds()
	}

	switch obs.Type {
	case domain.ObservationTypeGeneration:
		event.Type = domain.ReplayEventLLMCall
		event.Description = "LLM generation"
		event.Data.Model = obs.Model
		event.Data.Input = obs.Input
		event.Data.Output = obs.Output
		if obs.Usage != nil {
			event.Data.TokensInput = obs.Usage.PromptTokens
			event.Data.TokensOutput = obs.Usage.CompletionTokens
		}
		if obs.CalculatedCost != nil {
			event.Data.Cost = *obs.CalculatedCost
		}

	case domain.ObservationTypeSpan:
		// Check if it's a tool call based on name pattern
		if s.isToolCall(obs.Name) {
			event.Type = domain.ReplayEventToolCall
			event.Description = "Tool execution"
			event.Data.ToolName = obs.Name
			event.Data.Arguments = obs.Input
			event.Data.Result = obs.Output
		} else {
			event.Type = domain.ReplayEventAgentThought
			event.Description = "Agent operation"
			event.Data.Input = obs.Input
			event.Data.Output = obs.Output
		}

	case domain.ObservationTypeEvent:
		if obs.Level == "error" || obs.Level == "ERROR" {
			event.Type = domain.ReplayEventError
			event.Description = "Error occurred"
			event.Data.Error = obs.StatusMessage
		} else {
			event.Type = domain.ReplayEventAgentThought
			event.Description = "Agent event"
		}
	}

	return event
}

// fileOpToReplayEvent converts a file operation to a replay event
func (s *ReplayService) fileOpToReplayEvent(fileOp domain.FileOperation) domain.ReplayEvent {
	opDescription := map[string]string{
		"read":   "Read file",
		"write":  "Write file",
		"create": "Create file",
		"delete": "Delete file",
		"modify": "Modify file",
	}

	desc := opDescription[fileOp.Operation]
	if desc == "" {
		desc = fileOp.Operation
	}

	return domain.ReplayEvent{
		ID:          fileOp.ID.String(),
		Type:        domain.ReplayEventFileOperation,
		Timestamp:   fileOp.Timestamp,
		Title:       fileOp.FilePath,
		Description: desc,
		Status:      "success",
		Data: domain.ReplayEventData{
			FilePath:      fileOp.FilePath,
			Operation:     fileOp.Operation,
			Diff:          fileOp.Diff,
			ContentBefore: fileOp.ContentBefore,
			ContentAfter:  fileOp.ContentAfter,
		},
	}
}

// terminalCmdToReplayEvent converts a terminal command to a replay event
func (s *ReplayService) terminalCmdToReplayEvent(cmd domain.TerminalCommand) domain.ReplayEvent {
	status := "success"
	if cmd.ExitCode != nil && *cmd.ExitCode != 0 {
		status = "error"
	}

	var duration int64
	if cmd.EndTime != nil {
		duration = cmd.EndTime.Sub(cmd.StartTime).Milliseconds()
	}

	return domain.ReplayEvent{
		ID:          cmd.ID.String(),
		Type:        domain.ReplayEventTerminalCmd,
		Timestamp:   cmd.StartTime,
		Duration:    duration,
		Title:       cmd.Command,
		Description: "Terminal command",
		Status:      status,
		Data: domain.ReplayEventData{
			Command:    cmd.Command,
			WorkingDir: cmd.WorkingDirectory,
			ExitCode:   cmd.ExitCode,
			Stdout:     cmd.Stdout,
			Stderr:     cmd.Stderr,
		},
	}
}

// checkpointToReplayEvent converts a checkpoint to a replay event
func (s *ReplayService) checkpointToReplayEvent(cp domain.Checkpoint) domain.ReplayEvent {
	return domain.ReplayEvent{
		ID:          cp.ID.String(),
		Type:        domain.ReplayEventCheckpoint,
		Timestamp:   cp.CreatedAt,
		Title:       cp.Name,
		Description: "Checkpoint created",
		Status:      "success",
		Data: domain.ReplayEventData{
			CheckpointID: cp.ID.String(),
			FileManifest: cp.FileManifest,
			GitBranch:    cp.GitBranch,
			GitCommit:    cp.GitCommitSHA,
		},
	}
}

// buildGitContext builds git context from git links
func (s *ReplayService) buildGitContext(gitLinks []domain.GitLink) *domain.ReplayGitContext {
	if len(gitLinks) == 0 {
		return nil
	}

	// Use the first link for base context
	first := gitLinks[0]
	ctx := &domain.ReplayGitContext{
		Repository:    first.Repository,
		Branch:        first.Branch,
		StartCommit:   first.CommitSHA,
		FilesChanged:  make([]string, 0),
	}

	// Collect all changed files
	filesMap := make(map[string]bool)
	for _, link := range gitLinks {
		for _, f := range link.ChangedFiles {
			filesMap[f] = true
		}
		if link.CommitSHA != ctx.StartCommit {
			ctx.EndCommit = link.CommitSHA
		}
	}

	for f := range filesMap {
		ctx.FilesChanged = append(ctx.FilesChanged, f)
	}

	ctx.CommitsInRange = len(gitLinks)

	return ctx
}

// calculateSummary calculates summary statistics from events
func (s *ReplayService) calculateSummary(events []domain.ReplayEvent) domain.ReplaySummary {
	summary := domain.ReplaySummary{
		TotalEvents: len(events),
	}

	var totalDuration int64
	var durationCount int

	for _, event := range events {
		switch event.Type {
		case domain.ReplayEventLLMCall:
			summary.LLMCalls++
			summary.TotalTokens += event.Data.TokensInput + event.Data.TokensOutput
			summary.TotalCost += event.Data.Cost
		case domain.ReplayEventToolCall:
			summary.ToolCalls++
		case domain.ReplayEventFileOperation:
			summary.FileOperations++
		case domain.ReplayEventTerminalCmd:
			summary.TerminalCommands++
		case domain.ReplayEventCheckpoint:
			summary.Checkpoints++
		case domain.ReplayEventError:
			summary.Errors++
		}

		if event.Duration > 0 {
			totalDuration += event.Duration
			durationCount++
		}
	}

	if durationCount > 0 {
		summary.AverageLatency = float64(totalDuration) / float64(durationCount)
	}

	return summary
}

// observationStatus determines the status string for an observation
func (s *ReplayService) observationStatus(obs domain.Observation) string {
	if obs.Level == "error" || obs.Level == "ERROR" {
		return "error"
	}
	if obs.EndTime == nil {
		return "running"
	}
	return "success"
}

// isToolCall checks if an observation name indicates a tool call
func (s *ReplayService) isToolCall(name string) bool {
	// Common tool call patterns
	toolPrefixes := []string{"tool-", "tool_", "call_", "function_"}
	for _, prefix := range toolPrefixes {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// GetTimelineForTrace is a convenience method to get a complete timeline
func (s *ReplayService) GetTimelineForTrace(
	ctx context.Context,
	traceID uuid.UUID,
) (*domain.ReplayTimeline, error) {
	// This would be called by the handler after fetching all the data
	// from various repositories. For now, return an empty timeline.
	return &domain.ReplayTimeline{
		TraceID:   traceID,
		Events:    []domain.ReplayEvent{},
		Summary:   domain.ReplaySummary{},
	}, nil
}

// ExportTimeline exports a timeline in a portable format
func (s *ReplayService) ExportTimeline(timeline *domain.ReplayTimeline) *domain.ReplayExport {
	return &domain.ReplayExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Timeline:   *timeline,
	}
}
