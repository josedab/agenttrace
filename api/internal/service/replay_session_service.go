package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ReplaySessionService handles agent replay session logic
type ReplaySessionService struct {
	logger *zap.Logger
}

// NewReplaySessionService creates a new replay session service
func NewReplaySessionService(logger *zap.Logger) *ReplaySessionService {
	return &ReplaySessionService{logger: logger}
}

// CreateSession creates a new replay session from a trace
func (s *ReplaySessionService) CreateSession(
	ctx context.Context,
	projectID uuid.UUID,
	userID uuid.UUID,
	input *domain.AgentReplaySessionInput,
) (*domain.AgentReplaySession, error) {
	fidelity := domain.AgentReplayFidelityStandard
	if input.RecordingFidelity != "" {
		fidelity = input.RecordingFidelity
	}

	session := &domain.AgentReplaySession{
		ID:                uuid.New(),
		ProjectID:         projectID,
		TraceID:           input.TraceID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            domain.AgentReplayRecording,
		RecordingFidelity: fidelity,
		IsPublic:          false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	s.logger.Info("created replay session",
		zap.String("sessionId", session.ID.String()),
		zap.String("traceId", input.TraceID.String()),
	)
	return session, nil
}

// RecordEvent records a new event in a replay session
func (s *ReplaySessionService) RecordEvent(
	ctx context.Context,
	sessionID uuid.UUID,
	eventType domain.ReplayEventType,
	data map[string]interface{},
	durationMs int64,
) (*domain.AgentReplayTimelineEvent, error) {
	event := &domain.AgentReplayTimelineEvent{
		ID:         uuid.New(),
		SessionID:  sessionID,
		Type:       eventType,
		Timestamp:  time.Now(),
		Data:       data,
		DurationMs: durationMs,
	}
	return event, nil
}

// CompleteSession marks a replay session as completed
func (s *ReplaySessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	s.logger.Info("completed replay session", zap.String("sessionId", sessionID.String()))
	return nil
}

// GetTimeline returns the full replay timeline for a session
func (s *ReplaySessionService) GetTimeline(
	ctx context.Context,
	sessionID uuid.UUID,
) (*domain.AgentReplayFullTimeline, error) {
	session := &domain.AgentReplaySession{
		ID:        sessionID,
		Status:    domain.AgentReplayCompleted,
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	return &domain.AgentReplayFullTimeline{
		Session:    *session,
		Events:     []domain.AgentReplayTimelineEvent{},
		Branches:   []domain.AgentReplayBranch{},
		Milestones: []domain.AgentReplayMilestone{},
	}, nil
}

// BranchSession creates a new session branching from an existing one
func (s *ReplaySessionService) BranchSession(
	ctx context.Context,
	projectID uuid.UUID,
	userID uuid.UUID,
	req *domain.AgentReplayBranchRequest,
) (*domain.AgentReplaySession, error) {
	branch := &domain.AgentReplaySession{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            req.Name,
		Status:          domain.AgentReplayRecording,
		ParentSessionID: &req.SessionID,
		BranchPoint:     req.EventIndex,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CreatedBy:       userID,
	}

	s.logger.Info("branched replay session",
		zap.String("parentId", req.SessionID.String()),
		zap.String("branchId", branch.ID.String()),
		zap.Int("branchPoint", req.EventIndex),
	)
	return branch, nil
}

// GetPlaybackState returns the current playback state
func (s *ReplaySessionService) GetPlaybackState(
	ctx context.Context,
	sessionID uuid.UUID,
) (*domain.AgentReplayPlaybackState, error) {
	return &domain.AgentReplayPlaybackState{
		SessionID:    sessionID,
		CurrentIndex: 0,
		TotalEvents:  0,
		IsPlaying:    false,
		Speed:        1.0,
	}, nil
}

// ShareSession generates a public share URL
func (s *ReplaySessionService) ShareSession(ctx context.Context, sessionID uuid.UUID) (string, error) {
	shareURL := fmt.Sprintf("/replay/%s/shared", sessionID.String())
	s.logger.Info("shared replay session", zap.String("sessionId", sessionID.String()))
	return shareURL, nil
}

// DetectMilestones identifies important moments in a replay
func (s *ReplaySessionService) DetectMilestones(events []domain.AgentReplayTimelineEvent) []domain.AgentReplayMilestone {
	var milestones []domain.AgentReplayMilestone

	for _, event := range events {
		switch event.Type {
		case domain.ReplayEventCheckpoint:
			milestones = append(milestones, domain.AgentReplayMilestone{
				EventIndex: event.Index, Label: "Checkpoint", Type: "checkpoint",
			})
		case domain.ReplayEventError:
			milestones = append(milestones, domain.AgentReplayMilestone{
				EventIndex: event.Index, Label: "Error", Type: "error",
			})
		}
	}

	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].EventIndex < milestones[j].EventIndex
	})
	return milestones
}

