package domain

import (
	"github.com/google/uuid"
)

// ReasoningNodeType represents the type of a node in a reasoning tree
type ReasoningNodeType string

const (
	ReasoningNodeTypeInput     ReasoningNodeType = "INPUT"
	ReasoningNodeTypeReasoning ReasoningNodeType = "REASONING"
	ReasoningNodeTypeDecision  ReasoningNodeType = "DECISION"
	ReasoningNodeTypeAction    ReasoningNodeType = "ACTION"
	ReasoningNodeTypeToolCall  ReasoningNodeType = "TOOL_CALL"
	ReasoningNodeTypeOutput    ReasoningNodeType = "OUTPUT"
	ReasoningNodeTypeError     ReasoningNodeType = "ERROR"
)

// IsValid checks if the reasoning node type is valid
func (t ReasoningNodeType) IsValid() bool {
	switch t {
	case ReasoningNodeTypeInput, ReasoningNodeTypeReasoning, ReasoningNodeTypeDecision, ReasoningNodeTypeAction, ReasoningNodeTypeToolCall, ReasoningNodeTypeOutput, ReasoningNodeTypeError:
		return true
	}
	return false
}

// ReasoningTree represents the complete reasoning structure for a trace
type ReasoningTree struct {
	TraceID             string         `json:"traceId"`
	ProjectID           uuid.UUID      `json:"projectId"`
	RootNode            *ReasoningNode `json:"rootNode"`
	TotalNodes          int            `json:"totalNodes"`
	TotalDecisionPoints int            `json:"totalDecisionPoints"`
	MaxDepth            int            `json:"maxDepth"`
}

// ReasoningNode represents a single node in an agent's reasoning tree
type ReasoningNode struct {
	ID               string            `json:"id"`
	ParentID         string            `json:"parentId,omitempty"`
	Type             ReasoningNodeType `json:"type"`
	Label            string            `json:"label"`
	Input            any               `json:"input,omitempty"`
	Output           any               `json:"output,omitempty"`
	Reasoning        string            `json:"reasoning,omitempty"`
	Confidence       float64           `json:"confidence"`
	AlternativePaths []AlternativePath `json:"alternativePaths,omitempty"`
	Children         []*ReasoningNode  `json:"children,omitempty"`
	DurationMs       int64             `json:"durationMs"`
	Cost             float64           `json:"cost"`
	TokensUsed       int               `json:"tokensUsed"`
}

// AlternativePath represents an alternative decision path that was considered
type AlternativePath struct {
	Label       string  `json:"label"`
	Probability float64 `json:"probability"`
	WasChosen   bool    `json:"wasChosen"`
	Reason      string  `json:"reason"`
}

// ReasoningQuery represents query parameters for retrieving a reasoning tree
type ReasoningQuery struct {
	TraceID             string `json:"traceId"`
	MaxDepth            *int   `json:"maxDepth,omitempty"`
	IncludeAlternatives bool   `json:"includeAlternatives"`
}
