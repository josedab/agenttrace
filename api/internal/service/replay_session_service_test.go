package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type replaySessionRepositoryStub struct {
	sessions map[uuid.UUID]*domain.AgentReplaySession
	events   map[uuid.UUID][]domain.AgentReplayTimelineEvent
}

func (r *replaySessionRepositoryStub) Save(
	_ context.Context,
	session *domain.AgentReplaySession,
) error {
	copy := *session
	r.sessions[session.ID] = &copy
	return nil
}

func (r *replaySessionRepositoryStub) Update(
	_ context.Context,
	session *domain.AgentReplaySession,
) error {
	copy := *session
	r.sessions[session.ID] = &copy
	return nil
}

func (r *replaySessionRepositoryStub) GetByID(
	_ context.Context,
	projectID, id uuid.UUID,
) (*domain.AgentReplaySession, error) {
	session, ok := r.sessions[id]
	if !ok || session.ProjectID != projectID {
		return nil, apperrors.NotFound("replay session")
	}
	copy := *session
	return &copy, nil
}

func (r *replaySessionRepositoryStub) List(
	_ context.Context,
	projectID uuid.UUID,
	_ int,
) ([]domain.AgentReplaySession, error) {
	result := []domain.AgentReplaySession{}
	for _, session := range r.sessions {
		if session.ProjectID == projectID {
			result = append(result, *session)
		}
	}
	return result, nil
}

func (r *replaySessionRepositoryStub) SaveEvent(
	_ context.Context,
	event *domain.AgentReplayTimelineEvent,
) error {
	r.events[event.SessionID] = append(r.events[event.SessionID], *event)
	return nil
}

func (r *replaySessionRepositoryStub) ListEvents(
	_ context.Context,
	sessionID uuid.UUID,
) ([]domain.AgentReplayTimelineEvent, error) {
	return append([]domain.AgentReplayTimelineEvent(nil), r.events[sessionID]...), nil
}

func newReplaySessionTestService(
	projectID uuid.UUID,
	traceID string,
	timeline *domain.ReplayTimeline,
) (*ReplaySessionService, *replaySessionRepositoryStub) {
	repository := &replaySessionRepositoryStub{
		sessions: map[uuid.UUID]*domain.AgentReplaySession{},
		events:   map[uuid.UUID][]domain.AgentReplayTimelineEvent{},
	}
	service := NewReplaySessionService(
		zap.NewNop(),
		repository,
		&replayPlanTraceRepositoryStub{
			projectID: projectID,
			trace:     &domain.Trace{ID: traceID, ProjectID: projectID},
		},
		&replayTimelineProviderStub{timeline: timeline},
	)
	service.clock = func() time.Time {
		return time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	}
	return service, repository
}

func TestReplaySessionCreatesFromRealTimeline(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New()
	service, repository := newReplaySessionTestService(
		projectID,
		traceID.String(),
		&domain.ReplayTimeline{
			Duration: 1200,
			Summary: domain.ReplaySummary{
				TotalEvents:    3,
				FileOperations: 1,
				Checkpoints:    1,
			},
		},
	)

	session, err := service.CreateSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplaySessionInput{TraceID: traceID, Name: "Real trace replay"},
	)

	require.NoError(t, err)
	assert.Equal(t, domain.AgentReplayCompleted, session.Status)
	assert.Equal(t, 3, session.TotalEvents)
	assert.Equal(t, int64(1200), session.TotalDurationMs)
	assert.Equal(t, 1, session.FilesTracked)
	assert.NotNil(t, repository.sessions[session.ID])
}

