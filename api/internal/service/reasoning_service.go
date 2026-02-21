package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ReasoningService builds decision tree visualizations from traces by
// mapping observations to reasoning nodes and computing tree metrics.
type ReasoningService struct {
	logger       *zap.Logger
	queryService *QueryService
}

// NewReasoningService creates a new reasoning service
func NewReasoningService(
	logger *zap.Logger,
	queryService *QueryService,
) *ReasoningService {
	return &ReasoningService{
		logger:       logger,
		queryService: queryService,
	}
}

// BuildReasoningTree constructs a reasoning tree from all observations in a
// trace. Observations are mapped to node types as follows:
//   - GENERATION -> REASONING or DECISION (based on output keywords)
//   - SPAN with a tool name -> TOOL_CALL
//   - EVENT -> ACTION
//
// Parent-child relationships are derived from observation parentIDs. The
// method also counts decision points and computes the maximum tree depth.
func (s *ReasoningService) BuildReasoningTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ReasoningTree, error) {
	observations, err := s.queryService.GetObservationsByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observations for reasoning tree: %w", err)
	}

	if len(observations) == 0 {
		return &domain.ReasoningTree{
			TraceID:   traceID,
			ProjectID: projectID,
		}, nil
	}

	// Map observations to reasoning nodes
	nodeMap := make(map[string]*domain.ReasoningNode)
	for i := range observations {
		obs := &observations[i]
		node := s.observationToNode(obs)
		nodeMap[obs.ID] = node
	}

	// Build parent-child tree
	var roots []*domain.ReasoningNode
	for i := range observations {
		obs := &observations[i]
		node := nodeMap[obs.ID]

		if obs.ParentObservationID != nil && *obs.ParentObservationID != "" {
			if parent, ok := nodeMap[*obs.ParentObservationID]; ok {
				node.ParentID = parent.ID
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	// Use a synthetic root if there are multiple top-level nodes
	var rootNode *domain.ReasoningNode
	if len(roots) == 1 {
		rootNode = roots[0]
	} else {
		rootNode = &domain.ReasoningNode{
			ID:       traceID,
			Type:     domain.ReasoningNodeTypeInput,
			Label:    "Root",
			Children: roots,
		}
	}

	// Count decision points (nodes with >1 child or alternative paths)
	totalNodes := len(nodeMap)
	if rootNode.ID == traceID {
		totalNodes++ // synthetic root
	}

	decisionPoints := 0
	for _, node := range nodeMap {
		if len(node.Children) > 1 || len(node.AlternativePaths) > 0 {
			decisionPoints++
		}
	}

	maxDepth := s.computeDepth(rootNode)

	tree := &domain.ReasoningTree{
		TraceID:             traceID,
		ProjectID:           projectID,
		RootNode:            rootNode,
		TotalNodes:          totalNodes,
		TotalDecisionPoints: decisionPoints,
		MaxDepth:            maxDepth,
	}

	s.logger.Debug("built reasoning tree",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID),
		zap.Int("totalNodes", totalNodes),
		zap.Int("decisionPoints", decisionPoints),
	)

	return tree, nil
}

// GetNode retrieves a single reasoning node from a trace's reasoning tree.
func (s *ReasoningService) GetNode(ctx context.Context, projectID uuid.UUID, traceID, nodeID string) (*domain.ReasoningNode, error) {
	tree, err := s.BuildReasoningTree(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build reasoning tree for node lookup: %w", err)
	}

	node := s.findNode(tree.RootNode, nodeID)
	if node == nil {
		return nil, fmt.Errorf("node %s not found in trace %s", nodeID, traceID)
	}

	return node, nil
}

// decisionKeywords are terms that indicate a GENERATION output represents a
// decision rather than pure reasoning.
var decisionKeywords = []string{
	"decided", "decision", "choose", "chose", "selected", "selecting",
	"picked", "determined", "concluded", "resolved",
}

// observationToNode maps a single observation to a ReasoningNode based on
// its type and attributes.
func (s *ReasoningService) observationToNode(obs *domain.Observation) *domain.ReasoningNode {
	node := &domain.ReasoningNode{
		ID:         obs.ID,
		Label:      obs.Name,
		Input:      obs.Input,
		Output:     obs.Output,
		DurationMs: int64(obs.DurationMs),
		Cost:       obs.CostDetails.TotalCost,
		TokensUsed: int(obs.UsageDetails.TotalTokens),
	}

	switch obs.Type {
	case domain.ObservationTypeGeneration:
		node.Type = domain.ReasoningNodeTypeReasoning
		if s.containsDecisionKeywords(obs.Output) {
			node.Type = domain.ReasoningNodeTypeDecision
		}
	case domain.ObservationTypeSpan:
		if obs.Name != "" && strings.Contains(strings.ToLower(obs.Name), "tool") {
			node.Type = domain.ReasoningNodeTypeToolCall
		} else {
			node.Type = domain.ReasoningNodeTypeAction
		}
	case domain.ObservationTypeEvent:
		node.Type = domain.ReasoningNodeTypeAction
	default:
		node.Type = domain.ReasoningNodeTypeReasoning
	}

	return node
}

// containsDecisionKeywords checks whether the output text contains any
// decision-related keywords.
func (s *ReasoningService) containsDecisionKeywords(output string) bool {
	lower := strings.ToLower(output)
	for _, kw := range decisionKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// computeDepth returns the maximum depth of the reasoning tree.
func (s *ReasoningService) computeDepth(node *domain.ReasoningNode) int {
	if node == nil {
		return 0
	}
	maxChild := 0
	for _, child := range node.Children {
		d := s.computeDepth(child)
		if d > maxChild {
			maxChild = d
		}
	}
	return maxChild + 1
}

// findNode recursively searches for a node with the given ID.
func (s *ReasoningService) findNode(node *domain.ReasoningNode, id string) *domain.ReasoningNode {
	if node == nil {
		return nil
	}
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := s.findNode(child, id); found != nil {
			return found
		}
	}
	return nil
}
