package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ReplayPlanRepository defines safe replay plan persistence.
type ReplayPlanRepository interface {
	Create(ctx context.Context, plan *domain.ReplayPlan) error
	GetByID(ctx context.Context, projectID, planID uuid.UUID) (*domain.ReplayPlan, error)
	Update(ctx context.Context, plan *domain.ReplayPlan) error
	// TransitionStatus atomically applies a conditional status change and
	// returns the updated plan or a conflict error.
	TransitionStatus(
		ctx context.Context,
		projectID, planID uuid.UUID,
		transition domain.ReplayPlanTransition,
	) (*domain.ReplayPlan, error)
}

// ReplayPlanTraceRepository verifies trace ownership.
type ReplayPlanTraceRepository interface {
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
}

// ReplayPlanCheckpointRepository verifies checkpoint ownership.
type ReplayPlanCheckpointRepository interface {
	GetByID(
		ctx context.Context,
		projectID, checkpointID uuid.UUID,
	) (*domain.Checkpoint, error)
}

// ReplayTimelineProvider provides the existing coherent trace timeline.
type ReplayTimelineProvider interface {
	GetTimelineForTrace(
		ctx context.Context,
		projectID uuid.UUID,
		traceID string,
	) (*domain.ReplayTimeline, error)
}

// ReplaySandboxExecutor is intentionally optional; nil means host execution is forbidden.
type ReplaySandboxExecutor interface {
	Execute(ctx context.Context, plan *domain.ReplayPlan) (*domain.ReplayPlanResult, error)
}

// ReplayPlanService owns safe replay planning and state transitions.
type ReplayPlanService struct {
	repository       ReplayPlanRepository
	traceRepository  ReplayPlanTraceRepository
	checkpointRepo   ReplayPlanCheckpointRepository
	timelineProvider ReplayTimelineProvider
	sandboxExecutor  ReplaySandboxExecutor
	clock            func() time.Time
}

// NewReplayPlanService creates a safe replay planning service.
func NewReplayPlanService(
	repository ReplayPlanRepository,
	traceRepository ReplayPlanTraceRepository,
	checkpointRepository ReplayPlanCheckpointRepository,
	timelineProvider ReplayTimelineProvider,
	sandboxExecutor ReplaySandboxExecutor,
) *ReplayPlanService {
	return &ReplayPlanService{
		repository:       repository,
		traceRepository:  traceRepository,
		checkpointRepo:   checkpointRepository,
		timelineProvider: timelineProvider,
		sandboxExecutor:  sandboxExecutor,
		clock:            time.Now,
	}
}

// AssessCapabilities evaluates replay prerequisites without persisting a plan.
func (s *ReplayPlanService) AssessCapabilities(
	ctx context.Context,
	projectID uuid.UUID,
	traceID string,
	input domain.ReplayPlanInput,
) (*domain.ReplayCapabilityReport, error) {
	if _, err := s.traceRepository.GetByID(ctx, projectID, traceID); err != nil {
		return nil, err
	}

	timeline, err := s.timelineProvider.GetTimelineForTrace(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("build replay timeline: %w", err)
	}

	checkpoint, err := s.resolveCheckpoint(ctx, projectID, traceID, input.CheckpointID)
	if err != nil {
		return nil, err
	}

	report := buildReplayCapabilities(timeline, checkpoint != nil, s.sandboxExecutor != nil)
	applyRequestCapabilityChecks(&report, input)
	return &report, nil
}

// CreatePlan validates and persists a project-scoped replay plan.
func (s *ReplayPlanService) CreatePlan(
	ctx context.Context,
	projectID uuid.UUID,
	traceID string,
	createdBy *uuid.UUID,
	input domain.ReplayPlanInput,
) (*domain.ReplayPlan, error) {
	normalizeReplayPlanInput(&input)
	if err := validateReplayPlanInput(input); err != nil {
		return nil, err
	}

	capabilities, err := s.AssessCapabilities(ctx, projectID, traceID, input)
	if err != nil {
		return nil, err
	}

	status := domain.ReplayPlanReady
	if len(capabilities.UnsupportedReasons) > 0 {
		status = domain.ReplayPlanUnsupported
	}

	now := s.clock().UTC()
	plan := &domain.ReplayPlan{
		ID:           uuid.New(),
		ProjectID:    projectID,
		TraceID:      traceID,
		CheckpointID: input.CheckpointID,
		Status:       status,
		Request:      input,
		Capabilities: *capabilities,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repository.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("persist replay plan: %w", err)
	}
	return plan, nil
}

// GetPlan retrieves a plan only within the authorized project.
func (s *ReplayPlanService) GetPlan(
	ctx context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	return s.repository.GetByID(ctx, projectID, planID)
}

