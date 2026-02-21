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

// DebugSessionRepository defines repository operations needed for debug sessions
type DebugSessionRepository interface {
	Save(ctx context.Context, session *domain.DebugSession) error
	Get(ctx context.Context, id uuid.UUID) (*domain.DebugSession, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.DebugSession, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// DebugService extends replay with interactive debugging capabilities.
// It manages debug sessions, step-by-step state reconstruction, and annotations.
type DebugService struct {
	logger        *zap.Logger
	sessionRepo   DebugSessionRepository
	replayService *ReplayService
}

// NewDebugService creates a new debug service
func NewDebugService(
	logger *zap.Logger,
	sessionRepo DebugSessionRepository,
	replayService *ReplayService,
) *DebugService {
	return &DebugService{
		logger:        logger,
		sessionRepo:   sessionRepo,
		replayService: replayService,
	}
}

// CreateSession creates a new interactive debug session for a trace
func (s *DebugService) CreateSession(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, input *domain.CreateDebugSessionInput) (*domain.DebugSession, error) {
	// Build the timeline to determine total steps
	timeline, err := s.replayService.GetTimelineForTrace(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build timeline for debug session: %w", err)
	}

	now := time.Now()
	session := &domain.DebugSession{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     input.TraceID,
		UserID:      userID,
		Status:      domain.DebugSessionActive,
		CurrentStep: 0,
		TotalSteps:  len(timeline.Events),
		Breakpoints: input.Breakpoints,
		Annotations: []domain.DebugAnnotation{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save debug session: %w", err)
	}

	s.logger.Info("created debug session",
		zap.String("sessionId", session.ID.String()),
		zap.String("traceId", input.TraceID),
		zap.Int("totalSteps", session.TotalSteps),
	)

	return session, nil
}

// GetSession retrieves a debug session by ID
func (s *DebugService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.DebugSession, error) {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debug session: %w", err)
	}
	return session, nil
}

// GetStepState reconstructs the state at a specific step index by replaying
// file operations and accumulating cost/token data up to that step.
func (s *DebugService) GetStepState(ctx context.Context, projectID uuid.UUID, traceID string, stepIndex int) (*domain.DebugStepState, error) {
	timeline, err := s.replayService.GetTimelineForTrace(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build timeline: %w", err)
	}

	if stepIndex < 0 || stepIndex >= len(timeline.Events) {
		return nil, fmt.Errorf("step index %d out of range [0, %d)", stepIndex, len(timeline.Events))
	}

	// Reconstruct file tree state by replaying file operations up to stepIndex
	fileTree := make(map[string]*domain.FileTreeEntry)
	var modifiedFiles []domain.FileDiffEntry
	var costSoFar float64
	var tokensSoFar int
	var elapsedMs int64

	for i := 0; i <= stepIndex; i++ {
		event := timeline.Events[i]

		// Accumulate cost and tokens from LLM calls
		if event.Type == domain.ReplayEventLLMCall {
			costSoFar += event.Data.Cost
			tokensSoFar += event.Data.TokensInput + event.Data.TokensOutput
		}

		// Track file operations to build file tree
		if event.Type == domain.ReplayEventFileOperation && event.Data.FilePath != "" {
			entry := &domain.FileTreeEntry{
				Path:     event.Data.FilePath,
				Modified: event.Data.Operation == "update",
				Created:  event.Data.Operation == "create",
				Deleted:  event.Data.Operation == "delete",
			}
			fileTree[event.Data.FilePath] = entry

			if event.Data.Diff != "" {
				modifiedFiles = append(modifiedFiles, domain.FileDiffEntry{
					Path: event.Data.FilePath,
					Diff: event.Data.Diff,
				})
			}
		}

		if event.Duration > 0 {
			elapsedMs += event.Duration
		}
	}

	// Convert file tree map to slice
	fileTreeSlice := make([]domain.FileTreeEntry, 0, len(fileTree))
	for _, entry := range fileTree {
		fileTreeSlice = append(fileTreeSlice, *entry)
	}
	sort.Slice(fileTreeSlice, func(i, j int) bool {
		return fileTreeSlice[i].Path < fileTreeSlice[j].Path
	})

	state := &domain.DebugStepState{
		StepIndex:     stepIndex,
		Event:         timeline.Events[stepIndex],
		FileTree:      fileTreeSlice,
		ModifiedFiles: modifiedFiles,
		CostSoFar:     costSoFar,
		TokensSoFar:   tokensSoFar,
		ElapsedMs:     elapsedMs,
	}

	return state, nil
}

// AddAnnotation adds a developer note to a specific event within a debug session
func (s *DebugService) AddAnnotation(ctx context.Context, sessionID uuid.UUID, eventID string, userID uuid.UUID, content string) (*domain.DebugAnnotation, error) {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debug session: %w", err)
	}

	annotation := domain.DebugAnnotation{
		ID:        uuid.New(),
		EventID:   eventID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	session.Annotations = append(session.Annotations, annotation)
	session.UpdatedAt = time.Now()

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save annotation: %w", err)
	}

	s.logger.Info("added debug annotation",
		zap.String("sessionId", sessionID.String()),
		zap.String("eventId", eventID),
	)

	return &annotation, nil
}

// CompareTraces compares two trace replays, computing cost, latency, and token deltas,
// and identifying added, removed, and changed steps.
func (s *DebugService) CompareTraces(ctx context.Context, projectID uuid.UUID, traceIDA string, traceIDB string) (*domain.DebugCompareResult, error) {
	timelineA, err := s.replayService.GetTimelineForTrace(ctx, projectID, traceIDA)
	if err != nil {
		return nil, fmt.Errorf("failed to build timeline for trace A: %w", err)
	}

	timelineB, err := s.replayService.GetTimelineForTrace(ctx, projectID, traceIDB)
	if err != nil {
		return nil, fmt.Errorf("failed to build timeline for trace B: %w", err)
	}

	// Compute deltas
	costDelta := timelineB.Summary.TotalCost - timelineA.Summary.TotalCost
	tokenDelta := timelineB.Summary.TotalTokens - timelineA.Summary.TotalTokens
	latencyDelta := timelineB.Duration - timelineA.Duration

	// Identify differences by building event maps keyed by type+title
	eventsA := make(map[string]int)
	for i, e := range timelineA.Events {
		key := string(e.Type) + ":" + e.Title
		eventsA[key] = i
	}

	eventsB := make(map[string]int)
	for i, e := range timelineB.Events {
		key := string(e.Type) + ":" + e.Title
		eventsB[key] = i
	}

	var differences []domain.ComparisonDiff

	// Find removed steps (in A but not in B)
	for key, idxA := range eventsA {
		if _, ok := eventsB[key]; !ok {
			idx := idxA
			differences = append(differences, domain.ComparisonDiff{
				Type:    "removed",
				StepA:   &idx,
				Summary: fmt.Sprintf("Step removed: %s", timelineA.Events[idxA].Title),
			})
		}
	}

	// Find added steps (in B but not in A)
	for key, idxB := range eventsB {
		if _, ok := eventsA[key]; !ok {
			idx := idxB
			differences = append(differences, domain.ComparisonDiff{
				Type:    "added",
				StepB:   &idx,
				Summary: fmt.Sprintf("Step added: %s", timelineB.Events[idxB].Title),
			})
		}
	}

	// Find changed steps (in both but with different durations or costs)
	for key, idxA := range eventsA {
		if idxB, ok := eventsB[key]; ok {
			eA := timelineA.Events[idxA]
			eB := timelineB.Events[idxB]
			if eA.Duration != eB.Duration || eA.Data.Cost != eB.Data.Cost {
				a, b := idxA, idxB
				differences = append(differences, domain.ComparisonDiff{
					Type:    "changed",
					StepA:   &a,
					StepB:   &b,
					Summary: fmt.Sprintf("Step changed: %s (duration %dms->%dms)", eA.Title, eA.Duration, eB.Duration),
				})
			}
		}
	}

	result := &domain.DebugCompareResult{
		TraceA:       *timelineA,
		TraceB:       *timelineB,
		Differences:  differences,
		CostDelta:    costDelta,
		LatencyDelta: latencyDelta,
		TokenDelta:   tokenDelta,
	}

	return result, nil
}
