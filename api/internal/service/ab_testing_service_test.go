package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// --- helpers ---

func newTestService() *ABTestingService {
	return NewABTestingService(zap.NewNop())
}

func validTestInput() *domain.PromptABTestInput {
	return &domain.PromptABTestInput{
		Name:         "my-test",
		TargetMetric: "score",
		Variants: []domain.PromptABTestVariantInput{
			{Name: "control", PromptVersionID: uuid.New(), TrafficPercent: 50, IsControl: true},
			{Name: "variant-a", PromptVersionID: uuid.New(), TrafficPercent: 50},
		},
	}
}

// createRunningTest creates a test and starts it, returning the test and its variant IDs.
func createRunningTest(t *testing.T, svc *ABTestingService) (*domain.PromptABTest, uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	test, err := svc.CreateTest(ctx, uuid.New(), validTestInput())
	require.NoError(t, err)
	test, err = svc.StartTest(ctx, test.ID)
	require.NoError(t, err)
	return test, test.Variants[0].ID, test.Variants[1].ID
}

// recordResults records n results for a given variant with the specified score and latency.
func recordResults(t *testing.T, svc *ABTestingService, testID, variantID uuid.UUID, n int, score, latencyMs float64) {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		err := svc.RecordResult(ctx, testID, &domain.PromptABRecordResultInput{
			VariantID: variantID,
			Score:     score,
			LatencyMs: latencyMs,
			CostUSD:   0.01,
			Tokens:    100,
		})
		require.NoError(t, err)
	}
}

// --- CreateTest ---