// ExecutePlan executes only recorded generation replay or a configured sandbox executor.
// The ready-to-running transition is atomic, so concurrent requests for the same
// plan cannot both execute it.
func (s *ReplayPlanService) ExecutePlan(
	ctx context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	plan, err := s.claimPlanForExecution(ctx, projectID, planID)
	if err != nil {
		return nil, err
	}

	var result *domain.ReplayPlanResult
	switch plan.Request.Mode {
	case domain.ReplayModeRecordedGeneration:
		result, err = s.executeRecordedGeneration(ctx, plan)
	case domain.ReplayModeSandbox:
		if s.sandboxExecutor == nil {
			err = apperrors.Unprocessable("sandbox replay is not configured")
		} else {
			result, err = s.sandboxExecutor.Execute(ctx, plan)
		}
	default:
		err = apperrors.Validation("unsupported replay mode")
	}

	if err != nil {
		plan.ProjectID = projectID
		plan.Status = domain.ReplayPlanFailed
		plan.FailureReason = safeReplayFailure(err)
		plan.UpdatedAt = s.clock().UTC()
		// Persist the terminal failure on a detached context so a canceled
		// request cannot leave the plan stranded in "running" (which would also
		// block RetryPlan, since it only transitions from "failed").
		persistCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			replayFinalizeTimeout,
		)
		defer cancel()
		if updateErr := s.repository.Update(persistCtx, plan); updateErr != nil {
			return nil, fmt.Errorf(
				"replay failed: %v; persist failure state: %w",
				err,
				updateErr,
			)
		}
		return nil, err
	}

	plan.ProjectID = projectID
	plan.Result = result
	plan.Status = domain.ReplayPlanCompleted
	plan.FailureReason = ""
	plan.UpdatedAt = s.clock().UTC()
	// Terminal success persistence is likewise detached so an already-canceled
	// request context cannot drop the completed result.
	persistCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		replayFinalizeTimeout,
	)
	defer cancel()
	if err := s.repository.Update(persistCtx, plan); err != nil {
		return nil, fmt.Errorf("complete replay plan: %w", err)
	}
	return plan, nil
}

// replayFinalizeTimeout bounds the terminal persistence write that records a
// replay plan's completed or failed state. The write runs on a context detached
// from the request so a client cancellation during execution cannot leave the
// plan stuck in "running" and un-retryable.
const replayFinalizeTimeout = 5 * time.Second

// RetryPlan atomically returns a failed plan to ready. Running plans are never
// reclaimed automatically because elapsed time alone cannot prove that an
// external sandbox execution has stopped.
func (s *ReplayPlanService) RetryPlan(
	ctx context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	return s.repository.TransitionStatus(
		ctx,
		projectID,
		planID,
		domain.ReplayPlanTransition{
			From: domain.ReplayPlanFailed,
			To:   domain.ReplayPlanReady,
			At:   s.clock().UTC(),
		},
	)
}

// claimPlanForExecution wins exclusive ownership of a ready plan execution.
func (s *ReplayPlanService) claimPlanForExecution(
	ctx context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	return s.repository.TransitionStatus(
		ctx,
		projectID,
		planID,
		domain.ReplayPlanTransition{
			From: domain.ReplayPlanReady,
			To:   domain.ReplayPlanRunning,
			At:   s.clock().UTC(),
		},
	)
}

// GetComparison returns the original-versus-replay branch comparison.
func (s *ReplayPlanService) GetComparison(
	ctx context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlanComparison, error) {
	plan, err := s.repository.GetByID(ctx, projectID, planID)
	if err != nil {
		return nil, err
	}
	if plan.Status != domain.ReplayPlanCompleted || plan.Result == nil {
		return nil, apperrors.Conflict("replay comparison is available after completion")
	}
	return &plan.Result.Comparison, nil
}

func (s *ReplayPlanService) resolveCheckpoint(
	ctx context.Context,
	projectID uuid.UUID,
	traceID string,
	checkpointID *uuid.UUID,
) (*domain.Checkpoint, error) {
	if checkpointID == nil {
		return nil, nil
	}
	checkpoint, err := s.checkpointRepo.GetByID(ctx, projectID, *checkpointID)
	if err != nil {
		return nil, err
	}
	if !sameTraceID(checkpoint.TraceID, traceID) {
		return nil, apperrors.NotFound("checkpoint")
	}
	return checkpoint, nil
}

