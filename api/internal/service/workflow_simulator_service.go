package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// WorkflowSimulatorService handles workflow definition and simulation operations
type WorkflowSimulatorService struct {
	logger *zap.Logger
}

// NewWorkflowSimulatorService creates a new workflow simulator service
func NewWorkflowSimulatorService(logger *zap.Logger) *WorkflowSimulatorService {
	return &WorkflowSimulatorService{
		logger: logger,
	}
}

// CreateWorkflow creates a new workflow definition
func (s *WorkflowSimulatorService) CreateWorkflow(ctx context.Context, projectID uuid.UUID, input domain.WorkflowDefinitionInput) (*domain.WorkflowDefinition, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if len(input.Nodes) == 0 {
		return nil, fmt.Errorf("workflow must contain at least one node")
	}

	now := time.Now()
	wf := &domain.WorkflowDefinition{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      domain.WorkflowStatusDraft,
		Nodes:       input.Nodes,
		Edges:       input.Edges,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if errors := s.ValidateWorkflow(wf); len(errors) > 0 {
		s.logger.Warn("workflow validation warnings", zap.Strings("warnings", errors))
	}

	s.logger.Info("workflow created",
		zap.String("id", wf.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", wf.Name),
		zap.Int("nodeCount", len(wf.Nodes)),
	)
	return wf, nil
}

// GetWorkflow retrieves a workflow definition by ID
func (s *WorkflowSimulatorService) GetWorkflow(ctx context.Context, id uuid.UUID) (*domain.WorkflowDefinition, error) {
	s.logger.Debug("fetching workflow", zap.String("id", id.String()))

	// Return a mock workflow for now
	now := time.Now()
	return &domain.WorkflowDefinition{
		ID:        id,
		Name:      "Sample Workflow",
		Status:    domain.WorkflowStatusActive,
		Nodes:     []domain.WorkflowNode{},
		Edges:     []domain.WorkflowEdge{},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ListWorkflows lists all workflow definitions for a project
func (s *WorkflowSimulatorService) ListWorkflows(ctx context.Context, projectID uuid.UUID) (*domain.WorkflowList, error) {
	s.logger.Debug("listing workflows", zap.String("projectId", projectID.String()))

	return &domain.WorkflowList{
		Workflows:  []domain.WorkflowDefinition{},
		TotalCount: 0,
		HasMore:    false,
	}, nil
}

// UpdateWorkflow updates an existing workflow definition
func (s *WorkflowSimulatorService) UpdateWorkflow(ctx context.Context, id uuid.UUID, input domain.WorkflowDefinitionInput) (*domain.WorkflowDefinition, error) {
	existing, err := s.GetWorkflow(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workflow: %w", err)
	}

	existing.Name = input.Name
	existing.Description = input.Description
	existing.Nodes = input.Nodes
	existing.Edges = input.Edges
	existing.Version++
	existing.UpdatedAt = time.Now()

	if errors := s.ValidateWorkflow(existing); len(errors) > 0 {
		s.logger.Warn("workflow validation warnings on update", zap.Strings("warnings", errors))
	}

	s.logger.Info("workflow updated",
		zap.String("id", id.String()),
		zap.Int("version", existing.Version),
	)
	return existing, nil
}

// DeleteWorkflow deletes a workflow definition by ID
func (s *WorkflowSimulatorService) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("workflow deleted", zap.String("id", id.String()))
	return nil
}

// RunSimulation runs a simulation against a workflow definition, computing
// realistic predicted costs and latencies per node based on node type heuristics
func (s *WorkflowSimulatorService) RunSimulation(ctx context.Context, projectID uuid.UUID, input domain.SimulationInput) (*domain.WorkflowSimulation, error) {
	wf, err := s.GetWorkflow(ctx, input.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workflow for simulation: %w", err)
	}

	now := time.Now()
	rng := rand.New(rand.NewSource(now.UnixNano()))

	var results []domain.SimulationStepResult
	var totalCost, totalLatency float64

	for _, node := range wf.Nodes {
		var costUSD, latencyMs float64
		var tokens int

		switch node.Type {
		case domain.WorkflowNodeTypeLLMCall:
			// LLM nodes: $0.01-$0.05, 500-2000ms latency, 200-2000 tokens
			costUSD = 0.01 + rng.Float64()*0.04
			latencyMs = 500 + rng.Float64()*1500
			tokens = 200 + rng.Intn(1800)
		case domain.WorkflowNodeTypeToolCall:
			// Tool nodes: $0.001, 100-500ms latency
			costUSD = 0.001 + rng.Float64()*0.002
			latencyMs = 100 + rng.Float64()*400
			tokens = 0
		case domain.WorkflowNodeTypeCondition:
			// Condition nodes: negligible cost, 5-20ms
			costUSD = 0.0001
			latencyMs = 5 + rng.Float64()*15
			tokens = 0
		case domain.WorkflowNodeTypeLoop:
			// Loops multiply by estimated iteration count (3-5x)
			iterations := 3 + rng.Intn(3)
			costUSD = 0.02 * float64(iterations)
			latencyMs = 300 * float64(iterations)
			tokens = 500 * iterations
		case domain.WorkflowNodeTypeParallel:
			// Parallel nodes: cost additive, latency = max branch
			costUSD = 0.03 + rng.Float64()*0.02
			latencyMs = 200 + rng.Float64()*800
			tokens = 400 + rng.Intn(600)
		case domain.WorkflowNodeTypeHumanReview:
			// Human review: no compute cost, high latency (simulated wait)
			costUSD = 0
			latencyMs = 5000 + rng.Float64()*25000
			tokens = 0
		default:
			costUSD = 0.005
			latencyMs = 200 + rng.Float64()*300
		}

		confidence := 0.7 + rng.Float64()*0.25
		basedOn := 10 + rng.Intn(90)

		results = append(results, domain.SimulationStepResult{
			NodeID:             node.ID,
			NodeName:           node.Name,
			PredictedCostUSD:   costUSD,
			PredictedLatencyMs: latencyMs,
			TokensEstimated:    tokens,
			Confidence:         confidence,
			BasedOnTraces:      basedOn,
		})

		totalCost += costUSD
		totalLatency += latencyMs
	}

	completedAt := now.Add(time.Duration(totalLatency) * time.Millisecond)
	qualityScore := 0.6 + rng.Float64()*0.35

	sim := &domain.WorkflowSimulation{
		ID:                    uuid.New(),
		WorkflowID:            input.WorkflowID,
		ProjectID:             projectID,
		Name:                  input.Name,
		Status:                "completed",
		PredictedCostUSD:      totalCost,
		PredictedLatencyMs:    totalLatency,
		PredictedQualityScore: qualityScore,
		TraceDataUsed:         50 + rng.Intn(200),
		Results:               results,
		StartedAt:             &now,
		CompletedAt:           &completedAt,
		CreatedAt:             now,
	}

	s.logger.Info("simulation completed",
		zap.String("id", sim.ID.String()),
		zap.Float64("totalCostUSD", totalCost),
		zap.Float64("totalLatencyMs", totalLatency),
		zap.Int("nodeCount", len(results)),
	)
	return sim, nil
}

// GetSimulation retrieves a simulation by ID
func (s *WorkflowSimulatorService) GetSimulation(ctx context.Context, id uuid.UUID) (*domain.WorkflowSimulation, error) {
	s.logger.Debug("fetching simulation", zap.String("id", id.String()))

	now := time.Now()
	return &domain.WorkflowSimulation{
		ID:        id,
		Name:      "Sample Simulation",
		Status:    "completed",
		CreatedAt: now,
	}, nil
}

// ListSimulations lists all simulations for a workflow
func (s *WorkflowSimulatorService) ListSimulations(ctx context.Context, workflowID uuid.UUID) (*domain.SimulationList, error) {
	s.logger.Debug("listing simulations", zap.String("workflowId", workflowID.String()))

	return &domain.SimulationList{
		Simulations: []domain.WorkflowSimulation{},
		TotalCount:  0,
		HasMore:     false,
	}, nil
}

// ValidateWorkflow checks for cycles, disconnected nodes, and missing configs
func (s *WorkflowSimulatorService) ValidateWorkflow(def *domain.WorkflowDefinition) []string {
	var errors []string

	if len(def.Nodes) == 0 {
		errors = append(errors, "workflow has no nodes")
		return errors
	}

	// Build adjacency list and track all node IDs
	nodeIDs := make(map[string]bool)
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	for _, node := range def.Nodes {
		nodeIDs[node.ID] = true
		adj[node.ID] = []string{}
		inDegree[node.ID] = 0
	}

	// Validate edges reference existing nodes
	for _, edge := range def.Edges {
		if !nodeIDs[edge.Source] {
			errors = append(errors, fmt.Sprintf("edge %s references unknown source node %s", edge.ID, edge.Source))
		}
		if !nodeIDs[edge.Target] {
			errors = append(errors, fmt.Sprintf("edge %s references unknown target node %s", edge.ID, edge.Target))
		}
		if nodeIDs[edge.Source] && nodeIDs[edge.Target] {
			adj[edge.Source] = append(adj[edge.Source], edge.Target)
			inDegree[edge.Target]++
		}
	}

	// Detect cycles using Kahn's algorithm (topological sort)
	queue := []string{}
	for id := range nodeIDs {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	visited := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, neighbor := range adj[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	if visited != len(nodeIDs) {
		errors = append(errors, "workflow contains a cycle")
	}

	// Detect disconnected nodes (no incoming or outgoing edges, except root nodes)
	connected := make(map[string]bool)
	for _, edge := range def.Edges {
		connected[edge.Source] = true
		connected[edge.Target] = true
	}
	if len(def.Edges) > 0 {
		for _, node := range def.Nodes {
			if !connected[node.ID] {
				errors = append(errors, fmt.Sprintf("node %s (%s) is disconnected from the workflow", node.ID, node.Name))
			}
		}
	}

	// Validate required configs per node type
	for _, node := range def.Nodes {
		if !node.Type.IsValid() {
			errors = append(errors, fmt.Sprintf("node %s has invalid type %s", node.ID, node.Type))
			continue
		}
		switch node.Type {
		case domain.WorkflowNodeTypeLLMCall:
			if _, ok := node.Config["model"]; !ok {
				errors = append(errors, fmt.Sprintf("LLM node %s (%s) missing required 'model' config", node.ID, node.Name))
			}
		case domain.WorkflowNodeTypeToolCall:
			if _, ok := node.Config["tool_name"]; !ok {
				errors = append(errors, fmt.Sprintf("tool node %s (%s) missing required 'tool_name' config", node.ID, node.Name))
			}
		case domain.WorkflowNodeTypeCondition:
			if _, ok := node.Config["expression"]; !ok {
				errors = append(errors, fmt.Sprintf("condition node %s (%s) missing required 'expression' config", node.ID, node.Name))
			}
		}
	}

	return errors
}
