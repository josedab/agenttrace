package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ABTestingService handles A/B testing operations
type ABTestingService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	tests   map[uuid.UUID]*domain.PromptABTest
	results map[uuid.UUID][]domain.PromptABRecordResultInput // testID -> results
	assigns map[string]uuid.UUID                              // assignmentKey -> variantID (sticky)
}

// NewABTestingService creates a new A/B testing service
func NewABTestingService(logger *zap.Logger) *ABTestingService {
	return &ABTestingService{
		logger:  logger,
		tests:   make(map[uuid.UUID]*domain.PromptABTest),
		results: make(map[uuid.UUID][]domain.PromptABRecordResultInput),
		assigns: make(map[string]uuid.UUID),
	}
}

// CreateTest creates a new A/B test
func (s *ABTestingService) CreateTest(ctx context.Context, projectID uuid.UUID, input *domain.PromptABTestInput) (*domain.PromptABTest, error) {
	if input.Name == "" {
		return nil, apperrors.Validation("test name is required")
	}
	if len(input.Variants) < 2 {
		return nil, apperrors.Validation("at least 2 variants are required")
	}
	if input.TargetMetric == "" {
		return nil, apperrors.Validation("target metric is required")
	}

	// Validate traffic percentages sum to 100
	var totalTraffic float64
	for _, v := range input.Variants {
		totalTraffic += v.TrafficPercent
	}
	if math.Abs(totalTraffic-100.0) > 0.01 {
		return nil, apperrors.Validation("variant traffic percentages must sum to 100")
	}

	now := time.Now()
	testID := uuid.New()

	variants := make([]domain.PromptABTestVariant, len(input.Variants))
	for i, v := range input.Variants {
		variants[i] = domain.PromptABTestVariant{
			ID:              uuid.New(),
			Name:            v.Name,
			PromptVersionID: v.PromptVersionID,
			TrafficPercent:  v.TrafficPercent,
			IsControl:       v.IsControl,
			SampleCount:     0,
			Metrics:         domain.ABTestVariantMetrics{},
		}
	}

	trafficSplit := domain.PromptTrafficSplit{Method: "percentage"}
	if input.TrafficSplit != nil {
		trafficSplit = *input.TrafficSplit
	}

	confidenceLevel := 0.95
	if input.ConfidenceLevel > 0 {
		confidenceLevel = input.ConfidenceLevel
	}

	minSampleSize := 100
	if input.MinSampleSize > 0 {
		minSampleSize = input.MinSampleSize
	}

	test := &domain.PromptABTest{
		ID:               testID,
		ProjectID:        projectID,
		Name:             input.Name,
		Description:      input.Description,
		PromptID:         input.PromptID,
		Status:           domain.PromptABTestStatusDraft,
		Variants:         variants,
		TrafficSplit:     trafficSplit,
		TargetMetric:     input.TargetMetric,
		SecondaryMetrics: input.SecondaryMetrics,
		MinSampleSize:    minSampleSize,
		ConfidenceLevel:  confidenceLevel,
		AutoSelectWinner: input.AutoSelectWinner,
		GradualRollout:   input.GradualRollout,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.mu.Lock()
	s.tests[testID] = test
	s.results[testID] = []domain.PromptABRecordResultInput{}
	s.mu.Unlock()

	s.logger.Info("created A/B test", zap.String("testId", testID.String()), zap.String("name", input.Name))
	return test, nil
}

// GetTest returns a test by ID
func (s *ABTestingService) GetTest(ctx context.Context, testID uuid.UUID) (*domain.PromptABTest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}
	return test, nil
}

// ListTests lists all tests for a project
func (s *ABTestingService) ListTests(ctx context.Context, projectID uuid.UUID) ([]*domain.PromptABTest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tests []*domain.PromptABTest
	for _, t := range s.tests {
		if t.ProjectID == projectID {
			tests = append(tests, t)
		}
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt.After(tests[j].CreatedAt)
	})

	return tests, nil
}

// StartTest starts a test
func (s *ABTestingService) StartTest(ctx context.Context, testID uuid.UUID) (*domain.PromptABTest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	if test.Status != domain.PromptABTestStatusDraft && test.Status != domain.PromptABTestStatusPaused {
		return nil, apperrors.Validation("test can only be started from draft or paused status")
	}

	now := time.Now()
	test.Status = domain.PromptABTestStatusRunning
	test.StartedAt = &now
	test.UpdatedAt = now

	s.logger.Info("started A/B test", zap.String("testId", testID.String()))
	return test, nil
}

