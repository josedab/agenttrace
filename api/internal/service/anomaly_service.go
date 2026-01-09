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

// AnomalyService handles anomaly detection logic
type AnomalyService struct {
	logger *zap.Logger
}

// NewAnomalyService creates a new anomaly service
func NewAnomalyService(logger *zap.Logger) *AnomalyService {
	return &AnomalyService{
		logger: logger,
	}
}

// DefaultRuleConfig returns sensible defaults for a detection method
func (s *AnomalyService) DefaultRuleConfig(method domain.DetectionMethod) domain.AnomalyRuleConfig {
	config := domain.AnomalyRuleConfig{
		MinSamples:    30,
		LookbackHours: 24,
	}

	switch method {
	case domain.DetectionMethodZScore:
		config.ZScoreThreshold = 3.0
	case domain.DetectionMethodIQR:
		config.IQRMultiplier = 1.5
	case domain.DetectionMethodMAD:
		config.MADThreshold = 3.0
	case domain.DetectionMethodMovingAverage:
		config.WindowSize = 10
		config.Deviation = 0.2 // 20% deviation
	case domain.DetectionMethodExponentialEMA:
		config.Alpha = 0.3
		config.Deviation = 0.2
	case domain.DetectionMethodThreshold:
		// User must specify thresholds
	}

	return config
}

// DetectAnomaly runs anomaly detection on a set of data points
func (s *AnomalyService) DetectAnomaly(
	ctx context.Context,
	rule *domain.AnomalyRule,
	currentValue float64,
	historicalData []float64,
) (*domain.DetectionResult, error) {
	if len(historicalData) < rule.Config.MinSamples {
		return &domain.DetectionResult{
			IsAnomaly:   false,
			Description: fmt.Sprintf("Insufficient samples (%d < %d required)", len(historicalData), rule.Config.MinSamples),
		}, nil
	}

	// Calculate baseline statistics
	stats := s.CalculateBaselineStats(historicalData)

	var result *domain.DetectionResult
	var err error

	switch rule.Method {
	case domain.DetectionMethodZScore:
		result, err = s.detectZScore(currentValue, stats, rule.Config.ZScoreThreshold)
	case domain.DetectionMethodIQR:
		result, err = s.detectIQR(currentValue, stats, rule.Config.IQRMultiplier)
	case domain.DetectionMethodMAD:
		result, err = s.detectMAD(currentValue, stats, rule.Config.MADThreshold)
	case domain.DetectionMethodMovingAverage:
		result, err = s.detectMovingAverage(currentValue, historicalData, rule.Config.WindowSize, rule.Config.Deviation)
	case domain.DetectionMethodExponentialEMA:
		result, err = s.detectEMA(currentValue, historicalData, rule.Config.Alpha, rule.Config.Deviation)
	case domain.DetectionMethodThreshold:
		result, err = s.detectThreshold(currentValue, rule.Config.MinThreshold, rule.Config.MaxThreshold)
	default:
		return nil, fmt.Errorf("unsupported detection method: %s", rule.Method)
	}

	if err != nil {
		return nil, err
	}

	result.Method = rule.Method
	result.Value = currentValue
	result.Stats = stats

	// Determine severity based on score
	if result.IsAnomaly {
		result.Severity = s.determineSeverity(result.Score, result.Threshold, rule.Type)
	}

	return result, nil
}

// detectZScore detects anomalies using Z-score method
func (s *AnomalyService) detectZScore(value float64, stats domain.BaselineStats, threshold float64) (*domain.DetectionResult, error) {
	if stats.StdDev == 0 {
		return &domain.DetectionResult{
			IsAnomaly:   false,
			Score:       0,
			Threshold:   threshold,
			Expected:    stats.Mean,
			Description: "Standard deviation is zero, cannot compute Z-score",
		}, nil
	}

	zScore := math.Abs(value - stats.Mean) / stats.StdDev
	isAnomaly := zScore > threshold

	description := fmt.Sprintf("Z-score: %.2f (threshold: %.2f)", zScore, threshold)
	if isAnomaly {
		direction := "above"
		if value < stats.Mean {
			direction = "below"
		}
		description = fmt.Sprintf("Value %.2f is %.1f standard deviations %s mean (%.2f)", value, zScore, direction, stats.Mean)
	}

	return &domain.DetectionResult{
		IsAnomaly:   isAnomaly,
		Score:       zScore,
		Threshold:   threshold,
		Expected:    stats.Mean,
		Description: description,
	}, nil
}

