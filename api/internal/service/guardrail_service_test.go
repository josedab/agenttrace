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

func TestGuardrailCreateSelfHealingPolicy(t *testing.T) {
	repo := &mockGuardrailRepo{}
	svc := NewGuardrailService(zap.NewNop(), repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid policy with retry", func(t *testing.T) {
		input := &domain.SelfHealingPolicyInput{
			Name:   "Auto-retry on timeout",
			RuleID: uuid.New(),
			RemediationAction: domain.RemediationAction{
				Type:       "retry",
				MaxRetries: 3,
			},
			RetryPolicy: &domain.RetryPolicy{
				MaxAttempts:       3,
				InitialDelayMs:    100,
				MaxDelayMs:        5000,
				BackoffMultiplier: 2.0,
			},
		}
		policy, err := svc.CreateSelfHealingPolicy(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, policy)
		assert.NotEqual(t, uuid.Nil, policy.ID)
		assert.Equal(t, "Auto-retry on timeout", policy.Name)
		assert.Equal(t, "retry", policy.RemediationAction.Type)
		assert.True(t, policy.Enabled)
		assert.NotNil(t, policy.RetryPolicy)
		assert.Equal(t, 3, policy.RetryPolicy.MaxAttempts)
	})

	t.Run("policy with circuit breaker", func(t *testing.T) {
		input := &domain.SelfHealingPolicyInput{
			Name:   "Circuit breaker for model errors",
			RuleID: uuid.New(),
			RemediationAction: domain.RemediationAction{
				Type: "circuit_break",
			},
			CircuitBreaker: &domain.GuardrailCircuitBreakerConfig{
				FailureThreshold: 5,
				SuccessThreshold: 3,
				TimeoutSeconds:   60,
			},
		}
		policy, err := svc.CreateSelfHealingPolicy(ctx, projectID, input)
		require.NoError(t, err)
		assert.NotNil(t, policy.CircuitBreaker)
		assert.Equal(t, "closed", policy.CircuitBreaker.State) // default state
		assert.Equal(t, 5, policy.CircuitBreaker.FailureThreshold)
	})

	t.Run("policy with fallback chain", func(t *testing.T) {
		input := &domain.SelfHealingPolicyInput{
			Name:   "Model fallback chain",
			RuleID: uuid.New(),
			RemediationAction: domain.RemediationAction{
				Type: "fallback",
			},
			FallbackChain: []domain.FallbackStep{
				{Order: 1, Type: "model_switch", Description: "Try GPT-3.5", FallbackModel: "gpt-3.5-turbo"},
				{Order: 2, Type: "cache_response", Description: "Use cached response"},
				{Order: 3, Type: "default_response", Description: "Return default error"},
			},
		}
		policy, err := svc.CreateSelfHealingPolicy(ctx, projectID, input)
		require.NoError(t, err)
		assert.Len(t, policy.FallbackChain, 3)
		assert.Equal(t, "gpt-3.5-turbo", policy.FallbackChain[0].FallbackModel)
	})

	t.Run("empty name fails", func(t *testing.T) {
		input := &domain.SelfHealingPolicyInput{
			Name:   "",
			RuleID: uuid.New(),
		}
		_, err := svc.CreateSelfHealingPolicy(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})
}

func TestGuardrailEvaluatePipeline(t *testing.T) {
	repo := &mockGuardrailRepo{}
	svc := NewGuardrailService(zap.NewNop(), repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("passing all checks", func(t *testing.T) {
		input := &domain.EvalPipelineInput{
			TraceID:   "trace-001",
			Output:    "Valid response output",
			CostUSD:   0.05,
			LatencyMs: 1500,
		}
		result, err := svc.EvaluatePipeline(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Passed)
		assert.Equal(t, "trace-001", result.TraceID)
		assert.NotEmpty(t, result.Evaluations)
		assert.GreaterOrEqual(t, result.TotalLatencyMs, int64(0))
	})

	t.Run("cost limit violation triggers fallback", func(t *testing.T) {
		input := &domain.EvalPipelineInput{
			TraceID: "trace-002",
			CostUSD: 5.0, // exceeds $1.00 limit
		}
		result, err := svc.EvaluatePipeline(ctx, projectID, input)
		require.NoError(t, err)
		assert.True(t, result.Remediated)

		// Find the cost evaluation
		for _, eval := range result.Evaluations {
			if eval.RuleType == "cost_limit" {
				assert.False(t, eval.Passed)
				assert.True(t, eval.Remediated)
				assert.Contains(t, eval.ViolationMsg, "exceeds limit")
			}
		}
	})

	t.Run("latency budget violation", func(t *testing.T) {
		input := &domain.EvalPipelineInput{
			TraceID:   "trace-003",
			LatencyMs: 60000, // exceeds 30000ms budget
		}
		result, err := svc.EvaluatePipeline(ctx, projectID, input)
		require.NoError(t, err)
		assert.False(t, result.Passed)

		for _, eval := range result.Evaluations {
			if eval.RuleType == "latency_budget" {
				assert.False(t, eval.Passed)
				assert.Contains(t, eval.ViolationMsg, "exceeds budget")
			}
		}
	})

	t.Run("output validation - oversized output", func(t *testing.T) {
		hugeOutput := make([]byte, 200000)
		for i := range hugeOutput {
			hugeOutput[i] = 'x'
		}
		input := &domain.EvalPipelineInput{
			TraceID: "trace-004",
			Output:  string(hugeOutput),
		}
		result, err := svc.EvaluatePipeline(ctx, projectID, input)
		require.NoError(t, err)

		for _, eval := range result.Evaluations {
			if eval.RuleType == "output_validation" {
				assert.False(t, eval.Passed)
			}
		}
	})

	t.Run("minimal input - only trace ID", func(t *testing.T) {
		input := &domain.EvalPipelineInput{
			TraceID: "trace-005",
		}
		result, err := svc.EvaluatePipeline(ctx, projectID, input)
		require.NoError(t, err)
		assert.True(t, result.Passed)
		assert.Empty(t, result.Evaluations) // no checks triggered
	})
}

func TestGuardrailDashboardStats(t *testing.T) {
	repo := &mockGuardrailRepo{}
	svc := NewGuardrailService(zap.NewNop(), repo, nil)
	ctx := context.Background()

	stats, err := svc.GetGuardrailDashboardStats(ctx, uuid.New())
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalPolicies, 0)
	assert.GreaterOrEqual(t, stats.RemediationRate, 0.0)
}

func TestGuardrailAuditTrail(t *testing.T) {
	repo := &mockGuardrailRepo{}
	svc := NewGuardrailService(zap.NewNop(), repo, nil)
	ctx := context.Background()

	trail, err := svc.GetPolicyAuditTrail(ctx, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, trail)
}
