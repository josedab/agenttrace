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

// mockGuardrailRepo is a minimal mock for GuardrailRepository used in guardrail tests.
type mockGuardrailRepo struct {
	rules      []domain.GuardRule
	violations []domain.GuardViolation
	err        error
}

func (m *mockGuardrailRepo) SaveRule(_ context.Context, _ *domain.GuardRule) error { return nil }
func (m *mockGuardrailRepo) GetRuleByID(_ context.Context, _ uuid.UUID) (*domain.GuardRule, error) {
	return nil, nil
}
func (m *mockGuardrailRepo) UpdateRule(_ context.Context, _ *domain.GuardRule) error { return nil }
func (m *mockGuardrailRepo) DeleteRule(_ context.Context, _ uuid.UUID) error        { return nil }
func (m *mockGuardrailRepo) ListRules(_ context.Context, _ uuid.UUID) ([]domain.GuardRule, error) {
	return m.rules, m.err
}
func (m *mockGuardrailRepo) ListEnabledRules(_ context.Context, _ uuid.UUID) ([]domain.GuardRule, error) {
	// Return only enabled rules to mirror real behaviour
	var enabled []domain.GuardRule
	for _, r := range m.rules {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled, m.err
}
func (m *mockGuardrailRepo) SaveViolation(_ context.Context, v *domain.GuardViolation) error {
	m.violations = append(m.violations, *v)
	return nil
}
func (m *mockGuardrailRepo) ListViolations(_ context.Context, _ *domain.GuardViolationFilter, _, _ int) ([]domain.GuardViolation, int64, error) {
	return m.violations, int64(len(m.violations)), nil
}

func newGuardrailService(repo *mockGuardrailRepo) *GuardrailService {
	return NewGuardrailService(zap.NewNop(), repo, nil)
}

func TestGuardrailService_Evaluate(t *testing.T) {
	projectID := uuid.New()

	t.Run("cost limit violation", func(t *testing.T) {
		maxCost := 0.50
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeCostLimit,
					Config:  domain.GuardRuleConfig{MaxCostPerTrace: &maxCost},
					Action:  domain.GuardActionAlert,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-1", TotalCost: 1.25}
		result, err := svc.Evaluate(context.Background(), projectID, trace, nil)

		require.NoError(t, err)
		assert.True(t, result.Passed) // alert action does not block
		assert.Len(t, result.Violations, 1)
		assert.Contains(t, result.Violations[0].Details, "exceeds limit")
		assert.Len(t, repo.violations, 1)
	})

	t.Run("cost within limit produces no violation", func(t *testing.T) {
		maxCost := 5.0
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeCostLimit,
					Config:  domain.GuardRuleConfig{MaxCostPerTrace: &maxCost},
					Action:  domain.GuardActionAlert,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-1", TotalCost: 2.00}
		result, err := svc.Evaluate(context.Background(), projectID, trace, nil)

		require.NoError(t, err)
		assert.True(t, result.Passed)
		assert.Empty(t, result.Violations)
	})

	t.Run("latency limit violation", func(t *testing.T) {
		maxLatency := int64(1000)
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeLatencyLimit,
					Config:  domain.GuardRuleConfig{MaxLatencyMs: &maxLatency},
					Action:  domain.GuardActionBlock,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-2", DurationMs: 5000}
		result, err := svc.Evaluate(context.Background(), projectID, trace, nil)

		require.NoError(t, err)
		assert.False(t, result.Passed)
		assert.Len(t, result.Violations, 1)
		assert.Contains(t, result.Violations[0].Details, "latency")
	})

	t.Run("pattern block violation", func(t *testing.T) {
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypePatternBlock,
					Config:  domain.GuardRuleConfig{BlockedPatterns: []string{`secret_key`}},
					Action:  domain.GuardActionBlock,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-3"}
		observations := []domain.Observation{
			{ID: "obs-1", Input: "please use secret_key=abc", Output: "ok"},
		}
		result, err := svc.Evaluate(context.Background(), projectID, trace, observations)

		require.NoError(t, err)
		assert.False(t, result.Passed)
		assert.Len(t, result.Violations, 1)
		assert.Equal(t, domain.GuardViolationSeverityCritical, result.Violations[0].Severity)
	})

	t.Run("multiple rules all evaluated", func(t *testing.T) {
		maxCost := 0.10
		maxLatency := int64(100)
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeCostLimit,
					Config:  domain.GuardRuleConfig{MaxCostPerTrace: &maxCost},
					Action:  domain.GuardActionAlert,
				},
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeLatencyLimit,
					Config:  domain.GuardRuleConfig{MaxLatencyMs: &maxLatency},
					Action:  domain.GuardActionAlert,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-4", TotalCost: 1.0, DurationMs: 5000}
		result, err := svc.Evaluate(context.Background(), projectID, trace, nil)

		require.NoError(t, err)
		assert.Len(t, result.Violations, 2)
		assert.Len(t, repo.violations, 2)
	})

	t.Run("no violations when all rules pass", func(t *testing.T) {
		maxCost := 10.0
		maxLatency := int64(60000)
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeCostLimit,
					Config:  domain.GuardRuleConfig{MaxCostPerTrace: &maxCost},
					Action:  domain.GuardActionAlert,
				},
				{
					ID:      uuid.New(),
					Enabled: true,
					Type:    domain.GuardRuleTypeLatencyLimit,
					Config:  domain.GuardRuleConfig{MaxLatencyMs: &maxLatency},
					Action:  domain.GuardActionBlock,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-5", TotalCost: 0.01, DurationMs: 200}
		result, err := svc.Evaluate(context.Background(), projectID, trace, nil)

		require.NoError(t, err)
		assert.True(t, result.Passed)
		assert.Empty(t, result.Violations)
	})

	t.Run("disabled rules are skipped", func(t *testing.T) {
		maxCost := 0.01
		repo := &mockGuardrailRepo{
			rules: []domain.GuardRule{
				{
					ID:      uuid.New(),
					Enabled: false, // disabled
					Type:    domain.GuardRuleTypeCostLimit,
					Config:  domain.GuardRuleConfig{MaxCostPerTrace: &maxCost},
					Action:  domain.GuardActionBlock,
				},
			},
		}
		svc := newGuardrailService(repo)

		trace := &domain.Trace{ID: "trace-6", TotalCost: 100.0}
		result, err := svc.Evaluate(context.Background(), projectID, trace, nil)

		require.NoError(t, err)
		assert.True(t, result.Passed)
		assert.Empty(t, result.Violations)
		assert.Empty(t, repo.violations)
	})
}