func TestReplaySessionEnforcesProjectIsolation(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New()
	service, _ := newReplaySessionTestService(projectID, traceID.String(), &domain.ReplayTimeline{})
	session, err := service.CreateSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplaySessionInput{TraceID: traceID, Name: "Scoped"},
	)
	require.NoError(t, err)

	_, err = service.GetSession(context.Background(), uuid.New(), session.ID)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReplaySessionTimelineUsesRecordedTraceEvents(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New()
	eventTime := time.Date(2026, 7, 25, 9, 0, 0, 0, time.UTC)
	service, _ := newReplaySessionTestService(
		projectID,
		traceID.String(),
		&domain.ReplayTimeline{
			Events: []domain.ReplayEvent{
				{
					ID:        "generation-1",
					Type:      domain.ReplayEventLLMCall,
					Timestamp: eventTime,
					Title:     "Generate patch",
					Status:    "success",
					Data: domain.ReplayEventData{
						Model:        "gpt-4.1",
						TokensInput:  10,
						TokensOutput: 5,
					},
				},
			},
			Summary: domain.ReplaySummary{TotalEvents: 1, LLMCalls: 1},
		},
	)
	session, err := service.CreateSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplaySessionInput{TraceID: traceID, Name: "Timeline"},
	)
	require.NoError(t, err)

	timeline, err := service.GetTimeline(context.Background(), projectID, session.ID)

	require.NoError(t, err)
	require.Len(t, timeline.Events, 1)
	assert.Equal(t, domain.ReplayEventLLMCall, timeline.Events[0].Type)
	assert.Equal(t, "Generate patch", timeline.Events[0].Data["title"])
	assert.Equal(t, "gpt-4.1", timeline.Events[0].Data["model"])
}

func TestReplaySessionSnapshotReconstructsPersistedFileState(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New()
	service, _ := newReplaySessionTestService(projectID, traceID.String(), &domain.ReplayTimeline{})
	session, err := service.CreateSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplaySessionInput{TraceID: traceID, Name: "Files"},
	)
	require.NoError(t, err)

	_, err = service.RecordEvents(
		context.Background(),
		projectID,
		session.ID,
		[]domain.AgentReplayRecordEventInput{
			{
				Type: domain.ReplayEventFileOperation,
				FileDelta: &domain.ReplayFileDelta{
					Path:      "src/main.go",
					Operation: "create",
					After:     "package main",
				},
			},
			{
				Type: domain.ReplayEventLLMCall,
				Data: map[string]interface{}{
					"model":        "gpt-4.1",
					"tokensInput":  12,
					"tokensOutput": 8,
					"cost":         0.01,
				},
			},
		},
	)
	require.NoError(t, err)

	fileState, err := service.GetFileStateAt(context.Background(), projectID, session.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "package main", fileState.Files["src/main.go"])

	snapshot, err := service.GetReplaySnapshot(context.Background(), projectID, session.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4.1", snapshot.ActiveModel)
	assert.Equal(t, 20, snapshot.TotalTokens)
	assert.InDelta(t, 0.01, snapshot.TotalCost, 0.0001)
}

func TestReplaySessionBranchValidatesEventIndex(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New()
	service, _ := newReplaySessionTestService(
		projectID,
		traceID.String(),
		&domain.ReplayTimeline{
			Events: []domain.ReplayEvent{{
				ID:        "event-1",
				Type:      domain.ReplayEventCheckpoint,
				Timestamp: time.Date(2026, 7, 25, 9, 0, 0, 0, time.UTC),
			}},
			Summary: domain.ReplaySummary{TotalEvents: 1},
		},
	)
	session, err := service.CreateSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplaySessionInput{TraceID: traceID, Name: "Parent"},
	)
	require.NoError(t, err)

	branch, err := service.BranchSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplayBranchRequest{
			SessionID:  session.ID,
			EventIndex: 0,
			Name:       "Checkpoint branch",
		},
	)

	require.NoError(t, err)
	require.NotNil(t, branch.ParentSessionID)
	assert.Equal(t, session.ID, *branch.ParentSessionID)

	_, err = service.BranchSession(
		context.Background(),
		projectID,
		uuid.New(),
		&domain.AgentReplayBranchRequest{
			SessionID:  session.ID,
			EventIndex: 4,
			Name:       "Invalid branch",
		},
	)
	require.Error(t, err)
}