// detectIQR detects anomalies using Interquartile Range method
func (s *AnomalyService) detectIQR(value float64, stats domain.BaselineStats, multiplier float64) (*domain.DetectionResult, error) {
	lowerBound := stats.Q1 - multiplier*stats.IQR
	upperBound := stats.Q3 + multiplier*stats.IQR

	isAnomaly := value < lowerBound || value > upperBound

	// Score is distance from nearest boundary in IQR units
	var score float64
	if value < lowerBound {
		score = (lowerBound - value) / stats.IQR
	} else if value > upperBound {
		score = (value - upperBound) / stats.IQR
	}

	description := fmt.Sprintf("IQR bounds: [%.2f, %.2f]", lowerBound, upperBound)
	if isAnomaly {
		description = fmt.Sprintf("Value %.2f outside IQR bounds [%.2f, %.2f]", value, lowerBound, upperBound)
	}

	return &domain.DetectionResult{
		IsAnomaly:   isAnomaly,
		Score:       score,
		Threshold:   multiplier,
		Expected:    stats.Median,
		Description: description,
	}, nil
}

// detectMAD detects anomalies using Median Absolute Deviation
func (s *AnomalyService) detectMAD(value float64, stats domain.BaselineStats, threshold float64) (*domain.DetectionResult, error) {
	if stats.MAD == 0 {
		return &domain.DetectionResult{
			IsAnomaly:   false,
			Score:       0,
			Threshold:   threshold,
			Expected:    stats.Median,
			Description: "MAD is zero, cannot compute modified Z-score",
		}, nil
	}

	// Modified Z-score using MAD
	// 0.6745 is the scale factor for consistency with standard deviation
	modifiedZScore := 0.6745 * math.Abs(value-stats.Median) / stats.MAD
	isAnomaly := modifiedZScore > threshold

	description := fmt.Sprintf("Modified Z-score: %.2f (threshold: %.2f)", modifiedZScore, threshold)
	if isAnomaly {
		direction := "above"
		if value < stats.Median {
			direction = "below"
		}
		description = fmt.Sprintf("Value %.2f is significantly %s median (%.2f), modified Z-score: %.2f", value, direction, stats.Median, modifiedZScore)
	}

	return &domain.DetectionResult{
		IsAnomaly:   isAnomaly,
		Score:       modifiedZScore,
		Threshold:   threshold,
		Expected:    stats.Median,
		Description: description,
	}, nil
}

// detectMovingAverage detects anomalies using simple moving average
func (s *AnomalyService) detectMovingAverage(value float64, data []float64, windowSize int, deviation float64) (*domain.DetectionResult, error) {
	if len(data) < windowSize {
		windowSize = len(data)
	}

	// Calculate moving average of last windowSize points
	recentData := data[len(data)-windowSize:]
	var sum float64
	for _, v := range recentData {
		sum += v
	}
	movingAvg := sum / float64(len(recentData))

	// Calculate deviation percentage
	if movingAvg == 0 {
		return &domain.DetectionResult{
			IsAnomaly:   false,
			Score:       0,
			Threshold:   deviation,
			Expected:    movingAvg,
			Description: "Moving average is zero",
		}, nil
	}

	deviationPct := math.Abs(value-movingAvg) / movingAvg
	isAnomaly := deviationPct > deviation

	description := fmt.Sprintf("Deviation from moving avg: %.1f%% (threshold: %.1f%%)", deviationPct*100, deviation*100)
	if isAnomaly {
		direction := "above"
		if value < movingAvg {
			direction = "below"
		}
		description = fmt.Sprintf("Value %.2f is %.1f%% %s moving average (%.2f)", value, deviationPct*100, direction, movingAvg)
	}

	return &domain.DetectionResult{
		IsAnomaly:   isAnomaly,
		Score:       deviationPct,
		Threshold:   deviation,
		Expected:    movingAvg,
		Description: description,
	}, nil
}

// detectEMA detects anomalies using exponential moving average
func (s *AnomalyService) detectEMA(value float64, data []float64, alpha float64, deviation float64) (*domain.DetectionResult, error) {
	if len(data) == 0 {
		return &domain.DetectionResult{
			IsAnomaly:   false,
			Score:       0,
			Threshold:   deviation,
			Expected:    0,
			Description: "No historical data",
		}, nil
	}

	// Calculate EMA
	ema := data[0]
	for i := 1; i < len(data); i++ {
		ema = alpha*data[i] + (1-alpha)*ema
	}

	if ema == 0 {
		return &domain.DetectionResult{
			IsAnomaly:   false,
			Score:       0,
			Threshold:   deviation,
			Expected:    ema,
			Description: "EMA is zero",
		}, nil
	}

	deviationPct := math.Abs(value-ema) / ema
	isAnomaly := deviationPct > deviation

	description := fmt.Sprintf("Deviation from EMA: %.1f%% (threshold: %.1f%%)", deviationPct*100, deviation*100)
	if isAnomaly {
		direction := "above"
		if value < ema {
			direction = "below"
		}
		description = fmt.Sprintf("Value %.2f is %.1f%% %s EMA (%.2f)", value, deviationPct*100, direction, ema)
	}

	return &domain.DetectionResult{
		IsAnomaly:   isAnomaly,
		Score:       deviationPct,
		Threshold:   deviation,
		Expected:    ema,
		Description: description,
	}, nil
}

