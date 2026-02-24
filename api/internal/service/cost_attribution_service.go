package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type CostAttributionService struct {
	logger       *zap.Logger
	mu           sync.RWMutex
	attributions map[string][]domain.CostAttribution // projectID -> attributions
}

func NewCostAttributionService(logger *zap.Logger) *CostAttributionService {
	return &CostAttributionService{
		logger:       logger,
		attributions: make(map[string][]domain.CostAttribution),
	}
}

func (s *CostAttributionService) Attribute(ctx context.Context, projectID uuid.UUID, input domain.AttributionInput) (*domain.CostAttribution, error) {
	valueSaved := input.HoursSaved * input.HourlyRate
	costIncurred := input.HoursSaved * 0.15 * input.HourlyRate // assume AI cost is ~15% of manual
	roi := 0.0
	if costIncurred > 0 {
		roi = (valueSaved - costIncurred) / costIncurred * 100
	}

	attr := domain.CostAttribution{
		ID:                  uuid.New(),
		ProjectID:           projectID,
		TraceID:             input.TraceID,
		IssueRef:            input.IssueRef,
		IssueTitle:          input.IssueTitle,
		CostIncurred:        costIncurred,
		EstimatedValueSaved: valueSaved,
		HoursSaved:          input.HoursSaved,
		HourlyRate:          input.HourlyRate,
		ROI:                 roi,
		Category:            input.Category,
		CreatedAt:           time.Now(),
	}

	s.mu.Lock()
	s.attributions[projectID.String()] = append(s.attributions[projectID.String()], attr)
	s.mu.Unlock()

	s.logger.Info("cost attribution recorded", zap.String("projectId", projectID.String()), zap.String("issue", input.IssueRef))
	return &attr, nil
}

func (s *CostAttributionService) GetReport(ctx context.Context, projectID uuid.UUID, period domain.AttributionDateRange) (*domain.AttributionReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attrs := s.attributions[projectID.String()]

	report := &domain.AttributionReport{
		ProjectID:    projectID,
		Period:       period,
		Attributions: attrs,
		ByCategory:   make(map[string]domain.CategoryROI),
	}

	if attrs == nil {
		report.Attributions = []domain.CostAttribution{}
	}

	for _, a := range attrs {
		report.TotalCost += a.CostIncurred
		report.TotalValueSaved += a.EstimatedValueSaved

		cat := report.ByCategory[a.Category]
		cat.Category = a.Category
		cat.Cost += a.CostIncurred
		cat.Value += a.EstimatedValueSaved
		cat.TraceCount++
		report.ByCategory[a.Category] = cat
	}

	if report.TotalCost > 0 {
		report.OverallROI = (report.TotalValueSaved - report.TotalCost) / report.TotalCost * 100
	}

	// Compute per-category ROI
	for k, cat := range report.ByCategory {
		if cat.Cost > 0 {
			cat.ROI = (cat.Value - cat.Cost) / cat.Cost * 100
		}
		report.ByCategory[k] = cat
	}

	return report, nil
}

func (s *CostAttributionService) ListAttributions(ctx context.Context, projectID uuid.UUID) ([]domain.CostAttribution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attrs := s.attributions[projectID.String()]
	if attrs == nil {
		return []domain.CostAttribution{}, nil
	}
	return attrs, nil
}
