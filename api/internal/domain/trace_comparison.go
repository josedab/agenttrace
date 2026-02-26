package domain

import (
	"time"

	"github.com/google/uuid"
)

// TraceComparisonInput represents input for comparing multiple traces
type TraceComparisonInput struct {
	TraceIDs []string `json:"traceIds" validate:"required,min=2,max=10"`
	Metrics  []string `json:"metrics,omitempty"` // Optional: specific metrics to compare
}

// TraceComparisonMatrix represents a structured comparison of 2-N traces
type TraceComparisonMatrix struct {
	ID          uuid.UUID               `json:"id"`
	ProjectID   uuid.UUID               `json:"projectId"`
	TraceIDs    []string                `json:"traceIds"`
	Traces      []TraceComparisonEntry  `json:"traces"`
	Metrics     MetricComparisonGrid    `json:"metrics"`
	ToolUsage   ToolUsageComparison     `json:"toolUsage"`
	Topology    TopologyComparison      `json:"topology"`
	ShareToken  string                  `json:"shareToken,omitempty"`
	CreatedAt   time.Time               `json:"createdAt"`
	CreatedBy   uuid.UUID               `json:"createdBy"`
}

// TraceComparisonEntry represents a single trace in the comparison
type TraceComparisonEntry struct {
	TraceID     string            `json:"traceId"`
	Name        string            `json:"name"`
	StartTime   time.Time         `json:"startTime"`
	Duration    int64             `json:"durationMs"`
	TotalCost   float64           `json:"totalCost"`
	TotalTokens int               `json:"totalTokens"`
	SpanCount   int               `json:"spanCount"`
	ErrorCount  int               `json:"errorCount"`
	Level       string            `json:"level"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ToolCalls   []string          `json:"toolCalls"`
}

// MetricComparisonGrid provides a grid view of metrics across traces
type MetricComparisonGrid struct {
	Latency       MetricRow `json:"latency"`
	TotalCost     MetricRow `json:"totalCost"`
	TotalTokens   MetricRow `json:"totalTokens"`
	PromptTokens  MetricRow `json:"promptTokens"`
	OutputTokens  MetricRow `json:"outputTokens"`
	SpanCount     MetricRow `json:"spanCount"`
	ErrorRate     MetricRow `json:"errorRate"`
	QualityScore  MetricRow `json:"qualityScore,omitempty"`
}

// MetricRow represents a single metric across all compared traces
type MetricRow struct {
	Name    string                    `json:"name"`
	Unit    string                    `json:"unit"`
	Values  []MetricValue             `json:"values"`
	Stats   MetricComparisonStats     `json:"stats"`
}

// MetricValue represents a metric value for a specific trace
type MetricValue struct {
	TraceID string  `json:"traceId"`
	Value   float64 `json:"value"`
	Rank    int     `json:"rank"`
}

// MetricComparisonStats provides statistics across compared values
type MetricComparisonStats struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Avg       float64 `json:"avg"`
	Range     float64 `json:"range"`
	BestTrace string  `json:"bestTrace"` // Trace ID with best value
}

// ToolUsageComparison compares tool usage across traces
type ToolUsageComparison struct {
	AllTools     []string                      `json:"allTools"`
	ByTrace      map[string][]ToolCallSummary  `json:"byTrace"`
	Divergences  []ToolDivergence              `json:"divergences"`
}

// ToolCallSummary summarizes tool usage in a trace
type ToolCallSummary struct {
	ToolName   string  `json:"toolName"`
	CallCount  int     `json:"callCount"`
	TotalTime  int64   `json:"totalTimeMs"`
	AvgTime    float64 `json:"avgTimeMs"`
}

// ToolDivergence represents a difference in tool usage between traces
type ToolDivergence struct {
	ToolName    string   `json:"toolName"`
	Type        string   `json:"type"` // "missing_in", "count_diff", "time_diff"
	Description string   `json:"description"`
	TraceIDs    []string `json:"traceIds"`
}

// TopologyComparison compares the structure of trace spans
type TopologyComparison struct {
	CommonSpans    []string         `json:"commonSpans"`
	UniqueSpans    map[string][]string `json:"uniqueSpans"` // traceID -> span names unique to that trace
	StructureDiffs []StructureDiff  `json:"structureDiffs"`
}

// StructureDiff represents a structural difference between traces
type StructureDiff struct {
	Type        string   `json:"type"` // "depth_diff", "branching_diff", "order_diff"
	Description string   `json:"description"`
	TraceIDs    []string `json:"traceIds"`
}

// TraceComparisonExport represents an exportable comparison
type TraceComparisonExport struct {
	Format   string `json:"format"` // "pdf", "csv", "json"
	Data     []byte `json:"data"`
	Filename string `json:"filename"`
}
