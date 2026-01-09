package handler

import (
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OTelHandler handles OpenTelemetry-related HTTP requests
type OTelHandler struct {
	logger              *zap.Logger
	otelExporterService *service.OTelExporterService
}

// NewOTelHandler creates a new OpenTelemetry handler
func NewOTelHandler(
	logger *zap.Logger,
	otelExporterService *service.OTelExporterService,
) *OTelHandler {
	return &OTelHandler{
		logger:              logger,
		otelExporterService: otelExporterService,
	}
}

// ============================================================================
// OTLP Receiver Endpoints (for ingesting traces from OTel-instrumented apps)
// ============================================================================

// ReceiveTraces receives traces in OTLP format
// @Summary Receive OTLP traces
// @Description Receive traces in OpenTelemetry Protocol format
// @Tags otel
// @Accept json
// @Produce json
// @Param body body domain.OTelExportRequest true "OTLP trace data"
// @Success 200 {object} domain.OTelExportResponse
// @Failure 400 {object} ErrorResponse
// @Router /v1/traces [post]
func (h *OTelHandler) ReceiveTraces(c *fiber.Ctx) error {
	var request domain.OTelExportRequest
	if err := c.BodyParser(&request); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid OTLP request body", err)
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	// Process the incoming OTLP data
	var totalSpans int
	var rejectedSpans int64

	for _, resourceSpans := range request.ResourceSpans {
		// Extract resource attributes
		resourceAttrs := make(map[string]string)
		for k, v := range resourceSpans.Resource.Attributes {
			if s, ok := v.(string); ok {
				resourceAttrs[k] = s
			}
		}

		for _, scopeSpans := range resourceSpans.ScopeSpans {
			for _, span := range scopeSpans.Spans {
				totalSpans++

				// Convert OTLP span to AgentTrace observation
				_, err := h.convertOTelSpanToObservation(projectID, &span, resourceAttrs)
				if err != nil {
					h.logger.Warn("Failed to convert OTLP span",
						zap.String("spanId", span.SpanID),
						zap.Error(err),
					)
					rejectedSpans++
					continue
				}

				// In real implementation, save the observation to database
			}
		}
	}

	h.logger.Info("Received OTLP traces",
		zap.String("projectId", projectID.String()),
		zap.Int("totalSpans", totalSpans),
		zap.Int64("rejectedSpans", rejectedSpans),
	)

	response := domain.OTelExportResponse{}
	if rejectedSpans > 0 {
		response.PartialSuccess = &domain.OTelExportPartialSuccess{
			RejectedSpans: rejectedSpans,
			ErrorMessage:  "Some spans could not be processed",
		}
	}

	return c.JSON(response)
}

// convertOTelSpanToObservation converts an OTLP span to an AgentTrace observation
func (h *OTelHandler) convertOTelSpanToObservation(
	projectID uuid.UUID,
	span *domain.OTelSpan,
	resourceAttrs map[string]string,
) (*domain.Observation, error) {
	// Parse trace ID and span ID
	traceIDBytes, err := hex.DecodeString(span.TraceID)
	if err != nil {
		return nil, err
	}
	var traceID uuid.UUID
	if len(traceIDBytes) >= 16 {
		copy(traceID[:], traceIDBytes[:16])
	}

	spanIDBytes, err := hex.DecodeString(span.SpanID)
	if err != nil {
		return nil, err
	}
	var obsID uuid.UUID
	copy(obsID[:], spanIDBytes)
	// Fill remaining bytes with deterministic values
	for i := len(spanIDBytes); i < 16; i++ {
		obsID[i] = byte(i)
	}

	// Parse parent span ID
	var parentID *uuid.UUID
	if span.ParentSpanID != "" {
		parentBytes, err := hex.DecodeString(span.ParentSpanID)
		if err == nil {
			var pid uuid.UUID
			copy(pid[:], parentBytes)
			for i := len(parentBytes); i < 16; i++ {
				pid[i] = byte(i)
			}
			parentID = &pid
		}
	}

	// Determine observation type based on attributes
	obsType := "span"
	if _, ok := span.Attributes[domain.OTelAttrLLMRequestModel]; ok {
		obsType = "generation"
	}

	// Convert timestamps
	startTime := time.Unix(0, span.StartTimeUnixNano)
	endTime := time.Unix(0, span.EndTimeUnixNano)

	// Extract model info for generations
	var model *string
	var modelParams *map[string]any
	var usageDetails *map[string]any

	if obsType == "generation" {
		if m, ok := span.Attributes[domain.OTelAttrLLMRequestModel].(string); ok {
			model = &m
		}

		params := make(map[string]any)
		if temp, ok := span.Attributes[domain.OTelAttrLLMRequestTemperature]; ok {
			params["temperature"] = temp
		}
		if maxTokens, ok := span.Attributes[domain.OTelAttrLLMRequestMaxTokens]; ok {
			params["max_tokens"] = maxTokens
		}
		if len(params) > 0 {
			modelParams = &params
		}

		usage := make(map[string]any)
		if input, ok := span.Attributes[domain.OTelAttrLLMUsageInputTokens]; ok {
			usage["input"] = input
		}
		if output, ok := span.Attributes[domain.OTelAttrLLMUsageOutputTokens]; ok {
			usage["output"] = output
		}
		if len(usage) > 0 {
			usageDetails = &usage
		}
	}

	// Convert attributes to metadata
	metadata := make(map[string]any)
	for k, v := range span.Attributes {
		// Skip known attributes
		if k == domain.OTelAttrLLMRequestModel ||
			k == domain.OTelAttrLLMRequestTemperature ||
			k == domain.OTelAttrLLMRequestMaxTokens ||
			k == domain.OTelAttrLLMUsageInputTokens ||
			k == domain.OTelAttrLLMUsageOutputTokens {
			continue
		}
		metadata[k] = v
	}

	// Add resource attributes
	for k, v := range resourceAttrs {
		metadata["resource."+k] = v
	}

	// Determine level from status
	level := "DEFAULT"
	var statusMessage *string
	if span.Status.Code == domain.OTelStatusCodeError {
		level = "ERROR"
		if span.Status.Message != "" {
			statusMessage = &span.Status.Message
		}
	}

	// Calculate latency
	latencyMs := float64(span.EndTimeUnixNano-span.StartTimeUnixNano) / 1e6

	obs := &domain.Observation{
		ID:                  obsID,
		TraceID:             traceID,
		ProjectID:           projectID,
		Type:                obsType,
		Name:                span.Name,
		StartTime:           startTime,
		EndTime:             &endTime,
		Level:               level,
		StatusMessage:       statusMessage,
		Model:               model,
		ModelParameters:     modelParams,
		UsageDetails:        usageDetails,
		Metadata:            metadata,
		ParentObservationID: parentID,
		LatencyMs:           &latencyMs,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	return obs, nil
}

// ============================================================================
// Exporter Management Endpoints
// ============================================================================

// ListExporters returns all OTLP exporters for a project
// @Summary List OTLP exporters
// @Description Get all OTLP exporter configurations for a project
// @Tags otel
// @Accept json
// @Produce json
// @Success 200 {object} domain.OTelExporterList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/otel/exporters [get]
func (h *OTelHandler) ListExporters(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	h.logger.Debug("List OTLP exporters", zap.String("projectId", projectID.String()))

	// Return empty list for now
	result := domain.OTelExporterList{
		Exporters:  []domain.OTelExporter{},
		TotalCount: 0,
	}

	return c.JSON(result)
}

// GetExporter returns a specific OTLP exporter
// @Summary Get OTLP exporter
// @Description Get a specific OTLP exporter configuration
// @Tags otel
// @Accept json
// @Produce json
// @Param id path string true "Exporter ID"
// @Success 200 {object} domain.OTelExporter
// @Failure 404 {object} ErrorResponse
// @Router /api/public/otel/exporters/{id} [get]
func (h *OTelHandler) GetExporter(c *fiber.Ctx) error {
	exporterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid exporter ID", err)
	}

	h.logger.Debug("Get OTLP exporter", zap.String("exporterId", exporterID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Exporter not found", nil)
}

// CreateExporter creates a new OTLP exporter
// @Summary Create OTLP exporter
// @Description Create a new OTLP exporter configuration
// @Tags otel
// @Accept json
// @Produce json
// @Param exporter body domain.OTelExporterInput true "Exporter configuration"
// @Success 201 {object} domain.OTelExporter
// @Failure 400 {object} ErrorResponse
// @Router /api/public/otel/exporters [post]
func (h *OTelHandler) CreateExporter(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	userID := uuid.New() // In real implementation, get from auth context

	var input domain.OTelExporterInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	exporter, err := h.otelExporterService.CreateExporter(c.Context(), projectID, userID, &input)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return c.Status(fiber.StatusCreated).JSON(exporter)
}

// UpdateExporter updates an OTLP exporter
// @Summary Update OTLP exporter
// @Description Update an OTLP exporter configuration
// @Tags otel
// @Accept json
// @Produce json
// @Param id path string true "Exporter ID"
// @Param exporter body domain.OTelExporterInput true "Updated configuration"
// @Success 200 {object} domain.OTelExporter
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/otel/exporters/{id} [patch]
func (h *OTelHandler) UpdateExporter(c *fiber.Ctx) error {
	exporterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid exporter ID", err)
	}

	var input domain.OTelExporterInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Debug("Update OTLP exporter", zap.String("exporterId", exporterID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Exporter not found", nil)
}

// DeleteExporter deletes an OTLP exporter
// @Summary Delete OTLP exporter
// @Description Delete an OTLP exporter configuration
// @Tags otel
// @Accept json
// @Produce json
// @Param id path string true "Exporter ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Router /api/public/otel/exporters/{id} [delete]
func (h *OTelHandler) DeleteExporter(c *fiber.Ctx) error {
	exporterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid exporter ID", err)
	}

	h.logger.Info("Delete OTLP exporter", zap.String("exporterId", exporterID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Exporter not found", nil)
}

// ToggleExporter enables or disables an OTLP exporter
// @Summary Toggle OTLP exporter
// @Description Enable or disable an OTLP exporter
// @Tags otel
// @Accept json
// @Produce json
// @Param id path string true "Exporter ID"
// @Param body body ToggleExporterRequest true "Enable/disable"
// @Success 200 {object} domain.OTelExporter
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/otel/exporters/{id}/toggle [post]
func (h *OTelHandler) ToggleExporter(c *fiber.Ctx) error {
	exporterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid exporter ID", err)
	}

	var req ToggleExporterRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Info("Toggle OTLP exporter",
		zap.String("exporterId", exporterID.String()),
		zap.Bool("enabled", req.Enabled),
	)

	return errorResponse(c, fiber.StatusNotFound, "Exporter not found", nil)
}

// ToggleExporterRequest represents the request to toggle an exporter
type ToggleExporterRequest struct {
	Enabled bool `json:"enabled"`
}

// TestExporter tests an OTLP exporter connection
// @Summary Test OTLP exporter
// @Description Test the connection to an OTLP exporter
// @Tags otel
// @Accept json
// @Produce json
// @Param id path string true "Exporter ID"
// @Success 200 {object} TestExporterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/otel/exporters/{id}/test [post]
func (h *OTelHandler) TestExporter(c *fiber.Ctx) error {
	exporterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid exporter ID", err)
	}

	h.logger.Info("Test OTLP exporter", zap.String("exporterId", exporterID.String()))

	// In real implementation, fetch exporter and test
	return errorResponse(c, fiber.StatusNotFound, "Exporter not found", nil)
}

// TestExporterResponse represents the response from testing an exporter
type TestExporterResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Latency  int    `json:"latencyMs"`
}

// GetExporterStats returns statistics for an OTLP exporter
// @Summary Get exporter stats
// @Description Get export statistics for an OTLP exporter
// @Tags otel
// @Accept json
// @Produce json
// @Param id path string true "Exporter ID"
// @Success 200 {object} domain.OTelExporterStats
// @Failure 404 {object} ErrorResponse
// @Router /api/public/otel/exporters/{id}/stats [get]
func (h *OTelHandler) GetExporterStats(c *fiber.Ctx) error {
	exporterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid exporter ID", err)
	}

	h.logger.Debug("Get exporter stats", zap.String("exporterId", exporterID.String()))

	// Return mock stats
	stats := domain.OTelExporterStats{
		ExporterID:     exporterID,
		ExporterName:   "Unknown",
		TotalExported:  0,
		TotalErrors:    0,
		AvgLatencyMs:   0,
		ExportsLast24h: 0,
		ErrorsLast24h:  0,
	}

	return c.JSON(stats)
}

// GetDefaultConfig returns default configuration for common backends
// @Summary Get default config
// @Description Get default OTLP exporter configuration for common backends
// @Tags otel
// @Accept json
// @Produce json
// @Param backend query string true "Backend name (jaeger, zipkin, datadog, honeycomb, grafana-tempo, newrelic)"
// @Success 200 {object} domain.OTelExporterInput
// @Router /api/public/otel/defaults [get]
func (h *OTelHandler) GetDefaultConfig(c *fiber.Ctx) error {
	backend := c.Query("backend", "")
	if backend == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Backend parameter is required", nil)
	}

	config := h.otelExporterService.DefaultExporterConfig(backend)
	return c.JSON(config)
}

