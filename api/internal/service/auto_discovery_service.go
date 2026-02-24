package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AutoDiscoveryService handles auto-discovery of AI frameworks in projects
type AutoDiscoveryService struct {
	logger *zap.Logger
}

// NewAutoDiscoveryService creates a new auto-discovery service
func NewAutoDiscoveryService(logger *zap.Logger) *AutoDiscoveryService {
	return &AutoDiscoveryService{
		logger: logger,
	}
}

// ScanProject scans a project and returns discovered frameworks with mock detection results
func (s *AutoDiscoveryService) ScanProject(ctx context.Context, projectID uuid.UUID) (*domain.DiscoveryDashboard, error) {
	s.logger.Info("scanning project for AI frameworks", zap.String("projectId", projectID.String()))

	now := time.Now()

	langchainComponents := []domain.DiscoveredComponent{
		{
			Name:         "ChatOpenAI",
			Type:         "llm",
			CallCount:    1247,
			AvgLatencyMs: 850.5,
			FirstSeen:    now.Add(-72 * time.Hour),
			LastSeen:     now.Add(-5 * time.Minute),
		},
		{
			Name:         "RetrievalQA",
			Type:         "chain",
			CallCount:    523,
			AvgLatencyMs: 1200.3,
			FirstSeen:    now.Add(-48 * time.Hour),
			LastSeen:     now.Add(-10 * time.Minute),
		},
		{
			Name:         "ChromaDB",
			Type:         "vectorstore",
			CallCount:    891,
			AvgLatencyMs: 45.2,
			FirstSeen:    now.Add(-72 * time.Hour),
			LastSeen:     now.Add(-3 * time.Minute),
		},
	}

	openaiComponents := []domain.DiscoveredComponent{
		{
			Name:         "ChatCompletion",
			Type:         "llm",
			CallCount:    3421,
			AvgLatencyMs: 920.1,
			FirstSeen:    now.Add(-96 * time.Hour),
			LastSeen:     now.Add(-1 * time.Minute),
		},
		{
			Name:         "Embedding",
			Type:         "embedding",
			CallCount:    756,
			AvgLatencyMs: 120.4,
			FirstSeen:    now.Add(-96 * time.Hour),
			LastSeen:     now.Add(-15 * time.Minute),
		},
	}

	frameworks := []domain.DiscoveredFramework{
		{
			ID:               uuid.New(),
			ProjectID:        projectID,
			Framework:        domain.FrameworkTypeLangChain,
			Version:          "0.1.16",
			Status:           domain.DiscoveryStatusDetected,
			DetectedAt:       now,
			Components:       langchainComponents,
			AutoInstrumented: false,
			Config: domain.DiscoveryConfig{
				Enabled:      true,
				SamplingRate: 1.0,
				MaxDepth:     5,
			},
		},
		{
			ID:               uuid.New(),
			ProjectID:        projectID,
			Framework:        domain.FrameworkTypeOpenAI,
			Version:          "1.30.1",
			Status:           domain.DiscoveryStatusInstrumented,
			DetectedAt:       now.Add(-24 * time.Hour),
			Components:       openaiComponents,
			AutoInstrumented: true,
			Config: domain.DiscoveryConfig{
				Enabled:      true,
				SamplingRate: 1.0,
				MaxDepth:     3,
			},
		},
	}

	totalComponents := len(langchainComponents) + len(openaiComponents)
	instrumentedCount := len(openaiComponents) // only OpenAI is instrumented

	dashboard := &domain.DiscoveryDashboard{
		Frameworks:             frameworks,
		TotalComponents:        totalComponents,
		InstrumentedComponents: instrumentedCount,
		LastScanAt:             &now,
	}

	s.logger.Info("project scan completed",
		zap.String("projectId", projectID.String()),
		zap.Int("frameworksFound", len(frameworks)),
		zap.Int("totalComponents", totalComponents),
	)
	return dashboard, nil
}

// GetFramework retrieves a discovered framework by ID
func (s *AutoDiscoveryService) GetFramework(ctx context.Context, id uuid.UUID) (*domain.DiscoveredFramework, error) {
	s.logger.Debug("fetching framework", zap.String("id", id.String()))

	now := time.Now()
	return &domain.DiscoveredFramework{
		ID:               id,
		Framework:        domain.FrameworkTypeLangChain,
		Version:          "0.1.16",
		Status:           domain.DiscoveryStatusDetected,
		DetectedAt:       now,
		Components:       []domain.DiscoveredComponent{},
		AutoInstrumented: false,
		Config: domain.DiscoveryConfig{
			Enabled:      true,
			SamplingRate: 1.0,
			MaxDepth:     5,
		},
	}, nil
}

// UpdateConfig updates the discovery configuration for a project
func (s *AutoDiscoveryService) UpdateConfig(ctx context.Context, projectID uuid.UUID, config domain.DiscoveryConfig) error {
	if config.SamplingRate < 0 || config.SamplingRate > 1 {
		return fmt.Errorf("sampling rate must be between 0 and 1, got %f", config.SamplingRate)
	}
	if config.MaxDepth < 1 {
		return fmt.Errorf("max depth must be at least 1")
	}

	s.logger.Info("discovery config updated",
		zap.String("projectId", projectID.String()),
		zap.Bool("enabled", config.Enabled),
		zap.Float64("samplingRate", config.SamplingRate),
		zap.Int("maxDepth", config.MaxDepth),
	)
	return nil
}

// ToggleInstrumentation enables or disables auto-instrumentation for a framework
func (s *AutoDiscoveryService) ToggleInstrumentation(ctx context.Context, frameworkID uuid.UUID, enabled bool) error {
	s.logger.Info("instrumentation toggled",
		zap.String("frameworkId", frameworkID.String()),
		zap.Bool("enabled", enabled),
	)
	return nil
}
