package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// replayPlanRepositoryStub mirrors the PostgreSQL repository semantics: every
// read and write is project scoped and status transitions are compare-and-swap.
type replayPlanRepositoryStub struct {
	mu                     sync.Mutex
	plans                  map[uuid.UUID]*domain.ReplayPlan
	lastUpdateContextErr   error
	lastUpdateContextValue any
	lastUpdateDeadline     time.Time
	lastUpdateHasDeadline  bool
}

type replayFinalizeContextKey struct{}

func (r *replayPlanRepositoryStub) Create(_ context.Context, plan *domain.ReplayPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *plan
	r.plans[plan.ID] = &stored
	return nil
}

func (r *replayPlanRepositoryStub) GetByID(
	_ context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.get(projectID, planID)
}

func (r *replayPlanRepositoryStub) get(
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	plan, ok := r.plans[planID]
	if !ok || plan.ProjectID != projectID {
		return nil, apperrors.NotFound("replay plan")
	}
	copied := *plan
	return &copied, nil
}

func (r *replayPlanRepositoryStub) Update(ctx context.Context, plan *domain.ReplayPlan) error {
	// Mirror a real database driver: a canceled context aborts the write. This
	// lets tests prove that terminal persistence runs on a detached context.
	ctxErr := ctx.Err()
	ctxValue := ctx.Value(replayFinalizeContextKey{})
	deadline, hasDeadline := ctx.Deadline()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastUpdateContextErr = ctxErr
	r.lastUpdateContextValue = ctxValue
	r.lastUpdateDeadline = deadline
	r.lastUpdateHasDeadline = hasDeadline
	if ctxErr != nil {
		return ctxErr
	}
	existing, ok := r.plans[plan.ID]
	if !ok || existing.ProjectID != plan.ProjectID {
		return apperrors.NotFound("replay plan")
	}
	stored := *plan
	r.plans[plan.ID] = &stored
	return nil
}

func (r *replayPlanRepositoryStub) TransitionStatus(
	_ context.Context,
	projectID, planID uuid.UUID,
	transition domain.ReplayPlanTransition,
) (*domain.ReplayPlan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	plan, ok := r.plans[planID]
	if !ok || plan.ProjectID != projectID {
		return nil, apperrors.NotFound("replay plan")
	}
	if plan.Status != transition.From {
		return nil, apperrors.Conflict(
			"replay plan must be " + string(transition.From) +
				", current status is " + string(plan.Status),
		)
	}
	plan.Status = transition.To
	plan.UpdatedAt = transition.At
	plan.FailureReason = ""
	copied := *plan
	return &copied, nil
}

type replayPlanTraceRepositoryStub struct {
	projectID uuid.UUID
	trace     *domain.Trace
}

func (r *replayPlanTraceRepositoryStub) GetByID(
	_ context.Context,
	projectID uuid.UUID,
	traceID string,
) (*domain.Trace, error) {
	if projectID != r.projectID || !sameTraceID(traceID, r.trace.ID) {
		return nil, apperrors.NotFound("trace")
	}
	return r.trace, nil
}

type replayPlanCheckpointRepositoryStub struct {
	checkpoint *domain.Checkpoint
}

func (r *replayPlanCheckpointRepositoryStub) GetByID(
	_ context.Context,
	projectID, checkpointID uuid.UUID,
) (*domain.Checkpoint, error) {
	if r.checkpoint == nil ||
		r.checkpoint.ProjectID != projectID ||
		r.checkpoint.ID != checkpointID {
		return nil, apperrors.NotFound("checkpoint")
	}
	return r.checkpoint, nil
}

type replayTimelineProviderStub struct {
	timeline *domain.ReplayTimeline
}

func (p *replayTimelineProviderStub) GetTimelineForTrace(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) (*domain.ReplayTimeline, error) {
	return p.timeline, nil
}

