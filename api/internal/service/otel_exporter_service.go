package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/circuitbreaker"
)

// OTelExporterService handles OpenTelemetry exporting
type OTelExporterService struct {
	logger     *zap.Logger
	httpClient *http.Client
	cbRegistry *circuitbreaker.Registry

	// gRPC connection management
	grpcMu    sync.RWMutex
	grpcConns map[string]*grpc.ClientConn

	// Batch processing
	batchMu    sync.Mutex
	batches    map[uuid.UUID]*exportBatch
	stopCh     chan struct{}
}

// exportBatch holds spans waiting to be exported
type exportBatch struct {
	exporter *domain.OTelExporter
	spans    []*domain.OTelSpan
	mu       sync.Mutex
	lastSend time.Time
}

// NewOTelExporterService creates a new OTLP exporter service
func NewOTelExporterService(logger *zap.Logger) *OTelExporterService {
	registry := circuitbreaker.NewRegistry()

	svc := &OTelExporterService{
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cbRegistry: registry,
		grpcConns:  make(map[string]*grpc.ClientConn),
		batches:    make(map[uuid.UUID]*exportBatch),
		stopCh:     make(chan struct{}),
	}

	// Start batch processor
	go svc.processBatches()

	return svc
}

// getCircuitBreakerForEndpoint returns a circuit breaker for the given exporter endpoint
func (s *OTelExporterService) getCircuitBreakerForEndpoint(endpoint string) *circuitbreaker.CircuitBreaker {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		// Fallback to a default circuit breaker if URL parsing fails
		return s.cbRegistry.Get("otel:default", s.exporterCircuitBreakerConfig("default"))
	}

	host := parsedURL.Host
	if host == "" {
		// For gRPC endpoints that may not have scheme
		host = endpoint
	}
	return s.cbRegistry.Get("otel:"+host, s.exporterCircuitBreakerConfig(host))
}