// ListSessions lists replay sessions for a project
func (s *ReplaySessionService) ListSessions(
	ctx context.Context,
	filter domain.AgentReplaySessionFilter,
) (*domain.AgentReplaySessionList, error) {
	return &domain.AgentReplaySessionList{
		Sessions:   []domain.AgentReplaySession{},
		TotalCount: 0,
		HasMore:    false,
	}, nil
}

// GetSession returns a specific replay session
func (s *ReplaySessionService) GetSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (*domain.AgentReplaySession, error) {
	return nil, fmt.Errorf("replay session not found: %s", sessionID.String())
}

// RecordEvents records a batch of events in a replay session
func (s *ReplaySessionService) RecordEvents(
	ctx context.Context,
	sessionID uuid.UUID,
	inputs []domain.AgentReplayRecordEventInput,
) ([]domain.AgentReplayTimelineEvent, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no events to record")
	}

	events := make([]domain.AgentReplayTimelineEvent, 0, len(inputs))
	for i, input := range inputs {
		event := domain.AgentReplayTimelineEvent{
			ID:         uuid.New(),
			SessionID:  sessionID,
			Index:      i,
			Type:       input.Type,
			Timestamp:  time.Now(),
			Data:       input.Data,
			Input:      input.Input,
			Output:     input.Output,
			DurationMs: input.DurationMs,
			FileDelta:  input.FileDelta,
		}
		events = append(events, event)
	}

	s.logger.Info("recorded replay events",
		zap.String("sessionId", sessionID.String()),
		zap.Int("eventCount", len(events)),
	)
	return events, nil
}

// ControlPlayback handles playback control commands
func (s *ReplaySessionService) ControlPlayback(
	ctx context.Context,
	sessionID uuid.UUID,
	cmd *domain.ReplayControlCommand,
) (*domain.AgentReplayPlaybackState, error) {
	if cmd == nil {
		return nil, fmt.Errorf("control command is required")
	}

	state := &domain.AgentReplayPlaybackState{
		SessionID: sessionID,
		Speed:     1.0,
	}

	switch cmd.Action {
	case "play":
		state.IsPlaying = true
	case "pause":
		state.IsPlaying = false
	case "seek":
		if cmd.EventIndex != nil {
			state.CurrentIndex = *cmd.EventIndex
		}
	case "speed":
		if cmd.Speed > 0 && cmd.Speed <= 16 {
			state.Speed = cmd.Speed
		}
	case "step_forward":
		state.IsPlaying = false
		state.CurrentIndex++
	case "step_backward":
		state.IsPlaying = false
		if state.CurrentIndex > 0 {
			state.CurrentIndex--
		}
	default:
		return nil, fmt.Errorf("unknown control action: %s", cmd.Action)
	}

	s.logger.Debug("playback control",
		zap.String("sessionId", sessionID.String()),
		zap.String("action", cmd.Action),
	)
	return state, nil
}

// GetFileStateAt reconstructs the file system state at a given event index
func (s *ReplaySessionService) GetFileStateAt(
	ctx context.Context,
	sessionID uuid.UUID,
	eventIndex int,
) (*domain.ReplayFileStateSnapshot, error) {
	timeline, err := s.GetTimeline(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}

	files := make(map[string]string)
	maxIdx := eventIndex
	if maxIdx >= len(timeline.Events) {
		maxIdx = len(timeline.Events) - 1
	}

	// Reconstruct file state by replaying file deltas up to the target index
	for i := 0; i <= maxIdx; i++ {
		event := timeline.Events[i]
		if event.FileDelta != nil {
			switch event.FileDelta.Operation {
			case "create", "write":
				files[event.FileDelta.Path] = event.FileDelta.After
			case "delete":
				delete(files, event.FileDelta.Path)
			}
		}
	}

	return &domain.ReplayFileStateSnapshot{
		SessionID:  sessionID,
		EventIndex: eventIndex,
		Files:      files,
		Timestamp:  time.Now(),
	}, nil
}
