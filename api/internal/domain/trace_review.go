package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReviewStatus represents the status of a trace review
type TraceReviewStatus string

const (
	TraceReviewStatusOpen      TraceReviewStatus = "open"
	TraceReviewStatusInReview  TraceReviewStatus = "in_review"
	TraceReviewStatusApproved  TraceReviewStatus = "approved"
	TraceReviewStatusRejected  TraceReviewStatus = "rejected"
	TraceReviewStatusRerunReq  TraceReviewStatus = "rerun_requested"
)

// ReviewPriority represents the priority of a review
type ReviewPriority string

const (
	ReviewPriorityLow      ReviewPriority = "low"
	ReviewPriorityMedium   ReviewPriority = "medium"
	ReviewPriorityHigh     ReviewPriority = "high"
	ReviewPriorityCritical ReviewPriority = "critical"
)

// TraceReview represents a review of an agent trace
type TraceReview struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	TraceID     string         `json:"traceId"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Status      TraceReviewStatus   `json:"status"`
	Priority    ReviewPriority `json:"priority"`
	Assignee    *uuid.UUID     `json:"assignee,omitempty"`
	Reviewer    *uuid.UUID     `json:"reviewer,omitempty"`
	Comments    []ReviewComment `json:"comments"`
	Labels      []string       `json:"labels,omitempty"`
	DueAt       *time.Time     `json:"dueAt,omitempty"`
	CreatedBy   uuid.UUID      `json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	ResolvedAt  *time.Time     `json:"resolvedAt,omitempty"`
}

// ReviewComment represents a comment on a trace review
type ReviewComment struct {
	ID           uuid.UUID `json:"id"`
	ReviewID     uuid.UUID `json:"reviewId"`
	AuthorID     uuid.UUID `json:"authorId"`
	AuthorName   string    `json:"authorName"`
	Content      string    `json:"content"`
	ObservationID *string  `json:"observationId,omitempty"` // link to specific observation
	SpanPath     string    `json:"spanPath,omitempty"`
	IsResolved   bool      `json:"isResolved"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ReviewQueue represents a queue of pending reviews
type TraceReviewQueue struct {
	Reviews    []TraceReview `json:"reviews"`
	TotalCount int           `json:"totalCount"`
	OverdueSLA int           `json:"overdueSla"`
}

// TraceReviewInput represents input for creating a review
type TraceReviewInput struct {
	TraceID     string         `json:"traceId" validate:"required"`
	Title       string         `json:"title" validate:"required"`
	Description string         `json:"description,omitempty"`
	Priority    ReviewPriority `json:"priority,omitempty"`
	Assignee    *uuid.UUID     `json:"assignee,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
	DueAt       *time.Time     `json:"dueAt,omitempty"`
}

// ReviewCommentInput represents input for adding a comment
type ReviewCommentInput struct {
	Content       string  `json:"content" validate:"required"`
	ObservationID *string `json:"observationId,omitempty"`
	SpanPath      string  `json:"spanPath,omitempty"`
}
