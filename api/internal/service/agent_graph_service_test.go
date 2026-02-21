package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// mockQueryObservationRepo wraps MockObservationRepository to build a QueryService for graph tests.
func buildQueryServiceForGraphTests(observations []domain.Observation) *QueryService {
	obsRepo := new(MockObservationRepository)
	obsRepo.On("GetByTraceID", mock.Anything, mock.Anything, mock.Anything).
		Return(observations, nil)
	// Provide no-ops for other repo interfaces required by NewQueryService
	traceRepo := new(MockTraceRepository)
	return NewQueryService(traceRepo, obsRepo, nil, nil)
}

func TestAgentGraphService_BuildGraph(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()

	t.Run("single generation observation produces one node no edges", func(t *testing.T) {
		observations := []domain.Observation{
			{
				ID:        "obs-1",
				TraceID:   traceID,
				ProjectID: projectID,
				Type:      domain.ObservationTypeGeneration,
				Name:      "agent-alpha",
				Model:     "gpt-4",
				StartTime: time.Now().UTC(),
				EndTime:   ptrTime(time.Now().UTC().Add(time.Second)),
				UsageDetails: domain.UsageDetails{
					InputTokens:  100,
					OutputTokens: 50,
				},
				CostDetails: domain.CostDetails{TotalCost: 0.02},
				DurationMs:  1000,
			},
		}

		querySvc := buildQueryServiceForGraphTests(observations)
		svc := NewAgentGraphService(zap.NewNop(), querySvc)

		graph, err := svc.BuildGraph(context.Background(), projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, graph.Agents, 1)
		assert.Empty(t, graph.Edges)
		assert.Equal(t, "agent-alpha", graph.Agents[0].Name)
		assert.Equal(t, "gpt-4", graph.Agents[0].Model)
		assert.Equal(t, 150, graph.Agents[0].TokensUsed)
		assert.Equal(t, 0.02, graph.TotalCost)
	})

	t.Run("parent-child generations produce node and edge", func(t *testing.T) {
		parentID := "obs-parent"
		observations := []domain.Observation{
			{
				ID:        parentID,
				TraceID:   traceID,
				ProjectID: projectID,
				Type:      domain.ObservationTypeGeneration,
				Name:      "orchestrator",
				Model:     "gpt-4",
				StartTime: time.Now().UTC(),
				EndTime:   ptrTime(time.Now().UTC().Add(2 * time.Second)),
				UsageDetails: domain.UsageDetails{
					InputTokens:  200,
					OutputTokens: 100,
				},
				CostDetails: domain.CostDetails{TotalCost: 0.05},
				DurationMs:  2000,
			},
			{
				ID:                  "obs-child",
				TraceID:             traceID,
				ProjectID:           projectID,
				ParentObservationID: &parentID,
				Type:                domain.ObservationTypeGeneration,
				Name:                "worker",
				Model:               "gpt-4o-mini",
				StartTime:           time.Now().UTC(),
				EndTime:             ptrTime(time.Now().UTC().Add(time.Second)),
				UsageDetails: domain.UsageDetails{
					InputTokens:  50,
					OutputTokens: 25,
				},
				CostDetails: domain.CostDetails{TotalCost: 0.001},
				DurationMs:  1000,
			},
		}

		querySvc := buildQueryServiceForGraphTests(observations)
		svc := NewAgentGraphService(zap.NewNop(), querySvc)

		graph, err := svc.BuildGraph(context.Background(), projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, graph.Agents, 2)
		assert.Len(t, graph.Edges, 1)
		assert.Equal(t, parentID, graph.Edges[0].SourceID)
		assert.Equal(t, "obs-child", graph.Edges[0].TargetID)
	})

	t.Run("span observations are not included as nodes", func(t *testing.T) {
		observations := []domain.Observation{
			{
				ID:        "obs-span",
				TraceID:   traceID,
				ProjectID: projectID,
				Type:      domain.ObservationTypeSpan,
				Name:      "http-request",
				StartTime: time.Now().UTC(),
				DurationMs: 500,
			},
		}

		querySvc := buildQueryServiceForGraphTests(observations)
		svc := NewAgentGraphService(zap.NewNop(), querySvc)

		graph, err := svc.BuildGraph(context.Background(), projectID, traceID)

		require.NoError(t, err)
		assert.Empty(t, graph.Agents)
		assert.Empty(t, graph.Edges)
	})

	t.Run("cost and duration aggregation", func(t *testing.T) {
		parentID := "obs-a"
		observations := []domain.Observation{
			{
				ID:        "obs-a",
				TraceID:   traceID,
				ProjectID: projectID,
				Type:      domain.ObservationTypeGeneration,
				Name:      "agent-a",
				Model:     "gpt-4",
				StartTime: time.Now().UTC(),
				EndTime:   ptrTime(time.Now().UTC().Add(time.Second)),
				UsageDetails: domain.UsageDetails{
					InputTokens:  100,
					OutputTokens: 50,
				},
				CostDetails: domain.CostDetails{TotalCost: 0.03},
				DurationMs:  1000,
			},
			{
				ID:                  "obs-b",
				TraceID:             traceID,
				ProjectID:           projectID,
				ParentObservationID: &parentID,
				Type:                domain.ObservationTypeGeneration,
				Name:                "agent-b",
				Model:               "gpt-4o-mini",
				StartTime:           time.Now().UTC(),
				EndTime:             ptrTime(time.Now().UTC().Add(500 * time.Millisecond)),
				UsageDetails: domain.UsageDetails{
					InputTokens:  30,
					OutputTokens: 20,
				},
				CostDetails: domain.CostDetails{TotalCost: 0.005},
				DurationMs:  500,
			},
		}

		querySvc := buildQueryServiceForGraphTests(observations)
		svc := NewAgentGraphService(zap.NewNop(), querySvc)

		graph, err := svc.BuildGraph(context.Background(), projectID, traceID)

		require.NoError(t, err)
		assert.InDelta(t, 0.035, graph.TotalCost, 0.0001)
		assert.Equal(t, int64(1500), graph.TotalDurationMs)
	})

	t.Run("multi-level graph structure", func(t *testing.T) {
		rootID := "obs-root"
		midID := "obs-mid"
		observations := []domain.Observation{
			{
				ID:        rootID,
				TraceID:   traceID,
				ProjectID: projectID,
				Type:      domain.ObservationTypeGeneration,
				Name:      "root",
				Model:     "gpt-4",
				StartTime: time.Now().UTC(),
				EndTime:   ptrTime(time.Now().UTC().Add(3 * time.Second)),
				UsageDetails: domain.UsageDetails{InputTokens: 100, OutputTokens: 50},
				CostDetails:  domain.CostDetails{TotalCost: 0.04},
				DurationMs:   3000,
			},
			{
				ID:                  midID,
				TraceID:             traceID,
				ProjectID:           projectID,
				ParentObservationID: &rootID,
				Type:                domain.ObservationTypeGeneration,
				Name:                "middle",
				Model:               "gpt-4o",
				StartTime:           time.Now().UTC(),
				EndTime:             ptrTime(time.Now().UTC().Add(2 * time.Second)),
				UsageDetails: domain.UsageDetails{InputTokens: 80, OutputTokens: 40},
				CostDetails:  domain.CostDetails{TotalCost: 0.02},
				DurationMs:   2000,
			},
			{
				ID:                  "obs-leaf",
				TraceID:             traceID,
				ProjectID:           projectID,
				ParentObservationID: &midID,
				Type:                domain.ObservationTypeGeneration,
				Name:                "leaf",
				Model:               "gpt-4o-mini",
				StartTime:           time.Now().UTC(),
				EndTime:             ptrTime(time.Now().UTC().Add(time.Second)),
				UsageDetails: domain.UsageDetails{InputTokens: 20, OutputTokens: 10},
				CostDetails:  domain.CostDetails{TotalCost: 0.001},
				DurationMs:   1000,
			},
		}

		querySvc := buildQueryServiceForGraphTests(observations)
		svc := NewAgentGraphService(zap.NewNop(), querySvc)

		graph, err := svc.BuildGraph(context.Background(), projectID, traceID)

		require.NoError(t, err)
		assert.Len(t, graph.Agents, 3)
		assert.Len(t, graph.Edges, 2)
	})
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
