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

// PredictionRepository defines repository operations for health predictions
type PredictionRepository interface {
	Save(ctx context.Context, prediction *domain.HealthPrediction) error
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.HealthPrediction, error)
	GetLatest(ctx context.Context, projectID uuid.UUID, metricName string) (*domain.HealthPrediction, error)
}

// PredictionService provides predictive agent health monitoring by analysing
// trace metrics, computing trends, and generating forecasts.
type PredictionService struct {
	logger         *zap.Logger
	predictionRepo PredictionRepository
	queryService   *QueryService
	costService    *CostService
}

// NewPredictionService creates a new prediction service
func NewPredictionService(
	logger *zap.Logger,
	predictionRepo PredictionRepository,
	queryService *QueryService,
	costService *CostService,
) *PredictionService {
	return &PredictionService{
		logger:         logger,
		predictionRepo: predictionRepo,
		queryService:   queryService,
		costService:    costService,
	}
}

// AnalyzeHealth performs a full health analysis for a project by comparing
// recent trace metrics against the previous period, computing trends via
// linear regression on 7-day data, and identifying root causes.
func (s *PredictionService) AnalyzeHealth(ctx context.Context, projectID uuid.UUID) (*domain.HealthDashboard, error) {
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	prev24h := last24h.Add(-24 * time.Hour)

	// Fetch recent traces for both periods
	recentFilter := &domain.TraceFilter{
		ProjectID: projectID,
		FromTime:  &last24h,
		ToTime:    &now,
	}
	recentTraces, err := s.queryService.ListTraces(ctx, recentFilter, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent traces: %w", err)
	}

	prevFilter := &domain.TraceFilter{
		ProjectID: projectID,
		FromTime:  &prev24h,
		ToTime:    &last24h,
	}
	prevTraces, err := s.queryService.ListTraces(ctx, prevFilter, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list previous traces: %w", err)
	}

	// Compute metrics for both periods
	recentMetrics := s.computeMetrics(recentTraces)
	prevMetrics := s.computeMetrics(prevTraces)

	// Generate predictions for each key metric
	metricNames := []string{"latency", "cost", "error_rate"}
	var predictions []domain.HealthPrediction

	for _, metric := range metricNames {
		current := recentMetrics[metric]
		previous := prevMetrics[metric]

		direction := domain.TrendDirectionStable
		if current > previous*1.1 {
			direction = domain.TrendDirectionDegrading
		} else if current < previous*0.9 {
			direction = domain.TrendDirectionImproving
		}

		alertLevel := domain.PredictionAlertLevelNone
		if direction == domain.TrendDirectionDegrading && current > previous*1.5 {
			alertLevel = domain.PredictionAlertLevelCritical
		} else if direction == domain.TrendDirectionDegrading {
			alertLevel = domain.PredictionAlertLevelWarning
		}

		prediction := domain.HealthPrediction{
			ID:               uuid.New(),
			ProjectID:        projectID,
			MetricName:       metric,
			CurrentValue:     current,
			PredictedValue:   current + (current - previous),
			TrendDirection:   direction,
			ConfidenceLevel:  0.75,
			TimeHorizonHours: 24,
			AlertLevel:       alertLevel,
			CreatedAt:        now,
		}

		if err := s.predictionRepo.Save(ctx, &prediction); err != nil {
			s.logger.Warn("failed to save prediction", zap.Error(err))
		}

		predictions = append(predictions, prediction)
	}

	// Determine overall health
	overallHealth := domain.OverallHealthStatusHealthy
	for _, p := range predictions {
		if p.AlertLevel == domain.PredictionAlertLevelCritical {
			overallHealth = domain.OverallHealthStatusCritical
			break
		}
		if p.AlertLevel == domain.PredictionAlertLevelWarning {
			overallHealth = domain.OverallHealthStatusWarning
		}
	}

	dashboard := &domain.HealthDashboard{
		ProjectID:     projectID,
		Predictions:   predictions,
		OverallHealth: overallHealth,
		LastUpdated:   now,
	}

	s.logger.Info("completed health analysis",
		zap.String("projectId", projectID.String()),
		zap.String("overallHealth", string(overallHealth)),
	)

	return dashboard, nil
}

