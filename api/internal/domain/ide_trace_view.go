package domain

import (
	"time"

	"github.com/google/uuid"
)

// FileTraceMapping represents the mapping of traces to a source file
type FileTraceMapping struct {
	FilePath    string           `json:"filePath"`
	ProjectID   uuid.UUID       `json:"projectId"`
	Annotations []LineAnnotation `json:"annotations"`
	Summary     FileTraceSummary `json:"summary"`
}

// LineAnnotation represents a trace annotation on a specific line of code
type LineAnnotation struct {
	Line       int       `json:"line"`
	TraceID    uuid.UUID `json:"traceId"`
	TraceName  string    `json:"traceName"`
	Type       string    `json:"type"` // created, modified, read
	AgentName  string    `json:"agentName"`
	Cost       float64   `json:"cost"`
	LatencyMs  int64     `json:"latencyMs"`
	Timestamp  time.Time `json:"timestamp"`
	Confidence float64   `json:"confidence"`
}

// FileTraceSummary provides a summary of trace activity for a file
type FileTraceSummary struct {
	TotalTraces        int        `json:"totalTraces"`
	TotalModifications int        `json:"totalModifications"`
	TopAgents          []string   `json:"topAgents"`
	TotalCost          float64    `json:"totalCost"`
	AvgLatencyMs       float64    `json:"avgLatencyMs"`
	LastModified       *time.Time `json:"lastModified,omitempty"`
}

// IDETraceContext represents detailed trace context for IDE display
type IDETraceContext struct {
	TraceID      uuid.UUID    `json:"traceId"`
	TraceName    string       `json:"traceName"`
	AgentSession string       `json:"agentSession"`
	Reasoning    string       `json:"reasoning"`
	Cost         float64      `json:"cost"`
	LatencyMs    float64      `json:"latencyMs"`
	FileChanges  []IDEFileChange `json:"fileChanges"`
}

// IDEFileChange represents a change made to a file by a trace
type IDEFileChange struct {
	Path        string `json:"path"`
	Operation   string `json:"operation"`
	DiffSummary string `json:"diffSummary"`
}
