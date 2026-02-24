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

func TestCollabCreateReviewQueue(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCollabHubService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	queue, err := svc.CreateReviewQueue(ctx, projectID, "Safety Reviews", "Review queue for safety checks")
	require.NoError(t, err)
	assert.Equal(t, "Safety Reviews", queue.Name)
	assert.Equal(t, projectID, queue.ProjectID)
	assert.Equal(t, 0, queue.PendingCount)
	assert.NotEqual(t, uuid.Nil, queue.ID)

	// Empty name should fail
	_, err = svc.CreateReviewQueue(ctx, projectID, "", "desc")
	assert.Error(t, err)
}

func TestCollabAssignReview(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCollabHubService(logger)
	ctx := context.Background()

	queueID := uuid.New()
	traceID := uuid.New()
	assignTo := uuid.New()

	assignment, err := svc.AssignReview(ctx, queueID, traceID, assignTo)
	require.NoError(t, err)
	assert.Equal(t, queueID, assignment.QueueID)
	assert.Equal(t, traceID, assignment.TraceID)
	assert.Equal(t, assignTo, assignment.AssignedTo)
	assert.Equal(t, domain.ReviewStatusPending, assignment.Status)
}

func TestCollabCompleteReview(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCollabHubService(logger)
	ctx := context.Background()

	score := 0.85
	err := svc.CompleteReview(ctx, uuid.New(), "approved", "Looks good", &score)
	assert.NoError(t, err)

	err = svc.CompleteReview(ctx, uuid.New(), "rejected", "Needs improvement", nil)
	assert.NoError(t, err)
}

func TestCollabCompleteReviewInvalidStatus(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCollabHubService(logger)
	ctx := context.Background()

	// Invalid status should fail
	err := svc.CompleteReview(ctx, uuid.New(), "invalid_status", "feedback", nil)
	assert.Error(t, err)

	// Pending status should fail
	err = svc.CompleteReview(ctx, uuid.New(), "pending", "feedback", nil)
	assert.Error(t, err)

	// Score out of range should fail
	badScore := 1.5
	err = svc.CompleteReview(ctx, uuid.New(), "approved", "feedback", &badScore)
	assert.Error(t, err)
}

func TestCollabCreateQualityStandard(t *testing.T) {
	logger := zap.NewNop()
	svc := NewCollabHubService(logger)
	ctx := context.Background()
	projectID := uuid.New()

	rules := []domain.QualityRule{
		{Metric: "accuracy", Operator: "gte", Threshold: 0.9},
	}

	standard, err := svc.CreateQualityStandard(ctx, projectID, "Production Quality", rules)
	require.NoError(t, err)
	assert.Equal(t, "Production Quality", standard.Name)
	assert.Equal(t, projectID, standard.ProjectID)
	assert.True(t, standard.Enabled)
	assert.Len(t, standard.Rules, 1)

	// No rules should fail
	_, err = svc.CreateQualityStandard(ctx, projectID, "Empty", []domain.QualityRule{})
	assert.Error(t, err)

	// Empty name should fail
	_, err = svc.CreateQualityStandard(ctx, projectID, "", rules)
	assert.Error(t, err)

	// Invalid operator should fail
	_, err = svc.CreateQualityStandard(ctx, projectID, "Bad", []domain.QualityRule{
		{Metric: "accuracy", Operator: "invalid_op", Threshold: 0.9},
	})
	assert.Error(t, err)
}
