package domain

import (
	"time"

	"github.com/google/uuid"
)

// OTelExporterType represents the type of OTLP exporter
type OTelExporterType string

const (
	OTelExporterTypeGRPC OTelExporterType = "grpc"
	OTelExporterTypeHTTP OTelExporterType = "http"
)

// OTelExporterStatus represents the status of an exporter
type OTelExporterStatus string

const (
	OTelExporterStatusActive   OTelExporterStatus = "active"
	OTelExporterStatusInactive OTelExporterStatus = "inactive"
	OTelExporterStatusError    OTelExporterStatus = "error"
)

// OTelExporter represents an OpenTelemetry exporter configuration
type OTelExporter struct {
	ID        uuid.UUID          `json:"id"`
	ProjectID uuid.UUID          `json:"projectId"`
	Name      string             `json:"name"`
	Enabled   bool               `json:"enabled"`
	Type      OTelExporterType   `json:"type"`
	Status    OTelExporterStatus `json:"status"`

	// Connection settings
	Endpoint    string            `json:"endpoint"`
	Headers     map[string]string `json:"headers,omitempty"`
	Compression string            `json:"compression,omitempty"` // gzip, none
	Timeout     int               `json:"timeoutSeconds"`
	Insecure    bool              `json:"insecure"` // Skip TLS verification

	// TLS settings
	TLSConfig *OTelTLSConfig `json:"tlsConfig,omitempty"`

	// Batching settings
	BatchConfig OTelBatchConfig `json:"batchConfig"`

	// Retry settings
	RetryConfig OTelRetryConfig `json:"retryConfig"`

	// Resource attributes to add
	ResourceAttributes map[string]string `json:"resourceAttributes,omitempty"`

	// Filtering
	TraceNameFilter   string            `json:"traceNameFilter,omitempty"`
	MetadataFilters   map[string]string `json:"metadataFilters,omitempty"`
	SamplingRate      float64           `json:"samplingRate"` // 0-1, percentage of traces to export

	// Stats
	LastExportAt    *time.Time `json:"lastExportAt,omitempty"`
	ExportedCount   int64      `json:"exportedCount"`
	ErrorCount      int64      `json:"errorCount"`
	LastError       string     `json:"lastError,omitempty"`
	LastErrorAt     *time.Time `json:"lastErrorAt,omitempty"`

	// Audit
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy uuid.UUID `json:"createdBy"`
}

// OTelTLSConfig represents TLS configuration for OTLP exporters
type OTelTLSConfig struct {
	CertFile   string `json:"certFile,omitempty"`
	KeyFile    string `json:"keyFile,omitempty"`
	CAFile     string `json:"caFile,omitempty"`
	ServerName string `json:"serverName,omitempty"`
}

// OTelBatchConfig represents batching configuration
type OTelBatchConfig struct {
	MaxBatchSize      int `json:"maxBatchSize"`      // Max spans per batch
	MaxQueueSize      int `json:"maxQueueSize"`      // Max spans in queue
	BatchTimeout      int `json:"batchTimeoutMs"`    // Max time to wait before sending
	ExportTimeout     int `json:"exportTimeoutMs"`   // Timeout for export request
	ScheduleDelay     int `json:"scheduleDelayMs"`   // Delay between batch checks
}

// OTelRetryConfig represents retry configuration
type OTelRetryConfig struct {
	Enabled         bool `json:"enabled"`
	InitialInterval int  `json:"initialIntervalMs"`
	MaxInterval     int  `json:"maxIntervalMs"`
	MaxElapsedTime  int  `json:"maxElapsedTimeMs"`
	Multiplier      float64 `json:"multiplier"`
}

// OTelExporterInput represents input for creating/updating an exporter
type OTelExporterInput struct {
	Name               string             `json:"name" validate:"required,min=1,max=100"`
	Enabled            *bool              `json:"enabled,omitempty"`
	Type               OTelExporterType   `json:"type" validate:"required"`
	Endpoint           string             `json:"endpoint" validate:"required,url"`
	Headers            map[string]string  `json:"headers,omitempty"`
	Compression        string             `json:"compression,omitempty"`
	Timeout            *int               `json:"timeoutSeconds,omitempty"`
	Insecure           *bool              `json:"insecure,omitempty"`
	TLSConfig          *OTelTLSConfig     `json:"tlsConfig,omitempty"`
	BatchConfig        *OTelBatchConfig   `json:"batchConfig,omitempty"`
	RetryConfig        *OTelRetryConfig   `json:"retryConfig,omitempty"`
	ResourceAttributes map[string]string  `json:"resourceAttributes,omitempty"`
	TraceNameFilter    string             `json:"traceNameFilter,omitempty"`
	MetadataFilters    map[string]string  `json:"metadataFilters,omitempty"`
	SamplingRate       *float64           `json:"samplingRate,omitempty"`
}

// OTelExporterList represents a list of exporters
type OTelExporterList struct {
	Exporters  []OTelExporter `json:"exporters"`
	TotalCount int64          `json:"totalCount"`
}

// OTelSpan represents an OpenTelemetry span for export
type OTelSpan struct {
	TraceID           string            `json:"traceId"`
	SpanID            string            `json:"spanId"`
	ParentSpanID      string            `json:"parentSpanId,omitempty"`
	Name              string            `json:"name"`
	Kind              OTelSpanKind      `json:"kind"`
	StartTimeUnixNano int64             `json:"startTimeUnixNano"`
	EndTimeUnixNano   int64             `json:"endTimeUnixNano"`
	Attributes        map[string]any    `json:"attributes,omitempty"`
	Status            OTelSpanStatus    `json:"status"`
	Events            []OTelSpanEvent   `json:"events,omitempty"`
	Links             []OTelSpanLink    `json:"links,omitempty"`
}

