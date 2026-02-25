package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MultiAgentGraphService handles multi-agent collaboration graph logic
type MultiAgentGraphService struct {
	logger *zap.Logger
}

// NewMultiAgentGraphService creates a new multi-agent graph service
func NewMultiAgentGraphService(logger *zap.Logger) *MultiAgentGraphService {
	return &MultiAgentGraphService{
		logger: logger,
	}
}

// AnalyzeSession analyzes a trace to build a multi-agent collaboration session
func (s *MultiAgentGraphService) AnalyzeSession(ctx context.Context, projectID uuid.UUID, traceID uuid.UUID) (*domain.MultiAgentSession, error) {
	s.logger.Info("analyzing multi-agent session",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID.String()),
	)

	now := time.Now()
	agents := []domain.CollabAgentNode{}
	messages := []domain.CollabAgentMessage{}

	topology := s.DetectTopology(agents, messages)

	session := &domain.MultiAgentSession{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     traceID,
		Name:        "Session for trace " + traceID.String(),
		Topology:    topology,
		Agents:      agents,
		Messages:    messages,
		Bottlenecks: s.IdentifyBottlenecks(&domain.MultiAgentSession{Agents: agents, Messages: messages}),
		Status:      "completed",
		StartTime:   now,
		CreatedAt:   now,
	}

	return session, nil
}

// DetectTopology determines the collaboration topology from agents and messages.
// If one coordinator exists → hub_spoke, if all agents communicate with all → mesh,
// if messages flow sequentially → pipeline.
func (s *MultiAgentGraphService) DetectTopology(agents []domain.CollabAgentNode, messages []domain.CollabAgentMessage) domain.CollabTopology {
	if len(agents) == 0 {
		return domain.CollabTopologyPipeline
	}

	// Count coordinators
	coordinatorCount := 0
	for _, a := range agents {
		if a.Role == domain.AgentRoleCoordinator {
			coordinatorCount++
		}
	}

	// If exactly one coordinator, it's hub-spoke
	if coordinatorCount == 1 {
		return domain.CollabTopologyHubSpoke
	}

	// Build adjacency: track unique sender→receiver pairs
	type pair struct{ from, to uuid.UUID }
	seen := make(map[pair]bool)
	for _, m := range messages {
		seen[pair{m.FromAgentID, m.ToAgentID}] = true
	}

	n := len(agents)
	maxPairs := n * (n - 1)

	// If every agent talks to every other agent, it's mesh
	if maxPairs > 0 && len(seen) >= maxPairs {
		return domain.CollabTopologyMesh
	}

	// Check for sequential/pipeline pattern: each agent sends to at most one other
	outDegree := make(map[uuid.UUID]int)
	for p := range seen {
		outDegree[p.from]++
	}

	isPipeline := true
	for _, deg := range outDegree {
		if deg > 1 {
			isPipeline = false
			break
		}
	}
	if isPipeline && len(messages) > 0 {
		return domain.CollabTopologyPipeline
	}

	// Multiple coordinators → hierarchical
	if coordinatorCount > 1 {
		return domain.CollabTopologyHierarchical
	}

	return domain.CollabTopologyMesh
}

// IdentifyBottlenecks finds bottlenecks in a multi-agent session
func (s *MultiAgentGraphService) IdentifyBottlenecks(session *domain.MultiAgentSession) []domain.CollabBottleneck {
	bottlenecks := []domain.CollabBottleneck{}

	for _, agent := range session.Agents {
		// Flag agents with high average latency as bottlenecks
		if agent.AvgLatencyMs > 5000 {
			bottlenecks = append(bottlenecks, domain.CollabBottleneck{
				AgentID:      agent.ID,
				Type:         "high_latency",
				Description:  agent.Name + " has high average latency",
				Impact:       agent.AvgLatencyMs / 1000.0,
				SuggestedFix: "Consider optimizing prompts or using a faster model for " + agent.Name,
			})
		}

		// Flag agents handling too many tasks
		if agent.TaskCount > 20 {
			bottlenecks = append(bottlenecks, domain.CollabBottleneck{
				AgentID:      agent.ID,
				Type:         "overloaded",
				Description:  agent.Name + " is handling too many tasks",
				Impact:       float64(agent.TaskCount) / 10.0,
				SuggestedFix: "Distribute tasks across additional worker agents",
			})
		}
	}

	return bottlenecks
}

// ListSessions returns a paginated list of multi-agent sessions for a project
func (s *MultiAgentGraphService) ListSessions(ctx context.Context, projectID uuid.UUID) (*domain.MultiAgentSessionList, error) {
	s.logger.Info("listing multi-agent sessions", zap.String("projectId", projectID.String()))

	return &domain.MultiAgentSessionList{
		Sessions:   []domain.MultiAgentSession{},
		TotalCount: 0,
		HasMore:    false,
	}, nil
}

// GetSession returns a specific multi-agent session by ID
func (s *MultiAgentGraphService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.MultiAgentSession, error) {
	s.logger.Info("fetching multi-agent session", zap.String("sessionId", sessionID.String()))

	now := time.Now()
	session := &domain.MultiAgentSession{
		ID:          sessionID,
		Topology:    domain.CollabTopologyPipeline,
		Agents:      []domain.CollabAgentNode{},
		Messages:    []domain.CollabAgentMessage{},
		Bottlenecks: []domain.CollabBottleneck{},
		Status:      "completed",
		StartTime:   now,
		CreatedAt:   now,
	}

	return session, nil
}

// GetTopologyGraph builds a ReactFlow-compatible graph from a multi-agent session
func (s *MultiAgentGraphService) GetTopologyGraph(ctx context.Context, sessionID uuid.UUID) (*domain.TopologyGraph, error) {
session, err := s.GetSession(ctx, sessionID)
if err != nil {
return nil, err
}

graph := &domain.TopologyGraph{
SessionID: sessionID,
Layout:    "dagre",
}

for i, agent := range session.Agents {
graph.Nodes = append(graph.Nodes, domain.TopologyNode{
ID:    agent.ID.String(),
Type:  "agent",
Label: agent.Name,
Position: domain.TopologyPosition{X: float64(i%3) * 250, Y: float64(i/3) * 200},
Data: domain.TopologyNodeData{
Role: agent.Role, Framework: agent.Framework, Status: agent.Status,
TaskCount: agent.TaskCount, AvgLatencyMs: agent.AvgLatencyMs, TotalCostUsd: agent.TotalCostUsd,
},
})
}

edgeMap := make(map[string]*domain.TopologyEdge)
for _, msg := range session.Messages {
key := msg.FromAgentID.String() + "->" + msg.ToAgentID.String()
if edge, exists := edgeMap[key]; exists {
edge.Data.MessageCount++
} else {
edgeMap[key] = &domain.TopologyEdge{
ID: key, Source: msg.FromAgentID.String(), Target: msg.ToAgentID.String(),
Label: string(msg.Type), Type: string(msg.Type),
Animated: msg.Type == domain.MessageTypeDelegation,
Data:     domain.TopologyEdgeData{MessageCount: 1, AvgLatencyMs: float64(msg.DurationMs)},
}
}
}
for _, edge := range edgeMap {
graph.Edges = append(graph.Edges, *edge)
}

totalCost := 0.0
for _, a := range session.Agents {
totalCost += a.TotalCostUsd
}
graph.Stats = domain.TopologyStats{
TotalAgents: len(session.Agents), TotalMessages: len(session.Messages), TotalCost: totalCost,
}

return graph, nil
}