// PauseTest pauses a running test
func (s *ABTestingService) PauseTest(ctx context.Context, testID uuid.UUID) (*domain.PromptABTest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	if test.Status != domain.PromptABTestStatusRunning {
		return nil, apperrors.Validation("only running tests can be paused")
	}

	test.Status = domain.PromptABTestStatusPaused
	test.UpdatedAt = time.Now()

	s.logger.Info("paused A/B test", zap.String("testId", testID.String()))
	return test, nil
}

// StopTest completes a test
func (s *ABTestingService) StopTest(ctx context.Context, testID uuid.UUID) (*domain.PromptABTest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	if test.Status != domain.PromptABTestStatusRunning && test.Status != domain.PromptABTestStatusPaused {
		return nil, apperrors.Validation("only running or paused tests can be stopped")
	}

	now := time.Now()
	test.Status = domain.PromptABTestStatusCompleted
	test.EndedAt = &now
	test.UpdatedAt = now

	s.logger.Info("stopped A/B test", zap.String("testId", testID.String()))
	return test, nil
}

// AssignVariant assigns traffic to a variant using sticky assignment
func (s *ABTestingService) AssignVariant(ctx context.Context, testID uuid.UUID, assignmentKey string) (*domain.PromptVariantAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	if test.Status != domain.PromptABTestStatusRunning {
		return nil, apperrors.Validation("test must be running to assign variants")
	}

	stickyKey := fmt.Sprintf("%s:%s", testID.String(), assignmentKey)

	// Check for existing sticky assignment
	if variantID, exists := s.assigns[stickyKey]; exists {
		for _, v := range test.Variants {
			if v.ID == variantID {
				return &domain.PromptVariantAssignment{
					TestID:          testID,
					VariantID:       v.ID,
					VariantName:     v.Name,
					PromptVersionID: v.PromptVersionID,
					AssignmentKey:   assignmentKey,
				}, nil
			}
		}
	}

	// Deterministic assignment using hash of assignment key
	variant := s.selectVariant(test, assignmentKey)

	s.assigns[stickyKey] = variant.ID

	return &domain.PromptVariantAssignment{
		TestID:          testID,
		VariantID:       variant.ID,
		VariantName:     variant.Name,
		PromptVersionID: variant.PromptVersionID,
		AssignmentKey:   assignmentKey,
	}, nil
}

// selectVariant deterministically selects a variant based on assignment key and traffic percentages
func (s *ABTestingService) selectVariant(test *domain.PromptABTest, assignmentKey string) *domain.PromptABTestVariant {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", test.ID.String(), assignmentKey)))
	hashVal := float64(binary.BigEndian.Uint32(h[:4])) / float64(math.MaxUint32) * 100.0

	var cumulative float64
	for i := range test.Variants {
		cumulative += test.Variants[i].TrafficPercent
		if hashVal <= cumulative {
			return &test.Variants[i]
		}
	}
	return &test.Variants[len(test.Variants)-1]
}

// RecordResult records a measurement for a variant
func (s *ABTestingService) RecordResult(ctx context.Context, testID uuid.UUID, input *domain.PromptABRecordResultInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return apperrors.NotFound("A/B test")
	}

	if test.Status != domain.PromptABTestStatusRunning {
		return apperrors.Validation("test must be running to record results")
	}

	// Validate variant exists
	variantFound := false
	for i := range test.Variants {
		if test.Variants[i].ID == input.VariantID {
			test.Variants[i].SampleCount++
			variantFound = true
			break
		}
	}
	if !variantFound {
		return apperrors.Validation("variant not found in test")
	}

	s.results[testID] = append(s.results[testID], *input)

	// Update variant metrics
	s.recalculateVariantMetrics(test)

	return nil
}

// recalculateVariantMetrics updates aggregated metrics for all variants
func (s *ABTestingService) recalculateVariantMetrics(test *domain.PromptABTest) {
	results := s.results[test.ID]

	for i := range test.Variants {
		variantID := test.Variants[i].ID
		var scores []float64
		var totalLatency, totalCost float64
		var totalTokens int64
		var errorCount int

		for _, r := range results {
			if r.VariantID == variantID {
				scores = append(scores, r.Score)
				totalLatency += r.LatencyMs
				totalCost += r.CostUSD
				totalTokens += int64(r.Tokens)
				if r.IsError {
					errorCount++
				}
			}
		}

		n := len(scores)
		if n == 0 {
			continue
		}

		// Calculate mean
		var sum float64
		for _, s := range scores {
			sum += s
		}
		mean := sum / float64(n)

		// Calculate standard deviation
		var varianceSum float64
		for _, s := range scores {
			varianceSum += (s - mean) * (s - mean)
		}
		stdDev := math.Sqrt(varianceSum / float64(n))

		// Calculate P95 latency
		latencies := make([]float64, 0)
		for _, r := range results {
			if r.VariantID == variantID && r.LatencyMs > 0 {
				latencies = append(latencies, r.LatencyMs)
			}
		}
		sort.Float64s(latencies)
		var p95Latency float64
		if len(latencies) > 0 {
			idx := int(math.Ceil(float64(len(latencies))*0.95)) - 1
			if idx >= len(latencies) {
				idx = len(latencies) - 1
			}
			p95Latency = latencies[idx]
		}

		test.Variants[i].Metrics = domain.ABTestVariantMetrics{
			AvgScore:     mean,
			StdDeviation: stdDev,
			AvgLatencyMs: totalLatency / float64(n),
			AvgCostUSD:   totalCost / float64(n),
			ErrorRate:    float64(errorCount) / float64(n),
			P95LatencyMs: p95Latency,
			TotalTokens:  totalTokens,
		}
	}
}