// exporterCircuitBreakerConfig returns circuit breaker configuration for OTLP exporters
func (s *OTelExporterService) exporterCircuitBreakerConfig(name string) circuitbreaker.Config {
	return circuitbreaker.Config{
		Name:                "otel:" + name,
		MaxFailures:         5,                 // Open after 5 consecutive failures
		Timeout:             60 * time.Second,  // Try again after 1 minute
		MaxHalfOpenRequests: 1,
		OnStateChange: func(cbName string, from, to circuitbreaker.State) {
			s.logger.Info("OTLP exporter circuit breaker state changed",
				zap.String("circuit_breaker", cbName),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}
}

// CreateExporter creates a new OTLP exporter configuration
func (s *OTelExporterService) CreateExporter(
	ctx context.Context,
	projectID uuid.UUID,
	userID uuid.UUID,
	input *domain.OTelExporterInput,
) (*domain.OTelExporter, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Set defaults
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	timeout := 30
	if input.Timeout != nil {
		timeout = *input.Timeout
	}

	insecure := false
	if input.Insecure != nil {
		insecure = *input.Insecure
	}

	samplingRate := 1.0
	if input.SamplingRate != nil {
		samplingRate = *input.SamplingRate
	}

	batchConfig := domain.OTelBatchConfig{
		MaxBatchSize:  512,
		MaxQueueSize:  2048,
		BatchTimeout:  5000,
		ExportTimeout: 30000,
		ScheduleDelay: 1000,
	}
	if input.BatchConfig != nil {
		batchConfig = *input.BatchConfig
	}

	retryConfig := domain.OTelRetryConfig{
		Enabled:         true,
		InitialInterval: 1000,
		MaxInterval:     30000,
		MaxElapsedTime:  300000,
		Multiplier:      1.5,
	}
	if input.RetryConfig != nil {
		retryConfig = *input.RetryConfig
	}

	compression := "gzip"
	if input.Compression != "" {
		compression = input.Compression
	}

	now := time.Now()
	exporter := &domain.OTelExporter{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		Name:               input.Name,
		Enabled:            enabled,
		Type:               input.Type,
		Status:             domain.OTelExporterStatusActive,
		Endpoint:           input.Endpoint,
		Headers:            input.Headers,
		Compression:        compression,
		Timeout:            timeout,
		Insecure:           insecure,
		TLSConfig:          input.TLSConfig,
		BatchConfig:        batchConfig,
		RetryConfig:        retryConfig,
		ResourceAttributes: input.ResourceAttributes,
		TraceNameFilter:    input.TraceNameFilter,
		MetadataFilters:    input.MetadataFilters,
		SamplingRate:       samplingRate,
		ExportedCount:      0,
		ErrorCount:         0,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          userID,
	}

	s.logger.Info("Created OTLP exporter",
		zap.String("exporterId", exporter.ID.String()),
		zap.String("name", exporter.Name),
		zap.String("endpoint", exporter.Endpoint),
	)

	return exporter, nil
}

// DefaultExporterConfig returns sensible defaults for common backends
func (s *OTelExporterService) DefaultExporterConfig(backend string) *domain.OTelExporterInput {
	switch backend {
	case "jaeger":
		return &domain.OTelExporterInput{
			Name:     "Jaeger",
			Type:     domain.OTelExporterTypeGRPC,
			Endpoint: "localhost:4317",
		}
	case "zipkin":
		return &domain.OTelExporterInput{
			Name:     "Zipkin",
			Type:     domain.OTelExporterTypeHTTP,
			Endpoint: "http://localhost:9411/api/v2/spans",
		}
	case "datadog":
		return &domain.OTelExporterInput{
			Name:     "Datadog",
			Type:     domain.OTelExporterTypeHTTP,
			Endpoint: "https://trace.agent.datadoghq.com/api/v0.2/traces",
			Headers: map[string]string{
				"DD-API-KEY": "${DD_API_KEY}",
			},
		}
	case "honeycomb":
		return &domain.OTelExporterInput{
			Name:     "Honeycomb",
			Type:     domain.OTelExporterTypeGRPC,
			Endpoint: "api.honeycomb.io:443",
			Headers: map[string]string{
				"x-honeycomb-team": "${HONEYCOMB_API_KEY}",
			},
		}
	case "grafana-tempo":
		return &domain.OTelExporterInput{
			Name:     "Grafana Tempo",
			Type:     domain.OTelExporterTypeGRPC,
			Endpoint: "localhost:4317",
		}
	case "newrelic":
		return &domain.OTelExporterInput{
			Name:     "New Relic",
			Type:     domain.OTelExporterTypeGRPC,
			Endpoint: "otlp.nr-data.net:4317",
			Headers: map[string]string{
				"api-key": "${NEW_RELIC_API_KEY}",
			},
		}
	default:
		return &domain.OTelExporterInput{
			Name:     "OTLP Collector",
			Type:     domain.OTelExporterTypeGRPC,
			Endpoint: "localhost:4317",
		}
	}
}

// ConvertTraceToOTel converts an AgentTrace trace to OpenTelemetry format
func (s *OTelExporterService) ConvertTraceToOTel(
	trace *domain.Trace,
	observations []domain.Observation,
	resourceAttrs map[string]string,
) *domain.OTelResourceSpans {
	// Build resource
	resource := domain.OTelResource{
		Attributes: map[string]any{
			domain.OTelAttrServiceName:          "agenttrace",
			domain.OTelAttrServiceVersion:       "1.0.0",
			domain.OTelAttrAgentTraceProjectID:  trace.ProjectID.String(),
			domain.OTelAttrAgentTraceTraceID:    trace.ID,
			domain.OTelAttrAgentTraceTraceName:  trace.Name,
		},
	}

	// Add custom resource attributes
	for k, v := range resourceAttrs {
		resource.Attributes[k] = v
	}

	// Use trace ID directly as hex (it's already a hex string)
	traceIDHex := trace.ID

	// Convert observations to spans
	spans := make([]domain.OTelSpan, 0, len(observations)+1)

	// Root span for the trace
	rootSpan := s.createRootSpan(trace, traceIDHex)
	spans = append(spans, rootSpan)

	// Child spans for observations
	for _, obs := range observations {
		span := s.convertObservationToSpan(trace, &obs, traceIDHex)
		spans = append(spans, span)
	}

	return &domain.OTelResourceSpans{
		Resource: resource,
		ScopeSpans: []domain.OTelScopeSpans{
			{
				Scope: domain.OTelScope{
					Name:    "agenttrace",
					Version: "1.0.0",
				},
				Spans: spans,
			},
		},
	}
}

// createRootSpan creates the root span for a trace
func (s *OTelExporterService) createRootSpan(trace *domain.Trace, traceIDHex string) domain.OTelSpan {
	// Generate span ID from first 8 chars of trace ID
	spanIDHex := trace.ID
	if len(spanIDHex) > 16 {
		spanIDHex = spanIDHex[:16]
	}

	startTime := trace.StartTime.UnixNano()
	endTime := startTime
	if trace.EndTime != nil {
		endTime = trace.EndTime.UnixNano()
	}

	status := domain.OTelSpanStatus{Code: domain.OTelStatusCodeOK}
	if trace.Level == "ERROR" {
		status = domain.OTelSpanStatus{
			Code:    domain.OTelStatusCodeError,
			Message: "Trace completed with error",
		}
	}

	attrs := map[string]any{
		domain.OTelAttrAgentTraceTraceID:   trace.ID,
		domain.OTelAttrAgentTraceTraceName: trace.Name,
	}

	if trace.TotalCost > 0 {
		attrs[domain.OTelAttrAgentTraceCost] = trace.TotalCost
	}
	if trace.DurationMs > 0 {
		attrs[domain.OTelAttrAgentTraceLatencyMs] = trace.DurationMs
	}

	return domain.OTelSpan{
		TraceID:           traceIDHex,
		SpanID:            spanIDHex,
		Name:              trace.Name,
		Kind:              domain.OTelSpanKindServer,
		StartTimeUnixNano: startTime,
		EndTimeUnixNano:   endTime,
		Attributes:        attrs,
		Status:            status,
	}
}

// convertObservationToSpan converts an observation to an OTel span
func (s *OTelExporterService) convertObservationToSpan(
	trace *domain.Trace,
	obs *domain.Observation,
	traceIDHex string,
) domain.OTelSpan {
	// Use observation ID directly (it's already a string)
	spanIDHex := obs.ID
	if len(spanIDHex) > 16 {
		spanIDHex = spanIDHex[:16]
	}

	// Determine parent span ID
	parentSpanIDHex := trace.ID // Default to root
	if len(parentSpanIDHex) > 16 {
		parentSpanIDHex = parentSpanIDHex[:16]
	}
	if obs.ParentObservationID != nil {
		parentSpanIDHex = *obs.ParentObservationID
		if len(parentSpanIDHex) > 16 {
			parentSpanIDHex = parentSpanIDHex[:16]
		}
	}

	startTime := obs.StartTime.UnixNano()
	endTime := startTime
	if obs.EndTime != nil {
		endTime = obs.EndTime.UnixNano()
	}

	// Determine span kind based on observation type
	kind := domain.OTelSpanKindInternal
	switch obs.Type {
	case "generation":
		kind = domain.OTelSpanKindClient // LLM calls are client spans
	case "span":
		kind = domain.OTelSpanKindInternal
	}

	// Build attributes
	attrs := map[string]any{
		domain.OTelAttrAgentTraceSpanID:   obs.ID,
		domain.OTelAttrAgentTraceSpanType: string(obs.Type),
	}

	// Add LLM-specific attributes for generations
	if obs.Type == domain.ObservationTypeGeneration {
		if obs.Model != "" {
			attrs[domain.OTelAttrLLMRequestModel] = obs.Model
			attrs[domain.OTelAttrLLMResponseModel] = obs.Model
		}
		// Parse model parameters if present
		if obs.ModelParameters != "" {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(obs.ModelParameters), &params); err == nil {
				if temp, ok := params["temperature"]; ok {
					attrs[domain.OTelAttrLLMRequestTemperature] = temp
				}
				if maxTokens, ok := params["max_tokens"]; ok {
					attrs[domain.OTelAttrLLMRequestMaxTokens] = maxTokens
				}
			}
		}
		// Add token usage
		if obs.UsageDetails.InputTokens > 0 {
			attrs[domain.OTelAttrLLMUsageInputTokens] = obs.UsageDetails.InputTokens
		}
		if obs.UsageDetails.OutputTokens > 0 {
			attrs[domain.OTelAttrLLMUsageOutputTokens] = obs.UsageDetails.OutputTokens
		}
	}

	// Add cost and latency
	if obs.CostDetails.TotalCost > 0 {
		attrs[domain.OTelAttrAgentTraceCost] = obs.CostDetails.TotalCost
	}
	if obs.DurationMs > 0 {
		attrs[domain.OTelAttrAgentTraceLatencyMs] = obs.DurationMs
	}

	// Determine status
	status := domain.OTelSpanStatus{Code: domain.OTelStatusCodeOK}
	if obs.Level == "ERROR" {
		status = domain.OTelSpanStatus{
			Code:    domain.OTelStatusCodeError,
			Message: "Observation completed with error",
		}
		if obs.StatusMessage != "" {
			status.Message = obs.StatusMessage
		}
	}

	return domain.OTelSpan{
		TraceID:           traceIDHex,
		SpanID:            spanIDHex,
		ParentSpanID:      parentSpanIDHex,
		Name:              obs.Name,
		Kind:              kind,
		StartTimeUnixNano: startTime,
		EndTimeUnixNano:   endTime,
		Attributes:        attrs,
		Status:            status,
	}
}

// QueueSpansForExport adds spans to the export queue
func (s *OTelExporterService) QueueSpansForExport(
	exporter *domain.OTelExporter,
	spans []*domain.OTelSpan,
) error {
	if !exporter.Enabled {
		return nil
	}

	// Apply sampling
	if exporter.SamplingRate < 1.0 {
		if rand.Float64() > exporter.SamplingRate {
			return nil // Skip this batch
		}
	}

	s.batchMu.Lock()
	batch, exists := s.batches[exporter.ID]
	if !exists {
		batch = &exportBatch{
			exporter: exporter,
			spans:    make([]*domain.OTelSpan, 0, exporter.BatchConfig.MaxBatchSize),
			lastSend: time.Now(),
		}
		s.batches[exporter.ID] = batch
	}
	s.batchMu.Unlock()

	batch.mu.Lock()
	defer batch.mu.Unlock()

	batch.spans = append(batch.spans, spans...)

	// Check if batch should be sent immediately
	if len(batch.spans) >= exporter.BatchConfig.MaxBatchSize {
		go s.sendBatch(exporter, batch)
	}

	return nil
}

// processBatches periodically sends batches
func (s *OTelExporterService) processBatches() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.batchMu.Lock()
			for _, batch := range s.batches {
				batch.mu.Lock()
				shouldSend := len(batch.spans) > 0 &&
					time.Since(batch.lastSend).Milliseconds() >= int64(batch.exporter.BatchConfig.BatchTimeout)
				if shouldSend {
					go s.sendBatch(batch.exporter, batch)
				}
				batch.mu.Unlock()
			}
			s.batchMu.Unlock()
		}
	}
}

