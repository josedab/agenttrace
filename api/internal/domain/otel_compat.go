package domain

import (
	"time"

	"github.com/google/uuid"
)

// OTelExportFormat represents the export format for OTel destinations
type OTelExportFormat string

const (
	OTelExportFormatOTLPgRPC OTelExportFormat = "otlp_grpc"
	OTelExportFormatOTLPHTTP OTelExportFormat = "otlp_http"
	OTelExportFormatJaeger   OTelExportFormat = "jaeger"
	OTelExportFormatZipkin   OTelExportFormat = "zipkin"
)

// IsValid checks if the OTel export format is valid
func (f OTelExportFormat) IsValid() bool {
	switch f {
	case OTelExportFormatOTLPgRPC, OTelExportFormatOTLPHTTP, OTelExportFormatJaeger, OTelExportFormatZipkin:
		return true
	}
	return false
}

// OTelSemanticVersion represents the OTel semantic convention version
type OTelSemanticVersion string

const (
	OTelSemanticVersionV1_27  OTelSemanticVersion = "v1_27"
	OTelSemanticVersionV1_28  OTelSemanticVersion = "v1_28"
	OTelSemanticVersionLatest OTelSemanticVersion = "latest"
)

// IsValid checks if the OTel semantic version is valid
func (v OTelSemanticVersion) IsValid() bool {
	switch v {
	case OTelSemanticVersionV1_27, OTelSemanticVersionV1_28, OTelSemanticVersionLatest:
		return true
	}
	return false
}

// OTelExportDestination represents a configured OTel export destination
type OTelExportDestination struct {
	ID              uuid.UUID         `json:"id"`
	ProjectID       uuid.UUID         `json:"projectId"`
	Name            string            `json:"name"`
	Format          OTelExportFormat  `json:"format"`
	Endpoint        string            `json:"endpoint"`
	Headers         map[string]string `json:"headers,omitempty"`
	Enabled         bool              `json:"enabled"`
	TLSEnabled      bool              `json:"tlsEnabled"`
	SamplingRate    float64           `json:"samplingRate"`
	BatchSize       int               `json:"batchSize"`
	FlushIntervalMs int               `json:"flushIntervalMs"`
	LastExportAt    *time.Time        `json:"lastExportAt,omitempty"`
	ExportedCount   int64             `json:"exportedCount"`
	ErrorCount      int64             `json:"errorCount"`
	CreatedAt       time.Time         `json:"createdAt"`
}

// OTelMapping represents a mapping between AgentTrace fields and OTel attributes
type OTelMapping struct {
	AgentTraceField string `json:"agentTraceField"`
	OTelAttribute   string `json:"otelAttribute"`
	OTelNamespace   string `json:"otelNamespace"`
	Transform       string `json:"transform,omitempty"`
}

// OTelCollectorConfig represents a generated OTel collector configuration
type OTelCollectorConfig struct {
	ID             uuid.UUID              `json:"id"`
	ProjectID      uuid.UUID              `json:"projectId"`
	Receivers      []OTelCompatReceiverConfig `json:"receivers"`
	Exporters      []string               `json:"exporters"`
	Processors     []string               `json:"processors,omitempty"`
	PipelineConfig map[string]interface{} `json:"pipelineConfig,omitempty"`
	GeneratedAt    time.Time              `json:"generatedAt"`
}

// OTelCompatReceiverConfig represents configuration for an OTel receiver in the compat layer
type OTelCompatReceiverConfig struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Endpoint string `json:"endpoint"`
	Enabled  bool   `json:"enabled"`
}

// OTelCompatDashboard represents the OTel compatibility overview dashboard
type OTelCompatDashboard struct {
	ImportedTraces     int64                   `json:"importedTraces"`
	ExportedTraces     int64                   `json:"exportedTraces"`
	ActiveDestinations int                     `json:"activeDestinations"`
	SemanticVersion    OTelSemanticVersion     `json:"semanticVersion"`
	MappingCoverage    float64                 `json:"mappingCoverage"`
	Destinations       []OTelExportDestination `json:"destinations"`
}
