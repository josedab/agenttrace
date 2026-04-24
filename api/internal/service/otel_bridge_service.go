package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// OTelBridgeService manages the OpenTelemetry-native trace bridge
type OTelBridgeService struct {
	logger    *zap.Logger
	ingestion *IngestionService
	guard     OutboundGuard
	mu        sync.RWMutex
	configs   map[uuid.UUID]*domain.OTelBridgeConfig
	dests     map[uuid.UUID]*domain.ExportDestinationRef
	stats     map[uuid.UUID]*domain.OTelBridgeStats
}

// NewOTelBridgeService creates a new OTel bridge service.
// The outbound guard rejects export destination creation in no-egress mode.
func NewOTelBridgeService(
	logger *zap.Logger,
	ingestion *IngestionService,
	guard OutboundGuard,
) *OTelBridgeService {
	return &OTelBridgeService{
		logger:    logger,
		ingestion: ingestion,
		guard:     guard,
		configs:   make(map[uuid.UUID]*domain.OTelBridgeConfig),
		dests:     make(map[uuid.UUID]*domain.ExportDestinationRef),
		stats:     make(map[uuid.UUID]*domain.OTelBridgeStats),
	}
}

// GetConfig returns the bridge configuration for a project, creating a default if none exists
func (s *OTelBridgeService) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.OTelBridgeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		now := time.Now()
		cfg = &domain.OTelBridgeConfig{
			ID:                 uuid.New(),
			ProjectID:          projectID,
			ExportEnabled:      false,
			ImportEnabled:      true,
			ExportDestinations: []domain.ExportDestinationRef{},
			ImportMappings:     []domain.ImportMapping{},
			ResourceAttributes: map[string]string{
				"service.name": "agenttrace",
			},
			SamplingRate: 1.0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		s.configs[projectID] = cfg
		s.logger.Info("created default OTel bridge config",
			zap.String("projectId", projectID.String()),
		)
	}
	return cfg, nil
}

// UpdateConfig updates the bridge configuration for a project
func (s *OTelBridgeService) UpdateConfig(ctx context.Context, projectID uuid.UUID, input *domain.OTelBridgeConfigInput) (*domain.OTelBridgeConfig, error) {
	if input.ExportEnabled != nil && *input.ExportEnabled {
		if err := RequireOutbound(s.guard, EgressOTelExport); err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		return nil, fmt.Errorf("OTel bridge config not found for project: %s", projectID)
	}

	if input.ExportEnabled != nil {
		cfg.ExportEnabled = *input.ExportEnabled
	}
	if input.ImportEnabled != nil {
		cfg.ImportEnabled = *input.ImportEnabled
	}
	if input.ResourceAttributes != nil {
		cfg.ResourceAttributes = input.ResourceAttributes
	}
	if input.SamplingRate != nil {
		rate := *input.SamplingRate
		if rate < 0 || rate > 1 {
			return nil, fmt.Errorf("sampling rate must be between 0 and 1")
		}
		cfg.SamplingRate = rate
	}
	cfg.UpdatedAt = time.Now()

	s.logger.Info("updated OTel bridge config",
		zap.String("projectId", projectID.String()),
		zap.Bool("exportEnabled", cfg.ExportEnabled),
		zap.Bool("importEnabled", cfg.ImportEnabled),
	)
	return cfg, nil
}

// ListDestinations returns all export destinations for a project
func (s *OTelBridgeService) ListDestinations(ctx context.Context, projectID uuid.UUID) ([]domain.ExportDestinationRef, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		return []domain.ExportDestinationRef{}, nil
	}
	return cfg.ExportDestinations, nil
}