// detectThreshold detects anomalies using static thresholds
func (s *AnomalyService) detectThreshold(value float64, minThreshold *float64, maxThreshold *float64) (*domain.DetectionResult, error) {
	var isAnomaly bool
	var description string
	var score float64
	expected := value // No baseline for threshold detection

	if minThreshold != nil && value < *minThreshold {
		isAnomaly = true
		score = (*minThreshold - value) / *minThreshold
		description = fmt.Sprintf("Value %.2f below minimum threshold %.2f", value, *minThreshold)
		expected = *minThreshold
	} else if maxThreshold != nil && value > *maxThreshold {
		isAnomaly = true
		score = (value - *maxThreshold) / *maxThreshold
		description = fmt.Sprintf("Value %.2f above maximum threshold %.2f", value, *maxThreshold)
		expected = *maxThreshold
	} else {
		bounds := ""
		if minThreshold != nil {
			bounds = fmt.Sprintf("[%.2f, ", *minThreshold)
		} else {
			bounds = "(-inf, "
		}
		if maxThreshold != nil {
			bounds += fmt.Sprintf("%.2f]", *maxThreshold)
		} else {
			bounds += "inf)"
		}
		description = fmt.Sprintf("Value %.2f within thresholds %s", value, bounds)
	}

	var threshold float64
	if minThreshold != nil {
		threshold = *minThreshold
	}
	if maxThreshold != nil {
		threshold = *maxThreshold
	}

	return &domain.DetectionResult{
		IsAnomaly:   isAnomaly,
		Score:       score,
		Threshold:   threshold,
		Expected:    expected,
		Description: description,
	}, nil
}

// CalculateBaselineStats calculates statistical baseline from historical data
func (s *AnomalyService) CalculateBaselineStats(data []float64) domain.BaselineStats {
	if len(data) == 0 {
		return domain.BaselineStats{}
	}

	// Create sorted copy
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	n := len(sorted)

	// Mean
	var sum float64
	for _, v := range sorted {
		sum += v
	}
	mean := sum / float64(n)

	// Standard deviation
	var sumSquares float64
	for _, v := range sorted {
		diff := v - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(n))

	// Median
	var median float64
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	} else {
		median = sorted[n/2]
	}

	// Percentiles
	p95 := s.percentile(sorted, 95)
	p99 := s.percentile(sorted, 99)

	// Quartiles for IQR
	q1 := s.percentile(sorted, 25)
	q3 := s.percentile(sorted, 75)
	iqr := q3 - q1

	// Median Absolute Deviation
	deviations := make([]float64, n)
	for i, v := range sorted {
		deviations[i] = math.Abs(v - median)
	}
	sort.Float64s(deviations)
	var mad float64
	if n%2 == 0 {
		mad = (deviations[n/2-1] + deviations[n/2]) / 2
	} else {
		mad = deviations[n/2]
	}

	return domain.BaselineStats{
		Mean:   mean,
		StdDev: stdDev,
		Median: median,
		P95:    p95,
		P99:    p99,
		Min:    sorted[0],
		Max:    sorted[n-1],
		Q1:     q1,
		Q3:     q3,
		IQR:    iqr,
		MAD:    mad,
	}
}

// percentile calculates the p-th percentile of sorted data
func (s *AnomalyService) percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := (p / 100) * float64(len(sorted)-1)
	lower := int(rank)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	fraction := rank - float64(lower)
	return sorted[lower] + fraction*(sorted[upper]-sorted[lower])
}

// determineSeverity determines anomaly severity based on score
func (s *AnomalyService) determineSeverity(score float64, threshold float64, anomalyType domain.AnomalyType) domain.AnomalySeverity {
	// Ratio of how far beyond threshold
	ratio := score / threshold

	switch {
	case ratio > 3.0:
		return domain.AnomalySeverityCritical
	case ratio > 2.0:
		return domain.AnomalySeverityHigh
	case ratio > 1.5:
		return domain.AnomalySeverityMedium
	default:
		return domain.AnomalySeverityLow
	}
}

