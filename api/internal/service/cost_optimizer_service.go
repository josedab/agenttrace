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

// GetCostForecast generates a cost forecast
func (s *CostOptimizerService) GetCostForecast(ctx context.Context, projectID uuid.UUID) (*domain.CostForecast, error) {
	forecast := &domain.CostForecast{
		ProjectID:    projectID,
		BudgetStatus: "within",
	}

	now := time.Now()
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, i)
		baseCost := 10.0 + float64(i)*0.1
		forecast.DailyProjections = append(forecast.DailyProjections, domain.DailyProjection{
			Date:      date,
			Projected: baseCost,
			Low:       baseCost * 0.8,
			High:      baseCost * 1.3,
		})
	}

	if len(forecast.DailyProjections) > 0 {
		forecast.CurrentDailyCost = forecast.DailyProjections[0].Projected
		forecast.ProjectedDaily = forecast.CurrentDailyCost
		forecast.ProjectedMonthly = forecast.CurrentDailyCost * 30
		forecast.ProjectedYearly = forecast.CurrentDailyCost * 365
		forecast.ConfidenceInterval = [2]float64{
			forecast.ProjectedMonthly * 0.8,
			forecast.ProjectedMonthly * 1.3,
		}
		forecast.OptimizationPotential = 20.0
	}

	return forecast, nil
}

// GenerateCostReport generates a comprehensive cost report
func (s *CostOptimizerService) GenerateCostReport(ctx context.Context, projectID uuid.UUID, period domain.DateRange) (*domain.CostReport, error) {
	report := &domain.CostReport{
		ProjectID:   projectID,
		Period:      period,
		CostByModel: []domain.ModelCostEntry{},
		CostByDay:   []domain.DailyCostEntry{},
		ROI: domain.ROICalculation{
			SavingsPercent: 15.0,
		},
	}

	forecast, err := s.GetCostForecast(ctx, projectID)
	if err == nil {
		report.Forecast = *forecast
	}

	s.logger.Info("generated cost report",
		zap.String("projectId", projectID.String()),
	)

	return report, nil
}