// sendBatch sends a batch of spans to the exporter
func (s *OTelExporterService) sendBatch(exporter *domain.OTelExporter, batch *exportBatch) {
	batch.mu.Lock()
	if len(batch.spans) == 0 {
		batch.mu.Unlock()
		return
	}

	spans := batch.spans
	batch.spans = make([]*domain.OTelSpan, 0, exporter.BatchConfig.MaxBatchSize)
	batch.lastSend = time.Now()
	batch.mu.Unlock()

	// Build export request
	otelSpans := make([]domain.OTelSpan, len(spans))
	for i, span := range spans {
		otelSpans[i] = *span
	}

	request := domain.OTelExportRequest{
		ResourceSpans: []domain.OTelResourceSpans{
			{
				Resource: domain.OTelResource{
					Attributes: map[string]any{
						domain.OTelAttrServiceName: "agenttrace",
					},
				},
				ScopeSpans: []domain.OTelScopeSpans{
					{
						Scope: domain.OTelScope{Name: "agenttrace", Version: "1.0.0"},
						Spans: otelSpans,
					},
				},
			},
		},
	}

	// Send based on exporter type
	var err error
	switch exporter.Type {
	case domain.OTelExporterTypeHTTP:
		err = s.sendHTTP(exporter, &request)
	case domain.OTelExporterTypeGRPC:
		err = s.sendGRPC(exporter, &request)
	default:
		err = fmt.Errorf("unsupported exporter type: %s", exporter.Type)
	}

	if err != nil {
		s.logger.Error("Failed to export spans",
			zap.String("exporterId", exporter.ID.String()),
			zap.Error(err),
		)
		// Update error stats (in real impl, persist to DB)
	} else {
		s.logger.Debug("Exported spans",
			zap.String("exporterId", exporter.ID.String()),
			zap.Int("count", len(spans)),
		)
	}
}