// GetStatistics computes statistical analysis for a test
func (s *ABTestingService) GetStatistics(ctx context.Context, testID uuid.UUID) (*domain.PromptABTestStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	results := s.results[testID]

	// Group results by variant
	variantResults := make(map[uuid.UUID][]float64)
	for _, r := range results {
		variantResults[r.VariantID] = append(variantResults[r.VariantID], r.Score)
	}

	// Find control variant
	var controlIdx int
	for i, v := range test.Variants {
		if v.IsControl {
			controlIdx = i
			break
		}
	}

	controlScores := variantResults[test.Variants[controlIdx].ID]
	controlMean, controlStdDev := meanStdDev(controlScores)

	totalSamples := len(results)
	zAlpha := zScoreForConfidence(test.ConfidenceLevel)

	// Compute per-variant statistics
	variantStats := make([]domain.PromptVariantStatistic, len(test.Variants))
	var bestVariantIdx int
	var bestMean float64
	var overallPValue float64 = 1.0

	for i, v := range test.Variants {
		scores := variantResults[v.ID]
		mean, stdDev := meanStdDev(scores)
		n := len(scores)

		// Confidence interval
		var marginOfError float64
		if n > 0 {
			marginOfError = zAlpha * stdDev / math.Sqrt(float64(n))
		}

		// Improvement over control
		var improvement float64
		if controlMean > 0 {
			improvement = ((mean - controlMean) / controlMean) * 100.0
		}

		// Z-test for difference of means (comparing each variant to control)
		pValue := 1.0
		if !v.IsControl && n > 1 && len(controlScores) > 1 {
			pValue = zTestTwoSample(mean, controlMean, stdDev, controlStdDev, n, len(controlScores))
		}

		if pValue < overallPValue && !v.IsControl {
			overallPValue = pValue
		}

		isWinner := false
		if mean > bestMean {
			bestMean = mean
			bestVariantIdx = i
		}

		variantStats[i] = domain.PromptVariantStatistic{
			VariantID:       v.ID,
			VariantName:     v.Name,
			Mean:            mean,
			StdDev:          stdDev,
			ConfidenceLower: mean - marginOfError,
			ConfidenceUpper: mean + marginOfError,
			SampleSize:      n,
			IsWinner:        isWinner,
			Improvement:     improvement,
		}
	}

	// Mark winner
	variantStats[bestVariantIdx].IsWinner = true

	isSignificant := overallPValue < (1.0 - test.ConfidenceLevel)

	// Effect size (Cohen's d)
	var effect float64
	if controlStdDev > 0 {
		effect = (bestMean - controlMean) / controlStdDev
	}

	// Power analysis (approximate)
	powerAnalysis := computePower(effect, totalSamples, zAlpha)

	// Required samples estimation
	requiredSamples := estimateRequiredSamples(effect, zAlpha, 0.80)

	// Recommendation
	recommendation := "continue"
	if totalSamples < test.MinSampleSize {
		recommendation = "insufficient_data"
	} else if isSignificant {
		recommendation = "select_winner"
	}

	stats := &domain.PromptABTestStatistics{
		TestID:          testID,
		IsSignificant:   isSignificant,
		PValue:          overallPValue,
		ConfidenceLevel: test.ConfidenceLevel,
		Effect:          effect,
		PowerAnalysis:   powerAnalysis,
		RequiredSamples: requiredSamples,
		CurrentSamples:  totalSamples,
		VariantStats:    variantStats,
		Recommendation:  recommendation,
		AnalyzedAt:      time.Now(),
	}

	return stats, nil
}

// SelectWinner selects a winning variant for the test
func (s *ABTestingService) SelectWinner(ctx context.Context, testID uuid.UUID, variantID uuid.UUID) (*domain.PromptABTest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	// Validate variant exists
	found := false
	for _, v := range test.Variants {
		if v.ID == variantID {
			found = true
			break
		}
	}
	if !found {
		return nil, apperrors.Validation("variant not found in test")
	}

	now := time.Now()
	test.WinnerID = &variantID
	test.Status = domain.PromptABTestStatusCompleted
	test.EndedAt = &now
	test.UpdatedAt = now

	s.logger.Info("selected winner for A/B test",
		zap.String("testId", testID.String()),
		zap.String("variantId", variantID.String()))

	return test, nil
}

