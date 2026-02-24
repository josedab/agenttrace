package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestWorkflowCreateWorkflow(t *testing.T) {
	logger := zap.NewNop()
	svc := NewWorkflowSimulatorService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	input := domain.WorkflowDefinitionInput{
		Name: "Test Workflow",
		Nodes: []domain.WorkflowNode{
			{ID: "n1", Type: domain.WorkflowNodeTypeLLMCall, Name: "LLM Step", Config: map[string]interface{}{"model": "gpt-4"}},
		},
		Edges: []domain.WorkflowEdge{},
	}

	wf, err := svc.CreateWorkflow(ctx, projectID, input)
	require.NoError(t, err)
	assert.Equal(t, "Test Workflow", wf.Name)
	assert.Equal(t, projectID, wf.ProjectID)
	assert.Equal(t, domain.WorkflowStatusDraft, wf.Status)
	assert.Equal(t, 1, wf.Version)
	assert.Len(t, wf.Nodes, 1)

	// Empty name should fail
	_, err = svc.CreateWorkflow(ctx, projectID, domain.WorkflowDefinitionInput{Name: ""})
	assert.Error(t, err)

	// No nodes should fail
	_, err = svc.CreateWorkflow(ctx, projectID, domain.WorkflowDefinitionInput{Name: "X"})
	assert.Error(t, err)
}

func TestWorkflowRunSimulation(t *testing.T) {
	logger := zap.NewNop()
	svc := NewWorkflowSimulatorService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	// Create a workflow first to get a valid ID (GetWorkflow returns mock)
	wfID := uuid.New()
	input := domain.SimulationInput{
		WorkflowID: wfID,
		Name:       "Sim Run",
	}

	sim, err := svc.RunSimulation(ctx, projectID, input)
	require.NoError(t, err)
	assert.Equal(t, "completed", sim.Status)
	assert.Equal(t, projectID, sim.ProjectID)
	assert.Equal(t, wfID, sim.WorkflowID)
	// GetWorkflow returns empty nodes, so predictions should be zero
	assert.Equal(t, float64(0), sim.PredictedCostUSD)
	assert.Equal(t, float64(0), sim.PredictedLatencyMs)
	assert.NotNil(t, sim.StartedAt)
	assert.NotNil(t, sim.CompletedAt)
}

func TestWorkflowValidateWorkflow(t *testing.T) {
	logger := zap.NewNop()
	svc := NewWorkflowSimulatorService(logger)

	t.Run("cycle detection", func(t *testing.T) {
		wf := &domain.WorkflowDefinition{
			Nodes: []domain.WorkflowNode{
				{ID: "a", Type: domain.WorkflowNodeTypeLLMCall, Name: "A", Config: map[string]interface{}{"model": "gpt-4"}},
				{ID: "b", Type: domain.WorkflowNodeTypeLLMCall, Name: "B", Config: map[string]interface{}{"model": "gpt-4"}},
			},
			Edges: []domain.WorkflowEdge{
				{ID: "e1", Source: "a", Target: "b"},
				{ID: "e2", Source: "b", Target: "a"},
			},
		}
		errors := svc.ValidateWorkflow(wf)
		assert.Contains(t, errors, "workflow contains a cycle")
	})

	t.Run("disconnected nodes", func(t *testing.T) {
		wf := &domain.WorkflowDefinition{
			Nodes: []domain.WorkflowNode{
				{ID: "a", Type: domain.WorkflowNodeTypeLLMCall, Name: "A", Config: map[string]interface{}{"model": "gpt-4"}},
				{ID: "b", Type: domain.WorkflowNodeTypeLLMCall, Name: "B", Config: map[string]interface{}{"model": "gpt-4"}},
				{ID: "c", Type: domain.WorkflowNodeTypeLLMCall, Name: "C", Config: map[string]interface{}{"model": "gpt-4"}},
			},
			Edges: []domain.WorkflowEdge{
				{ID: "e1", Source: "a", Target: "b"},
			},
		}
		errors := svc.ValidateWorkflow(wf)
		found := false
		for _, e := range errors {
			if e == "node c (C) is disconnected from the workflow" {
				found = true
			}
		}
		assert.True(t, found, "expected disconnected node error, got: %v", errors)
	})
}

func TestWorkflowListWorkflows(t *testing.T) {
	logger := zap.NewNop()
	svc := NewWorkflowSimulatorService(logger)
	ctx := context.Background()

	list, err := svc.ListWorkflows(ctx, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Empty(t, list.Workflows)
	assert.Equal(t, int64(0), list.TotalCount)
	assert.False(t, list.HasMore)
}