// sendHTTP sends spans via HTTP/JSON
func (s *OTelExporterService) sendHTTP(exporter *domain.OTelExporter, request *domain.OTelExportRequest) error {
	// Marshal to JSON
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Compress if enabled
	var body io.Reader = bytes.NewReader(data)
	var contentEncoding string
	if exporter.Compression == "gzip" {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		if _, err := gzWriter.Write(data); err != nil {
			return fmt.Errorf("failed to compress: %w", err)
		}
		gzWriter.Close()
		body = &buf
		contentEncoding = "gzip"
	}

	// Create request
	req, err := http.NewRequest("POST", exporter.Endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if contentEncoding != "" {
		req.Header.Set("Content-Encoding", contentEncoding)
	}

	// Add custom headers
	for k, v := range exporter.Headers {
		// Expand environment variables
		if len(v) > 2 && v[0] == '$' && v[1] == '{' && v[len(v)-1] == '}' {
			envVar := v[2 : len(v)-1]
			v = os.Getenv(envVar)
		}
		req.Header.Set(k, v)
	}

	// Configure TLS if needed
	client := s.getHTTPClient(exporter)

	// Send request with circuit breaker protection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(exporter.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Get circuit breaker for this exporter's endpoint
	cb := s.getCircuitBreakerForEndpoint(exporter.Endpoint)

	var resp *http.Response
	err = cb.Execute(ctx, func() error {
		var httpErr error
		resp, httpErr = client.Do(req)
		if httpErr != nil {
			return fmt.Errorf("failed to send request: %w", httpErr)
		}
		return nil
	})

	// Check if circuit breaker blocked the request
	if err == circuitbreaker.ErrCircuitOpen {
		s.logger.Warn("OTLP export blocked by circuit breaker",
			zap.String("exporter_id", exporter.ID.String()),
			zap.String("endpoint", exporter.Endpoint),
		)
		return fmt.Errorf("circuit breaker open: OTLP endpoint temporarily unavailable")
	}
	if err == circuitbreaker.ErrTooManyRequests {
		return fmt.Errorf("circuit breaker half-open: too many concurrent requests to OTLP endpoint")
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("export failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendGRPC sends spans via gRPC using the OTLP protocol
func (s *OTelExporterService) sendGRPC(exporter *domain.OTelExporter, request *domain.OTelExportRequest) error {
	// Get or create gRPC connection
	conn, err := s.getGRPCConnection(exporter)
	if err != nil {
		return fmt.Errorf("failed to get gRPC connection: %w", err)
	}

	// Create trace service client
	client := v1.NewTraceServiceClient(conn)

	// Convert domain request to protobuf
	pbRequest, err := s.convertToProto(request)
	if err != nil {
		return fmt.Errorf("failed to convert request to protobuf: %w", err)
	}

	// Create context with timeout and headers
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(exporter.Timeout)*time.Second)
	defer cancel()

	// Add custom headers as metadata
	if len(exporter.Headers) > 0 {
		md := metadata.MD{}
		for k, v := range exporter.Headers {
			// Expand environment variables
			if len(v) > 2 && v[0] == '$' && v[1] == '{' && v[len(v)-1] == '}' {
				envVar := v[2 : len(v)-1]
				v = os.Getenv(envVar)
			}
			md.Set(k, v)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Get circuit breaker for this endpoint
	cb := s.getCircuitBreakerForEndpoint(exporter.Endpoint)

	// Send request with circuit breaker protection
	err = cb.Execute(ctx, func() error {
		resp, grpcErr := client.Export(ctx, pbRequest)
		if grpcErr != nil {
			return fmt.Errorf("gRPC export failed: %w", grpcErr)
		}

		// Check for partial success
		if resp.GetPartialSuccess() != nil && resp.GetPartialSuccess().GetRejectedSpans() > 0 {
			s.logger.Warn("OTLP gRPC export had partial success",
				zap.Int64("rejected_spans", resp.GetPartialSuccess().GetRejectedSpans()),
				zap.String("error_message", resp.GetPartialSuccess().GetErrorMessage()),
			)
		}

		return nil
	})

	// Check if circuit breaker blocked the request
	if err == circuitbreaker.ErrCircuitOpen {
		s.logger.Warn("OTLP gRPC export blocked by circuit breaker",
			zap.String("exporter_id", exporter.ID.String()),
			zap.String("endpoint", exporter.Endpoint),
		)
		return fmt.Errorf("circuit breaker open: OTLP gRPC endpoint temporarily unavailable")
	}
	if err == circuitbreaker.ErrTooManyRequests {
		return fmt.Errorf("circuit breaker half-open: too many concurrent requests to OTLP gRPC endpoint")
	}

	return err
}

// getGRPCConnection returns a cached or new gRPC connection for the exporter
func (s *OTelExporterService) getGRPCConnection(exporter *domain.OTelExporter) (*grpc.ClientConn, error) {
	// Check cache first
	s.grpcMu.RLock()
	conn, exists := s.grpcConns[exporter.Endpoint]
	s.grpcMu.RUnlock()

	if exists && conn != nil {
		return conn, nil
	}

	// Create new connection
	s.grpcMu.Lock()
	defer s.grpcMu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists := s.grpcConns[exporter.Endpoint]; exists && conn != nil {
		return conn, nil
	}

	// Build dial options
	opts := []grpc.DialOption{}

	if exporter.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Configure TLS
		tlsConfig := &tls.Config{}

		if exporter.TLSConfig != nil {
			// Load client certificate if specified
			if exporter.TLSConfig.CertFile != "" && exporter.TLSConfig.KeyFile != "" {
				cert, err := tls.LoadX509KeyPair(exporter.TLSConfig.CertFile, exporter.TLSConfig.KeyFile)
				if err != nil {
					return nil, fmt.Errorf("failed to load client certificate: %w", err)
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
			}

			// Load CA certificate if specified
			if exporter.TLSConfig.CAFile != "" {
				caCert, err := os.ReadFile(exporter.TLSConfig.CAFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read CA file: %w", err)
				}
				caCertPool := x509.NewCertPool()
				if !caCertPool.AppendCertsFromPEM(caCert) {
					return nil, fmt.Errorf("failed to parse CA certificate")
				}
				tlsConfig.RootCAs = caCertPool
			}

			// Set server name if specified
			if exporter.TLSConfig.ServerName != "" {
				tlsConfig.ServerName = exporter.TLSConfig.ServerName
			}
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Create the connection
	conn, err := grpc.NewClient(exporter.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	s.grpcConns[exporter.Endpoint] = conn
	s.logger.Info("Created gRPC connection for OTLP export",
		zap.String("endpoint", exporter.Endpoint),
		zap.String("exporter_id", exporter.ID.String()),
	)

	return conn, nil
}

// convertToProto converts domain OTelExportRequest to protobuf format
func (s *OTelExporterService) convertToProto(request *domain.OTelExportRequest) (*v1.ExportTraceServiceRequest, error) {
	pbResourceSpans := make([]*tracepb.ResourceSpans, 0, len(request.ResourceSpans))

	for _, rs := range request.ResourceSpans {
		pbResource := &resourcepb.Resource{
			Attributes: s.convertAttributesToProto(rs.Resource.Attributes),
		}

		pbScopeSpans := make([]*tracepb.ScopeSpans, 0, len(rs.ScopeSpans))
		for _, ss := range rs.ScopeSpans {
			pbScope := &commonpb.InstrumentationScope{
				Name:    ss.Scope.Name,
				Version: ss.Scope.Version,
			}

			pbSpans := make([]*tracepb.Span, 0, len(ss.Spans))
			for _, span := range ss.Spans {
				pbSpan, err := s.convertSpanToProto(&span)
				if err != nil {
					s.logger.Warn("Failed to convert span to protobuf",
						zap.String("span_id", span.SpanID),
						zap.Error(err),
					)
					continue
				}
				pbSpans = append(pbSpans, pbSpan)
			}

			pbScopeSpans = append(pbScopeSpans, &tracepb.ScopeSpans{
				Scope: pbScope,
				Spans: pbSpans,
			})
		}

		pbResourceSpans = append(pbResourceSpans, &tracepb.ResourceSpans{
			Resource:   pbResource,
			ScopeSpans: pbScopeSpans,
		})
	}

	return &v1.ExportTraceServiceRequest{
		ResourceSpans: pbResourceSpans,
	}, nil
}

// convertSpanToProto converts a domain OTelSpan to protobuf format
func (s *OTelExporterService) convertSpanToProto(span *domain.OTelSpan) (*tracepb.Span, error) {
	// Convert trace ID (hex string to bytes)
	traceID, err := hexToBytes(span.TraceID, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid trace ID: %w", err)
	}

	// Convert span ID (hex string to bytes)
	spanID, err := hexToBytes(span.SpanID, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid span ID: %w", err)
	}

	pbSpan := &tracepb.Span{
		TraceId:                traceID,
		SpanId:                 spanID,
		Name:                   span.Name,
		Kind:                   tracepb.Span_SpanKind(span.Kind),
		StartTimeUnixNano:      uint64(span.StartTimeUnixNano),
		EndTimeUnixNano:        uint64(span.EndTimeUnixNano),
		Attributes:             s.convertAttributesToProto(span.Attributes),
		Status: &tracepb.Status{
			Code:    tracepb.Status_StatusCode(span.Status.Code),
			Message: span.Status.Message,
		},
	}

	// Convert parent span ID if present
	if span.ParentSpanID != "" {
		parentSpanID, err := hexToBytes(span.ParentSpanID, 8)
		if err == nil {
			pbSpan.ParentSpanId = parentSpanID
		}
	}

	// Convert events
	if len(span.Events) > 0 {
		pbSpan.Events = make([]*tracepb.Span_Event, 0, len(span.Events))
		for _, event := range span.Events {
			pbSpan.Events = append(pbSpan.Events, &tracepb.Span_Event{
				Name:              event.Name,
				TimeUnixNano:      uint64(event.TimeUnixNano),
				Attributes:        s.convertAttributesToProto(event.Attributes),
			})
		}
	}

	// Convert links
	if len(span.Links) > 0 {
		pbSpan.Links = make([]*tracepb.Span_Link, 0, len(span.Links))
		for _, link := range span.Links {
			linkTraceID, err := hexToBytes(link.TraceID, 16)
			if err != nil {
				continue
			}
			linkSpanID, err := hexToBytes(link.SpanID, 8)
			if err != nil {
				continue
			}
			pbSpan.Links = append(pbSpan.Links, &tracepb.Span_Link{
				TraceId:    linkTraceID,
				SpanId:     linkSpanID,
				Attributes: s.convertAttributesToProto(link.Attributes),
			})
		}
	}

	return pbSpan, nil
}

// convertAttributesToProto converts map attributes to protobuf KeyValue slice
func (s *OTelExporterService) convertAttributesToProto(attrs map[string]any) []*commonpb.KeyValue {
	if len(attrs) == 0 {
		return nil
	}

	result := make([]*commonpb.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		kv := &commonpb.KeyValue{Key: k}

		switch val := v.(type) {
		case string:
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: val}}
		case bool:
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_BoolValue{BoolValue: val}}
		case int:
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(val)}}
		case int64:
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: val}}
		case float64:
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_DoubleValue{DoubleValue: val}}
		case float32:
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_DoubleValue{DoubleValue: float64(val)}}
		default:
			// Convert to string for unknown types
			kv.Value = &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: fmt.Sprintf("%v", val)}}
		}

		result = append(result, kv)
	}

	return result
}

