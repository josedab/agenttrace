package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// StreamingDashboardService handles real-time streaming dashboard operations
type StreamingDashboardService struct {
	logger    *zap.Logger
	streaming *StreamingService
	configs   map[uuid.UUID]*domain.DashboardConfig
}

// NewStreamingDashboardService creates a new streaming dashboard service
func NewStreamingDashboardService(logger *zap.Logger, streaming *StreamingService) *StreamingDashboardService {
	return &StreamingDashboardService{
		logger:    logger,
		streaming: streaming,
		configs:   make(map[uuid.UUID]*domain.DashboardConfig),
	}
}

// GetDashboard returns a live dashboard snapshot using streaming service data
func (s *StreamingDashboardService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.DashboardSnapshot, error) {
	s.logger.Info("fetching dashboard snapshot",
		zap.String("projectId", projectID.String()),
	)

	activeStreams := s.streaming.GetActiveStreams(projectID)

	var totalCost float64
	var totalTokens int64
	var errorCount int

	for _, stream := range activeStreams {
		totalCost += stream.TotalCost
		totalTokens += int64(stream.TotalTokens)
		if stream.ErrorCount > 0 {
			errorCount++
		}
	}

	topModels := make([]domain.ModelUsage, 0)

	snapshot := &domain.DashboardSnapshot{
		ProjectID:      projectID,
		ActiveSessions: len(activeStreams),
		TotalCost:      totalCost,
		TotalTokens:    totalTokens,
		ErrorCount:     errorCount,
		ActiveStreams:   activeStreams,
		TopModels:      topModels,
		Timestamp:      time.Now(),
	}

	s.logger.Info("dashboard snapshot built",
		zap.String("projectId", projectID.String()),
		zap.Int("activeSessions", snapshot.ActiveSessions),
	)

	return snapshot, nil
}

// GetConfig returns the dashboard configuration for a project, creating a default if none exists
func (s *StreamingDashboardService) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.DashboardConfig, error) {
	s.logger.Info("fetching dashboard config",
		zap.String("projectId", projectID.String()),
	)

	if config, ok := s.configs[projectID]; ok {
		return config, nil
	}

	now := time.Now()
	config := &domain.DashboardConfig{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      "Default Dashboard",
		Layout: []domain.StreamingDashboardWidget{
			{
				ID:   "widget-active-sessions",
				Type: "active_sessions",
				Position: domain.WidgetPosition{
					X: 0, Y: 0, Width: 4, Height: 2,
				},
				Config: map[string]any{"showCount": true},
			},
			{
				ID:   "widget-cost-ticker",
				Type: "cost_ticker",
				Position: domain.WidgetPosition{
					X: 4, Y: 0, Width: 4, Height: 2,
				},
				Config: map[string]any{"currency": "USD"},
			},
			{
				ID:   "widget-token-stream",
				Type: "token_stream",
				Position: domain.WidgetPosition{
					X: 8, Y: 0, Width: 4, Height: 2,
				},
				Config: map[string]any{"showRate": true},
			},
			{
				ID:   "widget-error-feed",
				Type: "error_feed",
				Position: domain.WidgetPosition{
					X: 0, Y: 2, Width: 6, Height: 3,
				},
				Config: map[string]any{"maxItems": 50},
			},
			{
				ID:   "widget-progress-bar",
				Type: "progress_bar",
				Position: domain.WidgetPosition{
					X: 6, Y: 2, Width: 6, Height: 3,
				},
				Config: map[string]any{"showPercentage": true},
			},
		},
		RefreshInterval: 5,
		IsDefault:       true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	s.configs[projectID] = config

	return config, nil
}

// UpdateConfig updates the dashboard configuration for a project
func (s *StreamingDashboardService) UpdateConfig(ctx context.Context, projectID uuid.UUID, config *domain.DashboardConfig) error {
	s.logger.Info("updating dashboard config",
		zap.String("projectId", projectID.String()),
		zap.String("configId", config.ID.String()),
	)

	config.ProjectID = projectID
	config.UpdatedAt = time.Now()
	s.configs[projectID] = config

	return nil
}
