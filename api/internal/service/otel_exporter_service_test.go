package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestOTelExporterService_CreateExporter(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	t.Run("creates exporter with valid input", func(t *testing.T) {
		projectID := uuid.New()
		userID := uuid.New()
		input := &domain.OTelExporterInput{
			Name:     "Test Exporter",
			Endpoint: "localhost:4317",
			Type:     domain.OTelExporterTypeGRPC,
		}

		exporter, err := svc.CreateExporter(nil, projectID, userID, input)
		require.NoError(t, err)
		assert.Equal(t, "Test Exporter", exporter.Name)
		assert.Equal(t, "localhost:4317", exporter.Endpoint)
		assert.True(t, exporter.Enabled)
		assert.Equal(t, userID, exporter.CreatedBy)
	})

	t.Run("returns error for missing name", func(t *testing.T) {
		_, err := svc.CreateExporter(nil, uuid.New(), uuid.New(), &domain.OTelExporterInput{
			Endpoint: "localhost:4317",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("returns error for missing endpoint", func(t *testing.T) {
		_, err := svc.CreateExporter(nil, uuid.New(), uuid.New(), &domain.OTelExporterInput{
			Name: "Test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("applies default batch and retry config", func(t *testing.T) {
		exporter, err := svc.CreateExporter(nil, uuid.New(), uuid.New(), &domain.OTelExporterInput{
			Name:     "Defaults",
			Endpoint: "localhost:4317",
		})
		require.NoError(t, err)
		assert.Equal(t, 512, exporter.BatchConfig.MaxBatchSize)
		assert.Equal(t, true, exporter.RetryConfig.Enabled)
		assert.Equal(t, "gzip", exporter.Compression)
		assert.Equal(t, 1.0, exporter.SamplingRate)
	})
}

func TestOTelExporterService_DefaultExporterConfig(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	tests := []struct {
		backend      string
		expectedName string
		expectedType domain.OTelExporterType
	}{
		{"jaeger", "Jaeger", domain.OTelExporterTypeGRPC},
		{"zipkin", "Zipkin", domain.OTelExporterTypeHTTP},
		{"datadog", "Datadog", domain.OTelExporterTypeHTTP},
		{"honeycomb", "Honeycomb", domain.OTelExporterTypeGRPC},
		{"unknown", "OTLP Collector", domain.OTelExporterTypeGRPC},
	}

	for _, tt := range tests {
		t.Run(tt.backend, func(t *testing.T) {
			config := svc.DefaultExporterConfig(tt.backend)
			assert.Equal(t, tt.expectedName, config.Name)
			assert.Equal(t, tt.expectedType, config.Type)
			assert.NotEmpty(t, config.Endpoint)
		})
	}
}

func TestOTelExporterService_ConvertTraceToOTel(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	now := time.Now()
	endTime := now.Add(5 * time.Second)
	trace := &domain.Trace{
		ID:        "abcdef1234567890abcdef1234567890",
		ProjectID: uuid.New(),
		Name:      "test-trace",
		StartTime: now,
		EndTime:   &endTime,
		Level:     "DEFAULT",
	}

	observations := []domain.Observation{
		{
			ID:        "obs1234567890abcdef",
			Type:      domain.ObservationTypeGeneration,
			Name:      "llm-call",
			StartTime: now,
			EndTime:   &endTime,
			Model:     "gpt-4",
			UsageDetails: domain.UsageDetails{
				InputTokens:  100,
				OutputTokens: 50,
			},
		},
	}

	result := svc.ConvertTraceToOTel(trace, observations, map[string]string{"env": "test"})
	require.NotNil(t, result)
	assert.Len(t, result.ScopeSpans, 1)
	assert.Len(t, result.ScopeSpans[0].Spans, 2) // root + 1 observation

	// Check resource attributes include custom ones
	assert.Equal(t, "test", result.Resource.Attributes["env"])
}

func TestOTelExporterService_QueueSpansForExport(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	t.Run("skips disabled exporter", func(t *testing.T) {
		exporter := &domain.OTelExporter{
			ID:           uuid.New(),
			Enabled:      false,
			SamplingRate: 1.0,
			BatchConfig:  domain.OTelBatchConfig{MaxBatchSize: 100},
		}
		err := svc.QueueSpansForExport(exporter, []*domain.OTelSpan{{}})
		assert.NoError(t, err)
	})

	t.Run("queues spans for enabled exporter", func(t *testing.T) {
		exporter := &domain.OTelExporter{
			ID:           uuid.New(),
			Enabled:      true,
			SamplingRate: 1.0,
			BatchConfig:  domain.OTelBatchConfig{MaxBatchSize: 100},
		}
		span := &domain.OTelSpan{
			TraceID: "abcdef1234567890abcdef1234567890",
			SpanID:  "abcdef1234567890",
			Name:    "test-span",
		}
		err := svc.QueueSpansForExport(exporter, []*domain.OTelSpan{span})
		assert.NoError(t, err)
	})
}

func TestOTelExporterService_GetHTTPClient_TLSErrors(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	t.Run("returns error for invalid cert file", func(t *testing.T) {
		exporter := &domain.OTelExporter{
			Timeout: 10,
			TLSConfig: &domain.OTelTLSConfig{
				CertFile: "/nonexistent/cert.pem",
				KeyFile:  "/nonexistent/key.pem",
			},
		}
		_, err := svc.getHTTPClient(exporter)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load client certificate")
	})

	t.Run("returns error for invalid CA file", func(t *testing.T) {
		exporter := &domain.OTelExporter{
			Timeout: 10,
			TLSConfig: &domain.OTelTLSConfig{
				CAFile: "/nonexistent/ca.pem",
			},
		}
		_, err := svc.getHTTPClient(exporter)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read CA file")
	})

	t.Run("succeeds with no TLS config", func(t *testing.T) {
		exporter := &domain.OTelExporter{Timeout: 10}
		client, err := svc.getHTTPClient(exporter)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("succeeds with insecure flag", func(t *testing.T) {
		exporter := &domain.OTelExporter{
			Timeout:  10,
			Insecure: true,
		}
		client, err := svc.getHTTPClient(exporter)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestOTelExporterService_HexToBytes(t *testing.T) {
	t.Run("valid hex", func(t *testing.T) {
		result, err := hexToBytes("abcdef1234567890", 8)
		require.NoError(t, err)
		assert.Len(t, result, 8)
	})

	t.Run("pads short hex", func(t *testing.T) {
		result, err := hexToBytes("ab", 8)
		require.NoError(t, err)
		assert.Len(t, result, 8)
	})

	t.Run("truncates long hex", func(t *testing.T) {
		result, err := hexToBytes("abcdef1234567890abcdef1234567890", 8)
		require.NoError(t, err)
		assert.Len(t, result, 8)
	})
}

func TestOTelExporterService_BatchErrorTracking(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	exporterID := uuid.New()
	assert.Equal(t, int64(0), svc.GetBatchErrorCount(exporterID))
}

func TestOTelExporterService_ConvertAttributesToProto(t *testing.T) {
	logger := zap.NewNop()
	svc := NewOTelExporterService(logger)
	defer svc.Stop()

	t.Run("converts various types", func(t *testing.T) {
		attrs := map[string]any{
			"str":   "value",
			"bool":  true,
			"int":   42,
			"int64": int64(100),
			"f64":   3.14,
			"other": struct{}{},
		}
		result := svc.convertAttributesToProto(attrs)
		assert.Len(t, result, 6)
	})

	t.Run("handles nil attrs", func(t *testing.T) {
		result := svc.convertAttributesToProto(nil)
		assert.Nil(t, result)
	})
}
