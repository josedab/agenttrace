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

func TestNewSandboxService(t *testing.T) {
	svc := NewSandboxService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestSandboxService_ReviewLifecycle(t *testing.T) {
	svc := NewSandboxService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("submits review with risk assessment", func(t *testing.T) {
		input := &domain.SandboxReviewInput{
			TraceID: uuid.New(),
			ProposedActions: []domain.SandboxAction{
				{ID: "a1", Type: "file_write", Target: "main.go", Description: "Write main file"},
				{ID: "a2", Type: "command_exec", Target: "go build", Description: "Build project"},
			},
		}

		review, err := svc.SubmitForReview(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, domain.SandboxStatusPending, review.Status)
		assert.Greater(t, review.RiskScore, 0.0)
		assert.NotEmpty(t, review.RiskLevel)
		assert.Len(t, review.ProposedActions, 2)
	})

	t.Run("assesses high risk for sensitive files", func(t *testing.T) {
		input := &domain.SandboxReviewInput{
			TraceID: uuid.New(),
			ProposedActions: []domain.SandboxAction{
				{ID: "a1", Type: "file_write", Target: ".env.secret", Description: "Write secrets"},
				{ID: "a2", Type: "file_delete", Target: "/important/data", Description: "Delete data"},
			},
		}

		review, err := svc.SubmitForReview(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "critical", review.RiskLevel)
		assert.GreaterOrEqual(t, review.RiskScore, 80.0)
	})

	t.Run("approves review", func(t *testing.T) {
		input := &domain.SandboxReviewInput{
			TraceID:         uuid.New(),
			ProposedActions: []domain.SandboxAction{{ID: "a1", Type: "file_write", Target: "readme.md"}},
		}
		review, _ := svc.SubmitForReview(ctx, projectID, input)

		decided, err := svc.ReviewDecision(ctx, review.ID, &domain.SandboxDecision{Action: "approve", Note: "LGTM"})
		require.NoError(t, err)
		assert.Equal(t, domain.SandboxStatusApproved, decided.Status)
		assert.NotNil(t, decided.ReviewedAt)
		for _, a := range decided.ProposedActions {
			assert.True(t, *a.Approved)
		}
	})

	t.Run("rejects review", func(t *testing.T) {
		input := &domain.SandboxReviewInput{
			TraceID:         uuid.New(),
			ProposedActions: []domain.SandboxAction{{ID: "a1", Type: "file_delete", Target: ".env"}},
		}
		review, _ := svc.SubmitForReview(ctx, projectID, input)

		decided, err := svc.ReviewDecision(ctx, review.ID, &domain.SandboxDecision{Action: "reject", Note: "Too risky"})
		require.NoError(t, err)
		assert.Equal(t, domain.SandboxStatusRejected, decided.Status)
	})

	t.Run("partial approval", func(t *testing.T) {
		input := &domain.SandboxReviewInput{
			TraceID: uuid.New(),
			ProposedActions: []domain.SandboxAction{
				{ID: "safe", Type: "file_write", Target: "main.go"},
				{ID: "risky", Type: "command_exec", Target: "rm -rf /"},
			},
		}
		review, _ := svc.SubmitForReview(ctx, projectID, input)

		decided, err := svc.ReviewDecision(ctx, review.ID, &domain.SandboxDecision{
			Action: "approve_partial", ActionIDs: []string{"safe"},
		})
		require.NoError(t, err)
		for _, a := range decided.ProposedActions {
			if a.ID == "safe" {
				assert.True(t, *a.Approved)
			} else {
				assert.False(t, *a.Approved)
			}
		}
	})

	t.Run("lists pending reviews", func(t *testing.T) {
		pending, err := svc.ListPendingReviews(ctx, projectID)
		require.NoError(t, err)
		for _, r := range pending {
			assert.Equal(t, domain.SandboxStatusPending, r.Status)
		}
	})
}

func TestSandboxService_Policies(t *testing.T) {
	svc := NewSandboxService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("creates policy with defaults", func(t *testing.T) {
		input := &domain.SandboxPolicyInput{Name: "default-policy"}
		policy, err := svc.CreatePolicy(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "default-policy", policy.Name)
		assert.Equal(t, "high_risk", policy.RequireReview)
		assert.False(t, policy.AllowNetwork)
	})

	t.Run("creates policy with custom settings", func(t *testing.T) {
		allowNet := true
		maxSize := int64(1048576)
		input := &domain.SandboxPolicyInput{
			Name: "custom", AllowNetwork: &allowNet,
			BlockedPaths: []string{".env", "*.key"},
			RequireReview: "always", MaxFileSize: &maxSize,
		}
		policy, err := svc.CreatePolicy(ctx, projectID, input)
		require.NoError(t, err)
		assert.True(t, policy.AllowNetwork)
		assert.Equal(t, "always", policy.RequireReview)
		assert.Equal(t, int64(1048576), policy.MaxFileSize)
	})

	t.Run("lists policies for project", func(t *testing.T) {
		policies, err := svc.ListPolicies(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, policies, 2)
	})
}

func TestSandboxService_Stats(t *testing.T) {
	svc := NewSandboxService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	// Create some reviews
	for i := 0; i < 3; i++ {
		svc.SubmitForReview(ctx, projectID, &domain.SandboxReviewInput{
			TraceID:         uuid.New(),
			ProposedActions: []domain.SandboxAction{{ID: "a", Type: "file_write", Target: "test.go"}},
		})
	}

	stats, err := svc.GetStats(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalReviews)
	assert.Equal(t, 3, stats.PendingReviews)
}