// CreateAnomaly creates an anomaly record from detection result
func (s *AnomalyService) CreateAnomaly(
	projectID uuid.UUID,
	rule *domain.AnomalyRule,
	result *domain.DetectionResult,
	context AnomalyContext,
) *domain.Anomaly {
	return &domain.Anomaly{
		ID:        uuid.New(),
		ProjectID: projectID,
		RuleID:    rule.ID,
		Type:      rule.Type,
		Severity:  result.Severity,

		DetectedAt:  time.Now(),
		Method:      result.Method,
		Score:       result.Score,
		Value:       result.Value,
		Expected:    result.Expected,
		Threshold:   result.Threshold,
		Description: result.Description,

		TraceID:       context.TraceID,
		TraceName:     context.TraceName,
		SpanID:        context.SpanID,
		SpanName:      context.SpanName,
		Metadata:      context.Metadata,
		TimeWindow:    context.TimeWindow,
		SampleCount:   context.SampleCount,
		BaselineStats: result.Stats,

		AlertsSent: []domain.AlertRecord{},
	}
}

// AnomalyContext provides context for anomaly creation
type AnomalyContext struct {
	TraceID     *uuid.UUID
	TraceName   string
	SpanID      *uuid.UUID
	SpanName    string
	Metadata    map[string]string
	TimeWindow  domain.TimeWindow
	SampleCount int
}

// CreateAlert creates an alert from an anomaly
func (s *AnomalyService) CreateAlert(anomaly *domain.Anomaly, rule *domain.AnomalyRule) *domain.Alert {
	var deviation float64
	if anomaly.Expected != 0 {
		deviation = ((anomaly.Value - anomaly.Expected) / anomaly.Expected) * 100
	}

	return &domain.Alert{
		ID:        uuid.New(),
		ProjectID: anomaly.ProjectID,
		AnomalyID: anomaly.ID,
		RuleID:    rule.ID,
		Status:    domain.AlertStatusActive,
		Severity:  anomaly.Severity,

		Title:       fmt.Sprintf("%s Anomaly Detected: %s", rule.Type, rule.Name),
		Description: anomaly.Description,
		Type:        anomaly.Type,

		CurrentValue:  anomaly.Value,
		ExpectedValue: anomaly.Expected,
		Deviation:     deviation,

		TriggeredAt: time.Now(),
		Notes:       []domain.AlertNote{},
	}
}

// FormatAlertMessage formats an alert for notification
func (s *AnomalyService) FormatAlertMessage(alert *domain.Alert, anomaly *domain.Anomaly) string {
	var msg string

	msg += fmt.Sprintf("**%s**\n\n", alert.Title)
	msg += fmt.Sprintf("%s\n\n", alert.Description)

	msg += "**Details:**\n"
	msg += fmt.Sprintf("- Severity: %s\n", alert.Severity)
	msg += fmt.Sprintf("- Current Value: %.2f\n", alert.CurrentValue)
	msg += fmt.Sprintf("- Expected Value: %.2f\n", alert.ExpectedValue)
	msg += fmt.Sprintf("- Deviation: %.1f%%\n", alert.Deviation)

	if anomaly.TraceName != "" {
		msg += fmt.Sprintf("- Trace: %s\n", anomaly.TraceName)
	}

	msg += fmt.Sprintf("\nDetected at: %s", alert.TriggeredAt.Format(time.RFC3339))

	return msg
}

// GetAnomalyStats calculates anomaly statistics for a project
func (s *AnomalyService) GetAnomalyStats(
	ctx context.Context,
	projectID uuid.UUID,
	anomalies []domain.Anomaly,
	activeAlerts int,
	period domain.TimeWindow,
) *domain.AnomalyStats {
	stats := &domain.AnomalyStats{
		ProjectID:      projectID,
		Period:         period,
		TotalAnomalies: len(anomalies),
		ActiveAlerts:   activeAlerts,
		BySeverity:     make(map[domain.AnomalySeverity]int),
		ByType:         make(map[domain.AnomalyType]int),
	}

	traceCount := make(map[string]int)

	for _, a := range anomalies {
		stats.BySeverity[a.Severity]++
		stats.ByType[a.Type]++
		if a.TraceName != "" {
			traceCount[a.TraceName]++
		}
	}

	// Top affected traces
	type traceEntry struct {
		name  string
		count int
	}
	var traces []traceEntry
	for name, count := range traceCount {
		traces = append(traces, traceEntry{name, count})
	}
	sort.Slice(traces, func(i, j int) bool {
		return traces[i].count > traces[j].count
	})

	for i := 0; i < len(traces) && i < 10; i++ {
		stats.TopAffectedTraces = append(stats.TopAffectedTraces, domain.TraceAnomalyCount{
			TraceName: traces[i].name,
			Count:     traces[i].count,
		})
	}

	return stats
}

// ShouldTriggerAlert determines if an alert should be triggered based on cooldown
func (s *AnomalyService) ShouldTriggerAlert(
	rule *domain.AnomalyRule,
	lastAlertTime *time.Time,
) bool {
	if lastAlertTime == nil {
		return true
	}

	cooldown := time.Duration(rule.Cooldown) * time.Minute
	return time.Since(*lastAlertTime) > cooldown
}
