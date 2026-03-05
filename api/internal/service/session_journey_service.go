package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// SessionJourneyService handles session-based trace journey operations
type SessionJourneyService struct {
	logger *zap.Logger
	query  *QueryService
}

// NewSessionJourneyService creates a new session journey service
func NewSessionJourneyService(logger *zap.Logger, query *QueryService) *SessionJourneyService {
	return &SessionJourneyService{
		logger: logger,
		query:  query,
	}
}

// GetJourney fetches all traces for a session and auto-detects phases
func (s *SessionJourneyService) GetJourney(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.SessionJourney, error) {
	s.logger.Info("building session journey",
		zap.String("projectId", projectID.String()),
		zap.String("sessionId", sessionID),
	)

	traces, err := s.query.GetSessionTraces(ctx, projectID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch session traces: %w", err)
	}

	sort.Slice(traces, func(i, j int) bool {
		return traces[i].StartTime.Before(traces[j].StartTime)
	})

	phases := s.detectPhases(traces)

	var totalDuration int64
	var totalCost float64
	var totalTokens int64
	status := "completed"

	for _, t := range traces {
		totalDuration += int64(t.DurationMs)
		totalCost += t.TotalCost
		totalTokens += t.TotalTokens
		if t.Level == domain.LevelError {
			status = "failed"
		}
	}

	// If any trace has no end time, session is still active
	if len(traces) > 0 && traces[len(traces)-1].EndTime == nil {
		status = "active"
	}

	journey := &domain.SessionJourney{
		ID:            uuid.New(),
		ProjectID:     projectID,
		SessionID:     sessionID,
		Phases:        phases,
		TotalDuration: totalDuration,
		TotalCost:     totalCost,
		TotalTokens:   totalTokens,
		TraceCount:    len(traces),
		Status:        status,
		DetectedAt:    time.Now(),
	}

	s.logger.Info("session journey built",
		zap.String("sessionId", sessionID),
		zap.Int("phases", len(phases)),
		zap.Int("traces", len(traces)),
	)

	return journey, nil
}

// GetPhases returns just the workflow phases for a session
func (s *SessionJourneyService) GetPhases(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.WorkflowPhase, error) {
	s.logger.Info("fetching session phases",
		zap.String("projectId", projectID.String()),
		zap.String("sessionId", sessionID),
	)

	journey, err := s.GetJourney(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}

	return journey.Phases, nil
}