func TestReplayPlanLifecycle(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	eventTime := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	repository := &replayPlanRepositoryStub{plans: map[uuid.UUID]*domain.ReplayPlan{}}
	service := NewReplayPlanService(
		repository,
		&replayPlanTraceRepositoryStub{
			projectID: projectID,
			trace:     &domain.Trace{ID: traceID, ProjectID: projectID},
		},
		&replayPlanCheckpointRepositoryStub{},
		&replayTimelineProviderStub{
			timeline: &domain.ReplayTimeline{
				Events: []domain.ReplayEvent{
					{
						ID:        "generation-1",
						Type:      domain.ReplayEventLLMCall,
						Timestamp: eventTime,
						Data: domain.ReplayEventData{
							Model:        "gpt-4.1",
							Output:       "recorded output",
							TokensInput:  10,
							TokensOutput: 5,
							Cost:         0.02,
						},
					},
					{
						ID:        "terminal-1",
						Type:      domain.ReplayEventTerminalCmd,
						Timestamp: eventTime.Add(time.Second),
					},
				},
				Summary: domain.ReplaySummary{
					LLMCalls:         1,
					TerminalCommands: 1,
				},
			},
		},
		nil,
	)
	service.clock = func() time.Time { return eventTime }

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{},
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanReady, plan.Status)
	assert.False(t, plan.Capabilities.CanExecuteInSandbox)
	assert.True(t, plan.Capabilities.HasTerminalCommands)

	completed, err := service.ExecutePlan(context.Background(), projectID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanCompleted, completed.Status)
	require.NotNil(t, completed.Result)
	assert.Len(t, completed.Result.Generations, 1)
	assert.Equal(t, 1, completed.Result.NonGenerationEventsSkipped)
	assert.True(t, completed.Result.Comparison.Equivalent)
	assert.Equal(t, float64(0), completed.Result.Comparison.ReplayProviderCost)

	comparison, err := service.GetComparison(context.Background(), projectID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, "recorded_equivalent", comparison.Verdict)
}

