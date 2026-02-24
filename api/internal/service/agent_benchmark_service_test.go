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

func TestAgentBenchmarkCreateSuite(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentBenchmarkService(logger)
	ctx := context.Background()

	projectID := uuid.New()
	tasks := []domain.BenchmarkTask{
		{
			ID:          uuid.New(),
			Name:        "Fix null pointer",
			Description: "Fix a null pointer dereference bug",
			Difficulty:  domain.BenchmarkDifficultyMedium,
		},
		{
			ID:          uuid.New(),
			Name:        "Add logging",
			Description: "Add structured logging to the service",
			Difficulty:  domain.BenchmarkDifficultyEasy,
		},
	}

	input := domain.AgentBenchmarkSuiteInput{
		Name:        "Code Quality Suite",
		Description: "Tests for code quality tasks",
		Category:    domain.AgentBenchmarkCategoryBugFix,
		Tasks:       tasks,
	}

	suite, err := svc.CreateSuite(ctx, projectID, input)
	require.NoError(t, err)
	require.NotNil(t, suite)

	assert.NotEqual(t, uuid.Nil, suite.ID)
	assert.Equal(t, projectID, suite.ProjectID)
	assert.Equal(t, "Code Quality Suite", suite.Name)
	assert.Equal(t, "Tests for code quality tasks", suite.Description)
	assert.Equal(t, domain.AgentBenchmarkCategoryBugFix, suite.Category)
	assert.Len(t, suite.Tasks, 2)
	assert.False(t, suite.CreatedAt.IsZero())
}

func TestAgentBenchmarkListSuites(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentBenchmarkService(logger)
	ctx := context.Background()

	suites, err := svc.ListSuites(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, suites)
}

func TestAgentBenchmarkRunBenchmark(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentBenchmarkService(logger)
	ctx := context.Background()

	suiteID := uuid.New()
	run, err := svc.RunBenchmark(ctx, suiteID, "copilot-agent", "gpt-4")
	require.NoError(t, err)
	require.NotNil(t, run)

	assert.NotEqual(t, uuid.Nil, run.ID)
	assert.Equal(t, suiteID, run.SuiteID)
	assert.Equal(t, "copilot-agent", run.AgentName)
	assert.Equal(t, "gpt-4", run.ModelName)
	assert.NotNil(t, run.Results)
	assert.NotNil(t, run.CompletedAt)
	assert.False(t, run.StartedAt.IsZero())
}

func TestAgentBenchmarkGetLeaderboard(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAgentBenchmarkService(logger)
	ctx := context.Background()

	suiteID := uuid.New()
	leaderboard, err := svc.GetLeaderboard(ctx, suiteID)
	require.NoError(t, err)
	require.NotNil(t, leaderboard)

	assert.Equal(t, suiteID, leaderboard.SuiteID)
	assert.NotNil(t, leaderboard.Entries)
	assert.Empty(t, leaderboard.Entries)
	assert.False(t, leaderboard.UpdatedAt.IsZero())
}