// hexToBytes converts a hex string to bytes with expected length
func hexToBytes(hexStr string, expectedLen int) ([]byte, error) {
	// Pad with zeros if too short
	for len(hexStr) < expectedLen*2 {
		hexStr = "0" + hexStr
	}

	// Truncate if too long
	if len(hexStr) > expectedLen*2 {
		hexStr = hexStr[:expectedLen*2]
	}

	return hex.DecodeString(hexStr)
}

// getHTTPClient returns an HTTP client configured for the exporter
func (s *OTelExporterService) getHTTPClient(exporter *domain.OTelExporter) *http.Client {
	transport := &http.Transport{}

	if exporter.TLSConfig != nil || exporter.Insecure {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: exporter.Insecure,
		}

		if exporter.TLSConfig != nil {
			// Load certificates if specified
			if exporter.TLSConfig.CertFile != "" && exporter.TLSConfig.KeyFile != "" {
				cert, err := tls.LoadX509KeyPair(exporter.TLSConfig.CertFile, exporter.TLSConfig.KeyFile)
				if err == nil {
					tlsConfig.Certificates = []tls.Certificate{cert}
				}
			}

			if exporter.TLSConfig.CAFile != "" {
				caCert, err := os.ReadFile(exporter.TLSConfig.CAFile)
				if err == nil {
					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCert)
					tlsConfig.RootCAs = caCertPool
				}
			}

			if exporter.TLSConfig.ServerName != "" {
				tlsConfig.ServerName = exporter.TLSConfig.ServerName
			}
		}

		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(exporter.Timeout) * time.Second,
	}
}

