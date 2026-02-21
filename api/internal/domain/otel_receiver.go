package domain

import "time"

// OTLPTraceRequest represents an incoming OTLP trace request
type OTLPTraceRequest struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

// ResourceSpan groups spans by their originating resource
type ResourceSpan struct {
	Resource   OTLPResource `json:"resource"`
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

// OTLPResource describes the entity producing telemetry
type OTLPResource struct {
	Attributes  map[string]any `json:"attributes,omitempty"`
	ServiceName string         `json:"serviceName,omitempty"`
}

// ScopeSpan groups spans by instrumentation scope
type ScopeSpan struct {
	Scope InstrumentationScope `json:"scope"`
	Spans []OTLPSpan           `json:"spans"`
}

// InstrumentationScope identifies the library producing telemetry
type InstrumentationScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// SpanKind represents the role of a span in a trace
type SpanKind int

const (
	SpanKindInternal SpanKind = 0
	SpanKindServer   SpanKind = 1
	SpanKindClient   SpanKind = 2
	SpanKindProducer SpanKind = 3
	SpanKindConsumer SpanKind = 4
)

// SpanStatus represents the status of a span
type SpanStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// SpanEvent represents a timed event within a span
type SpanEvent struct {
	Name       string         `json:"name"`
	Timestamp  time.Time      `json:"timestamp"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// OTLPSpan represents an incoming OTLP span
type OTLPSpan struct {
	TraceID      string         `json:"traceId"`
	SpanID       string         `json:"spanId"`
	ParentSpanID string         `json:"parentSpanId,omitempty"`
	Name         string         `json:"name"`
	Kind         SpanKind       `json:"kind"`
	StartTime    uint64         `json:"startTime"`    // Unix nanoseconds (from protobuf) or 0
	EndTime      uint64         `json:"endTime"`      // Unix nanoseconds (from protobuf) or 0
	Attributes   map[string]any `json:"attributes,omitempty"`
	Status       SpanStatus     `json:"status"`
	Events       []SpanEvent    `json:"events,omitempty"`
}

// StartTimeAsTime converts the StartTime nanoseconds to time.Time.
func (s OTLPSpan) StartTimeAsTime() time.Time {
	if s.StartTime == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(s.StartTime))
}

// EndTimeAsTime converts the EndTime nanoseconds to time.Time.
func (s OTLPSpan) EndTimeAsTime() time.Time {
	if s.EndTime == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(s.EndTime))
}

// OTLPTraceResponse represents the response to an OTLP trace request
type OTLPTraceResponse struct {
	PartialSuccess *PartialSuccess `json:"partialSuccess,omitempty"`
}

// PartialSuccess indicates that some spans were rejected
type PartialSuccess struct {
	RejectedSpans int64  `json:"rejectedSpans"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}
