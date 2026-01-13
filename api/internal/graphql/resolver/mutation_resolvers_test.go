package resolver

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/graphql/model"
)

func TestMutationResolver_CreateTrace_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background() // No project ID in context

	input := model.CreateTraceInput{
		Name: stringPtr("test-trace"),
	}

	trace, err := mutation.CreateTrace(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_UpdateTrace_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.UpdateTraceInput{
		Name: stringPtr("updated-trace"),
	}

	trace, err := mutation.UpdateTrace(ctx, "trace-123", input)
	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_DeleteTrace_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	success, err := mutation.DeleteTrace(ctx, "trace-123")
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateSpan_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateObservationInput{
		TraceId: "trace-123",
		Name:    stringPtr("test-span"),
	}

	obs, err := mutation.CreateSpan(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, obs)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateGeneration_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateGenerationInput{
		TraceId: "trace-123",
		Name:    stringPtr("test-generation"),
	}

	obs, err := mutation.CreateGeneration(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, obs)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateEvent_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateObservationInput{
		TraceId: "trace-123",
		Name:    stringPtr("test-event"),
	}

	obs, err := mutation.CreateEvent(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, obs)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_UpdateObservation_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.UpdateObservationInput{
		Name: stringPtr("updated-observation"),
	}

	obs, err := mutation.UpdateObservation(ctx, "obs-123", input)
	assert.Error(t, err)
	assert.Nil(t, obs)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateScore_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateScoreInput{
		TraceId: "trace-123",
		Name:    "accuracy",
	}

	score, err := mutation.CreateScore(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, score)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_UpdateScore_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.UpdateScoreInput{
		Value: float64Ptr(0.95),
	}

	score, err := mutation.UpdateScore(ctx, "score-123", input)
	assert.Error(t, err)
	assert.Nil(t, score)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_DeleteScore_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	success, err := mutation.DeleteScore(ctx, "score-123")
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_DeleteScore_NilService(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	projectID := uuid.New()
	ctx := context.WithValue(context.Background(), ContextKeyProjectID, projectID)

	success, err := mutation.DeleteScore(ctx, "score-123")
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "score service not configured")
}

func TestMutationResolver_CreatePrompt_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreatePromptInput{
		Name: "test-prompt",
		Type: "text",
	}

	prompt, err := mutation.CreatePrompt(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, prompt)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_UpdatePrompt_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.UpdatePromptInput{
		Description: stringPtr("Updated description"),
	}

	prompt, err := mutation.UpdatePrompt(ctx, "test-prompt", input)
	assert.Error(t, err)
	assert.Nil(t, prompt)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_DeletePrompt_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	success, err := mutation.DeletePrompt(ctx, "test-prompt")
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_SetPromptLabel_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	success, err := mutation.SetPromptLabel(ctx, "test-prompt", 1, "production")
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_RemovePromptLabel_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	success, err := mutation.RemovePromptLabel(ctx, "test-prompt", "production")
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateDataset_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateDatasetInput{
		Name: "test-dataset",
	}

	dataset, err := mutation.CreateDataset(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, dataset)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateEvaluator_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateEvaluatorInput{
		Name: "test-evaluator",
	}

	evaluator, err := mutation.CreateEvaluator(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, evaluator)
	assert.Contains(t, err.Error(), "project ID not found")
}

func TestMutationResolver_CreateOrganization_MissingUserID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	org, err := mutation.CreateOrganization(ctx, "test-org")
	assert.Error(t, err)
	assert.Nil(t, org)
	assert.Contains(t, err.Error(), "user ID not found")
}

func TestMutationResolver_CreateAPIKey_MissingProjectID(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	mutation := resolver.Mutation().(*mutationResolver)

	ctx := context.Background()

	input := model.CreateAPIKeyInput{
		Name:   "test-key",
		Scopes: []string{"read", "write"},
	}

	key, err := mutation.CreateAPIKey(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Contains(t, err.Error(), "project ID not found")
}

// Test helper functions

func TestStringValue(t *testing.T) {
	t.Run("nil returns empty string", func(t *testing.T) {
		result := stringValue(nil)
		assert.Equal(t, "", result)
	})

	t.Run("non-nil returns value", func(t *testing.T) {
		s := "test"
		result := stringValue(&s)
		assert.Equal(t, "test", result)
	})
}

func TestBoolValue(t *testing.T) {
	t.Run("nil returns default", func(t *testing.T) {
		result := boolValue(nil, true)
		assert.True(t, result)

		result = boolValue(nil, false)
		assert.False(t, result)
	})

	t.Run("non-nil returns value", func(t *testing.T) {
		b := true
		result := boolValue(&b, false)
		assert.True(t, result)

		b = false
		result = boolValue(&b, true)
		assert.False(t, result)
	})
}

func TestTimeValue(t *testing.T) {
	t.Run("nil returns default", func(t *testing.T) {
		def := time.Now()
		result := timeValue(nil, def)
		assert.Equal(t, def, result)
	})

	t.Run("non-nil returns value", func(t *testing.T) {
		val := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		def := time.Now()
		result := timeValue(&val, def)
		assert.Equal(t, val, result)
	})
}

func TestTimeValuePtr(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		result := timeValuePtr(nil)
		assert.Nil(t, result)
	})

	t.Run("non-nil returns same pointer", func(t *testing.T) {
		val := time.Now()
		result := timeValuePtr(&val)
		require.NotNil(t, result)
		assert.Equal(t, val, *result)
	})
}

func TestGenerateTraceID(t *testing.T) {
	id1 := generateTraceID()
	id2 := generateTraceID()

	assert.Len(t, id1, 32)
	assert.Len(t, id2, 32)
	assert.NotEqual(t, id1, id2)
}

func TestGenerateSpanID(t *testing.T) {
	id1 := generateSpanID()
	id2 := generateSpanID()

	assert.Len(t, id1, 16)
	assert.Len(t, id2, 16)
	assert.NotEqual(t, id1, id2)
}

func TestSlugify(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"test", "test"},
		{"Test Name", "test-name"},
		{"UPPERCASE", "uppercase"},
		{"with_underscores", "with-underscores"},
		{"multiple   spaces", "multiple-spaces"},
		{"special@chars!", "specialchars"},
		{"123 Numbers", "123-numbers"},
		{"  leading trailing  ", "leading-trailing"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := slugify(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
