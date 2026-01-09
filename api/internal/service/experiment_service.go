package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ExperimentService handles experiment (A/B testing) logic
type ExperimentService struct {
	logger *zap.Logger
}

// NewExperimentService creates a new experiment service
func NewExperimentService(logger *zap.Logger) *ExperimentService {
	return &ExperimentService{
		logger: logger,
	}
}

// CreateExperiment creates a new experiment
func (s *ExperimentService) CreateExperiment(
	ctx context.Context,
	projectID uuid.UUID,
	userID uuid.UUID,
	input *domain.ExperimentInput,
) (*domain.Experiment, error) {
	// Validate variant weights sum to 100
	var totalWeight float64
	for _, v := range input.Variants {
		totalWeight += v.Weight
	}
	if math.Abs(totalWeight-100) > 0.01 {
		return nil, fmt.Errorf("variant weights must sum to 100, got %.2f", totalWeight)
	}

	// Ensure at least one control variant
	hasControl := false
	for _, v := range input.Variants {
		if v.IsControl {
			hasControl = true
			break
		}
	}
	if !hasControl {
		return nil, fmt.Errorf("at least one variant must be marked as control")
	}

	experiment := &domain.Experiment{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Name:           input.Name,
		Description:    input.Description,
		Status:         domain.ExperimentStatusDraft,
		TargetMetric:   input.TargetMetric,
		TargetGoal:     input.TargetGoal,
		TrafficPercent: input.TrafficPercent,
		TraceNameFilter: input.TraceNameFilter,
		UserIDFilter:    input.UserIDFilter,
		MetadataFilters: input.MetadataFilters,
		MinDuration:    input.MinDuration,
		MinSamples:     input.MinSamples,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      userID,
		Variants:       make([]domain.ExperimentVariant, len(input.Variants)),
	}

	// Create variants
	for i, v := range input.Variants {
		experiment.Variants[i] = domain.ExperimentVariant{
			ID:           uuid.New(),
			ExperimentID: experiment.ID,
			Name:         v.Name,
			Description:  v.Description,
			Weight:       v.Weight,
			IsControl:    v.IsControl,
			Config:       v.Config,
			SampleCount:  0,
		}
	}

	s.logger.Info("Created experiment",
		zap.String("experimentId", experiment.ID.String()),
		zap.String("name", experiment.Name),
		zap.Int("variants", len(experiment.Variants)),
	)

	return experiment, nil
}

// StartExperiment starts an experiment
func (s *ExperimentService) StartExperiment(
	ctx context.Context,
	experiment *domain.Experiment,
) error {
	if experiment.Status != domain.ExperimentStatusDraft && experiment.Status != domain.ExperimentStatusPaused {
		return fmt.Errorf("experiment must be in draft or paused status to start")
	}

	now := time.Now()
	experiment.Status = domain.ExperimentStatusRunning
	experiment.StartedAt = &now
	experiment.UpdatedAt = now

	s.logger.Info("Started experiment",
		zap.String("experimentId", experiment.ID.String()),
	)

	return nil
}

// PauseExperiment pauses a running experiment
func (s *ExperimentService) PauseExperiment(
	ctx context.Context,
	experiment *domain.Experiment,
) error {
	if experiment.Status != domain.ExperimentStatusRunning {
		return fmt.Errorf("experiment must be running to pause")
	}

	experiment.Status = domain.ExperimentStatusPaused
	experiment.UpdatedAt = time.Now()

	return nil
}

// CompleteExperiment completes an experiment and calculates final results
func (s *ExperimentService) CompleteExperiment(
	ctx context.Context,
	experiment *domain.Experiment,
	metrics []domain.ExperimentMetric,
) error {
	if experiment.Status != domain.ExperimentStatusRunning && experiment.Status != domain.ExperimentStatusPaused {
		return fmt.Errorf("experiment must be running or paused to complete")
	}

	now := time.Now()
	experiment.Status = domain.ExperimentStatusCompleted
	experiment.EndedAt = &now
	experiment.UpdatedAt = now

	// Calculate results
	results := s.AnalyzeResults(experiment, metrics)
	experiment.Results = results

	// Determine winner if statistically significant
	for _, comp := range results.Comparisons {
		if comp.IsSignificant && comp.Winner != nil {
			for _, v := range experiment.Variants {
				if (comp.VariantAName == *comp.Winner && v.ID == comp.VariantA) ||
					(comp.VariantBName == *comp.Winner && v.ID == comp.VariantB) {
					experiment.WinningVariant = &v.ID
					break
				}
			}
		}
	}

	s.logger.Info("Completed experiment",
		zap.String("experimentId", experiment.ID.String()),
		zap.Bool("hasWinner", experiment.WinningVariant != nil),
	)

	return nil
}

