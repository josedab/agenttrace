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

func newTestCollaborationService() *CollaborationService {
	return NewCollaborationService(zap.NewNop(), nil, NewRealtimeService())
}

func TestCollaborationCreateReview(t *testing.T) {
	svc := newTestCollaborationService()
	ctx := context.Background()
	projectID := uuid.New()
	requestedBy := uuid.New()

	t.Run("valid review", func(t *testing.T) {
		input := &domain.CollabTraceReviewInput{
			TraceID:           "trace-123",
			Title:             "Review unusual agent behavior",
			Description:       "Agent took unexpected path in task execution",
			Priority:          "high",
			RequiredApprovals: 2,
		}
		review, err := svc.CreateReview(ctx, projectID, requestedBy, input)
		require.NoError(t, err)
		require.NotNil(t, review)
		assert.NotEqual(t, uuid.Nil, review.ID)
		assert.Equal(t, "trace-123", review.TraceID)
		assert.Equal(t, "Review unusual agent behavior", review.Title)
		assert.Equal(t, domain.ReviewStatusPending, review.Status)
		assert.Equal(t, 2, review.RequiredApprovals)
		assert.Equal(t, "high", review.Priority)
		assert.Equal(t, projectID, review.ProjectID)
		assert.Equal(t, requestedBy, review.RequestedBy)
	})

	t.Run("default priority and approvals", func(t *testing.T) {
		input := &domain.CollabTraceReviewInput{
			TraceID: "trace-456",
			Title:   "Basic review",
		}
		review, err := svc.CreateReview(ctx, projectID, requestedBy, input)
		require.NoError(t, err)
		assert.Equal(t, "medium", review.Priority)
		assert.Equal(t, 1, review.RequiredApprovals)
	})

	t.Run("empty title fails", func(t *testing.T) {
		input := &domain.CollabTraceReviewInput{
			TraceID: "trace-123",
			Title:   "",
		}
		_, err := svc.CreateReview(ctx, projectID, requestedBy, input)
		assert.Error(t, err)
	})

	t.Run("empty traceID fails", func(t *testing.T) {
		input := &domain.CollabTraceReviewInput{
			TraceID: "",
			Title:   "Test",
		}
		_, err := svc.CreateReview(ctx, projectID, requestedBy, input)
		assert.Error(t, err)
	})
}

func TestCollaborationAddReviewComment(t *testing.T) {
	svc := newTestCollaborationService()
	ctx := context.Background()
	reviewID := uuid.New()
	authorID := uuid.New()
	authorName := "alice"

	t.Run("valid comment", func(t *testing.T) {
		input := &domain.CollabReviewCommentInput{
			Content:  "This looks like a prompt drift issue. @alice can you confirm?",
			Mentions: []string{"alice"},
		}
		comment, err := svc.AddReviewComment(ctx, reviewID, authorID, authorName, input)
		require.NoError(t, err)
		require.NotNil(t, comment)
		assert.NotEqual(t, uuid.Nil, comment.ID)
		assert.Equal(t, reviewID, comment.ReviewID)
		assert.Contains(t, comment.Content, "prompt drift")
		assert.Contains(t, comment.Mentions, "alice")
		assert.False(t, comment.Resolved)
	})

	t.Run("empty content fails", func(t *testing.T) {
		input := &domain.CollabReviewCommentInput{Content: ""}
		_, err := svc.AddReviewComment(ctx, reviewID, authorID, authorName, input)
		assert.Error(t, err)
	})

	t.Run("threaded reply", func(t *testing.T) {
		parentID := uuid.New()
		input := &domain.CollabReviewCommentInput{
			Content:  "Yes, confirmed. The system prompt was modified.",
			ParentID: &parentID,
		}
		comment, err := svc.AddReviewComment(ctx, reviewID, authorID, authorName, input)
		require.NoError(t, err)
		assert.Equal(t, &parentID, comment.ParentID)
	})
}

func TestCollaborationCreateReviewQueue(t *testing.T) {
	svc := newTestCollaborationService()
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid queue", func(t *testing.T) {
		input := &domain.CollabReviewQueueInput{
			Name:           "Production Reviews",
			AssignmentRule: "round_robin",
			SLAHours:       24,
		}
		queue, err := svc.CreateReviewQueue(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, queue)
		assert.Equal(t, "Production Reviews", queue.Name)
		assert.Equal(t, "round_robin", queue.AssignmentRule)
		assert.Equal(t, 24, queue.SLAHours)
		assert.Equal(t, projectID, queue.ProjectID)
	})

	t.Run("default assignment rule and SLA", func(t *testing.T) {
		input := &domain.CollabReviewQueueInput{
			Name: "Default Queue",
		}
		queue, err := svc.CreateReviewQueue(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "round_robin", queue.AssignmentRule)
		assert.Equal(t, 24, queue.SLAHours)
	})

	t.Run("empty name fails", func(t *testing.T) {
		input := &domain.CollabReviewQueueInput{Name: ""}
		_, err := svc.CreateReviewQueue(ctx, projectID, input)
		assert.Error(t, err)
	})
}

func TestCollaborationNotificationIntegration(t *testing.T) {
	svc := newTestCollaborationService()
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid slack integration", func(t *testing.T) {
		input := &domain.NotificationIntegrationInput{
			Type:       "slack",
			Name:       "Team Notifications",
			WebhookURL: "https://hooks.slack.com/services/xxx",
			Events:     []string{"review_created", "approved"},
		}
		integration, err := svc.AddNotificationIntegration(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, integration)
		assert.Equal(t, "slack", integration.Type)
		assert.Equal(t, "Team Notifications", integration.Name)
		assert.True(t, integration.Enabled)
		assert.Equal(t, projectID, integration.ProjectID)
		assert.Equal(t, []string{"review_created", "approved"}, integration.Events)
	})

	t.Run("empty type fails", func(t *testing.T) {
		input := &domain.NotificationIntegrationInput{
			Type: "",
			Name: "Test",
		}
		_, err := svc.AddNotificationIntegration(ctx, projectID, input)
		assert.Error(t, err)
	})

	t.Run("empty name fails", func(t *testing.T) {
		input := &domain.NotificationIntegrationInput{
			Type: "slack",
			Name: "",
		}
		_, err := svc.AddNotificationIntegration(ctx, projectID, input)
		assert.Error(t, err)
	})
}
