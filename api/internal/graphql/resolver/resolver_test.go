package resolver

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/graphql/generated"
)

func TestNewResolver(t *testing.T) {
	logger := zap.NewNop()

	resolver := NewResolver(
		logger,
		nil, // queryService
		nil, // ingestionService
		nil, // scoreService
		nil, // promptService
		nil, // datasetService
		nil, // evalService
		nil, // authService
		nil, // orgService
		nil, // projectService
		nil, // costService
	)

	require.NotNil(t, resolver)
	// Logger is wrapped with Named("graphql"), so just check it's not nil
	assert.NotNil(t, resolver.Logger)
}

func TestContextKeyTypes(t *testing.T) {
	t.Run("context key project ID", func(t *testing.T) {
		assert.Equal(t, ContextKey("projectID"), ContextKeyProjectID)
	})

	t.Run("context key user ID", func(t *testing.T) {
		assert.Equal(t, ContextKey("userID"), ContextKeyUserID)
	})
}

func TestGetProjectIDFromContext(t *testing.T) {
	t.Run("returns error for empty context", func(t *testing.T) {
		ctx := context.Background()
		id, err := getProjectID(ctx)
		assert.Error(t, err)
		assert.Equal(t, uuid.UUID{}, id)
	})

	t.Run("returns project ID from context", func(t *testing.T) {
		projectID := uuid.New()
		ctx := context.WithValue(context.Background(), ContextKeyProjectID, projectID)
		id, err := getProjectID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, projectID, id)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Run("returns error for empty context", func(t *testing.T) {
		ctx := context.Background()
		id, err := getUserID(ctx)
		assert.Error(t, err)
		assert.Equal(t, uuid.UUID{}, id)
	})

	t.Run("returns user ID from context", func(t *testing.T) {
		userID := uuid.New()
		ctx := context.WithValue(context.Background(), ContextKeyUserID, userID)
		id, err := getUserID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, userID, id)
	})
}

func TestContextKeyString(t *testing.T) {
	key := ContextKey("testKey")
	assert.Equal(t, "testKey", string(key))
}

func TestResolver_Mutation(t *testing.T) {
	logger := zap.NewNop()
	resolver := NewResolver(logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	mutation := resolver.Mutation()
	require.NotNil(t, mutation)

	// Verify it implements MutationResolver
	_, ok := mutation.(generated.MutationResolver)
	assert.True(t, ok)
}

func TestContextValueExtraction(t *testing.T) {
	t.Run("extract project ID from context", func(t *testing.T) {
		projectID := uuid.New()
		ctx := context.WithValue(context.Background(), ContextKeyProjectID, projectID)

		extracted, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
		assert.True(t, ok)
		assert.Equal(t, projectID, extracted)
	})

	t.Run("extract user ID from context", func(t *testing.T) {
		userID := uuid.New()
		ctx := context.WithValue(context.Background(), ContextKeyUserID, userID)

		extracted, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
		assert.True(t, ok)
		assert.Equal(t, userID, extracted)
	})

	t.Run("missing project ID returns zero value", func(t *testing.T) {
		ctx := context.Background()

		extracted, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
		assert.False(t, ok)
		assert.Equal(t, uuid.UUID{}, extracted)
	})

	t.Run("wrong type in context returns zero value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ContextKeyProjectID, "not-a-uuid")

		extracted, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
		assert.False(t, ok)
		assert.Equal(t, uuid.UUID{}, extracted)
	})
}
