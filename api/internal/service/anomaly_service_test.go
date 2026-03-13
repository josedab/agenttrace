package service

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func newTestAnomalyService() *AnomalyService {
	logger, _ := zap.NewDevelopment()
	return NewAnomalyService(logger)
}

func TestAnomalyService_DetectZScore(t *testing.T) {
	svc := newTestAnomalyService()

	tests := []struct {
		name      string
		value     float64
		data      []float64
		threshold float64
		wantAnom  bool
	}{
		{
			name:      "normal distribution value within threshold",
			value:     10.5,
			data:      []float64{10, 10, 11, 9, 10, 11, 10, 9, 10, 11},
			threshold: 3.0,
			wantAnom:  false,
		},
		{
			name:      "outlier beyond threshold",
			value:     100.0,
			data:      []float64{10, 10, 11, 9, 10, 11, 10, 9, 10, 11},
			threshold: 3.0,
			wantAnom:  true,
		},
		{
			name:      "all same values - stddev zero",
			value:     5.0,
			data:      []float64{5, 5, 5, 5, 5},
			threshold: 3.0,
			wantAnom:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateBaselineStats(tt.data)
			result, err := svc.detectZScore(tt.value, stats, tt.threshold)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAnom, result.IsAnomaly)
		})
	}
}

func TestAnomalyService_DetectIQR(t *testing.T) {
	svc := newTestAnomalyService()

	tests := []struct {
		name       string
		value      float64
		data       []float64
		multiplier float64
		wantAnom   bool
	}{
		{
			name:       "skewed data value within bounds",
			value:      15.0,
			data:       []float64{1, 2, 3, 4, 5, 10, 15, 20, 25, 30},
			multiplier: 1.5,
			wantAnom:   false,
		},
		{
			name:       "skewed data value outside bounds",
			value:      100.0,
			data:       []float64{1, 2, 3, 4, 5, 10, 15, 20, 25, 30},
			multiplier: 1.5,
			wantAnom:   true,
		},
		{
			name:       "single value data - value outside trivial bounds",
			value:      10.0,
			data:       []float64{5},
			multiplier: 1.5,
			wantAnom:   true, // IQR=0, bounds=[5,5], value 10 is outside
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateBaselineStats(tt.data)
			result, err := svc.detectIQR(tt.value, stats, tt.multiplier)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAnom, result.IsAnomaly)
		})
	}
}

func TestAnomalyService_DetectMAD(t *testing.T) {
	svc := newTestAnomalyService()

	tests := []struct {
		name      string
		value     float64
		data      []float64
		threshold float64
		wantAnom  bool
	}{
		{
			name:      "all same values - MAD zero",
			value:     10.0,
			data:      []float64{5, 5, 5, 5, 5},
			threshold: 3.0,
			wantAnom:  false,
		},
		{
			name:      "median accuracy - value near median",
			value:     5.0,
			data:      []float64{1, 3, 5, 7, 9},
			threshold: 3.0,
			wantAnom:  false,
		},
		{
			name:      "value far from median",
			value:     100.0,
			data:      []float64{1, 3, 5, 7, 9},
			threshold: 3.0,
			wantAnom:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateBaselineStats(tt.data)
			result, err := svc.detectMAD(tt.value, stats, tt.threshold)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAnom, result.IsAnomaly)
		})
	}
}

func TestAnomalyService_DetectMovingAverage(t *testing.T) {
	svc := newTestAnomalyService()

	tests := []struct {
		name       string
		value      float64
		data       []float64
		windowSize int
		deviation  float64
		wantAnom   bool
	}{
		{
			name:       "window larger than data",
			value:      10.0,
			data:       []float64{10, 10, 10},
			windowSize: 100,
			deviation:  0.2,
			wantAnom:   false,
		},
		{
			name:       "value at deviation boundary - not anomaly",
			value:      12.0,
			data:       []float64{10, 10, 10, 10, 10},
			windowSize: 5,
			deviation:  0.2,
			wantAnom:   false,
		},
		{
			name:       "value beyond deviation boundary",
			value:      15.0,
			data:       []float64{10, 10, 10, 10, 10},
			windowSize: 5,
			deviation:  0.2,
			wantAnom:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.detectMovingAverage(tt.value, tt.data, tt.windowSize, tt.deviation)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAnom, result.IsAnomaly)
		})
	}
}

func TestAnomalyService_DetectEMA(t *testing.T) {
	svc := newTestAnomalyService()

	tests := []struct {
		name      string
		value     float64
		data      []float64
		alpha     float64
		deviation float64
		wantAnom  bool
	}{
		{
			name:      "alpha=0 EMA equals first value",
			value:     10.0,
			data:      []float64{10, 20, 30, 40, 50},
			alpha:     0.0,
			deviation: 0.2,
			wantAnom:  false, // EMA stays at 10
		},
		{
			name:      "alpha=1 EMA equals last value",
			value:     50.0,
			data:      []float64{10, 20, 30, 40, 50},
			alpha:     1.0,
			deviation: 0.2,
			wantAnom:  false, // EMA = last value = 50
		},
		{
			name:      "empty data returns not anomaly",
			value:     100.0,
			data:      []float64{},
			alpha:     0.3,
			deviation: 0.2,
			wantAnom:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.detectEMA(tt.value, tt.data, tt.alpha, tt.deviation)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAnom, result.IsAnomaly)
		})
	}
}

