package domain

import (
	"time"
)

// DistributedSpanKind represents the kind of a distributed span
type DistributedSpanKind string

const (
	DistributedSpanKindClient   DistributedSpanKind = "client"
	DistributedSpanKindServer   DistributedSpanKind = "server"
	DistributedSpanKindInternal DistributedSpanKind = "internal"
)

// DistributedSpan represents a span in a distributed trace
type DistributedSpan struct {
	SpanID        string            `json:"spanId"`
	ParentSpanID  string            `json:"parentSpanId,omitempty"`
	ServiceName   string            `json:"serviceName"`
	OperationName string            `json:"operationName"`
	StartTime     time.Time         `json:"startTime"`
	Duration      time.Duration     `json:"duration"`
	Status        string            `json:"status"`
	SpanKind      DistributedSpanKind `json:"spanKind"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// DistributedTrace represents a full distributed trace across services
type DistributedTrace struct {
	TraceID       string             `json:"traceId"`
	AgentSpans    []DistributedSpan  `json:"agentSpans"`
	ServiceSpans  []DistributedSpan  `json:"serviceSpans"`
	TotalServices int                `json:"totalServices"`
	CriticalPath  []string           `json:"criticalPath"`
	Bottleneck    *DistributedSpan   `json:"bottleneck,omitempty"`
}

// ServiceNodeType represents the type of a service node
type ServiceNodeType string

const (
	ServiceNodeTypeAgent    ServiceNodeType = "agent"
	ServiceNodeTypeAPI      ServiceNodeType = "api"
	ServiceNodeTypeDatabase ServiceNodeType = "database"
	ServiceNodeTypeCache    ServiceNodeType = "cache"
	ServiceNodeTypeExternal ServiceNodeType = "external"
)

// ServiceNode represents a service in the service map
type ServiceNode struct {
	Name         string          `json:"name"`
	Type         ServiceNodeType `json:"type"`
	RequestCount int             `json:"requestCount"`
	AvgLatencyMs float64         `json:"avgLatencyMs"`
	ErrorRate    float64         `json:"errorRate"`
}

// ServiceConnection represents a connection between services
type ServiceConnection struct {
	From         string  `json:"from"`
	To           string  `json:"to"`
	RequestCount int     `json:"requestCount"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
}

// ServiceMap represents the topology of services
type ServiceMap struct {
	Services    []ServiceNode       `json:"services"`
	Connections []ServiceConnection `json:"connections"`
}

// TraceCorrelationInput is the input for correlating traces
type TraceCorrelationInput struct {
	TraceID          string   `json:"traceId"`
	ExternalTraceIDs []string `json:"externalTraceIds"`
}
