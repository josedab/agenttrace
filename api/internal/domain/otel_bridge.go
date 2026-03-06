package domain

import (
	"time"

	"github.com/google/uuid"
)

// OTelBridgeConfig represents the configuration for an OpenTelemetry bridge
type OTelBridgeConfig struct {
	ID                 uuid.UUID            `json:"id"`
	ProjectID          uuid.UUID            `json:"projectId"`
	ExportEnabled      bool                 `json:"exportEnabled"`
	ImportEnabled      bool                 `json:"importEnabled"`
	ExportDestinations []ExportDestinationRef `json:"exportDestinations"`
	ImportMappings     []ImportMapping      `json:"importMappings"`
	ResourceAttributes map[string]string    `json:"resourceAttributes"`
	SamplingRate       float64              `json:"samplingRate"`
	CreatedAt          time.Time            `json:"createdAt"`
	UpdatedAt          time.Time            `json:"updatedAt"`
}

// ExportDestinationRef represents a reference to an export destination
type ExportDestinationRef struct {
	ID       uuid.UUID         `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Endpoint string            `json:"endpoint"`
	Protocol string            `json:"protocol"`
	Enabled  bool              `json:"enabled"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// ImportMapping represents a mapping for importing OpenTelemetry data
type ImportMapping struct {
	ID               uuid.UUID         `json:"id"`
	SourceService    string            `json:"sourceService"`
	AttributeMapping map[string]string `json:"attributeMapping"`
	SpanNameTemplate string            `json:"spanNameTemplate"`
	Enabled          bool              `json:"enabled"`
}

// OTelImportRequest represents a request to import OpenTelemetry data
type OTelImportRequest struct {
	ResourceSpans       []any `json:"resourceSpans"`
	CorrelateByTraceID  bool  `json:"correlateByTraceId"`
	CreateMissingTraces bool  `json:"createMissingTraces"`
}

// OTelBridgeStats represents statistics for the OpenTelemetry bridge
type OTelBridgeStats struct {
	ExportStats BridgeDirectionStats `json:"exportStats"`
	ImportStats BridgeDirectionStats `json:"importStats"`
	LastSync    time.Time            `json:"lastSync"`
}

// BridgeDirectionStats represents statistics for one direction of the bridge
type BridgeDirectionStats struct {
	TotalSpans   int64   `json:"totalSpans"`
	SuccessCount int64   `json:"successCount"`
	ErrorCount   int64   `json:"errorCount"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	Last24hCount int64   `json:"last24hCount"`
}

// OTelBridgeConfigInput represents input for updating the bridge configuration
type OTelBridgeConfigInput struct {
	ExportEnabled      *bool             `json:"exportEnabled,omitempty"`
	ImportEnabled      *bool             `json:"importEnabled,omitempty"`
	ResourceAttributes map[string]string `json:"resourceAttributes,omitempty"`
	SamplingRate       *float64          `json:"samplingRate,omitempty"`
}

// OTelDestinationInput represents input for creating or updating an export destination
type OTelDestinationInput struct {
	Name     string            `json:"name" validate:"required"`
	Type     string            `json:"type" validate:"required"`
	Endpoint string            `json:"endpoint" validate:"required"`
	Protocol string            `json:"protocol,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}