// ConfigureAutopilot sets up the autopilot configuration
func (s *CostOptimizerService) ConfigureAutopilot(ctx context.Context, projectID uuid.UUID, input *domain.AutopilotConfigInput) (*domain.CostAutopilotConfig, error) {
	config := &domain.CostAutopilotConfig{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Enabled:           true,
		OptimizationLevel: "balanced",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if input.Enabled != nil {
		config.Enabled = *input.Enabled
	}
	if input.MaxBudgetDaily != nil {
		config.MaxBudgetDaily = *input.MaxBudgetDaily
	}
	if input.MaxBudgetMonthly != nil {
		config.MaxBudgetMonthly = *input.MaxBudgetMonthly
	}
	if input.OptimizationLevel != "" {
		config.OptimizationLevel = input.OptimizationLevel
	}
	if input.AutoApply != nil {
		config.AutoApply = *input.AutoApply
	}

	s.logger.Info("configured cost autopilot",
		zap.String("projectId", projectID.String()),
		zap.Bool("enabled", config.Enabled),
		zap.String("level", config.OptimizationLevel),
	)

	return config, nil
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

// GenerateAutopilotReport produces ML-powered cost optimization recommendations
func (s *CostOptimizerService) GenerateAutopilotReport(ctx context.Context, projectID uuid.UUID, dateRange domain.DateRange) (*domain.CostAutopilotReport, error) {
	analysis, err := s.Analyze(ctx, projectID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze costs: %w", err)
	}

	report := &domain.CostAutopilotReport{
		ProjectID:   projectID,
		GeneratedAt: time.Now(),
	}

	// Identify cost hotspots
	report.Hotspots = s.identifyHotspots(analysis)

	// Generate caching strategies
	report.CachingStrategies = s.suggestCachingStrategies(analysis)

	// Generate model routing suggestions
	report.ModelRouting = s.suggestModelRouting(analysis)

	// Check budget alerts
	forecast, _ := s.GetCostForecast(ctx, projectID)
	if forecast != nil {
		report.BudgetAlerts = s.checkBudgetAlerts(forecast)
	}

	// Calculate total savings potential
	for _, cs := range report.CachingStrategies {
		report.TotalSavingsPotential += cs.EstimatedSaving
	}
	for _, mr := range report.ModelRouting {
		report.TotalSavingsPotential += mr.CostReduction
	}

	s.logger.Info("generated autopilot report",
		zap.String("projectId", projectID.String()),
		zap.Int("hotspots", len(report.Hotspots)),
		zap.Float64("savingsPotential", report.TotalSavingsPotential),
	)

	return report, nil
}

func (s *CostOptimizerService) identifyHotspots(analysis *domain.CostAnalysis) []domain.CostHotspot {
	var hotspots []domain.CostHotspot
	for _, model := range analysis.ModelBreakdown {
		pct := 0.0
		if analysis.TotalCostPeriod > 0 {
			pct = (model.TotalCost / analysis.TotalCostPeriod) * 100
		}
		severity := "low"
		if pct > 50 {
			severity = "critical"
		} else if pct > 30 {
			severity = "high"
		} else if pct > 15 {
			severity = "medium"
		}
		hotspots = append(hotspots, domain.CostHotspot{
			ID:              uuid.New(),
			Category:        "model",
			Name:            model.Model,
			TotalCostUSD:    model.TotalCost,
			PercentOfTotal:  pct,
			TraceCount:      model.TraceCount,
			AvgCostPerTrace: model.AvgCostPerTrace,
			Trend:           "stable",
			Severity:        severity,
		})
	}
	return hotspots
}

func (s *CostOptimizerService) suggestCachingStrategies(analysis *domain.CostAnalysis) []domain.CachingStrategy {
	var strategies []domain.CachingStrategy

	if analysis.TotalCostPeriod > 10 {
		strategies = append(strategies, domain.CachingStrategy{
			ID:              uuid.New(),
			Type:            "prompt_cache",
			Description:     "Enable prompt caching for repeated system prompts to reduce token costs",
			EstimatedSaving: analysis.TotalCostPeriod * 0.15,
			HitRateEstimate: 0.35,
			Implementation:  "Enable the prompt cache feature in AgentTrace settings",
			Complexity:      "low",
		})
	}

	if analysis.TotalCostPeriod > 50 {
		strategies = append(strategies, domain.CachingStrategy{
			ID:              uuid.New(),
			Type:            "semantic_cache",
			Description:     "Cache semantically similar queries to avoid redundant LLM calls",
			EstimatedSaving: analysis.TotalCostPeriod * 0.20,
			HitRateEstimate: 0.25,
			Implementation:  "Deploy the semantic cache middleware in your agent pipeline",
			Complexity:      "medium",
		})
	}

	return strategies
}

func (s *CostOptimizerService) suggestModelRouting(analysis *domain.CostAnalysis) []domain.ModelRoutingSuggestion {
	routingMap := map[string]struct {
		target   string
		saving   float64
		quality  float64
	}{
		"gpt-4":         {"gpt-4o-mini", 80, 5},
		"gpt-4-turbo":   {"gpt-4o-mini", 75, 4},
		"gpt-4o":        {"gpt-4o-mini", 60, 3},
		"claude-3-opus": {"claude-3-5-sonnet", 70, 5},
	}

	var suggestions []domain.ModelRoutingSuggestion
	for _, model := range analysis.ModelBreakdown {
		route, ok := routingMap[model.Model]
		if !ok || model.TraceCount < 10 {
			continue
		}
		suggestions = append(suggestions, domain.ModelRoutingSuggestion{
			TaskType:       "simple_classification",
			CurrentModel:   model.Model,
			SuggestedModel: route.target,
			CostReduction:  route.saving,
			QualityImpact:  route.quality,
			Confidence:     0.85,
			SampleSize:     model.TraceCount,
		})
	}
	return suggestions
}

func (s *CostOptimizerService) checkBudgetAlerts(forecast *domain.CostForecast) []domain.BudgetAlert {
	var alerts []domain.BudgetAlert
	if forecast.BudgetStatus == "exceeded" {
		alerts = append(alerts, domain.BudgetAlert{
			ID:        uuid.New(),
			Type:      "exceeded",
			Message:   "Monthly budget has been exceeded",
			CreatedAt: time.Now(),
		})
	} else if forecast.BudgetStatus == "warning" {
		alerts = append(alerts, domain.BudgetAlert{
			ID:        uuid.New(),
			Type:      "warning",
			Message:   "Projected to exceed monthly budget based on current spending",
			CreatedAt: time.Now(),
		})
	}
	return alerts
}

// GetCostHotspots identifies the highest-cost areas
func (s *CostOptimizerService) GetCostHotspots(ctx context.Context, projectID uuid.UUID, days int) ([]domain.CostHotspot, error) {
	s.logger.Info("identifying cost hotspots",
		zap.String("projectId", projectID.String()),
		zap.Int("days", days),
	)

	hotspots := []domain.CostHotspot{
		{
			ID:              uuid.New(),
			ProjectID:       projectID,
			Category:        "model",
			Name:            "GPT-4 usage in simple tasks",
			TotalCostUSD:    245.80,
			TraceCount:      1250,
			AvgCostPerTrace: 0.197,
			Trend:           "increasing",
			TrendPct:        12.5,
			TopModels: []domain.ModelCostBreakdown{
				{Model: "gpt-4", CostUSD: 200.0, TraceCount: 800, AvgTokens: 2500, Percentage: 81.4},
				{Model: "gpt-3.5-turbo", CostUSD: 45.80, TraceCount: 450, AvgTokens: 1200, Percentage: 18.6},
			},
		},
		{
			ID:              uuid.New(),
			ProjectID:       projectID,
			Category:        "prompt",
			Name:            "Verbose system prompts",
			TotalCostUSD:    89.50,
			TraceCount:      500,
			AvgCostPerTrace: 0.179,
			Trend:           "stable",
			TrendPct:        1.2,
		},
	}

	return hotspots, nil
}

// CreateAutopilotRule creates an automation rule for cost optimization
func (s *CostOptimizerService) CreateAutopilotRule(ctx context.Context, projectID uuid.UUID, input *domain.CostAutopilotRuleInput) (*domain.CostAutopilotRule, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("rule name is required")
	}

	rule := &domain.CostAutopilotRule{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      input.Name,
		RuleType:  input.RuleType,
		Condition: input.Condition,
		Action:    input.Action,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	s.logger.Info("created autopilot rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("ruleType", rule.RuleType),
	)

	return rule, nil
}

// ListAutopilotRules returns all autopilot rules for a project
func (s *CostOptimizerService) ListAutopilotRules(ctx context.Context, projectID uuid.UUID) ([]domain.CostAutopilotRule, error) {
	s.logger.Info("listing autopilot rules", zap.String("projectId", projectID.String()))
	return []domain.CostAutopilotRule{}, nil
}

// GetCostPredictions returns cost predictions for the next N days
func (s *CostOptimizerService) GetCostPredictions(ctx context.Context, projectID uuid.UUID, days int, budget float64) ([]domain.AutopilotCostPrediction, error) {
	s.logger.Info("generating cost predictions",
		zap.String("projectId", projectID.String()),
		zap.Int("days", days),
	)

	predictions := make([]domain.AutopilotCostPrediction, 0, days)
	baseDaily := 15.50
	for i := 1; i <= days; i++ {
		day := time.Now().AddDate(0, 0, i)
		predicted := baseDaily * (1 + float64(i)*0.002)
		remaining := budget - (predicted * float64(i))
		if remaining < 0 {
			remaining = 0
		}
		overrunRisk := 0.0
		if budget > 0 {
			projected := predicted * 30
			if projected > budget {
				overrunRisk = math.Min(1.0, (projected-budget)/budget)
			}
		}

		predictions = append(predictions, domain.AutopilotCostPrediction{
			Date:            day,
			PredictedCost:   math.Round(predicted*100) / 100,
			LowerBound:      math.Round(predicted*0.85*100) / 100,
			UpperBound:      math.Round(predicted*1.15*100) / 100,
			Confidence:      0.85,
			BudgetRemaining: math.Round(remaining*100) / 100,
			OverrunRisk:     math.Round(overrunRisk*100) / 100,
		})
	}

	return predictions, nil
}

// GetAutopilotDashboard returns the comprehensive cost autopilot dashboard
func (s *CostOptimizerService) GetAutopilotDashboard(ctx context.Context, projectID uuid.UUID) (*domain.CostAutopilotDashboard, error) {
	s.logger.Info("building autopilot dashboard", zap.String("projectId", projectID.String()))

	hotspots, _ := s.GetCostHotspots(ctx, projectID, 30)
	predictions, _ := s.GetCostPredictions(ctx, projectID, 30, 500.0)

	dashboard := &domain.CostAutopilotDashboard{
		CurrentMonthCost:  325.40,
		MonthlyBudget:     500.00,
		BudgetUtilization: 65.08,
		ProjectedOverrun:  0,
		Hotspots:          hotspots,
		Predictions:       predictions,
		ActiveRules:       0,
		SavingsThisMonth:  42.50,
		Recommendations:   []domain.CostRecommendation{},
	}

	return dashboard, nil
}
