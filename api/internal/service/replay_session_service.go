package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ReplaySessionRepository defines persistent replay session operations.
type ReplaySessionRepository interface {
	Save(ctx context.Context, session *domain.AgentReplaySession) error
	Update(ctx context.Context, session *domain.AgentReplaySession) error
	GetByID(
		ctx context.Context,
		projectID, id uuid.UUID,
	) (*domain.AgentReplaySession, error)
	List(
		ctx context.Context,
		projectID uuid.UUID,
		limit int,
	) ([]domain.AgentReplaySession, error)
	SaveEvent(ctx context.Context, event *domain.AgentReplayTimelineEvent) error
	ListEvents(
		ctx context.Context,
		sessionID uuid.UUID,
	) ([]domain.AgentReplayTimelineEvent, error)
}

// ReplaySessionService persists replay sessions and derives their timelines from real trace data.
type ReplaySessionService struct {
	logger           *zap.Logger
	repository       ReplaySessionRepository
	traceRepository  ReplayPlanTraceRepository
	timelineProvider ReplayTimelineProvider
	clock            func() time.Time
}

// NewReplaySessionService creates a repository-backed replay session service.
func NewReplaySessionService(
	logger *zap.Logger,
	repository ReplaySessionRepository,
	traceRepository ReplayPlanTraceRepository,
	timelineProvider ReplayTimelineProvider,
) *ReplaySessionService {
	return &ReplaySessionService{
		logger:           logger,
		repository:       repository,
		traceRepository:  traceRepository,
		timelineProvider: timelineProvider,
		clock:            time.Now,
	}
}

// CreateSession creates a replay session from an authorized trace.
func (s *ReplaySessionService) CreateSession(
	ctx context.Context,
	projectID uuid.UUID,
	userID uuid.UUID,
	input *domain.AgentReplaySessionInput,
) (*domain.AgentReplaySession, error) {
	if input == nil {
		return nil, apperrors.Validation("replay session input is required")
	}
	traceID := input.TraceID.String()
	if _, err := s.traceRepository.GetByID(ctx, projectID, traceID); err != nil {
		return nil, err
	}

	fidelity := input.RecordingFidelity
	if fidelity == "" {
		fidelity = domain.AgentReplayFidelityStandard
	}
	if fidelity != domain.AgentReplayFidelityFull &&
		fidelity != domain.AgentReplayFidelityStandard &&
		fidelity != domain.AgentReplayFidelityMinimal {
		return nil, apperrors.Validation("invalid recording fidelity")
	}

	timeline, err := s.timelineProvider.GetTimelineForTrace(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("build replay session timeline: %w", err)
	}

	now := s.clock().UTC()
	status := domain.AgentReplayRecording
	var endedAt *time.Time
	if timeline.Summary.TotalEvents > 0 {
		status = domain.AgentReplayCompleted
		ended := now
		endedAt = &ended
	}

	session := &domain.AgentReplaySession{
		ID:                uuid.New(),
		ProjectID:         projectID,
		TraceID:           input.TraceID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            status,
		RecordingFidelity: fidelity,
		TotalEvents:       timeline.Summary.TotalEvents,
		TotalDurationMs:   timeline.Duration,
		FilesTracked:      timeline.Summary.FileOperations,
		CheckpointCount:   timeline.Summary.Checkpoints,
		IsPublic:          false,
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         userID,
		EndedAt:           endedAt,
	}
	if err := s.repository.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("persist replay session: %w", err)
	}

	s.logger.Info("created replay session",
		zap.String("sessionId", session.ID.String()),
		zap.String("traceId", input.TraceID.String()),
		zap.String("projectId", projectID.String()),
	)
	return session, nil
}

// RecordEvent records a single event in a project-scoped replay session.
func (s *ReplaySessionService) RecordEvent(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
	eventType domain.ReplayEventType,
	data map[string]interface{},
	durationMs int64,
) (*domain.AgentReplayTimelineEvent, error) {
	events, err := s.RecordEvents(ctx, projectID, sessionID, []domain.AgentReplayRecordEventInput{{
		Type:       eventType,
		Data:       data,
		DurationMs: durationMs,
	}})
	if err != nil {
		return nil, err
	}
	return &events[0], nil
}

