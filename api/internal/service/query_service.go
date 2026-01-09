package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// QueryService handles trace and observation queries
type QueryService struct {
	traceRepo       TraceRepository
	observationRepo ObservationRepository
	scoreRepo       ScoreRepository
	sessionRepo     SessionRepository
}

// NewQueryService creates a new query service
func NewQueryService(
	traceRepo TraceRepository,
	observationRepo ObservationRepository,
	scoreRepo ScoreRepository,
	sessionRepo SessionRepository,
) *QueryService {
	return &QueryService{
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		scoreRepo:       scoreRepo,
		sessionRepo:     sessionRepo,
	}
}

// GetTrace retrieves a trace by ID with observations and scores
func (s *QueryService) GetTrace(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	trace, err := s.traceRepo.GetByID(ctx, projectID, traceID)
	if err != nil {
		return nil, err
	}

	// Load observations
	observations, err := s.observationRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observations: %w", err)
	}
	trace.Observations = observations

	// Load scores
	scores, err := s.scoreRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scores: %w", err)
	}
	trace.Scores = scores

	return trace, nil
}

// ListTraces retrieves traces with filtering and pagination
func (s *QueryService) ListTraces(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	return s.traceRepo.List(ctx, filter, limit, offset)
}

// GetObservation retrieves an observation by ID
func (s *QueryService) GetObservation(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error) {
	return s.observationRepo.GetByID(ctx, projectID, observationID)
}

// ListObservations retrieves observations with filtering
func (s *QueryService) ListObservations(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error) {
	return s.observationRepo.List(ctx, filter, limit, offset)
}

// GetObservationsByTraceID retrieves all observations for a trace
func (s *QueryService) GetObservationsByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error) {
	return s.observationRepo.GetByTraceID(ctx, projectID, traceID)
}

// GetObservationTree retrieves the observation tree for a trace
func (s *QueryService) GetObservationTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error) {
	return s.observationRepo.GetTree(ctx, projectID, traceID)
}

// GetSessionTraces retrieves traces for a session
func (s *QueryService) GetSessionTraces(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	return s.traceRepo.GetBySessionID(ctx, projectID, sessionID)
}

// ListSessions retrieves sessions with filtering and pagination
func (s *QueryService) ListSessions(ctx context.Context, filter *domain.SessionFilter, limit, offset int) (*domain.SessionList, error) {
	return s.sessionRepo.List(ctx, filter, limit, offset)
}

// GetSession retrieves a session by ID with aggregated metrics
func (s *QueryService) GetSession(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, projectID, sessionID)
	if err != nil {
		return nil, err
	}

	// Optionally load traces for the session
	traces, err := s.traceRepo.GetBySessionID(ctx, projectID, sessionID)
	if err == nil {
		session.Traces = traces
	}

	return session, nil
}

// SetBookmark sets the bookmark status of a trace
func (s *QueryService) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	return s.traceRepo.SetBookmark(ctx, projectID, traceID, bookmarked)
}

