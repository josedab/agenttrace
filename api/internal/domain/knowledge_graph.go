package domain

import (
	"time"

	"github.com/google/uuid"
)

type KnowledgeGraph struct {
	ProjectID uuid.UUID `json:"projectId"`
	Nodes     []KGNode  `json:"nodes"`
	Edges     []KGEdge  `json:"edges"`
	Stats     KGStats   `json:"stats"`
}

type KGNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"` // file, function, module, tool, agent, dependency
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Weight   int            `json:"weight"`
}

type KGEdge struct {
	Source       string    `json:"source"`
	Target       string    `json:"target"`
	Relationship string   `json:"relationship"` // modifies, calls, imports, depends_on, produces
	Weight       int       `json:"weight"`
	LastSeen     time.Time `json:"lastSeen"`
}

type KGStats struct {
	TotalNodes    int    `json:"totalNodes"`
	TotalEdges    int    `json:"totalEdges"`
	MostConnected string `json:"mostConnected"`
	Clusters      int    `json:"clusters"`
}

type KGQuery struct {
	NodeType  string `json:"nodeType,omitempty"`
	Depth     int    `json:"depth,omitempty"`
	FocusNode string `json:"focusNode,omitempty"`
}