// CompleteSession marks a replay session as completed.
func (s *ReplaySessionService) CompleteSession(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
) error {
	session, err := s.repository.GetByID(ctx, projectID, sessionID)
	if err != nil {
		return err
	}
	if session.Status == domain.AgentReplayFailed {
		return apperrors.Conflict("failed replay sessions cannot be completed")
	}

	now := s.clock().UTC()
	session.Status = domain.AgentReplayCompleted
	session.EndedAt = &now
	session.UpdatedAt = now
	if err := s.repository.Update(ctx, session); err != nil {
		return fmt.Errorf("complete replay session: %w", err)
	}
	return nil
}

// GetTimeline returns persisted events, or derives them from the source trace when none were recorded.
func (s *ReplaySessionService) GetTimeline(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
) (*domain.AgentReplayFullTimeline, error) {
	session, err := s.repository.GetByID(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}
	events, err := s.repository.ListEvents(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list replay session events: %w", err)
	}
	if len(events) == 0 {
		timeline, err := s.timelineProvider.GetTimelineForTrace(
			ctx,
			projectID,
			session.TraceID.String(),
		)
		if err != nil {
			return nil, fmt.Errorf("build replay timeline: %w", err)
		}
		events = replayEventsToSessionEvents(sessionID, timeline.Events)
	}

	return &domain.AgentReplayFullTimeline{
		Session:    *session,
		Events:     events,
		Branches:   replayBranches(*session),
		Milestones: s.DetectMilestones(events),
	}, nil
}

// BranchSession creates a persisted branch at a valid event index.
func (s *ReplaySessionService) BranchSession(
	ctx context.Context,
	projectID uuid.UUID,
	userID uuid.UUID,
	req *domain.AgentReplayBranchRequest,
) (*domain.AgentReplaySession, error) {
	if req == nil {
		return nil, apperrors.Validation("branch request is required")
	}
	parent, err := s.repository.GetByID(ctx, projectID, req.SessionID)
	if err != nil {
		return nil, err
	}
	timeline, err := s.GetTimeline(ctx, projectID, req.SessionID)
	if err != nil {
		return nil, err
	}
	if req.EventIndex < 0 || req.EventIndex >= len(timeline.Events) {
		return nil, apperrors.Validation("branch event index is out of range")
	}

	now := s.clock().UTC()
	branch := &domain.AgentReplaySession{
		ID:                uuid.New(),
		ProjectID:         projectID,
		TraceID:           parent.TraceID,
		Name:              req.Name,
		Status:            domain.AgentReplayCompleted,
		RecordingFidelity: parent.RecordingFidelity,
		TotalEvents:       req.EventIndex + 1,
		TotalDurationMs:   timeline.Events[req.EventIndex].Timestamp.Sub(parent.CreatedAt).Milliseconds(),
		ParentSessionID:   &parent.ID,
		BranchPoint:       req.EventIndex,
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         userID,
		EndedAt:           &now,
	}
	if branch.TotalDurationMs < 0 {
		branch.TotalDurationMs = 0
	}
	if err := s.repository.Save(ctx, branch); err != nil {
		return nil, fmt.Errorf("persist replay branch: %w", err)
	}
	return branch, nil
}

// GetPlaybackState derives playback bounds from the real timeline.
func (s *ReplaySessionService) GetPlaybackState(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
) (*domain.AgentReplayPlaybackState, error) {
	timeline, err := s.GetTimeline(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}
	return &domain.AgentReplayPlaybackState{
		SessionID:   sessionID,
		TotalEvents: len(timeline.Events),
		IsPlaying:   false,
		Speed:       1,
		TotalMs:     timeline.Session.TotalDurationMs,
	}, nil
}

// ShareSession is retained for compatibility; secure token sharing is owned by ShareLinkService.
func (s *ReplaySessionService) ShareSession(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
) (string, error) {
	if _, err := s.repository.GetByID(ctx, projectID, sessionID); err != nil {
		return "", err
	}
	return "", apperrors.Unprocessable(
		"legacy replay sharing is disabled; create a redacted share link instead",
	)
}

// DetectMilestones identifies checkpoints and errors in a replay.
func (s *ReplaySessionService) DetectMilestones(
	events []domain.AgentReplayTimelineEvent,
) []domain.AgentReplayMilestone {
	milestones := make([]domain.AgentReplayMilestone, 0)
	for _, event := range events {
		switch event.Type {
		case domain.ReplayEventCheckpoint:
			milestones = append(milestones, domain.AgentReplayMilestone{
				EventIndex: event.Index,
				Label:      "Checkpoint",
				Type:       "checkpoint",
			})
		case domain.ReplayEventError:
			milestones = append(milestones, domain.AgentReplayMilestone{
				EventIndex: event.Index,
				Label:      "Error",
				Type:       "error",
			})
		}
	}
	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].EventIndex < milestones[j].EventIndex
	})
	return milestones
}

