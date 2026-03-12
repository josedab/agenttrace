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

// CustomMetricsService manages user-defined custom metrics
type CustomMetricsService struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	metrics    map[uuid.UUID]*domain.CustomMetric
	dashboards map[uuid.UUID]*domain.MetricDashboard
	alerts     map[uuid.UUID]*domain.MetricAlert
}

// NewCustomMetricsService creates a new custom metrics service
func NewCustomMetricsService(logger *zap.Logger) *CustomMetricsService {
	return &CustomMetricsService{
		logger:     logger,
		metrics:    make(map[uuid.UUID]*domain.CustomMetric),
		dashboards: make(map[uuid.UUID]*domain.MetricDashboard),
		alerts:     make(map[uuid.UUID]*domain.MetricAlert),
	}
}

// CreateMetric creates a new custom metric
func (s *CustomMetricsService) CreateMetric(ctx context.Context, projectID uuid.UUID, input *domain.CustomMetricInput) (*domain.CustomMetric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric := &domain.CustomMetric{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		Description:     input.Description,
		Query:           input.Query,
		Unit:            input.Unit,
		Aggregation:     input.Aggregation,
		RefreshInterval: input.RefreshInterval,
		Enabled:         true,
		CreatedAt:       time.Now(),
	}

	s.metrics[metric.ID] = metric
	s.logger.Info("created custom metric", zap.String("id", metric.ID.String()), zap.String("name", input.Name))
	return metric, nil
}

// ListMetrics returns all custom metrics for a project
func (s *CustomMetricsService) ListMetrics(ctx context.Context, projectID uuid.UUID) ([]*domain.CustomMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.CustomMetric
	for _, m := range s.metrics {
		if m.ProjectID == projectID {
			result = append(result, m)
		}
	}

	if result == nil {
		result = []*domain.CustomMetric{}
	}
	return result, nil
}

// GetMetricValues returns mock time series data for a metric
func (s *CustomMetricsService) GetMetricValues(ctx context.Context, metricID uuid.UUID, from time.Time, to time.Time) ([]domain.CustomMetricValue, error) {
	s.logger.Debug("getting metric values", zap.String("metricId", metricID.String()))

	duration := to.Sub(from)
	points := 24
	interval := duration / time.Duration(points)

	values := make([]domain.CustomMetricValue, points)
	rng := rand.New(rand.NewSource(int64(metricID.ID())))
	baseValue := 50.0 + rng.Float64()*50.0

	for i := 0; i < points; i++ {
		values[i] = domain.CustomMetricValue{
			MetricID:  metricID,
			Timestamp: from.Add(interval * time.Duration(i)),
			Value:     math.Round((baseValue+rng.Float64()*20.0-10.0)*100) / 100,
		}
	}

	return values, nil
}

// CreateDashboard creates a new metric dashboard
func (s *CustomMetricsService) CreateDashboard(ctx context.Context, projectID uuid.UUID, name string, widgets []domain.DashboardWidget) (*domain.MetricDashboard, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range widgets {
		if widgets[i].ID == uuid.Nil {
			widgets[i].ID = uuid.New()
		}
	}

	dashboard := &domain.MetricDashboard{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
		Widgets:   widgets,
		CreatedAt: time.Now(),
	}

	s.dashboards[dashboard.ID] = dashboard
	s.logger.Info("created dashboard", zap.String("id", dashboard.ID.String()), zap.String("name", name))
	return dashboard, nil
}

// ListDashboards returns all dashboards for a project
func (s *CustomMetricsService) ListDashboards(ctx context.Context, projectID uuid.UUID) ([]*domain.MetricDashboard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.MetricDashboard
	for _, d := range s.dashboards {
		if d.ProjectID == projectID {
			result = append(result, d)
		}
	}

	if result == nil {
		result = []*domain.MetricDashboard{}
	}
	return result, nil
}

// CreateAlert creates a new metric alert
func (s *CustomMetricsService) CreateAlert(ctx context.Context, projectID uuid.UUID, input *domain.MetricAlertInput) (*domain.MetricAlert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert := &domain.MetricAlert{
		ID:        uuid.New(),
		MetricID:  input.MetricID,
		Condition: input.Condition,
		Threshold: input.Threshold,
		Channel:   input.Channel,
		Enabled:   true,
	}

	s.alerts[alert.ID] = alert
	s.logger.Info("created metric alert", zap.String("id", alert.ID.String()))
	return alert, nil
}

// ListAlerts returns all alerts for a project
func (s *CustomMetricsService) ListAlerts(ctx context.Context, projectID uuid.UUID) ([]*domain.MetricAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.MetricAlert
	for _, a := range s.alerts {
		result = append(result, a)
	}

	if result == nil {
		result = []*domain.MetricAlert{}
	}
	return result, nil
}
