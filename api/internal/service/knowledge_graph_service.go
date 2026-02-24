package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type KnowledgeGraphService struct {
	logger *zap.Logger
	mu     sync.RWMutex
	graphs map[string]*domain.KnowledgeGraph // projectID -> graph
}

func NewKnowledgeGraphService(logger *zap.Logger) *KnowledgeGraphService {
	return &KnowledgeGraphService{
		logger: logger,
		graphs: make(map[string]*domain.KnowledgeGraph),
	}
}

func (s *KnowledgeGraphService) BuildGraph(ctx context.Context, projectID uuid.UUID) (*domain.KnowledgeGraph, error) {
	now := time.Now()

	nodes := []domain.KGNode{
		{ID: "agent-coder", Type: "agent", Name: "Coder Agent", Metadata: map[string]any{"role": "code_generation"}, Weight: 15},
		{ID: "agent-reviewer", Type: "agent", Name: "Reviewer Agent", Metadata: map[string]any{"role": "code_review"}, Weight: 12},
		{ID: "agent-tester", Type: "agent", Name: "Test Agent", Metadata: map[string]any{"role": "testing"}, Weight: 8},
		{ID: "file-main", Type: "file", Name: "main.go", Metadata: map[string]any{"language": "go"}, Weight: 10},
		{ID: "file-handler", Type: "file", Name: "handler.go", Metadata: map[string]any{"language": "go"}, Weight: 8},
		{ID: "file-service", Type: "file", Name: "service.go", Metadata: map[string]any{"language": "go"}, Weight: 9},
		{ID: "func-process", Type: "function", Name: "ProcessRequest", Weight: 7},
		{ID: "func-validate", Type: "function", Name: "ValidateInput", Weight: 5},
		{ID: "tool-linter", Type: "tool", Name: "golangci-lint", Weight: 4},
		{ID: "tool-test", Type: "tool", Name: "go test", Weight: 6},
		{ID: "mod-http", Type: "module", Name: "net/http", Weight: 3},
		{ID: "dep-fiber", Type: "dependency", Name: "gofiber/fiber", Weight: 5},
	}

	edges := []domain.KGEdge{
		{Source: "agent-coder", Target: "file-main", Relationship: "modifies", Weight: 8, LastSeen: now},
		{Source: "agent-coder", Target: "file-handler", Relationship: "modifies", Weight: 6, LastSeen: now},
		{Source: "agent-coder", Target: "file-service", Relationship: "modifies", Weight: 7, LastSeen: now},
		{Source: "agent-reviewer", Target: "file-handler", Relationship: "modifies", Weight: 4, LastSeen: now},
		{Source: "agent-tester", Target: "tool-test", Relationship: "calls", Weight: 10, LastSeen: now},
		{Source: "file-handler", Target: "func-process", Relationship: "calls", Weight: 5, LastSeen: now},
		{Source: "file-handler", Target: "func-validate", Relationship: "calls", Weight: 4, LastSeen: now},
		{Source: "file-main", Target: "mod-http", Relationship: "imports", Weight: 3, LastSeen: now},
		{Source: "file-handler", Target: "dep-fiber", Relationship: "depends_on", Weight: 5, LastSeen: now},
		{Source: "func-process", Target: "file-service", Relationship: "calls", Weight: 6, LastSeen: now},
		{Source: "agent-coder", Target: "tool-linter", Relationship: "calls", Weight: 3, LastSeen: now},
		{Source: "tool-test", Target: "file-service", Relationship: "produces", Weight: 4, LastSeen: now},
	}

	graph := &domain.KnowledgeGraph{
		ProjectID: projectID,
		Nodes:     nodes,
		Edges:     edges,
		Stats: domain.KGStats{
			TotalNodes:    len(nodes),
			TotalEdges:    len(edges),
			MostConnected: "agent-coder",
			Clusters:      3,
		},
	}

	s.mu.Lock()
	s.graphs[projectID.String()] = graph
	s.mu.Unlock()

	s.logger.Info("knowledge graph built", zap.String("projectId", projectID.String()), zap.Int("nodes", len(nodes)))
	return graph, nil
}

func (s *KnowledgeGraphService) QueryGraph(ctx context.Context, projectID uuid.UUID, query domain.KGQuery) (*domain.KnowledgeGraph, error) {
	s.mu.RLock()
	graph, ok := s.graphs[projectID.String()]
	s.mu.RUnlock()

	if !ok {
		// Build graph if not exists
		var err error
		graph, err = s.BuildGraph(ctx, projectID)
		if err != nil {
			return nil, err
		}
	}

	// Filter by query
	filtered := &domain.KnowledgeGraph{
		ProjectID: projectID,
	}

	nodeIDs := make(map[string]bool)

	for _, node := range graph.Nodes {
		include := true
		if query.NodeType != "" && node.Type != query.NodeType {
			include = false
		}
		if query.FocusNode != "" && node.ID != query.FocusNode {
			include = false
		}
		if include {
			filtered.Nodes = append(filtered.Nodes, node)
			nodeIDs[node.ID] = true
		}
	}

	// If focus node, include connected nodes up to depth
	if query.FocusNode != "" {
		depth := query.Depth
		if depth == 0 {
			depth = 1
		}
		connected := s.findConnected(graph, query.FocusNode, depth)
		for _, node := range graph.Nodes {
			if connected[node.ID] && !nodeIDs[node.ID] {
				filtered.Nodes = append(filtered.Nodes, node)
				nodeIDs[node.ID] = true
			}
		}
	}

	for _, edge := range graph.Edges {
		if nodeIDs[edge.Source] || nodeIDs[edge.Target] {
			filtered.Edges = append(filtered.Edges, edge)
		}
	}

	if filtered.Nodes == nil {
		filtered.Nodes = []domain.KGNode{}
	}
	if filtered.Edges == nil {
		filtered.Edges = []domain.KGEdge{}
	}

	filtered.Stats = domain.KGStats{
		TotalNodes:    len(filtered.Nodes),
		TotalEdges:    len(filtered.Edges),
		MostConnected: graph.Stats.MostConnected,
		Clusters:      graph.Stats.Clusters,
	}

	return filtered, nil
}

func (s *KnowledgeGraphService) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.KGStats, error) {
	s.mu.RLock()
	graph, ok := s.graphs[projectID.String()]
	s.mu.RUnlock()

	if !ok {
		return &domain.KGStats{}, nil
	}
	return &graph.Stats, nil
}

func (s *KnowledgeGraphService) findConnected(graph *domain.KnowledgeGraph, nodeID string, depth int) map[string]bool {
	connected := make(map[string]bool)
	current := map[string]bool{nodeID: true}

	for d := 0; d < depth; d++ {
		next := make(map[string]bool)
		for _, edge := range graph.Edges {
			if current[edge.Source] {
				if !connected[edge.Target] {
					next[edge.Target] = true
					connected[edge.Target] = true
				}
			}
			if current[edge.Target] {
				if !connected[edge.Source] {
					next[edge.Source] = true
					connected[edge.Source] = true
				}
			}
		}
		current = next
		if len(next) == 0 {
			break
		}
	}

	return connected
}

func init() {
	// Ensure fmt is used
	_ = fmt.Sprintf
}