// OTelSpanKind represents the span kind
type OTelSpanKind int

const (
	OTelSpanKindUnspecified OTelSpanKind = 0
	OTelSpanKindInternal    OTelSpanKind = 1
	OTelSpanKindServer      OTelSpanKind = 2
	OTelSpanKindClient      OTelSpanKind = 3
	OTelSpanKindProducer    OTelSpanKind = 4
	OTelSpanKindConsumer    OTelSpanKind = 5
)

// OTelSpanStatus represents the span status
type OTelSpanStatus struct {
	Code    OTelStatusCode `json:"code"`
	Message string         `json:"message,omitempty"`
}

// OTelStatusCode represents the status code
type OTelStatusCode int

const (
	OTelStatusCodeUnset OTelStatusCode = 0
	OTelStatusCodeOK    OTelStatusCode = 1
	OTelStatusCodeError OTelStatusCode = 2
)

// OTelSpanEvent represents a span event
type OTelSpanEvent struct {
	Name              string         `json:"name"`
	TimeUnixNano      int64          `json:"timeUnixNano"`
	Attributes        map[string]any `json:"attributes,omitempty"`
}

// OTelSpanLink represents a link to another span
type OTelSpanLink struct {
	TraceID    string         `json:"traceId"`
	SpanID     string         `json:"spanId"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// OTelResource represents resource information
type OTelResource struct {
	Attributes map[string]any `json:"attributes"`
}

// OTelScope represents instrumentation scope
type OTelScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// OTelResourceSpans represents spans grouped by resource
type OTelResourceSpans struct {
	Resource   OTelResource    `json:"resource"`
	ScopeSpans []OTelScopeSpans `json:"scopeSpans"`
}

// OTelScopeSpans represents spans grouped by scope
type OTelScopeSpans struct {
	Scope OTelScope  `json:"scope"`
	Spans []OTelSpan `json:"spans"`
}

// OTelExportRequest represents an OTLP export request
type OTelExportRequest struct {
	ResourceSpans []OTelResourceSpans `json:"resourceSpans"`
}

// OTelExportResponse represents an OTLP export response
type OTelExportResponse struct {
	PartialSuccess *OTelExportPartialSuccess `json:"partialSuccess,omitempty"`
}

// OTelExportPartialSuccess represents partial success info
type OTelExportPartialSuccess struct {
	RejectedSpans int64  `json:"rejectedSpans"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}

// OTelReceiverConfig represents configuration for receiving OTLP data
type OTelReceiverConfig struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	Enabled   bool      `json:"enabled"`

	// GRPC settings
	GRPCEnabled bool   `json:"grpcEnabled"`
	GRPCPort    int    `json:"grpcPort"`

	// HTTP settings
	HTTPEnabled bool   `json:"httpEnabled"`
	HTTPPort    int    `json:"httpPort"`
	HTTPPath    string `json:"httpPath"` // Default: /v1/traces

	// Authentication
	RequireAuth bool     `json:"requireAuth"`
	APIKeys     []string `json:"apiKeys,omitempty"` // Allowed API keys

	// Processing
	MaxBatchSize int `json:"maxBatchSize"`

	// Audit
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OTelExporterStats represents export statistics
type OTelExporterStats struct {
	ExporterID      uuid.UUID `json:"exporterId"`
	ExporterName    string    `json:"exporterName"`
	TotalExported   int64     `json:"totalExported"`
	TotalErrors     int64     `json:"totalErrors"`
	LastExportAt    *time.Time `json:"lastExportAt,omitempty"`
	LastErrorAt     *time.Time `json:"lastErrorAt,omitempty"`
	AvgLatencyMs    float64   `json:"avgLatencyMs"`
	ExportsLast24h  int64     `json:"exportsLast24h"`
	ErrorsLast24h   int64     `json:"errorsLast24h"`
}

// Standard OpenTelemetry semantic conventions for LLM observability
const (
	// LLM attributes (following OpenTelemetry semantic conventions)
	OTelAttrLLMSystem         = "gen_ai.system"           // e.g., "openai", "anthropic"
	OTelAttrLLMRequestModel   = "gen_ai.request.model"    // e.g., "gpt-4"
	OTelAttrLLMResponseModel  = "gen_ai.response.model"   // Actual model used
	OTelAttrLLMRequestMaxTokens = "gen_ai.request.max_tokens"
	OTelAttrLLMRequestTemperature = "gen_ai.request.temperature"
	OTelAttrLLMRequestTopP    = "gen_ai.request.top_p"
	OTelAttrLLMUsageInputTokens = "gen_ai.usage.input_tokens"
	OTelAttrLLMUsageOutputTokens = "gen_ai.usage.output_tokens"
	OTelAttrLLMResponseFinishReason = "gen_ai.response.finish_reasons"

	// AgentTrace custom attributes
	OTelAttrAgentTraceTraceID    = "agenttrace.trace.id"
	OTelAttrAgentTraceSpanID     = "agenttrace.span.id"
	OTelAttrAgentTraceProjectID  = "agenttrace.project.id"
	OTelAttrAgentTraceTraceName  = "agenttrace.trace.name"
	OTelAttrAgentTraceSpanType   = "agenttrace.span.type"
	OTelAttrAgentTraceCost       = "agenttrace.cost"
	OTelAttrAgentTraceLatencyMs  = "agenttrace.latency_ms"

	// Service attributes
	OTelAttrServiceName    = "service.name"
	OTelAttrServiceVersion = "service.version"
	OTelAttrDeploymentEnv  = "deployment.environment"
)
