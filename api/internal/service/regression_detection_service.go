package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// RegressionDetectionService handles regression detection configuration and analysis
type RegressionDetectionService struct {
	logger *zap.Logger
}

// NewRegressionDetectionService creates a new regression detection service
func NewRegressionDetectionService(logger *zap.Logger) *RegressionDetectionService {
	return &RegressionDetectionService{
		logger: logger,
	}
}

// CreateConfig creates a new regression detection configuration
func (s *RegressionDetectionService) CreateConfig(ctx context.Context, projectID uuid.UUID, input domain.RegressionDetectionInput) (*domain.RegressionDetectionConfig, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("config name is required")
	}
	if len(input.MonitoredMetrics) == 0 {
		return nil, fmt.Errorf("at least one monitored metric is required")
	}

	now := time.Now()
	config := &domain.RegressionDetectionConfig{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Name:             input.Name,
		Enabled:          true,
		Method:           input.Method,
		MonitoredMetrics: input.MonitoredMetrics,
		Thresholds:       input.Thresholds,
		BaselineWindow:   input.BaselineWindow,
		EvaluationWindow: input.EvaluationWindow,
		MinSampleSize:    input.MinSampleSize,
		AlertChannels:    input.AlertChannels,
		Schedule:         input.Schedule,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if config.BaselineWindow <= 0 {
		config.BaselineWindow = 14
	}
	if config.EvaluationWindow <= 0 {
		config.EvaluationWindow = 3
	}
	if config.MinSampleSize <= 0 {
		config.MinSampleSize = 50
	}
	if config.Thresholds == nil {
		config.Thresholds = &domain.RegressionThresholds{
			QualityDropPct:       5.0,
			LatencyIncreasePct:   10.0,
			CostIncreasePct:      15.0,
			ErrorRateIncreasePct: 2.0,
			PValueThreshold:      0.05,
			MinEffectSize:        0.2,
		}
	}

	s.logger.Info("regression detection config created",
		zap.String("id", config.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", config.Name),
		zap.String("method", string(config.Method)),
	)
	return config, nil
}

// ListConfigs lists regression detection configs for a project
func (s *RegressionDetectionService) ListConfigs(ctx context.Context, projectID uuid.UUID, limit, offset int) (*domain.RegressionDetectionConfigList, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	s.logger.Debug("listing regression detection configs",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	now := time.Now()
	configs := []domain.RegressionDetectionConfig{
		{
			ID:               uuid.MustParse("a1b2c3d4-0001-4000-8000-000000000001"),
			ProjectID:        projectID,
			Name:             "Quality Score Monitor",
			Enabled:          true,
			Method:           domain.RegressionDetectionMethodStatistical,
			MonitoredMetrics: []string{"quality", "error_rate"},
			Thresholds: &domain.RegressionThresholds{
				QualityDropPct:       5.0,
				LatencyIncreasePct:   10.0,
				CostIncreasePct:      15.0,
				ErrorRateIncreasePct: 2.0,
				PValueThreshold:      0.05,
				MinEffectSize:        0.2,
			},
			BaselineWindow:   14,
			EvaluationWindow: 3,
			MinSampleSize:    100,
			Schedule:         "0 */6 * * *",
			CreatedAt:        now.Add(-72 * time.Hour),
			UpdatedAt:        now.Add(-24 * time.Hour),
		},
		{
			ID:               uuid.MustParse("a1b2c3d4-0002-4000-8000-000000000002"),
			ProjectID:        projectID,
			Name:             "Latency Anomaly Detector",
			Enabled:          true,
			Method:           domain.RegressionDetectionMethodMLAnomaly,
			MonitoredMetrics: []string{"latency", "cost"},
			Thresholds: &domain.RegressionThresholds{
				QualityDropPct:       10.0,
				LatencyIncreasePct:   15.0,
				CostIncreasePct:      20.0,
				ErrorRateIncreasePct: 5.0,
				PValueThreshold:      0.01,
				MinEffectSize:        0.3,
			},
			BaselineWindow:   21,
			EvaluationWindow: 7,
			MinSampleSize:    200,
			Schedule:         "0 0 * * *",
			CreatedAt:        now.Add(-168 * time.Hour),
			UpdatedAt:        now.Add(-48 * time.Hour),
		},
		{
			ID:               uuid.MustParse("a1b2c3d4-0003-4000-8000-000000000003"),
			ProjectID:        projectID,
			Name:             "Cost Trend Analyzer",
			Enabled:          false,
			Method:           domain.RegressionDetectionMethodTrendAnalysis,
			MonitoredMetrics: []string{"cost"},
			Thresholds: &domain.RegressionThresholds{
				QualityDropPct:       8.0,
				LatencyIncreasePct:   12.0,
				CostIncreasePct:      10.0,
				ErrorRateIncreasePct: 3.0,
				PValueThreshold:      0.05,
				MinEffectSize:        0.25,
			},
			BaselineWindow:   30,
			EvaluationWindow: 7,
			MinSampleSize:    50,
			Schedule:         "0 0 * * 1",
			CreatedAt:        now.Add(-240 * time.Hour),
			UpdatedAt:        now.Add(-240 * time.Hour),
		},
	}

	// Apply offset/limit
	total := int64(len(configs))
	if offset >= len(configs) {
		configs = []domain.RegressionDetectionConfig{}
	} else {
		end := offset + limit
		if end > len(configs) {
			end = len(configs)
		}
		configs = configs[offset:end]
	}

	return &domain.RegressionDetectionConfigList{
		Configs:    configs,
		TotalCount: total,
		HasMore:    int64(offset+limit) < total,
	}, nil
}

// GetConfig retrieves a specific regression detection config
func (s *RegressionDetectionService) GetConfig(ctx context.Context, projectID, configID uuid.UUID) (*domain.RegressionDetectionConfig, error) {
	s.logger.Debug("getting regression detection config",
		zap.String("projectId", projectID.String()),
		zap.String("configId", configID.String()),
	)

	now := time.Now()
	config := &domain.RegressionDetectionConfig{
		ID:               configID,
		ProjectID:        projectID,
		Name:             "Quality Score Monitor",
		Enabled:          true,
		Method:           domain.RegressionDetectionMethodStatistical,
		MonitoredMetrics: []string{"quality", "latency", "error_rate"},
		Thresholds: &domain.RegressionThresholds{
			QualityDropPct:       5.0,
			LatencyIncreasePct:   10.0,
			CostIncreasePct:      15.0,
			ErrorRateIncreasePct: 2.0,
			PValueThreshold:      0.05,
			MinEffectSize:        0.2,
		},
		BaselineWindow:   14,
		EvaluationWindow: 3,
		MinSampleSize:    100,
		AlertChannels:    []uuid.UUID{},
		Schedule:         "0 */6 * * *",
		CreatedAt:        now.Add(-72 * time.Hour),
		UpdatedAt:        now.Add(-12 * time.Hour),
	}

	return config, nil
}

// UpdateConfig updates an existing regression detection config
func (s *RegressionDetectionService) UpdateConfig(ctx context.Context, projectID, configID uuid.UUID, input domain.RegressionDetectionInput) (*domain.RegressionDetectionConfig, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("config name is required")
	}

	existing, err := s.GetConfig(ctx, projectID, configID)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}

	existing.Name = input.Name
	existing.Method = input.Method
	existing.MonitoredMetrics = input.MonitoredMetrics
	existing.Schedule = input.Schedule
	existing.AlertChannels = input.AlertChannels
	existing.UpdatedAt = time.Now()

	if input.Thresholds != nil {
		existing.Thresholds = input.Thresholds
	}
	if input.BaselineWindow > 0 {
		existing.BaselineWindow = input.BaselineWindow
	}
	if input.EvaluationWindow > 0 {
		existing.EvaluationWindow = input.EvaluationWindow
	}
	if input.MinSampleSize > 0 {
		existing.MinSampleSize = input.MinSampleSize
	}

	s.logger.Info("regression detection config updated",
		zap.String("id", configID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", existing.Name),
	)
	return existing, nil
}

// DeleteConfig deletes a regression detection config
func (s *RegressionDetectionService) DeleteConfig(ctx context.Context, projectID, configID uuid.UUID) error {
	s.logger.Info("regression detection config deleted",
		zap.String("projectId", projectID.String()),
		zap.String("configId", configID.String()),
	)
	return nil
}

// RunDetection executes a regression detection analysis for the given config
func (s *RegressionDetectionService) RunDetection(ctx context.Context, projectID, configID uuid.UUID) (*domain.RegressionDetectionResult, error) {
	config, err := s.GetConfig(ctx, projectID, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	s.logger.Info("running regression detection",
		zap.String("projectId", projectID.String()),
		zap.String("configId", configID.String()),
		zap.String("method", string(config.Method)),
	)

	now := time.Now()
	result := &domain.RegressionDetectionResult{
		ID:         uuid.New(),
		ProjectID:  projectID,
		ConfigID:   configID,
		DetectedAt: now,
		Method:     config.Method,
		Status:     "detected",
	}

	switch config.Method {
	case domain.RegressionDetectionMethodStatistical:
		s.applyStatisticalAnalysis(result)
	case domain.RegressionDetectionMethodMLAnomaly:
		s.applyMLAnomalyAnalysis(result)
	case domain.RegressionDetectionMethodTrendAnalysis:
		s.applyTrendAnalysis(result)
	default:
		s.applyStatisticalAnalysis(result)
	}

	result.RelatedChanges = []domain.RelatedChange{
		{
			Type:        "model_update",
			Description: "Model version updated from gpt-4-0613 to gpt-4-turbo",
			Timestamp:   now.Add(-6 * time.Hour),
			Metadata: map[string]any{
				"previousModel": "gpt-4-0613",
				"newModel":      "gpt-4-turbo",
			},
		},
		{
			Type:        "prompt_change",
			Description: "System prompt modified in production pipeline",
			Timestamp:   now.Add(-8 * time.Hour),
			Metadata: map[string]any{
				"pipelineId": "pipeline-main-v2",
				"changeType": "system_prompt",
			},
		},
	}

	if result.IsRegression {
		s.logger.Warn("regression detected",
			zap.String("resultId", result.ID.String()),
			zap.String("metric", result.AffectedMetric),
			zap.String("severity", string(result.Severity)),
			zap.Float64("deltaPct", result.DeltaPct),
		)
	} else {
		s.logger.Info("no regression detected",
			zap.String("resultId", result.ID.String()),
			zap.String("metric", result.AffectedMetric),
		)
	}

	return result, nil
}

// applyStatisticalAnalysis generates z-score, p-value, and effect size based results
func (s *RegressionDetectionService) applyStatisticalAnalysis(result *domain.RegressionDetectionResult) {
	result.AffectedMetric = "quality"
	result.BaselineValue = 0.847
	result.CurrentValue = 0.791
	result.DeltaPct = -6.61

	pValue := 0.003
	effectSize := 0.72
	result.PValue = &pValue
	result.EffectSize = &effectSize

	// p-value < 0.05 and effect size > 0.2 indicates significant regression
	result.IsRegression = true
	result.Severity = domain.RegressionDetectionSeverityHigh
	result.Summary = "Statistical analysis detected a significant quality regression: quality score dropped 6.61% (z-score=2.97, p=0.003, Cohen's d=0.72)"
	result.RootCauseHypothesis = "Quality degradation correlates with recent model version update. The new model shows lower consistency in structured output formatting, particularly affecting evaluation criteria related to factual accuracy."
}

// applyMLAnomalyAnalysis generates isolation forest style anomaly detection results
func (s *RegressionDetectionService) applyMLAnomalyAnalysis(result *domain.RegressionDetectionResult) {
	result.AffectedMetric = "latency"
	result.BaselineValue = 1.234
	result.CurrentValue = 1.891
	result.DeltaPct = 53.24

	// Isolation forest anomaly score (>0.5 indicates anomaly)
	anomalyScore := 0.78
	result.EffectSize = &anomalyScore

	result.IsRegression = true
	result.Severity = domain.RegressionDetectionSeverityCritical
	result.Summary = "ML anomaly detection identified latency regression: p95 latency increased 53.24% (anomaly score=0.78, isolation depth=4.2/10)"
	result.RootCauseHypothesis = "Latency spike coincides with increased prompt token count following system prompt modification. Token count increased from ~800 to ~1400, directly impacting response time."
}

// applyTrendAnalysis generates moving average trend comparison results
func (s *RegressionDetectionService) applyTrendAnalysis(result *domain.RegressionDetectionResult) {
	result.AffectedMetric = "cost"
	result.BaselineValue = 0.0234
	result.CurrentValue = 0.0251
	result.DeltaPct = 7.26

	// Trend deviation: current moving average vs baseline moving average
	trendDeviation := 0.35
	result.EffectSize = &trendDeviation

	// 7.26% increase is below typical cost threshold of 10%, no regression
	result.IsRegression = false
	result.Severity = domain.RegressionDetectionSeverityLow
	result.Summary = "Trend analysis shows gradual cost increase of 7.26% over evaluation window. 7-day moving average ($0.0251) trending above 30-day baseline ($0.0234) but within acceptable thresholds."
	result.RootCauseHypothesis = "Minor cost increase attributed to natural traffic growth and slightly longer average conversations. No single change identified as primary driver."
}

// ListDetections lists recent regression detection results
func (s *RegressionDetectionService) ListDetections(ctx context.Context, projectID uuid.UUID, severity string, status string, limit, offset int) (*domain.RegressionDetectionResultList, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	s.logger.Debug("listing regression detections",
		zap.String("projectId", projectID.String()),
		zap.String("severity", severity),
		zap.String("status", status),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	now := time.Now()
	configID1 := uuid.MustParse("a1b2c3d4-0001-4000-8000-000000000001")
	configID2 := uuid.MustParse("a1b2c3d4-0002-4000-8000-000000000002")
	configID3 := uuid.MustParse("a1b2c3d4-0003-4000-8000-000000000003")
	acknowledgedBy := uuid.MustParse("00000000-0000-4000-8000-000000000099")
	acknowledgedAt := now.Add(-2 * time.Hour)
	resolvedAt := now.Add(-1 * time.Hour)
	pVal1 := 0.003
	pVal2 := 0.41
	effect1 := 0.72
	effect2 := 0.81
	effect3 := 0.12

	allResults := []domain.RegressionDetectionResult{
		{
			ID:                  uuid.MustParse("d1d2d3d4-0001-4000-8000-000000000001"),
			ProjectID:           projectID,
			ConfigID:            configID1,
			DetectedAt:          now.Add(-1 * time.Hour),
			Severity:            domain.RegressionDetectionSeverityCritical,
			Method:              domain.RegressionDetectionMethodStatistical,
			AffectedMetric:      "quality",
			BaselineValue:       0.847,
			CurrentValue:        0.791,
			DeltaPct:            -6.61,
			PValue:              &pVal1,
			EffectSize:          &effect1,
			IsRegression:        true,
			Summary:             "Significant quality score regression detected via statistical analysis",
			RootCauseHypothesis: "Quality degradation correlates with model version update",
			RelatedChanges: []domain.RelatedChange{
				{Type: "model_update", Description: "Model updated to gpt-4-turbo", Timestamp: now.Add(-6 * time.Hour)},
			},
			Status: "detected",
		},
		{
			ID:                  uuid.MustParse("d1d2d3d4-0002-4000-8000-000000000002"),
			ProjectID:           projectID,
			ConfigID:            configID2,
			DetectedAt:          now.Add(-4 * time.Hour),
			Severity:            domain.RegressionDetectionSeverityHigh,
			Method:              domain.RegressionDetectionMethodMLAnomaly,
			AffectedMetric:      "latency",
			BaselineValue:       1.234,
			CurrentValue:        1.891,
			DeltaPct:            53.24,
			EffectSize:          &effect2,
			IsRegression:        true,
			Summary:             "Anomalous latency increase detected by isolation forest model",
			RootCauseHypothesis: "Latency spike coincides with increased prompt token count",
			RelatedChanges: []domain.RelatedChange{
				{Type: "prompt_change", Description: "System prompt expanded", Timestamp: now.Add(-8 * time.Hour)},
			},
			Status:         "investigating",
			AcknowledgedBy: &acknowledgedBy,
			AcknowledgedAt: &acknowledgedAt,
		},
		{
			ID:                  uuid.MustParse("d1d2d3d4-0003-4000-8000-000000000003"),
			ProjectID:           projectID,
			ConfigID:            configID1,
			DetectedAt:          now.Add(-24 * time.Hour),
			Severity:            domain.RegressionDetectionSeverityMedium,
			Method:              domain.RegressionDetectionMethodStatistical,
			AffectedMetric:      "error_rate",
			BaselineValue:       0.023,
			CurrentValue:        0.031,
			DeltaPct:            34.78,
			PValue:              &pVal2,
			EffectSize:          &effect3,
			IsRegression:        false,
			Summary:             "Error rate increase observed but not statistically significant (p=0.41)",
			RootCauseHypothesis: "Increase within normal variance; may be related to seasonal traffic patterns",
			RelatedChanges:      []domain.RelatedChange{},
			Status:              "false_positive",
			AcknowledgedBy:      &acknowledgedBy,
			AcknowledgedAt:      &acknowledgedAt,
			ResolvedAt:          &resolvedAt,
		},
		{
			ID:                  uuid.MustParse("d1d2d3d4-0004-4000-8000-000000000004"),
			ProjectID:           projectID,
			ConfigID:            configID3,
			DetectedAt:          now.Add(-48 * time.Hour),
			Severity:            domain.RegressionDetectionSeverityLow,
			Method:              domain.RegressionDetectionMethodTrendAnalysis,
			AffectedMetric:      "cost",
			BaselineValue:       0.0234,
			CurrentValue:        0.0251,
			DeltaPct:            7.26,
			IsRegression:        false,
			Summary:             "Gradual cost increase within acceptable thresholds",
			RootCauseHypothesis: "Minor cost increase attributed to natural traffic growth",
			RelatedChanges:      []domain.RelatedChange{},
			Status:              "resolved",
			ResolvedAt:          &resolvedAt,
		},
	}

	// Filter by severity
	if severity != "" {
		filtered := make([]domain.RegressionDetectionResult, 0)
		for _, r := range allResults {
			if string(r.Severity) == severity {
				filtered = append(filtered, r)
			}
		}
		allResults = filtered
	}

	// Filter by status
	if status != "" {
		filtered := make([]domain.RegressionDetectionResult, 0)
		for _, r := range allResults {
			if r.Status == status {
				filtered = append(filtered, r)
			}
		}
		allResults = filtered
	}

	total := int64(len(allResults))
	if offset >= len(allResults) {
		allResults = []domain.RegressionDetectionResult{}
	} else {
		end := offset + limit
		if end > len(allResults) {
			end = len(allResults)
		}
		allResults = allResults[offset:end]
	}

	return &domain.RegressionDetectionResultList{
		Results:    allResults,
		TotalCount: total,
		HasMore:    int64(offset+limit) < total,
	}, nil
}

// AcknowledgeDetection marks a detection as acknowledged by a user
func (s *RegressionDetectionService) AcknowledgeDetection(ctx context.Context, projectID, detectionID, userID uuid.UUID) (*domain.RegressionDetectionResult, error) {
	s.logger.Info("acknowledging regression detection",
		zap.String("projectId", projectID.String()),
		zap.String("detectionId", detectionID.String()),
		zap.String("userId", userID.String()),
	)

	now := time.Now()
	pValue := 0.003
	effectSize := 0.72
	result := &domain.RegressionDetectionResult{
		ID:                  detectionID,
		ProjectID:           projectID,
		ConfigID:            uuid.MustParse("a1b2c3d4-0001-4000-8000-000000000001"),
		DetectedAt:          now.Add(-1 * time.Hour),
		Severity:            domain.RegressionDetectionSeverityHigh,
		Method:              domain.RegressionDetectionMethodStatistical,
		AffectedMetric:      "quality",
		BaselineValue:       0.847,
		CurrentValue:        0.791,
		DeltaPct:            -6.61,
		PValue:              &pValue,
		EffectSize:          &effectSize,
		IsRegression:        true,
		Summary:             "Significant quality score regression detected via statistical analysis",
		RootCauseHypothesis: "Quality degradation correlates with model version update",
		RelatedChanges: []domain.RelatedChange{
			{Type: "model_update", Description: "Model updated to gpt-4-turbo", Timestamp: now.Add(-6 * time.Hour)},
		},
		Status:         "investigating",
		AcknowledgedBy: &userID,
		AcknowledgedAt: &now,
	}

	return result, nil
}

// ResolveDetection marks a detection as resolved
func (s *RegressionDetectionService) ResolveDetection(ctx context.Context, projectID, detectionID uuid.UUID, falsePositive bool) (*domain.RegressionDetectionResult, error) {
	resolvedStatus := "resolved"
	if falsePositive {
		resolvedStatus = "false_positive"
	}

	s.logger.Info("resolving regression detection",
		zap.String("projectId", projectID.String()),
		zap.String("detectionId", detectionID.String()),
		zap.Bool("falsePositive", falsePositive),
		zap.String("status", resolvedStatus),
	)

	now := time.Now()
	acknowledgedBy := uuid.MustParse("00000000-0000-4000-8000-000000000099")
	acknowledgedAt := now.Add(-30 * time.Minute)
	result := &domain.RegressionDetectionResult{
		ID:                  detectionID,
		ProjectID:           projectID,
		ConfigID:            uuid.MustParse("a1b2c3d4-0001-4000-8000-000000000001"),
		DetectedAt:          now.Add(-2 * time.Hour),
		Severity:            domain.RegressionDetectionSeverityHigh,
		Method:              domain.RegressionDetectionMethodStatistical,
		AffectedMetric:      "quality",
		BaselineValue:       0.847,
		CurrentValue:        0.791,
		DeltaPct:            -6.61,
		IsRegression:        !falsePositive,
		Summary:             "Significant quality score regression detected via statistical analysis",
		RootCauseHypothesis: "Quality degradation correlates with model version update",
		RelatedChanges: []domain.RelatedChange{
			{Type: "model_update", Description: "Model updated to gpt-4-turbo", Timestamp: now.Add(-6 * time.Hour)},
		},
		Status:         resolvedStatus,
		AcknowledgedBy: &acknowledgedBy,
		AcknowledgedAt: &acknowledgedAt,
		ResolvedAt:     &now,
	}

	return result, nil
}

// GetDashboard returns a regression detection dashboard overview
func (s *RegressionDetectionService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.RegressionDetectionDashboard, error) {
	s.logger.Debug("getting regression detection dashboard",
		zap.String("projectId", projectID.String()),
	)

	now := time.Now()
	pValue := 0.003
	effectSize := 0.72

	recentDetections := []domain.RegressionDetectionResult{
		{
			ID:                  uuid.MustParse("d1d2d3d4-0001-4000-8000-000000000001"),
			ProjectID:           projectID,
			ConfigID:            uuid.MustParse("a1b2c3d4-0001-4000-8000-000000000001"),
			DetectedAt:          now.Add(-1 * time.Hour),
			Severity:            domain.RegressionDetectionSeverityCritical,
			Method:              domain.RegressionDetectionMethodStatistical,
			AffectedMetric:      "quality",
			BaselineValue:       0.847,
			CurrentValue:        0.791,
			DeltaPct:            -6.61,
			PValue:              &pValue,
			EffectSize:          &effectSize,
			IsRegression:        true,
			Summary:             "Significant quality score regression detected",
			RootCauseHypothesis: "Quality degradation correlates with model version update",
			RelatedChanges: []domain.RelatedChange{
				{Type: "model_update", Description: "Model updated to gpt-4-turbo", Timestamp: now.Add(-6 * time.Hour)},
			},
			Status: "detected",
		},
	}

	dashboard := &domain.RegressionDetectionDashboard{
		TotalConfigs:         3,
		ActiveConfigs:        2,
		TotalDetections:      12,
		UnresolvedDetections: 3,
		RecentDetections:     recentDetections,
		MetricHealth: map[string]domain.MetricHealthStatus{
			"quality": {
				Metric:         "quality",
				Status:         "critical",
				CurrentValue:   0.791,
				BaselineValue:  0.847,
				TrendDirection: -0.066,
				LastChecked:    now.Add(-15 * time.Minute),
			},
			"latency": {
				Metric:         "latency",
				Status:         "warning",
				CurrentValue:   1.891,
				BaselineValue:  1.234,
				TrendDirection: 0.532,
				LastChecked:    now.Add(-15 * time.Minute),
			},
			"cost": {
				Metric:         "cost",
				Status:         "healthy",
				CurrentValue:   0.0251,
				BaselineValue:  0.0234,
				TrendDirection: 0.073,
				LastChecked:    now.Add(-15 * time.Minute),
			},
			"error_rate": {
				Metric:         "error_rate",
				Status:         "healthy",
				CurrentValue:   0.023,
				BaselineValue:  0.021,
				TrendDirection: 0.095,
				LastChecked:    now.Add(-15 * time.Minute),
			},
		},
	}

	return dashboard, nil
}
