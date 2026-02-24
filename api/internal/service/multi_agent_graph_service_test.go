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

func TestMultiAgentGraphAnalyzeSession(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	traceID := uuid.New()

	session, err := svc.AnalyzeSession(ctx, projectID, traceID)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.Equal(t, projectID, session.ProjectID)
	assert.Equal(t, traceID, session.TraceID)
	assert.Contains(t, session.Name, traceID.String())
	assert.Equal(t, "completed", session.Status)
	assert.NotNil(t, session.Agents)
	assert.NotNil(t, session.Messages)
	assert.NotNil(t, session.Bottlenecks)
	assert.False(t, session.CreatedAt.IsZero())
}

func TestMultiAgentGraphDetectTopologyHubSpoke(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)

	agents := []domain.CollabAgentNode{
		{ID: uuid.New(), Name: "Coordinator", Role: domain.AgentRoleCoordinator},
		{ID: uuid.New(), Name: "Worker1", Role: domain.AgentRoleWorker},
		{ID: uuid.New(), Name: "Worker2", Role: domain.AgentRoleWorker},
	}

	messages := []domain.CollabAgentMessage{
		{FromAgentID: agents[0].ID, ToAgentID: agents[1].ID},
		{FromAgentID: agents[0].ID, ToAgentID: agents[2].ID},
	}

	topology := svc.DetectTopology(agents, messages)
	assert.Equal(t, domain.CollabTopologyHubSpoke, topology)
}

func TestMultiAgentGraphDetectTopologyMesh(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)

	a1 := uuid.New()
	a2 := uuid.New()
	a3 := uuid.New()

	agents := []domain.CollabAgentNode{
		{ID: a1, Name: "Agent1", Role: domain.AgentRoleWorker},
		{ID: a2, Name: "Agent2", Role: domain.AgentRoleWorker},
		{ID: a3, Name: "Agent3", Role: domain.AgentRoleWorker},
	}

	// All agents communicate with all others (n*(n-1) = 6 pairs)
	messages := []domain.CollabAgentMessage{
		{FromAgentID: a1, ToAgentID: a2},
		{FromAgentID: a1, ToAgentID: a3},
		{FromAgentID: a2, ToAgentID: a1},
		{FromAgentID: a2, ToAgentID: a3},
		{FromAgentID: a3, ToAgentID: a1},
		{FromAgentID: a3, ToAgentID: a2},
	}

	topology := svc.DetectTopology(agents, messages)
	assert.Equal(t, domain.CollabTopologyMesh, topology)
}

func TestMultiAgentGraphDetectTopologyPipeline(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)

	a1 := uuid.New()
	a2 := uuid.New()
	a3 := uuid.New()

	agents := []domain.CollabAgentNode{
		{ID: a1, Name: "Agent1", Role: domain.AgentRoleWorker},
		{ID: a2, Name: "Agent2", Role: domain.AgentRoleWorker},
		{ID: a3, Name: "Agent3", Role: domain.AgentRoleWorker},
	}

	// Sequential: A->B->C (each sends to at most one other)
	messages := []domain.CollabAgentMessage{
		{FromAgentID: a1, ToAgentID: a2},
		{FromAgentID: a2, ToAgentID: a3},
	}

	topology := svc.DetectTopology(agents, messages)
	assert.Equal(t, domain.CollabTopologyPipeline, topology)
}

func TestMultiAgentGraphIdentifyBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)

	session := &domain.MultiAgentSession{
		Agents: []domain.CollabAgentNode{
			{ID: uuid.New(), Name: "FastAgent", AvgLatencyMs: 100, TaskCount: 5},
			{ID: uuid.New(), Name: "SlowAgent", AvgLatencyMs: 8000, TaskCount: 3},
			{ID: uuid.New(), Name: "OverloadedAgent", AvgLatencyMs: 200, TaskCount: 25},
		},
		Messages: []domain.CollabAgentMessage{},
	}

	bottlenecks := svc.IdentifyBottlenecks(session)
	require.Len(t, bottlenecks, 2)

	// SlowAgent should be flagged for high latency
	assert.Equal(t, "high_latency", bottlenecks[0].Type)
	assert.Contains(t, bottlenecks[0].Description, "SlowAgent")

	// OverloadedAgent should be flagged for too many tasks
	assert.Equal(t, "overloaded", bottlenecks[1].Type)
	assert.Contains(t, bottlenecks[1].Description, "OverloadedAgent")
}

func TestMultiAgentGraphListSessions(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)
	ctx := context.Background()

	list, err := svc.ListSessions(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, list)

	assert.Empty(t, list.Sessions)
	assert.Equal(t, int64(0), list.TotalCount)
	assert.False(t, list.HasMore)
}

func TestMultiAgentGraphGetSession(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMultiAgentGraphService(logger)
	ctx := context.Background()

	sessionID := uuid.New()
	session, err := svc.GetSession(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, domain.CollabTopologyPipeline, session.Topology)
	assert.Equal(t, "completed", session.Status)
	assert.NotNil(t, session.Agents)
	assert.NotNil(t, session.Messages)
	assert.NotNil(t, session.Bottlenecks)
}