func (s *ReplayPlanService) executeRecordedGeneration(
	ctx context.Context,
	plan *domain.ReplayPlan,
) (*domain.ReplayPlanResult, error) {
	timeline, err := s.timelineProvider.GetTimelineForTrace(ctx, plan.ProjectID, plan.TraceID)
	if err != nil {
		return nil, fmt.Errorf("build replay timeline: %w", err)
	}
	checkpoint, err := s.resolveCheckpoint(
		ctx,
		plan.ProjectID,
		plan.TraceID,
		plan.CheckpointID,
	)
	if err != nil {
		return nil, err
	}

	startedAt := s.clock().UTC()
	var checkpointTime *time.Time
	if checkpoint != nil {
		checkpointTime = &checkpoint.CreatedAt
	}

	result := &domain.ReplayPlanResult{
		StartedAt:   startedAt,
		Generations: []domain.ReplayGenerationResult{},
	}
	var referenceTokens int
	var referenceCost float64

	for _, event := range timeline.Events {
		if checkpointTime != nil && event.Timestamp.Before(*checkpointTime) {
			continue
		}
		if event.Type != domain.ReplayEventLLMCall {
			result.NonGenerationEventsSkipped++
			continue
		}

		outputJSON, err := json.Marshal(event.Data.Output)
		if err != nil {
			return nil, fmt.Errorf("hash recorded generation output: %w", err)
		}
		outputHash := sha256.Sum256(outputJSON)
		tokens := event.Data.TokensInput + event.Data.TokensOutput
		referenceTokens += tokens
		referenceCost += event.Data.Cost
		result.Generations = append(result.Generations, domain.ReplayGenerationResult{
			EventID:       event.ID,
			Model:         event.Data.Model,
			OutputSHA256:  hex.EncodeToString(outputHash[:]),
			Tokens:        tokens,
			ReferenceCost: event.Data.Cost,
		})
	}

	if len(result.Generations) == 0 {
		return nil, apperrors.Unprocessable("no recorded generations are available after the checkpoint")
	}

	result.CompletedAt = s.clock().UTC()
	result.Comparison = domain.ReplayPlanComparison{
		OriginalGenerationCount: len(result.Generations),
		ReplayGenerationCount:   len(result.Generations),
		OriginalTokens:          referenceTokens,
		ReplayTokens:            referenceTokens,
		OriginalCost:            referenceCost,
		ReplayProviderCost:      0,
		Equivalent:              true,
		Verdict:                 "recorded_equivalent",
		Notes: []string{
			"Recorded generation outputs were replayed deterministically by content hash.",
			"No terminal commands, tools, file writes, or arbitrary code were executed.",
			"Provider cost is zero because no external model call was made.",
		},
	}
	return result, nil
}

func buildReplayCapabilities(
	timeline *domain.ReplayTimeline,
	hasCheckpoint, hasSandbox bool,
) domain.ReplayCapabilityReport {
	report := domain.ReplayCapabilityReport{
		CanInspectTimeline:          true,
		CanReplayRecordedGeneration: timeline.Summary.LLMCalls > 0,
		CanExecuteInSandbox:         hasSandbox,
		HasCheckpoint:               hasCheckpoint,
		HasFileOperations:           timeline.Summary.FileOperations > 0,
		HasTerminalCommands:         timeline.Summary.TerminalCommands > 0,
		GenerationCount:             timeline.Summary.LLMCalls,
		UnsupportedReasons:          []string{},
		SafetyNotice:                "AgentTrace never executes recorded terminal commands, tools, or file writes on the API host.",
	}
	return report
}

func applyRequestCapabilityChecks(
	report *domain.ReplayCapabilityReport,
	input domain.ReplayPlanInput,
) {
	switch input.Mode {
	case domain.ReplayModeRecordedGeneration:
		if !report.CanReplayRecordedGeneration {
			report.UnsupportedReasons = append(
				report.UnsupportedReasons,
				"the selected trace has no recorded model generations",
			)
		}
		if input.ModelOverride != "" || input.PromptOverride != "" || input.Temperature != nil {
			report.UnsupportedReasons = append(
				report.UnsupportedReasons,
				"model or prompt overrides require a configured sandbox/model provider",
			)
		}
	case domain.ReplayModeSandbox:
		if !report.CanExecuteInSandbox {
			report.UnsupportedReasons = append(
				report.UnsupportedReasons,
				"sandbox execution is not configured",
			)
		}
	default:
		report.UnsupportedReasons = append(report.UnsupportedReasons, "unsupported replay mode")
	}
}

func normalizeReplayPlanInput(input *domain.ReplayPlanInput) {
	if input.Mode == "" {
		input.Mode = domain.ReplayModeRecordedGeneration
	}
}

func validateReplayPlanInput(input domain.ReplayPlanInput) error {
	switch input.Mode {
	case domain.ReplayModeRecordedGeneration, domain.ReplayModeSandbox:
	default:
		return apperrors.Validation("mode must be recorded_generation or sandbox")
	}
	if input.Temperature != nil && (*input.Temperature < 0 || *input.Temperature > 2) {
		return apperrors.Validation("temperature must be between 0 and 2")
	}
	return nil
}

func sameTraceID(left, right string) bool {
	normalize := func(value string) string {
		return strings.ToLower(strings.ReplaceAll(value, "-", ""))
	}
	return normalize(left) == normalize(right)
}

func safeReplayFailure(err error) string {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		return appErr.Message
	}
	return "replay failed"
}
