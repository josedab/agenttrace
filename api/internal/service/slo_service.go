package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// SLOService manages agent performance SLOs
type SLOService struct {
	logger *zap.Logger
	mu     sync.RWMutex
	slos   map[string]*domain.SLO
}

// NewSLOService creates a new SLO service
func NewSLOService(logger *zap.Logger) *SLOService {
	return &SLOService{
		logger: logger,
		slos:   make(map[string]*domain.SLO),
	}
}

// CreateSLO creates a new SLO
func (s *SLOService) CreateSLO(ctx context.Context, projectID string, input *domain.SLOInput) (*domain.SLO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slo := &domain.SLO{
		ID:        fmt.Sprintf("slo_%d", time.Now().UnixNano()),
		ProjectID: projectID,
		AgentName: input.AgentName,
		Name:      input.Name,
		Metric:    input.Metric,
		Target:    input.Target,
		Window:    input.Window,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	s.slos[slo.ID] = slo
	s.logger.Info("created SLO",
		zap.String("id", slo.ID),
		zap.String("name", slo.Name),
		zap.String("metric", slo.Metric),
	)
	return slo, nil
}

// ListSLOs returns all SLOs for a project
func (s *SLOService) ListSLOs(ctx context.Context, projectID string) ([]domain.SLO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.SLO
	for _, slo := range s.slos {
		if slo.ProjectID == projectID || projectID == "" {
			result = append(result, *slo)
		}
	}
	if result == nil {
		result = []domain.SLO{}
	}
	return result, nil
}

// GetSLOStatus computes the current compliance status of an SLO
func (s *SLOService) GetSLOStatus(ctx context.Context, sloID string) (*domain.SLOStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	slo, ok := s.slos[sloID]
	if !ok {
		return nil, fmt.Errorf("SLO not found: %s", sloID)
	}

	// Simulate current value based on metric type
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var currentValue float64
	switch slo.Metric {
	case "latency_p99":
		currentValue = slo.Target * (0.7 + r.Float64()*0.5) // 70-120% of target
	case "success_rate":
		currentValue = math.Min(100, slo.Target*(0.95+r.Float64()*0.08)) // 95-103% of target
	case "cost_per_trace":
		currentValue = slo.Target * (0.6 + r.Float64()*0.6) // 60-120% of target
	case "uptime":
		currentValue = math.Min(100, slo.Target*(0.98+r.Float64()*0.03)) // 98-101% of target
	default:
		currentValue = slo.Target * (0.8 + r.Float64()*0.4)
	}

	// For latency and cost, lower is better; for success_rate and uptime, higher is better
	compliant := true
	switch slo.Metric {
	case "latency_p99", "cost_per_trace":
		compliant = currentValue <= slo.Target
	default:
		compliant = currentValue >= slo.Target
	}

	// Error budget: how much of the allowed error is remaining
	errorBudgetRemaining := 100.0
	burnRate := 0.0
	violationCount := 0

	if !compliant {
		var deviation float64
		switch slo.Metric {
		case "latency_p99", "cost_per_trace":
			deviation = (currentValue - slo.Target) / slo.Target
		default:
			deviation = (slo.Target - currentValue) / slo.Target
		}
		errorBudgetRemaining = math.Max(0, 100.0-deviation*100.0*10)
		burnRate = deviation * 10.0
		violationCount = 1 + r.Intn(5)
	} else {
		errorBudgetRemaining = 50.0 + r.Float64()*50.0
		burnRate = r.Float64() * 0.5
	}

	return &domain.SLOStatus{
		SLOID:                sloID,
		CurrentValue:         math.Round(currentValue*100) / 100,
		Target:               slo.Target,
		Compliant:            compliant,
		ErrorBudgetRemaining: math.Round(errorBudgetRemaining*100) / 100,
		BurnRate:             math.Round(burnRate*100) / 100,
		ViolationCount:       violationCount,
		LastChecked:          time.Now(),
	}, nil
}

// GetReport generates an SLO compliance report for a project
func (s *SLOService) GetReport(ctx context.Context, projectID string) (*domain.SLOReport, error) {
	slos, err := s.ListSLOs(ctx, projectID)
	if err != nil {
		return nil, err
	}

	report := &domain.SLOReport{
		ProjectID: projectID,
		SLOs:      []domain.SLOStatus{},
		AtRisk:    []string{},
	}

	compliantCount := 0
	for _, slo := range slos {
		status, err := s.GetSLOStatus(ctx, slo.ID)
		if err != nil {
			continue
		}
		report.SLOs = append(report.SLOs, *status)
		if status.Compliant {
			compliantCount++
		}
		if status.ErrorBudgetRemaining < 30 {
			report.AtRisk = append(report.AtRisk, slo.Name)
		}
	}

	if len(slos) > 0 {
		report.OverallCompliance = math.Round(float64(compliantCount)/float64(len(slos))*10000) / 100
	}

	return report, nil
}

// GetHistory returns mock time series data for an SLO
func (s *SLOService) GetHistory(ctx context.Context, sloID string, from, to time.Time) ([]domain.SLOHistoryPoint, error) {
	s.mu.RLock()
	slo, ok := s.slos[sloID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("SLO not found: %s", sloID)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var points []domain.SLOHistoryPoint

	// Generate hourly data points
	interval := time.Hour
	current := from
	for current.Before(to) {
		var value float64
		switch slo.Metric {
		case "latency_p99":
			value = slo.Target * (0.6 + r.Float64()*0.6)
		case "success_rate":
			value = math.Min(100, slo.Target*(0.93+r.Float64()*0.1))
		case "cost_per_trace":
			value = slo.Target * (0.5 + r.Float64()*0.7)
		case "uptime":
			value = math.Min(100, slo.Target*(0.97+r.Float64()*0.04))
		default:
			value = slo.Target * (0.8 + r.Float64()*0.4)
		}

		compliant := true
		switch slo.Metric {
		case "latency_p99", "cost_per_trace":
			compliant = value <= slo.Target
		default:
			compliant = value >= slo.Target
		}

		points = append(points, domain.SLOHistoryPoint{
			Timestamp: current,
			Value:     math.Round(value*100) / 100,
			Compliant: compliant,
		})

		current = current.Add(interval)
	}

	return points, nil
}