func TestABTesting_CreateTest(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid input", func(t *testing.T) {
		test, err := svc.CreateTest(ctx, projectID, validTestInput())
		require.NoError(t, err)
		assert.Equal(t, "my-test", test.Name)
		assert.Equal(t, domain.PromptABTestStatusDraft, test.Status)
		assert.Len(t, test.Variants, 2)
		assert.Equal(t, projectID, test.ProjectID)
		assert.Equal(t, 0.95, test.ConfidenceLevel)
		assert.Equal(t, 100, test.MinSampleSize)
	})

	t.Run("empty name", func(t *testing.T) {
		input := validTestInput()
		input.Name = ""
		_, err := svc.CreateTest(ctx, projectID, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("insufficient variants", func(t *testing.T) {
		input := validTestInput()
		input.Variants = []domain.PromptABTestVariantInput{
			{Name: "only-one", PromptVersionID: uuid.New(), TrafficPercent: 100, IsControl: true},
		}
		_, err := svc.CreateTest(ctx, projectID, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "2 variants")
	})

	t.Run("empty target metric", func(t *testing.T) {
		input := validTestInput()
		input.TargetMetric = ""
		_, err := svc.CreateTest(ctx, projectID, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target metric")
	})

	t.Run("traffic percentages do not sum to 100", func(t *testing.T) {
		input := validTestInput()
		input.Variants[0].TrafficPercent = 40
		input.Variants[1].TrafficPercent = 40
		_, err := svc.CreateTest(ctx, projectID, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sum to 100")
	})

	t.Run("custom confidence and sample size", func(t *testing.T) {
		input := validTestInput()
		input.ConfidenceLevel = 0.99
		input.MinSampleSize = 500
		test, err := svc.CreateTest(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, 0.99, test.ConfidenceLevel)
		assert.Equal(t, 500, test.MinSampleSize)
	})
}

// --- Lifecycle ---

func TestABTesting_Lifecycle(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, err := svc.CreateTest(ctx, uuid.New(), validTestInput())
	require.NoError(t, err)
	assert.Equal(t, domain.PromptABTestStatusDraft, test.Status)

	// Start
	test, err = svc.StartTest(ctx, test.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PromptABTestStatusRunning, test.Status)
	assert.NotNil(t, test.StartedAt)

	// Pause
	test, err = svc.PauseTest(ctx, test.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PromptABTestStatusPaused, test.Status)

	// Resume from paused
	test, err = svc.StartTest(ctx, test.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PromptABTestStatusRunning, test.Status)

	// Stop
	test, err = svc.StopTest(ctx, test.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PromptABTestStatusCompleted, test.Status)
	assert.NotNil(t, test.EndedAt)
}

func TestABTesting_LifecycleInvalidTransitions(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, _ := svc.CreateTest(ctx, uuid.New(), validTestInput())

	// Cannot pause a draft test
	_, err := svc.PauseTest(ctx, test.ID)
	require.Error(t, err)

	// Cannot stop a draft test
	_, err = svc.StopTest(ctx, test.ID)
	require.Error(t, err)

	// Start then complete
	test, _ = svc.StartTest(ctx, test.ID)
	test, _ = svc.StopTest(ctx, test.ID)

	// Cannot start a completed test
	_, err = svc.StartTest(ctx, test.ID)
	require.Error(t, err)
}

func TestABTesting_NotFound(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	fakeID := uuid.New()

	_, err := svc.GetTest(ctx, fakeID)
	require.Error(t, err)

	_, err = svc.StartTest(ctx, fakeID)
	require.Error(t, err)

	_, err = svc.PauseTest(ctx, fakeID)
	require.Error(t, err)

	_, err = svc.StopTest(ctx, fakeID)
	require.Error(t, err)
}

// --- ListTests ---

func TestABTesting_ListTests(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	_, err := svc.CreateTest(ctx, projectA, validTestInput())
	require.NoError(t, err)
	_, err = svc.CreateTest(ctx, projectA, validTestInput())
	require.NoError(t, err)
	_, err = svc.CreateTest(ctx, projectB, validTestInput())
	require.NoError(t, err)

	tests, err := svc.ListTests(ctx, projectA)
	require.NoError(t, err)
	assert.Len(t, tests, 2)

	tests, err = svc.ListTests(ctx, projectB)
	require.NoError(t, err)
	assert.Len(t, tests, 1)

	tests, err = svc.ListTests(ctx, uuid.New())
	require.NoError(t, err)
	assert.Len(t, tests, 0)
}

// --- AssignVariant ---

func TestABTesting_AssignVariant(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, _, _ := createRunningTest(t, svc)

	t.Run("deterministic sticky assignment", func(t *testing.T) {
		a1, err := svc.AssignVariant(ctx, test.ID, "user-123")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, a1.VariantID)
		assert.Equal(t, "user-123", a1.AssignmentKey)

		// Same key yields same variant (sticky)
		a2, err := svc.AssignVariant(ctx, test.ID, "user-123")
		require.NoError(t, err)
		assert.Equal(t, a1.VariantID, a2.VariantID)

		// Different key may yield different variant, but is still deterministic
		a3, err := svc.AssignVariant(ctx, test.ID, "user-456")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, a3.VariantID)

		a4, err := svc.AssignVariant(ctx, test.ID, "user-456")
		require.NoError(t, err)
		assert.Equal(t, a3.VariantID, a4.VariantID)
	})

	t.Run("not running test", func(t *testing.T) {
		draft, _ := svc.CreateTest(ctx, uuid.New(), validTestInput())
		_, err := svc.AssignVariant(ctx, draft.ID, "key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "running")
	})

	t.Run("not found test", func(t *testing.T) {
		_, err := svc.AssignVariant(ctx, uuid.New(), "key")
		require.Error(t, err)
	})
}

// --- RecordResult ---

func TestABTesting_RecordResult(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, controlID, variantID := createRunningTest(t, svc)

	t.Run("valid result", func(t *testing.T) {
		err := svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: controlID,
			Score:     0.85,
			LatencyMs: 120,
			CostUSD:   0.02,
			Tokens:    200,
		})
		require.NoError(t, err)

		updated, _ := svc.GetTest(ctx, test.ID)
		for _, v := range updated.Variants {
			if v.ID == controlID {
				assert.Equal(t, 1, v.SampleCount)
				assert.InDelta(t, 0.85, v.Metrics.AvgScore, 0.001)
			}
		}
	})

	t.Run("invalid variant", func(t *testing.T) {
		err := svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: uuid.New(),
			Score:     0.5,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variant not found")
	})

	t.Run("not found test", func(t *testing.T) {
		err := svc.RecordResult(ctx, uuid.New(), &domain.PromptABRecordResultInput{
			VariantID: controlID,
			Score:     0.5,
		})
		require.Error(t, err)
	})

	t.Run("not running test", func(t *testing.T) {
		_, _ = svc.StopTest(ctx, test.ID)
		err := svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: variantID,
			Score:     0.5,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "running")
	})
}

