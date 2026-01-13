package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/graphql/model"
)

func TestTracesInput(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		input := model.TracesInput{}
		assert.Nil(t, input.Limit)
		assert.Nil(t, input.Cursor)
		assert.Nil(t, input.UserID)
		assert.Nil(t, input.SessionID)
		assert.Nil(t, input.Name)
		assert.Empty(t, input.Tags)
	})

	t.Run("with values", func(t *testing.T) {
		limit := 25
		cursor := "cursor-123"
		userID := "user-456"
		sessionID := "session-789"
		name := "trace-name"
		now := time.Now()
		version := "1.0.0"
		release := "release-1"
		orderBy := "timestamp"
		order := domain.SortOrderDesc

		input := model.TracesInput{
			Limit:         &limit,
			Cursor:        &cursor,
			UserID:        &userID,
			SessionID:     &sessionID,
			Name:          &name,
			Tags:          []string{"tag1", "tag2"},
			FromTimestamp: &now,
			ToTimestamp:   &now,
			Version:       &version,
			Release:       &release,
			OrderBy:       &orderBy,
			Order:         &order,
		}

		assert.Equal(t, 25, *input.Limit)
		assert.Equal(t, "cursor-123", *input.Cursor)
		assert.Equal(t, "user-456", *input.UserID)
		assert.Equal(t, "session-789", *input.SessionID)
		assert.Equal(t, "trace-name", *input.Name)
		assert.Len(t, input.Tags, 2)
	})
}

func TestTraceConnection(t *testing.T) {
	t.Run("empty connection", func(t *testing.T) {
		conn := model.TraceConnection{
			Edges:      []*model.TraceEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.Empty(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})
}

func TestPageInfo(t *testing.T) {
	t.Run("no pagination", func(t *testing.T) {
		info := model.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		}

		assert.False(t, info.HasNextPage)
		assert.False(t, info.HasPreviousPage)
		assert.Nil(t, info.StartCursor)
		assert.Nil(t, info.EndCursor)
	})

	t.Run("with cursors", func(t *testing.T) {
		start := "start-cursor"
		end := "end-cursor"

		info := model.PageInfo{
			HasNextPage:     true,
			HasPreviousPage: true,
			StartCursor:     &start,
			EndCursor:       &end,
		}

		assert.True(t, info.HasNextPage)
		assert.True(t, info.HasPreviousPage)
		assert.Equal(t, "start-cursor", *info.StartCursor)
		assert.Equal(t, "end-cursor", *info.EndCursor)
	})
}

func TestObservationsInput(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		input := model.ObservationsInput{}
		assert.Nil(t, input.TraceID)
		assert.Nil(t, input.ParentObservationID)
		assert.Nil(t, input.Type)
		assert.Nil(t, input.Name)
		assert.Nil(t, input.Limit)
		assert.Nil(t, input.Cursor)
	})

	t.Run("with trace ID", func(t *testing.T) {
		traceID := "trace-123"
		limit := 10

		input := model.ObservationsInput{
			TraceID: &traceID,
			Limit:   &limit,
		}

		assert.Equal(t, "trace-123", *input.TraceID)
		assert.Equal(t, 10, *input.Limit)
	})
}

func TestScoresInput(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		input := model.ScoresInput{}
		assert.Nil(t, input.TraceID)
		assert.Nil(t, input.ObservationID)
		assert.Nil(t, input.Name)
		assert.Nil(t, input.Source)
		assert.Nil(t, input.Limit)
		assert.Nil(t, input.Cursor)
	})
}

func TestSessionsInput(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		input := model.SessionsInput{}
		assert.Nil(t, input.Limit)
		assert.Nil(t, input.Cursor)
		assert.Nil(t, input.FromTimestamp)
		assert.Nil(t, input.ToTimestamp)
	})
}

func TestPromptsInput(t *testing.T) {
	t.Run("with filters", func(t *testing.T) {
		name := "prompt-name"
		label := "production"
		limit := 20

		input := model.PromptsInput{
			Name:  &name,
			Label: &label,
			Tags:  []string{"tag1"},
			Limit: &limit,
		}

		assert.Equal(t, "prompt-name", *input.Name)
		assert.Equal(t, "production", *input.Label)
		assert.Len(t, input.Tags, 1)
		assert.Equal(t, 20, *input.Limit)
	})
}

func TestDatasetsInput(t *testing.T) {
	t.Run("with name filter", func(t *testing.T) {
		name := "dataset-name"
		limit := 15

		input := model.DatasetsInput{
			Name:  &name,
			Limit: &limit,
		}

		assert.Equal(t, "dataset-name", *input.Name)
		assert.Equal(t, 15, *input.Limit)
	})
}

func TestEvaluatorsInput(t *testing.T) {
	t.Run("with filters", func(t *testing.T) {
		enabled := true
		limit := 25

		input := model.EvaluatorsInput{
			Enabled: &enabled,
			Limit:   &limit,
		}

		assert.True(t, *input.Enabled)
		assert.Equal(t, 25, *input.Limit)
	})
}

func TestMetricsInput(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now()

		input := model.MetricsInput{
			FromTimestamp: from,
			ToTimestamp:   to,
		}

		assert.False(t, input.FromTimestamp.IsZero())
		assert.False(t, input.ToTimestamp.IsZero())
	})

	t.Run("with optional filters", func(t *testing.T) {
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now()
		userID := "user-123"
		sessionID := "session-456"
		name := "trace-name"

		input := model.MetricsInput{
			FromTimestamp: from,
			ToTimestamp:   to,
			UserId:        &userID,
			SessionId:     &sessionID,
			Name:          &name,
			Tags:          []string{"tag1", "tag2"},
		}

		assert.Equal(t, "user-123", *input.UserId)
		assert.Equal(t, "session-456", *input.SessionId)
		assert.Equal(t, "trace-name", *input.Name)
		assert.Len(t, input.Tags, 2)
	})
}

