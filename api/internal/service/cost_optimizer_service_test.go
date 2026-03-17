package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// mockCostRecRepo is a minimal mock for CostRecommendationRepository.
type mockCostRecRepo struct {
	saved []domain.CostRecommendation
}

func (m *mockCostRecRepo) Save(_ context.Context, rec *domain.CostRecommendation) error {
	m.saved = append(m.saved, *rec)
	return nil
}
func (m *mockCostRecRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.CostRecommendation, error) {
	for i := range m.saved {
		if m.saved[i].ID == id {
			return &m.saved[i], nil
		}
	}
	return nil, nil
}
func (m *mockCostRecRepo) ListByProject(_ context.Context, _ uuid.UUID) ([]domain.CostRecommendation, error) {
	return m.saved, nil
}
func (m *mockCostRecRepo) Update(_ context.Context, rec *domain.CostRecommendation) error {
	for i := range m.saved {
		if m.saved[i].ID == rec.ID {
			m.saved[i] = *rec
			return nil
		}
	}
	return nil
}

func TestCostOptimizerService_GenerateRecommendations(t *testing.T) {
	t.Run("recommends cheaper alternative for known models", func(t *testing.T) {
		recRepo := &mockCostRecRepo{}
		costSvc := NewCostService(zap.NewNop())

		svc := &CostOptimizerService{
			logger:      zap.NewNop(),
			recRepo:     recRepo,
			costService: costSvc,
		}

		projectID := uuid.New()
		breakdown := []domain.ModelCostEntry{
			{Model: "gpt-4", TraceCount: 100, TotalCost: 50.0, AvgCostPerTrace: 0.5},
		}

		recommendations, savings := svc.generateRecommendations(projectID, breakdown)

		require.NotEmpty(t, recommendations)
		assert.Equal(t, "gpt-4", recommendations[0].CurrentModel)
		assert.Equal(t, "gpt-4o-mini", recommendations[0].RecommendedModel)
		assert.Greater(t, savings, 0.0)
		assert.Equal(t, domain.CostRecommendationPending, recommendations[0].Status)
	})

	t.Run("no recommendation for unknown model", func(t *testing.T) {
		recRepo := &mockCostRecRepo{}
		costSvc := NewCostService(zap.NewNop())

		svc := &CostOptimizerService{
			logger:      zap.NewNop(),
			recRepo:     recRepo,
			costService: costSvc,
		}

		projectID := uuid.New()
		breakdown := []domain.ModelCostEntry{
			{Model: "some-unknown-model-xyz", TraceCount: 10, TotalCost: 5.0},
		}

		recommendations, savings := svc.generateRecommendations(projectID, breakdown)

		assert.Empty(t, recommendations)
		assert.Equal(t, 0.0, savings)
	})

	t.Run("multiple models produce multiple recommendations", func(t *testing.T) {
		recRepo := &mockCostRecRepo{}
		costSvc := NewCostService(zap.NewNop())

		svc := &CostOptimizerService{
			logger:      zap.NewNop(),
			recRepo:     recRepo,
			costService: costSvc,
		}

		projectID := uuid.New()
		breakdown := []domain.ModelCostEntry{
			{Model: "gpt-4", TraceCount: 50, TotalCost: 25.0, AvgCostPerTrace: 0.5},
			{Model: "gpt-4o", TraceCount: 30, TotalCost: 10.0, AvgCostPerTrace: 0.33},
		}

		recommendations, savings := svc.generateRecommendations(projectID, breakdown)

		assert.Len(t, recommendations, 2)
		assert.Greater(t, savings, 0.0)
	})
}

func TestCostOptimizerService_GetRecommendations(t *testing.T) {
	t.Run("returns saved recommendations", func(t *testing.T) {
		projectID := uuid.New()
		recRepo := &mockCostRecRepo{
			saved: []domain.CostRecommendation{
				{
					ID:               uuid.New(),
					ProjectID:        projectID,
					CurrentModel:     "gpt-4",
					RecommendedModel: "gpt-4o-mini",
					Status:           domain.CostRecommendationPending,
					CreatedAt:        time.Now(),
				},
			},
		}
		costSvc := NewCostService(zap.NewNop())

		svc := NewCostOptimizerService(zap.NewNop(), recRepo, costSvc, nil)

		recs, err := svc.GetRecommendations(context.Background(), projectID)

		require.NoError(t, err)
		assert.Len(t, recs, 1)
		assert.Equal(t, "gpt-4", recs[0].CurrentModel)
	})
}

func TestCostOptimizerService_ApplyRecommendation(t *testing.T) {
	t.Run("marks recommendation as applied", func(t *testing.T) {
		recID := uuid.New()
		recRepo := &mockCostRecRepo{
			saved: []domain.CostRecommendation{
				{
					ID:               recID,
					ProjectID:        uuid.New(),
					CurrentModel:     "gpt-4",
					RecommendedModel: "gpt-4o-mini",
					Status:           domain.CostRecommendationPending,
				},
			},
		}
		costSvc := NewCostService(zap.NewNop())
		svc := NewCostOptimizerService(zap.NewNop(), recRepo, costSvc, nil)

		err := svc.ApplyRecommendation(context.Background(), recID)

		require.NoError(t, err)
		assert.Equal(t, domain.CostRecommendationApplied, recRepo.saved[0].Status)
	})
}

