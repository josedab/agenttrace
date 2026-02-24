package domain

import (
	"time"

	"github.com/google/uuid"
)

// KGNodeType represents the type of a node in the knowledge graph
type KGNodeType string

const (
	KGNodeTypeFile     KGNodeType = "file"
	KGNodeTypeFunction KGNodeType = "function"
	KGNodeTypeClass    KGNodeType = "class"
	KGNodeTypePackage  KGNodeType = "package"
	KGNodeTypeModule   KGNodeType = "module"
	KGNodeTypeVariable KGNodeType = "variable"
)

// KGEdgeType represents the type of an edge in the knowledge graph
type KGEdgeType string

const (
	KGEdgeTypeImports   KGEdgeType = "imports"
	KGEdgeTypeCalls     KGEdgeType = "calls"
	KGEdgeTypeModifies  KGEdgeType = "modifies"
	KGEdgeTypeReads     KGEdgeType = "reads"
	KGEdgeTypeInherits  KGEdgeType = "inherits"
	KGEdgeTypeDependsOn KGEdgeType = "depends_on"
)

// KnowledgeGraphView represents a view of the agent knowledge graph
type KnowledgeGraphView struct {
	ProjectID   uuid.UUID `json:"projectId"`
	Nodes       []AgentKGNode `json:"nodes"`
	Edges       []AgentKGEdge `json:"edges"`
	Stats       AgentKGStats  `json:"stats"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// AgentKGNode represents a node in the agent knowledge graph
type AgentKGNode struct {
	ID           string                 `json:"id"`
	Type         KGNodeType             `json:"type"`
	Name         string                 `json:"name"`
	FilePath     string                 `json:"filePath"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	TraceCount   int                    `json:"traceCount"`
	LastAccessed *time.Time             `json:"lastAccessed,omitempty"`
}

// AgentKGEdge represents an edge in the agent knowledge graph
type AgentKGEdge struct {
	Source   string      `json:"source"`
	Target   string      `json:"target"`
	Type     KGEdgeType  `json:"type"`
	Weight   float64     `json:"weight"`
	TraceIDs []uuid.UUID `json:"traceIds"`
}

// AgentKGStats provides statistics about the agent knowledge graph
type AgentKGStats struct {
	TotalNodes        int      `json:"totalNodes"`
	TotalEdges        int      `json:"totalEdges"`
	FilesCovered      int      `json:"filesCovered"`
	FunctionsCovered  int      `json:"functionsCovered"`
	AvgDepth          int      `json:"avgDepth"`
	MostConnected     []string `json:"mostConnected"`
}

// KGEvolution represents the evolution of the knowledge graph over time
type KGEvolution struct {
	ProjectID uuid.UUID    `json:"projectId"`
	Snapshots []KGSnapshot `json:"snapshots"`
}

// KGSnapshot represents a snapshot of knowledge graph changes at a point in time
type KGSnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	NodesAdded   int       `json:"nodesAdded"`
	EdgesAdded   int       `json:"edgesAdded"`
	NodesRemoved int       `json:"nodesRemoved"`
	NewPatterns  []string  `json:"newPatterns,omitempty"`
}

// AgentKGQuery represents a query for the agent knowledge graph
type AgentKGQuery struct {
	ProjectID uuid.UUID    `json:"projectId"`
	FocusNode string       `json:"focusNode"`
	Depth     int          `json:"depth"`
	NodeTypes []KGNodeType `json:"nodeTypes,omitempty"`
	EdgeTypes []KGEdgeType `json:"edgeTypes,omitempty"`
}
