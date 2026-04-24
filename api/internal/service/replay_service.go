package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ReplayTraceRepository defines trace repository operations needed for replay
type ReplayTraceRepository interface {
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
}

// ReplayObservationRepository defines observation repository operations needed for replay
type ReplayObservationRepository interface {
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error)
}

// ReplayFileOperationRepository defines file operation repository operations needed for replay
type ReplayFileOperationRepository interface {
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.FileOperation, error)
}

// ReplayTerminalCommandRepository defines terminal command repository operations needed for replay
type ReplayTerminalCommandRepository interface {
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TerminalCommand, error)
}

// ReplayCheckpointRepository defines checkpoint repository operations needed for replay
type ReplayCheckpointRepository interface {
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Checkpoint, error)
}

// ReplayGitLinkRepository defines git link repository operations needed for replay
type ReplayGitLinkRepository interface {
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.GitLink, error)
}

// ReplayService handles building replay timelines from trace data
type ReplayService struct {
	logger              *zap.Logger
	traceRepo           ReplayTraceRepository
	observationRepo     ReplayObservationRepository
	fileOperationRepo   ReplayFileOperationRepository
	terminalCommandRepo ReplayTerminalCommandRepository
	checkpointRepo      ReplayCheckpointRepository
	gitLinkRepo         ReplayGitLinkRepository
}

