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

func TestNewOrchestrationDebuggerService(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestOrchestrationDebuggerService_CreateSession(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New()

	session, err := svc.CreateSession(ctx, projectID, &domain.OrchestrationSessionInput{TraceID: traceID})
	require.NoError(t, err)
	assert.Equal(t, projectID, session.ProjectID)
	assert.Equal(t, traceID, session.TraceID)
	assert.Equal(t, "running", session.Status)
	assert.Equal(t, 0, session.CurrentStep)
	assert.Greater(t, len(session.Agents), 0)
	assert.Greater(t, len(session.Messages), 0)
	assert.Equal(t, len(session.Messages), session.TotalSteps)
}

func TestOrchestrationDebuggerService_GetSession(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	ctx := context.Background()

	t.Run("existing session", func(t *testing.T) {
		session, err := svc.CreateSession(ctx, uuid.New(), &domain.OrchestrationSessionInput{TraceID: uuid.New()})
		require.NoError(t, err)

		got, err := svc.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetSession(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestOrchestrationDebuggerService_ExecuteCommand(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name       string
		action     string
		stepCount  int
		wantStatus string
		checkStep  func(t *testing.T, before, after int)
	}{
		{
			name:       "step advances by 1",
			action:     "step",
			stepCount:  1,
			wantStatus: "running",
			checkStep:  func(t *testing.T, before, after int) { assert.Equal(t, before+1, after) },
		},
		{
			name:       "inspect does not change step",
			action:     "inspect",
			wantStatus: "running",
			checkStep:  func(t *testing.T, before, after int) { assert.Equal(t, before, after) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			session, _ := svc.CreateSession(ctx, uuid.New(), &domain.OrchestrationSessionInput{TraceID: uuid.New()})
			beforeStep := session.CurrentStep

			updated, err := svc.ExecuteCommand(ctx, session.ID, &domain.DebugCommand{
				Action:    tc.action,
				StepCount: tc.stepCount,
			})
			require.NoError(t, err)
			tc.checkStep(t, beforeStep, updated.CurrentStep)
		})
	}

	t.Run("continue runs to completion without breakpoints", func(t *testing.T) {
		session, _ := svc.CreateSession(ctx, uuid.New(), &domain.OrchestrationSessionInput{TraceID: uuid.New()})
		updated, err := svc.ExecuteCommand(ctx, session.ID, &domain.DebugCommand{Action: "continue"})
		require.NoError(t, err)
		assert.Equal(t, "completed", updated.Status)
		assert.Equal(t, updated.TotalSteps, updated.CurrentStep)
	})

	t.Run("unknown action returns error", func(t *testing.T) {
		session, _ := svc.CreateSession(ctx, uuid.New(), &domain.OrchestrationSessionInput{TraceID: uuid.New()})
		_, err := svc.ExecuteCommand(ctx, session.ID, &domain.DebugCommand{Action: "invalid"})
		assert.Error(t, err)
	})

	t.Run("not found session", func(t *testing.T) {
		_, err := svc.ExecuteCommand(ctx, uuid.New(), &domain.DebugCommand{Action: "step"})
		assert.Error(t, err)
	})
}

func TestOrchestrationDebuggerService_AddBreakpoint(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, uuid.New(), &domain.OrchestrationSessionInput{TraceID: uuid.New()})

	updated, err := svc.AddBreakpoint(ctx, session.ID, &domain.AgentBreakpoint{
		AgentID:   "worker-code",
		Condition: "on_message",
		Enabled:   true,
	})
	require.NoError(t, err)
	assert.Len(t, updated.Breakpoints, 1)
	assert.Equal(t, "worker-code", updated.Breakpoints[0].AgentID)
	assert.NotEmpty(t, updated.Breakpoints[0].ID)
}

func TestOrchestrationDebuggerService_ContinueWithBreakpoint(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, uuid.New(), &domain.OrchestrationSessionInput{TraceID: uuid.New()})

	// Add breakpoint on an agent that appears in messages
	_, err := svc.AddBreakpoint(ctx, session.ID, &domain.AgentBreakpoint{
		AgentID:   "worker-code",
		Condition: "on_message",
		Enabled:   true,
	})
	require.NoError(t, err)

	updated, err := svc.ExecuteCommand(ctx, session.ID, &domain.DebugCommand{Action: "continue"})
	require.NoError(t, err)
	assert.Equal(t, "paused", updated.Status)
	assert.Greater(t, updated.CurrentStep, 0)
}

func TestOrchestrationDebuggerService_ListSessions(t *testing.T) {
	svc := NewOrchestrationDebuggerService(zap.NewNop())
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	_, _ = svc.CreateSession(ctx, projectA, &domain.OrchestrationSessionInput{TraceID: uuid.New()})
	_, _ = svc.CreateSession(ctx, projectA, &domain.OrchestrationSessionInput{TraceID: uuid.New()})
	_, _ = svc.CreateSession(ctx, projectB, &domain.OrchestrationSessionInput{TraceID: uuid.New()})

	t.Run("filters by project", func(t *testing.T) {
		sessionsA, err := svc.ListSessions(ctx, projectA)
		require.NoError(t, err)
		assert.Len(t, sessionsA, 2)

		sessionsB, err := svc.ListSessions(ctx, projectB)
		require.NoError(t, err)
		assert.Len(t, sessionsB, 1)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		sessions, err := svc.ListSessions(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, sessions)
	})
}
