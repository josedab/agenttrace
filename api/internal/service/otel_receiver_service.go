package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// OTelReceiverService handles incoming OTLP traces and converts them to AgentTrace format
type OTelReceiverService struct {
	logger       *zap.Logger
	ingestionSvc *IngestionService
}

// NewOTelReceiverService creates a new OTel receiver service
func NewOTelReceiverService(
	logger *zap.Logger,
	ingestionSvc *IngestionService,
) *OTelReceiverService {
	return &OTelReceiverService{
		logger:       logger,
		ingestionSvc: ingestionSvc,
	}
}

// ReceiveTraces processes incoming OTLP traces and converts them to AgentTrace format
func (s *OTelReceiverService) ReceiveTraces(ctx context.Context, projectID uuid.UUID, request *domain.OTLPTraceRequest) (*domain.OTLPTraceResponse, error) {
	var rejectedSpans int64

	for _, rs := range request.ResourceSpans {
		traceInput := s.MapResourceToTrace(&rs.Resource)

		for _, ss := range rs.ScopeSpans {
			for _, span := range ss.Spans {
				obsInput := s.MapSpanToObservation(&span)
				if obsInput == nil {
					rejectedSpans++
					continue
				}

				// Use resource-level trace name if available
				if traceInput.Name != "" && (obsInput.Name == nil || *obsInput.Name == "") {
					obsInput.Name = &traceInput.Name
				}

				_, err := s.ingestionSvc.IngestObservation(ctx, projectID, obsInput)
				if err != nil {
					s.logger.Warn("failed to ingest OTLP span",
						zap.String("spanId", span.SpanID),
						zap.Error(err),
					)
					rejectedSpans++
				}
			}
		}
	}

	resp := &domain.OTLPTraceResponse{}
	if rejectedSpans > 0 {
		resp.PartialSuccess = &domain.PartialSuccess{
			RejectedSpans: rejectedSpans,
			ErrorMessage:  fmt.Sprintf("%d spans rejected during ingestion", rejectedSpans),
		}
	}

	return resp, nil
}

// ReceiveMetrics is a stub for future OTLP metrics ingestion
func (s *OTelReceiverService) ReceiveMetrics(ctx context.Context, projectID uuid.UUID, request any) (any, error) {
	s.logger.Debug("OTLP metrics receiver called (stub)", zap.String("projectId", projectID.String()))
	return map[string]string{"status": "not_implemented"}, nil
}

// ReceiveLogs is a stub for future OTLP logs ingestion
func (s *OTelReceiverService) ReceiveLogs(ctx context.Context, projectID uuid.UUID, request any) (any, error) {
	s.logger.Debug("OTLP logs receiver called (stub)", zap.String("projectId", projectID.String()))
	return map[string]string{"status": "not_implemented"}, nil
}

// MapSpanToObservation converts an OTLP span to an AgentTrace ObservationInput
func (s *OTelReceiverService) MapSpanToObservation(span *domain.OTLPSpan) *domain.ObservationInput {
	if span == nil {
		return nil
	}

	obsType := domain.ObservationTypeSpan
	// Detect LLM generation spans via semantic conventions
	if _, ok := span.Attributes[domain.OTelAttrLLMRequestModel]; ok {
		obsType = domain.ObservationTypeGeneration
	}
	if _, ok := span.Attributes[domain.OTelAttrLLMSystem]; ok {
		obsType = domain.ObservationTypeGeneration
	}

	startTime := span.StartTimeAsTime()
	endTime := span.EndTimeAsTime()
	spanID := span.SpanID
	traceID := span.TraceID
	spanName := span.Name

	input := &domain.ObservationInput{
		ID:        &spanID,
		TraceID:   &traceID,
		Type:      &obsType,
		Name:      &spanName,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	if span.ParentSpanID != "" {
		input.ParentObservationID = &span.ParentSpanID
	}

	// Map LLM-specific attributes
	if model, ok := span.Attributes[domain.OTelAttrLLMRequestModel].(string); ok {
		input.Model = &model
	}

	// Map metadata from remaining attributes
	metadata := make(map[string]any)
	for k, v := range span.Attributes {
		metadata[k] = v
	}
	if len(metadata) > 0 {
		input.Metadata = metadata
	}

	// Map error status
	if span.Status.Code == 2 { // ERROR
		level := domain.LevelError
		input.Level = &level
		if span.Status.Message != "" {
			msg := span.Status.Message
			input.StatusMessage = &msg
		}
	}

	return input
}

// MapResourceToTrace extracts trace-level information from an OTLP resource
func (s *OTelReceiverService) MapResourceToTrace(resource *domain.OTLPResource) *domain.TraceInput {
	input := &domain.TraceInput{}

	// Use service name as trace name
	if resource.ServiceName != "" {
		input.Name = resource.ServiceName
	} else if name, ok := resource.Attributes[domain.OTelAttrServiceName].(string); ok {
		input.Name = name
	}

	// Copy resource attributes as metadata
	if len(resource.Attributes) > 0 {
		metadata := make(map[string]any)
		for k, v := range resource.Attributes {
			metadata[k] = v
		}
		input.Metadata = metadata
	}

	return input
}