func TestMetrics(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		metrics := model.Metrics{
			TraceCount:       0,
			ObservationCount: 0,
			TotalCost:        0.0,
			TotalTokens:      0,
			ModelUsage:       []*model.ModelUsage{},
		}

		assert.Equal(t, 0, metrics.TraceCount)
		assert.Equal(t, 0, metrics.ObservationCount)
		assert.Equal(t, 0.0, metrics.TotalCost)
		assert.Equal(t, 0, metrics.TotalTokens)
		assert.Empty(t, metrics.ModelUsage)
	})

	t.Run("with latency metrics", func(t *testing.T) {
		avg := 150.5
		p50 := 120.0
		p95 := 350.0
		p99 := 500.0

		metrics := model.Metrics{
			TraceCount:  100,
			TotalCost:   25.50,
			TotalTokens: 50000,
			AvgLatency:  &avg,
			P50Latency:  &p50,
			P95Latency:  &p95,
			P99Latency:  &p99,
		}

		assert.Equal(t, 100, metrics.TraceCount)
		assert.Equal(t, 25.50, metrics.TotalCost)
		assert.Equal(t, 50000, metrics.TotalTokens)
		assert.Equal(t, 150.5, *metrics.AvgLatency)
		assert.Equal(t, 120.0, *metrics.P50Latency)
		assert.Equal(t, 350.0, *metrics.P95Latency)
		assert.Equal(t, 500.0, *metrics.P99Latency)
	})
}

func TestModelUsage(t *testing.T) {
	t.Run("model usage structure", func(t *testing.T) {
		usage := model.ModelUsage{
			Model:            "gpt-4",
			Count:            50,
			PromptTokens:     10000,
			CompletionTokens: 5000,
			TotalTokens:      15000,
			Cost:             12.50,
		}

		assert.Equal(t, "gpt-4", usage.Model)
		assert.Equal(t, 50, usage.Count)
		assert.Equal(t, 10000, usage.PromptTokens)
		assert.Equal(t, 5000, usage.CompletionTokens)
		assert.Equal(t, 15000, usage.TotalTokens)
		assert.Equal(t, 12.50, usage.Cost)
	})
}

func TestDailyCostsInput(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		input := model.DailyCostsInput{
			FromDate: "2024-01-01",
			ToDate:   "2024-01-31",
		}

		assert.Equal(t, "2024-01-01", input.FromDate)
		assert.Equal(t, "2024-01-31", input.ToDate)
	})

	t.Run("with groupBy", func(t *testing.T) {
		groupBy := "model"
		input := model.DailyCostsInput{
			FromDate: "2024-01-01",
			ToDate:   "2024-01-31",
			GroupBy:  &groupBy,
		}

		assert.Equal(t, "model", *input.GroupBy)
	})
}

func TestDailyCost(t *testing.T) {
	t.Run("daily cost structure", func(t *testing.T) {
		cost := model.DailyCost{
			Date:       "2024-01-15",
			TotalCost:  5.25,
			TraceCount: 25,
			ModelCosts: []*model.ModelCost{
				{Model: "gpt-4", Cost: 3.50},
				{Model: "gpt-3.5-turbo", Cost: 1.75},
			},
		}

		assert.Equal(t, "2024-01-15", cost.Date)
		assert.Equal(t, 5.25, cost.TotalCost)
		assert.Equal(t, 25, cost.TraceCount)
		assert.Len(t, cost.ModelCosts, 2)
	})
}

func TestModelCost(t *testing.T) {
	t.Run("model cost structure", func(t *testing.T) {
		cost := model.ModelCost{
			Model: "claude-3-opus",
			Cost:  15.75,
			Count: 30,
		}

		assert.Equal(t, "claude-3-opus", cost.Model)
		assert.Equal(t, 15.75, cost.Cost)
		assert.Equal(t, 30, cost.Count)
	})
}

func TestConnectionTypes(t *testing.T) {
	t.Run("ObservationConnection", func(t *testing.T) {
		conn := model.ObservationConnection{
			Edges:      []*model.ObservationEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.NotNil(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})

	t.Run("ScoreConnection", func(t *testing.T) {
		conn := model.ScoreConnection{
			Edges:      []*model.ScoreEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.NotNil(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})

	t.Run("SessionConnection", func(t *testing.T) {
		conn := model.SessionConnection{
			Edges:      []*model.SessionEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.NotNil(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})

	t.Run("PromptConnection", func(t *testing.T) {
		conn := model.PromptConnection{
			Edges:      []*model.PromptEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.NotNil(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})

	t.Run("DatasetConnection", func(t *testing.T) {
		conn := model.DatasetConnection{
			Edges:      []*model.DatasetEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.NotNil(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})

	t.Run("EvaluatorConnection", func(t *testing.T) {
		conn := model.EvaluatorConnection{
			Edges:      []*model.EvaluatorEdge{},
			PageInfo:   &model.PageInfo{},
			TotalCount: 0,
		}

		assert.NotNil(t, conn.Edges)
		assert.NotNil(t, conn.PageInfo)
		assert.Equal(t, 0, conn.TotalCount)
	})
}
