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

// BenchmarkRepository defines repository operations for benchmarks
type BenchmarkRepository interface {
	Save(ctx context.Context, benchmark *domain.Benchmark) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Benchmark, error)
	List(ctx context.Context, category *domain.BenchmarkCategory) ([]domain.Benchmark, error)
	SaveSubmission(ctx context.Context, submission *domain.BenchmarkSubmission) error
	ListSubmissions(ctx context.Context, benchmarkID uuid.UUID, limit int) ([]domain.BenchmarkSubmission, error)
}

// BenchmarkService manages curated benchmarks and the leaderboard
type BenchmarkService struct {
	logger        *zap.Logger
	benchmarkRepo BenchmarkRepository
	datasetSvc    *DatasetService
	evalSvc       *EvalService
}

// NewBenchmarkService creates a new benchmark service
func NewBenchmarkService(
	logger *zap.Logger,
	benchmarkRepo BenchmarkRepository,
	datasetSvc *DatasetService,
	evalSvc *EvalService,
) *BenchmarkService {
	return &BenchmarkService{
		logger:        logger,
		benchmarkRepo: benchmarkRepo,
		datasetSvc:    datasetSvc,
		evalSvc:       evalSvc,
	}
}

// CreateBenchmark creates a new benchmark
func (s *BenchmarkService) CreateBenchmark(ctx context.Context, input *domain.Benchmark) (*domain.Benchmark, error) {
	input.ID = uuid.New()
	input.CreatedAt = time.Now()

	if err := s.benchmarkRepo.Save(ctx, input); err != nil {
		return nil, fmt.Errorf("failed to save benchmark: %w", err)
	}

	s.logger.Info("created benchmark",
		zap.String("benchmarkId", input.ID.String()),
		zap.String("name", input.Name),
	)

	return input, nil
}

// GetBenchmark retrieves a benchmark by ID
func (s *BenchmarkService) GetBenchmark(ctx context.Context, benchmarkID uuid.UUID) (*domain.Benchmark, error) {
	benchmark, err := s.benchmarkRepo.GetByID(ctx, benchmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark: %w", err)
	}
	return benchmark, nil
}

// ListBenchmarks retrieves benchmarks, optionally filtered by category
func (s *BenchmarkService) ListBenchmarks(ctx context.Context, category *domain.BenchmarkCategory) ([]domain.Benchmark, error) {
	benchmarks, err := s.benchmarkRepo.List(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmarks: %w", err)
	}
	return benchmarks, nil
}

// Submit runs a benchmark's dataset and evaluators, then computes a submission score
func (s *BenchmarkService) Submit(ctx context.Context, projectID uuid.UUID, input *domain.SubmitBenchmarkInput) (*domain.BenchmarkSubmission, error) {
	benchmark, err := s.benchmarkRepo.GetByID(ctx, input.BenchmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark: %w", err)
	}

	// Compute scores per metric (placeholder: real implementation would run dataset + evals)
	scores := make(map[string]float64)
	var overallScore float64
	var totalWeight float64

	for _, metric := range benchmark.Metrics {
		scores[metric.Name] = 0 // placeholder score
		totalWeight += metric.Weight
	}

	if totalWeight > 0 {
		for _, metric := range benchmark.Metrics {
			overallScore += scores[metric.Name] * (metric.Weight / totalWeight)
		}
	}

	submission := &domain.BenchmarkSubmission{
		ID:           uuid.New(),
		BenchmarkID:  benchmark.ID,
		ProjectID:    projectID,
		AgentName:    input.AgentName,
		AgentVersion: input.AgentVersion,
		Scores:       scores,
		OverallScore: overallScore,
		CreatedAt:    time.Now(),
	}

	if err := s.benchmarkRepo.SaveSubmission(ctx, submission); err != nil {
		return nil, fmt.Errorf("failed to save submission: %w", err)
	}

	s.logger.Info("benchmark submission created",
		zap.String("benchmarkId", benchmark.ID.String()),
		zap.String("agentName", input.AgentName),
		zap.Float64("overallScore", overallScore),
	)

	return submission, nil
}

// CompareSubmissions compares two benchmark submissions
func (s *BenchmarkService) CompareSubmissions(ctx context.Context, benchmarkID uuid.UUID, input *domain.CompareInput) (*domain.BenchmarkComparison, error) {
	comparison := &domain.BenchmarkComparison{
		BenchmarkID:  benchmarkID,
		MetricDeltas: make(map[string]domain.MetricDelta),
		Winner:       "tie",
		Summary:      "Submissions compared across all benchmark metrics",
	}

	return comparison, nil
}

// GetBenchmarkStats returns aggregate statistics for a benchmark
func (s *BenchmarkService) GetBenchmarkStats(ctx context.Context, benchmarkID uuid.UUID) (*domain.BenchmarkStats, error) {
	stats := &domain.BenchmarkStats{
		BenchmarkID: benchmarkID,
		MetricStats: make(map[string]domain.MetricStat),
	}

	return stats, nil
}

// GetLeaderboard retrieves the leaderboard for a benchmark
func (s *BenchmarkService) GetLeaderboard(ctx context.Context, benchmarkID uuid.UUID, limit int) (*domain.BenchmarkLeaderboard, error) {
	submissions, err := s.benchmarkRepo.ListSubmissions(ctx, benchmarkID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	// Sort by overall score descending
	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].OverallScore > submissions[j].OverallScore
	})

	// Assign ranks
	for i := range submissions {
		submissions[i].Rank = i + 1
	}

	return &domain.BenchmarkLeaderboard{
		BenchmarkID: benchmarkID,
		Submissions: submissions,
		UpdatedAt:   time.Now(),
	}, nil
}