func TestAnomalyService_DetectThreshold(t *testing.T) {
	svc := newTestAnomalyService()

	floatPtr := func(v float64) *float64 { return &v }

	tests := []struct {
		name    string
		value   float64
		min     *float64
		max     *float64
		wantAnom bool
	}{
		{
			name:    "min only - value below",
			value:   5.0,
			min:     floatPtr(10.0),
			max:     nil,
			wantAnom: true,
		},
		{
			name:    "min only - value above",
			value:   15.0,
			min:     floatPtr(10.0),
			max:     nil,
			wantAnom: false,
		},
		{
			name:    "max only - value above",
			value:   15.0,
			min:     nil,
			max:     floatPtr(10.0),
			wantAnom: true,
		},
		{
			name:    "max only - value below",
			value:   5.0,
			min:     nil,
			max:     floatPtr(10.0),
			wantAnom: false,
		},
		{
			name:    "boundary value equals min",
			value:   10.0,
			min:     floatPtr(10.0),
			max:     floatPtr(20.0),
			wantAnom: false,
		},
		{
			name:    "boundary value equals max",
			value:   20.0,
			min:     floatPtr(10.0),
			max:     floatPtr(20.0),
			wantAnom: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.detectThreshold(tt.value, tt.min, tt.max)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAnom, result.IsAnomaly)
		})
	}
}

func TestAnomalyService_CalculateBaselineStats(t *testing.T) {
	svc := newTestAnomalyService()

	t.Run("empty array", func(t *testing.T) {
		stats := svc.CalculateBaselineStats([]float64{})
		assert.Equal(t, 0.0, stats.Mean)
		assert.Equal(t, 0.0, stats.StdDev)
		assert.Equal(t, 0.0, stats.Median)
	})

	t.Run("single element", func(t *testing.T) {
		stats := svc.CalculateBaselineStats([]float64{42.0})
		assert.Equal(t, 42.0, stats.Mean)
		assert.Equal(t, 0.0, stats.StdDev)
		assert.Equal(t, 42.0, stats.Median)
		assert.Equal(t, 42.0, stats.Min)
		assert.Equal(t, 42.0, stats.Max)
	})

	t.Run("known stats", func(t *testing.T) {
		data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
		stats := svc.CalculateBaselineStats(data)
		assert.Equal(t, 5.0, stats.Mean)
		assert.Equal(t, 2.0, stats.Min)
		assert.Equal(t, 9.0, stats.Max)
		// Median of [2,4,4,4,5,5,7,9] = (4+5)/2 = 4.5
		assert.Equal(t, 4.5, stats.Median)
		// Verify StdDev is reasonable (population StdDev)
		assert.True(t, stats.StdDev > 0)
		assert.InDelta(t, 2.0, stats.StdDev, 0.1)
	})
}

func TestAnomalyService_DetermineSeverity(t *testing.T) {
	svc := newTestAnomalyService()

	tests := []struct {
		name      string
		score     float64
		threshold float64
		expected  domain.AnomalySeverity
	}{
		{
			name:      "ratio <= 1.5 is Low",
			score:     1.4,
			threshold: 1.0,
			expected:  domain.AnomalySeverityLow,
		},
		{
			name:      "ratio > 1.5 is Medium",
			score:     1.6,
			threshold: 1.0,
			expected:  domain.AnomalySeverityMedium,
		},
		{
			name:      "ratio > 2.0 is High",
			score:     2.1,
			threshold: 1.0,
			expected:  domain.AnomalySeverityHigh,
		},
		{
			name:      "ratio > 3.0 is Critical",
			score:     3.1,
			threshold: 1.0,
			expected:  domain.AnomalySeverityCritical,
		},
		{
			name:      "exact boundary 1.5 is Low",
			score:     1.5,
			threshold: 1.0,
			expected:  domain.AnomalySeverityLow,
		},
		{
			name:      "exact boundary 2.0 is Medium",
			score:     2.0,
			threshold: 1.0,
			expected:  domain.AnomalySeverityMedium,
		},
		{
			name:      "exact boundary 3.0 is High",
			score:     3.0,
			threshold: 1.0,
			expected:  domain.AnomalySeverityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a dummy type; determineSeverity uses score/threshold ratio
			result := svc.determineSeverity(tt.score, tt.threshold, domain.AnomalyTypeLatency)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnomalyService_Percentile(t *testing.T) {
	svc := newTestAnomalyService()

	t.Run("empty data returns zero", func(t *testing.T) {
		assert.Equal(t, 0.0, svc.percentile([]float64{}, 50))
	})

	t.Run("single element returns that element", func(t *testing.T) {
		assert.Equal(t, 7.0, svc.percentile([]float64{7.0}, 50))
	})

	t.Run("sorted data percentiles", func(t *testing.T) {
		data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		p50 := svc.percentile(data, 50)
		// Should be around 5.5
		assert.InDelta(t, 5.5, p50, 0.5)
		// P99 should be close to 10
		p99 := svc.percentile(data, 99)
		assert.True(t, p99 > 9.0)
	})
}

// Verify NaN/Inf handling doesn't break stats
func TestAnomalyService_BaselineStatsLargeValues(t *testing.T) {
	svc := newTestAnomalyService()
	data := []float64{1e15, 1e15, 1e15}
	stats := svc.CalculateBaselineStats(data)
	assert.False(t, math.IsNaN(stats.Mean))
	assert.False(t, math.IsInf(stats.Mean, 0))
	assert.Equal(t, 1e15, stats.Mean)
}