// ListSessions lists persisted replay sessions for one project.
func (s *ReplaySessionService) ListSessions(
	ctx context.Context,
	filter domain.AgentReplaySessionFilter,
) (*domain.AgentReplaySessionList, error) {
	sessions, err := s.repository.List(ctx, filter.ProjectID, 100)
	if err != nil {
		return nil, fmt.Errorf("list replay sessions: %w", err)
	}

	filtered := make([]domain.AgentReplaySession, 0, len(sessions))
	for _, session := range sessions {
		if filter.TraceID != nil && session.TraceID != *filter.TraceID {
			continue
		}
		if filter.Status != nil && session.Status != *filter.Status {
			continue
		}
		if filter.IsPublic != nil && session.IsPublic != *filter.IsPublic {
			continue
		}
		filtered = append(filtered, session)
	}

	return &domain.AgentReplaySessionList{
		Sessions:   filtered,
		TotalCount: int64(len(filtered)),
		HasMore:    false,
	}, nil
}

// GetSession retrieves a session only within the authorized project.
func (s *ReplaySessionService) GetSession(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
) (*domain.AgentReplaySession, error) {
	return s.repository.GetByID(ctx, projectID, sessionID)
}

// RecordEvents appends persisted replay events with stable indexes.
func (s *ReplaySessionService) RecordEvents(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
	inputs []domain.AgentReplayRecordEventInput,
) ([]domain.AgentReplayTimelineEvent, error) {
	if len(inputs) == 0 {
		return nil, apperrors.Validation("at least one event is required")
	}
	session, err := s.repository.GetByID(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}
	existing, err := s.repository.ListEvents(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list existing replay events: %w", err)
	}

	now := s.clock().UTC()
	events := make([]domain.AgentReplayTimelineEvent, 0, len(inputs))
	for index, input := range inputs {
		event := domain.AgentReplayTimelineEvent{
			ID:         uuid.New(),
			SessionID:  sessionID,
			Index:      len(existing) + index,
			Type:       input.Type,
			Timestamp:  now.Add(time.Duration(index) * time.Nanosecond),
			Data:       input.Data,
			Input:      input.Input,
			Output:     input.Output,
			DurationMs: input.DurationMs,
			FileDelta:  input.FileDelta,
		}
		if event.Data == nil {
			event.Data = map[string]interface{}{}
		}
		if err := s.repository.SaveEvent(ctx, &event); err != nil {
			return nil, fmt.Errorf("persist replay event: %w", err)
		}
		events = append(events, event)
	}

	session.Status = domain.AgentReplayRecording
	session.TotalEvents = len(existing) + len(events)
	session.UpdatedAt = now
	session.EndedAt = nil
	if err := s.repository.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("update replay session event count: %w", err)
	}
	return events, nil
}

// ControlPlayback validates playback commands against the real timeline.
func (s *ReplaySessionService) ControlPlayback(
	ctx context.Context,
	projectID, sessionID uuid.UUID,
	command *domain.ReplayControlCommand,
) (*domain.AgentReplayPlaybackState, error) {
	if command == nil {
		return nil, apperrors.Validation("control command is required")
	}
	state, err := s.GetPlaybackState(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}

	switch command.Action {
	case "play":
		state.IsPlaying = true
	case "pause":
		state.IsPlaying = false
	case "seek":
		if command.EventIndex == nil ||
			*command.EventIndex < 0 ||
			*command.EventIndex >= state.TotalEvents {
			return nil, apperrors.Validation("seek event index is out of range")
		}
		state.CurrentIndex = *command.EventIndex
	case "speed":
		if command.Speed < 0.5 || command.Speed > 16 {
			return nil, apperrors.Validation("speed must be between 0.5 and 16")
		}
		state.Speed = command.Speed
	case "step_forward":
		if state.TotalEvents > 0 {
			state.CurrentIndex = 1
			if state.CurrentIndex >= state.TotalEvents {
				state.CurrentIndex = state.TotalEvents - 1
			}
		}
	case "step_backward":
		state.CurrentIndex = 0
	default:
		return nil, apperrors.Validation("unknown playback action")
	}
	return state, nil
}