// GetSupportedBackends returns list of supported backends with configs
// @Summary Get supported backends
// @Description Get list of supported OTLP backends with default configurations
// @Tags otel
// @Accept json
// @Produce json
// @Success 200 {array} BackendInfo
// @Router /api/public/otel/backends [get]
func (h *OTelHandler) GetSupportedBackends(c *fiber.Ctx) error {
	backends := []BackendInfo{
		{
			ID:          "jaeger",
			Name:        "Jaeger",
			Description: "Open source distributed tracing",
			Type:        "grpc",
			DocsURL:     "https://www.jaegertracing.io/docs/",
		},
		{
			ID:          "zipkin",
			Name:        "Zipkin",
			Description: "Distributed tracing system",
			Type:        "http",
			DocsURL:     "https://zipkin.io/",
		},
		{
			ID:          "grafana-tempo",
			Name:        "Grafana Tempo",
			Description: "High-scale distributed tracing backend",
			Type:        "grpc",
			DocsURL:     "https://grafana.com/docs/tempo/",
		},
		{
			ID:          "datadog",
			Name:        "Datadog",
			Description: "Cloud monitoring platform",
			Type:        "http",
			DocsURL:     "https://docs.datadoghq.com/tracing/",
		},
		{
			ID:          "honeycomb",
			Name:        "Honeycomb",
			Description: "Observability for distributed systems",
			Type:        "grpc",
			DocsURL:     "https://docs.honeycomb.io/",
		},
		{
			ID:          "newrelic",
			Name:        "New Relic",
			Description: "Full-stack observability platform",
			Type:        "grpc",
			DocsURL:     "https://docs.newrelic.com/docs/distributed-tracing/",
		},
		{
			ID:          "otel-collector",
			Name:        "OpenTelemetry Collector",
			Description: "Vendor-agnostic telemetry collector",
			Type:        "grpc",
			DocsURL:     "https://opentelemetry.io/docs/collector/",
		},
	}

	return c.JSON(backends)
}

// BackendInfo represents information about a supported backend
type BackendInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	DocsURL     string `json:"docsUrl"`
}