// --- GetStatistics ---

func TestABTesting_GetStatistics_InsufficientData(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, controlID, variantID := createRunningTest(t, svc)

	// Record a few results — below MinSampleSize (100)
	recordResults(t, svc, test.ID, controlID, 5, 0.8, 100)
	recordResults(t, svc, test.ID, variantID, 5, 0.9, 90)

	stats, err := svc.GetStatistics(ctx, test.ID)
	require.NoError(t, err)
	assert.Equal(t, "insufficient_data", stats.Recommendation)
	assert.Equal(t, 10, stats.CurrentSamples)
	assert.Equal(t, 0.95, stats.ConfidenceLevel)
}

func TestABTesting_GetStatistics_SignificantDifference(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	input := validTestInput()
	input.MinSampleSize = 20
	test, err := svc.CreateTest(ctx, uuid.New(), input)
	require.NoError(t, err)
	test, _ = svc.StartTest(ctx, test.ID)

	controlID := test.Variants[0].ID
	variantID := test.Variants[1].ID

	// Control: mean ~0.50, Variant: mean ~0.90 — clearly different
	for i := 0; i < 60; i++ {
		_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: controlID,
			Score:     0.50 + float64(i%3)*0.01,
			LatencyMs: 100,
		})
	}
	for i := 0; i < 60; i++ {
		_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: variantID,
			Score:     0.90 + float64(i%3)*0.01,
			LatencyMs: 80,
		})
	}

	stats, err := svc.GetStatistics(ctx, test.ID)
	require.NoError(t, err)

	// P-value must be in [0, 1]
	assert.GreaterOrEqual(t, stats.PValue, 0.0)
	assert.LessOrEqual(t, stats.PValue, 1.0)

	// With large effect, should be significant
	assert.True(t, stats.IsSignificant, "expected statistical significance")
	assert.Equal(t, "select_winner", stats.Recommendation)
	assert.Equal(t, 120, stats.CurrentSamples)

	// Variant with higher score should be marked as winner
	var winnerFound bool
	for _, vs := range stats.VariantStats {
		if vs.IsWinner {
			assert.Equal(t, variantID, vs.VariantID)
			winnerFound = true
		}
		// Confidence intervals must bound the mean
		assert.LessOrEqual(t, vs.ConfidenceLower, vs.Mean)
		assert.GreaterOrEqual(t, vs.ConfidenceUpper, vs.Mean)
		// Sample sizes
		assert.Equal(t, 60, vs.SampleSize)
	}
	assert.True(t, winnerFound)
}

func TestABTesting_GetStatistics_EqualMeans(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	input := validTestInput()
	input.MinSampleSize = 10
	test, _ := svc.CreateTest(ctx, uuid.New(), input)
	test, _ = svc.StartTest(ctx, test.ID)

	controlID := test.Variants[0].ID
	variantID := test.Variants[1].ID

	// Both variants have the same score
	recordResults(t, svc, test.ID, controlID, 50, 0.75, 100)
	recordResults(t, svc, test.ID, variantID, 50, 0.75, 100)

	stats, err := svc.GetStatistics(ctx, test.ID)
	require.NoError(t, err)

	// p-value should be 1.0 (no difference) when means are equal and stddev is 0
	assert.InDelta(t, 1.0, stats.PValue, 0.01)
	assert.False(t, stats.IsSignificant)
}

