package domain

import (
	"time"

	"github.com/google/uuid"
)

// DebugQueryType represents the type of AI debug query
type DebugQueryType string

const (
	DebugQueryTypeRootCause  DebugQueryType = "root_cause"
	DebugQueryTypeExplain    DebugQueryType = "explain"
	DebugQueryTypeSuggestFix DebugQueryType = "suggest_fix"
	DebugQueryTypeCompare    DebugQueryType = "compare"
	DebugQueryTypeOptimize   DebugQueryType = "optimize"
)

// IsValid checks if the debug query type is valid
func (t DebugQueryType) IsValid() bool {
	switch t {
	case DebugQueryTypeRootCause, DebugQueryTypeExplain, DebugQueryTypeSuggestFix, DebugQueryTypeCompare, DebugQueryTypeOptimize:
		return true
	}
	return false
}

// DebugQuery represents an AI-powered debug query against a trace
type DebugQuery struct {
	ID        uuid.UUID      `json:"id"`
	ProjectID uuid.UUID      `json:"projectId"`
	TraceID   uuid.UUID      `json:"traceId"`
	Query     string         `json:"query"`
	QueryType DebugQueryType `json:"queryType"`
	Context   DebugContext   `json:"context"`
	Response  DebugResponse  `json:"response"`
	CreatedAt time.Time      `json:"createdAt"`
	CreatedBy uuid.UUID      `json:"createdBy"`
}

// DebugContext represents the context provided to the AI debugger
type DebugContext struct {
	TraceSummary   string             `json:"traceSummary"`
	Observations   []DebugObservation `json:"observations"`
	FileOperations []string           `json:"fileOperations,omitempty"`
	Errors         []string           `json:"errors,omitempty"`
	GitContext     *DebugGitContext   `json:"gitContext,omitempty"`
	CostTotal      float64           `json:"costTotal"`
	DurationMs     int64             `json:"durationMs"`
}

// DebugObservation represents a summarized observation for debug context
type DebugObservation struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Model      string `json:"model,omitempty"`
	DurationMs int64  `json:"durationMs"`
	Status     string `json:"status"`
	TokensUsed int    `json:"tokensUsed,omitempty"`
	Error      string `json:"error,omitempty"`
}

// DebugGitContext represents git-related context for debugging
type DebugGitContext struct {
	Branch       string   `json:"branch"`
	CommitSHA    string   `json:"commitSha"`
	FilesChanged []string `json:"filesChanged,omitempty"`
}

// DebugResponse represents the AI debugger's response
type DebugResponse struct {
	Answer         string        `json:"answer"`
	RootCauses     []RootCause   `json:"rootCauses,omitempty"`
	SuggestedFixes []SuggestedFix `json:"suggestedFixes,omitempty"`
	Confidence     float64       `json:"confidence"`
	SourceTraceIDs []uuid.UUID   `json:"sourceTraceIds,omitempty"`
	GeneratedAt    time.Time     `json:"generatedAt"`
}

// RootCause represents an identified root cause of an issue
type RootCause struct {
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Evidence    string  `json:"evidence"`
	Category    string  `json:"category"`
}

// SuggestedFix represents a suggested fix for an identified issue
type SuggestedFix struct {
	Description string `json:"description"`
	CodeSnippet string `json:"codeSnippet,omitempty"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
}
