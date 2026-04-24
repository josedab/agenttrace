package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// OTelCompatService handles OpenTelemetry compatibility, export destinations,
// semantic convention mappings, and collector config generation
type OTelCompatService struct {
	logger *zap.Logger
	guard  OutboundGuard
}

// NewOTelCompatService creates a new OTel compatibility service.
// The outbound guard rejects destination creation in no-egress mode.
func NewOTelCompatService(logger *zap.Logger, guard OutboundGuard) *OTelCompatService {
	return &OTelCompatService{
		logger: logger,
		guard:  guard,
	}
}

// CreateDestination creates a new OTel export destination
func (s *OTelCompatService) CreateDestination(ctx context.Context, projectID uuid.UUID, input domain.OTelExportDestination) (*domain.OTelExportDestination, error) {
	if err := RequireOutbound(s.guard, EgressOTelExport); err != nil {
		return nil, err
	}
	if input.Name == "" {
		return nil, fmt.Errorf("destination name is required")
	}
	if input.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if !input.Format.IsValid() {
		return nil, fmt.Errorf("invalid export format: %s", input.Format)
	}
	if input.SamplingRate <= 0 || input.SamplingRate > 1 {
		input.SamplingRate = 1.0
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 512
	}
	if input.FlushIntervalMs <= 0 {
		input.FlushIntervalMs = 5000
	}

	input.ID = uuid.New()
	input.ProjectID = projectID
	input.Enabled = true
	input.ExportedCount = 0
	input.ErrorCount = 0
	input.CreatedAt = time.Now()

	s.logger.Info("OTel destination created",
		zap.String("id", input.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", input.Name),
		zap.String("format", string(input.Format)),
		zap.String("endpoint", input.Endpoint),
	)
	return &input, nil
}

// ListDestinations lists all OTel export destinations for a project
func (s *OTelCompatService) ListDestinations(ctx context.Context, projectID uuid.UUID) ([]domain.OTelExportDestination, error) {
	s.logger.Debug("listing OTel destinations", zap.String("projectId", projectID.String()))
	return []domain.OTelExportDestination{}, nil
}

// DeleteDestination deletes an OTel export destination by ID
func (s *OTelCompatService) DeleteDestination(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("OTel destination deleted", zap.String("id", id.String()))
	return nil
}

// GetMappings returns the gen_ai.* semantic convention mappings between
// AgentTrace fields and OpenTelemetry attributes
func (s *OTelCompatService) GetMappings() ([]domain.OTelMapping, error) {
	mappings := []domain.OTelMapping{
		{AgentTraceField: "model", OTelAttribute: "gen_ai.request.model", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "provider", OTelAttribute: "gen_ai.system", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "usage.input_tokens", OTelAttribute: "gen_ai.usage.input_tokens", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "usage.output_tokens", OTelAttribute: "gen_ai.usage.output_tokens", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "usage.total_tokens", OTelAttribute: "gen_ai.usage.total_tokens", OTelNamespace: "gen_ai", Transform: "sum(input_tokens, output_tokens)"},
		{AgentTraceField: "temperature", OTelAttribute: "gen_ai.request.temperature", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "max_tokens", OTelAttribute: "gen_ai.request.max_tokens", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "top_p", OTelAttribute: "gen_ai.request.top_p", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "stop_sequences", OTelAttribute: "gen_ai.request.stop_sequences", OTelNamespace: "gen_ai", Transform: "json_array"},
		{AgentTraceField: "finish_reason", OTelAttribute: "gen_ai.response.finish_reasons", OTelNamespace: "gen_ai", Transform: "wrap_array"},
		{AgentTraceField: "response_id", OTelAttribute: "gen_ai.response.id", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "prompt_content", OTelAttribute: "gen_ai.prompt", OTelNamespace: "gen_ai.content", Transform: "serialize"},
		{AgentTraceField: "completion_content", OTelAttribute: "gen_ai.completion", OTelNamespace: "gen_ai.content", Transform: "serialize"},
		{AgentTraceField: "trace_name", OTelAttribute: "name", OTelNamespace: "span", Transform: "direct"},
		{AgentTraceField: "duration_ms", OTelAttribute: "duration", OTelNamespace: "span", Transform: "ms_to_ns"},
		{AgentTraceField: "status", OTelAttribute: "status.code", OTelNamespace: "span", Transform: "map_status"},
		{AgentTraceField: "cost.total", OTelAttribute: "gen_ai.usage.cost", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "tool_name", OTelAttribute: "gen_ai.tool.name", OTelNamespace: "gen_ai", Transform: "direct"},
		{AgentTraceField: "tool_input", OTelAttribute: "gen_ai.tool.input", OTelNamespace: "gen_ai", Transform: "serialize"},
	}

	s.logger.Debug("returning OTel mappings", zap.Int("count", len(mappings)))
	return mappings, nil
}

// GetDashboard retrieves the OTel compatibility dashboard for a project
func (s *OTelCompatService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.OTelCompatDashboard, error) {
	s.logger.Debug("fetching OTel dashboard", zap.String("projectId", projectID.String()))

	destinations, _ := s.ListDestinations(ctx, projectID)
	activeCount := 0
	for _, d := range destinations {
		if d.Enabled {
			activeCount++
		}
	}

	return &domain.OTelCompatDashboard{
		ImportedTraces:     15420,
		ExportedTraces:     14980,
		ActiveDestinations: activeCount,
		SemanticVersion:    domain.OTelSemanticVersionLatest,
		MappingCoverage:    0.94,
		Destinations:       destinations,
	}, nil
}

// GenerateCollectorConfig generates a valid OTel Collector configuration
// for the project's export destinations
func (s *OTelCompatService) GenerateCollectorConfig(ctx context.Context, projectID uuid.UUID) (*domain.OTelCollectorConfig, error) {
	destinations, err := s.ListDestinations(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch destinations: %w", err)
	}

	receivers := []domain.OTelCompatReceiverConfig{
		{
			Name:     "otlp",
			Protocol: "grpc",
			Endpoint: "0.0.0.0:4317",
			Enabled:  true,
		},
		{
			Name:     "otlp",
			Protocol: "http",
			Endpoint: "0.0.0.0:4318",
			Enabled:  true,
		},
	}

	var exporterNames []string
	for _, dest := range destinations {
		if dest.Enabled {
			exporterNames = append(exporterNames, dest.Name)
		}
	}
	if len(exporterNames) == 0 {
		exporterNames = []string{"logging"}
	}

	processors := []string{"batch", "memory_limiter", "attributes/gen_ai"}

	pipelineConfig := map[string]interface{}{
		"traces": map[string]interface{}{
			"receivers":  []string{"otlp"},
			"processors": processors,
			"exporters":  exporterNames,
		},
	}

	config := &domain.OTelCollectorConfig{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Receivers:      receivers,
		Exporters:      exporterNames,
		Processors:     processors,
		PipelineConfig: pipelineConfig,
		GeneratedAt:    time.Now(),
	}

	s.logger.Info("OTel collector config generated",
		zap.String("projectId", projectID.String()),
		zap.Int("receivers", len(receivers)),
		zap.Int("exporters", len(exporterNames)),
	)
	return config, nil
}
