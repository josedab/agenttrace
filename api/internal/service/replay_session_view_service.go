package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// BuildUnifiedTimeline transforms real replay events into the unified view model.
func (s *ReplaySessionService) BuildUnifiedTimeline(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
) ([]domain.UnifiedTimelineEvent, error) {
	timeline, err := s.GetTimeline(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.UnifiedTimelineEvent, 0, len(timeline.Events))
	for _, event := range timeline.Events {
		endTime := event.Timestamp.Add(time.Duration(event.DurationMs) * time.Millisecond)
		unified := domain.UnifiedTimelineEvent{
			ID:          uuid.NewSHA1(uuid.NameSpaceOID, []byte(event.ID.String())),
			SessionID:   sessionID,
			EventType:   string(event.Type),
			Category:    replayEventCategory(event.Type),
			Title:       replayEventTitle(event),
			Description: stringValue(event.Data["description"]),
			StartTime:   event.Timestamp,
			EndTime:     &endTime,
			DurationMs:  event.DurationMs,
			Metadata:    event.Data,
			Status:      stringValueOr(event.Data["status"], "success"),
			Model:       stringValue(event.Data["model"]),
			CostUSD:     floatValue(event.Data["cost"]),
			TokensUsed:  int(floatValue(event.Data["tokensInput"]) + floatValue(event.Data["tokensOutput"])),
		}
		if event.FileDelta != nil {
			unified.FileDelta = &domain.UnifiedFileDelta{
				FilePath:    event.FileDelta.Path,
				ChangeType:  event.FileDelta.Operation,
				DiffPreview: event.FileDelta.DiffPatch,
			}
		}
		result = append(result, unified)
	}
	return result, nil
}

// GetReplaySnapshot reconstructs deterministic state up to an event index.
func (s *ReplaySessionService) GetReplaySnapshot(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
	eventIndex int,
) (*domain.ReplaySnapshot, error) {
	timeline, err := s.GetTimeline(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}
	if eventIndex < 0 || eventIndex >= len(timeline.Events) {
		return nil, apperrors.Validation("event index is out of range")
	}

	snapshot := &domain.ReplaySnapshot{
		EventIndex:  eventIndex,
		Timestamp:   timeline.Events[eventIndex].Timestamp,
		FileStates:  map[string]string{},
		EventCounts: map[string]int{},
	}
	for index := 0; index <= eventIndex; index++ {
		event := timeline.Events[index]
		snapshot.ElapsedMs += event.DurationMs
		snapshot.EventCounts[string(event.Type)]++
		snapshot.TotalCost += floatValue(event.Data["cost"])
		snapshot.TotalTokens += int(
			floatValue(event.Data["tokensInput"]) + floatValue(event.Data["tokensOutput"]),
		)
		if model := stringValue(event.Data["model"]); model != "" {
			snapshot.ActiveModel = model
		}
		applyFileDelta(snapshot.FileStates, event.FileDelta)
	}
	return snapshot, nil
}

// AddAnnotation validates session and event ownership before returning an annotation.
func (s *ReplaySessionService) AddAnnotation(
	ctx context.Context,
	projectID, sessionID, userID uuid.UUID,
	input *domain.ReplayAnnotationInput,
) (*domain.ReplayAnnotation, error) {
	if input == nil || input.Content == "" {
		return nil, apperrors.Validation("annotation content is required")
	}
	timeline, err := s.GetTimeline(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}
	found := false
	for _, event := range timeline.Events {
		if event.ID == input.EventID {
			found = true
			break
		}
	}
	if !found {
		return nil, apperrors.NotFound("replay event")
	}

	return &domain.ReplayAnnotation{
		ID:        uuid.New(),
		UserID:    userID,
		Content:   input.Content,
		EventID:   input.EventID,
		Timestamp: s.clock().UTC(),
	}, nil
}

// GetFileStateAt reconstructs the file state at a given event index.
func (s *ReplaySessionService) GetFileStateAt(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
	eventIndex int,
) (*domain.ReplayFileStateSnapshot, error) {
	timeline, err := s.GetTimeline(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}
	if eventIndex < 0 || eventIndex >= len(timeline.Events) {
		return nil, apperrors.Validation("event index is out of range")
	}

	files := make(map[string]string)
	for index := 0; index <= eventIndex; index++ {
		applyFileDelta(files, timeline.Events[index].FileDelta)
	}
	return &domain.ReplayFileStateSnapshot{
		SessionID:  sessionID,
		EventIndex: eventIndex,
		Files:      files,
		Timestamp:  timeline.Events[eventIndex].Timestamp,
	}, nil
}

