package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockCostBudgetRepository is a mock implementation of the CostBudgetRepository
type MockCostBudgetRepository struct {
	mock.Mock
}

func (m *MockCostBudgetRepository) Save(ctx context.Context, budget *domain.CostBudget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockCostBudgetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CostBudget, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CostBudget), args.Error(1)
}

func (m *MockCostBudgetRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.CostBudget, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CostBudget), args.Error(1)
}

func (m *MockCostBudgetRepository) Update(ctx context.Context, budget *domain.CostBudget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockCostBudgetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCostBudgetService_CheckBudget(t *testing.T) {
	t.Run("allows when under budget", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000, // $1000
				CurrentSpendCents: 50000,  // $500
				AutoAction:        domain.BudgetAutoActionBlock,
				Enabled:           true,
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 1000) // $10 additional

		require.NoError(t, err)
		assert.True(t, allowed, "should allow when under budget")
		assert.Equal(t, domain.BudgetAutoActionNone, action)
	})

	t.Run("returns auto-action when over budget", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000, // $1000
				CurrentSpendCents: 95000,  // $950
				AutoAction:        domain.BudgetAutoActionBlock,
				Enabled:           true,
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 10000) // $100 additional -> exceeds

		require.NoError(t, err)
		assert.False(t, allowed, "should deny when over budget")
		assert.Equal(t, domain.BudgetAutoActionBlock, action)
	})

	t.Run("returns THROTTLE action when configured", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 50000,
				CurrentSpendCents: 49000,
				AutoAction:        domain.BudgetAutoActionThrottle,
				Enabled:           true,
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 2000)

		require.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, domain.BudgetAutoActionThrottle, action)
	})

	t.Run("allows when no budgets exist", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		repo.On("ListByProject", mock.Anything, projectID).Return([]domain.CostBudget{}, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 50000)

		require.NoError(t, err)
		assert.True(t, allowed, "should allow when no budgets exist")
		assert.Equal(t, domain.BudgetAutoActionNone, action)
	})

	t.Run("allows when budget is disabled", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 10000,
				CurrentSpendCents: 50000, // over limit but disabled
				AutoAction:        domain.BudgetAutoActionBlock,
				Enabled:           false,
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 5000)

		require.NoError(t, err)
		assert.True(t, allowed, "should allow when budget is disabled")
		assert.Equal(t, domain.BudgetAutoActionNone, action)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		repo.On("ListByProject", mock.Anything, projectID).Return(nil, fmt.Errorf("database error"))

		_, _, err := svc.CheckBudget(ctx, projectID, 1000)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list budgets")
	})
}

func TestCostBudgetService_EvaluateThresholds(t *testing.T) {
	t.Run("returns crossed thresholds that are not yet notified", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000,
				CurrentSpendCents: 85000, // 85%
				AutoAction:        domain.BudgetAutoActionAlertOnly,
				Enabled:           true,
				AlertThresholds: []domain.BudgetThreshold{
					{Percent: 50, NotifyVia: "email", Notified: true},  // already notified
					{Percent: 80, NotifyVia: "slack", Notified: false}, // should be crossed
					{Percent: 100, NotifyVia: "webhook", Notified: false}, // not yet crossed
				},
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.CostBudget")).Return(nil)

		crossed, err := svc.EvaluateThresholds(ctx, projectID)

		require.NoError(t, err)
		require.Len(t, crossed, 1)
		assert.Equal(t, 80, crossed[0].Percent)
		assert.Equal(t, "slack", crossed[0].NotifyVia)
	})

	t.Run("returns multiple crossed thresholds", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000,
				CurrentSpendCents: 100000, // 100%
				AutoAction:        domain.BudgetAutoActionBlock,
				Enabled:           true,
				AlertThresholds: []domain.BudgetThreshold{
					{Percent: 50, NotifyVia: "email", Notified: false},
					{Percent: 80, NotifyVia: "slack", Notified: false},
					{Percent: 100, NotifyVia: "webhook", Notified: false},
				},
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.CostBudget")).Return(nil)

		crossed, err := svc.EvaluateThresholds(ctx, projectID)

		require.NoError(t, err)
		assert.Len(t, crossed, 3)
	})

	t.Run("returns empty when no thresholds crossed", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000,
				CurrentSpendCents: 10000, // 10%
				AutoAction:        domain.BudgetAutoActionNone,
				Enabled:           true,
				AlertThresholds: []domain.BudgetThreshold{
					{Percent: 50, NotifyVia: "email", Notified: false},
					{Percent: 80, NotifyVia: "slack", Notified: false},
				},
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		crossed, err := svc.EvaluateThresholds(ctx, projectID)

		require.NoError(t, err)
		assert.Empty(t, crossed)
	})

	t.Run("skips disabled budgets", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000,
				CurrentSpendCents: 90000, // 90% but disabled
				Enabled:           false,
				AlertThresholds: []domain.BudgetThreshold{
					{Percent: 50, NotifyVia: "email", Notified: false},
					{Percent: 80, NotifyVia: "slack", Notified: false},
				},
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		crossed, err := svc.EvaluateThresholds(ctx, projectID)

		require.NoError(t, err)
		assert.Empty(t, crossed)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		repo.On("ListByProject", mock.Anything, projectID).Return(nil, fmt.Errorf("database error"))

		_, err := svc.EvaluateThresholds(ctx, projectID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list budgets")
	})
}

func TestCostBudgetService_CreateBudget(t *testing.T) {
	t.Run("creates budget with correct defaults", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.CostBudget")).Return(nil)

		input := domain.CostBudgetInput{
			Name:              "Production Budget",
			MonthlyLimitCents: 500000,
			AlertThresholds: []domain.BudgetThreshold{
				{Percent: 80, NotifyVia: "email"},
			},
			AutoAction: domain.BudgetAutoActionAlertOnly,
		}

		result, err := svc.CreateBudget(ctx, projectID, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Production Budget", result.Name)
		assert.Equal(t, int64(500000), result.MonthlyLimitCents)
		assert.Equal(t, domain.BudgetAutoActionAlertOnly, result.AutoAction)
		assert.True(t, result.Enabled)
		assert.Equal(t, projectID, result.ProjectID)
		assert.NotEqual(t, uuid.Nil, result.ID)
	})
}

func TestCostBudgetService_ForecastExhaustion(t *testing.T) {
	t.Run("budget at exact limit denies additional spend", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000,
				CurrentSpendCents: 100000, // exactly at limit
				AutoAction:        domain.BudgetAutoActionSwitchModel,
				Enabled:           true,
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 1) // even $0.01 more

		require.NoError(t, err)
		assert.False(t, allowed)
		assert.Equal(t, domain.BudgetAutoActionSwitchModel, action)
	})

	t.Run("budget allows spend at exactly the remaining amount", func(t *testing.T) {
		repo := new(MockCostBudgetRepository)
		svc := NewCostBudgetService(zap.NewNop(), repo, nil, nil)

		projectID := uuid.New()
		ctx := context.Background()

		budgets := []domain.CostBudget{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				MonthlyLimitCents: 100000,
				CurrentSpendCents: 90000,
				AutoAction:        domain.BudgetAutoActionBlock,
				Enabled:           true,
			},
		}

		repo.On("ListByProject", mock.Anything, projectID).Return(budgets, nil)

		allowed, action, err := svc.CheckBudget(ctx, projectID, 10000) // exactly at limit

		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, domain.BudgetAutoActionNone, action)
	})
}