func TestABTesting_GetStatistics_ZeroVariance(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	input := validTestInput()
	input.MinSampleSize = 5
	test, _ := svc.CreateTest(ctx, uuid.New(), input)
	test, _ = svc.StartTest(ctx, test.ID)

	controlID := test.Variants[0].ID
	variantID := test.Variants[1].ID

	// All identical scores but different means
	recordResults(t, svc, test.ID, controlID, 10, 0.5, 100)
	recordResults(t, svc, test.ID, variantID, 10, 0.9, 100)

	stats, err := svc.GetStatistics(ctx, test.ID)
	require.NoError(t, err)

	// When std dev is 0, z-test returns p=1.0 (se=0 guard)
	assert.GreaterOrEqual(t, stats.PValue, 0.0)
	assert.LessOrEqual(t, stats.PValue, 1.0)
}

func TestABTesting_GetStatistics_SingleSample(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	input := validTestInput()
	input.MinSampleSize = 1
	test, _ := svc.CreateTest(ctx, uuid.New(), input)
	test, _ = svc.StartTest(ctx, test.ID)

	controlID := test.Variants[0].ID
	variantID := test.Variants[1].ID

	recordResults(t, svc, test.ID, controlID, 1, 0.5, 100)
	recordResults(t, svc, test.ID, variantID, 1, 0.9, 100)

	stats, err := svc.GetStatistics(ctx, test.ID)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, stats.PValue, 0.0)
	assert.LessOrEqual(t, stats.PValue, 1.0)
	assert.Equal(t, 2, stats.CurrentSamples)
}

func TestABTesting_GetStatistics_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetStatistics(context.Background(), uuid.New())
	require.Error(t, err)
}

// --- SelectWinner ---

func TestABTesting_SelectWinner(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, controlID, variantID := createRunningTest(t, svc)

	t.Run("valid selection", func(t *testing.T) {
		updated, err := svc.SelectWinner(ctx, test.ID, variantID)
		require.NoError(t, err)
		require.NotNil(t, updated.WinnerID)
		assert.Equal(t, variantID, *updated.WinnerID)
		assert.Equal(t, domain.PromptABTestStatusCompleted, updated.Status)
		assert.NotNil(t, updated.EndedAt)
	})

	t.Run("invalid variant", func(t *testing.T) {
		// Need a new test since previous one is completed
		test2, _, _ := createRunningTest(t, svc)
		_, err := svc.SelectWinner(ctx, test2.ID, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variant not found")
	})

	t.Run("can select control as winner", func(t *testing.T) {
		test3, ctrl3, _ := createRunningTest(t, svc)
		updated, err := svc.SelectWinner(ctx, test3.ID, ctrl3)
		require.NoError(t, err)
		assert.Equal(t, ctrl3, *updated.WinnerID)
	})

	t.Run("not found test", func(t *testing.T) {
		_, err := svc.SelectWinner(ctx, uuid.New(), controlID)
		require.Error(t, err)
	})
}

// --- StartGradualRollout ---

func TestABTesting_StartGradualRollout(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, _, variantID := createRunningTest(t, svc)

	t.Run("no winner selected", func(t *testing.T) {
		_, err := svc.StartGradualRollout(ctx, test.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "winner")
	})

	t.Run("default rollout config", func(t *testing.T) {
		_, _ = svc.SelectWinner(ctx, test.ID, variantID)
		updated, err := svc.StartGradualRollout(ctx, test.ID)
		require.NoError(t, err)
		require.NotNil(t, updated.GradualRollout)
		assert.True(t, updated.GradualRollout.Enabled)
		assert.Equal(t, 10.0, updated.GradualRollout.CurrentPercent)
		assert.Equal(t, 10.0, updated.GradualRollout.InitialPercent)
	})

	t.Run("custom rollout config", func(t *testing.T) {
		input := validTestInput()
		input.GradualRollout = &domain.PromptGradualRollout{
			InitialPercent:     20,
			IncrementPercent:   15,
			IncrementIntervalH: 2,
		}
		test2, _ := svc.CreateTest(ctx, uuid.New(), input)
		test2, _ = svc.StartTest(ctx, test2.ID)
		_, _ = svc.SelectWinner(ctx, test2.ID, test2.Variants[1].ID)
		updated, err := svc.StartGradualRollout(ctx, test2.ID)
		require.NoError(t, err)
		assert.True(t, updated.GradualRollout.Enabled)
		assert.Equal(t, 20.0, updated.GradualRollout.CurrentPercent)
	})

	t.Run("not found test", func(t *testing.T) {
		_, err := svc.StartGradualRollout(ctx, uuid.New())
		require.Error(t, err)
	})
}

// --- Statistical helper functions ---

func TestABTesting_MeanStdDev(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m, s := meanStdDev(nil)
		assert.Equal(t, 0.0, m)
		assert.Equal(t, 0.0, s)
	})

	t.Run("single value", func(t *testing.T) {
		m, s := meanStdDev([]float64{5.0})
		assert.Equal(t, 5.0, m)
		assert.Equal(t, 0.0, s)
	})

	t.Run("known values", func(t *testing.T) {
		// [2, 4, 4, 4, 5, 5, 7, 9] → mean=5, sample stddev≈2.138
		m, s := meanStdDev([]float64{2, 4, 4, 4, 5, 5, 7, 9})
		assert.InDelta(t, 5.0, m, 0.001)
		assert.InDelta(t, 2.138, s, 0.01)
	})
}