// AssignVariant assigns a trace to a variant using consistent hashing
func (s *ExperimentService) AssignVariant(
	experiment *domain.Experiment,
	traceID uuid.UUID,
	userID string,
) (*domain.ExperimentAssignment, error) {
	if experiment.Status != domain.ExperimentStatusRunning {
		return nil, fmt.Errorf("experiment is not running")
	}

	// Use consistent hashing based on user ID (or trace ID if no user)
	hashInput := userID
	if hashInput == "" {
		hashInput = traceID.String()
	}
	hashInput = experiment.ID.String() + ":" + hashInput

	hash := sha256.Sum256([]byte(hashInput))
	hashValue := binary.BigEndian.Uint64(hash[:8])

	// Check if trace should be included based on traffic percentage
	trafficHash := float64(hashValue%10000) / 100.0
	if trafficHash >= experiment.TrafficPercent {
		return nil, nil // Not included in experiment
	}

	// Select variant based on weights
	variantHash := float64(hashValue%10000) / 100.0
	var cumulativeWeight float64
	var selectedVariant *domain.ExperimentVariant

	for i := range experiment.Variants {
		cumulativeWeight += experiment.Variants[i].Weight
		if variantHash < cumulativeWeight {
			selectedVariant = &experiment.Variants[i]
			break
		}
	}

	if selectedVariant == nil {
		selectedVariant = &experiment.Variants[len(experiment.Variants)-1]
	}

	return &domain.ExperimentAssignment{
		ExperimentID:  experiment.ID,
		VariantID:     selectedVariant.ID,
		TraceID:       traceID,
		AssignedAt:    time.Now(),
		VariantConfig: selectedVariant.Config,
	}, nil
}

// AnalyzeResults performs statistical analysis on experiment metrics
func (s *ExperimentService) AnalyzeResults(
	experiment *domain.Experiment,
	metrics []domain.ExperimentMetric,
) *domain.ExperimentResults {
	results := &domain.ExperimentResults{
		AnalyzedAt:     time.Now(),
		TotalSamples:   len(metrics),
		VariantResults: make([]domain.VariantResult, len(experiment.Variants)),
		Comparisons:    make([]domain.VariantComparison, 0),
	}

	// Group metrics by variant
	variantMetrics := make(map[uuid.UUID][]float64)
	for _, m := range metrics {
		variantMetrics[m.VariantID] = append(variantMetrics[m.VariantID], m.MetricValue)
	}

	// Calculate per-variant statistics
	for i, variant := range experiment.Variants {
		values := variantMetrics[variant.ID]
		result := domain.VariantResult{
			VariantID:   variant.ID,
			VariantName: variant.Name,
			SampleCount: len(values),
		}

		if len(values) > 0 {
			result.Mean = s.mean(values)
			result.StdDev = s.stdDev(values, result.Mean)
			result.Median = s.percentile(values, 50)
			result.P95 = s.percentile(values, 95)
			result.P99 = s.percentile(values, 99)
			result.Min = s.min(values)
			result.Max = s.max(values)
		}

		results.VariantResults[i] = result
	}

	// Perform pairwise comparisons (each variant vs control)
	var controlVariant *domain.ExperimentVariant
	for i := range experiment.Variants {
		if experiment.Variants[i].IsControl {
			controlVariant = &experiment.Variants[i]
			break
		}
	}

	if controlVariant != nil {
		controlValues := variantMetrics[controlVariant.ID]

		for _, variant := range experiment.Variants {
			if variant.ID == controlVariant.ID {
				continue
			}

			treatmentValues := variantMetrics[variant.ID]
			comparison := s.compareVariants(
				controlVariant.ID, controlVariant.Name, controlValues,
				variant.ID, variant.Name, treatmentValues,
				experiment.TargetGoal,
			)
			results.Comparisons = append(results.Comparisons, comparison)
		}
	}

	// Determine recommended action
	results.RecommendedAction = s.determineRecommendation(experiment, results)
	results.Confidence = s.calculateOverallConfidence(results)

	return results
}

