package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func newOTelReceiverService() *OTelReceiverService {
	return NewOTelReceiverService(zap.NewNop(), nil)
}

func TestOTelReceiverService_MapResourceToTrace(t *testing.T) {
	svc := newOTelReceiverService()

	t.Run("maps service name from ServiceName field", func(t *testing.T) {
		resource := &domain.OTLPResource{
			ServiceName: "my-agent-service",
			Attributes:  map[string]any{"env": "production"},
		}

		traceInput := svc.MapResourceToTrace(resource)

		assert.Equal(t, "my-agent-service", traceInput.Name)
		assert.Equal(t, "production", traceInput.Metadata.(map[string]any)["env"])
	})

	t.Run("maps service name from attributes when ServiceName empty", func(t *testing.T) {
		resource := &domain.OTLPResource{
			Attributes: map[string]any{
				domain.OTelAttrServiceName: "attr-service",
			},
		}

		traceInput := svc.MapResourceToTrace(resource)

		assert.Equal(t, "attr-service", traceInput.Name)
	})

	t.Run("handles empty resource", func(t *testing.T) {
		resource := &domain.OTLPResource{}

		traceInput := svc.MapResourceToTrace(resource)

		assert.Empty(t, traceInput.Name)
		assert.Nil(t, traceInput.Metadata)
	})
}

func TestOTelReceiverService_MapSpanToObservation(t *testing.T) {
	svc := newOTelReceiverService()

	t.Run("maps basic span fields", func(t *testing.T) {
		now := uint64(time.Now().UnixNano())
		span := &domain.OTLPSpan{
			TraceID:   "trace-abc",
			SpanID:    "span-123",
			Name:      "llm-call",
			StartTime: now,
			EndTime:   now + uint64(2*time.Second),
		}

		obs := svc.MapSpanToObservation(span)

		require.NotNil(t, obs)
		assert.Equal(t, "span-123", *obs.ID)
		assert.Equal(t, "trace-abc", *obs.TraceID)
		assert.Equal(t, "llm-call", *obs.Name)
		assert.Equal(t, domain.ObservationTypeSpan, *obs.Type)
	})

	t.Run("detects generation from gen_ai.request.model attribute", func(t *testing.T) {
		span := &domain.OTLPSpan{
			TraceID:   "trace-abc",
			SpanID:    "span-gen",
			Name:      "openai-call",
			StartTime: uint64(time.Now().UnixNano()),
			EndTime:   uint64(time.Now().Add(time.Second).UnixNano()),
			Attributes: map[string]any{
				domain.OTelAttrLLMRequestModel: "gpt-4",
			},
		}

		obs := svc.MapSpanToObservation(span)

		require.NotNil(t, obs)
		assert.Equal(t, domain.ObservationTypeGeneration, *obs.Type)
		assert.Equal(t, "gpt-4", *obs.Model)
	})

	t.Run("detects generation from gen_ai.system attribute", func(t *testing.T) {
		span := &domain.OTLPSpan{
			TraceID:   "trace-abc",
			SpanID:    "span-sys",
			Name:      "anthropic-call",
			StartTime: uint64(time.Now().UnixNano()),
			EndTime:   uint64(time.Now().Add(time.Second).UnixNano()),
			Attributes: map[string]any{
				domain.OTelAttrLLMSystem: "anthropic",
			},
		}

		obs := svc.MapSpanToObservation(span)

		require.NotNil(t, obs)
		assert.Equal(t, domain.ObservationTypeGeneration, *obs.Type)
	})

	t.Run("sets parent observation ID from parent span", func(t *testing.T) {
		span := &domain.OTLPSpan{
			TraceID:      "trace-abc",
			SpanID:       "span-child",
			ParentSpanID: "span-parent",
			Name:         "child-span",
			StartTime:    uint64(time.Now().UnixNano()),
			EndTime:      uint64(time.Now().Add(time.Second).UnixNano()),
		}

		obs := svc.MapSpanToObservation(span)

		require.NotNil(t, obs)
		require.NotNil(t, obs.ParentObservationID)
		assert.Equal(t, "span-parent", *obs.ParentObservationID)
	})

	t.Run("maps error status", func(t *testing.T) {
		span := &domain.OTLPSpan{
			TraceID:   "trace-abc",
			SpanID:    "span-err",
			Name:      "error-span",
			StartTime: uint64(time.Now().UnixNano()),
			EndTime:   uint64(time.Now().Add(time.Second).UnixNano()),
			Status: domain.SpanStatus{
				Code:    2, // ERROR
				Message: "something failed",
			},
		}

		obs := svc.MapSpanToObservation(span)

		require.NotNil(t, obs)
		require.NotNil(t, obs.Level)
		assert.Equal(t, domain.LevelError, *obs.Level)
		require.NotNil(t, obs.StatusMessage)
		assert.Equal(t, "something failed", *obs.StatusMessage)
	})

	t.Run("returns nil for nil span", func(t *testing.T) {
		obs := svc.MapSpanToObservation(nil)
		assert.Nil(t, obs)
	})
}

func TestOTelReceiverService_ReceiveTraces(t *testing.T) {
	t.Run("handles empty request gracefully", func(t *testing.T) {
		svc := newOTelReceiverService()

		request := &domain.OTLPTraceRequest{
			ResourceSpans: []domain.ResourceSpan{},
		}

		resp, err := svc.ReceiveTraces(context.Background(), uuid.New(), request)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Nil(t, resp.PartialSuccess)
	})

	t.Run("handles request with no spans in scope", func(t *testing.T) {
		svc := newOTelReceiverService()

		request := &domain.OTLPTraceRequest{
			ResourceSpans: []domain.ResourceSpan{
				{
					Resource: domain.OTLPResource{ServiceName: "test-svc"},
					ScopeSpans: []domain.ScopeSpan{
						{Spans: []domain.OTLPSpan{}},
					},
				},
			},
		}

		resp, err := svc.ReceiveTraces(context.Background(), uuid.New(), request)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Nil(t, resp.PartialSuccess)
	})
}
