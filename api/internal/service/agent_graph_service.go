package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AgentGraphService builds graph visualizations from multi-agent traces
type AgentGraphService struct {
	logger       *zap.Logger
	queryService *QueryService
}

// NewAgentGraphService creates a new agent graph service
func NewAgentGraphService(
	logger *zap.Logger,
	queryService *QueryService,
) *AgentGraphService {
	return &AgentGraphService{
		logger:       logger,
		queryService: queryService,
	}
}

// BuildGraph constructs an agent graph from a trace.
// It identifies agents from GENERATION observations, builds edges from
// parent-child relationships, and calculates per-agent cost/token stats.
func (s *AgentGraphService) BuildGraph(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.AgentGraph, error) {
	observations, err := s.queryService.GetObservationsByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observations: %w", err)
	}

	traceUUID, err := uuid.Parse(traceID)
	if err != nil {
		return nil, fmt.Errorf("invalid trace ID: %w", err)
	}

	graph := &domain.AgentGraph{
		TraceID:   traceUUID,
		ProjectID: projectID,
		Agents:    make([]domain.AgentNode, 0),
		Edges:     make([]domain.AgentEdge, 0),
	}

	// Index observations by ID for parent lookups
	obsMap := make(map[string]*domain.Observation, len(observations))
	for i := range observations {
		obsMap[observations[i].ID] = &observations[i]
	}

	// Build nodes from generation observations
	nodeMap := make(map[string]*domain.AgentNode)
	for _, obs := range observations {
		if obs.Type != domain.ObservationTypeGeneration {
			continue
		}

		tokensUsed := int(obs.UsageDetails.InputTokens + obs.UsageDetails.OutputTokens)
		cost := obs.CostDetails.TotalCost

		node := domain.AgentNode{
			ID:            obs.ID,
			Name:          obs.Name,
			Type:          s.inferNodeType(obs),
			ObservationID: obs.ID,
			Model:         obs.Model,
			TokensUsed:    tokensUsed,
			Cost:          cost,
			DurationMs:    int64(obs.DurationMs),
			Status:        s.nodeStatus(obs),
		}

		graph.Agents = append(graph.Agents, node)
		nodeMap[obs.ID] = &node
		graph.TotalCost += cost
		graph.TotalDurationMs += int64(obs.DurationMs)
	}

	// Build edges from parent-child relationships
	edgeTracker := make(map[string]*domain.AgentEdge)
	for _, obs := range observations {
		if obs.ParentObservationID == nil {
			continue
		}

		parentID := *obs.ParentObservationID
		childID := obs.ID

		// Only create edges between generation nodes
		if _, parentIsNode := nodeMap[parentID]; !parentIsNode {
			continue
		}
		if _, childIsNode := nodeMap[childID]; !childIsNode {
			continue
		}

		edgeKey := parentID + "->" + childID
		edge, exists := edgeTracker[edgeKey]
		if !exists {
			edge = &domain.AgentEdge{
				SourceID: parentID,
				TargetID: childID,
			}
			edgeTracker[edgeKey] = edge
		}
		edge.MessageCount++
		edge.TokenCount += int(obs.UsageDetails.InputTokens + obs.UsageDetails.OutputTokens)
	}

	for _, edge := range edgeTracker {
		graph.Edges = append(graph.Edges, *edge)
	}

	s.logger.Info("built agent graph",
		zap.String("traceId", traceID),
		zap.Int("nodes", len(graph.Agents)),
		zap.Int("edges", len(graph.Edges)),
	)

	return graph, nil
}

// CompareGraphs compares the agent graphs of two traces
func (s *AgentGraphService) CompareGraphs(ctx context.Context, projectID uuid.UUID, traceIDA string, traceIDB string) (*domain.AgentGraphComparison, error) {
	graphA, err := s.BuildGraph(ctx, projectID, traceIDA)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph for trace A: %w", err)
	}

	graphB, err := s.BuildGraph(ctx, projectID, traceIDB)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph for trace B: %w", err)
	}

	// Find added and removed nodes
	nodesA := make(map[string]bool)
	for _, node := range graphA.Agents {
		nodesA[node.Name] = true
	}

	nodesB := make(map[string]bool)
	for _, node := range graphB.Agents {
		nodesB[node.Name] = true
	}

	var addedNodes, removedNodes []string
	for name := range nodesB {
		if !nodesA[name] {
			addedNodes = append(addedNodes, name)
		}
	}
	for name := range nodesA {
		if !nodesB[name] {
			removedNodes = append(removedNodes, name)
		}
	}

	comparison := &domain.AgentGraphComparison{
		GraphA:       *graphA,
		GraphB:       *graphB,
		AddedNodes:   addedNodes,
		RemovedNodes: removedNodes,
		CostDelta:    graphB.TotalCost - graphA.TotalCost,
		LatencyDelta: graphB.TotalDurationMs - graphA.TotalDurationMs,
	}

	return comparison, nil
}

// inferNodeType determines the agent node type from an observation
func (s *AgentGraphService) inferNodeType(obs domain.Observation) domain.AgentNodeType {
	if obs.ParentObservationID == nil {
		return domain.AgentNodeOrchestrator
	}
	if obs.Type == domain.ObservationTypeSpan {
		return domain.AgentNodeTool
	}
	return domain.AgentNodeWorker
}

// nodeStatus determines the status string for an agent node
func (s *AgentGraphService) nodeStatus(obs domain.Observation) string {
	if obs.Level == "error" || obs.Level == "ERROR" {
		return "error"
	}
	if obs.EndTime == nil {
		return "running"
	}
	return "success"
}
