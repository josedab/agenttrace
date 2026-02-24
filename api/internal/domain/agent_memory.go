package domain

import (
	"time"

	"github.com/google/uuid"
)

// MemoryItemType represents the type of a memory item
type MemoryItemType string

const (
	MemoryItemTypeSystem     MemoryItemType = "system"
	MemoryItemTypeUser       MemoryItemType = "user"
	MemoryItemTypeAssistant  MemoryItemType = "assistant"
	MemoryItemTypeToolResult MemoryItemType = "tool_result"
	MemoryItemTypeContext    MemoryItemType = "context"
)

// MemoryItem represents a single item in the agent's context window
type MemoryItem struct {
	Type       MemoryItemType `json:"type"`
	Content    string         `json:"content"`
	TokenCount int            `json:"tokenCount"`
	Retained   bool           `json:"retained"`
}

// MemorySnapshot captures the state of an agent's context window at a point in time
type MemorySnapshot struct {
	ID             uuid.UUID    `json:"id"`
	TraceID        uuid.UUID    `json:"traceId"`
	ProjectID      uuid.UUID    `json:"projectId"`
	StepIndex      int          `json:"stepIndex"`
	TotalTokens    int          `json:"totalTokens"`
	UsedTokens     int          `json:"usedTokens"`
	AvailableTokens int         `json:"availableTokens"`
	RetainedItems  []MemoryItem `json:"retainedItems"`
	TruncatedItems []MemoryItem `json:"truncatedItems"`
	UtilizationPct float64      `json:"utilizationPct"`
	Timestamp      time.Time    `json:"timestamp"`
}

// MemoryTimeline tracks context window utilization across an agent trace
type MemoryTimeline struct {
	TraceID                 uuid.UUID             `json:"traceId"`
	Snapshots               []MemorySnapshot      `json:"snapshots"`
	PeakUtilization         float64               `json:"peakUtilization"`
	AvgUtilization          float64               `json:"avgUtilization"`
	TruncationEvents        int                   `json:"truncationEvents"`
	OptimizationSuggestions []string              `json:"optimizationSuggestions"`
}

// MemoryOptimizationTechnique represents an optimization approach
type MemoryOptimizationTechnique string

const (
	MemoryOptCompression   MemoryOptimizationTechnique = "compression"
	MemoryOptSummarization MemoryOptimizationTechnique = "summarization"
	MemoryOptRAG           MemoryOptimizationTechnique = "rag"
	MemoryOptSlidingWindow MemoryOptimizationTechnique = "sliding_window"
)

// MemoryOptimization describes a potential optimization for context window usage
type MemoryOptimization struct {
	Technique           MemoryOptimizationTechnique `json:"technique"`
	Description         string                      `json:"description"`
	EstimatedSavingsPct float64                     `json:"estimatedSavingsPct"`
	Confidence          float64                     `json:"confidence"`
}

// MemoryAnalysisInput is the input for memory analysis
type MemoryAnalysisInput struct {
	TraceID string `json:"traceId"`
}
