package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CollaborationRepository defines repository operations for collaboration data
type CollaborationRepository interface {
	SaveAnnotation(ctx context.Context, annotation *domain.TraceAnnotation) error
	GetAnnotationByID(ctx context.Context, id uuid.UUID) (*domain.TraceAnnotation, error)
	ListAnnotations(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TraceAnnotation, error)
	UpdateAnnotation(ctx context.Context, annotation *domain.TraceAnnotation) error
	SaveSharedSession(ctx context.Context, session *domain.SharedSession) error
	GetSharedSessionByID(ctx context.Context, id uuid.UUID) (*domain.SharedSession, error)
}

// CollaborationService manages real-time collaboration on traces
type CollaborationService struct {
	logger      *zap.Logger
	collabRepo  CollaborationRepository
	realtimeSvc *RealtimeService
}

// NewCollaborationService creates a new collaboration service
func NewCollaborationService(
	logger *zap.Logger,
	collabRepo CollaborationRepository,
	realtimeSvc *RealtimeService,
) *CollaborationService {
	return &CollaborationService{
		logger:      logger,
		collabRepo:  collabRepo,
		realtimeSvc: realtimeSvc,
	}
}

// UpdatePresence tracks which user is viewing which trace
func (s *CollaborationService) UpdatePresence(ctx context.Context, presence *domain.UserPresence) error {
	presence.LastSeen = time.Now()

	// Broadcast presence update to other viewers
	s.realtimeSvc.Publish(ctx, presence.ProjectID, "presence_update", presence)

	return nil
}

// GetPresence retrieves all users currently viewing a trace
func (s *CollaborationService) GetPresence(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.UserPresence, error) {
	// In a real implementation, this would query an in-memory store or cache
	// for active presence records filtered by projectID and traceID
	return []domain.UserPresence{}, nil
}

// AddAnnotation creates a new annotation on a trace event
func (s *CollaborationService) AddAnnotation(ctx context.Context, projectID uuid.UUID, annotation *domain.TraceAnnotation) (*domain.TraceAnnotation, error) {
	annotation.ID = uuid.New()
	annotation.ProjectID = projectID
	annotation.CreatedAt = time.Now()

	if err := s.collabRepo.SaveAnnotation(ctx, annotation); err != nil {
		return nil, fmt.Errorf("failed to save annotation: %w", err)
	}

	// Broadcast annotation to collaborators
	s.realtimeSvc.Publish(ctx, projectID, "annotation_added", annotation)

	s.logger.Info("annotation added",
		zap.String("annotationId", annotation.ID.String()),
		zap.String("traceId", annotation.TraceID),
		zap.String("userId", annotation.UserID.String()),
	)

	return annotation, nil
}

// ListAnnotations retrieves all annotations for a trace
func (s *CollaborationService) ListAnnotations(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TraceAnnotation, error) {
	annotations, err := s.collabRepo.ListAnnotations(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations: %w", err)
	}
	return annotations, nil
}

// ResolveAnnotation marks an annotation as resolved
func (s *CollaborationService) ResolveAnnotation(ctx context.Context, annotationID uuid.UUID, userID uuid.UUID) error {
	annotation, err := s.collabRepo.GetAnnotationByID(ctx, annotationID)
	if err != nil {
		return fmt.Errorf("failed to get annotation: %w", err)
	}

	now := time.Now()
	annotation.ResolvedAt = &now

	if err := s.collabRepo.UpdateAnnotation(ctx, annotation); err != nil {
		return fmt.Errorf("failed to resolve annotation: %w", err)
	}

	// Broadcast resolution to collaborators
	s.realtimeSvc.Publish(ctx, annotation.ProjectID, "annotation_resolved", annotation)

	return nil
}

// CreateSharedSession creates a new shared collaboration session on a trace
func (s *CollaborationService) CreateSharedSession(ctx context.Context, projectID uuid.UUID, input *domain.SharedSession) (*domain.SharedSession, error) {
	input.ID = uuid.New()
	input.ProjectID = projectID
	input.CreatedAt = time.Now()

	if err := s.collabRepo.SaveSharedSession(ctx, input); err != nil {
		return nil, fmt.Errorf("failed to create shared session: %w", err)
	}

	s.logger.Info("shared session created",
		zap.String("sessionId", input.ID.String()),
		zap.String("traceId", input.TraceID),
	)

	return input, nil
}

