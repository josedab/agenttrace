package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ScorecardRepository defines repository operations for agent scorecards
type ScorecardRepository interface {
	Save(ctx context.Context, scorecard *domain.AgentScorecard) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentScorecard, error)
	List(ctx context.Context, projectID uuid.UUID, agentName string) ([]domain.AgentScorecard, error)
	SaveConfig(ctx context.Context, config *domain.ScorecardConfig) error
	GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.ScorecardConfig, error)
}

// ScorecardService generates automated agent performance scorecards by
// aggregating trace metrics and computing period-over-period trends.
type ScorecardService struct {
	logger       *zap.Logger
	scorecardRepo ScorecardRepository
	queryService *QueryService
}

// NewScorecardService creates a new scorecard service
func NewScorecardService(
	logger *zap.Logger,
	scorecardRepo ScorecardRepository,
	queryService *QueryService,
) *ScorecardService {
	return &ScorecardService{
		logger:        logger,
		scorecardRepo: scorecardRepo,
		queryService:  queryService,
	}
}

// Generate creates a scorecard for the given agent and period by querying
// trace data, computing metrics and trends, assigning a grade, and
// generating a summary.
func (s *ScorecardService) Generate(ctx context.Context, projectID uuid.UUID, input domain.ScorecardInput) (*domain.AgentScorecard, error) {
	now := time.Now()
	periodStart, periodEnd := s.computePeriodBounds(now, input.Period)

	// Query traces for the current period
	currentMetrics, err := s.computeMetrics(ctx, projectID, input.AgentName, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to compute current period metrics: %w", err)
	}

	// Query traces for the previous period (for trend comparison)
	prevDuration := periodEnd.Sub(periodStart)
	prevStart := periodStart.Add(-prevDuration)
	prevEnd := periodStart
	prevMetrics, err := s.computeMetrics(ctx, projectID, input.AgentName, prevStart, prevEnd)
	if err != nil {
		s.logger.Warn("failed to compute previous period metrics, trends will be zero",
			zap.Error(err),
		)
		prevMetrics = &domain.ScorecardMetrics{}
	}

	trends := s.computeTrends(currentMetrics, prevMetrics)
	grade := s.assignGrade(currentMetrics)
	summary := s.generateSummary(input.AgentName, input.Period, currentMetrics, grade)

	scorecard := &domain.AgentScorecard{
		ID:          uuid.New(),
		ProjectID:   projectID,
		AgentName:   input.AgentName,
		Period:      input.Period,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Metrics:     *currentMetrics,
		Trends:      trends,
		Grade:       grade,
		Summary:     summary,
		CreatedAt:   now,
	}

	if err := s.scorecardRepo.Save(ctx, scorecard); err != nil {
		return nil, fmt.Errorf("failed to save scorecard: %w", err)
	}

	s.logger.Info("generated agent scorecard",
		zap.String("projectId", projectID.String()),
		zap.String("agentName", input.AgentName),
		zap.String("grade", grade),
	)

	return scorecard, nil
}

// GetScorecard retrieves a scorecard by ID.
func (s *ScorecardService) GetScorecard(ctx context.Context, scorecardID uuid.UUID) (*domain.AgentScorecard, error) {
	scorecard, err := s.scorecardRepo.GetByID(ctx, scorecardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scorecard: %w", err)
	}
	return scorecard, nil
}

// ListScorecards retrieves scorecards for a project, optionally filtered by
// agent name.
func (s *ScorecardService) ListScorecards(ctx context.Context, projectID uuid.UUID, agentName string) ([]domain.AgentScorecard, error) {
	scorecards, err := s.scorecardRepo.List(ctx, projectID, agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to list scorecards: %w", err)
	}
	return scorecards, nil
}

// ConfigureAutoGeneration saves or updates the auto-generation configuration
// for scorecards.
func (s *ScorecardService) ConfigureAutoGeneration(ctx context.Context, config domain.ScorecardConfig) error {
	if err := s.scorecardRepo.SaveConfig(ctx, &config); err != nil {
		return fmt.Errorf("failed to save scorecard config: %w", err)
	}

	s.logger.Info("configured scorecard auto-generation",
		zap.String("projectId", config.ProjectID.String()),
		zap.String("agentName", config.AgentName),
		zap.String("period", config.Period),
		zap.Bool("enabled", config.Enabled),
	)

	return nil
}

// GetConfig retrieves the scorecard auto-generation configuration for a
// project.
func (s *ScorecardService) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.ScorecardConfig, error) {
	config, err := s.scorecardRepo.GetConfig(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scorecard config: %w", err)
	}
	return config, nil
}