// compareVariants performs statistical comparison between two variants
func (s *ExperimentService) compareVariants(
	variantAID uuid.UUID, variantAName string, valuesA []float64,
	variantBID uuid.UUID, variantBName string, valuesB []float64,
	targetGoal string,
) domain.VariantComparison {
	comparison := domain.VariantComparison{
		VariantA:     variantAID,
		VariantB:     variantBID,
		VariantAName: variantAName,
		VariantBName: variantBName,
	}

	if len(valuesA) == 0 || len(valuesB) == 0 {
		return comparison
	}

	meanA := s.mean(valuesA)
	meanB := s.mean(valuesB)

	comparison.MeanDifference = meanB - meanA
	if meanA != 0 {
		comparison.PercentChange = ((meanB - meanA) / meanA) * 100
	}

	// Perform Welch's t-test
	comparison.PValue = s.welchTTest(valuesA, valuesB)
	comparison.IsSignificant = comparison.PValue < 0.05
	comparison.ConfidenceLevel = 1 - comparison.PValue

	// Determine winner based on goal
	if comparison.IsSignificant {
		if targetGoal == "minimize" {
			if meanB < meanA {
				winner := variantBName
				comparison.Winner = &winner
			} else {
				winner := variantAName
				comparison.Winner = &winner
			}
		} else { // maximize
			if meanB > meanA {
				winner := variantBName
				comparison.Winner = &winner
			} else {
				winner := variantAName
				comparison.Winner = &winner
			}
		}
	}

	return comparison
}

// Statistical helper functions

func (s *ExperimentService) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *ExperimentService) stdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

func (s *ExperimentService) percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := (p / 100) * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

func (s *ExperimentService) min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minVal := values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

func (s *ExperimentService) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

// welchTTest performs Welch's t-test and returns the p-value
func (s *ExperimentService) welchTTest(group1, group2 []float64) float64 {
	n1 := float64(len(group1))
	n2 := float64(len(group2))

	if n1 < 2 || n2 < 2 {
		return 1.0 // Not enough samples
	}

	mean1 := s.mean(group1)
	mean2 := s.mean(group2)
	var1 := s.variance(group1, mean1)
	var2 := s.variance(group2, mean2)

	// Welch's t-statistic
	se := math.Sqrt(var1/n1 + var2/n2)
	if se == 0 {
		return 1.0
	}
	t := (mean1 - mean2) / se

	// Welch-Satterthwaite degrees of freedom
	num := math.Pow(var1/n1+var2/n2, 2)
	denom := math.Pow(var1/n1, 2)/(n1-1) + math.Pow(var2/n2, 2)/(n2-1)
	df := num / denom

	// Approximate p-value using normal distribution for large samples
	// For a proper implementation, use a t-distribution
	return s.tDistributionPValue(math.Abs(t), df)
}

func (s *ExperimentService) variance(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values)-1)
}

// tDistributionPValue approximates p-value for t-distribution
// This is a simplified approximation - use a proper statistical library in production
func (s *ExperimentService) tDistributionPValue(t, df float64) float64 {
	// For large df, approximate with normal distribution
	if df > 100 {
		// Two-tailed p-value from standard normal
		return 2 * (1 - s.normalCDF(t))
	}

	// Simple approximation for smaller df
	// In production, use gonum or similar statistical library
	x := df / (df + t*t)
	pValue := s.incompleteBeta(df/2, 0.5, x)
	return pValue
}

// normalCDF approximates the standard normal CDF
func (s *ExperimentService) normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// incompleteBeta is a simplified approximation of the incomplete beta function
func (s *ExperimentService) incompleteBeta(a, b, x float64) float64 {
	// Simplified approximation - use gonum for production
	if x < 0 || x > 1 {
		return 0
	}
	return math.Pow(x, a) * math.Pow(1-x, b) / (a * s.beta(a, b))
}

func (s *ExperimentService) beta(a, b float64) float64 {
	lg1, _ := math.Lgamma(a)
	lg2, _ := math.Lgamma(b)
	lg3, _ := math.Lgamma(a + b)
	return math.Exp(lg1 + lg2 - lg3)
}

func (s *ExperimentService) determineRecommendation(
	experiment *domain.Experiment,
	results *domain.ExperimentResults,
) string {
	// Check if we have enough samples
	minSamples := 100
	if experiment.MinSamples != nil {
		minSamples = *experiment.MinSamples
	}

	for _, vr := range results.VariantResults {
		if vr.SampleCount < minSamples {
			return "Continue collecting data - insufficient samples"
		}
	}

	// Check for significant winner
	for _, comp := range results.Comparisons {
		if comp.IsSignificant && comp.Winner != nil {
			return fmt.Sprintf("Roll out %s - statistically significant improvement (p=%.4f)",
				*comp.Winner, comp.PValue)
		}
	}

	// No significant difference
	if results.Confidence < 0.8 {
		return "Continue collecting data - confidence level too low"
	}

	return "No significant difference detected - consider keeping current implementation"
}

func (s *ExperimentService) calculateOverallConfidence(results *domain.ExperimentResults) float64 {
	if len(results.Comparisons) == 0 {
		return 0
	}

	var totalConfidence float64
	for _, comp := range results.Comparisons {
		totalConfidence += comp.ConfidenceLevel
	}

	return totalConfidence / float64(len(results.Comparisons))
}