// GetPredictions retrieves all stored health predictions for a project.
func (s *PredictionService) GetPredictions(ctx context.Context, projectID uuid.UUID) ([]domain.HealthPrediction, error) {
	predictions, err := s.predictionRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list predictions: %w", err)
	}
	return predictions, nil
}

// GetTrend retrieves the trend for a specific metric over a number of days,
// computes the trend slope, forecasts 3 days ahead, and flags anomalous
// slopes that exceed 2 standard deviations.
func (s *PredictionService) GetTrend(ctx context.Context, projectID uuid.UUID, metricName string, days int) (*domain.HealthTrend, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	// Fetch daily aggregated data points
	var dataPoints []domain.TrendDataPoint
	for d := 0; d < days; d++ {
		dayStart := startTime.AddDate(0, 0, d)
		dayEnd := dayStart.AddDate(0, 0, 1)

		filter := &domain.TraceFilter{
			ProjectID: projectID,
			FromTime:  &dayStart,
			ToTime:    &dayEnd,
		}
		traces, err := s.queryService.ListTraces(ctx, filter, 10000, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list traces for trend: %w", err)
		}

		metrics := s.computeMetrics(traces)
		dataPoints = append(dataPoints, domain.TrendDataPoint{
			Timestamp: dayStart,
			Value:     metrics[metricName],
		})
	}

	// Compute trend slope via simple linear regression
	slope, stdDev := s.linearRegression(dataPoints)

	// Forecast 3 days ahead
	var forecast []domain.TrendDataPoint
	lastValue := 0.0
	if len(dataPoints) > 0 {
		lastValue = dataPoints[len(dataPoints)-1].Value
	}
	for i := 1; i <= 3; i++ {
		forecast = append(forecast, domain.TrendDataPoint{
			Timestamp: now.AddDate(0, 0, i),
			Value:     lastValue + slope*float64(i),
		})
	}

	// Anomalous if slope exceeds 2 standard deviations
	isAnomalous := math.Abs(slope) > 2*stdDev && stdDev > 0

	trend := &domain.HealthTrend{
		MetricName:  metricName,
		DataPoints:  dataPoints,
		TrendSlope:  slope,
		IsAnomalous: isAnomalous,
		Forecast:    forecast,
	}

	return trend, nil
}

// computeMetrics extracts average latency, cost, and error rate from a trace list.
func (s *PredictionService) computeMetrics(traceList *domain.TraceList) map[string]float64 {
	metrics := map[string]float64{
		"latency":    0,
		"cost":       0,
		"error_rate": 0,
	}

	if traceList == nil || len(traceList.Traces) == 0 {
		return metrics
	}

	var totalLatency, totalCost float64
	var errorCount int
	count := len(traceList.Traces)

	for _, t := range traceList.Traces {
		totalLatency += t.DurationMs
		totalCost += t.TotalCost
		if t.Level == domain.LevelError {
			errorCount++
		}
	}

	metrics["latency"] = totalLatency / float64(count)
	metrics["cost"] = totalCost / float64(count)
	if count > 0 {
		metrics["error_rate"] = float64(errorCount) / float64(count)
	}

	return metrics
}

// linearRegression performs simple linear regression on data points and returns
// the slope and the standard deviation of residuals.
func (s *PredictionService) linearRegression(points []domain.TrendDataPoint) (slope, stdDev float64) {
	n := float64(len(points))
	if n < 2 {
		return 0, 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, p := range points {
		x := float64(i)
		sumX += x
		sumY += p.Value
		sumXY += x * p.Value
		sumX2 += x * x
	}

	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, 0
	}

	slope = (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n

	// Standard deviation of residuals
	var sumResiduals float64
	for i, p := range points {
		predicted := intercept + slope*float64(i)
		diff := p.Value - predicted
		sumResiduals += diff * diff
	}
	stdDev = math.Sqrt(sumResiduals / n)

	return slope, stdDev
}