// StartGradualRollout begins gradual rollout for the winning variant
func (s *ABTestingService) StartGradualRollout(ctx context.Context, testID uuid.UUID) (*domain.PromptABTest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	test, ok := s.tests[testID]
	if !ok {
		return nil, apperrors.NotFound("A/B test")
	}

	if test.WinnerID == nil {
		return nil, apperrors.Validation("a winner must be selected before starting gradual rollout")
	}

	if test.GradualRollout == nil {
		test.GradualRollout = &domain.PromptGradualRollout{
			Enabled:            true,
			InitialPercent:     10,
			IncrementPercent:   10,
			IncrementIntervalH: 1,
			CurrentPercent:     10,
			AutoComplete:       true,
		}
	} else {
		test.GradualRollout.Enabled = true
		test.GradualRollout.CurrentPercent = test.GradualRollout.InitialPercent
	}

	test.UpdatedAt = time.Now()

	s.logger.Info("started gradual rollout for A/B test", zap.String("testId", testID.String()))
	return test, nil
}

// --- Statistical helper functions ---

// meanStdDev computes mean and standard deviation
func meanStdDev(values []float64) (float64, float64) {
	n := len(values)
	if n == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)

	if n == 1 {
		return mean, 0
	}

	var varianceSum float64
	for _, v := range values {
		varianceSum += (v - mean) * (v - mean)
	}
	stdDev := math.Sqrt(varianceSum / float64(n-1))

	return mean, stdDev
}

// zTestTwoSample performs a two-sample z-test and returns the p-value
func zTestTwoSample(mean1, mean2, std1, std2 float64, n1, n2 int) float64 {
	if n1 == 0 || n2 == 0 {
		return 1.0
	}

	se := math.Sqrt((std1*std1)/float64(n1) + (std2*std2)/float64(n2))
	if se == 0 {
		return 1.0
	}

	z := math.Abs(mean1-mean2) / se

	// Two-tailed p-value using normal CDF approximation
	p := 2.0 * (1.0 - normalCDF(z))
	return p
}

// normalCDF approximates the cumulative distribution function of the standard normal
// using the Abramowitz and Stegun approximation (formula 26.2.17)
func normalCDF(x float64) float64 {
	if x < 0 {
		return 1.0 - normalCDF(-x)
	}

	const (
		a1 = 0.254829592
		a2 = -0.284496736
		a3 = 1.421413741
		a4 = -1.453152027
		a5 = 1.061405429
		p  = 0.3275911
	)

	t := 1.0 / (1.0 + p*x)
	t2 := t * t
	t3 := t2 * t
	t4 := t3 * t
	t5 := t4 * t

	return 1.0 - (a1*t+a2*t2+a3*t3+a4*t4+a5*t5)*math.Exp(-x*x/2.0)
}

// zScoreForConfidence returns the z-score for a given confidence level
func zScoreForConfidence(confidence float64) float64 {
	// Common confidence levels
	switch {
	case confidence >= 0.99:
		return 2.576
	case confidence >= 0.975:
		return 2.241
	case confidence >= 0.95:
		return 1.960
	case confidence >= 0.90:
		return 1.645
	default:
		return 1.960
	}
}

// computePower approximates statistical power
func computePower(effectSize float64, n int, zAlpha float64) float64 {
	if effectSize == 0 || n == 0 {
		return 0
	}

	// Power = P(Z > zAlpha - effectSize * sqrt(n/2))
	noncentrality := effectSize * math.Sqrt(float64(n)/2.0)
	zBeta := zAlpha - noncentrality
	power := normalCDF(-zBeta)

	if power > 1.0 {
		power = 1.0
	}
	if power < 0 {
		power = 0
	}
	return power
}

// estimateRequiredSamples estimates the total required sample size for desired power
func estimateRequiredSamples(effectSize, zAlpha, desiredPower float64) int {
	if effectSize == 0 {
		return 0
	}

	zBeta := zScoreForPower(desiredPower)
	// n = 2 * ((zAlpha + zBeta) / effectSize)^2
	n := 2.0 * math.Pow((zAlpha+zBeta)/effectSize, 2)

	return int(math.Ceil(n))
}

// zScoreForPower returns the z-score for a desired power level
func zScoreForPower(power float64) float64 {
	switch {
	case power >= 0.95:
		return 1.645
	case power >= 0.90:
		return 1.282
	case power >= 0.80:
		return 0.842
	default:
		return 0.842
	}
}
