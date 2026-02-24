package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestReplaySessionCreateSession(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	userID := uuid.New()
	traceID := uuid.New()

	input := &domain.AgentReplaySessionInput{
		TraceID:     traceID,
		Name:        "Test Session",
		Description: "A test replay session",
	}

	session, err := svc.CreateSession(ctx, projectID, userID, input)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.Equal(t, projectID, session.ProjectID)
	assert.Equal(t, traceID, session.TraceID)
	assert.Equal(t, "Test Session", session.Name)
	assert.Equal(t, "A test replay session", session.Description)
	assert.Equal(t, domain.AgentReplayRecording, session.Status)
	assert.Equal(t, domain.AgentReplayFidelityStandard, session.RecordingFidelity)
	assert.False(t, session.IsPublic)
	assert.Equal(t, userID, session.CreatedBy)
	assert.False(t, session.CreatedAt.IsZero())
}

func TestReplaySessionCreateSessionCustomFidelity(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	input := &domain.AgentReplaySessionInput{
		TraceID:           uuid.New(),
		Name:              "Full Fidelity Session",
		RecordingFidelity: domain.AgentReplayFidelityFull,
	}

	session, err := svc.CreateSession(ctx, uuid.New(), uuid.New(), input)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, domain.AgentReplayFidelityFull, session.RecordingFidelity)
}

func TestReplaySessionRecordEvent(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	sessionID := uuid.New()
	data := map[string]interface{}{"key": "value"}

	event, err := svc.RecordEvent(ctx, sessionID, domain.ReplayEventLLMCall, data, 150)
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, domain.ReplayEventLLMCall, event.Type)
	assert.Equal(t, int64(150), event.DurationMs)
	assert.Equal(t, "value", event.Data["key"])
	assert.False(t, event.Timestamp.IsZero())
}

func TestReplaySessionCompleteSession(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	err := svc.CompleteSession(ctx, uuid.New())
	require.NoError(t, err)
}

func TestReplaySessionGetTimeline(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	sessionID := uuid.New()
	timeline, err := svc.GetTimeline(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, timeline)

	assert.Equal(t, sessionID, timeline.Session.ID)
	assert.Equal(t, domain.AgentReplayCompleted, timeline.Session.Status)
	assert.Empty(t, timeline.Events)
	assert.Empty(t, timeline.Branches)
	assert.Empty(t, timeline.Milestones)
}

func TestReplaySessionBranchSession(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()

	req := &domain.AgentReplayBranchRequest{
		SessionID:  parentID,
		EventIndex: 5,
		Name:       "Branch at step 5",
	}

	branch, err := svc.BranchSession(ctx, projectID, userID, req)
	require.NoError(t, err)
	require.NotNil(t, branch)

	assert.NotEqual(t, uuid.Nil, branch.ID)
	assert.Equal(t, projectID, branch.ProjectID)
	assert.Equal(t, "Branch at step 5", branch.Name)
	assert.Equal(t, domain.AgentReplayRecording, branch.Status)
	require.NotNil(t, branch.ParentSessionID)
	assert.Equal(t, parentID, *branch.ParentSessionID)
	assert.Equal(t, 5, branch.BranchPoint)
	assert.Equal(t, userID, branch.CreatedBy)
}

func TestReplaySessionGetPlaybackState(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	sessionID := uuid.New()
	state, err := svc.GetPlaybackState(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, state)

	assert.Equal(t, sessionID, state.SessionID)
	assert.Equal(t, 0, state.CurrentIndex)
	assert.Equal(t, 0, state.TotalEvents)
	assert.False(t, state.IsPlaying)
	assert.Equal(t, 1.0, state.Speed)
}

func TestReplaySessionShareSession(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	sessionID := uuid.New()
	shareURL, err := svc.ShareSession(ctx, sessionID)
	require.NoError(t, err)

	assert.Contains(t, shareURL, sessionID.String())
	assert.Contains(t, shareURL, "/replay/")
	assert.Contains(t, shareURL, "/shared")
}

func TestReplaySessionDetectMilestones(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)

	events := []domain.AgentReplayTimelineEvent{
		{Index: 0, Type: domain.ReplayEventLLMCall},
		{Index: 1, Type: domain.ReplayEventCheckpoint},
		{Index: 2, Type: domain.ReplayEventToolCall},
		{Index: 3, Type: domain.ReplayEventError},
		{Index: 4, Type: domain.ReplayEventCheckpoint},
	}

	milestones := svc.DetectMilestones(events)
	require.Len(t, milestones, 3)

	assert.Equal(t, 1, milestones[0].EventIndex)
	assert.Equal(t, "Checkpoint", milestones[0].Label)
	assert.Equal(t, "checkpoint", milestones[0].Type)

	assert.Equal(t, 3, milestones[1].EventIndex)
	assert.Equal(t, "Error", milestones[1].Label)
	assert.Equal(t, "error", milestones[1].Type)

	assert.Equal(t, 4, milestones[2].EventIndex)
	assert.Equal(t, "checkpoint", milestones[2].Type)
}

func TestReplaySessionDetectMilestonesEmpty(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)

	milestones := svc.DetectMilestones([]domain.AgentReplayTimelineEvent{})
	assert.Empty(t, milestones)
}

func TestReplaySessionListSessions(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	filter := domain.AgentReplaySessionFilter{
		ProjectID: uuid.New(),
	}

	list, err := svc.ListSessions(ctx, filter)
	require.NoError(t, err)
	require.NotNil(t, list)

	assert.Empty(t, list.Sessions)
	assert.Equal(t, int64(0), list.TotalCount)
	assert.False(t, list.HasMore)
}

func TestReplaySessionGetSessionNotFound(t *testing.T) {
	logger := zap.NewNop()
	svc := NewReplaySessionService(logger)
	ctx := context.Background()

	session, err := svc.GetSession(ctx, uuid.New())
	assert.Nil(t, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replay session not found")
}