func replayEventsToSessionEvents(
	sessionID uuid.UUID,
	events []domain.ReplayEvent,
) []domain.AgentReplayTimelineEvent {
	result := make([]domain.AgentReplayTimelineEvent, 0, len(events))
	for index, event := range events {
		data := replayEventDataMap(event.Data)
		data["title"] = event.Title
		data["description"] = event.Description
		data["status"] = event.Status

		var fileDelta *domain.ReplayFileDelta
		if event.Type == domain.ReplayEventFileOperation {
			fileDelta = &domain.ReplayFileDelta{
				Path:      event.Data.FilePath,
				Operation: event.Data.Operation,
				DiffPatch: event.Data.Diff,
				Before:    event.Data.ContentBefore,
				After:     event.Data.ContentAfter,
			}
		}

		result = append(result, domain.AgentReplayTimelineEvent{
			ID:         uuid.NewSHA1(uuid.NameSpaceOID, []byte(event.ID)),
			SessionID:  sessionID,
			Index:      index,
			Type:       event.Type,
			Timestamp:  event.Timestamp,
			Data:       data,
			Input:      event.Data.Input,
			Output:     event.Data.Output,
			DurationMs: event.Duration,
			FileDelta:  fileDelta,
		})
	}
	return result
}

func replayEventDataMap(data domain.ReplayEventData) map[string]interface{} {
	return map[string]interface{}{
		"model":        data.Model,
		"input":        data.Input,
		"output":       data.Output,
		"tokensInput":  data.TokensInput,
		"tokensOutput": data.TokensOutput,
		"cost":         data.Cost,
		"toolName":     data.ToolName,
		"arguments":    data.Arguments,
		"result":       data.Result,
		"filePath":     data.FilePath,
		"operation":    data.Operation,
		"diff":         data.Diff,
		"command":      data.Command,
		"workingDir":   data.WorkingDir,
		"exitCode":     data.ExitCode,
		"stdout":       data.Stdout,
		"stderr":       data.Stderr,
		"checkpointId": data.CheckpointID,
		"fileManifest": data.FileManifest,
		"gitCommit":    data.GitCommit,
		"gitBranch":    data.GitBranch,
		"gitMessage":   data.GitMessage,
		"changedFiles": data.ChangedFiles,
		"error":        data.Error,
		"errorType":    data.ErrorType,
		"stackTrace":   data.StackTrace,
	}
}

func replayBranches(session domain.AgentReplaySession) []domain.AgentReplayBranch {
	if session.ParentSessionID == nil {
		return []domain.AgentReplayBranch{}
	}
	return []domain.AgentReplayBranch{{
		SessionID:  session.ID,
		EventIndex: session.BranchPoint,
		Name:       session.Name,
		CreatedAt:  session.CreatedAt,
	}}
}

func replayEventCategory(eventType domain.ReplayEventType) string {
	switch eventType {
	case domain.ReplayEventUserInput:
		return "user_input"
	case domain.ReplayEventCheckpoint, domain.ReplayEventGitOperation:
		return "system"
	default:
		return "agent_action"
	}
}

func replayEventTitle(event domain.AgentReplayTimelineEvent) string {
	if title := stringValue(event.Data["title"]); title != "" {
		return title
	}
	return string(event.Type)
}

func applyFileDelta(files map[string]string, delta *domain.ReplayFileDelta) {
	if delta == nil || delta.Path == "" {
		return
	}
	switch delta.Operation {
	case "create", "write", "update", "edit":
		files[delta.Path] = delta.After
	case "delete":
		delete(files, delta.Path)
	}
}

func stringValue(value interface{}) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func stringValueOr(value interface{}, fallback string) string {
	if result := stringValue(value); result != "" {
		return result
	}
	return fallback
}

func floatValue(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case uint64:
		return float64(typed)
	default:
		return 0
	}
}