func TestReplayPlanRejectsCrossProjectAccess(t *testing.T) {
	projectID := uuid.New()
	otherProjectID := uuid.New()
	traceID := uuid.New().String()
	repository := &replayPlanRepositoryStub{plans: map[uuid.UUID]*domain.ReplayPlan{}}
	service := NewReplayPlanService(
		repository,
		&replayPlanTraceRepositoryStub{
			projectID: projectID,
			trace:     &domain.Trace{ID: traceID, ProjectID: projectID},
		},
		&replayPlanCheckpointRepositoryStub{},
		&replayTimelineProviderStub{
			timeline: &domain.ReplayTimeline{
				Summary: domain.ReplaySummary{LLMCalls: 1},
			},
		},
		nil,
	)

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{},
	)
	require.NoError(t, err)

	_, err = service.GetPlan(context.Background(), otherProjectID, plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReplayPlanMarksUnsupportedPrerequisites(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	service := NewReplayPlanService(
		&replayPlanRepositoryStub{plans: map[uuid.UUID]*domain.ReplayPlan{}},
		&replayPlanTraceRepositoryStub{
			projectID: projectID,
			trace:     &domain.Trace{ID: traceID, ProjectID: projectID},
		},
		&replayPlanCheckpointRepositoryStub{},
		&replayTimelineProviderStub{timeline: &domain.ReplayTimeline{}},
		nil,
	)

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{Mode: domain.ReplayModeSandbox},
	)

	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanUnsupported, plan.Status)
	assert.Contains(t, plan.Capabilities.UnsupportedReasons, "sandbox execution is not configured")

	_, err = service.ExecutePlan(context.Background(), projectID, plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestReplayPlanCheckpointMustBelongToTrace(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	checkpointID := uuid.New()
	service := NewReplayPlanService(
		&replayPlanRepositoryStub{plans: map[uuid.UUID]*domain.ReplayPlan{}},
		&replayPlanTraceRepositoryStub{
			projectID: projectID,
			trace:     &domain.Trace{ID: traceID, ProjectID: projectID},
		},
		&replayPlanCheckpointRepositoryStub{
			checkpoint: &domain.Checkpoint{
				ID:        checkpointID,
				ProjectID: projectID,
				TraceID:   uuid.New().String(),
			},
		},
		&replayTimelineProviderStub{timeline: &domain.ReplayTimeline{}},
		nil,
	)

	_, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{CheckpointID: &checkpointID},
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// failingSandboxExecutor makes a replay fail after it has been claimed so the
// failure path can be observed.
type failingSandboxExecutor struct {
	calls int32
}

func (e *failingSandboxExecutor) Execute(
	_ context.Context,
	_ *domain.ReplayPlan,
) (*domain.ReplayPlanResult, error) {
	atomic.AddInt32(&e.calls, 1)
	return nil, apperrors.Unprocessable("sandbox rejected the replay")
}

// blockingSandboxExecutor keeps an execution in flight until it is released,
// which exposes concurrent execution attempts.
type blockingSandboxExecutor struct {
	started  chan struct{}
	release  chan struct{}
	attempts int32
}

func (e *blockingSandboxExecutor) Execute(
	_ context.Context,
	_ *domain.ReplayPlan,
) (*domain.ReplayPlanResult, error) {
	atomic.AddInt32(&e.attempts, 1)
	e.started <- struct{}{}
	<-e.release
	now := time.Now().UTC()
	return &domain.ReplayPlanResult{
		StartedAt:   now,
		CompletedAt: now,
		Generations: []domain.ReplayGenerationResult{{EventID: "generation-1"}},
	}, nil
}

func newSandboxReplayService(
	t *testing.T,
	projectID uuid.UUID,
	traceID string,
	executor ReplaySandboxExecutor,
) (*ReplayPlanService, *replayPlanRepositoryStub) {
	t.Helper()
	repository := &replayPlanRepositoryStub{plans: map[uuid.UUID]*domain.ReplayPlan{}}
	service := NewReplayPlanService(
		repository,
		&replayPlanTraceRepositoryStub{
			projectID: projectID,
			trace:     &domain.Trace{ID: traceID, ProjectID: projectID},
		},
		&replayPlanCheckpointRepositoryStub{},
		&replayTimelineProviderStub{
			timeline: &domain.ReplayTimeline{
				Summary: domain.ReplaySummary{LLMCalls: 1},
			},
		},
		executor,
	)
	return service, repository
}

func TestReplayPlanExecutesOnlyOnceUnderConcurrency(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	executor := &blockingSandboxExecutor{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	service, _ := newSandboxReplayService(t, projectID, traceID, executor)

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{Mode: domain.ReplayModeSandbox},
	)
	require.NoError(t, err)
	require.Equal(t, domain.ReplayPlanReady, plan.Status)

	firstDone := make(chan error, 1)
	go func() {
		_, execErr := service.ExecutePlan(context.Background(), projectID, plan.ID)
		firstDone <- execErr
	}()

	// The first execution is now inside the sandbox executor and the plan row is
	// already marked running, so a second request must be refused.
	<-executor.started

	_, secondErr := service.ExecutePlan(context.Background(), projectID, plan.ID)
	require.Error(t, secondErr)
	assert.True(t, apperrors.IsConflict(secondErr))
	assert.Contains(t, secondErr.Error(), string(domain.ReplayPlanRunning))

	close(executor.release)
	require.NoError(t, <-firstDone)
	assert.Equal(t, int32(1), atomic.LoadInt32(&executor.attempts))

	completed, err := service.GetPlan(context.Background(), projectID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanCompleted, completed.Status)
}

func TestReplayPlanPersistsFailureWithinProject(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	executor := &failingSandboxExecutor{}
	service, _ := newSandboxReplayService(t, projectID, traceID, executor)

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{Mode: domain.ReplayModeSandbox},
	)
	require.NoError(t, err)

	_, err = service.ExecutePlan(context.Background(), projectID, plan.ID)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&executor.calls))

	failed, err := service.GetPlan(context.Background(), projectID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanFailed, failed.Status)
	assert.Equal(t, "sandbox rejected the replay", failed.FailureReason)

	// The failed plan stays invisible and unusable outside its project.
	_, err = service.GetPlan(context.Background(), uuid.New(), plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	_, err = service.ExecutePlan(context.Background(), uuid.New(), plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReplayPlanNeverReclaimsRunningExecutionByAge(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	executor := &failingSandboxExecutor{}
	service, repository := newSandboxReplayService(t, projectID, traceID, executor)

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{Mode: domain.ReplayModeSandbox},
	)
	require.NoError(t, err)

	repository.mu.Lock()
	repository.plans[plan.ID].Status = domain.ReplayPlanRunning
	repository.plans[plan.ID].UpdatedAt = time.Now().UTC().Add(-24 * time.Hour)
	repository.mu.Unlock()

	_, err = service.ExecutePlan(context.Background(), projectID, plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, int32(0), atomic.LoadInt32(&executor.calls))
}

func TestReplayPlanRetriesFailedPlanWithinProject(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	service, _ := newSandboxReplayService(t, projectID, traceID, &failingSandboxExecutor{})

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{Mode: domain.ReplayModeSandbox},
	)
	require.NoError(t, err)

	_, err = service.ExecutePlan(context.Background(), projectID, plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsUnprocessable(err))

	_, err = service.RetryPlan(context.Background(), uuid.New(), plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	retried, err := service.RetryPlan(context.Background(), projectID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanReady, retried.Status)
	assert.Empty(t, retried.FailureReason)

	_, err = service.RetryPlan(context.Background(), projectID, plan.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

// cancelingSandboxExecutor cancels the request context mid-execution and then
// returns an error, reproducing a client disconnect during replay.
type cancelingSandboxExecutor struct {
	cancel                context.CancelFunc
	calls                 int32
	contextValue          any
	contextErrAfterCancel error
}

func (e *cancelingSandboxExecutor) Execute(
	ctx context.Context,
	_ *domain.ReplayPlan,
) (*domain.ReplayPlanResult, error) {
	atomic.AddInt32(&e.calls, 1)
	e.contextValue = ctx.Value(replayFinalizeContextKey{})
	// Simulate the caller canceling the request while the executor is running.
	e.cancel()
	e.contextErrAfterCancel = ctx.Err()
	return nil, apperrors.Unprocessable("sandbox aborted mid-run")
}

// TestReplayPlanPersistsFailureAfterRequestCancellation proves that a replay
// whose request context is canceled during execution still persists its
// terminal "failed" state (on a detached context), so the plan does not get
// stuck in "running" and can subsequently be retried.
func TestReplayPlanPersistsFailureAfterRequestCancellation(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()

	ctx, cancel := context.WithCancel(context.WithValue(
		context.Background(),
		replayFinalizeContextKey{},
		"request-value",
	))
	defer cancel()
	executor := &cancelingSandboxExecutor{cancel: cancel}
	service, repository := newSandboxReplayService(t, projectID, traceID, executor)

	plan, err := service.CreatePlan(
		context.Background(),
		projectID,
		traceID,
		nil,
		domain.ReplayPlanInput{Mode: domain.ReplayModeSandbox},
	)
	require.NoError(t, err)

	// Execute with the cancellable context; the executor cancels it and errors.
	_, err = service.ExecutePlan(ctx, projectID, plan.ID)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&executor.calls))
	assert.Equal(t, "request-value", executor.contextValue)
	assert.ErrorIs(t, executor.contextErrAfterCancel, context.Canceled)
	require.Error(t, ctx.Err(), "request context should have been canceled")

	// Despite the cancellation, the terminal failure was persisted, so the plan
	// is "failed" (not left stranded in "running").
	repository.mu.Lock()
	stored := repository.plans[plan.ID]
	status := stored.Status
	updateContextErr := repository.lastUpdateContextErr
	updateContextValue := repository.lastUpdateContextValue
	updateDeadline := repository.lastUpdateDeadline
	updateHasDeadline := repository.lastUpdateHasDeadline
	repository.mu.Unlock()
	assert.Equal(t, domain.ReplayPlanFailed, status)
	assert.NoError(t, updateContextErr, "terminal persistence must ignore request cancellation")
	assert.Equal(t, "request-value", updateContextValue, "detached context must retain request values")
	require.True(t, updateHasDeadline, "terminal persistence must remain time bounded")
	assert.Greater(t, time.Until(updateDeadline), time.Duration(0))
	assert.LessOrEqual(t, time.Until(updateDeadline), replayFinalizeTimeout)

	// And a failed plan can be retried back to ready.
	retried, err := service.RetryPlan(context.Background(), projectID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ReplayPlanReady, retried.Status)
}