// TestExporter tests the connection to an exporter
func (s *OTelExporterService) TestExporter(ctx context.Context, exporter *domain.OTelExporter) error {
	// Create a test span
	testSpan := &domain.OTelSpan{
		TraceID:           "00000000000000000000000000000001",
		SpanID:            "0000000000000001",
		Name:              "agenttrace.test",
		Kind:              domain.OTelSpanKindInternal,
		StartTimeUnixNano: time.Now().UnixNano(),
		EndTimeUnixNano:   time.Now().UnixNano(),
		Attributes: map[string]any{
			"test": true,
		},
		Status: domain.OTelSpanStatus{Code: domain.OTelStatusCodeOK},
	}

	request := domain.OTelExportRequest{
		ResourceSpans: []domain.OTelResourceSpans{
			{
				Resource: domain.OTelResource{
					Attributes: map[string]any{
						domain.OTelAttrServiceName: "agenttrace-test",
					},
				},
				ScopeSpans: []domain.OTelScopeSpans{
					{
						Scope: domain.OTelScope{Name: "agenttrace", Version: "1.0.0"},
						Spans: []domain.OTelSpan{*testSpan},
					},
				},
			},
		},
	}

	switch exporter.Type {
	case domain.OTelExporterTypeHTTP:
		return s.sendHTTP(exporter, &request)
	case domain.OTelExporterTypeGRPC:
		return s.sendGRPC(exporter, &request)
	default:
		return fmt.Errorf("unsupported exporter type: %s", exporter.Type)
	}
}

// Stop stops the exporter service and closes all connections
func (s *OTelExporterService) Stop() {
	close(s.stopCh)

	// Close all gRPC connections
	s.grpcMu.Lock()
	defer s.grpcMu.Unlock()

	for endpoint, conn := range s.grpcConns {
		if err := conn.Close(); err != nil {
			s.logger.Warn("Failed to close gRPC connection",
				zap.String("endpoint", endpoint),
				zap.Error(err),
			)
		}
	}
	s.grpcConns = make(map[string]*grpc.ClientConn)
}