// GetSharedSession retrieves a shared session by ID
func (s *CollaborationService) GetSharedSession(ctx context.Context, sessionID uuid.UUID) (*domain.SharedSession, error) {
	session, err := s.collabRepo.GetSharedSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared session: %w", err)
	}
	return session, nil
}

// CreateDiscussionThread creates a new discussion thread
func (s *CollaborationService) CreateDiscussionThread(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, userName string, input *domain.DiscussionInput) (*domain.DiscussionThread, error) {
	thread := &domain.DiscussionThread{
		ID:             uuid.New(),
		TraceID:        input.TraceID,
		ObservationID:  input.ObservationID,
		Title:          input.Title,
		Status:         "open",
		CreatedBy:      userID,
		CreatedByName:  userName,
		ParticipantIDs: []uuid.UUID{userID},
		Messages: []domain.ThreadMessage{
			{
				ID:        uuid.New(),
				UserID:    userID,
				UserName:  userName,
				Content:   input.InitialMessage,
				CreatedAt: time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.logger.Info("created discussion thread",
		zap.String("threadId", thread.ID.String()),
		zap.String("traceId", thread.TraceID.String()),
	)

	return thread, nil
}

// AddThreadMessage adds a message to a discussion thread
func (s *CollaborationService) AddThreadMessage(ctx context.Context, threadID, userID uuid.UUID, userName, content string, mentions []uuid.UUID) (*domain.ThreadMessage, error) {
	msg := &domain.ThreadMessage{
		ID:        uuid.New(),
		ThreadID:  threadID,
		UserID:    userID,
		UserName:  userName,
		Content:   content,
		Mentions:  mentions,
		CreatedAt: time.Now(),
	}

	s.logger.Info("added thread message",
		zap.String("threadId", threadID.String()),
		zap.String("userId", userID.String()),
	)

	return msg, nil
}

// CreateEvalQueue creates a new shared evaluation queue
func (s *CollaborationService) CreateEvalQueue(ctx context.Context, projectID uuid.UUID, input *domain.EvalQueueInput) (*domain.EvalQueue, error) {
	queue := &domain.EvalQueue{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Assignees:   input.Assignees,
		TraceIDs:    input.TraceIDs,
		Status:      "active",
		Progress: domain.EvalQueueProgress{
			Total:   len(input.TraceIDs),
			Pending: len(input.TraceIDs),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.logger.Info("created eval queue",
		zap.String("queueId", queue.ID.String()),
		zap.Int("traceCount", len(input.TraceIDs)),
	)

	return queue, nil
}

// CreateReview creates a new trace review request
func (s *CollaborationService) CreateReview(ctx context.Context, projectID, requestedBy uuid.UUID, input *domain.CollabTraceReviewInput) (*domain.TraceReviewRequest, error) {
	if input.TraceID == "" {
		return nil, fmt.Errorf("traceID is required")
	}
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	now := time.Now()
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}
	requiredApprovals := input.RequiredApprovals
	if requiredApprovals == 0 {
		requiredApprovals = 1
	}

	review := &domain.TraceReviewRequest{
		ID:                uuid.New(),
		ProjectID:         projectID,
		TraceID:           input.TraceID,
		Title:             input.Title,
		Description:       input.Description,
		RequestedBy:       requestedBy,
		AssignedTo:        input.AssignedTo,
		Status:            domain.ReviewStatusPending,
		Priority:          priority,
		Labels:            input.Labels,
		ApprovalCount:     0,
		RequiredApprovals: requiredApprovals,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.realtimeSvc.Publish(ctx, projectID, "review_created", review)

	s.logger.Info("created trace review",
		zap.String("reviewId", review.ID.String()),
		zap.String("traceId", input.TraceID),
	)

	return review, nil
}

// GetReview retrieves a trace review by ID
func (s *CollaborationService) GetReview(ctx context.Context, reviewID uuid.UUID) (*domain.TraceReviewRequest, error) {
	// In production this would query the database
	s.logger.Info("get review", zap.String("reviewId", reviewID.String()))
	return &domain.TraceReviewRequest{
		ID:     reviewID,
		Status: domain.ReviewStatusPending,
	}, nil
}

// ListReviews lists trace reviews for a project
func (s *CollaborationService) ListReviews(ctx context.Context, projectID uuid.UUID, status string) ([]domain.TraceReviewRequest, error) {
	s.logger.Info("listing reviews",
		zap.String("projectId", projectID.String()),
		zap.String("status", status),
	)
	return []domain.TraceReviewRequest{}, nil
}

// ApproveReview approves a trace review
func (s *CollaborationService) ApproveReview(ctx context.Context, reviewID, approverID uuid.UUID) (*domain.TraceReviewRequest, error) {
	review, err := s.GetReview(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	review.ApprovalCount++
	review.UpdatedAt = time.Now()
	if review.ApprovalCount >= review.RequiredApprovals {
		review.Status = domain.ReviewStatusApproved
	} else {
		review.Status = domain.ReviewStatusInReview
	}

	s.realtimeSvc.Publish(ctx, review.ProjectID, "review_approved", review)

	s.logger.Info("review approved",
		zap.String("reviewId", reviewID.String()),
		zap.String("approver", approverID.String()),
		zap.Int("approvalCount", review.ApprovalCount),
	)

	return review, nil
}

// AddReviewComment adds a threaded comment to a review
func (s *CollaborationService) AddReviewComment(ctx context.Context, reviewID, authorID uuid.UUID, authorName string, input *domain.CollabReviewCommentInput) (*domain.CollabReviewComment, error) {
	if input.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	now := time.Now()
	comment := &domain.CollabReviewComment{
		ID:         uuid.New(),
		ReviewID:   reviewID,
		ParentID:   input.ParentID,
		AuthorID:   authorID,
		AuthorName: authorName,
		Content:    input.Content,
		Mentions:   input.Mentions,
		SpanID:     input.SpanID,
		Resolved:   false,
		Reactions:  map[string]int{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.logger.Info("added review comment",
		zap.String("reviewId", reviewID.String()),
		zap.String("commentId", comment.ID.String()),
		zap.String("authorId", authorID.String()),
	)

	return comment, nil
}

// ListReviewComments lists comments for a review
func (s *CollaborationService) ListReviewComments(ctx context.Context, reviewID uuid.UUID) ([]domain.CollabReviewComment, error) {
	s.logger.Info("listing review comments", zap.String("reviewId", reviewID.String()))
	return []domain.CollabReviewComment{}, nil
}

// CreateReviewQueue creates a new review queue with assignment rules
func (s *CollaborationService) CreateReviewQueue(ctx context.Context, projectID uuid.UUID, input *domain.CollabReviewQueueInput) (*domain.CollabReviewQueue, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	assignmentRule := input.AssignmentRule
	if assignmentRule == "" {
		assignmentRule = "round_robin"
	}
	slaHours := input.SLAHours
	if slaHours == 0 {
		slaHours = 24
	}

	queue := &domain.CollabReviewQueue{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		AssignmentRule:  assignmentRule,
		Reviewers:       input.Reviewers,
		AutoAssign:      input.AutoAssign,
		SLAHours:        slaHours,
		EscalationHours: input.EscalationHours,
		PendingCount:    0,
		AvgReviewTimeH:  0,
		CreatedAt:       time.Now(),
	}

	s.logger.Info("created review queue",
		zap.String("queueId", queue.ID.String()),
		zap.String("name", input.Name),
	)

	return queue, nil
}

// ListReviewQueues lists review queues for a project
func (s *CollaborationService) ListReviewQueues(ctx context.Context, projectID uuid.UUID) ([]domain.CollabReviewQueue, error) {
	s.logger.Info("listing review queues", zap.String("projectId", projectID.String()))
	return []domain.CollabReviewQueue{}, nil
}

// AddNotificationIntegration adds a notification integration (Slack/Teams/GitHub)
func (s *CollaborationService) AddNotificationIntegration(ctx context.Context, projectID uuid.UUID, input *domain.NotificationIntegrationInput) (*domain.NotificationIntegration, error) {
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	integration := &domain.NotificationIntegration{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Type:       input.Type,
		Name:       input.Name,
		WebhookURL: input.WebhookURL,
		ChannelID:  input.ChannelID,
		Enabled:    true,
		Events:     input.Events,
		CreatedAt:  time.Now(),
	}

	s.logger.Info("added notification integration",
		zap.String("integrationId", integration.ID.String()),
		zap.String("type", input.Type),
	)

	return integration, nil
}

// ListIntegrations lists notification integrations for a project
func (s *CollaborationService) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]domain.NotificationIntegration, error) {
	s.logger.Info("listing integrations", zap.String("projectId", projectID.String()))
	return []domain.NotificationIntegration{}, nil
}