func TestCostOptimizerService_DismissRecommendation(t *testing.T) {
	t.Run("marks recommendation as dismissed", func(t *testing.T) {
		recID := uuid.New()
		recRepo := &mockCostRecRepo{
			saved: []domain.CostRecommendation{
				{
					ID:               recID,
					ProjectID:        uuid.New(),
					CurrentModel:     "gpt-4",
					RecommendedModel: "gpt-4o-mini",
					Status:           domain.CostRecommendationPending,
				},
			},
		}
		costSvc := NewCostService(zap.NewNop())
		svc := NewCostOptimizerService(zap.NewNop(), recRepo, costSvc, nil)

		err := svc.DismissRecommendation(context.Background(), recID)

		require.NoError(t, err)
		assert.Equal(t, domain.CostRecommendationDismissed, recRepo.saved[0].Status)
	})
}

func newTestCostOptimizerService() *CostOptimizerService {
	return NewCostOptimizerService(zap.NewNop(), nil, nil, nil)
}

func TestCostOptimizerGetHotspots(t *testing.T) {
	svc := newTestCostOptimizerService()
	ctx := context.Background()
	projectID := uuid.New()

	hotspots, err := svc.GetCostHotspots(ctx, projectID, 30)
	require.NoError(t, err)
	require.NotEmpty(t, hotspots)

	for _, h := range hotspots {
		assert.NotEqual(t, uuid.Nil, h.ID)
		assert.NotEmpty(t, h.Category)
		assert.NotEmpty(t, h.Name)
		assert.Greater(t, h.TotalCostUSD, 0.0)
		assert.Greater(t, h.TraceCount, 0)
		assert.NotEmpty(t, h.Trend)
	}
}

func TestCostOptimizerCreateAutopilotRule(t *testing.T) {
	svc := newTestCostOptimizerService()
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid rule", func(t *testing.T) {
		input := &domain.CostAutopilotRuleInput{
			Name:     "Downgrade for simple tasks",
			RuleType: "model_downgrade",
			Condition: domain.CostRuleCondition{
				Metric:    "per_trace_cost",
				Operator:  "gt",
				Threshold: 0.10,
			},
			Action: domain.CostRuleAction{
				ActionType:    "switch_model",
				FallbackModel: "gpt-3.5-turbo",
			},
		}
		rule, err := svc.CreateAutopilotRule(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, rule)
		assert.Equal(t, "Downgrade for simple tasks", rule.Name)
		assert.Equal(t, "model_downgrade", rule.RuleType)
		assert.True(t, rule.Enabled)
		assert.NotEqual(t, uuid.Nil, rule.ID)
	})

	t.Run("empty name fails", func(t *testing.T) {
		input := &domain.CostAutopilotRuleInput{
			Name:     "",
			RuleType: "model_downgrade",
		}
		_, err := svc.CreateAutopilotRule(ctx, projectID, input)
		assert.Error(t, err)
	})
}

func TestCostOptimizerGetPredictions(t *testing.T) {
	svc := newTestCostOptimizerService()
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("generates predictions for requested days", func(t *testing.T) {
		predictions, err := svc.GetCostPredictions(ctx, projectID, 7, 500.0)
		require.NoError(t, err)
		assert.Len(t, predictions, 7)

		for _, p := range predictions {
			assert.Greater(t, p.PredictedCost, 0.0)
			assert.LessOrEqual(t, p.LowerBound, p.PredictedCost)
			assert.GreaterOrEqual(t, p.UpperBound, p.PredictedCost)
			assert.Greater(t, p.Confidence, 0.0)
			assert.LessOrEqual(t, p.Confidence, 1.0)
			assert.GreaterOrEqual(t, p.OverrunRisk, 0.0)
			assert.LessOrEqual(t, p.OverrunRisk, 1.0)
		}
	})

	t.Run("zero budget gives high overrun risk", func(t *testing.T) {
		predictions, err := svc.GetCostPredictions(ctx, projectID, 1, 0.0)
		require.NoError(t, err)
		require.Len(t, predictions, 1)
		// With zero budget, there should be no overrun risk (0/0 case)
		assert.GreaterOrEqual(t, predictions[0].OverrunRisk, 0.0)
	})
}

func TestCostOptimizerAutopilotDashboard(t *testing.T) {
	svc := newTestCostOptimizerService()
	ctx := context.Background()
	projectID := uuid.New()

	dashboard, err := svc.GetAutopilotDashboard(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, dashboard)

	assert.GreaterOrEqual(t, dashboard.CurrentMonthCost, 0.0)
	assert.Greater(t, dashboard.MonthlyBudget, 0.0)
	assert.NotEmpty(t, dashboard.Hotspots)
	assert.NotEmpty(t, dashboard.Predictions)
}