// computePeriodBounds returns the start and end times for the given period
// relative to the reference time.
func (s *ScorecardService) computePeriodBounds(ref time.Time, period string) (time.Time, time.Time) {
	switch period {
	case "weekly":
		end := ref.Truncate(24 * time.Hour)
		start := end.AddDate(0, 0, -7)
		return start, end
	case "monthly":
		end := ref.Truncate(24 * time.Hour)
		start := end.AddDate(0, -1, 0)
		return start, end
	default:
		end := ref.Truncate(24 * time.Hour)
		start := end.AddDate(0, 0, -7)
		return start, end
	}
}

// computeMetrics aggregates trace data for the given agent and time range.
func (s *ScorecardService) computeMetrics(ctx context.Context, projectID uuid.UUID, agentName string, start, end time.Time) (*domain.ScorecardMetrics, error) {
	filter := &domain.TraceFilter{
		ProjectID: projectID,
		Name:      &agentName,
		FromTime:  &start,
		ToTime:    &end,
	}

	traces, err := s.queryService.ListTraces(ctx, filter, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list traces: %w", err)
	}

	metrics := &domain.ScorecardMetrics{}
	if traces == nil || len(traces.Traces) == 0 {
		return metrics, nil
	}

	var totalLatency float64
	var totalTokens uint64
	var totalCostCents int64
	var errorCount int
	latencies := make([]float64, 0, len(traces.Traces))

	for _, t := range traces.Traces {
		metrics.TotalTraces++
		totalLatency += t.DurationMs
		totalTokens += t.TotalTokens
		totalCostCents += int64(t.TotalCost * 100)
		latencies = append(latencies, t.DurationMs)

		if t.Level == domain.LevelError {
			errorCount++
		}
	}

	if metrics.TotalTraces > 0 {
		metrics.AvgLatencyMs = totalLatency / float64(metrics.TotalTraces)
		metrics.AvgTokensPerTrace = int(totalTokens / uint64(metrics.TotalTraces))
		metrics.TotalCostCents = totalCostCents
		metrics.CostPerTrace = float64(totalCostCents) / float64(metrics.TotalTraces) / 100
		metrics.ErrorRate = float64(errorCount) / float64(metrics.TotalTraces)
		metrics.SuccessRate = 1.0 - metrics.ErrorRate
		metrics.P95LatencyMs = s.percentile(latencies, 0.95)
	}

	return metrics, nil
}

// computeTrends calculates period-over-period deltas.
func (s *ScorecardService) computeTrends(current, previous *domain.ScorecardMetrics) domain.ScorecardTrends {
	return domain.ScorecardTrends{
		SuccessRateDelta: current.SuccessRate - previous.SuccessRate,
		LatencyDelta:     current.AvgLatencyMs - previous.AvgLatencyMs,
		CostDelta:        current.CostPerTrace - previous.CostPerTrace,
		ErrorRateDelta:   current.ErrorRate - previous.ErrorRate,
		VolumeChange:     safeDelta(float64(previous.TotalTraces), float64(current.TotalTraces)),
	}
}

// assignGrade assigns a letter grade based on success rate and cost per trace.
// A: >90% success and <$0.10/trace, B: >80%, C: >70%, D: >60%, F: <60%.
func (s *ScorecardService) assignGrade(m *domain.ScorecardMetrics) string {
	if m.SuccessRate > 0.90 && m.CostPerTrace < 0.10 {
		return "A"
	}
	if m.SuccessRate > 0.80 {
		return "B"
	}
	if m.SuccessRate > 0.70 {
		return "C"
	}
	if m.SuccessRate > 0.60 {
		return "D"
	}
	return "F"
}

// generateSummary creates a human-readable summary for the scorecard.
func (s *ScorecardService) generateSummary(agentName, period string, m *domain.ScorecardMetrics, grade string) string {
	return fmt.Sprintf(
		"Agent %q received a grade of %s for the %s period. "+
			"Processed %d traces with a %.1f%% success rate, "+
			"avg latency of %.0fms, and $%.4f cost per trace.",
		agentName, grade, period,
		m.TotalTraces, m.SuccessRate*100,
		m.AvgLatencyMs, m.CostPerTrace,
	)
}

// percentile returns the p-th percentile from a sorted slice of values.
func (s *ScorecardService) percentile(values []float64, p float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	// Simple nearest-rank method
	sorted := make([]float64, n)
	copy(sorted, values)
	// Insertion sort for simplicity
	for i := 1; i < n; i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	rank := int(math.Ceil(p * float64(n)))
	if rank >= n {
		rank = n - 1
	}
	return sorted[rank]
}

// safeDelta computes a percentage change, returning 0 when the base is zero.
func safeDelta(base, current float64) float64 {
	if base == 0 {
		if current == 0 {
			return 0
		}
		return 1.0
	}
	return (current - base) / base
}
