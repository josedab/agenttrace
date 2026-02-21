package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CostRecommendationRepository defines repository operations for cost recommendations
type CostRecommendationRepository interface {
	Save(ctx context.Context, rec *domain.CostRecommendation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CostRecommendation, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.CostRecommendation, error)
	Update(ctx context.Context, rec *domain.CostRecommendation) error
}

// CostOptimizerService analyzes trace costs and recommends model substitutions
type CostOptimizerService struct {
	logger       *zap.Logger
	recRepo      CostRecommendationRepository
	costService  *CostService
	queryService *QueryService
}

// NewCostOptimizerService creates a new cost optimizer service
func NewCostOptimizerService(
	logger *zap.Logger,
	recRepo CostRecommendationRepository,
	costService *CostService,
	queryService *QueryService,
) *CostOptimizerService {
	return &CostOptimizerService{
		logger:       logger,
		recRepo:      recRepo,
		costService:  costService,
		queryService: queryService,
	}
}

// Analyze performs a cost analysis for a project over a date range.
// It groups traces by model, calculates per-model costs, and identifies
// expensive models with cheaper alternatives that maintain quality.
func (s *CostOptimizerService) Analyze(ctx context.Context, projectID uuid.UUID, dateRange domain.DateRange) (*domain.CostAnalysis, error) {
	// Fetch traces for the period
	filter := &domain.TraceFilter{
		ProjectID: projectID,
		FromTime:  &dateRange.Start,
		ToTime:    &dateRange.End,
	}

	traceList, err := s.queryService.ListTraces(ctx, filter, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list traces for analysis: %w", err)
	}

	// Group traces by model and calculate costs
	modelStats := make(map[string]*domain.ModelCostEntry)
	var totalCost float64

	for _, trace := range traceList.Traces {
		for _, obs := range trace.Observations {
			if obs.Model == "" {
				continue
			}
			entry, ok := modelStats[obs.Model]
			if !ok {
				entry = &domain.ModelCostEntry{Model: obs.Model}
				modelStats[obs.Model] = entry
			}
			entry.TraceCount++
			entry.TotalCost += obs.CostDetails.TotalCost
			totalCost += obs.CostDetails.TotalCost
		}
	}

	// Calculate averages and build breakdown
	breakdown := make([]domain.ModelCostEntry, 0, len(modelStats))
	for _, entry := range modelStats {
		if entry.TraceCount > 0 {
			entry.AvgCostPerTrace = entry.TotalCost / float64(entry.TraceCount)
		}
		breakdown = append(breakdown, *entry)
	}

	// Generate recommendations by finding cheaper alternatives
	recommendations, potentialSavings := s.generateRecommendations(projectID, breakdown)

	// Persist new recommendations
	for i := range recommendations {
		if err := s.recRepo.Save(ctx, &recommendations[i]); err != nil {
			s.logger.Warn("failed to save recommendation", zap.Error(err))
		}
	}

	analysis := &domain.CostAnalysis{
		ProjectID:        projectID,
		TotalCostPeriod:  totalCost,
		ModelBreakdown:   breakdown,
		Recommendations:  recommendations,
		PotentialSavings: potentialSavings,
	}

	s.logger.Info("completed cost analysis",
		zap.String("projectId", projectID.String()),
		zap.Float64("totalCost", totalCost),
		zap.Float64("potentialSavings", potentialSavings),
	)

	return analysis, nil
}

// generateRecommendations identifies expensive models and suggests cheaper alternatives
func (s *CostOptimizerService) generateRecommendations(projectID uuid.UUID, breakdown []domain.ModelCostEntry) ([]domain.CostRecommendation, float64) {
	// Map of known cheaper alternatives maintaining reasonable quality
	alternatives := map[string]string{
		"gpt-4":          "gpt-4o-mini",
		"gpt-4-turbo":    "gpt-4o-mini",
		"gpt-4o":         "gpt-4o-mini",
		"claude-3-opus":  "claude-3-5-sonnet",
		"claude-3-sonnet": "claude-3-5-haiku",
	}

	var recommendations []domain.CostRecommendation
	var potentialSavings float64

	for _, entry := range breakdown {
		alt, ok := alternatives[entry.Model]
		if !ok {
			continue
		}

		// Estimate savings based on known pricing ratios
		altPricing := s.costService.GetPricing(alt)
		currentPricing := s.costService.GetPricing(entry.Model)
		if altPricing == nil || currentPricing == nil || currentPricing.InputPricePer1K == 0 {
			continue
		}

		ratio := altPricing.InputPricePer1K / currentPricing.InputPricePer1K
		estimatedSavings := entry.TotalCost * (1 - ratio)
		potentialSavings += estimatedSavings

		rec := domain.CostRecommendation{
			ID:                       uuid.New(),
			ProjectID:                projectID,
			CurrentModel:             entry.Model,
			RecommendedModel:         alt,
			TraceCount:               entry.TraceCount,
			EstimatedSavingsPerMonth: estimatedSavings,
			QualityImpactEstimate:    0.05, // Conservative 5% quality impact estimate
			Confidence:               0.8,
			Status:                   domain.CostRecommendationPending,
			CreatedAt:                time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, potentialSavings
}

// GetRecommendations retrieves all cost recommendations for a project
func (s *CostOptimizerService) GetRecommendations(ctx context.Context, projectID uuid.UUID) ([]domain.CostRecommendation, error) {
	recs, err := s.recRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list recommendations: %w", err)
	}
	return recs, nil
}

// ApplyRecommendation marks a recommendation as applied
func (s *CostOptimizerService) ApplyRecommendation(ctx context.Context, recommendationID uuid.UUID) error {
	rec, err := s.recRepo.GetByID(ctx, recommendationID)
	if err != nil {
		return fmt.Errorf("failed to get recommendation: %w", err)
	}

	rec.Status = domain.CostRecommendationApplied
	if err := s.recRepo.Update(ctx, rec); err != nil {
		return fmt.Errorf("failed to update recommendation: %w", err)
	}

	s.logger.Info("applied cost recommendation",
		zap.String("recommendationId", recommendationID.String()),
		zap.String("from", rec.CurrentModel),
		zap.String("to", rec.RecommendedModel),
	)

	return nil
}

// DismissRecommendation marks a recommendation as dismissed
func (s *CostOptimizerService) DismissRecommendation(ctx context.Context, recommendationID uuid.UUID) error {
	rec, err := s.recRepo.GetByID(ctx, recommendationID)
	if err != nil {
		return fmt.Errorf("failed to get recommendation: %w", err)
	}

	rec.Status = domain.CostRecommendationDismissed
	if err := s.recRepo.Update(ctx, rec); err != nil {
		return fmt.Errorf("failed to update recommendation: %w", err)
	}

	s.logger.Info("dismissed cost recommendation",
		zap.String("recommendationId", recommendationID.String()),
	)

	return nil
}
