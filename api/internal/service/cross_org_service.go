package service

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

var mockIndustryData = map[string]map[string]struct{ avg, p50, p90 float64 }{
	"performance": {
		"latency_p99_ms":     {avg: 450, p50: 380, p90: 820},
		"throughput_rps":     {avg: 1200, p50: 1000, p90: 3500},
		"error_rate_percent": {avg: 2.1, p50: 1.5, p90: 5.0},
	},
	"cost": {
		"cost_per_trace_usd": {avg: 0.015, p50: 0.012, p90: 0.035},
		"monthly_spend_usd":  {avg: 2500, p50: 1800, p90: 8000},
	},
	"quality": {
		"test_coverage_percent": {avg: 68, p50: 65, p90: 92},
		"bug_escape_rate":       {avg: 3.2, p50: 2.8, p90: 7.5},
	},
}

type CrossOrgService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	submissions map[string][]domain.CrossOrgSubmission // projectID -> submissions
}

func NewCrossOrgService(logger *zap.Logger) *CrossOrgService {
	return &CrossOrgService{
		logger:      logger,
		submissions: make(map[string][]domain.CrossOrgSubmission),
	}
}

func (s *CrossOrgService) Submit(ctx context.Context, projectID uuid.UUID, input domain.CrossOrgSubmissionInput) (*domain.CrossOrgSubmission, error) {
	sub := domain.CrossOrgSubmission{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Category:    input.Category,
		Metrics:     input.Metrics,
		Anonymized:  true,
		SubmittedAt: time.Now(),
	}

	s.mu.Lock()
	s.submissions[projectID.String()] = append(s.submissions[projectID.String()], sub)
	s.mu.Unlock()

	s.logger.Info("cross-org submission received", zap.String("projectId", projectID.String()), zap.String("category", input.Category))
	return &sub, nil
}

func (s *CrossOrgService) GetReport(ctx context.Context, projectID uuid.UUID) (*domain.CrossOrgReport, error) {
	var benchmarks []domain.CrossOrgBenchmark
	var strongAreas, weakAreas []string
	var totalPercentile float64
	count := 0

	for category, metrics := range mockIndustryData {
		for metricName, data := range metrics {
			percentile := 40 + rand.Float64()*50 // 40-90 range
			b := domain.CrossOrgBenchmark{
				ID:                uuid.New(),
				Category:          category,
				MetricName:        metricName,
				Percentile:        math.Round(percentile*10) / 10,
				IndustryAvg:       data.avg,
				IndustryP50:       data.p50,
				IndustryP90:       data.p90,
				AnonymousRank:     int(math.Round((100-percentile)/100*150)) + 1,
				TotalParticipants: 150,
				UpdatedAt:         time.Now(),
			}
			benchmarks = append(benchmarks, b)
			totalPercentile += percentile
			count++

			if percentile >= 75 {
				strongAreas = append(strongAreas, category+"/"+metricName)
			} else if percentile < 40 {
				weakAreas = append(weakAreas, category+"/"+metricName)
			}
		}
	}

	overall := 0.0
	if count > 0 {
		overall = math.Round(totalPercentile/float64(count)*10) / 10
	}

	return &domain.CrossOrgReport{
		ProjectID:         projectID,
		Benchmarks:        benchmarks,
		OverallPercentile: overall,
		StrongAreas:       strongAreas,
		WeakAreas:         weakAreas,
	}, nil
}

func (s *CrossOrgService) GetIndustryStats(ctx context.Context, category string) (map[string]struct{ Avg, P50, P90 float64 }, error) {
	result := make(map[string]struct{ Avg, P50, P90 float64 })
	if data, ok := mockIndustryData[category]; ok {
		for metric, vals := range data {
			result[metric] = struct{ Avg, P50, P90 float64 }{Avg: vals.avg, P50: vals.p50, P90: vals.p90}
		}
	}
	return result, nil
}
