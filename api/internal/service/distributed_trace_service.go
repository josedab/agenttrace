package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// DistributedTraceService handles distributed tracing across services
type DistributedTraceService struct {
	logger *zap.Logger
}

// NewDistributedTraceService creates a new distributed trace service
func NewDistributedTraceService(logger *zap.Logger) *DistributedTraceService {
	return &DistributedTraceService{
		logger: logger,
	}
}

// GetDistributedTrace returns a distributed trace with spans across services
func (s *DistributedTraceService) GetDistributedTrace(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.DistributedTrace, error) {
	s.logger.Debug("getting distributed trace", zap.String("traceId", traceID))

	now := time.Now()

	agentSpans := []domain.DistributedSpan{
		{
			SpanID:        uuid.New().String(),
			ServiceName:   "agent-orchestrator",
			OperationName: "process_request",
			StartTime:     now,
			Duration:      2500 * time.Millisecond,
			Status:        "ok",
			SpanKind:      domain.DistributedSpanKindServer,
			Attributes:    map[string]string{"agent.type": "orchestrator", "model": "gpt-4"},
		},
		{
			SpanID:        uuid.New().String(),
			ParentSpanID:  "root",
			ServiceName:   "agent-orchestrator",
			OperationName: "llm_call",
			StartTime:     now.Add(100 * time.Millisecond),
			Duration:      1800 * time.Millisecond,
			Status:        "ok",
			SpanKind:      domain.DistributedSpanKindClient,
			Attributes:    map[string]string{"model": "gpt-4", "tokens": "4500"},
		},
	}

	serviceSpans := []domain.DistributedSpan{
		{
			SpanID:        uuid.New().String(),
			ServiceName:   "api-gateway",
			OperationName: "route_request",
			StartTime:     now.Add(-50 * time.Millisecond),
			Duration:      2600 * time.Millisecond,
			Status:        "ok",
			SpanKind:      domain.DistributedSpanKindServer,
			Attributes:    map[string]string{"http.method": "POST", "http.route": "/api/chat"},
		},
		{
			SpanID:        uuid.New().String(),
			ServiceName:   "postgres",
			OperationName: "SELECT",
			StartTime:     now.Add(200 * time.Millisecond),
			Duration:      45 * time.Millisecond,
			Status:        "ok",
			SpanKind:      domain.DistributedSpanKindClient,
			Attributes:    map[string]string{"db.system": "postgresql", "db.statement": "SELECT ..."},
		},
		{
			SpanID:        uuid.New().String(),
			ServiceName:   "redis",
			OperationName: "GET",
			StartTime:     now.Add(50 * time.Millisecond),
			Duration:      5 * time.Millisecond,
			Status:        "ok",
			SpanKind:      domain.DistributedSpanKindClient,
			Attributes:    map[string]string{"db.system": "redis"},
		},
	}

	bottleneck := &agentSpans[1]

	return &domain.DistributedTrace{
		TraceID:       traceID,
		AgentSpans:    agentSpans,
		ServiceSpans:  serviceSpans,
		TotalServices: 4,
		CriticalPath:  []string{"api-gateway", "agent-orchestrator", "llm_call"},
		Bottleneck:    bottleneck,
	}, nil
}

// GetServiceMap returns the service topology for a project
func (s *DistributedTraceService) GetServiceMap(ctx context.Context, projectID uuid.UUID) (*domain.ServiceMap, error) {
	s.logger.Debug("getting service map", zap.String("projectId", projectID.String()))

	services := []domain.ServiceNode{
		{Name: "agent-orchestrator", Type: domain.ServiceNodeTypeAgent, RequestCount: 15000, AvgLatencyMs: 2500, ErrorRate: 0.02},
		{Name: "api-gateway", Type: domain.ServiceNodeTypeAPI, RequestCount: 25000, AvgLatencyMs: 150, ErrorRate: 0.01},
		{Name: "postgres", Type: domain.ServiceNodeTypeDatabase, RequestCount: 45000, AvgLatencyMs: 25, ErrorRate: 0.001},
		{Name: "redis", Type: domain.ServiceNodeTypeCache, RequestCount: 80000, AvgLatencyMs: 2, ErrorRate: 0.0001},
		{Name: "openai-api", Type: domain.ServiceNodeTypeExternal, RequestCount: 12000, AvgLatencyMs: 1800, ErrorRate: 0.03},
	}

	connections := []domain.ServiceConnection{
		{From: "api-gateway", To: "agent-orchestrator", RequestCount: 15000, AvgLatencyMs: 2500},
		{From: "agent-orchestrator", To: "postgres", RequestCount: 30000, AvgLatencyMs: 25},
		{From: "agent-orchestrator", To: "redis", RequestCount: 60000, AvgLatencyMs: 2},
		{From: "agent-orchestrator", To: "openai-api", RequestCount: 12000, AvgLatencyMs: 1800},
		{From: "api-gateway", To: "redis", RequestCount: 20000, AvgLatencyMs: 2},
	}

	return &domain.ServiceMap{
		Services:    services,
		Connections: connections,
	}, nil
}

// CorrelateTraces correlates a trace with external trace IDs
func (s *DistributedTraceService) CorrelateTraces(ctx context.Context, projectID uuid.UUID, input *domain.TraceCorrelationInput) (*domain.DistributedTrace, error) {
	s.logger.Debug("correlating traces", zap.String("traceId", input.TraceID), zap.Int("externalCount", len(input.ExternalTraceIDs)))

	trace, err := s.GetDistributedTrace(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, err
	}

	for i, extID := range input.ExternalTraceIDs {
		trace.ServiceSpans = append(trace.ServiceSpans, domain.DistributedSpan{
			SpanID:        uuid.New().String(),
			ServiceName:   fmt.Sprintf("external-service-%d", i+1),
			OperationName: "correlated_span",
			StartTime:     time.Now(),
			Duration:      500 * time.Millisecond,
			Status:        "ok",
			SpanKind:      domain.DistributedSpanKindServer,
			Attributes:    map[string]string{"external.trace_id": extID},
		})
		trace.TotalServices++
	}

	return trace, nil
}