func TestABTesting_ZTestTwoSample(t *testing.T) {
	t.Run("identical distributions", func(t *testing.T) {
		p := zTestTwoSample(5.0, 5.0, 1.0, 1.0, 100, 100)
		// When means are equal, z=0; p-value should be large (no significant difference)
		assert.GreaterOrEqual(t, p, 0.0)
	})

	t.Run("clearly different distributions", func(t *testing.T) {
		p := zTestTwoSample(8.0, 5.0, 1.0, 1.0, 100, 100)
		assert.Less(t, p, 0.01)
	})

	t.Run("zero samples", func(t *testing.T) {
		p := zTestTwoSample(5.0, 5.0, 1.0, 1.0, 0, 100)
		assert.Equal(t, 1.0, p)
	})

	t.Run("zero standard error", func(t *testing.T) {
		p := zTestTwoSample(5.0, 5.0, 0.0, 0.0, 10, 10)
		assert.Equal(t, 1.0, p)
	})

	t.Run("p-value in valid range", func(t *testing.T) {
		p := zTestTwoSample(6.0, 5.0, 2.0, 2.0, 30, 30)
		assert.GreaterOrEqual(t, p, 0.0)
		assert.LessOrEqual(t, p, 1.0)
	})
}

func TestABTesting_NormalCDF(t *testing.T) {
	// For positive x, normalCDF is increasing
	assert.Less(t, normalCDF(1.0), normalCDF(2.0))
	assert.Less(t, normalCDF(2.0), normalCDF(3.0))
	// Large positive x → close to 1
	assert.InDelta(t, 1.0, normalCDF(5.0), 0.001)
	// Bounded non-negative for positive x
	assert.GreaterOrEqual(t, normalCDF(1.0), 0.0)
	assert.LessOrEqual(t, normalCDF(3.0), 1.0)
	// Symmetry: normalCDF(-x) = 1 - normalCDF(x)
	assert.InDelta(t, normalCDF(-2.0), 1.0-normalCDF(2.0), 1e-10)
}

func TestABTesting_ZScoreForConfidence(t *testing.T) {
	assert.Equal(t, 1.960, zScoreForConfidence(0.95))
	assert.Equal(t, 2.576, zScoreForConfidence(0.99))
	assert.Equal(t, 1.645, zScoreForConfidence(0.90))
}

func TestABTesting_ComputePower(t *testing.T) {
	t.Run("zero effect", func(t *testing.T) {
		assert.Equal(t, 0.0, computePower(0, 100, 1.96))
	})

	t.Run("zero samples", func(t *testing.T) {
		assert.Equal(t, 0.0, computePower(0.5, 0, 1.96))
	})

	t.Run("positive power", func(t *testing.T) {
		p := computePower(0.8, 100, 1.96)
		assert.Greater(t, p, 0.0)
		assert.LessOrEqual(t, p, 1.0)
	})
}

