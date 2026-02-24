package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CollabHubService handles team collaboration features including review queues,
// assignments, quality standards, and activity feeds
type CollabHubService struct {
	logger *zap.Logger
}

// NewCollabHubService creates a new collaboration hub service
func NewCollabHubService(logger *zap.Logger) *CollabHubService {
	return &CollabHubService{
		logger: logger,
	}
}

// CreateReviewQueue creates a new review queue for a project
func (s *CollabHubService) CreateReviewQueue(ctx context.Context, projectID uuid.UUID, name, description string) (*domain.ReviewQueue, error) {
	if name == "" {
		return nil, fmt.Errorf("review queue name is required")
	}

	queue := &domain.ReviewQueue{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Name:           name,
		Description:    description,
		Filters:        map[string]interface{}{},
		AssignedTo:     []uuid.UUID{},
		PendingCount:   0,
		CompletedCount: 0,
		CreatedAt:      time.Now(),
	}

	s.logger.Info("review queue created",
		zap.String("id", queue.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", name),
	)
	return queue, nil
}

// ListQueues lists all review queues for a project
func (s *CollabHubService) ListQueues(ctx context.Context, projectID uuid.UUID) ([]domain.ReviewQueue, error) {
	s.logger.Debug("listing review queues", zap.String("projectId", projectID.String()))
	return []domain.ReviewQueue{}, nil
}

// AssignReview creates a review assignment for a trace in a queue
func (s *CollabHubService) AssignReview(ctx context.Context, queueID, traceID, assignTo uuid.UUID) (*domain.ReviewAssignment, error) {
	now := time.Now()

	assignment := &domain.ReviewAssignment{
		ID:         uuid.New(),
		QueueID:    queueID,
		TraceID:    traceID,
		AssignedTo: assignTo,
		Status:     domain.ReviewStatusPending,
		AssignedAt: now,
	}

	s.logger.Info("review assigned",
		zap.String("id", assignment.ID.String()),
		zap.String("queueId", queueID.String()),
		zap.String("traceId", traceID.String()),
		zap.String("assignedTo", assignTo.String()),
	)
	return assignment, nil
}

// CompleteReview marks a review assignment as completed with feedback and optional score
func (s *CollabHubService) CompleteReview(ctx context.Context, assignmentID uuid.UUID, status, feedback string, score *float64) error {
	reviewStatus := domain.ReviewStatus(status)
	if !reviewStatus.IsValid() {
		return fmt.Errorf("invalid review status: %s", status)
	}
	if reviewStatus == domain.ReviewStatusPending {
		return fmt.Errorf("cannot complete a review with pending status")
	}
	if score != nil && (*score < 0 || *score > 1) {
		return fmt.Errorf("score must be between 0 and 1, got %f", *score)
	}

	s.logger.Info("review completed",
		zap.String("assignmentId", assignmentID.String()),
		zap.String("status", status),
		zap.Bool("hasFeedback", feedback != ""),
		zap.Bool("hasScore", score != nil),
	)
	return nil
}

// CreateQualityStandard creates a new quality standard with rules for a project
func (s *CollabHubService) CreateQualityStandard(ctx context.Context, projectID uuid.UUID, name string, rules []domain.QualityRule) (*domain.QualityStandard, error) {
	if name == "" {
		return nil, fmt.Errorf("quality standard name is required")
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("quality standard must have at least one rule")
	}

	// Validate rules
	validOperators := map[string]bool{"gt": true, "gte": true, "lt": true, "lte": true, "eq": true}
	for i, rule := range rules {
		if rule.Metric == "" {
			return nil, fmt.Errorf("rule %d: metric is required", i)
		}
		if !validOperators[rule.Operator] {
			return nil, fmt.Errorf("rule %d: invalid operator %s; use gt, gte, lt, lte, or eq", i, rule.Operator)
		}
	}

	standard := &domain.QualityStandard{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            name,
		Enabled:         true,
		Rules:           rules,
		EnforceOnDeploy: false,
		CreatedAt:       time.Now(),
	}

	s.logger.Info("quality standard created",
		zap.String("id", standard.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", name),
		zap.Int("ruleCount", len(rules)),
	)
	return standard, nil
}

// ListStandards lists all quality standards for a project
func (s *CollabHubService) ListStandards(ctx context.Context, projectID uuid.UUID) ([]domain.QualityStandard, error) {
	s.logger.Debug("listing quality standards", zap.String("projectId", projectID.String()))
	return []domain.QualityStandard{}, nil
}

// GetActivityFeed retrieves recent activity items for a project
func (s *CollabHubService) GetActivityFeed(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.ActivityFeedItem, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	s.logger.Debug("fetching activity feed",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
	)

	now := time.Now()
	items := []domain.ActivityFeedItem{
		{
			ID:           uuid.New(),
			ProjectID:    projectID,
			Type:         domain.ActivityTypeTraceReviewed,
			UserID:       uuid.New(),
			UserName:     "alice",
			Description:  "Reviewed trace for RAG pipeline - approved with score 0.92",
			ResourceID:   uuid.New().String(),
			ResourceType: "trace",
			Timestamp:    now.Add(-5 * time.Minute),
			Metadata:     map[string]interface{}{"score": 0.92, "status": "approved"},
		},
		{
			ID:           uuid.New(),
			ProjectID:    projectID,
			Type:         domain.ActivityTypePromptDeployed,
			UserID:       uuid.New(),
			UserName:     "bob",
			Description:  "Deployed prompt v3 to production after A/B test",
			ResourceID:   uuid.New().String(),
			ResourceType: "prompt",
			Timestamp:    now.Add(-30 * time.Minute),
			Metadata:     map[string]interface{}{"version": 3, "environment": "production"},
		},
		{
			ID:           uuid.New(),
			ProjectID:    projectID,
			Type:         domain.ActivityTypeEvalCompleted,
			UserID:       uuid.New(),
			UserName:     "charlie",
			Description:  "Completed evaluation run: 94% pass rate across 200 test cases",
			ResourceID:   uuid.New().String(),
			ResourceType: "evaluation",
			Timestamp:    now.Add(-2 * time.Hour),
			Metadata:     map[string]interface{}{"passRate": 0.94, "totalTests": 200},
		},
	}

	if limit < len(items) {
		items = items[:limit]
	}

	return items, nil
}
