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
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// OTelExporterService handles OpenTelemetry exporting
type OTelExporterService struct {
	logger     *zap.Logger
	httpClient *http.Client

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
	svc := &OTelExporterService{
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		batches:    make(map[uuid.UUID]*exportBatch),
		stopCh:     make(chan struct{}),
	}

	// Start batch processor
	go svc.processBatches()

	return svc
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
			domain.OTelAttrAgentTraceTraceID:    trace.ID.String(),
			domain.OTelAttrAgentTraceTraceName:  trace.Name,
		},
	}

	// Add custom resource attributes
	for k, v := range resourceAttrs {
		resource.Attributes[k] = v
	}

	// Convert trace ID to 16-byte hex
	traceIDBytes := trace.ID[:]
	traceIDHex := hex.EncodeToString(traceIDBytes)

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
	spanIDHex := hex.EncodeToString(trace.ID[:8])

	startTime := trace.CreatedAt.UnixNano()
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
		domain.OTelAttrAgentTraceTraceID:   trace.ID.String(),
		domain.OTelAttrAgentTraceTraceName: trace.Name,
	}

	if trace.TotalCost != nil {
		attrs[domain.OTelAttrAgentTraceCost] = *trace.TotalCost
	}
	if trace.LatencyMs != nil {
		attrs[domain.OTelAttrAgentTraceLatencyMs] = *trace.LatencyMs
	}

	// Add metadata as attributes
	for k, v := range trace.Metadata {
		attrs["agenttrace.metadata."+k] = v
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
	spanIDHex := hex.EncodeToString(obs.ID[:8])

	// Determine parent span ID
	parentSpanIDHex := hex.EncodeToString(trace.ID[:8]) // Default to root
	if obs.ParentObservationID != nil {
		parentSpanIDHex = hex.EncodeToString((*obs.ParentObservationID)[:8])
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
		domain.OTelAttrAgentTraceSpanID:   obs.ID.String(),
		domain.OTelAttrAgentTraceSpanType: obs.Type,
	}

	// Add LLM-specific attributes for generations
	if obs.Type == "generation" {
		if obs.Model != nil {
			attrs[domain.OTelAttrLLMRequestModel] = *obs.Model
			attrs[domain.OTelAttrLLMResponseModel] = *obs.Model
		}
		if obs.ModelParameters != nil {
			if temp, ok := (*obs.ModelParameters)["temperature"]; ok {
				attrs[domain.OTelAttrLLMRequestTemperature] = temp
			}
			if maxTokens, ok := (*obs.ModelParameters)["max_tokens"]; ok {
				attrs[domain.OTelAttrLLMRequestMaxTokens] = maxTokens
			}
		}
		if obs.UsageDetails != nil {
			if input, ok := (*obs.UsageDetails)["input"]; ok {
				attrs[domain.OTelAttrLLMUsageInputTokens] = input
			}
			if output, ok := (*obs.UsageDetails)["output"]; ok {
				attrs[domain.OTelAttrLLMUsageOutputTokens] = output
			}
		}
	}

	// Add cost and latency
	if obs.CalculatedTotalCost != nil {
		attrs[domain.OTelAttrAgentTraceCost] = *obs.CalculatedTotalCost
	}
	if obs.LatencyMs != nil {
		attrs[domain.OTelAttrAgentTraceLatencyMs] = *obs.LatencyMs
	}

	// Add metadata
	for k, v := range obs.Metadata {
		attrs["agenttrace.metadata."+k] = v
	}

	// Determine status
	status := domain.OTelSpanStatus{Code: domain.OTelStatusCodeOK}
	if obs.Level == "ERROR" {
		status = domain.OTelSpanStatus{
			Code:    domain.OTelStatusCodeError,
			Message: "Observation completed with error",
		}
		if obs.StatusMessage != nil {
			status.Message = *obs.StatusMessage
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

	// Send request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(exporter.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("export failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendGRPC sends spans via gRPC (placeholder - would use actual gRPC client)
func (s *OTelExporterService) sendGRPC(exporter *domain.OTelExporter, request *domain.OTelExportRequest) error {
	// In a real implementation, this would use the OTLP gRPC client
	// For now, fall back to HTTP if gRPC endpoint looks like HTTP
	s.logger.Debug("gRPC export - falling back to HTTP for prototype",
		zap.String("endpoint", exporter.Endpoint),
	)

	// Convert gRPC endpoint to HTTP for prototype
	httpExporter := *exporter
	httpExporter.Type = domain.OTelExporterTypeHTTP
	if httpExporter.Endpoint[0] != 'h' {
		httpExporter.Endpoint = "http://" + httpExporter.Endpoint + "/v1/traces"
	}

	return s.sendHTTP(&httpExporter, request)
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

// Stop stops the exporter service
func (s *OTelExporterService) Stop() {
	close(s.stopCh)
}
