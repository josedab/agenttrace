package domain

import (
	"time"

	"github.com/google/uuid"
)

// AgentRole represents the role of an agent in a multi-agent collaboration
type AgentRole string

const (
	AgentRoleCoordinator AgentRole = "coordinator"
	AgentRoleWorker      AgentRole = "worker"
	AgentRoleReviewer    AgentRole = "reviewer"
	AgentRolePlanner     AgentRole = "planner"
	AgentRoleExecutor    AgentRole = "executor"
)

// MessageType represents the type of message exchanged between agents
type MessageType string

const (
	MessageTypeDelegation MessageType = "delegation"
	MessageTypeResponse   MessageType = "response"
	MessageTypeConsensus  MessageType = "consensus"
	MessageTypeDebate     MessageType = "debate"
	MessageTypeFeedback   MessageType = "feedback"
)

// CollabTopology represents the collaboration topology of a multi-agent session
type CollabTopology string

const (
	CollabTopologyPipeline     CollabTopology = "pipeline"
	CollabTopologyHubSpoke     CollabTopology = "hub_spoke"
	CollabTopologyMesh         CollabTopology = "mesh"
	CollabTopologyDebate       CollabTopology = "debate"
	CollabTopologyHierarchical CollabTopology = "hierarchical"
)

// MultiAgentSession represents a multi-agent collaboration session
type MultiAgentSession struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	TraceID     uuid.UUID      `json:"traceId"`
	Name        string         `json:"name"`
	Topology    CollabTopology `json:"topology"`
	Agents      []CollabAgentNode    `json:"agents"`
	Messages    []CollabAgentMessage `json:"messages"`
	Bottlenecks []CollabBottleneck `json:"bottlenecks,omitempty"`
	Status      string         `json:"status"`
	StartTime   time.Time      `json:"startTime"`
	EndTime     *time.Time     `json:"endTime,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
}

// CollabAgentNode represents an agent participating in a multi-agent session
type CollabAgentNode struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Role         AgentRole `json:"role"`
	Framework    string    `json:"framework"`
	TaskCount    int       `json:"taskCount"`
	AvgLatencyMs float64  `json:"avgLatencyMs"`
	TotalCostUsd float64  `json:"totalCostUsd"`
	Status       string   `json:"status"`
}

// CollabAgentMessage represents a message exchanged between agents in collaboration
type CollabAgentMessage struct {
	ID          uuid.UUID   `json:"id"`
	FromAgentID uuid.UUID   `json:"fromAgentId"`
	ToAgentID   uuid.UUID   `json:"toAgentId"`
	Type        MessageType `json:"type"`
	Content     string      `json:"content"`
	Timestamp   time.Time   `json:"timestamp"`
	DurationMs  int64       `json:"durationMs"`
}

// CollabBottleneck represents a detected bottleneck in a multi-agent session
type CollabBottleneck struct {
	AgentID      uuid.UUID `json:"agentId"`
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	Impact       float64   `json:"impact"`
	SuggestedFix string    `json:"suggestedFix"`
}

// MultiAgentSessionInput represents input for creating/updating a multi-agent session
type MultiAgentSessionInput struct {
	Name     string         `json:"name" validate:"required,min=1,max=200"`
	TraceID  uuid.UUID      `json:"traceId" validate:"required"`
	Topology CollabTopology `json:"topology" validate:"required"`
	Agents   []CollabAgentNode `json:"agents,omitempty"`
}

// MultiAgentSessionFilter represents filter options for querying multi-agent sessions
type MultiAgentSessionFilter struct {
	ProjectID uuid.UUID
	Topology  *CollabTopology
	Status    *string
	StartTime *time.Time
	EndTime   *time.Time
	TraceID   *uuid.UUID
}

// MultiAgentSessionList represents a paginated list of multi-agent sessions
type MultiAgentSessionList struct {
	Sessions   []MultiAgentSession `json:"sessions"`
	TotalCount int64               `json:"totalCount"`
	HasMore    bool                `json:"hasMore"`
}

// TopologyGraph represents the graph data for topology visualization (ReactFlow compatible)
type TopologyGraph struct {
	SessionID uuid.UUID          `json:"sessionId"`
	Nodes     []TopologyNode     `json:"nodes"`
	Edges     []TopologyEdge     `json:"edges"`
	Stats     TopologyStats      `json:"stats"`
	Layout    string             `json:"layout"` // dagre, force, circular
}

// TopologyNode represents a node in the topology graph
type TopologyNode struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"` // agent, tool, external
	Label    string            `json:"label"`
	Position TopologyPosition  `json:"position"`
	Data     TopologyNodeData  `json:"data"`
}

// TopologyPosition represents x,y coordinates for graph layout
type TopologyPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// TopologyNodeData contains data for a topology node
type TopologyNodeData struct {
	Role         AgentRole `json:"role"`
	Framework    string    `json:"framework,omitempty"`
	Model        string    `json:"model,omitempty"`
	Status       string    `json:"status"`
	TaskCount    int       `json:"taskCount"`
	AvgLatencyMs float64  `json:"avgLatencyMs"`
	TotalCostUsd float64  `json:"totalCostUsd"`
	ErrorRate    float64   `json:"errorRate"`
}

// TopologyEdge represents a connection between nodes
type TopologyEdge struct {
	ID        string           `json:"id"`
	Source    string           `json:"source"`
	Target    string           `json:"target"`
	Label     string           `json:"label,omitempty"`
	Type      string           `json:"type"` // delegation, response, data_flow
	Animated  bool             `json:"animated"`
	Data      TopologyEdgeData `json:"data"`
}

// TopologyEdgeData contains data for a topology edge
type TopologyEdgeData struct {
	MessageCount int     `json:"messageCount"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	TotalTokens  int64   `json:"totalTokens"`
}

// TopologyStats contains aggregate stats for the topology
type TopologyStats struct {
	TotalAgents    int     `json:"totalAgents"`
	TotalMessages  int     `json:"totalMessages"`
	TotalCost      float64 `json:"totalCost"`
	AvgLatency     float64 `json:"avgLatencyMs"`
	DelegationDepth int    `json:"delegationDepth"`
	ParallelPaths  int     `json:"parallelPaths"`
}