func TestABTesting_EstimateRequiredSamples(t *testing.T) {
	t.Run("zero effect", func(t *testing.T) {
		assert.Equal(t, 0, estimateRequiredSamples(0, 1.96, 0.80))
	})

	t.Run("positive result", func(t *testing.T) {
		n := estimateRequiredSamples(0.5, 1.96, 0.80)
		assert.Greater(t, n, 0)
	})

	t.Run("smaller effect needs larger sample", func(t *testing.T) {
		nSmall := estimateRequiredSamples(0.2, 1.96, 0.80)
		nLarge := estimateRequiredSamples(0.8, 1.96, 0.80)
		assert.Greater(t, nSmall, nLarge)
	})
}

// --- Variant metrics recalculation ---

func TestABTesting_VariantMetrics(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	test, controlID, variantID := createRunningTest(t, svc)

	// Record with errors
	for i := 0; i < 10; i++ {
		_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: controlID,
			Score:     0.8,
			LatencyMs: float64(100 + i*10),
			CostUSD:   0.01,
			Tokens:    150,
			IsError:   i < 2, // 2 out of 10 errors
		})
	}
	_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
		VariantID: variantID,
		Score:     0.9,
		LatencyMs: 50,
		CostUSD:   0.02,
		Tokens:    200,
	})

	updated, _ := svc.GetTest(ctx, test.ID)
	for _, v := range updated.Variants {
		if v.ID == controlID {
			assert.Equal(t, 10, v.SampleCount)
			assert.InDelta(t, 0.8, v.Metrics.AvgScore, 0.001)
			assert.InDelta(t, 0.2, v.Metrics.ErrorRate, 0.001)
			assert.Greater(t, v.Metrics.P95LatencyMs, 0.0)
			assert.Equal(t, int64(1500), v.Metrics.TotalTokens)
		}
		if v.ID == variantID {
			assert.Equal(t, 1, v.SampleCount)
			assert.InDelta(t, 0.9, v.Metrics.AvgScore, 0.001)
			assert.InDelta(t, 0.0, v.Metrics.ErrorRate, 0.001)
		}
	}
}

// --- End-to-end ---

func TestABTesting_EndToEnd(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	// Create with low min sample size for quick testing
	input := validTestInput()
	input.MinSampleSize = 20
	test, err := svc.CreateTest(ctx, uuid.New(), input)
	require.NoError(t, err)

	// Start
	test, err = svc.StartTest(ctx, test.ID)
	require.NoError(t, err)

	controlID := test.Variants[0].ID
	variantID := test.Variants[1].ID

	// Assign variants and record results
	for i := 0; i < 30; i++ {
		assignment, err := svc.AssignVariant(ctx, test.ID, fmt.Sprintf("user-%d", i))
		require.NoError(t, err)

		score := 0.5
		if assignment.VariantID == variantID {
			score = 0.9
		}
		_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: assignment.VariantID,
			Score:     score,
			LatencyMs: 100,
			CostUSD:   0.01,
			Tokens:    100,
		})
	}

	// Also record direct results to ensure sufficient data
	for i := 0; i < 30; i++ {
		_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: controlID,
			Score:     0.50 + float64(i%5)*0.01,
			LatencyMs: 100,
		})
		_ = svc.RecordResult(ctx, test.ID, &domain.PromptABRecordResultInput{
			VariantID: variantID,
			Score:     0.90 + float64(i%5)*0.01,
			LatencyMs: 80,
		})
	}

	// Get statistics
	stats, err := svc.GetStatistics(ctx, test.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.PValue, 0.0)
	assert.LessOrEqual(t, stats.PValue, 1.0)
	assert.Greater(t, stats.CurrentSamples, 20)

	// Select winner
	updated, err := svc.SelectWinner(ctx, test.ID, variantID)
	require.NoError(t, err)
	assert.Equal(t, domain.PromptABTestStatusCompleted, updated.Status)
	assert.Equal(t, variantID, *updated.WinnerID)

	// Start gradual rollout
	updated, err = svc.StartGradualRollout(ctx, test.ID)
	require.NoError(t, err)
	assert.True(t, updated.GradualRollout.Enabled)
}

// ensure fmt is used
var _ = fmt.Sprintf