// AddDestination adds an export destination to a project's bridge config
func (s *OTelBridgeService) AddDestination(ctx context.Context, projectID uuid.UUID, input *domain.OTelDestinationInput) (*domain.ExportDestinationRef, error) {
	if err := RequireOutbound(s.guard, EgressOTelExport); err != nil {
		return nil, err
	}
	if input.Name == "" {
		return nil, fmt.Errorf("destination name is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("destination type is required")
	}
	if input.Endpoint == "" {
		return nil, fmt.Errorf("destination endpoint is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		return nil, fmt.Errorf("OTel bridge config not found for project: %s", projectID)
	}

	protocol := input.Protocol
	if protocol == "" {
		protocol = "grpc"
	}

	dest := domain.ExportDestinationRef{
		ID:       uuid.New(),
		Name:     input.Name,
		Type:     input.Type,
		Endpoint: input.Endpoint,
		Protocol: protocol,
		Enabled:  true,
		Headers:  input.Headers,
	}

	cfg.ExportDestinations = append(cfg.ExportDestinations, dest)
	s.dests[dest.ID] = &dest
	cfg.UpdatedAt = time.Now()

	s.logger.Info("added OTel export destination",
		zap.String("id", dest.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", dest.Name),
		zap.String("endpoint", dest.Endpoint),
	)
	return &dest, nil
}

// RemoveDestination removes an export destination
func (s *OTelBridgeService) RemoveDestination(
	ctx context.Context,
	projectID, destID uuid.UUID,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		return apperrors.NotFound("OTel export destination")
	}

	for index, destination := range cfg.ExportDestinations {
		if destination.ID != destID {
			continue
		}
		cfg.ExportDestinations = append(
			cfg.ExportDestinations[:index],
			cfg.ExportDestinations[index+1:]...,
		)
		cfg.UpdatedAt = time.Now()
		delete(s.dests, destID)
		s.logger.Info(
			"removed OTel export destination",
			zap.String("id", destID.String()),
			zap.String("projectId", projectID.String()),
		)
		return nil
	}
	return apperrors.NotFound("OTel export destination")
}

// ImportSpans imports OpenTelemetry spans and returns the count imported
func (s *OTelBridgeService) ImportSpans(ctx context.Context, projectID uuid.UUID, input *domain.OTelImportRequest) (int, error) {
	if len(input.ResourceSpans) == 0 {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	imported := len(input.ResourceSpans)

	// Update import stats
	st, ok := s.stats[projectID]
	if !ok {
		st = &domain.OTelBridgeStats{}
		s.stats[projectID] = st
	}
	st.ImportStats.TotalSpans += int64(imported)
	st.ImportStats.SuccessCount += int64(imported)
	st.ImportStats.Last24hCount += int64(imported)
	st.LastSync = time.Now()

	s.logger.Info("imported OTel spans",
		zap.String("projectId", projectID.String()),
		zap.Int("count", imported),
		zap.Bool("correlateByTraceId", input.CorrelateByTraceID),
	)
	return imported, nil
}

// GetStats returns bridge statistics for a project
func (s *OTelBridgeService) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.OTelBridgeStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.stats[projectID]
	if !ok {
		return &domain.OTelBridgeStats{
			ExportStats: domain.BridgeDirectionStats{},
			ImportStats: domain.BridgeDirectionStats{},
			LastSync:    time.Time{},
		}, nil
	}
	return st, nil
}

// GenerateCollectorConfig generates an OTel Collector YAML configuration
// that can be used with the standard OpenTelemetry Collector to forward
// spans to AgentTrace and other configured destinations
func (s *OTelBridgeService) GenerateCollectorConfig(ctx context.Context, projectID uuid.UUID, apiKey string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		return "", fmt.Errorf("OTel bridge config not found for project: %s", projectID)
	}

	// Build exporters section
	exporters := `  otlp/agenttrace:
    endpoint: "${AGENTTRACE_API_URL}"
    headers:
      Authorization: "Bearer ${AGENTTRACE_API_KEY}"
    tls:
      insecure: false
`

	for _, dest := range cfg.ExportDestinations {
		if !dest.Enabled {
			continue
		}
		exporterName := fmt.Sprintf("  otlp/%s", dest.Name)
		exporters += fmt.Sprintf(`%s:
    endpoint: "%s"
    protocol: %s
`, exporterName, dest.Endpoint, dest.Protocol)

		if len(dest.Headers) > 0 {
			exporters += "    headers:\n"
			for k, v := range dest.Headers {
				exporters += fmt.Sprintf("      %s: \"%s\"\n", k, v)
			}
		}
	}

	// Build pipeline destination list
	pipelineExporters := "otlp/agenttrace"
	for _, dest := range cfg.ExportDestinations {
		if dest.Enabled {
			pipelineExporters += fmt.Sprintf(", otlp/%s", dest.Name)
		}
	}

	yamlConfig := fmt.Sprintf(`# Auto-generated OpenTelemetry Collector config for AgentTrace
# Generated at: %s
# Project: %s

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 5s
    send_batch_size: 512
  resource:
    attributes:
      - key: service.name
        value: "${SERVICE_NAME:-agenttrace}"
        action: upsert

exporters:
%s
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, resource]
      exporters: [%s]
`, time.Now().Format(time.RFC3339), projectID, exporters, pipelineExporters)

	s.logger.Info("generated OTel collector config",
		zap.String("projectId", projectID.String()),
		zap.Int("destinationCount", len(cfg.ExportDestinations)),
	)

	return yamlConfig, nil
}