// UpdateTrace updates a trace with the given input
func (s *QueryService) UpdateTrace(ctx context.Context, projectID uuid.UUID, traceID string, input *domain.TraceUpdateInput) (*domain.Trace, error) {
	// Get existing trace
	trace, err := s.traceRepo.GetByID(ctx, projectID, traceID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Name != nil {
		trace.Name = *input.Name
	}
	if input.UserID != nil {
		trace.UserID = *input.UserID
	}
	if input.SessionID != nil {
		trace.SessionID = *input.SessionID
	}
	if input.Release != nil {
		trace.Release = *input.Release
	}
	if input.Version != nil {
		trace.Version = *input.Version
	}
	if input.Tags != nil {
		trace.Tags = input.Tags
	}
	if input.Public != nil {
		trace.Public = *input.Public
	}
	if input.Level != nil {
		trace.Level = *input.Level
	}
	if input.StatusMessage != nil {
		trace.StatusMessage = *input.StatusMessage
	}
	if input.EndTime != nil {
		trace.EndTime = input.EndTime
		if trace.StartTime.Before(*input.EndTime) {
			trace.DurationMs = float64(input.EndTime.Sub(trace.StartTime).Milliseconds())
		}
	}
	if input.Bookmarked != nil {
		trace.Bookmarked = *input.Bookmarked
	}
	if input.GitCommitSha != nil {
		trace.GitCommitSha = *input.GitCommitSha
	}
	if input.GitBranch != nil {
		trace.GitBranch = *input.GitBranch
	}
	if input.GitRepoURL != nil {
		trace.GitRepoURL = *input.GitRepoURL
	}

	// Update in repository
	if err := s.traceRepo.Update(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to update trace: %w", err)
	}

	return trace, nil
}

// GetTraceStats calculates statistics for traces matching a filter
func (s *QueryService) GetTraceStats(ctx context.Context, filter *domain.TraceFilter) (*TraceStats, error) {
	list, err := s.traceRepo.List(ctx, filter, 10000, 0)
	if err != nil {
		return nil, err
	}

	stats := &TraceStats{
		TotalCount: list.TotalCount,
	}

	if len(list.Traces) == 0 {
		return stats, nil
	}

	var totalDuration, totalCost float64
	var totalTokens uint64
	var errorCount int64

	for _, trace := range list.Traces {
		totalDuration += trace.DurationMs
		totalCost += trace.TotalCost
		totalTokens += trace.TotalTokens
		if trace.Level == domain.LevelError {
			errorCount++
		}
	}

	stats.AvgDuration = totalDuration / float64(len(list.Traces))
	stats.TotalCost = totalCost
	stats.TotalTokens = totalTokens
	stats.ErrorCount = errorCount
	stats.ErrorRate = float64(errorCount) / float64(len(list.Traces)) * 100

	return stats, nil
}

// TraceStats represents aggregated trace statistics
type TraceStats struct {
	TotalCount  int64   `json:"totalCount"`
	AvgDuration float64 `json:"avgDuration"`
	TotalCost   float64 `json:"totalCost"`
	TotalTokens uint64  `json:"totalTokens"`
	ErrorCount  int64   `json:"errorCount"`
	ErrorRate   float64 `json:"errorRate"`
}

// GetGenerationStats calculates statistics for generations
func (s *QueryService) GetGenerationStats(ctx context.Context, projectID uuid.UUID, model *string) (*GenerationStats, error) {
	genType := domain.ObservationTypeGeneration
	filter := &domain.ObservationFilter{
		ProjectID: projectID,
		Type:      &genType,
	}
	if model != nil {
		filter.Model = model
	}

	observations, _, err := s.observationRepo.List(ctx, filter, 10000, 0)
	if err != nil {
		return nil, err
	}

	stats := &GenerationStats{
		TotalCount: int64(len(observations)),
		ByModel:    make(map[string]*ModelStats),
	}

	for _, obs := range observations {
		if obs.Model == "" {
			continue
		}

		modelStats, ok := stats.ByModel[obs.Model]
		if !ok {
			modelStats = &ModelStats{Model: obs.Model}
			stats.ByModel[obs.Model] = modelStats
		}

		modelStats.Count++
		modelStats.TotalLatency += obs.DurationMs
		modelStats.TotalInputTokens += int64(obs.UsageDetails.InputTokens)
		modelStats.TotalOutputTokens += int64(obs.UsageDetails.OutputTokens)
		modelStats.TotalCost += obs.CostDetails.TotalCost
	}

	// Calculate averages
	for _, modelStats := range stats.ByModel {
		if modelStats.Count > 0 {
			modelStats.AvgLatency = modelStats.TotalLatency / float64(modelStats.Count)
		}
	}

	return stats, nil
}

// GenerationStats represents aggregated generation statistics
type GenerationStats struct {
	TotalCount int64                 `json:"totalCount"`
	ByModel    map[string]*ModelStats `json:"byModel"`
}

// ModelStats represents statistics for a specific model
type ModelStats struct {
	Model             string  `json:"model"`
	Count             int64   `json:"count"`
	TotalLatency      float64 `json:"totalLatency"`
	AvgLatency        float64 `json:"avgLatency"`
	TotalInputTokens  int64   `json:"totalInputTokens"`
	TotalOutputTokens int64   `json:"totalOutputTokens"`
	TotalCost         float64 `json:"totalCost"`
}
