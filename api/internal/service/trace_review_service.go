package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceReviewService manages collaborative trace reviews
type TraceReviewService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	reviews map[uuid.UUID]*domain.TraceReview
}

// NewTraceReviewService creates a new trace review service
func NewTraceReviewService(logger *zap.Logger) *TraceReviewService {
	return &TraceReviewService{
		logger:  logger,
		reviews: make(map[uuid.UUID]*domain.TraceReview),
	}
}

// CreateReview creates a new trace review
func (s *TraceReviewService) CreateReview(ctx context.Context, projectID, userID uuid.UUID, input *domain.TraceReviewInput) (*domain.TraceReview, error) {
	if input.TraceID == "" {
		return nil, fmt.Errorf("trace ID is required")
	}
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	priority := input.Priority
	if priority == "" {
		priority = domain.ReviewPriorityMedium
	}

	review := &domain.TraceReview{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     input.TraceID,
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TraceReviewStatusOpen,
		Priority:    priority,
		Assignee:    input.Assignee,
		Comments:    []domain.ReviewComment{},
		Labels:      input.Labels,
		DueAt:       input.DueAt,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.reviews[review.ID] = review
	s.mu.Unlock()

	s.logger.Info("trace review created",
		zap.String("reviewId", review.ID.String()),
		zap.String("traceId", review.TraceID),
		zap.String("priority", string(review.Priority)),
	)

	return review, nil
}

// GetReview retrieves a review by ID
func (s *TraceReviewService) GetReview(ctx context.Context, id uuid.UUID) (*domain.TraceReview, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	review, exists := s.reviews[id]
	if !exists {
		return nil, fmt.Errorf("review not found")
	}
	return review, nil
}

// ListReviews lists reviews for a project
func (s *TraceReviewService) ListReviews(ctx context.Context, projectID uuid.UUID, status *domain.TraceReviewStatus) ([]domain.TraceReview, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var reviews []domain.TraceReview
	for _, review := range s.reviews {
		if review.ProjectID == projectID {
			if status != nil && review.Status != *status {
				continue
			}
			reviews = append(reviews, *review)
		}
	}

	sort.Slice(reviews, func(i, j int) bool {
		return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
	})

	return reviews, nil
}

// UpdateReview updates a review
func (s *TraceReviewService) UpdateReview(ctx context.Context, id uuid.UUID, input *domain.TraceReviewInput) (*domain.TraceReview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	review, exists := s.reviews[id]
	if !exists {
		return nil, fmt.Errorf("review not found")
	}

	if input.Title != "" {
		review.Title = input.Title
	}
	if input.Description != "" {
		review.Description = input.Description
	}
	if input.Priority != "" {
		review.Priority = input.Priority
	}
	if input.Assignee != nil {
		review.Assignee = input.Assignee
	}
	if len(input.Labels) > 0 {
		review.Labels = input.Labels
	}
	review.UpdatedAt = time.Now()

	return review, nil
}

// AddComment adds a comment to a review
func (s *TraceReviewService) AddComment(ctx context.Context, reviewID, userID uuid.UUID, input *domain.ReviewCommentInput) (*domain.ReviewComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	review, exists := s.reviews[reviewID]
	if !exists {
		return nil, fmt.Errorf("review not found")
	}

	if input.Content == "" {
		return nil, fmt.Errorf("comment content is required")
	}

	comment := domain.ReviewComment{
		ID:            uuid.New(),
		ReviewID:      reviewID,
		AuthorID:      userID,
		AuthorName:    "reviewer",
		Content:       input.Content,
		ObservationID: input.ObservationID,
		SpanPath:      input.SpanPath,
		CreatedAt:     time.Now(),
	}

	review.Comments = append(review.Comments, comment)
	review.Status = domain.TraceReviewStatusInReview
	review.UpdatedAt = time.Now()

	s.logger.Info("review comment added",
		zap.String("reviewId", reviewID.String()),
		zap.Int("totalComments", len(review.Comments)),
	)

	return &comment, nil
}

// Approve approves a trace review
func (s *TraceReviewService) Approve(ctx context.Context, reviewID, userID uuid.UUID) (*domain.TraceReview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	review, exists := s.reviews[reviewID]
	if !exists {
		return nil, fmt.Errorf("review not found")
	}

	review.Status = domain.TraceReviewStatusApproved
	review.Reviewer = &userID
	now := time.Now()
	review.ResolvedAt = &now
	review.UpdatedAt = now

	s.logger.Info("trace review approved",
		zap.String("reviewId", reviewID.String()),
	)

	return review, nil
}

// Reject rejects a trace review
func (s *TraceReviewService) Reject(ctx context.Context, reviewID, userID uuid.UUID, reason string) (*domain.TraceReview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	review, exists := s.reviews[reviewID]
	if !exists {
		return nil, fmt.Errorf("review not found")
	}

	review.Status = domain.TraceReviewStatusRejected
	review.Reviewer = &userID
	now := time.Now()
	review.ResolvedAt = &now
	review.UpdatedAt = now

	if reason != "" {
		review.Comments = append(review.Comments, domain.ReviewComment{
			ID:         uuid.New(),
			ReviewID:   reviewID,
			AuthorID:   userID,
			AuthorName: "reviewer",
			Content:    fmt.Sprintf("Rejected: %s", reason),
			CreatedAt:  now,
		})
	}

	return review, nil
}

// GetQueue returns the review queue
func (s *TraceReviewService) GetQueue(ctx context.Context, projectID uuid.UUID) (*domain.TraceReviewQueue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []domain.TraceReview
	overdue := 0
	now := time.Now()

	for _, review := range s.reviews {
		if review.ProjectID == projectID && (review.Status == domain.TraceReviewStatusOpen || review.Status == domain.TraceReviewStatusInReview) {
			pending = append(pending, *review)
			if review.DueAt != nil && review.DueAt.Before(now) {
				overdue++
			}
		}
	}

	// Sort by priority then creation time
	priorityOrder := map[domain.ReviewPriority]int{
		domain.ReviewPriorityCritical: 0,
		domain.ReviewPriorityHigh:     1,
		domain.ReviewPriorityMedium:   2,
		domain.ReviewPriorityLow:      3,
	}

	sort.Slice(pending, func(i, j int) bool {
		pi := priorityOrder[pending[i].Priority]
		pj := priorityOrder[pending[j].Priority]
		if pi != pj {
			return pi < pj
		}
		return pending[i].CreatedAt.Before(pending[j].CreatedAt)
	})

	return &domain.TraceReviewQueue{
		Reviews:    pending,
		TotalCount: len(pending),
		OverdueSLA: overdue,
	}, nil
}