// NewReplayService creates a new replay service
func NewReplayService(
	logger *zap.Logger,
	traceRepo ReplayTraceRepository,
	observationRepo ReplayObservationRepository,
	fileOperationRepo ReplayFileOperationRepository,
	terminalCommandRepo ReplayTerminalCommandRepository,
	checkpointRepo ReplayCheckpointRepository,
	gitLinkRepo ReplayGitLinkRepository,
) *ReplayService {
	return &ReplayService{
		logger:              logger,
		traceRepo:           traceRepo,
		observationRepo:     observationRepo,
		fileOperationRepo:   fileOperationRepo,
		terminalCommandRepo: terminalCommandRepo,
		checkpointRepo:      checkpointRepo,
		gitLinkRepo:         gitLinkRepo,
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
	// Parse trace ID string to UUID
	traceUUID, err := uuid.Parse(trace.ID)
	if err != nil {
		return nil, err
	}

	timeline := &domain.ReplayTimeline{
		TraceID:   traceUUID,
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
		ID:        obs.ID,
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
		if obs.UsageDetails.InputTokens > 0 || obs.UsageDetails.OutputTokens > 0 {
			event.Data.TokensInput = int(obs.UsageDetails.InputTokens)
			event.Data.TokensOutput = int(obs.UsageDetails.OutputTokens)
		}
		if obs.CostDetails.TotalCost > 0 {
			event.Data.Cost = obs.CostDetails.TotalCost
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
	opDescription := map[domain.FileOperationType]string{
		domain.FileOperationRead:   "Read file",
		domain.FileOperationUpdate: "Write file",
		domain.FileOperationCreate: "Create file",
		domain.FileOperationDelete: "Delete file",
	}

	desc := opDescription[fileOp.Operation]
	if desc == "" {
		desc = string(fileOp.Operation)
	}

	status := "success"
	if !fileOp.Success {
		status = "error"
	}

	return domain.ReplayEvent{
		ID:          fileOp.ID.String(),
		Type:        domain.ReplayEventFileOperation,
		Timestamp:   fileOp.StartedAt,
		Title:       fileOp.FilePath,
		Description: desc,
		Status:      status,
		Data: domain.ReplayEventData{
			FilePath:  fileOp.FilePath,
			Operation: string(fileOp.Operation),
			Diff:      fileOp.DiffPreview,
		},
	}
}

// terminalCmdToReplayEvent converts a terminal command to a replay event
func (s *ReplayService) terminalCmdToReplayEvent(cmd domain.TerminalCommand) domain.ReplayEvent {
	status := "success"
	if cmd.ExitCode != 0 {
		status = "error"
	}

	var duration int64
	if cmd.CompletedAt != nil {
		duration = cmd.CompletedAt.Sub(cmd.StartedAt).Milliseconds()
	}

	exitCode := int(cmd.ExitCode)
	return domain.ReplayEvent{
		ID:          cmd.ID.String(),
		Type:        domain.ReplayEventTerminalCmd,
		Timestamp:   cmd.StartedAt,
		Duration:    duration,
		Title:       cmd.Command,
		Description: "Terminal command",
		Status:      status,
		Data: domain.ReplayEventData{
			Command:    cmd.Command,
			WorkingDir: cmd.WorkingDirectory,
			ExitCode:   &exitCode,
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
			FileManifest: cp.FilesChanged,
			GitBranch:    cp.GitBranch,
			GitCommit:    cp.GitCommitSha,
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
		Repository:   first.RepoURL,
		Branch:       first.Branch,
		StartCommit:  first.CommitSha,
		FilesChanged: make([]string, 0),
	}

	// Collect all changed files from added, modified, and deleted
	filesMap := make(map[string]bool)
	for _, link := range gitLinks {
		for _, f := range link.FilesAdded {
			filesMap[f] = true
		}
		for _, f := range link.FilesModified {
			filesMap[f] = true
		}
		for _, f := range link.FilesDeleted {
			filesMap[f] = true
		}
		if link.CommitSha != ctx.StartCommit {
			ctx.EndCommit = link.CommitSha
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

// GetTimelineForTrace fetches all data and builds a complete timeline for a trace
func (s *ReplayService) GetTimelineForTrace(
	ctx context.Context,
	projectID uuid.UUID,
	traceID string,
) (*domain.ReplayTimeline, error) {
	// Fetch the trace
	trace, err := s.traceRepo.GetByID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	// Fetch all related data
	observations, err := s.observationRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		s.logger.Warn("failed to get observations for replay", zap.String("traceId", traceID), zap.Error(err))
		observations = []domain.Observation{}
	}

	fileOps, err := s.fileOperationRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		s.logger.Warn("failed to get file operations for replay", zap.String("traceId", traceID), zap.Error(err))
		fileOps = []domain.FileOperation{}
	}

	terminalCmds, err := s.terminalCommandRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		s.logger.Warn("failed to get terminal commands for replay", zap.String("traceId", traceID), zap.Error(err))
		terminalCmds = []domain.TerminalCommand{}
	}

	checkpoints, err := s.checkpointRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		s.logger.Warn("failed to get checkpoints for replay", zap.String("traceId", traceID), zap.Error(err))
		checkpoints = []domain.Checkpoint{}
	}

	gitLinks, err := s.gitLinkRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		s.logger.Warn("failed to get git links for replay", zap.String("traceId", traceID), zap.Error(err))
		gitLinks = []domain.GitLink{}
	}

	// Build the timeline
	return s.BuildTimeline(ctx, trace, observations, fileOps, terminalCmds, checkpoints, gitLinks)
}

// GenerateReproductionScript generates an executable script from a trace
func (s *ReplayService) GenerateReproductionScript(ctx context.Context, projectID uuid.UUID, traceID uuid.UUID, input *domain.ReproductionInput) (*domain.ReproductionScript, error) {
	storedTraceID := strings.ReplaceAll(traceID.String(), "-", "")
	timeline, err := s.GetTimelineForTrace(ctx, projectID, storedTraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}

	script := &domain.ReproductionScript{
		ID:        uuid.New(),
		TraceID:   traceID,
		Format:    input.Format,
		Language:  string(input.Format),
		Config:    input.Config,
		CreatedAt: time.Now(),
	}

	switch input.Format {
	case domain.ReproFormatPython:
		script.Script = s.generatePythonScript(timeline, input.Config)
	case domain.ReproFormatTypeScript:
		script.Script = s.generateTypeScriptScript(timeline, input.Config)
	case domain.ReproFormatJSON:
		script.Script = s.generateJSONExport(timeline)
	default:
		script.Script = s.generateShellScript(timeline, input.Config)
	}

	s.logger.Info("generated reproduction script",
		zap.String("traceId", traceID.String()),
		zap.String("format", string(input.Format)),
	)

	return script, nil
}

// CompareReplays performs A/B comparison of two traces
func (s *ReplayService) CompareReplays(ctx context.Context, projectID uuid.UUID, traceIDA, traceIDB uuid.UUID) (*domain.ReplayComparison, error) {
	timelineA, err := s.GetTimelineForTrace(ctx, projectID, traceIDA.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline A: %w", err)
	}

	timelineB, err := s.GetTimelineForTrace(ctx, projectID, traceIDB.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline B: %w", err)
	}

	comparison := &domain.ReplayComparison{
		ID:        uuid.New(),
		TraceIDA:  traceIDA,
		TraceIDB:  traceIDB,
		TimelineA: *timelineA,
		TimelineB: *timelineB,
		CreatedAt: time.Now(),
	}

	comparison.Differences = s.findDifferences(timelineA, timelineB)
	comparison.Summary = s.calculateComparisonSummary(timelineA, timelineB)

	return comparison, nil
}

func (s *ReplayService) generatePythonScript(timeline *domain.ReplayTimeline, config domain.ReproductionConfig) string {
	var sb strings.Builder
	sb.WriteString("#!/usr/bin/env python3\n")
	sb.WriteString("\"\"\"Auto-generated reproduction script from AgentTrace\"\"\"\n")
	sb.WriteString("# Trace: " + timeline.TraceID.String() + "\n")
	sb.WriteString("# Generated at: " + time.Now().Format(time.RFC3339) + "\n\n")
	sb.WriteString("from agenttrace import AgentTrace\n\n")
	sb.WriteString("client = AgentTrace()\n")
	sb.WriteString("trace = client.trace(name=\"" + timeline.TraceName + "\")\n\n")

	for _, event := range timeline.Events {
		switch event.Type {
		case domain.ReplayEventLLMCall:
			model := event.Data.Model
			if config.SwapModel != "" {
				model = config.SwapModel
			}
			sb.WriteString(fmt.Sprintf("# Step: %s\n", event.Title))
			sb.WriteString(fmt.Sprintf("generation = trace.generation(name=\"%s\", model=\"%s\")\n", event.Title, model))
			sb.WriteString("generation.end()\n\n")
		case domain.ReplayEventToolCall:
			sb.WriteString(fmt.Sprintf("# Tool call: %s\n", event.Data.ToolName))
			sb.WriteString(fmt.Sprintf("span = trace.span(name=\"%s\")\n", event.Data.ToolName))
			sb.WriteString("span.end()\n\n")
		}
	}

	sb.WriteString("trace.end()\nclient.flush()\n")
	sb.WriteString("print(\"Reproduction complete\")\n")
	return sb.String()
}

func (s *ReplayService) generateTypeScriptScript(timeline *domain.ReplayTimeline, config domain.ReproductionConfig) string {
	var sb strings.Builder
	sb.WriteString("#!/usr/bin/env npx ts-node\n")
	sb.WriteString("// Auto-generated reproduction script from AgentTrace\n")
	sb.WriteString("// Trace: " + timeline.TraceID.String() + "\n\n")
	sb.WriteString("import { AgentTrace } from 'agenttrace';\n\n")
	sb.WriteString("const client = new AgentTrace();\n\n")
	sb.WriteString("async function reproduce() {\n")
	sb.WriteString("  const trace = client.trace({ name: \"" + timeline.TraceName + "\" });\n\n")

	for _, event := range timeline.Events {
		if event.Type == domain.ReplayEventLLMCall {
			model := event.Data.Model
			if config.SwapModel != "" {
				model = config.SwapModel
			}
			sb.WriteString(fmt.Sprintf("  // %s\n", event.Title))
			sb.WriteString(fmt.Sprintf("  const gen = trace.generation({ name: \"%s\", model: \"%s\" });\n", event.Title, model))
			sb.WriteString("  gen.end();\n\n")
		}
	}

	sb.WriteString("  trace.end();\n")
	sb.WriteString("  await client.flush();\n")
	sb.WriteString("  console.log('Reproduction complete');\n")
	sb.WriteString("}\n\nreproduce();\n")
	return sb.String()
}

func (s *ReplayService) generateShellScript(timeline *domain.ReplayTimeline, config domain.ReproductionConfig) string {
	var sb strings.Builder
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Auto-generated reproduction script from AgentTrace\n")
	sb.WriteString("# Trace: " + timeline.TraceID.String() + "\n\n")
	sb.WriteString("echo \"Starting reproduction...\"\n\n")

	for _, event := range timeline.Events {
		if event.Type == domain.ReplayEventTerminalCmd && event.Data.Command != "" {
			sb.WriteString(fmt.Sprintf("# %s\n", event.Title))
			sb.WriteString(fmt.Sprintf("echo \"Executing: %s\"\n", event.Data.Command))
			sb.WriteString(event.Data.Command + "\n\n")
		}
	}

	sb.WriteString("echo \"Reproduction complete\"\n")
	return sb.String()
}

func (s *ReplayService) generateJSONExport(timeline *domain.ReplayTimeline) string {
	data, _ := json.Marshal(domain.ReplayExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Timeline:   *timeline,
	})
	return string(data)
}

func (s *ReplayService) findDifferences(a, b *domain.ReplayTimeline) []domain.ReplayDifference {
	var diffs []domain.ReplayDifference

	maxLen := len(a.Events)
	if len(b.Events) > maxLen {
		maxLen = len(b.Events)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(a.Events) {
			diffs = append(diffs, domain.ReplayDifference{
				EventIndexA: -1,
				EventIndexB: i,
				Type:        "added",
				Description: fmt.Sprintf("Event '%s' only in trace B", b.Events[i].Title),
				Impact:      "medium",
			})
		} else if i >= len(b.Events) {
			diffs = append(diffs, domain.ReplayDifference{
				EventIndexA: i,
				EventIndexB: -1,
				Type:        "removed",
				Description: fmt.Sprintf("Event '%s' only in trace A", a.Events[i].Title),
				Impact:      "medium",
			})
		} else if a.Events[i].Type != b.Events[i].Type {
			diffs = append(diffs, domain.ReplayDifference{
				EventIndexA: i,
				EventIndexB: i,
				Type:        "modified",
				Description: fmt.Sprintf("Event type changed: %s → %s", a.Events[i].Type, b.Events[i].Type),
				Impact:      "high",
			})
		}
	}

	return diffs
}

func (s *ReplayService) calculateComparisonSummary(a, b *domain.ReplayTimeline) domain.ComparisonSummary {
	summary := domain.ComparisonSummary{
		CostDelta:       b.Summary.TotalCost - a.Summary.TotalCost,
		LatencyDelta:    b.Duration - a.Duration,
		TokenDelta:      b.Summary.TotalTokens - a.Summary.TotalTokens,
		EventCountDelta: b.Summary.TotalEvents - a.Summary.TotalEvents,
	}

	if a.Summary.TotalCost > 0 {
		summary.CostDeltaPct = summary.CostDelta / a.Summary.TotalCost * 100
	}
	if a.Duration > 0 {
		summary.LatencyDeltaPct = float64(summary.LatencyDelta) / float64(a.Duration) * 100
	}

	aBetter, bBetter := 0, 0
	if summary.CostDelta < 0 {
		bBetter++
	} else if summary.CostDelta > 0 {
		aBetter++
	}
	if summary.LatencyDelta < 0 {
		bBetter++
	} else if summary.LatencyDelta > 0 {
		aBetter++
	}

	if bBetter > aBetter {
		summary.Verdict = "b_better"
	} else if aBetter > bBetter {
		summary.Verdict = "a_better"
	} else {
		summary.Verdict = "equivalent"
	}

	return summary
}

// ExportTimeline exports a timeline in a portable format
func (s *ReplayService) ExportTimeline(timeline *domain.ReplayTimeline) *domain.ReplayExport {
	return &domain.ReplayExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Timeline:   *timeline,
	}
}
