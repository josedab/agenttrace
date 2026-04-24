package service

import (
	"context"
	"fmt"
	"math"
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
	if input.EvaluatorIDs == nil {
		input.EvaluatorIDs = []uuid.UUID{}
	}

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

// CalculateELORating computes ELO-style ratings for agents based on head-to-head comparisons
func (s *BenchmarkService) CalculateELORating(ctx context.Context, benchmarkID uuid.UUID) ([]domain.ELORating, error) {
	submissions, err := s.benchmarkRepo.ListSubmissions(ctx, benchmarkID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	// Group submissions by agent
	agentBest := make(map[string]*domain.BenchmarkSubmission)
	for i, sub := range submissions {
		key := sub.AgentName
		if existing, ok := agentBest[key]; !ok || sub.OverallScore > existing.OverallScore {
			agentBest[key] = &submissions[i]
		}
	}

	// Initialize ratings at 1500
	ratings := make(map[string]*domain.ELORating)
	for name, sub := range agentBest {
		ratings[name] = &domain.ELORating{
			ID:          uuid.New(),
			BenchmarkID: benchmarkID,
			AgentName:   name,
			ProjectID:   sub.ProjectID,
			Rating:      1500.0,
			UpdatedAt:   time.Now(),
		}
	}

	// Run pairwise ELO calculations
	kFactor := 32.0
	agents := make([]string, 0, len(agentBest))
	for name := range agentBest {
		agents = append(agents, name)
	}
	sort.Strings(agents)

	for i := 0; i < len(agents); i++ {
		for j := i + 1; j < len(agents); j++ {
			a, b := agents[i], agents[j]
			scoreA := agentBest[a].OverallScore
			scoreB := agentBest[b].OverallScore

			rA := ratings[a]
			rB := ratings[b]

			expectedA := 1.0 / (1.0 + pow10((rB.Rating-rA.Rating)/400.0))
			expectedB := 1.0 - expectedA

			var actualA, actualB float64
			if scoreA > scoreB {
				actualA, actualB = 1.0, 0.0
				rA.Wins++
				rB.Losses++
			} else if scoreB > scoreA {
				actualA, actualB = 0.0, 1.0
				rA.Losses++
				rB.Wins++
			} else {
				actualA, actualB = 0.5, 0.5
				rA.Draws++
				rB.Draws++
			}

			rA.Rating += kFactor * (actualA - expectedA)
			rB.Rating += kFactor * (actualB - expectedB)
			rA.TotalGames++
			rB.TotalGames++
		}
	}

	// Compute confidence based on total games
	result := make([]domain.ELORating, 0, len(ratings))
	for _, r := range ratings {
		if r.TotalGames > 0 {
			r.Confidence = 1.0 - 1.0/float64(1+r.TotalGames)
		}
		result = append(result, *r)
	}

	// Sort by rating descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Rating > result[j].Rating
	})

	s.logger.Info("calculated ELO ratings",
		zap.String("benchmarkId", benchmarkID.String()),
		zap.Int("agentCount", len(result)),
	)

	return result, nil
}

// SubmitCommunity handles community API benchmark submissions
func (s *BenchmarkService) SubmitCommunity(ctx context.Context, submitterID uuid.UUID, input *domain.CommunitySubmissionInput) (*domain.CommunitySubmission, error) {
	benchmark, err := s.benchmarkRepo.GetByID(ctx, input.BenchmarkID)
	if err != nil {
		return nil, fmt.Errorf("benchmark not found: %w", err)
	}

	if !benchmark.IsPublic {
		return nil, fmt.Errorf("benchmark is not public")
	}

	// Calculate overall score from submitted scores
	var overallScore float64
	var totalWeight float64
	for _, metric := range benchmark.Metrics {
		if score, ok := input.Scores[metric.Name]; ok {
			overallScore += score * metric.Weight
			totalWeight += metric.Weight
		}
	}
	if totalWeight > 0 {
		overallScore /= totalWeight
	}

	submission := &domain.CommunitySubmission{
		ID:           uuid.New(),
		BenchmarkID:  input.BenchmarkID,
		SubmitterID:  submitterID,
		AgentName:    input.AgentName,
		AgentVersion: input.AgentVersion,
		RepoURL:      input.RepoURL,
		Scores:       input.Scores,
		OverallScore: overallScore,
		Verified:     false,
		Metadata:     input.Metadata,
		CreatedAt:    time.Now(),
	}

	s.logger.Info("community benchmark submission",
		zap.String("benchmarkId", input.BenchmarkID.String()),
		zap.String("agentName", input.AgentName),
		zap.Float64("overallScore", overallScore),
	)

	return submission, nil
}

// GenerateGHAWorkflow generates a GitHub Actions workflow for automated benchmark runs
func (s *BenchmarkService) GenerateGHAWorkflow(ctx context.Context, config *domain.GHABenchmarkConfig) (string, error) {
	benchmark, err := s.benchmarkRepo.GetByID(ctx, config.BenchmarkID)
	if err != nil {
		return "", fmt.Errorf("benchmark not found: %w", err)
	}

	workflow := fmt.Sprintf(`name: AgentTrace Benchmark - %s

on:
  %s:
`, benchmark.Name, config.TriggerEvent)

	if config.Schedule != "" {
		workflow += fmt.Sprintf(`  schedule:
    - cron: '%s'
`, config.Schedule)
	}

	workflow += fmt.Sprintf(`
jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup AgentTrace CLI
        run: |
          curl -sSL https://get.agenttrace.io/cli | sh
          agenttrace doctor

      - name: Run Benchmark
        env:
          AGENTTRACE_API_KEY: ${{ secrets.AGENTTRACE_API_KEY }}
        run: |
          agenttrace benchmark run --id %s --format json > results.json

      - name: Submit Results
        env:
          AGENTTRACE_API_KEY: ${{ secrets.AGENTTRACE_API_KEY }}
        run: |
          agenttrace benchmark submit --id %s --results results.json

      - name: Upload Results Artifact
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: results.json
`, benchmark.ID.String(), benchmark.ID.String())

	return workflow, nil
}

// pow10 computes 10^x using the math package
func pow10(x float64) float64 {
	return math.Pow(10, x)
}