// ListRecentJourneys returns recent session journeys for a project
func (s *SessionJourneyService) ListRecentJourneys(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.SessionJourney, error) {
	s.logger.Info("listing recent journeys",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
	)

	if limit <= 0 {
		limit = 10
	}

	sessions, err := s.query.ListSessions(ctx, &domain.SessionFilter{
		ProjectID: projectID,
	}, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	journeys := make([]domain.SessionJourney, 0, len(sessions.Sessions))
	for _, sess := range sessions.Sessions {
		journey, err := s.GetJourney(ctx, projectID, sess.ID)
		if err != nil {
			s.logger.Warn("failed to build journey for session",
				zap.String("sessionId", sess.ID),
				zap.Error(err),
			)
			continue
		}
		journeys = append(journeys, *journey)
	}

	return journeys, nil
}

func (s *SessionJourneyService) detectPhases(traces []domain.Trace) []domain.WorkflowPhase {
	if len(traces) == 0 {
		return []domain.WorkflowPhase{}
	}

	type phaseGroup struct {
		name       string
		traceIDs   []string
		startTime  time.Time
		endTime    time.Time
		cost       float64
		tokens     int64
		errorCount int
		toolCalls  int
		files      int
		confidence float64
	}

	var groups []*phaseGroup
	var current *phaseGroup

	for i, trace := range traces {
		phaseName := s.classifyTrace(trace, i, len(traces))

		// Start a new phase if the name changes or there's a significant time gap
		startNew := current == nil || current.name != phaseName
		if !startNew && i > 0 {
			gap := trace.StartTime.Sub(traces[i-1].StartTime)
			if gap > 5*time.Minute {
				startNew = true
			}
		}

		if startNew {
			if current != nil {
				groups = append(groups, current)
			}
			current = &phaseGroup{
				name:      phaseName,
				startTime: trace.StartTime,
			}
		}

		current.traceIDs = append(current.traceIDs, trace.ID)
		if trace.EndTime != nil {
			current.endTime = *trace.EndTime
		} else {
			current.endTime = trace.StartTime.Add(time.Duration(trace.DurationMs) * time.Millisecond)
		}
		current.cost += trace.TotalCost
		current.tokens += trace.TotalTokens
		if trace.Level == domain.LevelError {
			current.errorCount++
		}
		current.confidence = s.classifyConfidence(trace, phaseName)
	}
	if current != nil {
		groups = append(groups, current)
	}

	phases := make([]domain.WorkflowPhase, 0, len(groups))
	for _, g := range groups {
		endTime := g.endTime
		phase := domain.WorkflowPhase{
			Name:       g.name,
			StartTime:  g.startTime,
			EndTime:    &endTime,
			DurationMs: endTime.Sub(g.startTime).Milliseconds(),
			TraceIDs:   g.traceIDs,
			Metrics: domain.PhaseMetrics{
				Cost:          g.cost,
				Tokens:        g.tokens,
				ErrorCount:    g.errorCount,
				ToolCallCount: g.toolCalls,
				FilesModified: g.files,
			},
			Confidence: g.confidence,
		}
		phases = append(phases, phase)
	}

	return phases
}

func (s *SessionJourneyService) classifyTrace(trace domain.Trace, index, total int) string {
	name := strings.ToLower(trace.Name)

	// Check for debugging phase (errors or debug keywords)
	if trace.Level == domain.LevelError || strings.Contains(name, "debug") || strings.Contains(name, "fix") {
		return "debugging"
	}

	// Check for review phase
	if strings.Contains(name, "review") || strings.Contains(name, "evaluate") {
		return "review"
	}

	// Check for testing phase
	if strings.Contains(name, "test") || strings.Contains(name, "verify") || strings.Contains(name, "check") {
		return "testing"
	}

	// Check for planning phase (early traces or planning keywords)
	if strings.Contains(name, "plan") || strings.Contains(name, "think") || strings.Contains(name, "reason") {
		return "planning"
	}

	// Early traces default to planning
	if total > 3 && index < total/5 {
		return "planning"
	}

	// Check for implementation phase (file modifications, code generation)
	if strings.Contains(name, "code") || strings.Contains(name, "generate") ||
		strings.Contains(name, "write") || strings.Contains(name, "edit") ||
		strings.Contains(name, "create") || strings.Contains(name, "implement") {
		return "implementation"
	}

	// Default middle traces to implementation
	return "implementation"
}

func (s *SessionJourneyService) classifyConfidence(trace domain.Trace, phaseName string) float64 {
	name := strings.ToLower(trace.Name)

	// High confidence when keywords directly match
	switch phaseName {
	case "planning":
		if strings.Contains(name, "plan") || strings.Contains(name, "think") || strings.Contains(name, "reason") {
			return 0.9
		}
	case "implementation":
		if strings.Contains(name, "code") || strings.Contains(name, "generate") || strings.Contains(name, "write") {
			return 0.9
		}
	case "testing":
		if strings.Contains(name, "test") || strings.Contains(name, "verify") {
			return 0.9
		}
	case "debugging":
		if trace.Level == domain.LevelError || strings.Contains(name, "debug") {
			return 0.95
		}
	case "review":
		if strings.Contains(name, "review") || strings.Contains(name, "evaluate") {
			return 0.9
		}
	}

	// Lower confidence for positional heuristics
	return 0.6
}
