package handler

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OTLPGRPCServer implements the OpenTelemetry OTLP gRPC trace receiver.
type OTLPGRPCServer struct {
	v1.UnimplementedTraceServiceServer
	logger       *zap.Logger
	receiverSvc  *service.OTelReceiverService
	authService  *service.AuthService
	grpcServer   *grpc.Server
}

// NewOTLPGRPCServer creates a new OTLP gRPC server.
func NewOTLPGRPCServer(
	logger *zap.Logger,
	receiverSvc *service.OTelReceiverService,
	authService *service.AuthService,
) *OTLPGRPCServer {
	srv := &OTLPGRPCServer{
		logger:      logger.Named("otlp-grpc"),
		receiverSvc: receiverSvc,
		authService: authService,
	}

	srv.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(srv.authInterceptor),
	)
	v1.RegisterTraceServiceServer(srv.grpcServer, srv)

	return srv
}

// Start begins listening for OTLP gRPC requests.
func (s *OTLPGRPCServer) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.Info("OTLP gRPC server starting", zap.Int("port", port))
	return s.grpcServer.Serve(listener)
}

// Stop gracefully stops the gRPC server.
func (s *OTLPGRPCServer) Stop() {
	s.grpcServer.GracefulStop()
}

// Export implements the OTLP TraceService Export RPC.
func (s *OTLPGRPCServer) Export(
	ctx context.Context,
	req *v1.ExportTraceServiceRequest,
) (*v1.ExportTraceServiceResponse, error) {
	projectID, ok := ctx.Value(ctxKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "project ID not found in context")
	}

	if req == nil || len(req.ResourceSpans) == 0 {
		return &v1.ExportTraceServiceResponse{}, nil
	}

	// Convert protobuf to our domain model
	otlpReq := convertProtoToRequest(req)

	// Process via the receiver service
	result, err := s.receiverSvc.ReceiveTraces(ctx, projectID, otlpReq)
	if err != nil {
		s.logger.Error("failed to process OTLP traces",
			zap.Error(err),
			zap.Int("resourceSpans", len(req.ResourceSpans)),
		)
		return nil, status.Error(codes.Internal, "failed to process traces")
	}

	resp := &v1.ExportTraceServiceResponse{}
	if result.PartialSuccess != nil && result.PartialSuccess.RejectedSpans > 0 {
		resp.PartialSuccess = &v1.ExportTracePartialSuccess{
			RejectedSpans: result.PartialSuccess.RejectedSpans,
			ErrorMessage:  result.PartialSuccess.ErrorMessage,
		}
	}

	return resp, nil
}

type contextKey string

const ctxKeyProjectID contextKey = "projectID"

// authInterceptor validates the API key from gRPC metadata.
func (s *OTLPGRPCServer) authInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Extract API key from authorization header or x-api-key
	var apiKey string
	if vals := md.Get("authorization"); len(vals) > 0 {
		apiKey = vals[0]
		if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
			apiKey = apiKey[7:]
		}
	}
	if apiKey == "" {
		if vals := md.Get("x-api-key"); len(vals) > 0 {
			apiKey = vals[0]
		}
	}

	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "API key required")
	}

	// Validate key — extract public/secret parts
	projectID, err := s.authService.ValidateAPIKeyPublicOnly(ctx, apiKey)
	if err != nil || projectID == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid API key")
	}

	ctx = context.WithValue(ctx, ctxKeyProjectID, *projectID)
	return handler(ctx, req)
}

// convertProtoToRequest converts protobuf OTLP to our domain model.
func convertProtoToRequest(req *v1.ExportTraceServiceRequest) *domain.OTLPTraceRequest {
	result := &domain.OTLPTraceRequest{}

	for _, rs := range req.ResourceSpans {
		resourceSpan := domain.ResourceSpan{
			Resource: convertResourceProto(rs.Resource),
		}

		for _, ss := range rs.ScopeSpans {
			scopeSpan := domain.ScopeSpan{
				Scope: domain.InstrumentationScope{
					Name:    ss.Scope.GetName(),
					Version: ss.Scope.GetVersion(),
				},
			}

			for _, span := range ss.Spans {
				scopeSpan.Spans = append(scopeSpan.Spans, convertSpan(span))
			}

			resourceSpan.ScopeSpans = append(resourceSpan.ScopeSpans, scopeSpan)
		}

		result.ResourceSpans = append(result.ResourceSpans, resourceSpan)
	}

	return result
}

func convertResourceProto(r *resourcepb.Resource) domain.OTLPResource {
	if r == nil {
		return domain.OTLPResource{}
	}

	attrs := convertAttributes(r.Attributes)
	serviceName, _ := attrs["service.name"].(string)

	return domain.OTLPResource{
		ServiceName: serviceName,
		Attributes:  attrs,
	}
}

func convertSpan(s *tracepb.Span) domain.OTLPSpan {
	if s == nil {
		return domain.OTLPSpan{}
	}

	span := domain.OTLPSpan{
		TraceID:      fmt.Sprintf("%x", s.TraceId),
		SpanID:       fmt.Sprintf("%x", s.SpanId),
		ParentSpanID: fmt.Sprintf("%x", s.ParentSpanId),
		Name:         s.Name,
		Kind:         domain.SpanKind(s.Kind),
		StartTime:    s.StartTimeUnixNano,
		EndTime:      s.EndTimeUnixNano,
		Attributes:   convertAttributes(s.Attributes),
	}

	if s.Status != nil {
		span.Status = domain.SpanStatus{
			Code:    int(s.Status.Code),
			Message: s.Status.Message,
		}
	}

	for _, event := range s.Events {
		span.Events = append(span.Events, domain.SpanEvent{
			Name:       event.Name,
			Attributes: convertAttributes(event.Attributes),
		})
	}

	return span
}

func convertAttributes(attrs []*commonpb.KeyValue) map[string]any {
	result := make(map[string]any, len(attrs))
	for _, kv := range attrs {
		if kv.Value != nil {
			switch v := kv.Value.Value.(type) {
			case *commonpb.AnyValue_StringValue:
				result[kv.Key] = v.StringValue
			case *commonpb.AnyValue_IntValue:
				result[kv.Key] = v.IntValue
			case *commonpb.AnyValue_DoubleValue:
				result[kv.Key] = v.DoubleValue
			case *commonpb.AnyValue_BoolValue:
				result[kv.Key] = v.BoolValue
			}
		}
	}
	return result
}
